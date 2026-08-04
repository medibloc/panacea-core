package harness

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	rpcclient "github.com/cometbft/cometbft/rpc/client"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type quorumRPCStub struct {
	rpcclient.Client
	status       *coretypes.ResultStatus
	netInfo      *coretypes.ResultNetInfo
	block        *coretypes.ResultBlock
	blockResults *coretypes.ResultBlockResults
}

func (s *quorumRPCStub) Status(context.Context) (*coretypes.ResultStatus, error) {
	return s.status, nil
}

func (s *quorumRPCStub) NetInfo(context.Context) (*coretypes.ResultNetInfo, error) {
	return s.netInfo, nil
}

func (s *quorumRPCStub) Block(context.Context, *int64) (*coretypes.ResultBlock, error) {
	return s.block, nil
}

func (s *quorumRPCStub) BlockResults(context.Context, *int64) (*coretypes.ResultBlockResults, error) {
	return s.blockResults, nil
}

func TestWaitForQuorumNodeStateRejectsNodeWithoutConnectedPeers(t *testing.T) {
	t.Parallel()

	const targetHeight int64 = 42
	node := newQuorumTestNode(newQuorumTestChain(), 0, 0, targetHeight)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := waitForQuorumNodeState(ctx, targetHeight, node)

	require.ErrorContains(t, err, "connected peers")
}

func TestWaitForQuorumAgreementRecordsConnectedPeersForEveryNode(t *testing.T) {
	t.Parallel()

	const targetHeight int64 = 42
	chain := newQuorumTestChain()
	store, err := newArtifactStore(t.Name(), "run-123456789abc", Config{ArtifactRoot: trustedArtifactTempDir(t)})
	require.NoError(t, err)
	network := &Network{artifacts: store}
	nodes := []*cosmos.ChainNode{
		newQuorumTestNode(chain, 0, 2, targetHeight),
		newQuorumTestNode(chain, 1, 3, targetHeight),
	}

	_, err = network.WaitForQuorumAgreement(context.Background(), "peer-evidence", targetHeight, nodes...)
	require.NoError(t, err)
	raw, err := os.ReadFile(filepath.Join(store.dir, "queries", "quorum-observations.jsonl"))
	require.NoError(t, err)
	var observation struct {
		Evidence struct {
			States []QuorumNodeState `json:"states"`
		} `json:"evidence"`
	}
	require.NoError(t, json.Unmarshal(raw, &observation))
	require.Len(t, observation.Evidence.States, len(nodes))
	require.Equal(t, 2, observation.Evidence.States[0].Peers)
	require.Equal(t, 3, observation.Evidence.States[1].Peers)
}

func newQuorumTestChain() *cosmos.CosmosChain {
	return cosmos.NewCosmosChain(
		"quorum-peer-evidence",
		ibc.ChainConfig{ChainID: "quorum-peer-evidence"},
		2,
		0,
		zap.NewNop(),
	)
}

func newQuorumTestNode(
	chain *cosmos.CosmosChain,
	index int,
	peers int,
	targetHeight int64,
) *cosmos.ChainNode {
	return &cosmos.ChainNode{
		Index:     index,
		Chain:     chain,
		Validator: true,
		TestName:  "quorum-peer-evidence",
		Client: &quorumRPCStub{
			status: &coretypes.ResultStatus{
				SyncInfo: coretypes.SyncInfo{LatestBlockHeight: targetHeight},
			},
			netInfo: &coretypes.ResultNetInfo{NPeers: peers},
			block: &coretypes.ResultBlock{
				BlockID: cmttypes.BlockID{Hash: []byte{0xaa, 0xbb}},
				Block: &cmttypes.Block{Header: cmttypes.Header{
					Height: targetHeight,
				}},
			},
			blockResults: &coretypes.ResultBlockResults{
				Height:  targetHeight,
				AppHash: []byte{0xcc, 0xdd},
			},
		},
	}
}
