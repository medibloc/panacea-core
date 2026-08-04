package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

const (
	txCommitPollInterval = 500 * time.Millisecond
	txStderrMaxBytes     = 64 << 10
)

// TxResult is the stable subset of the SDK broadcast/query-tx JSON contract
// used by the real-node suites. Raw retains the complete response as evidence.
type TxResult struct {
	Height    string          `json:"height"`
	TxHash    string          `json:"txhash"`
	Codespace string          `json:"codespace"`
	Code      uint32          `json:"code"`
	RawLog    string          `json:"raw_log"`
	Events    []TxEvent       `json:"events"`
	Raw       json.RawMessage `json:"-"`
}

// TxLifecycleResult keeps the sync-broadcast CheckTx response distinct from
// the later committed FinalizeBlock result. Committed remains nil until the
// transaction is observed through query-tx.
type TxLifecycleResult struct {
	CheckTx   *TxResult
	Committed *TxResult
}

// TxBatchRequest describes one independently signed transaction in a batch.
// A signer (node plus key name) may appear only once so concurrent broadcasts
// cannot race account sequence allocation.
type TxBatchRequest struct {
	Step    string
	Node    *cosmos.ChainNode
	KeyName string
	Command []string
}

type txBatchBroadcastResult struct {
	requestID string
	lifecycle *TxLifecycleResult
	err       error
}

type txBatchCommitResult struct {
	committed *TxResult
	err       error
}

// Result preserves the historical BroadcastAndWaitTx return semantics:
// committed evidence wins, CheckTx rejection returns its broadcast evidence,
// and an accepted transaction that was not observed committed returns nil.
func (r *TxLifecycleResult) Result() *TxResult {
	if r == nil {
		return nil
	}
	if r.Committed != nil {
		return r.Committed
	}
	if r.CheckTx != nil && r.CheckTx.Code != 0 {
		return r.CheckTx
	}
	return nil
}

// TxEvent is one typed or legacy ABCI event returned by query tx.
type TxEvent struct {
	Type       string             `json:"type"`
	Attributes []TxEventAttribute `json:"attributes"`
}

// TxEventAttribute is one event key/value pair. Typed-event string values are
// JSON quoted by the SDK, while some legacy events return plain strings.
type TxEventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Index bool   `json:"index,omitempty"`
}

// HeightInt64 parses a committed height. It returns zero for broadcast
// responses and malformed or absent heights.
func (r TxResult) HeightInt64() int64 {
	height, _ := strconv.ParseInt(r.Height, 10, 64)
	return height
}

// FindEvent returns the first exact event-type match.
func (r TxResult) FindEvent(eventType string) (TxEvent, bool) {
	for _, event := range r.Events {
		if event.Type == eventType {
			return event, true
		}
	}
	return TxEvent{}, false
}

// Attribute returns a normalized string attribute or an empty string when the
// key is absent. JSON-quoted typed-event values are decoded once.
func (e TxEvent) Attribute(key string) string {
	for _, attribute := range e.Attributes {
		if attribute.Key != key {
			continue
		}
		var decoded string
		if json.Unmarshal([]byte(attribute.Value), &decoded) == nil {
			return decoded
		}
		return attribute.Value
	}
	return ""
}

// BroadcastAndWaitTx records a safe logical request, preserves the sync
// broadcast response as CheckTx evidence, polls query-tx through the explicit
// full node, and verifies the committed execution result.
func (n *Network) BroadcastAndWaitTx(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	keyName string,
	command ...string,
) (*TxResult, error) {
	lifecycle, err := n.BroadcastAndWaitTxLifecycle(ctx, step, node, keyName, command...)
	return lifecycle.Result(), err
}

