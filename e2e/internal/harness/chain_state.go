package harness

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

const statePollInterval = 250 * time.Millisecond

// BlockEvidence identifies one committed history point on one concrete node.
// Matching app hashes alone are insufficient: BlockID also proves that nodes
// observed the same CometBFT history.
type BlockEvidence struct {
	RecordedAt  time.Time `json:"recorded_at"`
	Node        string    `json:"node"`
	Height      int64     `json:"height"`
	StateHeight int64     `json:"state_height"`
	BlockID     string    `json:"block_id"`
	AppHash     string    `json:"app_hash"`
}

// ValidatorPower is the stable validator-set projection retained by the
// consensus suite.
type ValidatorPower struct {
	Address string `json:"address"`
	Power   int64  `json:"power"`
}

// NodeBlock queries an exact height from the supplied node and records it.
func (n *Network) NodeBlock(ctx context.Context, node *cosmos.ChainNode, height int64) (BlockEvidence, error) {
	if node == nil {
		return BlockEvidence{}, errors.New("node is required")
	}
	if height <= 0 {
		return BlockEvidence{}, fmt.Errorf("block height must be positive, got %d", height)
	}
	result, err := node.Client.Block(ctx, &height)
	if err != nil {
		return BlockEvidence{}, fmt.Errorf("query block %d from %s: %w", height, node.Name(), err)
	}
	if result == nil || result.Block == nil {
		return BlockEvidence{}, fmt.Errorf("query block %d from %s returned no block", height, node.Name())
	}
	evidence := BlockEvidence{
		RecordedAt:  time.Now().UTC(),
		Node:        node.Name(),
		Height:      result.Block.Height,
		StateHeight: result.Block.Height - 1,
		BlockID:     strings.ToUpper(fmt.Sprintf("%X", []byte(result.BlockID.Hash))),
		AppHash:     strings.ToUpper(fmt.Sprintf("%X", []byte(result.Block.AppHash))),
	}
	if evidence.Height != height {
		return BlockEvidence{}, fmt.Errorf("node %s returned height %d, want %d", node.Name(), evidence.Height, height)
	}
	if evidence.BlockID == "" || evidence.AppHash == "" {
		return BlockEvidence{}, fmt.Errorf("node %s returned incomplete block evidence at %d", node.Name(), height)
	}
	if err := n.artifacts.appendJSONLine("nodes/history.jsonl", evidence); err != nil {
		return BlockEvidence{}, fmt.Errorf("record block evidence: %w", err)
	}
	return evidence, nil
}

// NodeLatestBlock resolves and records the latest block visible to one node.
func (n *Network) NodeLatestBlock(ctx context.Context, node *cosmos.ChainNode) (BlockEvidence, error) {
	if node == nil {
		return BlockEvidence{}, errors.New("node is required")
	}
	height, err := node.Height(ctx)
	if err != nil {
		return BlockEvidence{}, fmt.Errorf("query latest height from %s: %w", node.Name(), err)
	}
	return n.NodeBlock(ctx, node, height)
}

// RequireSameHistoryAtHeight proves block-ID and app-hash equality across all
// supplied nodes at one committed height.
func (n *Network) RequireSameHistoryAtHeight(
	ctx context.Context,
	height int64,
	nodes ...*cosmos.ChainNode,
) ([]BlockEvidence, error) {
	if len(nodes) < 2 {
		return nil, errors.New("at least two nodes are required for history comparison")
	}
	evidence := make([]BlockEvidence, 0, len(nodes))
	for _, node := range nodes {
		block, err := n.NodeBlock(ctx, node, height)
		if err != nil {
			return evidence, err
		}
		evidence = append(evidence, block)
	}
	if err := validateSameHistory(evidence); err != nil {
		return evidence, err
	}
	return evidence, nil
}

