package harness

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"
)

func TestWaitForCometStateSyncProvidersProbesInParallelWithBoundedContexts(t *testing.T) {
	network, store := newCometStateSyncProviderTestNetwork(t, "run-111111111111")
	chain := newQuorumTestChain()
	sources := []*cosmos.ChainNode{
		newQuorumTestNode(chain, 0, 1, 30),
		newQuorumTestNode(chain, 1, 1, 30),
	}
	sourceIndexes := map[*cosmos.ChainNode]int{sources[0]: 0, sources[1]: 1}

	const probeTimeout = 40 * time.Millisecond
	var mu sync.Mutex
	calls := make([]int, len(sources))
	firstActive := 0
	firstActiveMax := 0
	firstStarted := 0
	allFirstStarted := make(chan struct{})
	var closeAllFirstStarted sync.Once
	probeDeadlineRemaining := make([]time.Duration, 0, 4)
	missingProbeDeadline := false

	readLogs := func(ctx context.Context, source *cosmos.ChainNode, _ time.Time) ([]byte, error) {
		index := sourceIndexes[source]
		deadline, hasDeadline := ctx.Deadline()

		mu.Lock()
		calls[index]++
		attempt := calls[index]
		if hasDeadline {
			probeDeadlineRemaining = append(probeDeadlineRemaining, time.Until(deadline))
		} else {
			missingProbeDeadline = true
		}
		if attempt == 1 {
			firstActive++
			if firstActive > firstActiveMax {
				firstActiveMax = firstActive
			}
			firstStarted++
			if firstStarted == len(sources) {
				closeAllFirstStarted.Do(func() { close(allFirstStarted) })
			}
			mu.Unlock()

			select {
			case <-allFirstStarted:
			case <-ctx.Done():
			}

			mu.Lock()
			firstActive--
			mu.Unlock()
			return nil, fmt.Errorf("transient provider %d log error", index)
		}
		mu.Unlock()

		return []byte("INF completed state snapshot height=30 format=3\n"), nil
	}

	evidence, err := network.waitForCometStateSyncProvidersWithHooks(
		context.Background(),
		sources,
		time.Date(2026, 8, 5, 1, 2, 3, 0, time.UTC),
		time.Second,
		"recovery/state-sync/provider",
		cometStateSyncProviderPollHooks{
			readLogs:     readLogs,
			probeTimeout: probeTimeout,
			pollInterval: time.Millisecond,
		},
	)
	require.NoError(t, err)
	require.Len(t, evidence, len(sources))

	mu.Lock()
	finalCalls := append([]int(nil), calls...)
	finalFirstActiveMax := firstActiveMax
	finalProbeDeadlines := append([]time.Duration(nil), probeDeadlineRemaining...)
	finalMissingProbeDeadline := missingProbeDeadline
	mu.Unlock()
	require.Equal(t, []int{2, 2}, finalCalls)
	require.Equal(t, len(sources), finalFirstActiveMax)
	require.False(t, finalMissingProbeDeadline)
	require.Len(t, finalProbeDeadlines, 4)
	for _, remaining := range finalProbeDeadlines {
		require.Positive(t, remaining)
		require.LessOrEqual(t, remaining, probeTimeout+10*time.Millisecond)
	}
	for _, provider := range evidence {
		require.Equal(t, []int64{30}, provider.CompletedSnapshotHeights)
		require.Empty(t, provider.LogError)
	}
	for index := range sources {
		raw, readErr := os.ReadFile(filepath.Join(store.dir, "recovery", "state-sync", fmt.Sprintf("provider-%d.log", index)))
		require.NoError(t, readErr)
		require.Contains(t, string(raw), "completed state snapshot height=30")
	}
}

