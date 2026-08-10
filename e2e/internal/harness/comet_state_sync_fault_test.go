package harness

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLatestCommonProviderSnapshotHeightRequiresEveryProvider(t *testing.T) {
	providers := []CometStateSyncProviderEvidence{
		{CompletedSnapshotHeights: []int64{10, 20, 30}},
		{CompletedSnapshotHeights: []int64{10, 30, 40}},
	}
	height, ok := latestCommonProviderSnapshotHeight(providers)
	require.True(t, ok)
	require.Equal(t, int64(30), height)

	_, ok = latestCommonProviderSnapshotHeight([]CometStateSyncProviderEvidence{
		{CompletedSnapshotHeights: []int64{20}},
		{CompletedSnapshotHeights: []int64{30}},
	})
	require.False(t, ok)
}

func TestUnavailableProviderFailureRequiresSetupAndTransportEvidence(t *testing.T) {
	logs := parseCometStateSyncUnavailableProviderLogs([]byte(
		`ERR failed to start state sync err="failed to set up light client state provider: failed to fetch light block: Get \"http://127.0.0.1:1/commit?height=20\": dial tcp 127.0.0.1:1: connect: connection refused"`,
	))
	require.True(t, logs.LightClientSetupFailed)
	require.True(t, logs.ProviderTransportFailure)
	require.False(t, logs.UnexpectedSuccess)
	require.NoError(t, validateCometStateSyncUnavailableProviderFailure(logs, 3*time.Second, 10*time.Second))

	missingTransport := logs
	missingTransport.ProviderTransportFailure = false
	require.ErrorContains(
		t,
		validateCometStateSyncUnavailableProviderFailure(missingTransport, 3*time.Second, 10*time.Second),
		"transport",
	)
	success := parseCometStateSyncUnavailableProviderLogs([]byte(
		"ERR failed to set up light client state provider: connection refused\nINF Snapshot restored height=30\n",
	))
	require.ErrorContains(
		t,
		validateCometStateSyncUnavailableProviderFailure(success, time.Second, 10*time.Second),
		"unexpectedly restored",
	)
}

func TestCorruptedChunkFailureRequiresTwoChecksumRejectionsAndPeerExhaustion(t *testing.T) {
	logs := parseCometStateSyncCorruptedChunkLogs([]byte(`
ERR chunk checksum mismatch; rejecting sender and requesting refetch chunk=0 sender=peer-a
ERR chunk checksum mismatch; rejecting sender and requesting refetch chunk=0 sender=peer-b
ERR No valid peers found for snapshot height=30 format=3
`))
	require.Equal(t, 2, logs.ChecksumMismatches)
	require.True(t, logs.NoValidPeers)
	require.False(t, logs.UnexpectedSuccess)
	require.NoError(t, validateCometStateSyncCorruptedChunkFailure(logs, 8*time.Second, 15*time.Second))

	oneProvider := logs
	oneProvider.ChecksumMismatches = 1
	require.ErrorContains(
		t,
		validateCometStateSyncCorruptedChunkFailure(oneProvider, 8*time.Second, 15*time.Second),
		"at least 2",
	)
	success := parseCometStateSyncCorruptedChunkLogs([]byte(
		"ERR chunk checksum mismatch; rejecting sender and requesting refetch\n" +
			"ERR chunk checksum mismatch; rejecting sender and requesting refetch\n" +
			"ERR No valid peers found for snapshot\n" +
			"INF Verified ABCI app height=30\n",
	))
	require.ErrorContains(
		t,
		validateCometStateSyncCorruptedChunkFailure(success, 8*time.Second, 15*time.Second),
		"unexpectedly restored",
	)
}

func TestParseCometStateSyncProviderSnapshotInterval(t *testing.T) {
	interval, err := parseCometStateSyncProviderSnapshotInterval([]byte(`
[state-sync]
snapshot-interval = 30
snapshot-keep-recent = 3
`))
	require.NoError(t, err)
	require.Equal(t, uint64(30), interval)

	_, err = parseCometStateSyncProviderSnapshotInterval([]byte("[state-sync]\nsnapshot-interval = 0\n"))
	require.ErrorContains(t, err, "snapshot-interval")
}

func TestParseAndSelectCommonCometStateSyncChunkZero(t *testing.T) {
	homeA := "/var/cosmos-chain/panacea-a"
	homeB := "/var/cosmos-chain/panacea-b"
	providerA, err := parseCometStateSyncSnapshotChunkPaths(homeA, 30, []byte(
		homeA+"/data/snapshots/30/3/1\n"+
			homeA+"/data/snapshots/30/3/0\n"+
			homeA+"/data/snapshots/30/2/0\n",
	))
	require.NoError(t, err)
	require.Len(t, providerA, 3)
	require.Equal(t, uint32(2), providerA[0].Format)
	require.Equal(t, uint32(0), providerA[0].Chunk)

	providerB, err := parseCometStateSyncSnapshotChunkPaths(homeB, 30, []byte(
		homeB+"/data/snapshots/30/3/0\n"+
			homeB+"/data/snapshots/30/3/1\n",
	))
	require.NoError(t, err)
	format, selected, err := selectCommonCometStateSyncChunkZero([][]cometStateSyncChunkPath{providerA, providerB})
	require.NoError(t, err)
	require.Equal(t, uint32(3), format)
	require.Equal(t, "data/snapshots/30/3/0", selected[0].RelativePath)
	require.Equal(t, "data/snapshots/30/3/0", selected[1].RelativePath)

	_, err = parseCometStateSyncSnapshotChunkPaths(homeA, 30, []byte("/tmp/30/3/0\n"))
	require.ErrorContains(t, err, "escapes")
	_, _, err = selectCommonCometStateSyncChunkZero([][]cometStateSyncChunkPath{
		{{Format: 2, Chunk: 0}},
		{{Format: 3, Chunk: 0}},
	})
	require.ErrorContains(t, err, "common format")
}

func TestRenderedCometStateSyncPEXMustBeDisabledForFaults(t *testing.T) {
	require.NoError(t, validateRenderedCometStateSyncPEXDisabled([]byte("[p2p]\npex = false\n")))
	require.ErrorContains(t, validateRenderedCometStateSyncPEXDisabled([]byte("[p2p]\npex = true\n")), "want false")
}

func TestCometStateSyncFaultTimeoutIsBounded(t *testing.T) {
	timeout, err := normalizeCometStateSyncFaultTimeout(0)
	require.NoError(t, err)
	require.Equal(t, defaultCometStateSyncFaultTimeout, timeout)
	_, err = normalizeCometStateSyncFaultTimeout(5 * time.Second)
	require.Error(t, err)
	_, err = normalizeCometStateSyncFaultTimeout(90 * time.Second)
	require.Error(t, err)
	require.ErrorContains(
		t,
		validateCometStateSyncFaultDeadline(17*time.Second, 10*time.Second, "fault"),
		"exceeded bounded deadline",
	)
}
