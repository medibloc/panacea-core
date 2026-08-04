package harness

import (
	"context"
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

var systemEvidenceStepPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// ZeroHeightInvariantEvidence proves that a disposable node process loaded a
// concrete historical application version and completed the application's
// zero-height export path. Panacea runs CrisisKeeper.AssertInvariants on that
// path before mutating the in-memory export state; no exported state is
// committed back to the source database.
type ZeroHeightInvariantEvidence struct {
	RecordedAt          time.Time `json:"recorded_at"`
	Step                string    `json:"step"`
	Node                string    `json:"node"`
	SourceHeight        int64     `json:"source_height"`
	OutputBytes         int       `json:"output_bytes"`
	CanonicalDigest     string    `json:"canonical_digest"`
	ForZeroHeight       bool      `json:"for_zero_height"`
	AllInvariantsPassed bool      `json:"all_invariants_passed"`
}

// GenesisValidationEvidence proves that the running target binary accepted an
// exported document through its public `genesis validate` command.
type GenesisValidationEvidence struct {
	RecordedAt      time.Time `json:"recorded_at"`
	Step            string    `json:"step"`
	Node            string    `json:"node"`
	Bytes           int       `json:"bytes"`
	CanonicalDigest string    `json:"canonical_digest"`
	Stdout          string    `json:"stdout,omitempty"`
}

func validateSystemEvidenceStep(step string) error {
	if !systemEvidenceStepPattern.MatchString(step) {
		return fmt.Errorf("system evidence step %q must match %s", step, systemEvidenceStepPattern)
	}
	return nil
}

// AssertValidatorInvariantsAtHeight gracefully stops one validator while the
// remaining quorum continues, executes a zero-height export against the
// requested closed application version, and restarts/catches up the validator.
func (n *Network) AssertValidatorInvariantsAtHeight(
	ctx context.Context,
	step string,
	validatorIndex int,
	height int64,
) (ZeroHeightInvariantEvidence, error) {
	if n == nil || n.Chain == nil {
		return ZeroHeightInvariantEvidence{}, errors.New("invariant export requires a network")
	}
	if err := validateSystemEvidenceStep(step); err != nil {
		return ZeroHeightInvariantEvidence{}, err
	}
	if validatorIndex < 0 || validatorIndex >= len(n.Chain.Validators) {
		return ZeroHeightInvariantEvidence{}, fmt.Errorf("validator index %d is out of range", validatorIndex)
	}
	if height <= 0 {
		return ZeroHeightInvariantEvidence{}, fmt.Errorf("invariant export height must be positive, got %d", height)
	}
	node := n.Chain.Validators[validatorIndex]
	beforeHeight, err := node.Height(ctx)
	if err != nil {
		return ZeroHeightInvariantEvidence{}, fmt.Errorf("query %s before invariant export: %w", node.Name(), err)
	}
	if beforeHeight < height {
		return ZeroHeightInvariantEvidence{}, fmt.Errorf(
			"validator %s height %d is behind invariant export height %d",
			node.Name(),
			beforeHeight,
			height,
		)
	}

	n.txMu.Lock()
	defer n.txMu.Unlock()
	relativeDocument := "system-invariant-" + step + ".json"
	absoluteDocument := path.Join(node.HomeDir(), relativeDocument)
	var evidence ZeroHeightInvariantEvidence
	operationErr := runRecoveryStoppedOperation(
		func() error { return n.stopNodeForRecovery(ctx, "system-invariant-"+step, node) },
		func() error {
			_, actionErr := n.runRecoveryAction(
				"system-invariant-"+step,
				"zero-height-invariant-export",
				node,
				func() ([]byte, []byte, error) {
					return node.ExecBin(
						ctx,
						"export",
						"--height", fmt.Sprintf("%d", height),
						"--for-zero-height",
						"--output-document", absoluteDocument,
						"--home", node.HomeDir(),
					)
				},
			)
			if actionErr != nil {
				return actionErr
			}
			contents, readErr := node.ReadFile(ctx, relativeDocument)
			if readErr != nil {
				return fmt.Errorf("read zero-height invariant export: %w", readErr)
			}
			digest, digestErr := CanonicalGenesisDigest(contents)
			if digestErr != nil {
				return fmt.Errorf("digest zero-height invariant export: %w", digestErr)
			}
			evidence = ZeroHeightInvariantEvidence{
				RecordedAt:          time.Now().UTC(),
				Step:                step,
				Node:                node.Name(),
				SourceHeight:        height,
				OutputBytes:         len(contents),
				CanonicalDigest:     digest,
				ForZeroHeight:       true,
				AllInvariantsPassed: true,
			}
			if err := n.artifacts.write(
				"upgrade/system-modules/invariants/"+step+"-zero-height-export.json",
				contents,
			); err != nil {
				return fmt.Errorf("record zero-height invariant export: %w", err)
			}
			if err := n.artifacts.writeJSON(
				"upgrade/system-modules/invariants/"+step+"-evidence.json",
				evidence,
			); err != nil {
				return fmt.Errorf("record zero-height invariant evidence: %w", err)
			}
			return nil
		},
		func() error { return n.startNodeAfterRecovery(ctx, "system-invariant-"+step, node) },
	)
	if operationErr != nil {
		return evidence, fmt.Errorf("assert invariants on %s at height %d: %w", node.Name(), height, operationErr)
	}
	if err := n.WaitForNodeHeight(ctx, node, beforeHeight+1); err != nil {
		return evidence, fmt.Errorf("wait for %s after invariant export: %w", node.Name(), err)
	}
	return evidence, nil
}

// ValidateGenesisDocument writes an exported document to a disposable node
// volume and validates it with the target binary's public genesis command.
func (n *Network) ValidateGenesisDocument(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	contents []byte,
) (GenesisValidationEvidence, error) {
	if n == nil || n.Chain == nil {
		return GenesisValidationEvidence{}, errors.New("genesis validation requires a network")
	}
	if err := validateSystemEvidenceStep(step); err != nil {
		return GenesisValidationEvidence{}, err
	}
	if node == nil {
		return GenesisValidationEvidence{}, errors.New("genesis validation node is required")
	}
	if len(strings.TrimSpace(string(contents))) == 0 {
		return GenesisValidationEvidence{}, errors.New("genesis validation document is empty")
	}
	digest, err := CanonicalGenesisDigest(contents)
	if err != nil {
		return GenesisValidationEvidence{}, fmt.Errorf("canonical genesis validation document: %w", err)
	}
	relativeDocument := "system-genesis-validation-" + step + ".json"
	if err := node.WriteFile(ctx, contents, relativeDocument); err != nil {
		return GenesisValidationEvidence{}, fmt.Errorf("stage genesis validation document: %w", err)
	}
	absoluteDocument := path.Join(node.HomeDir(), relativeDocument)

	n.txMu.Lock()
	defer n.txMu.Unlock()
	stdout, actionErr := n.runRecoveryAction(
		"system-genesis-validation-"+step,
		"genesis-validate",
		node,
		func() ([]byte, []byte, error) {
			return node.ExecBin(ctx, "genesis", "validate-genesis", absoluteDocument, "--home", node.HomeDir())
		},
	)
	if actionErr != nil {
		return GenesisValidationEvidence{}, fmt.Errorf("validate exported genesis with %s: %w", node.Name(), actionErr)
	}
	evidence := GenesisValidationEvidence{
		RecordedAt:      time.Now().UTC(),
		Step:            step,
		Node:            node.Name(),
		Bytes:           len(contents),
		CanonicalDigest: digest,
		Stdout:          strings.TrimSpace(string(stdout)),
	}
	if err := n.artifacts.writeJSON(
		"upgrade/system-modules/exports/"+step+"-validation.json",
		evidence,
	); err != nil {
		return evidence, fmt.Errorf("record genesis validation evidence: %w", err)
	}
	return evidence, nil
}
