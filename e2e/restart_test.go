package e2e_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestPersistenceRestartAndExport(t *testing.T) {
	if os.Getenv("PANACEA_E2E_RESTART") != "1" {
		t.Skip("set PANACEA_E2E_RESTART=1 or use ./scripts/e2e/run.sh restart")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	network, err := harness.Start(ctx, t, harness.Config{
		Image:              harness.CurrentImage(),
		NumValidators:      1,
		NumFullNodes:       1,
		SnapshotInterval:   5,
		SnapshotKeepRecent: 2,
	})
	require.NoError(t, err)
	defer network.RecordTestPanic()
	require.NoError(t, network.WaitForHeight(ctx, 5))
	require.NoError(t, network.WaitForFullNode(ctx, 5))

	creator := buildAndFundNFTWallet(t, ctx, network, "restart-creator")
	firstOwner := buildAndFundNFTWallet(t, ctx, network, "restart-owner-first")
	finalOwner := buildAndFundNFTWallet(t, ctx, network, "restart-owner-final")
	validator := network.Chain.Validators[0]
	fullNode := network.Chain.FullNodes[0]
	classID := creator.FormattedAddress() + ":restart.class"
	nftID := "persist.1"

	_, err = network.BroadcastAndWaitTx(
		ctx, "restart-create-class", validator, creator.KeyName(),
		"nft", "create-class", "restart.class", "Restart Class", "RST",
		"owner-transferable", "true", "2",
		"--gas", "500000", "--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	mintTx, err := network.BroadcastAndWaitTx(
		ctx, "restart-mint", validator, creator.KeyName(),
		"nft", "mint", classID, nftID, firstOwner.FormattedAddress(),
		"--uri", lifecycleNFTURI,
		"--uri-hash", lifecycleNFTURIHash,
		"--data", lifecycleDataJSON,
		"--gas", "500000", "--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	beforeRestart := queryNFTRecord(t, ctx, network, "restart-before-graceful", classID, nftID)
	require.NotNil(t, beforeRestart.NFTRecord.Live)

	graceful, err := network.GracefulRestartNode(ctx, validator)
	require.NoError(t, err)
	require.Greater(t, graceful.After.Height, graceful.Before.Height)
	afterGraceful := queryNFTRecord(t, ctx, network, "restart-after-graceful", classID, nftID)
	require.Equal(t, beforeRestart, afterGraceful)
	require.NoError(t, network.WaitForFullNode(ctx, graceful.After.Height))

	_, err = network.BroadcastAndWaitTx(
		ctx, "restart-send", validator, firstOwner.KeyName(),
		"nft", "send", classID, nftID, finalOwner.FormattedAddress(),
		"--gas", "500000", "--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	_, err = network.BroadcastAndWaitTx(
		ctx, "restart-revoke", validator, creator.KeyName(),
		"nft", "revoke", classID, nftID,
		"--gas", "500000", "--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	_, err = network.BroadcastAndWaitTx(
		ctx, "restart-burn", validator, finalOwner.KeyName(),
		"nft", "burn", classID, nftID,
		"--gas", "500000", "--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	tombstone := queryNFTRecord(t, ctx, network, "restart-tombstone-before-sigkill", classID, nftID)
	require.NotNil(t, tombstone.NFTRecord.BurnTombstone)

	forced, err := network.ForceKillAndRestartNode(ctx, fullNode)
	require.NoError(t, err)
	require.Greater(t, forced.After.Height, forced.Before.Height)
	afterForced := queryNFTRecord(t, ctx, network, "restart-tombstone-after-sigkill", classID, nftID)
	require.Equal(t, tombstone, afterForced)

	exportHeight := mintTx.HeightInt64()
	latest, err := network.Chain.Height(ctx)
	require.NoError(t, err)
	if latest > exportHeight {
		exportHeight = latest
	}
	require.NoError(t, network.WaitForFullNode(ctx, exportHeight))
	_, err = network.RequireSameHistoryAtHeight(ctx, exportHeight, validator, fullNode)
	require.NoError(t, err)

	require.NoError(t, network.Chain.StopAllNodes(ctx))
	exported, err := network.ExportStateTwiceAtHeight(ctx, exportHeight)
	require.NoError(t, err)
	require.NotEmpty(t, exported)
	require.NoError(t, network.Chain.StartAllNodes(ctx))
	require.NoError(t, network.WaitForHeight(ctx, exportHeight+1))
	require.NoError(t, network.WaitForFullNode(ctx, exportHeight+1))
	_, err = network.RequireSameHistoryAtHeight(ctx, exportHeight, validator, fullNode)
	require.NoError(t, err)
	afterExportRestart := queryNFTRecord(t, ctx, network, "restart-tombstone-after-export", classID, nftID)
	require.Equal(t, tombstone, afterExportRestart)
}
