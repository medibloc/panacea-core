package e2e_test

import (
	"context"
	"os"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestPortableApplicationSnapshotRestoreAndFreshFullNodeSync(t *testing.T) {
	if os.Getenv("PANACEA_E2E_RESTART") != "1" {
		t.Skip("set PANACEA_E2E_RESTART=1 or use ./scripts/e2e/run.sh restart")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	network, err := harness.Start(ctx, t, harness.Config{
		Image:         harness.CurrentImage(),
		NumValidators: 1,
		NumFullNodes:  1,
	})
	require.NoError(t, err)
	defer network.RecordTestPanic()
	require.NoError(t, network.WaitForHeight(ctx, 5))
	require.NoError(t, network.WaitForFullNode(ctx, 5))

	users := interchaintest.GetAndFundTestUsers(
		t,
		ctx,
		"snapshot-restore",
		sdkmath.NewInt(2_000_000),
		network.Chain,
	)
	recipient, err := network.BuildWallet(ctx, "snapshot-recipient", "")
	require.NoError(t, err)
	committed, err := network.BroadcastAndWaitTx(
		ctx,
		"snapshot-bank-send",
		network.Chain.Validators[0],
		users[0].KeyName(),
		"bank", "send",
		users[0].KeyName(),
		recipient.FormattedAddress(),
		"4321umed",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	require.Positive(t, committed.HeightInt64())

	before, err := network.QueryFullNodeBalance(ctx, recipient.FormattedAddress(), "umed")
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(4321), before)

	restored, err := network.RestoreFullNodeApplicationFromPortableSnapshot(
		ctx,
		"portable-application-snapshot",
		0,
	)
	require.NoError(t, err)
	require.Positive(t, restored.Snapshot.Height)
	require.NotEmpty(t, restored.Archive)
	require.NotEmpty(t, restored.BackupDB)
	require.Len(t, restored.RestoredHistory, 2)
	require.Len(t, restored.CaughtUpHistory, 2)
	require.NotEmpty(t, restored.Application.AppHash)

	afterRestore, err := network.QueryFullNodeBalance(ctx, recipient.FormattedAddress(), "umed")
	require.NoError(t, err)
	require.Equal(t, before, afterRestore)

	synced, err := network.AddAndSyncFreshFullNode(ctx, "fresh-full-node-block-sync")
	require.NoError(t, err)
	require.Equal(t, 1, synced.PreviousNodeCount)
	require.Len(t, synced.History, 2)
	require.Positive(t, synced.TargetHeight)

	freshNode := network.Chain.FullNodes[synced.PreviousNodeCount]
	response, err := banktypes.NewQueryClient(freshNode.GrpcConn).Balance(
		ctx,
		&banktypes.QueryBalanceRequest{
			Address: recipient.FormattedAddress(),
			Denom:   "umed",
		},
	)
	require.NoError(t, err)
	require.NotNil(t, response.Balance)
	require.Equal(t, before, response.Balance.Amount)
}