func validateSameHistory(evidence []BlockEvidence) error {
	if len(evidence) < 2 {
		return errors.New("at least two block observations are required")
	}
	want := evidence[0]
	for _, observed := range evidence[1:] {
		if observed.Height != want.Height {
			return fmt.Errorf("node %s height %d differs from %s height %d", observed.Node, observed.Height, want.Node, want.Height)
		}
		if !strings.EqualFold(observed.BlockID, want.BlockID) {
			return fmt.Errorf("node %s block ID %s differs from %s block ID %s at height %d", observed.Node, observed.BlockID, want.Node, want.BlockID, want.Height)
		}
		if !strings.EqualFold(observed.AppHash, want.AppHash) {
			return fmt.Errorf("node %s app hash %s differs from %s app hash %s at height %d", observed.Node, observed.AppHash, want.Node, want.AppHash, want.Height)
		}
	}
	return nil
}

// WaitForNodeHeight waits on one explicitly chosen node.
func (n *Network) WaitForNodeHeight(ctx context.Context, node *cosmos.ChainNode, target int64) error {
	if node == nil {
		return errors.New("node is required")
	}
	return waitForHeight(ctx, target, node.Height)
}

// WaitForStableHeight allows an in-flight block and then requires the node's
// height to remain unchanged for the full stability window.
func (n *Network) WaitForStableHeight(
	ctx context.Context,
	node *cosmos.ChainNode,
	stabilityWindow time.Duration,
) (int64, error) {
	if node == nil {
		return 0, errors.New("node is required")
	}
	if stabilityWindow <= 0 {
		return 0, errors.New("stability window must be positive")
	}
	lastHeight, err := node.Height(ctx)
	if err != nil {
		return 0, fmt.Errorf("read initial height from %s: %w", node.Name(), err)
	}
	stableSince := time.Now()
	ticker := time.NewTicker(statePollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("wait for %s height to remain stable for %s (last height %d): %w", node.Name(), stabilityWindow, lastHeight, ctx.Err())
		case <-ticker.C:
			height, queryErr := node.Height(ctx)
			if queryErr != nil {
				return 0, fmt.Errorf("read height from %s while checking halt: %w", node.Name(), queryErr)
			}
			if height != lastHeight {
				lastHeight = height
				stableSince = time.Now()
				continue
			}
			if time.Since(stableSince) >= stabilityWindow {
				return lastHeight, nil
			}
		}
	}
}

// ValidatorSet records the entire validator set at an exact height.
func (n *Network) ValidatorSet(
	ctx context.Context,
	node *cosmos.ChainNode,
	height int64,
) ([]ValidatorPower, error) {
	if node == nil {
		return nil, errors.New("node is required")
	}
	perPage := 100
	result, err := node.Client.Validators(ctx, &height, nil, &perPage)
	if err != nil {
		return nil, fmt.Errorf("query validator set at height %d from %s: %w", height, node.Name(), err)
	}
	if result == nil || result.Total != len(result.Validators) {
		return nil, fmt.Errorf("validator response at height %d is incomplete: total=%d returned=%d", height, result.Total, len(result.Validators))
	}
	powers := make([]ValidatorPower, 0, len(result.Validators))
	for _, validator := range result.Validators {
		if validator == nil {
			return nil, fmt.Errorf("validator response at height %d contains nil entry", height)
		}
		powers = append(powers, ValidatorPower{
			Address: strings.ToUpper(fmt.Sprintf("%X", []byte(validator.Address))),
			Power:   validator.VotingPower,
		})
	}
	if err := n.artifacts.appendJSONLine("nodes/validator-sets.jsonl", map[string]any{
		"recorded_at": time.Now().UTC(),
		"node":        node.Name(),
		"height":      height,
		"validators":  powers,
	}); err != nil {
		return nil, fmt.Errorf("record validator set: %w", err)
	}
	return powers, nil
}

// WriteArtifactJSON and AppendArtifactJSON expose the artifact sink to suites
// while retaining the store's path-containment checks.
func (n *Network) WriteArtifactJSON(relativePath string, value any) error {
	return n.artifacts.writeJSON(relativePath, value)
}

func (n *Network) AppendArtifactJSON(relativePath string, value any) error {
	return n.artifacts.appendJSONLine(relativePath, value)
}

func (n *Network) WriteArtifact(relativePath string, contents []byte) error {
	return n.artifacts.write(relativePath, contents)
}
