package harness

import (
	"context"
	"errors"
	"testing"
	"time"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"
)

type upgradeCheckpointRPCStub struct {
	status          *coretypes.ResultStatus
	blocks          map[int64]*coretypes.ResultBlock
	requestedBlocks []int64
}

func (s *upgradeCheckpointRPCStub) Status(context.Context) (*coretypes.ResultStatus, error) {
	return s.status, nil
}

func (s *upgradeCheckpointRPCStub) Block(_ context.Context, height *int64) (*coretypes.ResultBlock, error) {
	if height == nil {
		return nil, errors.New("explicit block height is required")
	}
	s.requestedBlocks = append(s.requestedBlocks, *height)
	return s.blocks[*height], nil
}

func TestUpgradeCheckpointObservationValidate(t *testing.T) {
	t.Parallel()
	valid := UpgradeCheckpointObservation{
		ObservedAt:    time.Now().UTC(),
		Node:          "fullnode-0",
		QueryBoundary: UpgradeCheckpointQueryBoundaryCometBFTRPC,
		Height:        77,
		BlockID:       "AABB",
		AppHash:       "CCDD",
	}
	require.NoError(t, valid.Validate())

	for _, test := range []struct {
		name   string
		mutate func(*UpgradeCheckpointObservation)
		want   string
	}{
		{name: "time", mutate: func(value *UpgradeCheckpointObservation) { value.ObservedAt = time.Time{} }, want: "observed_at"},
		{name: "node", mutate: func(value *UpgradeCheckpointObservation) { value.Node = "" }, want: "node"},
		{name: "boundary", mutate: func(value *UpgradeCheckpointObservation) { value.QueryBoundary = "cli" }, want: "query_boundary"},
		{name: "height", mutate: func(value *UpgradeCheckpointObservation) { value.Height = 0 }, want: "height"},
		{name: "block ID", mutate: func(value *UpgradeCheckpointObservation) { value.BlockID = "" }, want: "block_id"},
		{name: "app hash", mutate: func(value *UpgradeCheckpointObservation) { value.AppHash = "" }, want: "app_hash"},
	} {
		t.Run(test.name, func(t *testing.T) {
			candidate := valid
			test.mutate(&candidate)
			require.ErrorContains(t, candidate.Validate(), test.want)
		})
	}
}

func TestCaptureUpgradeCheckpointObservationUsesLatestApplicationHash(t *testing.T) {
	rpc := &upgradeCheckpointRPCStub{
		status: &coretypes.ResultStatus{SyncInfo: coretypes.SyncInfo{
			LatestBlockHeight: 42,
			LatestAppHash:     []byte{0xcc, 0xdd},
		}},
		blocks: map[int64]*coretypes.ResultBlock{
			42: {
				BlockID: cmttypes.BlockID{Hash: []byte{0xaa, 0xbb}},
				Block: &cmttypes.Block{Header: cmttypes.Header{
					Height:  42,
					AppHash: []byte{0x00, 0x11},
				}},
			},
		},
	}

	observation, err := captureUpgradeCheckpointObservationRPC(context.Background(), "fullnode-0", rpc, 0)
	require.NoError(t, err)
	require.Equal(t, []int64{42}, rpc.requestedBlocks)
	require.Equal(t, int64(42), observation.Height)
	require.Equal(t, "AABB", observation.BlockID)
	require.Equal(t, "CCDD", observation.AppHash)
}

func TestCaptureUpgradeCheckpointObservationUsesHistoricalCarrierAppHash(t *testing.T) {
	rpc := &upgradeCheckpointRPCStub{
		status: &coretypes.ResultStatus{SyncInfo: coretypes.SyncInfo{
			LatestBlockHeight: 43,
			LatestAppHash:     []byte{0xee, 0xff},
		}},
		blocks: map[int64]*coretypes.ResultBlock{
			41: {
				BlockID: cmttypes.BlockID{Hash: []byte{0xaa, 0xbb}},
				Block:   &cmttypes.Block{Header: cmttypes.Header{Height: 41}},
			},
			42: {
				BlockID: cmttypes.BlockID{Hash: []byte{0x00, 0x11}},
				Block: &cmttypes.Block{Header: cmttypes.Header{
					Height:  42,
					AppHash: []byte{0xcc, 0xdd},
				}},
			},
		},
	}

	observation, err := captureUpgradeCheckpointObservationRPC(context.Background(), "fullnode-0", rpc, 41)
	require.NoError(t, err)
	require.Equal(t, []int64{41, 42}, rpc.requestedBlocks)
	require.Equal(t, int64(41), observation.Height)
	require.Equal(t, "AABB", observation.BlockID)
	require.Equal(t, "CCDD", observation.AppHash)
}

func TestWaitForUpgradeApplicationHeightRetriesOnlyFutureHeight(t *testing.T) {
	transient := errors.New("rpc error: code = Unknown desc = codespace sdk code 26: invalid height: cannot query with height in the future; please provide a valid height")
	attempts := 0
	err := waitForUpgradeApplicationHeight(context.Background(), 42, time.Nanosecond, func(context.Context) error {
		attempts++
		if attempts == 1 {
			return transient
		}
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 2, attempts)

	permanent := errors.New("rpc error: code = Unavailable desc = connection refused")
	attempts = 0
	err = waitForUpgradeApplicationHeight(context.Background(), 42, time.Nanosecond, func(context.Context) error {
		attempts++
		return permanent
	})
	require.ErrorContains(t, err, permanent.Error())
	require.Equal(t, 1, attempts)
}
