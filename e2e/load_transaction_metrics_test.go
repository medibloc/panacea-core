package e2e_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestLoadTxSampleKeepsCheckTxSeparateFromFinalizeBlock(t *testing.T) {
	started := time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC)
	finished := started.Add(time.Second)
	lifecycle := &harness.TxLifecycleResult{
		CheckTx: &harness.TxResult{Height: "0", TxHash: "ABC", Code: 0},
		Committed: &harness.TxResult{
			Height: "17",
			TxHash: "ABC",
			Code:   9,
			Raw:    json.RawMessage(`{"gas_wanted":"10","gas_used":"8"}`),
		},
	}
	deliverErr := errors.New("FinalizeBlock rejected transaction")

	sample, err := loadTxSampleFromLifecycle("mint", started, finished, lifecycle, deliverErr)
	require.ErrorIs(t, err, deliverErr)
	require.True(t, sample.CheckTxAccepted)
	require.True(t, sample.Committed)
	require.True(t, sample.Failed)
	require.Equal(t, "ABC", sample.TxHash)
	require.Equal(t, int64(17), sample.Height)
	require.Equal(t, int64(10), sample.GasWanted)
	require.Equal(t, int64(8), sample.GasUsed)

	summary, summaryErr := harness.SummarizeLoadTransactions([]harness.LoadTxSample{sample})
	require.NoError(t, summaryErr)
	require.Equal(t, 1, summary.Submitted)
	require.Equal(t, 1, summary.CheckTxAccepted)
	require.Equal(t, 1, summary.Committed)
	require.Equal(t, 1, summary.Failed)
}

func TestLoadTxSampleKeepsAcceptedButUnobservedTransactionSeparate(t *testing.T) {
	started := time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC)
	finished := started.Add(time.Second)
	commitErr := context.DeadlineExceeded
	lifecycle := &harness.TxLifecycleResult{
		CheckTx: &harness.TxResult{Height: "0", TxHash: "TIMEOUT", Code: 0},
	}

	sample, err := loadTxSampleFromLifecycle("mint", started, finished, lifecycle, commitErr)
	require.ErrorIs(t, err, commitErr)
	require.True(t, sample.CheckTxAccepted)
	require.False(t, sample.Committed)
	require.True(t, sample.Failed)
	require.Empty(t, sample.TxHash)
	require.Zero(t, sample.Height)

	summary, summaryErr := harness.SummarizeLoadTransactions([]harness.LoadTxSample{sample})
	require.NoError(t, summaryErr)
	require.Equal(t, 1, summary.Submitted)
	require.Equal(t, 1, summary.CheckTxAccepted)
	require.Zero(t, summary.Committed)
	require.Equal(t, 1, summary.Failed)
}

func TestLoadTxSampleDoesNotAcceptMissingCheckTxEvidence(t *testing.T) {
	started := time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC)
	lifecycle := &harness.TxLifecycleResult{
		Committed: &harness.TxResult{
			Height: "17",
			TxHash: "ABC",
			Code:   0,
			Raw:    json.RawMessage(`{"gas_wanted":"10","gas_used":"8"}`),
		},
	}

	sample, err := loadTxSampleFromLifecycle("mint", started, started.Add(time.Second), lifecycle, nil)

	require.NoError(t, err)
	require.False(t, sample.CheckTxAccepted)
	require.True(t, sample.Committed)
	require.False(t, sample.Failed)
}
