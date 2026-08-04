package harness

import (
	"strings"
	"testing"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
)

func TestCometStateSyncPlanUsesExactTwoRPCServersAndSafeDefaults(t *testing.T) {
	trustHash := strings.Repeat("ab", 32)
	plan, err := newCometStateSyncPlan(
		[]string{"http://panacea-val-0:26657", "http://panacea-val-1:26657"},
		42,
		trustHash,
		cometStateSyncPlanOptions{},
	)
	require.NoError(t, err)
	require.Equal(t, int64(42), plan.TrustHeight)
	require.Equal(t, strings.ToUpper(trustHash), plan.TrustHash)
	require.Equal(t, defaultCometStateSyncTrustPeriod, plan.TrustPeriod)
	require.Equal(t, defaultCometStateSyncDiscoveryTime, plan.DiscoveryTime)
	require.Equal(t, defaultCometStateSyncChunkRequestTimeout, plan.ChunkRequestTimeout)
	require.Equal(t, defaultCometStateSyncChunkFetchers, plan.ChunkFetchers)

	override := plan.tomlOverride()
	stateSync, ok := override["statesync"].(testutil.Toml)
	require.True(t, ok)
	require.Equal(t, true, stateSync["enable"])
	require.Equal(t, "http://panacea-val-0:26657,http://panacea-val-1:26657", stateSync["rpc_servers"])
	require.Equal(t, int64(42), stateSync["trust_height"])
	require.Equal(t, strings.ToUpper(trustHash), stateSync["trust_hash"])
	require.Equal(t, "168h0m0s", stateSync["trust_period"])
	require.Equal(t, "5s", stateSync["discovery_time"])
	require.Equal(t, "10s", stateSync["chunk_request_timeout"])
	require.Equal(t, "4", stateSync["chunk_fetchers"])

	rendered := []byte(`[statesync]
enable = true
rpc_servers = "http://panacea-val-0:26657,http://panacea-val-1:26657"
trust_height = 42
trust_hash = "` + strings.ToUpper(trustHash) + `"
trust_period = "168h0m0s"
discovery_time = "5s"
chunk_request_timeout = "10s"
chunk_fetchers = "4"
`)
	require.NoError(t, validateRenderedCometStateSyncConfig(rendered, plan))

	broken := strings.Replace(string(rendered), "trust_height = 42", "trust_height = 41", 1)
	require.ErrorContains(t, validateRenderedCometStateSyncConfig([]byte(broken), plan), "trust_height")
}

func TestCometStateSyncPlanRejectsUnsafeInputs(t *testing.T) {
	validHash := strings.Repeat("01", 32)
	testCases := []struct {
		name    string
		servers []string
		height  int64
		hash    string
		options cometStateSyncPlanOptions
	}{
		{name: "one RPC", servers: []string{"http://one:26657"}, height: 2, hash: validHash},
		{name: "duplicate RPC", servers: []string{"http://one:26657", "http://one:26657"}, height: 2, hash: validHash},
		{name: "non HTTP RPC", servers: []string{"tcp://one:26657", "http://two:26657"}, height: 2, hash: validHash},
		{name: "zero height", servers: []string{"http://one:26657", "http://two:26657"}, height: 0, hash: validHash},
		{name: "short hash", servers: []string{"http://one:26657", "http://two:26657"}, height: 2, hash: "abcd"},
		{name: "short discovery", servers: []string{"http://one:26657", "http://two:26657"}, height: 2, hash: validHash, options: cometStateSyncPlanOptions{DiscoveryTime: time.Second}},
		{name: "short chunk timeout", servers: []string{"http://one:26657", "http://two:26657"}, height: 2, hash: validHash, options: cometStateSyncPlanOptions{ChunkRequestTimeout: time.Second}},
		{name: "negative fetchers", servers: []string{"http://one:26657", "http://two:26657"}, height: 2, hash: validHash, options: cometStateSyncPlanOptions{ChunkFetchers: -1}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := newCometStateSyncPlan(testCase.servers, testCase.height, testCase.hash, testCase.options)
			require.Error(t, err)
		})
	}
}

func TestFreshCometStateSyncInventoryRejectsApplicationAndCometState(t *testing.T) {
	home := "/var/cosmos-chain/panacea-run"
	inventory, err := parseStateSyncDataInventory(
		home,
		[]byte(home+"/data/priv_validator_state.json\n"),
	)
	require.NoError(t, err)
	require.Equal(t, []string{"data/priv_validator_state.json"}, inventory)

	inventory, err = parseStateSyncDataInventory(home, nil)
	require.NoError(t, err)
	require.Empty(t, inventory)

	for _, forbidden := range []string{
		"application.db",
		"blockstore.db",
		"state.db",
		"tx_index.db",
		"evidence.db",
		"snapshots",
		"cs.wal",
	} {
		t.Run(forbidden, func(t *testing.T) {
			output := []byte(home + "/data/priv_validator_state.json\n" + home + "/data/" + forbidden + "\n")
			_, err := parseStateSyncDataInventory(home, output)
			require.ErrorContains(t, err, "pre-existing state")
		})
	}
	_, err = parseStateSyncDataInventory(home, []byte("/tmp/application.db\n"))
	require.ErrorContains(t, err, "escapes data directory")
}

