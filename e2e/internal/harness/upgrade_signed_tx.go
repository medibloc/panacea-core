package harness

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

func upgradeSignedTxContainerPath(homeDir, relativePath string) (string, error) {
	if strings.TrimSpace(homeDir) == "" {
		return "", errors.New("node home directory is required")
	}
	if strings.TrimSpace(relativePath) == "" || path.IsAbs(relativePath) {
		return "", errors.New("signed transaction path must be relative")
	}
	clean := path.Clean(relativePath)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("signed transaction path escapes node home: %q", relativePath)
	}
	return path.Join(homeDir, clean), nil
}

func classifyAcceptedCheckTx(result TxResult) error {
	if strings.TrimSpace(result.Height) != "0" {
		if result.HeightInt64() > 0 {
			return fmt.Errorf("transaction reached committed height %s instead of returning CheckTx evidence", result.Height)
		}
		return fmt.Errorf("CheckTx acceptance returned invalid broadcast height %q", result.Height)
	}
	if result.Code != 0 {
		return fmt.Errorf(
			"CheckTx failed: codespace=%s code=%d raw_log=%s",
			result.Codespace,
			result.Code,
			result.RawLog,
		)
	}
	return nil
}

// BroadcastSignedTxFileCheckTx sync-broadcasts a prebuilt signed transaction
// from the node home and returns only accepted CheckTx evidence. It deliberately
// does not poll query-tx or wait for the transaction to commit.
func (n *Network) BroadcastSignedTxFileCheckTx(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	relativePath string,
) (*TxResult, error) {
	n.txMu.Lock()
	defer n.txMu.Unlock()

	_, result, err := n.broadcastSignedTxFileCheckTxLocked(
		ctx,
		step,
		node,
		relativePath,
		"checktx_acceptance",
		"",
		0,
	)
	return result, err
}

// BroadcastSignedTxFileExpectCheckTxFailure sync-broadcasts a prebuilt signed
// transaction and proves that CheckTx rejected it with the exact stable
// codespace and code. An expected rejection is returned as evidence and does
// not wait for, or claim, a committed transaction.
func (n *Network) BroadcastSignedTxFileExpectCheckTxFailure(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	relativePath string,
	expectedCodespace string,
	expectedCode uint32,
) (*TxResult, error) {
	if strings.TrimSpace(expectedCodespace) == "" {
		return nil, errors.New("expected CheckTx codespace is required")
	}
	if expectedCode == 0 {
		return nil, errors.New("expected CheckTx code must be nonzero")
	}

	n.txMu.Lock()
	defer n.txMu.Unlock()

	_, result, err := n.broadcastSignedTxFileCheckTxLocked(
		ctx,
		step,
		node,
		relativePath,
		"checktx_rejection",
		expectedCodespace,
		expectedCode,
	)
	return result, err
}

func (n *Network) broadcastSignedTxFileCheckTxLocked(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	relativePath string,
	expectedOutcome string,
	expectedCodespace string,
	expectedCode uint32,
) (string, *TxResult, error) {
	if strings.TrimSpace(step) == "" {
		return "", nil, errors.New("transaction step is required")
	}
	if node == nil {
		return "", nil, errors.New("transaction node is required")
	}
	containerPath, err := upgradeSignedTxContainerPath(node.HomeDir(), relativePath)
	if err != nil {
		return "", nil, err
	}

	requestID := fmt.Sprintf("%s-%d", step, time.Now().UTC().UnixNano())
	request := map[string]any{
		"request_id":       requestID,
		"recorded_at":      time.Now().UTC(),
		"step":             step,
		"node":             node.Name(),
		"arguments":        []string{"broadcast-signed-file", relativePath},
		"expected_outcome": expectedOutcome,
	}
	if expectedCode != 0 {
		request["expected_codespace"] = expectedCodespace
		request["expected_code"] = expectedCode
	}
	if err := n.artifacts.appendJSONLine("tx/requests.jsonl", request); err != nil {
		n.artifacts.recordFailure("record-tx-request", err)
		return requestID, nil, fmt.Errorf("record signed transaction request %s: %w", step, err)
	}

	stdout, stderr, execErr := node.Exec(ctx, node.NodeCommand(
		"tx", "broadcast", containerPath,
		"--broadcast-mode", "sync",
		"--output", "json",
	), node.Chain.Config().Env)
	broadcastEvidence := map[string]any{
		"request_id":       requestID,
		"recorded_at":      time.Now().UTC(),
		"step":             step,
		"stdout":           jsonOrString(stdout),
		"stderr":           boundedString(stderr, txStderrMaxBytes),
		"exec_error":       errorString(execErr),
		"expected_outcome": expectedOutcome,
	}
	if expectedCode != 0 {
		broadcastEvidence["expected_codespace"] = expectedCodespace
		broadcastEvidence["expected_code"] = expectedCode
	}
	if err := n.artifacts.appendJSONLine("tx/broadcast-results.jsonl", broadcastEvidence); err != nil {
		n.artifacts.recordFailure("record-tx-broadcast", err)
		return requestID, nil, fmt.Errorf("record signed transaction broadcast %s: %w", step, err)
	}
	if execErr != nil {
		err := fmt.Errorf("broadcast signed transaction %s: %w: %s", step, execErr, boundedString(stderr, txStderrMaxBytes))
		n.artifacts.recordFailure("broadcast-tx-"+step, err)
		return requestID, nil, err
	}

	broadcast, err := parseTxResult(stdout)
	if err != nil {
		err = fmt.Errorf("decode CheckTx response for %s: %w", step, err)
		n.artifacts.recordFailure("decode-checktx-"+step, err)
		return requestID, nil, err
	}
	if expectedCode == 0 {
		err = classifyAcceptedCheckTx(broadcast)
	} else {
		err = classifyExpectedCheckTxFailure(broadcast, expectedCodespace, expectedCode)
	}
	if err != nil {
		err = fmt.Errorf("CheckTx %s: %w", step, err)
		n.artifacts.recordFailure("checktx-"+step, err)
		return requestID, &broadcast, err
	}
	return requestID, &broadcast, nil
}

