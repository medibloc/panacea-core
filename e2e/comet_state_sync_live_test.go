package e2e_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

// TestActualCometStateSyncAndBadTrustHash is an opt-in current-version vertical
// slice. The connected upgrade-deep scenario separately calls the same
// successful-join method after switching its existing source nodes;
// unavailable RPC providers, corrupted chunks, and a bad trust hash are
// retained here as bounded fault contracts.
func TestActualCometStateSyncAndBadTrustHash(t *testing.T) {
	if os.Getenv("PANACEA_E2E_STATE_SYNC") != "1" {
		t.Skip("set PANACEA_E2E_STATE_SYNC=1 to run the CometBFT state-sync vertical slice")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()
	providerSince := time.Now().UTC()
	image := harness.CurrentImage()
	network, err := harness.Start(ctx, t, harness.Config{
		Image:              image,
		NumValidators:      2,
		NumFullNodes:       0,
		TimeoutCommit:      "1s",
		SnapshotInterval:   30,
		SnapshotKeepRecent: 3,
	})
	require.NoError(t, err)
	defer network.RecordTestPanic()
	require.NoError(t, network.WaitForHeight(ctx, 32))

	queryAddress, err := network.Chain.Validators[0].AccountKeyBech32(ctx, "validator")
	require.NoError(t, err)
	request := harness.CometStateSyncRequest{
		Step:                  "current-comet-state-sync",
		RPCSources:            network.Chain.Validators,
		ExpectedImage:         image,
		QueryCommand:          []string{"bank", "balances", queryAddress},
		ProviderSnapshotSince: providerSince,
		ProviderWaitTimeout:   45 * time.Second,
		CompletionTimeout:     3 * time.Minute,
	}
	corrupted, err := network.ExpectCometStateSyncCorruptedChunks(ctx, harness.CometStateSyncCorruptedChunkRequest{
		StateSync: harness.CometStateSyncRequest{
			Step:                  "current-comet-state-sync-corrupted-chunks",
			RPCSources:            request.RPCSources,
			ExpectedImage:         request.ExpectedImage,
			ProviderSnapshotSince: providerSince,
			ProviderWaitTimeout:   request.ProviderWaitTimeout,
		},
		SnapshotInterval: 30,
		FailureTimeout:   20 * time.Second,
	})
	require.NoError(t, err)
	require.Equal(t, "actual-cometbft-state-sync-corrupted-chunks", corrupted.Mode)
	require.GreaterOrEqual(t, corrupted.Logs.ChecksumMismatches, 2)
	require.True(t, corrupted.Logs.NoValidPeers)
	require.False(t, corrupted.Logs.UnexpectedSuccess)
	require.Len(t, corrupted.Mutations, 2)
	for _, mutation := range corrupted.Mutations {
		require.NotEqual(t, mutation.OriginalSHA256, mutation.CorruptedSHA256)
		require.Equal(t, mutation.OriginalSHA256, mutation.RestoredSHA256)
		require.True(t, mutation.Restored)
	}
	require.True(t, corrupted.Rejected)
	require.True(t, corrupted.NodeStopped)

	synced, err := network.RunCometStateSync(ctx, request)
	require.NoError(t, err)
	require.Equal(t, "actual-cometbft-state-sync", synced.Mode)
	require.Len(t, synced.Sources, 2)
	require.Len(t, synced.TrustHistory, 2)
	require.True(t, synced.StateSyncLogs.DiscoveredSnapshot)
	require.True(t, synced.StateSyncLogs.AcceptedSnapshot)
	require.Positive(t, synced.StateSyncLogs.FetchedChunks)
	require.Positive(t, synced.StateSyncLogs.AppliedChunks)
	require.True(t, synced.StateSyncLogs.VerifiedABCIApp)
	require.True(t, synced.StateSyncLogs.RestoredSnapshot)
	require.Greater(t, synced.StateSyncLogs.SnapshotHeight, int64(1))
	require.False(t, synced.BeforeRestart.Restored.CatchingUp)
	require.False(t, synced.AfterRestart.Restored.CatchingUp)
	require.True(t, synced.GenesisBlockUnavailable)
	require.True(t, synced.RestartSkippedStateSync)
	require.NotEmpty(t, synced.Queries.CurrentBefore.Response)
	require.NotEmpty(t, synced.Queries.HistoricalBefore.Response)
	require.Equal(t, synced.Queries.CurrentBefore.Response, synced.Queries.CurrentAfter.Response)
	require.Equal(t, synced.Queries.HistoricalBefore.Response, synced.Queries.HistoricalAfter.Response)
	require.True(t, synced.NodeStopped)

	rejected, err := network.ExpectCometStateSyncBadTrustHash(ctx, request, 30*time.Second)
	require.NoError(t, err)
	require.Equal(t, "actual-cometbft-state-sync-bad-trust-hash", rejected.Mode)
	require.NotEqual(t, rejected.OriginalTrustHash, rejected.MutatedTrustHash)
	require.True(t, rejected.Logs.RejectedTrustHash)
	require.False(t, rejected.Logs.UnexpectedSuccess)
	require.True(t, rejected.Rejected)
	require.True(t, rejected.NodeStopped)

	unavailable, err := network.ExpectCometStateSyncUnavailableProviders(ctx, harness.CometStateSyncRequest{
		Step:                  "current-comet-state-sync-unavailable-providers",
		RPCSources:            request.RPCSources,
		ExpectedImage:         request.ExpectedImage,
		ProviderSnapshotSince: providerSince,
		ProviderWaitTimeout:   request.ProviderWaitTimeout,
	}, 20*time.Second)
	require.NoError(t, err)
	require.Equal(t, "actual-cometbft-state-sync-unavailable-providers", unavailable.Mode)
	require.True(t, unavailable.Logs.LightClientSetupFailed)
	require.True(t, unavailable.Logs.ProviderTransportFailure)
	require.False(t, unavailable.Logs.UnexpectedSuccess)
	require.True(t, unavailable.Rejected)
	require.True(t, unavailable.NodeStopped)
}
