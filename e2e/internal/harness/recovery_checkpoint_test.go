package harness

import (
	"context"
	"testing"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"
)

type recoveryRPCStub struct {
	status          *coretypes.ResultStatus
	block           *coretypes.ResultBlock
	requestedHeight int64
}

func (s *recoveryRPCStub) Status(context.Context) (*coretypes.ResultStatus, error) {
	return s.status, nil
}

func (s *recoveryRPCStub) Block(_ context.Context, height *int64) (*coretypes.ResultBlock, error) {
	if height != nil {
		s.requestedHeight = *height
	}
	return s.block, nil
}

func TestValidateRecoveryContinuityAcceptsStableHistoryAndProgressedHead(t *testing.T) {
	before := RecoveryCheckpoint{
		Node:    "fullnode-0",
		Height:  42,
		BlockID: "AABB",
		AppHash: "CCDD",
	}
	afterSameHeight := before
	afterSameHeight.Node = "fullnode-0-restarted"
	progressed := RecoveryCheckpoint{
		Node:    "fullnode-0-restarted",
		Height:  45,
		BlockID: "EEFF",
		AppHash: "0011",
	}

	require.NoError(t, ValidateRecoveryContinuity(before, afterSameHeight, progressed))
}

func TestCaptureRecoveryCheckpointUsesLatestCommittedHeight(t *testing.T) {
	rpc := &recoveryRPCStub{
		status: &coretypes.ResultStatus{SyncInfo: coretypes.SyncInfo{LatestBlockHeight: 42}},
		block: &coretypes.ResultBlock{
			BlockID: cmttypes.BlockID{Hash: []byte{0xaa, 0xbb}},
			Block: &cmttypes.Block{Header: cmttypes.Header{
				Height:  42,
				AppHash: []byte{0xcc, 0xdd},
			}},
		},
	}

	checkpoint, err := captureRecoveryCheckpoint(context.Background(), "after-restart", "fullnode-0", rpc, 0)
	require.NoError(t, err)
	require.Equal(t, int64(42), rpc.requestedHeight)
	require.Equal(t, int64(42), checkpoint.Height)
	require.Equal(t, "AABB", checkpoint.BlockID)
	require.Equal(t, "CCDD", checkpoint.AppHash)
}
