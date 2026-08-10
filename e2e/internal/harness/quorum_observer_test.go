package harness

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestVerifyCommonCommitmentAcceptsSameHeightHashes(t *testing.T) {
	t.Parallel()

	agreement, err := VerifyCommonCommitment(42, []QuorumCommitment{
		{Node: "validator-0", Height: 42, BlockHash: "AABB", AppHash: "CCDD"},
		{Node: "validator-1", Height: 42, BlockHash: "AABB", AppHash: "CCDD"},
		{Node: "fullnode-0", Height: 42, BlockHash: "AABB", AppHash: "CCDD"},
	})

	require.NoError(t, err)
	require.Equal(t, int64(42), agreement.Height)
	require.Equal(t, "AABB", agreement.BlockHash)
	require.Equal(t, "CCDD", agreement.AppHash)
	require.Equal(t, []string{"validator-0", "validator-1", "fullnode-0"}, agreement.Nodes)
}

func TestVerifyCommonCommitmentRejectsDuplicateNodeEvidence(t *testing.T) {
	t.Parallel()

	_, err := VerifyCommonCommitment(42, []QuorumCommitment{
		{Node: "validator-0", Height: 42, BlockHash: "AABB", AppHash: "CCDD"},
		{Node: "validator-0", Height: 42, BlockHash: "AABB", AppHash: "CCDD"},
	})

	require.ErrorContains(t, err, "duplicate")
}

func TestVerifyCommonCommitmentRejectsDifferentApplicationState(t *testing.T) {
	t.Parallel()

	_, err := VerifyCommonCommitment(42, []QuorumCommitment{
		{Node: "validator-0", Height: 42, BlockHash: "AABB", AppHash: "CCDD"},
		{Node: "fullnode-0", Height: 42, BlockHash: "AABB", AppHash: "EEFF"},
	})

	require.ErrorContains(t, err, "differs")
}

func TestQuorumObserverWaitForProgressRejectsHeightOverflow(t *testing.T) {
	t.Parallel()

	observer, err := NewQuorumObserver(time.Millisecond)
	require.NoError(t, err)

	_, err = observer.WaitForProgress(
		context.Background(),
		math.MaxInt64,
		1,
		func(context.Context) (int64, error) { return math.MaxInt64, nil },
	)

	require.ErrorContains(t, err, "overflows")
}

func TestQuorumObserverWaitForProgressReturnsAfterMinimumBlocks(t *testing.T) {
	t.Parallel()

	observer, err := NewQuorumObserver(time.Millisecond)
	require.NoError(t, err)
	heights := []int64{10, 11, 13}
	next := 0
	readHeight := func(context.Context) (int64, error) {
		height := heights[next]
		if next < len(heights)-1 {
			next++
		}
		return height, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	window, err := observer.WaitForProgress(ctx, 10, 3, readHeight)

	require.NoError(t, err)
	require.Equal(t, int64(10), window.StartHeight)
	require.Equal(t, int64(13), window.EndHeight)
	require.Equal(t, int64(13), window.TargetHeight)
	require.GreaterOrEqual(t, len(window.Samples), 3)
}

func TestQuorumObserverObserveStallAcceptsStableHeightForWholeWindow(t *testing.T) {
	t.Parallel()

	observer, err := NewQuorumObserver(time.Millisecond)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	window, err := observer.ObserveStall(
		ctx,
		3*time.Millisecond,
		5*time.Millisecond,
		func(context.Context) (int64, error) { return 21, nil },
	)

	require.NoError(t, err)
	require.Equal(t, int64(21), window.StartHeight)
	require.Equal(t, int64(21), window.EndHeight)
	require.GreaterOrEqual(t, len(window.Samples), 2)
}

func TestQuorumObserverObserveStallRejectsCommittedHeightAdvance(t *testing.T) {
	t.Parallel()

	observer, err := NewQuorumObserver(100 * time.Microsecond)
	require.NoError(t, err)
	reads := 0
	readHeight := func(context.Context) (int64, error) {
		reads++
		if reads <= 2 {
			return 21, nil
		}
		return 22, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	window, err := observer.ObserveStall(ctx, time.Nanosecond, 20*time.Millisecond, readHeight)

	require.ErrorContains(t, err, "advanced")
	require.Equal(t, int64(21), window.StartHeight)
	require.Equal(t, int64(22), window.EndHeight)
}