// BroadcastAndWaitTxBatch broadcasts independently signed transactions at the
// same time before waiting for any commit. This is required for short voting
// periods where serial BroadcastAndWaitTx calls can consume the entire window.
// Results retain request order and every request keeps the normal CheckTx and
// committed FinalizeBlock artifact lifecycle.
func (n *Network) BroadcastAndWaitTxBatch(
	ctx context.Context,
	requests []TxBatchRequest,
) ([]*TxResult, error) {
	n.txMu.Lock()
	defer n.txMu.Unlock()

	if err := validateTxBatchRequests(requests); err != nil {
		return nil, err
	}

	broadcasts := runConcurrentTxBatch(len(requests), func(index int) txBatchBroadcastResult {
		request := requests[index]
		requestID, lifecycle, err := n.broadcastTxCheck(
			ctx,
			request.Step,
			request.Node,
			request.KeyName,
			request.Command...,
		)
		return txBatchBroadcastResult{requestID: requestID, lifecycle: lifecycle, err: err}
	})
	commits := runConcurrentTxBatch(len(requests), func(index int) txBatchCommitResult {
		broadcast := broadcasts[index]
		if broadcast.err != nil || broadcast.lifecycle == nil || broadcast.lifecycle.CheckTx == nil {
			return txBatchCommitResult{}
		}
		committed, err := n.waitForCommittedTx(
			ctx,
			broadcast.requestID,
			requests[index].Step,
			broadcast.lifecycle.CheckTx.TxHash,
		)
		return txBatchCommitResult{committed: committed, err: err}
	})

	results := make([]*TxResult, len(requests))
	errs := make([]error, 0)
	for index, request := range requests {
		lifecycle := broadcasts[index].lifecycle
		if lifecycle != nil {
			lifecycle.Committed = commits[index].committed
			results[index] = lifecycle.Result()
		}
		if broadcasts[index].err != nil {
			n.artifacts.recordFailure("transaction-"+request.Step, broadcasts[index].err)
			errs = append(errs, broadcasts[index].err)
			continue
		}
		if commits[index].err != nil {
			n.artifacts.recordFailure("transaction-"+request.Step, commits[index].err)
			errs = append(errs, commits[index].err)
			continue
		}
		if results[index] == nil {
			err := fmt.Errorf("transaction %s has no committed result", request.Step)
			n.artifacts.recordFailure("committed-result-"+request.Step, err)
			errs = append(errs, err)
			continue
		}
		if results[index].Code != 0 {
			err := fmt.Errorf(
				"FinalizeBlock %s failed: codespace=%s code=%d raw_log=%s",
				request.Step,
				results[index].Codespace,
				results[index].Code,
				results[index].RawLog,
			)
			n.artifacts.recordFailure("finalize-block-"+request.Step, err)
			errs = append(errs, err)
		}
	}
	return results, errors.Join(errs...)
}

// BroadcastAndWaitTxLifecycle returns both transaction lifecycle boundaries
// for callers that must report CheckTx and committed execution independently.
func (n *Network) BroadcastAndWaitTxLifecycle(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	keyName string,
	command ...string,
) (*TxLifecycleResult, error) {
	lifecycle, err := n.broadcastAndWaitTxLifecycle(ctx, step, node, keyName, command...)
	if err != nil {
		n.artifacts.recordFailure("transaction-"+step, err)
		return lifecycle, err
	}
	result := lifecycle.Committed
	if result == nil {
		err = fmt.Errorf("transaction %s has no committed result", step)
		n.artifacts.recordFailure("committed-result-"+step, err)
		return lifecycle, err
	}
	if result.Code != 0 {
		err = fmt.Errorf(
			"FinalizeBlock %s failed: codespace=%s code=%d raw_log=%s",
			step,
			result.Codespace,
			result.Code,
			result.RawLog,
		)
		n.artifacts.recordFailure("finalize-block-"+step, err)
		return lifecycle, err
	}
	return lifecycle, nil
}

