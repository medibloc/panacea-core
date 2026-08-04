package e2e_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestFourValidatorQuorumFailureAndRecovery(t *testing.T) {
	if os.Getenv("PANACEA_E2E_CONSENSUS") != "1" {
		t.Skip("set PANACEA_E2E_CONSENSUS=1 or use ./scripts/e2e/run.sh consensus")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	network, err := harness.Start(ctx, t, harness.Config{
		Image:         harness.CurrentImage(),
		NumValidators: 4,
		NumFullNodes:  1,
		TimeoutCommit: "1s",
	})
	require.NoError(t, err)
	defer network.RecordTestPanic()

	validators := network.Chain.Validators
	fullNode := network.Chain.FullNodes[0]
	require.NoError(t, network.WaitForHeight(ctx, 5))
	require.NoError(t, network.WaitForFullNode(ctx, 5))
	powers, err := network.ValidatorSet(ctx, fullNode, 5)
	require.NoError(t, err)
	require.Len(t, powers, 4)
	require.Positive(t, powers[0].Power)
	for _, validator := range powers[1:] {
		require.Equal(t, powers[0].Power, validator.Power, "validators must have equal voting power")
	}
	_, err = network.RequireSameHistoryAtHeight(ctx, 5, network.Chain.Nodes()...)
	require.NoError(t, err)

	creator := buildAndFundNFTWallet(t, ctx, network, "consensus-creator")
	firstOwner := buildAndFundNFTWallet(t, ctx, network, "consensus-owner-first")
	finalOwner := buildAndFundNFTWallet(t, ctx, network, "consensus-owner-final")
	classID := creator.FormattedAddress() + ":consensus.class"
	nftID := "quorum.1"
	_, err = network.BroadcastAndWaitTx(
		ctx, "consensus-create-class", validators[0], creator.KeyName(),
		"nft", "create-class", "consensus.class", "Consensus Class", "QRM",
		"owner-transferable", "true", "2",
		"--gas", "500000", "--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	_, err = network.BroadcastAndWaitTx(
		ctx, "consensus-mint-before-disruption", validators[0], creator.KeyName(),
		"nft", "mint", classID, nftID, firstOwner.FormattedAddress(),
		"--data", lifecycleDataJSON,
		"--gas", "500000", "--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	beforeDisruption := queryNFTRecord(t, ctx, network, "consensus-before-disruption", classID, nftID)
	require.NotNil(t, beforeDisruption.NFTRecord.Live)

	beforeFirstStop, err := fullNode.Height(ctx)
	require.NoError(t, err)
	require.NoError(t, validators[0].StopContainer(ctx))
	require.NoError(t, network.AppendArtifactJSON("nodes/disruptions.jsonl", map[string]any{
		"phase": "one-validator-stopped", "node": validators[0].Name(), "height": beforeFirstStop,
	}))
	require.NoError(t, network.WaitForNodeHeight(ctx, fullNode, beforeFirstStop+3))

	beforeSecondStop, err := fullNode.Height(ctx)
	require.NoError(t, err)
	require.NoError(t, validators[1].StopContainer(ctx))
	require.NoError(t, network.AppendArtifactJSON("nodes/disruptions.jsonl", map[string]any{
		"phase": "quorum-lost", "node": validators[1].Name(), "height": beforeSecondStop,
	}))
	haltCtx, haltCancel := context.WithTimeout(ctx, 20*time.Second)
	haltedHeight, err := network.WaitForStableHeight(haltCtx, fullNode, 5*time.Second)
	haltCancel()
	require.NoError(t, err)
	require.GreaterOrEqual(t, haltedHeight, beforeSecondStop)

	require.NoError(t, validators[0].StartContainer(ctx))
	require.NoError(t, network.WaitForNodeHeight(ctx, fullNode, haltedHeight+3))
	require.NoError(t, network.AppendArtifactJSON("nodes/disruptions.jsonl", map[string]any{
		"phase": "quorum-restored", "node": validators[0].Name(), "halted_height": haltedHeight,
	}))

	_, err = network.BroadcastAndWaitTx(
		ctx, "consensus-send-after-recovery", validators[0], firstOwner.KeyName(),
		"nft", "send", classID, nftID, finalOwner.FormattedAddress(),
		"--gas", "500000", "--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	afterRecovery := queryNFTRecord(t, ctx, network, "consensus-after-recovery", classID, nftID)
	require.NotNil(t, afterRecovery.NFTRecord.Live)
	require.Equal(t, finalOwner.FormattedAddress(), afterRecovery.NFTRecord.Live.Owner)

	require.NoError(t, validators[1].StartContainer(ctx))
	target, err := fullNode.Height(ctx)
	require.NoError(t, err)
	target += 3
	for _, node := range network.Chain.Nodes() {
		require.NoError(t, network.WaitForNodeHeight(ctx, node, target), node.Name())
	}
	_, err = network.RequireSameHistoryAtHeight(ctx, target, network.Chain.Nodes()...)
	require.NoError(t, err)

	for attempt := 1; attempt <= 2; attempt++ {
		restart, restartErr := network.GracefulRestartNode(ctx, fullNode)
		require.NoError(t, restartErr)
		require.Greater(t, restart.After.Height, restart.Before.Height)
		finalHeight := restart.After.Height
		for _, validator := range validators {
			require.NoError(t, network.WaitForNodeHeight(ctx, validator, finalHeight))
		}
		_, restartErr = network.RequireSameHistoryAtHeight(ctx, finalHeight, network.Chain.Nodes()...)
		require.NoError(t, restartErr, "full-node restart attempt %d", attempt)
	}
}
