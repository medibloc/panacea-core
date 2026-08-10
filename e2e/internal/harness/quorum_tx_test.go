package harness

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestVerifyQuorumNoCommitEvidenceAcceptsReachableRecentNotFound(t *testing.T) {
	t.Parallel()

	observedUntil := time.Now().UTC()
	err := VerifyQuorumNoCommitEvidence([]QuorumTxLookupSample{
		{
			ObservedAt:     observedUntil.Add(-100 * time.Millisecond),
			FullNodeHeight: 42,
			LookupError:    "tx (AABB) not found",
		},
	}, observedUntil, time.Second)

	require.NoError(t, err)
}

func TestVerifyQuorumNoCommitEvidenceRejectsCommittedTransaction(t *testing.T) {
	t.Parallel()

	observedUntil := time.Now().UTC()
	err := VerifyQuorumNoCommitEvidence([]QuorumTxLookupSample{
		{
			ObservedAt:     observedUntil,
			FullNodeHeight: 43,
			Committed:      true,
			CommitHeight:   43,
		},
	}, observedUntil, time.Second)

	require.ErrorContains(t, err, "committed")
}

func TestVerifyQuorumNoCommitEvidenceRejectsGenericLookupFailure(t *testing.T) {
	t.Parallel()

	observedUntil := time.Now().UTC()
	err := VerifyQuorumNoCommitEvidence([]QuorumTxLookupSample{
		{
			ObservedAt:     observedUntil,
			FullNodeHeight: 42,
			LookupError:    "connection reset by peer",
		},
	}, observedUntil, time.Second)

	require.ErrorContains(t, err, "never returned")
}

func TestVerifyQuorumNoCommitEvidenceRejectsNonTransactionNotFound(t *testing.T) {
	t.Parallel()

	observedUntil := time.Now().UTC()
	err := VerifyQuorumNoCommitEvidence([]QuorumTxLookupSample{
		{
			ObservedAt:     observedUntil,
			FullNodeHeight: 42,
			LookupError:    "RPC endpoint not found",
		},
	}, observedUntil, time.Second)

	require.ErrorContains(t, err, "never returned")
}
