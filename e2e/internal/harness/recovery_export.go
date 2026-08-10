package harness

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

// DeterministicGenesisExport is artifact evidence for two independent state
// exports taken from the same closed database version.
type DeterministicGenesisExport struct {
	RecordedAt time.Time `json:"recorded_at"`
	Step       string    `json:"step"`
	Node       string    `json:"node"`
	Height     int64     `json:"height"`
	Digest     string    `json:"digest"`
	Bytes      int       `json:"bytes"`
	// Contents is retained in memory for an isolated bootstrap proof. It is
	// intentionally excluded from compact evidence; the public export itself
	// is written separately as recovery/exported-genesis.json.
	Contents []byte `json:"-"`
}

// CanonicalGenesisDigest returns a stable digest for semantically identical
// genesis JSON documents. It preserves JSON number spelling while removing
// formatting and object-key-order differences.
func CanonicalGenesisDigest(contents []byte) (string, error) {
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return "", fmt.Errorf("decode genesis JSON: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return "", fmt.Errorf("decode genesis JSON: multiple JSON values")
		}
		return "", fmt.Errorf("decode trailing genesis JSON: %w", err)
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("encode canonical genesis JSON: %w", err)
	}
	digest := sha256.Sum256(canonical)
	return fmt.Sprintf("%X", digest[:]), nil
}

// ExportValidatorGenesisDeterministically gracefully closes one validator DB,
// exports the same height twice, and requires equal canonical genesis digests.
// The full first export and compact evidence are preserved as run artifacts.
func (n *Network) ExportValidatorGenesisDeterministically(
	ctx context.Context,
	step string,
	validatorIndex int,
	height int64,
) (DeterministicGenesisExport, error) {
	if validatorIndex < 0 || validatorIndex >= len(n.Chain.Validators) {
		return DeterministicGenesisExport{}, fmt.Errorf("validator index %d is out of range", validatorIndex)
	}
	if height <= 0 {
		return DeterministicGenesisExport{}, fmt.Errorf("export height must be positive, got %d", height)
	}
	node := n.Chain.Validators[validatorIndex]
	n.txMu.Lock()
	defer n.txMu.Unlock()

	var evidence DeterministicGenesisExport
	err := runRecoveryStoppedOperation(
		func() error { return n.stopNodeForRecovery(ctx, step, node) },
		func() error {
			var err error
			evidence, err = n.exportGenesisDeterministically(
				ctx,
				step,
				node,
				height,
				"exported-genesis",
				"export-evidence.json",
			)
			return err
		},
		func() error { return n.startNodeAfterRecovery(ctx, step, node) },
	)
	if err != nil {
		return evidence, fmt.Errorf("deterministic genesis export on %s: %w", node.Name(), err)
	}
	return evidence, nil
}

// StopAndExportValidatorGenesisDeterministically stops every source process,
// exports the validator's latest closed DB version twice, and deliberately
// leaves the source stopped. This is the safe export for StartFromExport: the
// copied consensus key has never signed at or above the exported
// initial_height.
func (n *Network) StopAndExportValidatorGenesisDeterministically(
	ctx context.Context,
	step string,
	validatorIndex int,
) (DeterministicGenesisExport, error) {
	if validatorIndex < 0 || validatorIndex >= len(n.Chain.Validators) {
		return DeterministicGenesisExport{}, fmt.Errorf("validator index %d is out of range", validatorIndex)
	}
	node := n.Chain.Validators[validatorIndex]
	n.txMu.Lock()
	defer n.txMu.Unlock()

	if _, err := n.runRecoveryAction(step, "stop-source-chain", node, func() ([]byte, []byte, error) {
		var stopErrors []error
		stoppedNodeNames := make([]string, 0, len(n.Chain.Nodes()))
		for _, sourceNode := range n.Chain.Nodes() {
			if err := sourceNode.StopContainer(ctx); err != nil {
				stopErrors = append(stopErrors, fmt.Errorf("stop %s: %w", sourceNode.Name(), err))
				continue
			}
			stoppedNodeNames = append(stoppedNodeNames, sourceNode.Name())
		}
		n.artifacts.markIntentionallyStopped(stoppedNodeNames...)
		return nil, nil, errors.Join(stopErrors...)
	}); err != nil {
		return DeterministicGenesisExport{}, fmt.Errorf("stop source chain before terminal export: %w", err)
	}
	privateValidatorState, err := node.ReadFile(ctx, "data/priv_validator_state.json")
	if err != nil {
		return DeterministicGenesisExport{}, fmt.Errorf("read stopped source private-validator state: %w", err)
	}
	lastSignedHeight, err := privateValidatorLastSignedHeight(privateValidatorState)
	if err != nil {
		return DeterministicGenesisExport{}, err
	}
	evidence, err := n.exportGenesisDeterministically(
		ctx,
		step,
		node,
		-1,
		"import-exported-genesis",
		"import-export-evidence.json",
	)
	if err != nil {
		return evidence, fmt.Errorf("terminal deterministic genesis export on %s (source remains stopped): %w", node.Name(), err)
	}
	safeToReuse := lastSignedHeight <= evidence.Height
	artifactErr := n.artifacts.writeJSON("recovery/import-consensus-safety.json", map[string]any{
		"recorded_at":                 time.Now().UTC(),
		"step":                        step,
		"node":                        node.Name(),
		"last_signed_height":          lastSignedHeight,
		"exported_app_height":         evidence.Height,
		"exported_initial_height":     evidence.Height + 1,
		"safe_to_reuse_validator_key": safeToReuse,
	})
	if !safeToReuse {
		return evidence, errors.Join(
			fmt.Errorf(
				"source validator already signed height %d beyond exported app height %d; refusing consensus-key reuse",
				lastSignedHeight,
				evidence.Height,
			),
			artifactErr,
		)
	}
	if artifactErr != nil {
		return evidence, fmt.Errorf("record import consensus safety: %w", artifactErr)
	}
	return evidence, nil
}

