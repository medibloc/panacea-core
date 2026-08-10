package harness

import (
	"context"
	"fmt"
	"time"

	rpcclient "github.com/cometbft/cometbft/rpc/client"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

// RecoveryCheckpoint identifies one committed block and its application state
// root. A restart is continuous when the historical checkpoint is unchanged
// and a later checkpoint can still be committed.
type RecoveryCheckpoint struct {
	RecordedAt time.Time `json:"recorded_at"`
	Step       string    `json:"step"`
	Node       string    `json:"node"`
	Height     int64     `json:"height"`
	BlockID    string    `json:"block_id"`
	AppHash    string    `json:"app_hash"`
}

type recoveryBlockClient interface {
	rpcclient.StatusClient
	Block(context.Context, *int64) (*coretypes.ResultBlock, error)
}

// CaptureRecoveryCheckpoint reads an exact committed block through the
// selected node's CometBFT RPC client and appends it to the run artifacts.
// Pass height zero to capture the node's latest committed height.
func (n *Network) CaptureRecoveryCheckpoint(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	height int64,
) (RecoveryCheckpoint, error) {
	if node == nil {
		return RecoveryCheckpoint{}, fmt.Errorf("capture recovery checkpoint %s: node is nil", step)
	}
	if node.Client == nil {
		return RecoveryCheckpoint{}, fmt.Errorf("capture recovery checkpoint %s on %s: RPC client is nil", step, node.Name())
	}
	checkpoint, err := captureRecoveryCheckpoint(ctx, step, node.Name(), node.Client, height)
	if err != nil {
		err = fmt.Errorf("capture recovery checkpoint %s on %s: %w", step, node.Name(), err)
		n.artifacts.recordFailure("recovery-checkpoint", err)
		return RecoveryCheckpoint{}, err
	}
	if err := n.artifacts.appendJSONLine("recovery/checkpoints.jsonl", checkpoint); err != nil {
		err = fmt.Errorf("record recovery checkpoint %s on %s: %w", step, node.Name(), err)
		n.artifacts.recordFailure("recovery-checkpoint-artifact", err)
		return RecoveryCheckpoint{}, err
	}
	return checkpoint, nil
}

func captureRecoveryCheckpoint(
	ctx context.Context,
	step string,
	nodeName string,
	client recoveryBlockClient,
	height int64,
) (RecoveryCheckpoint, error) {
	if height < 0 {
		return RecoveryCheckpoint{}, fmt.Errorf("height must not be negative: %d", height)
	}
	status, err := client.Status(ctx)
	if err != nil {
		return RecoveryCheckpoint{}, fmt.Errorf("query status: %w", err)
	}
	if status == nil {
		return RecoveryCheckpoint{}, fmt.Errorf("query status returned an empty result")
	}
	if height == 0 {
		height = status.SyncInfo.LatestBlockHeight
	}
	if height <= 0 {
		return RecoveryCheckpoint{}, fmt.Errorf("latest committed height is invalid: %d", height)
	}
	if status.SyncInfo.LatestBlockHeight < height {
		return RecoveryCheckpoint{}, fmt.Errorf(
			"requested height %d is ahead of node height %d",
			height,
			status.SyncInfo.LatestBlockHeight,
		)
	}
	block, err := client.Block(ctx, &height)
	if err != nil {
		return RecoveryCheckpoint{}, fmt.Errorf("query block %d: %w", height, err)
	}
	if block == nil || block.Block == nil {
		return RecoveryCheckpoint{}, fmt.Errorf("query block %d returned an empty result", height)
	}
	if block.Block.Header.Height != height {
		return RecoveryCheckpoint{}, fmt.Errorf("query block %d returned height %d", height, block.Block.Header.Height)
	}
	checkpoint := RecoveryCheckpoint{
		RecordedAt: time.Now().UTC(),
		Step:       step,
		Node:       nodeName,
		Height:     height,
		BlockID:    fmt.Sprintf("%X", []byte(block.BlockID.Hash)),
		AppHash:    fmt.Sprintf("%X", block.Block.Header.AppHash),
	}
	if checkpoint.BlockID == "" {
		return RecoveryCheckpoint{}, fmt.Errorf("block %d has an empty block ID", height)
	}
	if checkpoint.AppHash == "" {
		return RecoveryCheckpoint{}, fmt.Errorf("block %d has an empty app hash", height)
	}
	return checkpoint, nil
}

// ValidateRecoveryContinuity proves that a node preserved an already committed
// block across a restart and subsequently observed a newer committed block.
func ValidateRecoveryContinuity(before, afterSameHeight, progressed RecoveryCheckpoint) error {
	for _, observed := range []struct {
		label      string
		checkpoint RecoveryCheckpoint
	}{
		{label: "before", checkpoint: before},
		{label: "after same height", checkpoint: afterSameHeight},
		{label: "progressed", checkpoint: progressed},
	} {
		label := observed.label
		checkpoint := observed.checkpoint
		if checkpoint.Height <= 0 {
			return fmt.Errorf("%s checkpoint has invalid height %d", label, checkpoint.Height)
		}
		if checkpoint.BlockID == "" {
			return fmt.Errorf("%s checkpoint is missing block ID", label)
		}
		if checkpoint.AppHash == "" {
			return fmt.Errorf("%s checkpoint is missing app hash", label)
		}
	}
	if before.Height != afterSameHeight.Height {
		return fmt.Errorf("historical checkpoint height changed: before=%d after=%d", before.Height, afterSameHeight.Height)
	}
	if before.BlockID != afterSameHeight.BlockID {
		return fmt.Errorf("block ID changed at height %d: before=%s after=%s", before.Height, before.BlockID, afterSameHeight.BlockID)
	}
	if before.AppHash != afterSameHeight.AppHash {
		return fmt.Errorf("app hash changed at height %d: before=%s after=%s", before.Height, before.AppHash, afterSameHeight.AppHash)
	}
	if progressed.Height <= afterSameHeight.Height {
		return fmt.Errorf("height did not progress after restart: historical=%d latest=%d", afterSameHeight.Height, progressed.Height)
	}
	return nil
}