func TestCometStateSyncLogContractRequiresSnapshotChunksAndABCIProof(t *testing.T) {
	logs := []byte(`INF Discovered new snapshot height=20 format=1
INF Snapshot accepted, restoring height=20 format=1
INF Fetching snapshot chunk height=20 chunk=0
INF Fetching snapshot chunk height=20 chunk=1
INF Applied snapshot chunk to ABCI app height=20 chunk=0
INF Applied snapshot chunk to ABCI app height=20 chunk=1
INF Verified ABCI app height=20 appHash=CAFE
INF Snapshot restored height=20 format=1
`)
	evidence := parseCometStateSyncLogs(logs)
	require.NoError(t, evidence.validate())
	require.True(t, evidence.DiscoveredSnapshot)
	require.True(t, evidence.AcceptedSnapshot)
	require.Equal(t, 2, evidence.FetchedChunks)
	require.Equal(t, 2, evidence.AppliedChunks)
	require.True(t, evidence.VerifiedABCIApp)
	require.True(t, evidence.RestoredSnapshot)
	require.Equal(t, int64(20), evidence.SnapshotHeight)
	require.Len(t, evidence.MatchedLines, 8)

	blockSyncOnly := parseCometStateSyncLogs([]byte("INF executed block height=20\nINF committed state height=20\n"))
	require.ErrorContains(t, blockSyncOnly.validate(), "Discovered new snapshot")

	missingApply := parseCometStateSyncLogs([]byte(strings.ReplaceAll(string(logs), "Applied snapshot chunk to ABCI app", "applied ordinary block")))
	require.ErrorContains(t, missingApply.validate(), "Applied snapshot chunk to ABCI app")
}

func TestCometStateSyncLogHeightAcceptsStructuredLogOrdering(t *testing.T) {
	logs := []byte(`{"height":25,"msg":"Snapshot restored","module":"statesync"}`)
	evidence := parseCometStateSyncLogs(logs)
	require.True(t, evidence.RestoredSnapshot)
	require.Equal(t, int64(25), evidence.SnapshotHeight)
}

func TestCometStateSyncProviderContractRequiresCompletedAndVerifiableSnapshot(t *testing.T) {
	since := time.Date(2026, 8, 4, 1, 2, 3, 0, time.UTC)
	providerA := parseCometStateSyncProviderLogs(
		"provider-a",
		since,
		"provider-a.log",
		[]byte("INF completed state snapshot height=10 format=1\nINF completed state snapshot height=\x1b[0m20 format=1\n"),
	)
	providerB := parseCometStateSyncProviderLogs(
		"provider-b",
		since,
		"provider-b.log",
		[]byte("INF ordinary block height=21\n"),
	)
	require.Equal(t, []int64{10, 20}, providerA.CompletedSnapshotHeights)
	require.Empty(t, providerB.CompletedSnapshotHeights)

	height, ok := usableProviderSnapshotHeight([]CometStateSyncProviderEvidence{providerA, providerB}, 21)
	require.True(t, ok)
	require.Equal(t, int64(10), height)
	height, ok = usableProviderSnapshotHeight([]CometStateSyncProviderEvidence{providerA, providerB}, 22)
	require.True(t, ok)
	require.Equal(t, int64(20), height)
}

func TestCometStateSyncBadTrustHashContractIsBoundedAndCannotSucceed(t *testing.T) {
	original := strings.Repeat("00", 32)
	mutated, err := mutateCometStateSyncTrustHash(original)
	require.NoError(t, err)
	require.NotEqual(t, original, mutated)
	require.Len(t, mutated, 64)

	logs := parseCometStateSyncBadTrustLogs([]byte(
		`ERR failed to start state sync err="failed to set up light client state provider: expected header's hash 0100, but got FF00"`,
	))
	require.True(t, logs.RejectedTrustHash)
	require.False(t, logs.UnexpectedSuccess)
	require.NoError(t, validateCometStateSyncBadTrustFailure(logs, 12*time.Second, 15*time.Second))
	require.ErrorContains(
		t,
		validateCometStateSyncBadTrustFailure(logs, 22*time.Second, 15*time.Second),
		"exceeded bounded deadline",
	)

	successLogs := parseCometStateSyncBadTrustLogs([]byte(
		"ERR trusted header hash does not match\nINF Verified ABCI app height=20\nINF Snapshot restored height=20\n",
	))
	require.ErrorContains(
		t,
		validateCometStateSyncBadTrustFailure(successLogs, time.Second, 15*time.Second),
		"unexpectedly restored",
	)
	require.ErrorContains(
		t,
		validateCometStateSyncBadTrustFailure(CometStateSyncBadTrustLogEvidence{}, time.Second, 15*time.Second),
		"do not prove",
	)
}

func TestCometStateSyncQueryContinuityIsSemanticAndPinsHarnessFlags(t *testing.T) {
	require.NoError(t, validateCometStateSyncQueryCommand([]string{"bank", "balances", "panacea1abc"}))
	for _, forbidden := range []string{"--height", "--node", "--home", "--output"} {
		require.ErrorContains(
			t,
			validateCometStateSyncQueryCommand([]string{"bank", "total", forbidden}),
			forbidden,
		)
	}
	require.ErrorContains(
		t,
		validateCometStateSyncQueryCommand([]string{"bank", "total", "--height=77"}),
		"--height",
	)
	require.NoError(
		t,
		validateCometStateSyncQueryContinuity(
			[]byte(`{"balances":[{"denom":"umed","amount":"42"}],"pagination":{"total":"1"}}`),
			[]byte(`{"pagination":{"total":"1"},"balances":[{"amount":"42","denom":"umed"}]}`),
			"historical",
		),
	)
	require.ErrorContains(
		t,
		validateCometStateSyncQueryContinuity(
			[]byte(`{"amount":"42"}`),
			[]byte(`{"amount":"41"}`),
			"current",
		),
		"changed across",
	)
}
