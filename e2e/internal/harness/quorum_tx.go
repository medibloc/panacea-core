package harness

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

const quorumFixedGas = "500000"

// QuorumPendingTx retains the accepted CheckTx response while consensus is
// halted so the same hash can be proven uncommitted and later committed after
// quorum recovery.
type QuorumPendingTx struct {
	RequestID string   `json:"request_id"`
	Step      string   `json:"step"`
	TxHash    string   `json:"tx_hash"`
	CheckTx   TxResult `json:"check_tx"`
}

// QuorumTxLookupSample records both full-node reachability and transaction
// index visibility. A lookup error only proves absence when it is CometBFT's
// explicit tx-not-found response.
type QuorumTxLookupSample struct {
	ObservedAt     time.Time `json:"observed_at"`
	FullNodeHeight int64     `json:"full_node_height,omitempty"`
	CatchingUp     bool      `json:"catching_up,omitempty"`
	LookupError    string    `json:"lookup_error,omitempty"`
	Committed      bool      `json:"committed"`
	CommitHeight   int64     `json:"commit_height,omitempty"`
}

// VerifyQuorumNoCommitEvidence requires a live, caught-up full node to have
// returned CometBFT's explicit tx-not-found result near the end of a bounded
// window, and rejects any sample that reports a commit.
func VerifyQuorumNoCommitEvidence(
	samples []QuorumTxLookupSample,
	observedUntil time.Time,
	maximumEvidenceAge time.Duration,
) error {
	if observedUntil.IsZero() {
		return errors.New("no-commit observation end time is required")
	}
	if maximumEvidenceAge <= 0 {
		return errors.New("maximum no-commit evidence age must be positive")
	}
	var latestNotFound time.Time
	for _, sample := range samples {
		if sample.Committed {
			return fmt.Errorf("transaction committed at height %d during no-commit observation", sample.CommitHeight)
		}
		if sample.FullNodeHeight <= 0 || sample.CatchingUp || sample.ObservedAt.IsZero() {
			continue
		}
		if !isQuorumTxNotFound(sample.LookupError) {
			continue
		}
		if sample.ObservedAt.After(observedUntil) {
			return fmt.Errorf("tx-not-found evidence timestamp %s is after observation end %s", sample.ObservedAt, observedUntil)
		}
		if sample.ObservedAt.After(latestNotFound) {
			latestNotFound = sample.ObservedAt
		}
	}
	if latestNotFound.IsZero() {
		return errors.New("full node never returned recent explicit tx-not-found evidence")
	}
	if age := observedUntil.Sub(latestNotFound); age > maximumEvidenceAge {
		return fmt.Errorf("latest full-node tx-not-found evidence is stale: age=%s", age.Round(time.Millisecond))
	}
	return nil
}

func isQuorumTxNotFound(message string) bool {
	normalized := strings.ToLower(message)
	return strings.Contains(normalized, "tx (") && strings.Contains(normalized, ") not found")
}

// BroadcastQuorumTxAndObserveNoCommit broadcasts with fixed gas in sync mode,
// proves CheckTx acceptance, then holds the harness transaction mutex while an
// explicit full node remains reachable and reports tx-not-found throughout a
// bounded consensus-halt window.
func (n *Network) BroadcastQuorumTxAndObserveNoCommit(
	ctx context.Context,
	observationWindow time.Duration,
	step string,
	node *cosmos.ChainNode,
	keyName string,
	command ...string,
) (*QuorumPendingTx, error) {
	if n == nil || n.Chain == nil {
		return nil, errors.New("quorum network is required")
	}
	if observationWindow <= 0 {
		return nil, errors.New("no-commit observation window must be positive")
	}
	if strings.TrimSpace(step) == "" {
		return nil, errors.New("transaction step is required")
	}
	if node == nil {
		return nil, errors.New("transaction node is required")
	}
	if strings.TrimSpace(keyName) == "" {
		return nil, errors.New("transaction key name is required")
	}
	if len(command) == 0 {
		return nil, errors.New("transaction command is required")
	}
	if len(n.Chain.FullNodes) == 0 {
		return nil, errors.New("network has no explicit full node")
	}
	for _, argument := range command {
		if argument == "--gas" || strings.HasPrefix(argument, "--gas=") ||
			argument == "--broadcast-mode" || strings.HasPrefix(argument, "--broadcast-mode=") {
			return nil, fmt.Errorf("quorum transaction flags are harness-owned: %s", argument)
		}
	}

	n.txMu.Lock()
	defer n.txMu.Unlock()

	arguments := append([]string(nil), command...)
	arguments = append(arguments, "--gas", quorumFixedGas, "--broadcast-mode", "sync")
	requestID := fmt.Sprintf("%s-%d", step, time.Now().UTC().UnixNano())
	request := map[string]any{
		"request_id":  requestID,
		"recorded_at": time.Now().UTC(),
		"step":        step,
		"node":        node.Name(),
		"key_name":    keyName,
		"arguments":   arguments,
		"expectation": "checktx-accepted-and-not-committed",
	}
	if err := n.artifacts.appendJSONLine("tx/requests.jsonl", request); err != nil {
		return nil, fmt.Errorf("record quorum transaction request %s: %w", step, err)
	}

	stdout, stderr, execErr := node.Exec(ctx, node.TxCommand(keyName, arguments...), node.Chain.Config().Env)
	broadcastEvidence := map[string]any{
		"request_id":  requestID,
		"recorded_at": time.Now().UTC(),
		"step":        step,
		"stdout":      jsonOrString(stdout),
		"stderr":      boundedString(stderr, txStderrMaxBytes),
		"exec_error":  quorumErrorString(execErr),
	}
	if err := n.artifacts.appendJSONLine("tx/broadcast-results.jsonl", broadcastEvidence); err != nil {
		return nil, fmt.Errorf("record quorum transaction broadcast %s: %w", step, err)
	}
	if execErr != nil {
		err := fmt.Errorf("broadcast quorum transaction %s: %w: %s", step, execErr, boundedString(stderr, txStderrMaxBytes))
		n.artifacts.recordFailure("quorum-broadcast-"+step, err)
		return nil, err
	}

	broadcast, err := parseTxResult(stdout)
	if err != nil {
		err = fmt.Errorf("decode quorum CheckTx response for %s: %w", step, err)
		n.artifacts.recordFailure("quorum-decode-checktx-"+step, err)
		return nil, err
	}
	if broadcast.Code != 0 {
		err = fmt.Errorf(
			"quorum CheckTx %s failed: codespace=%s code=%d raw_log=%s",
			step,
			broadcast.Codespace,
			broadcast.Code,
			broadcast.RawLog,
		)
		n.artifacts.recordFailure("quorum-checktx-"+step, err)
		return nil, err
	}
	hash, err := hex.DecodeString(broadcast.TxHash)
	if err != nil {
		return nil, fmt.Errorf("decode quorum transaction hash %q: %w", broadcast.TxHash, err)
	}
	if len(hash) == 0 {
		return nil, errors.New("quorum transaction hash decoded to empty bytes")
	}
	pending := &QuorumPendingTx{
		RequestID: requestID,
		Step:      step,
		TxHash:    broadcast.TxHash,
		CheckTx:   broadcast,
	}

	samples, operationErr := n.observeQuorumTxNotCommitted(ctx, observationWindow, hash)
	recordErr := n.recordQuorumEvidence(step, "transaction-not-committed", map[string]any{
		"pending":            pending,
		"observation_window": observationWindow.String(),
		"samples":            samples,
	}, operationErr)
	return pending, errors.Join(operationErr, recordErr)
}