// BroadcastTxExpectCLIRejection records a transaction attempt that must be
// rejected by command-side validation before CheckTx. It keeps expected CLI
// failures out of the artifact failure summary while preserving the raw
// request and bounded subprocess evidence.
func (n *Network) BroadcastTxExpectCLIRejection(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	keyName string,
	expectedDiagnostic string,
	command ...string,
) error {
	n.txMu.Lock()
	defer n.txMu.Unlock()

	if strings.TrimSpace(step) == "" {
		return errors.New("transaction step is required")
	}
	if node == nil {
		return errors.New("transaction node is required")
	}
	if strings.TrimSpace(keyName) == "" {
		return errors.New("transaction key name is required")
	}
	if strings.TrimSpace(expectedDiagnostic) == "" {
		return errors.New("expected CLI diagnostic is required")
	}
	if len(command) == 0 {
		return errors.New("transaction command is required")
	}

	requestID := fmt.Sprintf("%s-%d", step, time.Now().UTC().UnixNano())
	request := map[string]any{
		"request_id":       requestID,
		"recorded_at":      time.Now().UTC(),
		"step":             step,
		"node":             node.Name(),
		"key_name":         keyName,
		"arguments":        append([]string(nil), command...),
		"expected_outcome": "cli_rejection",
	}
	if err := n.artifacts.appendJSONLine("tx/requests.jsonl", request); err != nil {
		return fmt.Errorf("record expected CLI rejection request %s: %w", step, err)
	}

	stdout, stderr, execErr := node.Exec(ctx, node.TxCommand(keyName, command...), node.Chain.Config().Env)
	evidence := map[string]any{
		"request_id":          requestID,
		"recorded_at":         time.Now().UTC(),
		"step":                step,
		"stdout":              jsonOrString(stdout),
		"stderr":              boundedString(stderr, txStderrMaxBytes),
		"exec_error":          errorString(execErr),
		"expected_diagnostic": expectedDiagnostic,
	}
	if err := n.artifacts.appendJSONLine("tx/broadcast-results.jsonl", evidence); err != nil {
		return fmt.Errorf("record expected CLI rejection result %s: %w", step, err)
	}
	if err := classifyExpectedCLIRejection(execErr, stdout, stderr, expectedDiagnostic); err != nil {
		err = fmt.Errorf("CLI rejection %s: %w", step, err)
		n.artifacts.recordFailure("cli-rejection-"+step, err)
		return err
	}
	return nil
}

func classifyExpectedCLIRejection(execErr error, stdout, stderr []byte, expectedDiagnostic string) error {
	if execErr == nil {
		return errors.New("transaction unexpectedly succeeded before CheckTx")
	}
	if strings.TrimSpace(string(stdout)) != "" {
		return fmt.Errorf(
			"CLI rejection returned transaction output instead of failing before CheckTx: %s",
			boundedString(stdout, txStderrMaxBytes),
		)
	}
	if !strings.Contains(string(stderr), expectedDiagnostic) &&
		!strings.Contains(execErr.Error(), expectedDiagnostic) {
		return fmt.Errorf(
			"CLI rejection is missing expected diagnostic %q: %v: %s",
			expectedDiagnostic,
			execErr,
			boundedString(stderr, txStderrMaxBytes),
		)
	}
	return nil
}

