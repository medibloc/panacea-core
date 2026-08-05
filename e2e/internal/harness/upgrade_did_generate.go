package harness

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

// GenerateDIDAuthenticatedTx asks the running historical binary to build an
// unsigned Cosmos transaction for an existing DID key. The DID keystore
// password is read only inside the node and neither it nor private key material
// is returned or written to artifacts.
func (n *Network) GenerateDIDAuthenticatedTx(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	keyName string,
	command ...string,
) (json.RawMessage, error) {
	n.txMu.Lock()
	defer n.txMu.Unlock()

	if strings.TrimSpace(step) == "" {
		return nil, errors.New("DID generate-only step is required")
	}
	if node == nil {
		return nil, errors.New("DID generate-only node is required")
	}
	if strings.TrimSpace(keyName) == "" {
		return nil, errors.New("DID generate-only key name is required")
	}
	if len(command) == 0 {
		return nil, errors.New("DID generate-only command is required")
	}
	for _, argument := range command {
		if argument == "--generate-only" {
			return nil, errors.New("DID generate-only flag is managed by the harness")
		}
	}

	arguments := append(append([]string(nil), command...), "--generate-only")
	requestID := fmt.Sprintf("%s-%d", step, time.Now().UTC().UnixNano())
	if err := n.artifacts.appendJSONLine("tx/requests.jsonl", map[string]any{
		"request_id":  requestID,
		"recorded_at": time.Now().UTC(),
		"step":        step,
		"node":        node.Name(),
		"key_name":    keyName,
		"arguments":   arguments,
		"stdin_mode":  "node-local-disposable-password",
		"operation":   "generate-only",
	}); err != nil {
		return nil, fmt.Errorf("record DID generate-only request %s: %w", step, err)
	}

	txCommand := node.TxCommand(keyName, arguments...)
	shellCommand := legacyDIDAuthenticatedCommand(
		path.Join(node.HomeDir(), legacyDIDPasswordFileName),
		txCommand,
	)
	stdout, stderr, execErr := node.Exec(ctx, shellCommand, node.Chain.Config().Env)
	safeStdout := sanitizeLegacyDIDOutput(stdout)
	safeStderr := sanitizeLegacyDIDOutput(stderr)
	safeExecErr := sanitizeLegacyDIDError(execErr)
	validJSON := json.Valid(safeStdout)
	recordErr := n.artifacts.appendJSONLine("tx/generated-results.jsonl", map[string]any{
		"request_id":  requestID,
		"recorded_at": time.Now().UTC(),
		"step":        step,
		"stdout":      jsonOrString(safeStdout),
		"stderr":      boundedString(safeStderr, txStderrMaxBytes),
		"exec_error":  errorString(safeExecErr),
		"valid_json":  validJSON,
	})
	if recordErr != nil {
		return nil, fmt.Errorf("record DID generate-only result %s: %w", step, recordErr)
	}
	if execErr != nil {
		return nil, fmt.Errorf(
			"generate DID transaction %s: %w: %s",
			step,
			safeExecErr,
			boundedString(safeStderr, txStderrMaxBytes),
		)
	}
	if !validJSON {
		return nil, fmt.Errorf("DID generate-only step %s returned invalid JSON", step)
	}
	return append(json.RawMessage(nil), safeStdout...), nil
}
