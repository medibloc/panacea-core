package harness

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

// NodeCLIQuery records and executes a JSON query against an explicitly chosen
// node. Upgrade-halt tests use this boundary while the historical full node is
// deliberately kept on the old image as a mempool carrier.
func (n *Network) NodeCLIQuery(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	command ...string,
) (json.RawMessage, error) {
	if strings.TrimSpace(step) == "" {
		return nil, errors.New("node query step is required")
	}
	if node == nil {
		return nil, errors.New("node query target is required")
	}
	if len(command) == 0 {
		return nil, errors.New("node query command is required")
	}

	stdout, stderr, queryErr := node.ExecQuery(ctx, command...)
	recordErr := n.recordQuery(queryRecord{
		Boundary: "node-cli",
		Step:     step,
		Request: map[string]any{
			"node":      node.Name(),
			"arguments": append([]string(nil), command...),
		},
		Response: jsonOrString(stdout),
		Stderr:   boundedString(stderr, txStderrMaxBytes),
		Error:    errorString(queryErr),
	})
	if recordErr != nil || queryErr != nil {
		return nil, errors.Join(
			recordErr,
			func() error {
				if queryErr == nil {
					return nil
				}
				return fmt.Errorf(
					"node CLI query %s on %s: %w: %s",
					step,
					node.Name(),
					queryErr,
					boundedString(stderr, txStderrMaxBytes),
				)
			}(),
		)
	}
	trimmed := strings.TrimSpace(string(stdout))
	if !json.Valid([]byte(trimmed)) {
		return nil, fmt.Errorf("node CLI query %s on %s returned invalid JSON", step, node.Name())
	}
	return append(json.RawMessage(nil), trimmed...), nil
}

// WaitForCommittedTxOnNode polls an explicitly chosen upgraded node for one
// exact hash. It does not depend on the full node being upgraded or available.
func (n *Network) WaitForCommittedTxOnNode(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	txHash string,
) (*TxResult, error) {
	if strings.TrimSpace(step) == "" {
		return nil, errors.New("committed transaction query step is required")
	}
	if node == nil {
		return nil, errors.New("committed transaction query node is required")
	}
	if strings.TrimSpace(txHash) == "" {
		return nil, errors.New("committed transaction hash is required")
	}

	ticker := time.NewTicker(txCommitPollInterval)
	defer ticker.Stop()
	var lastErr error
	for attempt := 1; ; attempt++ {
		stdout, stderr, queryErr := node.ExecQuery(ctx, "tx", txHash)
		if queryErr == nil {
			result, decodeErr := parseTxResult(stdout)
			if decodeErr == nil && !strings.EqualFold(result.TxHash, txHash) {
				decodeErr = fmt.Errorf("query tx returned hash %s, want %s", result.TxHash, txHash)
			}
			if decodeErr == nil && result.HeightInt64() <= 0 {
				decodeErr = fmt.Errorf("query tx returned non-committed height %q", result.Height)
			}
			if decodeErr == nil {
				recordErr := n.artifacts.appendJSONLine("tx/committed-results.jsonl", map[string]any{
					"recorded_at": time.Now().UTC(),
					"step":        step,
					"node":        node.Name(),
					"tx_hash":     txHash,
					"attempt":     attempt,
					"result":      json.RawMessage(result.Raw),
				})
				if recordErr != nil {
					return nil, fmt.Errorf("record committed transaction %s: %w", step, recordErr)
				}
				return &result, nil
			}
			lastErr = decodeErr
		} else {
			lastErr = fmt.Errorf("%w: %s", queryErr, boundedString(stderr, txStderrMaxBytes))
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf(
				"wait for committed transaction %s (%s) on %s: last error=%v: %w",
				step,
				txHash,
				node.Name(),
				lastErr,
				ctx.Err(),
			)
		case <-ticker.C:
		}
	}
}