// BroadcastTxExpectCheckTxFailure proves that command-side construction and
// signing succeeded but the transaction was rejected by CheckTx with the
// expected stable codespace and code. Expected rejections are preserved as
// broadcast evidence without being added to the artifact failure summary.
func (n *Network) BroadcastTxExpectCheckTxFailure(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	keyName string,
	expectedCodespace string,
	expectedCode uint32,
	command ...string,
) (*TxResult, error) {
	n.txMu.Lock()
	defer n.txMu.Unlock()

	if strings.TrimSpace(step) == "" {
		return nil, errors.New("transaction step is required")
	}
	if node == nil {
		return nil, errors.New("transaction node is required")
	}
	if strings.TrimSpace(keyName) == "" {
		return nil, errors.New("transaction key name is required")
	}
	if strings.TrimSpace(expectedCodespace) == "" {
		return nil, errors.New("expected CheckTx codespace is required")
	}
	if expectedCode == 0 {
		return nil, errors.New("expected CheckTx code must be nonzero")
	}
	if len(command) == 0 {
		return nil, errors.New("transaction command is required")
	}

	requestID := fmt.Sprintf("%s-%d", step, time.Now().UTC().UnixNano())
	request := map[string]any{
		"request_id":         requestID,
		"recorded_at":        time.Now().UTC(),
		"step":               step,
		"node":               node.Name(),
		"key_name":           keyName,
		"arguments":          append([]string(nil), command...),
		"expected_outcome":   "checktx_rejection",
		"expected_codespace": expectedCodespace,
		"expected_code":      expectedCode,
	}
	if err := n.artifacts.appendJSONLine("tx/requests.jsonl", request); err != nil {
		return nil, fmt.Errorf("record expected CheckTx rejection request %s: %w", step, err)
	}

	stdout, stderr, execErr := node.Exec(ctx, node.TxCommand(keyName, command...), node.Chain.Config().Env)
	evidence := map[string]any{
		"request_id":         requestID,
		"recorded_at":        time.Now().UTC(),
		"step":               step,
		"stdout":             jsonOrString(stdout),
		"stderr":             boundedString(stderr, txStderrMaxBytes),
		"exec_error":         errorString(execErr),
		"expected_outcome":   "checktx_rejection",
		"expected_codespace": expectedCodespace,
		"expected_code":      expectedCode,
	}
	if err := n.artifacts.appendJSONLine("tx/broadcast-results.jsonl", evidence); err != nil {
		return nil, fmt.Errorf("record expected CheckTx rejection result %s: %w", step, err)
	}
	if execErr != nil {
		err := fmt.Errorf(
			"broadcast transaction expected to reach CheckTx %s: %w: %s",
			step,
			execErr,
			boundedString(stderr, txStderrMaxBytes),
		)
		n.artifacts.recordFailure("expected-checktx-broadcast-"+step, err)
		return nil, err
	}

	result, err := parseTxResult(stdout)
	if err != nil {
		err = fmt.Errorf("decode expected CheckTx rejection for %s: %w", step, err)
		n.artifacts.recordFailure("expected-checktx-decode-"+step, err)
		return nil, err
	}
	if err := classifyExpectedCheckTxFailure(result, expectedCodespace, expectedCode); err != nil {
		err = fmt.Errorf("CheckTx rejection %s: %w", step, err)
		n.artifacts.recordFailure("expected-checktx-"+step, err)
		return &result, err
	}
	return &result, nil
}

func classifyExpectedCheckTxFailure(result TxResult, expectedCodespace string, expectedCode uint32) error {
	if result.Code == 0 {
		return errors.New("transaction unexpectedly passed CheckTx")
	}
	if strings.TrimSpace(result.Height) != "0" {
		if result.HeightInt64() > 0 {
			return fmt.Errorf("transaction reached committed height %s instead of stopping at CheckTx", result.Height)
		}
		return fmt.Errorf("CheckTx rejection returned invalid broadcast height %q", result.Height)
	}
	if result.Codespace != expectedCodespace || result.Code != expectedCode {
		return fmt.Errorf(
			"returned codespace=%s code=%d, want codespace=%s code=%d",
			result.Codespace,
			result.Code,
			expectedCodespace,
			expectedCode,
		)
	}
	return nil
}

// BroadcastAndWaitTxExpectDeliverFailure proves that CheckTx accepted a
// transaction but committed execution rejected it with the expected stable
// codespace and code. This is used for atomicity and permanent-ID tests.
func (n *Network) BroadcastAndWaitTxExpectDeliverFailure(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	keyName string,
	expectedCodespace string,
	expectedCode uint32,
	command ...string,
) (*TxResult, error) {
	lifecycle, err := n.broadcastAndWaitTxLifecycle(ctx, step, node, keyName, command...)
	result := lifecycle.Result()
	if err != nil {
		return result, err
	}
	if result.Code != expectedCode || result.Codespace != expectedCodespace {
		return result, fmt.Errorf(
			"FinalizeBlock %s returned codespace=%s code=%d, want codespace=%s code=%d",
			step,
			result.Codespace,
			result.Code,
			expectedCodespace,
			expectedCode,
		)
	}
	return result, nil
}

func (n *Network) broadcastAndWaitTxLifecycle(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	keyName string,
	command ...string,
) (*TxLifecycleResult, error) {
	n.txMu.Lock()
	defer n.txMu.Unlock()

	requestID, lifecycle, err := n.broadcastTxCheck(ctx, step, node, keyName, command...)
	if err != nil {
		return lifecycle, err
	}
	committed, err := n.waitForCommittedTx(ctx, requestID, step, lifecycle.CheckTx.TxHash)
	lifecycle.Committed = committed
	if err != nil {
		return lifecycle, err
	}
	return lifecycle, nil
}