func TestWaitForCometStateSyncProvidersDoesNotReprobeSatisfiedProviders(t *testing.T) {
	network, _ := newCometStateSyncProviderTestNetwork(t, "run-222222222222")
	chain := newQuorumTestChain()
	sources := []*cosmos.ChainNode{
		newQuorumTestNode(chain, 0, 1, 30),
		newQuorumTestNode(chain, 1, 1, 30),
	}
	sourceIndexes := map[*cosmos.ChainNode]int{sources[0]: 0, sources[1]: 1}

	var mu sync.Mutex
	calls := make([]int, len(sources))
	readLogs := func(_ context.Context, source *cosmos.ChainNode, _ time.Time) ([]byte, error) {
		index := sourceIndexes[source]
		mu.Lock()
		calls[index]++
		attempt := calls[index]
		mu.Unlock()

		if index == 1 && attempt == 1 {
			return nil, errors.New("provider 1 is temporarily unavailable")
		}
		return []byte(fmt.Sprintf("INF completed state snapshot height=%d format=3\n", 30+index)), nil
	}

	evidence, err := network.waitForCometStateSyncProvidersWithHooks(
		context.Background(),
		sources,
		time.Date(2026, 8, 5, 2, 3, 4, 0, time.UTC),
		time.Second,
		"recovery/state-sync/cache-provider",
		cometStateSyncProviderPollHooks{
			readLogs:     readLogs,
			probeTimeout: 50 * time.Millisecond,
			pollInterval: time.Millisecond,
		},
	)
	require.NoError(t, err)

	mu.Lock()
	finalCalls := append([]int(nil), calls...)
	mu.Unlock()
	require.Equal(t, []int{1, 2}, finalCalls)
	require.Equal(t, []int64{30}, evidence[0].CompletedSnapshotHeights)
	require.Equal(t, []int64{31}, evidence[1].CompletedSnapshotHeights)
	require.Empty(t, evidence[0].LogError)
	require.Empty(t, evidence[1].LogError)
}

func TestWaitForCometStateSyncProvidersPreservesLastErrorsAndOverallDeadline(t *testing.T) {
	network, _ := newCometStateSyncProviderTestNetwork(t, "run-333333333333")
	chain := newQuorumTestChain()
	sources := []*cosmos.ChainNode{
		newQuorumTestNode(chain, 0, 1, 30),
		newQuorumTestNode(chain, 1, 1, 30),
	}
	sourceIndexes := map[*cosmos.ChainNode]int{sources[0]: 0, sources[1]: 1}

	var calls [2]atomic.Int32
	releaseBlockedProbes := make(chan struct{})
	allSecondProbesStarted := make(chan struct{})
	var closeAllSecondProbesStarted sync.Once
	readLogs := func(_ context.Context, source *cosmos.ChainNode, _ time.Time) ([]byte, error) {
		index := sourceIndexes[source]
		attempt := calls[index].Add(1)
		if attempt == 1 {
			return nil, fmt.Errorf("provider-%d first probe failure", index)
		}
		if calls[0].Load() >= 2 && calls[1].Load() >= 2 {
			closeAllSecondProbesStarted.Do(func() { close(allSecondProbesStarted) })
		}
		<-releaseBlockedProbes
		return nil, fmt.Errorf("provider-%d blocked probe released", index)
	}

	type waitResult struct {
		evidence []CometStateSyncProviderEvidence
		err      error
	}
	resultCh := make(chan waitResult, 1)
	go func() {
		evidence, err := network.waitForCometStateSyncProvidersWithHooks(
			context.Background(),
			sources,
			time.Date(2026, 8, 5, 3, 4, 5, 0, time.UTC),
			500*time.Millisecond,
			"recovery/state-sync/deadline-provider",
			cometStateSyncProviderPollHooks{
				readLogs:     readLogs,
				probeTimeout: 20 * time.Millisecond,
				pollInterval: time.Millisecond,
			},
		)
		resultCh <- waitResult{evidence: evidence, err: err}
	}()

	select {
	case <-allSecondProbesStarted:
	case <-time.After(250 * time.Millisecond):
		close(releaseBlockedProbes)
		<-resultCh
		t.Fatal("provider polling did not start a second concurrent probe round")
	}

	var result waitResult
	select {
	case result = <-resultCh:
		close(releaseBlockedProbes)
	case <-time.After(time.Second):
		close(releaseBlockedProbes)
		<-resultCh
		t.Fatal("provider polling did not preserve its overall deadline")
	}

	require.Error(t, result.err)
	require.ErrorIs(t, result.err, context.DeadlineExceeded)
	require.Contains(t, result.err.Error(), sources[0].Name())
	require.Contains(t, result.err.Error(), sources[1].Name())
	require.Contains(t, result.err.Error(), "provider-0 first probe failure")
	require.Contains(t, result.err.Error(), "provider-1 first probe failure")
	require.Equal(t, int32(2), calls[0].Load())
	require.Equal(t, int32(2), calls[1].Load())
	require.Len(t, result.evidence, len(sources))
	require.True(t, strings.Contains(result.evidence[0].LogError, "provider-0 first probe failure"))
	require.True(t, strings.Contains(result.evidence[1].LogError, "provider-1 first probe failure"))
}

func newCometStateSyncProviderTestNetwork(t *testing.T, runID string) (*Network, *artifactStore) {
	t.Helper()
	store, err := newArtifactStore(
		"comet-state-sync-provider",
		runID,
		artifactTestConfig(filepath.Join(trustedArtifactTempDir(t), "artifacts")),
	)
	require.NoError(t, err)
	return &Network{artifacts: store}, store
}
