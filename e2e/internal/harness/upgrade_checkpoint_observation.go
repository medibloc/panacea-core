package harness

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"google.golang.org/grpc/metadata"
)

const UpgradeCheckpointQueryBoundaryCometBFTRPC = "cometbft-rpc"

const upgradeCheckpointApplicationPollInterval = 100 * time.Millisecond

// UpgradeCheckpointObservation binds a semantic state checkpoint to one
// concrete committed application height observed through one named node.
// Domain queries may use CLI, gRPC, or REST; the block ID at H and application
// hash after H are independently read from CometBFT RPC so neither can be
// inferred from a client-side query flag. For a historical H, block H+1 carries
// the application hash committed after H.
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
// application height from node. Passing height zero resolves the node's latest
// height. CometBFT RPC can briefly advertise a new height before BaseApp serves
// that height, so the method probes the node's pinned gRPC state before callers
// issue their semantic checkpoint queries.
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
	if node.GrpcConn == nil {
		return UpgradeCheckpointObservation{}, fmt.Errorf("capture upgrade checkpoint observation %s on %s: gRPC connection is nil", step, node.Name())
	}
	observation, err := captureUpgradeCheckpointObservationRPC(ctx, node.Name(), node.Client, height)
	if err != nil {
		return UpgradeCheckpointObservation{}, fmt.Errorf("capture upgrade checkpoint observation %s on %s: %w", step, node.Name(), err)
	}
	probe := func(probeCtx context.Context) error {
		pinnedCtx := metadata.AppendToOutgoingContext(
			probeCtx,
			"x-cosmos-block-height",
			fmt.Sprintf("%d", observation.Height),
		)
		_, probeErr := banktypes.NewQueryClient(node.GrpcConn).Params(pinnedCtx, &banktypes.QueryParamsRequest{})
		return probeErr
	}
	if err := waitForUpgradeApplicationHeight(
		ctx,
		observation.Height,
		upgradeCheckpointApplicationPollInterval,
		probe,
	); err != nil {
		return UpgradeCheckpointObservation{}, fmt.Errorf("capture upgrade checkpoint observation %s on %s: %w", step, node.Name(), err)
	}
	observation.ObservedAt = time.Now().UTC()
	if err := observation.Validate(); err != nil {
		return UpgradeCheckpointObservation{}, err
	}
	return observation, nil
}

func captureUpgradeCheckpointObservationRPC(
	ctx context.Context,
	nodeName string,
	client recoveryBlockClient,
	height int64,
) (UpgradeCheckpointObservation, error) {
	if height < 0 {
		return UpgradeCheckpointObservation{}, fmt.Errorf("height must not be negative: %d", height)
	}
	status, err := client.Status(ctx)
	if err != nil {
		return UpgradeCheckpointObservation{}, fmt.Errorf("query status: %w", err)
	}
	if status == nil {
		return UpgradeCheckpointObservation{}, errors.New("query status returned an empty result")
	}
	latestHeight := status.SyncInfo.LatestBlockHeight
	if height == 0 {
		height = latestHeight
	}
	if height <= 0 {
		return UpgradeCheckpointObservation{}, fmt.Errorf("latest committed height is invalid: %d", height)
	}
	if latestHeight < height {
		return UpgradeCheckpointObservation{}, fmt.Errorf(
			"requested height %d is ahead of node height %d",
			height,
			latestHeight,
		)
	}
	block, err := client.Block(ctx, &height)
	if err != nil {
		return UpgradeCheckpointObservation{}, fmt.Errorf("query block %d: %w", height, err)
	}
	if block == nil || block.Block == nil {
		return UpgradeCheckpointObservation{}, fmt.Errorf("query block %d returned an empty result", height)
	}
	if block.Block.Header.Height != height {
		return UpgradeCheckpointObservation{}, fmt.Errorf("query block %d returned height %d", height, block.Block.Header.Height)
	}

	appHash := status.SyncInfo.LatestAppHash
	if latestHeight > height {
		carrierHeight := height + 1
		carrier, carrierErr := client.Block(ctx, &carrierHeight)
		if carrierErr != nil {
			return UpgradeCheckpointObservation{}, fmt.Errorf("query app-hash carrier block %d: %w", carrierHeight, carrierErr)
		}
		if carrier == nil || carrier.Block == nil {
			return UpgradeCheckpointObservation{}, fmt.Errorf("query app-hash carrier block %d returned an empty result", carrierHeight)
		}
		if carrier.Block.Header.Height != carrierHeight {
			return UpgradeCheckpointObservation{}, fmt.Errorf(
				"query app-hash carrier block %d returned height %d",
				carrierHeight,
				carrier.Block.Header.Height,
			)
		}
		appHash = carrier.Block.Header.AppHash
	}

	observation := UpgradeCheckpointObservation{
		ObservedAt:    time.Now().UTC(),
		Node:          nodeName,
		QueryBoundary: UpgradeCheckpointQueryBoundaryCometBFTRPC,
		Height:        height,
		BlockID:       fmt.Sprintf("%X", []byte(block.BlockID.Hash)),
		AppHash:       fmt.Sprintf("%X", []byte(appHash)),
	}
	if err := observation.Validate(); err != nil {
		return UpgradeCheckpointObservation{}, err
	}
	return observation, nil
}

func waitForUpgradeApplicationHeight(
	ctx context.Context,
	height int64,
	pollInterval time.Duration,
	probe func(context.Context) error,
) error {
	if height <= 0 {
		return fmt.Errorf("application height must be positive, got %d", height)
	}
	if pollInterval <= 0 {
		return fmt.Errorf("application height poll interval must be positive, got %s", pollInterval)
	}
	if probe == nil {
		return errors.New("application height probe is required")
	}
	for {
		err := probe(ctx)
		if err == nil {
			return nil
		}
		if !isFutureApplicationHeightError(err) {
			return fmt.Errorf("probe application height %d: %w", height, err)
		}
		timer := time.NewTimer(pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("wait for application height %d: %w", height, ctx.Err())
		case <-timer.C:
		}
	}
}

func isFutureApplicationHeightError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "invalid height") && strings.Contains(message, "height in the future")
}