func (n *Network) broadcastTxCheck(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	keyName string,
	command ...string,
) (string, *TxLifecycleResult, error) {
	if strings.TrimSpace(step) == "" {
		return "", nil, errors.New("transaction step is required")
	}
	if node == nil {
		return "", nil, errors.New("transaction node is required")
	}
	if strings.TrimSpace(keyName) == "" {
		return "", nil, errors.New("transaction key name is required")
	}
	if len(command) == 0 {
		return "", nil, errors.New("transaction command is required")
	}

	requestID := fmt.Sprintf("%s-%d", step, time.Now().UTC().UnixNano())
	request := map[string]any{
		"request_id":  requestID,
		"recorded_at": time.Now().UTC(),
		"step":        step,
		"node":        node.Name(),
		"key_name":    keyName,
		"arguments":   append([]string(nil), command...),
	}
	if err := n.artifacts.appendJSONLine("tx/requests.jsonl", request); err != nil {
		n.artifacts.recordFailure("record-tx-request", err)
		return requestID, nil, fmt.Errorf("record transaction request %s: %w", step, err)
	}

	stdout, stderr, execErr := node.Exec(ctx, node.TxCommand(keyName, command...), node.Chain.Config().Env)
	broadcastEvidence := map[string]any{
		"request_id":  requestID,
		"recorded_at": time.Now().UTC(),
		"step":        step,
		"stdout":      jsonOrString(stdout),
		"stderr":      boundedString(stderr, txStderrMaxBytes),
		"exec_error":  errorString(execErr),
	}
	if err := n.artifacts.appendJSONLine("tx/broadcast-results.jsonl", broadcastEvidence); err != nil {
		n.artifacts.recordFailure("record-tx-broadcast", err)
		return requestID, nil, fmt.Errorf("record transaction broadcast %s: %w", step, err)
	}
	if execErr != nil {
		err := fmt.Errorf("broadcast transaction %s: %w: %s", step, execErr, boundedString(stderr, txStderrMaxBytes))
		n.artifacts.recordFailure("broadcast-tx-"+step, err)
		return requestID, nil, err
	}

	broadcast, err := parseTxResult(stdout)
	if err != nil {
		err = fmt.Errorf("decode CheckTx response for %s: %w", step, err)
		n.artifacts.recordFailure("decode-checktx-"+step, err)
		return requestID, nil, err
	}
	lifecycle := &TxLifecycleResult{CheckTx: &broadcast}
	if broadcast.Code != 0 {
		err = fmt.Errorf(
			"CheckTx %s failed: codespace=%s code=%d raw_log=%s",
			step,
			broadcast.Codespace,
			broadcast.Code,
			broadcast.RawLog,
		)
		n.artifacts.recordFailure("checktx-"+step, err)
		return requestID, lifecycle, err
	}
	return requestID, lifecycle, nil
}

func validateTxBatchRequests(requests []TxBatchRequest) error {
	if len(requests) == 0 {
		return errors.New("transaction batch is empty")
	}
	seenSigners := make(map[string]int, len(requests))
	for index, request := range requests {
		if strings.TrimSpace(request.Step) == "" {
			return fmt.Errorf("transaction batch request %d has no step", index)
		}
		if request.Node == nil {
			return fmt.Errorf("transaction batch request %d has no node", index)
		}
		if strings.TrimSpace(request.KeyName) == "" {
			return fmt.Errorf("transaction batch request %d has no key name", index)
		}
		if len(request.Command) == 0 {
			return fmt.Errorf("transaction batch request %d has no command", index)
		}
		signer := request.Node.Name() + "\x00" + request.KeyName
		if first, duplicate := seenSigners[signer]; duplicate {
			return fmt.Errorf(
				"transaction batch requests %d and %d use the same signer %s/%s",
				first,
				index,
				request.Node.Name(),
				request.KeyName,
			)
		}
		seenSigners[signer] = index
	}
	return nil
}

func runConcurrentTxBatch[T any](count int, execute func(index int) T) []T {
	if count <= 0 {
		return nil
	}
	results := make([]T, count)
	var wait sync.WaitGroup
	wait.Add(count)
	for index := 0; index < count; index++ {
		index := index
		go func() {
			defer wait.Done()
			results[index] = execute(index)
		}()
	}
	wait.Wait()
	return results
}