func (n *Network) observeQuorumTxNotCommitted(
	ctx context.Context,
	observationWindow time.Duration,
	hash []byte,
) ([]QuorumTxLookupSample, error) {
	fullNode := n.Chain.FullNodes[0]
	windowCtx, cancel := context.WithTimeout(ctx, observationWindow)
	defer cancel()
	ticker := time.NewTicker(quorumPollInterval)
	defer ticker.Stop()

	samples := make([]QuorumTxLookupSample, 0)
	for {
		sample := QuorumTxLookupSample{}
		status, statusErr := fullNode.Client.Status(windowCtx)
		if statusErr == nil && status != nil {
			sample.FullNodeHeight = status.SyncInfo.LatestBlockHeight
			sample.CatchingUp = status.SyncInfo.CatchingUp
			result, lookupErr := fullNode.Client.Tx(windowCtx, hash, false)
			sample.LookupError = quorumErrorString(lookupErr)
			if lookupErr == nil {
				if result == nil {
					return samples, errors.New("full-node tx lookup returned an empty successful result")
				}
				sample.Committed = true
				sample.CommitHeight = result.Height
				sample.ObservedAt = time.Now().UTC()
				samples = append(samples, sample)
				return samples, fmt.Errorf(
					"transaction %X committed unexpectedly at height %d during quorum-loss window",
					hash,
					result.Height,
				)
			}
		} else {
			sample.LookupError = quorumErrorString(statusErr)
			if statusErr == nil {
				sample.LookupError = "full-node status returned an empty result"
			}
		}
		sample.ObservedAt = time.Now().UTC()
		samples = append(samples, sample)

		select {
		case <-ctx.Done():
			return samples, fmt.Errorf("observe uncommitted transaction: %w", ctx.Err())
		case <-windowCtx.Done():
			if ctx.Err() != nil {
				return samples, fmt.Errorf("observe uncommitted transaction: %w", ctx.Err())
			}
			if err := VerifyQuorumNoCommitEvidence(
				samples,
				time.Now().UTC(),
				2*quorumPollInterval,
			); err != nil {
				return samples, err
			}
			return samples, nil
		case <-ticker.C:
		}
	}
}

// WaitForQuorumTxCommit proves that the exact CheckTx-accepted transaction
// commits successfully after validator quorum is restored.
func (n *Network) WaitForQuorumTxCommit(ctx context.Context, pending *QuorumPendingTx) (*TxResult, error) {
	if pending == nil || strings.TrimSpace(pending.TxHash) == "" {
		return nil, errors.New("pending quorum transaction is required")
	}
	n.txMu.Lock()
	defer n.txMu.Unlock()

	committed, operationErr := n.waitForCommittedTx(ctx, pending.RequestID, pending.Step, pending.TxHash)
	if operationErr == nil && committed == nil {
		operationErr = errors.New("committed quorum transaction result is empty")
	}
	if operationErr == nil && committed.Code != 0 {
		operationErr = fmt.Errorf(
			"recovered quorum transaction %s failed: codespace=%s code=%d raw_log=%s",
			pending.Step,
			committed.Codespace,
			committed.Code,
			committed.RawLog,
		)
	}
	recordErr := n.recordQuorumEvidence(pending.Step, "transaction-committed-after-recovery", map[string]any{
		"pending":   pending,
		"committed": committed,
	}, operationErr)
	return committed, errors.Join(operationErr, recordErr)
}