func (n *Network) exportGenesisDeterministically(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	requestedHeight int64,
	artifactStem string,
	evidenceName string,
) (DeterministicGenesisExport, error) {
	export := func(action string) ([]byte, error) {
		var contents []byte
		_, err := n.runRecoveryAction(step, action, node, func() ([]byte, []byte, error) {
			exported, exportErr := node.ExportState(ctx, requestedHeight)
			contents = []byte(exported)
			return []byte(fmt.Sprintf("exported %d bytes", len(contents))), nil, exportErr
		})
		return contents, err
	}
	first, err := export("genesis-export-first")
	if err != nil {
		return DeterministicGenesisExport{}, fmt.Errorf("first genesis export at height %d: %w", requestedHeight, err)
	}
	second, err := export("genesis-export-second")
	if err != nil {
		return DeterministicGenesisExport{}, fmt.Errorf("second genesis export at height %d: %w", requestedHeight, err)
	}
	firstHeight, err := exportedGenesisAppHeight(first)
	if err != nil {
		return DeterministicGenesisExport{}, fmt.Errorf("first genesis export height: %w", err)
	}
	secondHeight, err := exportedGenesisAppHeight(second)
	if err != nil {
		return DeterministicGenesisExport{}, fmt.Errorf("second genesis export height: %w", err)
	}
	if firstHeight != secondHeight {
		return DeterministicGenesisExport{}, fmt.Errorf("genesis export height changed: first=%d second=%d", firstHeight, secondHeight)
	}
	if requestedHeight > 0 && firstHeight != requestedHeight {
		return DeterministicGenesisExport{}, fmt.Errorf("genesis export returned app height %d, want %d", firstHeight, requestedHeight)
	}
	firstDigest, err := CanonicalGenesisDigest(first)
	if err != nil {
		return DeterministicGenesisExport{}, fmt.Errorf("first genesis export: %w", err)
	}
	secondDigest, err := CanonicalGenesisDigest(second)
	if err != nil {
		return DeterministicGenesisExport{}, fmt.Errorf("second genesis export: %w", err)
	}
	if firstDigest != secondDigest {
		_ = n.artifacts.write("recovery/"+artifactStem+"-first.json", first)
		_ = n.artifacts.write("recovery/"+artifactStem+"-second.json", second)
		return DeterministicGenesisExport{}, fmt.Errorf(
			"canonical genesis export changed at height %d: first=%s second=%s",
			firstHeight,
			firstDigest,
			secondDigest,
		)
	}
	evidence := DeterministicGenesisExport{
		RecordedAt: time.Now().UTC(),
		Step:       step,
		Node:       node.Name(),
		Height:     firstHeight,
		Digest:     firstDigest,
		Bytes:      len(first),
		Contents:   append([]byte(nil), first...),
	}
	if err := n.artifacts.write("recovery/"+artifactStem+".json", first); err != nil {
		return evidence, fmt.Errorf("record exported genesis: %w", err)
	}
	if err := n.artifacts.writeJSON("recovery/"+evidenceName, evidence); err != nil {
		return evidence, fmt.Errorf("record genesis export evidence: %w", err)
	}
	return evidence, nil
}

func exportedGenesisAppHeight(contents []byte) (int64, error) {
	document, err := decodeJSONObject(contents, "exported genesis")
	if err != nil {
		return 0, err
	}
	initialHeight, err := decodeJSONInt64(document["initial_height"], "exported genesis initial_height")
	if err != nil {
		return 0, err
	}
	if initialHeight <= 1 {
		return 0, fmt.Errorf("exported genesis initial_height must be greater than 1, got %d", initialHeight)
	}
	return initialHeight - 1, nil
}

func privateValidatorLastSignedHeight(contents []byte) (int64, error) {
	document, err := decodeJSONObject(contents, "private-validator state")
	if err != nil {
		return 0, err
	}
	height, err := decodeJSONInt64(document["height"], "private-validator state height")
	if err != nil {
		return 0, err
	}
	if height < 0 {
		return 0, fmt.Errorf("private-validator state height must not be negative, got %d", height)
	}
	return height, nil
}

func decodeJSONInt64(raw json.RawMessage, label string) (int64, error) {
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		height, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%s is invalid: %q", label, text)
		}
		return height, nil
	}
	var height int64
	if err := json.Unmarshal(raw, &height); err != nil {
		return 0, fmt.Errorf("%s is invalid: %s", label, bytes.TrimSpace(raw))
	}
	return height, nil
}