func (n *Network) waitForCommittedTx(ctx context.Context, requestID, step, txHash string) (*TxResult, error) {
	if len(n.Chain.FullNodes) == 0 {
		return nil, errors.New("network has no full node for committed transaction query")
	}
	node := n.Chain.FullNodes[0]
	ticker := time.NewTicker(txCommitPollInterval)
	defer ticker.Stop()

	var lastErr error
	for {
		stdout, stderr, err := node.ExecQuery(ctx, "tx", txHash)
		if err == nil {
			result, decodeErr := parseTxResult(stdout)
			if decodeErr == nil {
				if !strings.EqualFold(result.TxHash, txHash) {
					decodeErr = fmt.Errorf("query tx returned hash %s, want %s", result.TxHash, txHash)
				} else if result.HeightInt64() <= 0 {
					decodeErr = fmt.Errorf("query tx returned non-committed height %q", result.Height)
				}
			}
			if decodeErr == nil {
				evidence := map[string]any{
					"request_id":  requestID,
					"recorded_at": time.Now().UTC(),
					"step":        step,
					"tx_hash":     txHash,
					"result":      json.RawMessage(result.Raw),
				}
				if artifactErr := n.artifacts.appendJSONLine("tx/committed-results.jsonl", evidence); artifactErr != nil {
					return nil, fmt.Errorf("record committed transaction %s: %w", step, artifactErr)
				}
				return &result, nil
			}
			lastErr = decodeErr
		} else {
			lastErr = fmt.Errorf("%w: %s", err, boundedString(stderr, txStderrMaxBytes))
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("wait for committed transaction %s (%s): last error=%v: %w", step, txHash, lastErr, ctx.Err())
		case <-ticker.C:
		}
	}
}

func parseTxResult(contents []byte) (TxResult, error) {
	var raw struct {
		Height    json.RawMessage `json:"height"`
		TxHash    string          `json:"txhash"`
		Codespace string          `json:"codespace"`
		Code      json.RawMessage `json:"code"`
		RawLog    string          `json:"raw_log"`
		Events    []TxEvent       `json:"events"`
		Logs      []struct {
			Events []TxEvent `json:"events"`
		} `json:"logs"`
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.UseNumber()
	if err := decoder.Decode(&raw); err != nil {
		return TxResult{}, fmt.Errorf("decode transaction JSON: %w", err)
	}
	if strings.TrimSpace(raw.TxHash) == "" {
		return TxResult{}, errors.New("transaction JSON has no txhash")
	}
	if len(raw.Code) == 0 || bytes.Equal(raw.Code, []byte("null")) {
		return TxResult{}, errors.New("transaction JSON has no code")
	}
	height, _ := rawJSONScalar(raw.Height)
	codeText, _ := rawJSONScalar(raw.Code)
	var code uint64
	if codeText != "" {
		parsed, err := strconv.ParseUint(codeText, 10, 32)
		if err != nil {
			return TxResult{}, fmt.Errorf("decode transaction code %q: %w", codeText, err)
		}
		code = parsed
	}
	events := raw.Events
	if len(events) == 0 {
		for _, log := range raw.Logs {
			events = append(events, log.Events...)
		}
	}
	return TxResult{
		Height:    height,
		TxHash:    raw.TxHash,
		Codespace: raw.Codespace,
		Code:      uint32(code),
		RawLog:    raw.RawLog,
		Events:    events,
		Raw:       append(json.RawMessage(nil), contents...),
	}, nil
}

func rawJSONScalar(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return "", nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}
	var number json.Number
	if err := json.Unmarshal(raw, &number); err == nil {
		return number.String(), nil
	}
	return "", fmt.Errorf("expected JSON string or number, got %s", raw)
}

func jsonOrString(contents []byte) any {
	trimmed := bytes.TrimSpace(contents)
	if json.Valid(trimmed) {
		return json.RawMessage(append([]byte(nil), trimmed...))
	}
	return boundedString(contents, txStderrMaxBytes)
}

func boundedString(contents []byte, maximum int) string {
	if len(contents) <= maximum {
		return string(contents)
	}
	return string(contents[:maximum]) + "\n[truncated]"
}