// BroadcastSignedTxFileAndWaitDeliverFailure broadcasts a prebuilt signed
// transaction from the node home. Keeping the transaction as a file supports
// historical message types whose CLI command was intentionally removed in the
// new release, including bytes generated and signed before an upgrade.
func (n *Network) BroadcastSignedTxFileAndWaitDeliverFailure(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	relativePath string,
	expectedCodespace string,
	expectedCode uint32,
) (*TxResult, error) {
	if strings.TrimSpace(expectedCodespace) == "" || expectedCode == 0 {
		return nil, errors.New("expected deliver failure codespace and non-zero code are required")
	}
	return n.broadcastSignedTxFileAndWait(ctx, step, node, relativePath, expectedCodespace, expectedCode)
}

// BroadcastSignedTxFileAndWait broadcasts a prebuilt signed transaction and
// requires both CheckTx and the committed FinalizeBlock result to succeed. It
// is the positive counterpart of BroadcastSignedTxFileAndWaitDeliverFailure
// and is used to prove that exact bytes signed by an old binary remain valid
// after an in-place upgrade.
func (n *Network) BroadcastSignedTxFileAndWait(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	relativePath string,
) (*TxResult, error) {
	return n.broadcastSignedTxFileAndWait(ctx, step, node, relativePath, "", 0)
}

func (n *Network) broadcastSignedTxFileAndWait(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	relativePath string,
	expectedCodespace string,
	expectedCode uint32,
) (*TxResult, error) {
	n.txMu.Lock()
	defer n.txMu.Unlock()

	if expectedCode != 0 && strings.TrimSpace(expectedCodespace) == "" {
		return nil, errors.New("expected deliver failure codespace and non-zero code are required")
	}
	requestID, broadcast, err := n.broadcastSignedTxFileCheckTxLocked(
		ctx,
		step,
		node,
		relativePath,
		"checktx_acceptance",
		"",
		0,
	)
	if err != nil {
		return broadcast, err
	}

	committed, err := n.waitForCommittedTx(ctx, requestID, step, broadcast.TxHash)
	if err != nil {
		return committed, err
	}
	if expectedCode == 0 {
		if committed.Code != 0 {
			err = fmt.Errorf(
				"FinalizeBlock %s failed: codespace=%s code=%d raw_log=%s",
				step,
				committed.Codespace,
				committed.Code,
				committed.RawLog,
			)
			n.artifacts.recordFailure("finalize-block-"+step, err)
			return committed, err
		}
		return committed, nil
	}
	if committed.Code != expectedCode || committed.Codespace != expectedCodespace {
		err = fmt.Errorf(
			"FinalizeBlock %s returned codespace=%s code=%d, want codespace=%s code=%d",
			step,
			committed.Codespace,
			committed.Code,
			expectedCodespace,
			expectedCode,
		)
		n.artifacts.recordFailure("finalize-block-"+step, err)
		return committed, err
	}
	return committed, nil
}
