package harness

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

const UpgradeCheckpointQueryBoundaryCometBFTRPC = "cometbft-rpc"

// UpgradeCheckpointObservation binds a semantic state checkpoint to one
// concrete committed block observed through one named node. Domain queries
// may use CLI, gRPC, or REST; the block identity itself is independently read
// from CometBFT RPC so height, block ID, and app hash cannot be inferred from
// a client-side query flag.
type UpgradeCheckpointObservation struct {
	ObservedAt    time.Time `json:"observed_at"`
	Node          string    `json:"node"`
	QueryBoundary string    `json:"query_boundary"`
	Height        int64     `json:"height"`
	BlockID       string    `json:"block_id"`
	AppHash       string    `json:"app_hash"`
}

func (o UpgradeCheckpointObservation) Validate() error {
	if o.ObservedAt.IsZero() {
		return errors.New("checkpoint observation observed_at is required")
	}
	if strings.TrimSpace(o.Node) == "" {
		return errors.New("checkpoint observation node is required")
	}
	if o.QueryBoundary != UpgradeCheckpointQueryBoundaryCometBFTRPC {
		return fmt.Errorf("checkpoint observation query_boundary %q, want %q", o.QueryBoundary, UpgradeCheckpointQueryBoundaryCometBFTRPC)
	}
	if o.Height <= 0 {
		return fmt.Errorf("checkpoint observation height must be positive, got %d", o.Height)
	}
	if strings.TrimSpace(o.BlockID) == "" || strings.TrimSpace(o.AppHash) == "" {
		return errors.New("checkpoint observation block_id and app_hash are required")
	}
	return nil
}

// CaptureUpgradeCheckpointObservation reads the exact requested committed
// block from node. Passing height zero resolves the node's latest height.
func (n *Network) CaptureUpgradeCheckpointObservation(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	height int64,
) (UpgradeCheckpointObservation, error) {
	if node == nil {
		return UpgradeCheckpointObservation{}, fmt.Errorf("capture upgrade checkpoint observation %s: node is nil", step)
	}
	if node.Client == nil {
		return UpgradeCheckpointObservation{}, fmt.Errorf("capture upgrade checkpoint observation %s on %s: RPC client is nil", step, node.Name())
	}
	checkpoint, err := captureRecoveryCheckpoint(ctx, step, node.Name(), node.Client, height)
	if err != nil {
		return UpgradeCheckpointObservation{}, fmt.Errorf("capture upgrade checkpoint observation %s on %s: %w", step, node.Name(), err)
	}
	observation := UpgradeCheckpointObservation{
		ObservedAt:    checkpoint.RecordedAt,
		Node:          checkpoint.Node,
		QueryBoundary: UpgradeCheckpointQueryBoundaryCometBFTRPC,
		Height:        checkpoint.Height,
		BlockID:       checkpoint.BlockID,
		AppHash:       checkpoint.AppHash,
	}
	if err := observation.Validate(); err != nil {
		return UpgradeCheckpointObservation{}, err
	}
	return observation, nil
}
