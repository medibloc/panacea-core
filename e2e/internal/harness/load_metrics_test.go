package harness

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSummarizeLoadQueriesUsesNearestRankPercentiles(t *testing.T) {
	t.Parallel()

	started := time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC)
	samples := make([]LoadQuerySample, 100)
	for index := range samples {
		samples[index] = LoadQuerySample{
			StartedAt:  started.Add(time.Duration(index) * time.Millisecond),
			FinishedAt: started.Add(time.Duration(index+1) * time.Millisecond),
			Boundary:   "rest",
			Dataset:    "classes",
			Success:    true,
		}
	}
	samples[3].Success = false
	samples[3].TimedOut = true
	samples[3].StatusCode = 504

	summary, err := SummarizeLoadQueries(samples)
	require.NoError(t, err)
	require.Equal(t, 100, summary.Requests)
	require.Equal(t, 99, summary.Successes)
	require.Equal(t, 1, summary.Failures)
	require.Equal(t, 1, summary.Timeouts)
	require.Equal(t, int64(100), summary.WindowMilliseconds)
	require.InDelta(t, 1000, summary.ThroughputPerSecond, 0.001)
	require.InDelta(t, 1, summary.LatencyMilliseconds.P50, 0.001)
	require.InDelta(t, 1, summary.LatencyMilliseconds.P95, 0.001)
	require.InDelta(t, 1, summary.LatencyMilliseconds.P99, 0.001)
}

func TestSummarizeLoadQueriesRejectsImpossibleTiming(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	_, err := SummarizeLoadQueries([]LoadQuerySample{{
		StartedAt:  now,
		FinishedAt: now.Add(-time.Millisecond),
	}})
	require.ErrorContains(t, err, "finishes before it starts")
}

func TestSummarizeLoadTransactionsCapturesLifecycleCountsAndGas(t *testing.T) {
	t.Parallel()

	started := time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC)
	samples := []LoadTxSample{
		{SubmittedAt: started, FinishedAt: started.Add(10 * time.Millisecond), CheckTxAccepted: true, Committed: true, GasWanted: 10, GasUsed: 8},
		{SubmittedAt: started.Add(time.Millisecond), FinishedAt: started.Add(21 * time.Millisecond), CheckTxAccepted: true, Committed: true, GasWanted: 20, GasUsed: 17},
		{SubmittedAt: started.Add(2 * time.Millisecond), FinishedAt: started.Add(32 * time.Millisecond), Failed: true},
	}

	summary, err := SummarizeLoadTransactions(samples)
	require.NoError(t, err)
	require.Equal(t, 3, summary.Submitted)
	require.Equal(t, 2, summary.CheckTxAccepted)
	require.Equal(t, 2, summary.Committed)
	require.Equal(t, 1, summary.Failed)
	require.EqualValues(t, 30, summary.GasWanted)
	require.EqualValues(t, 25, summary.GasUsed)
	require.InDelta(t, 10, summary.CommitLatencyMilliseconds.P50, 0.001)
	require.InDelta(t, 20, summary.CommitLatencyMilliseconds.P95, 0.001)
	require.InDelta(t, 20, summary.CommitLatencyMilliseconds.P99, 0.001)
}

func TestSummarizeLoadBlocksCapturesIntervalsGasAndMissedSignatures(t *testing.T) {
	t.Parallel()

	samples := []LoadBlockSample{
		{Height: 10, IntervalMilliseconds: 1000, Transactions: 2, FailedTransactions: 1, GasWanted: 20, GasUsed: 15},
		{Height: 11, IntervalMilliseconds: 1500, Transactions: 3, GasWanted: 30, GasUsed: 25, MissedSignatures: 1},
	}
	summary := SummarizeLoadBlocks(samples)
	require.Equal(t, 2, summary.Blocks)
	require.Equal(t, 5, summary.Transactions)
	require.Equal(t, 1, summary.FailedTransactions)
	require.EqualValues(t, 50, summary.GasWanted)
	require.EqualValues(t, 40, summary.GasUsed)
	require.Equal(t, 1, summary.MissedSignatures)
	require.InDelta(t, 1000, summary.IntervalMilliseconds.P50, 0.001)
	require.InDelta(t, 1500, summary.IntervalMilliseconds.P95, 0.001)
}

func TestDecodeTxGasAcceptsStringAndNumericJSON(t *testing.T) {
	t.Parallel()

	wanted, used, err := DecodeTxGas(json.RawMessage(`{"gas_wanted":"500000","gas_used":12345}`))
	require.NoError(t, err)
	require.EqualValues(t, 500000, wanted)
	require.EqualValues(t, 12345, used)

	_, _, err = DecodeTxGas(json.RawMessage(`{"height":"7"}`))
	require.ErrorContains(t, err, "gas_wanted")
}

func TestRewriteQueryGasLimitPreservesTheRestOfAppConfig(t *testing.T) {
	t.Parallel()

	input := []byte("minimum-gas-prices = \"0umed\"\nquery-gas-limit = 10000000 # bounded queries\n[api]\nenable = true\n")
	original := append([]byte(nil), input...)
	rewritten, previous, err := RewriteQueryGasLimit(input, 1)
	require.NoError(t, err)
	require.EqualValues(t, 10_000_000, previous)
	require.Equal(t, "minimum-gas-prices = \"0umed\"\nquery-gas-limit = 1 # bounded queries\n[api]\nenable = true\n", string(rewritten))
	require.Equal(t, original, input, "input must remain immutable")
}

func TestRewriteQueryGasLimitRequiresExactlyOneNumericSetting(t *testing.T) {
	t.Parallel()

	_, _, err := RewriteQueryGasLimit([]byte("[api]\nenable = true\n"), 1)
	require.ErrorContains(t, err, "exactly one")

	_, _, err = RewriteQueryGasLimit([]byte("query-gas-limit = 10\nquery-gas-limit = 20\n"), 1)
	require.ErrorContains(t, err, "exactly one")

	_, _, err = RewriteQueryGasLimit([]byte("query-gas-limit = \"unlimited\"\n"), 1)
	require.ErrorContains(t, err, "numeric")
}
