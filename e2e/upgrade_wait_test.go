package e2e_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUpgradeHaltWaitBudgetCoversSlowChaosFixture(t *testing.T) {
	const (
		chaosUpgradeLeadBlocks = int64(55)
		chaosTimeoutCommit     = 2 * time.Second
	)

	minimumBudget := time.Duration(chaosUpgradeLeadBlocks)*chaosTimeoutCommit*2 +
		upgradeHaltStabilityWindow + upgradeHaltProbeTimeout
	require.GreaterOrEqual(t, upgradeHaltWaitTimeout, minimumBudget)
}

func TestUpgradeHaltTrackerAcceptsHealthyRunningSDKHalt(t *testing.T) {
	const (
		planName      = "v2.3.0"
		upgradeHeight = int64(55)
	)
	window := 5 * time.Second
	tracker, err := newUpgradeHaltTracker(planName, upgradeHeight, window)
	require.NoError(t, err)

	observations := validUpgradeHaltObservations(planName, upgradeHeight)
	startedAt := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	complete, err := tracker.Observe(startedAt, observations, nil)
	require.NoError(t, err)
	require.False(t, complete)

	complete, err = tracker.Observe(startedAt.Add(window-time.Nanosecond), observations, nil)
	require.NoError(t, err)
	require.False(t, complete)

	complete, err = tracker.Observe(startedAt.Add(window), observations, nil)
	require.NoError(t, err)
	require.True(t, complete)
	require.Equal(t, observations, tracker.Evidence())
}

func TestUpgradeHaltTrackerRejectsConsensusAdvancePastBoundary(t *testing.T) {
	tracker, err := newUpgradeHaltTracker("v2.3.0", 55, 5*time.Second)
	require.NoError(t, err)
	observations := validUpgradeHaltObservations("v2.3.0", 55)
	advanced := observations["validator-0"]
	advanced.ConsensusHeight = 56
	observations[advanced.Node] = advanced

	complete, err := tracker.Observe(time.Now(), observations, nil)

	require.False(t, complete)
	require.ErrorContains(t, err, "validator-0 consensus height 56 advanced past upgrade height 55")
}

func TestUpgradeHaltTrackerRequiresExactPreUpgradeApplicationHeight(t *testing.T) {
	for _, applicationHeight := range []int64{53, 55} {
		t.Run(time.Unix(applicationHeight, 0).UTC().Format(time.RFC3339), func(t *testing.T) {
			tracker, err := newUpgradeHaltTracker("v2.3.0", 55, 5*time.Second)
			require.NoError(t, err)
			observations := validUpgradeHaltObservations("v2.3.0", 55)
			invalid := observations["validator-0"]
			invalid.ApplicationHeight = applicationHeight
			observations[invalid.Node] = invalid

			complete, err := tracker.Observe(time.Now(), observations, nil)

			require.False(t, complete)
			require.ErrorContains(t, err, "validator-0 application height")
		})
	}
}

func TestUpgradeHaltTrackerRequiresExactUpgradePlan(t *testing.T) {
	for _, test := range []struct {
		name       string
		planName   string
		planHeight int64
	}{
		{name: "wrong name", planName: "v2.4.0", planHeight: 55},
		{name: "wrong height", planName: "v2.3.0", planHeight: 56},
	} {
		t.Run(test.name, func(t *testing.T) {
			tracker, err := newUpgradeHaltTracker("v2.3.0", 55, 5*time.Second)
			require.NoError(t, err)
			observations := validUpgradeHaltObservations("v2.3.0", 55)
			invalid := observations["validator-0"]
			invalid.PlanName = test.planName
			invalid.PlanHeight = test.planHeight
			observations[invalid.Node] = invalid

			complete, err := tracker.Observe(time.Now(), observations, nil)

			require.False(t, complete)
			require.ErrorContains(t, err, "validator-0 upgrade plan")
			require.ErrorContains(t, err, `want "v2.3.0" at 55`)
		})
	}
}

func TestUpgradeHaltTrackerRejectsUnhealthyContainerLifecycle(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*upgradeHaltObservation)
	}{
		{name: "not running", mutate: func(observation *upgradeHaltObservation) {
			observation.ContainerRunning = false
			observation.ContainerStatus = "exited"
		}},
		{name: "paused", mutate: func(observation *upgradeHaltObservation) {
			observation.ContainerPaused = true
		}},
		{name: "wrong status", mutate: func(observation *upgradeHaltObservation) {
			observation.ContainerStatus = "restarting"
		}},
		{name: "restarting", mutate: func(observation *upgradeHaltObservation) {
			observation.ContainerRestarting = true
		}},
		{name: "oom killed", mutate: func(observation *upgradeHaltObservation) {
			observation.ContainerOOMKilled = true
		}},
		{name: "dead", mutate: func(observation *upgradeHaltObservation) {
			observation.ContainerDead = true
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			tracker, err := newUpgradeHaltTracker("v2.3.0", 55, 5*time.Second)
			require.NoError(t, err)
			observations := validUpgradeHaltObservations("v2.3.0", 55)
			invalid := observations["validator-0"]
			test.mutate(&invalid)
			observations[invalid.Node] = invalid

			complete, err := tracker.Observe(time.Now(), observations, nil)

			require.False(t, complete)
			require.ErrorContains(t, err, "validator-0 has unhealthy Docker lifecycle")
		})
	}
}

func TestUpgradeHaltTrackerRejectsEmptyApplicationHash(t *testing.T) {
	tracker, err := newUpgradeHaltTracker("v2.3.0", 55, 5*time.Second)
	require.NoError(t, err)
	observations := validUpgradeHaltObservations("v2.3.0", 55)
	invalid := observations["validator-0"]
	invalid.ApplicationHash = "  "
	observations[invalid.Node] = invalid

	complete, err := tracker.Observe(time.Now(), observations, nil)

	require.False(t, complete)
	require.ErrorContains(t, err, "validator-0 application hash is empty")
}

func TestUpgradeHaltTrackerRejectsDivergentApplicationHash(t *testing.T) {
	tracker, err := newUpgradeHaltTracker("v2.3.0", 55, 5*time.Second)
	require.NoError(t, err)
	observations := validUpgradeHaltObservations("v2.3.0", 55)
	invalid := observations["validator-0"]
	invalid.ApplicationHash = "DEADBEEF"
	observations[invalid.Node] = invalid

	complete, err := tracker.Observe(time.Now(), observations, nil)

	require.False(t, complete)
	require.ErrorContains(t, err, "application hash")
	require.ErrorContains(t, err, "differs")
}

func TestUpgradeHaltTrackerRequiresBlockHeaderToCommitApplicationState(t *testing.T) {
	tracker, err := newUpgradeHaltTracker("v2.3.0", 55, 5*time.Second)
	require.NoError(t, err)
	observations := validUpgradeHaltObservations("v2.3.0", 55)
	invalid := observations["validator-0"]
	invalid.ConsensusAppHash = "DEADBEEF"
	observations[invalid.Node] = invalid

	complete, err := tracker.Observe(time.Now(), observations, nil)

	require.False(t, complete)
	require.ErrorContains(t, err, "block 55 app hash DEADBEEF")
	require.ErrorContains(t, err, "application hash AABBCCDD at height 54")
}

func TestUpgradeHaltTrackerRejectsDivergentBoundaryBlock(t *testing.T) {
	tracker, err := newUpgradeHaltTracker("v2.3.0", 55, 5*time.Second)
	require.NoError(t, err)
	observations := validUpgradeHaltObservations("v2.3.0", 55)
	invalid := observations["validator-0"]
	invalid.ConsensusBlockID = "DEADBEEF"
	observations[invalid.Node] = invalid

	complete, err := tracker.Observe(time.Now(), observations, nil)

	require.False(t, complete)
	require.ErrorContains(t, err, "boundary block/app commitment")
	require.ErrorContains(t, err, "differs")
}

func TestUpgradeHaltTrackerRejectsBoundaryCommitmentChangeDuringWindow(t *testing.T) {
	window := 5 * time.Second
	tracker, err := newUpgradeHaltTracker("v2.3.0", 55, window)
	require.NoError(t, err)
	observations := validUpgradeHaltObservations("v2.3.0", 55)
	startedAt := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	complete, err := tracker.Observe(startedAt, observations, nil)
	require.NoError(t, err)
	require.False(t, complete)

	changed := cloneUpgradeHaltObservations(observations)
	for node, observation := range changed {
		observation.ConsensusBlockID = "55667788"
		changed[node] = observation
	}
	complete, err = tracker.Observe(startedAt.Add(window), changed, nil)

	require.False(t, complete)
	require.ErrorContains(t, err, "upgrade-boundary commitment changed during the stability window")
	require.Equal(t, observations, tracker.Evidence())
}

func TestUpgradeHaltTrackerReadGapResetsStabilityAndPreservesLastGoodEvidence(t *testing.T) {
	const upgradeHeight = int64(55)
	window := 5 * time.Second
	tracker, err := newUpgradeHaltTracker("v2.3.0", upgradeHeight, window)
	require.NoError(t, err)
	observations := validUpgradeHaltObservations("v2.3.0", upgradeHeight)
	startedAt := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)

	complete, err := tracker.Observe(startedAt, observations, nil)
	require.NoError(t, err)
	require.False(t, complete)

	complete, err = tracker.Observe(startedAt.Add(4*time.Second), nil, errors.New("validator-2 RPC unavailable"))
	require.NoError(t, err)
	require.False(t, complete)
	require.Equal(t, observations, tracker.Evidence())

	complete, err = tracker.Observe(startedAt.Add(window), observations, nil)
	require.NoError(t, err)
	require.False(t, complete, "a read gap must restart the uninterrupted window")

	complete, err = tracker.Observe(startedAt.Add(2*window), observations, nil)
	require.NoError(t, err)
	require.True(t, complete)
}

func TestUpgradeHaltTrackerPreBoundaryObservationResetsStability(t *testing.T) {
	window := 5 * time.Second
	tracker, err := newUpgradeHaltTracker("v2.3.0", 55, window)
	require.NoError(t, err)
	valid := validUpgradeHaltObservations("v2.3.0", 55)
	startedAt := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	complete, err := tracker.Observe(startedAt, valid, nil)
	require.NoError(t, err)
	require.False(t, complete)

	preBoundary := cloneUpgradeHaltObservations(valid)
	observation := preBoundary["validator-0"]
	observation.ConsensusHeight = 54
	preBoundary[observation.Node] = observation
	complete, err = tracker.Observe(startedAt.Add(4*time.Second), preBoundary, nil)
	require.NoError(t, err)
	require.False(t, complete)
	require.Equal(t, valid, tracker.Evidence())

	complete, err = tracker.Observe(startedAt.Add(window), valid, nil)
	require.NoError(t, err)
	require.False(t, complete, "a pre-boundary observation must restart the uninterrupted window")
	complete, err = tracker.Observe(startedAt.Add(2*window), valid, nil)
	require.NoError(t, err)
	require.True(t, complete)
}

func TestUpgradeHaltTrackerEarlierConsensusAndApplicationStateResetStability(t *testing.T) {
	window := 5 * time.Second
	tracker, err := newUpgradeHaltTracker("v2.3.0", 55, window)
	require.NoError(t, err)
	valid := validUpgradeHaltObservations("v2.3.0", 55)
	startedAt := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	complete, err := tracker.Observe(startedAt, valid, nil)
	require.NoError(t, err)
	require.False(t, complete)

	earlier := cloneUpgradeHaltObservations(valid)
	for node, observation := range earlier {
		observation.ConsensusHeight = 53
		observation.ApplicationHeight = 53
		earlier[node] = observation
	}
	complete, err = tracker.Observe(startedAt.Add(4*time.Second), earlier, nil)
	require.NoError(t, err)
	require.False(t, complete)
	require.Equal(t, valid, tracker.Evidence())

	complete, err = tracker.Observe(startedAt.Add(window), valid, nil)
	require.NoError(t, err)
	require.False(t, complete)
	complete, err = tracker.Observe(startedAt.Add(2*window), valid, nil)
	require.NoError(t, err)
	require.True(t, complete)
}

func TestUpgradeHaltTrackerRequiresObservationKeyToMatchNode(t *testing.T) {
	tracker, err := newUpgradeHaltTracker("v2.3.0", 55, 5*time.Second)
	require.NoError(t, err)
	observations := validUpgradeHaltObservations("v2.3.0", 55)
	observation := observations["validator-0"]
	delete(observations, "validator-0")
	observations["wrong-map-key"] = observation

	complete, err := tracker.Observe(time.Now(), observations, nil)

	require.False(t, complete)
	require.ErrorContains(t, err, `observation key "wrong-map-key" does not match node "validator-0"`)
}

func TestUpgradeHaltTrackerNodeSetGapResetsStability(t *testing.T) {
	window := 5 * time.Second
	tracker, err := newUpgradeHaltTracker("v2.3.0", 55, window)
	require.NoError(t, err)
	valid := validUpgradeHaltObservations("v2.3.0", 55)
	startedAt := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	complete, err := tracker.Observe(startedAt, valid, nil)
	require.NoError(t, err)
	require.False(t, complete)

	incomplete := cloneUpgradeHaltObservations(valid)
	delete(incomplete, "validator-2")
	complete, err = tracker.Observe(startedAt.Add(window), incomplete, nil)
	require.NoError(t, err)
	require.False(t, complete)
	require.Equal(t, valid, tracker.Evidence())

	complete, err = tracker.Observe(startedAt.Add(window), valid, nil)
	require.NoError(t, err)
	require.False(t, complete)
	complete, err = tracker.Observe(startedAt.Add(2*window), valid, nil)
	require.NoError(t, err)
	require.True(t, complete)
}

func TestUpgradeHaltTrackerKeepsNodeSetEstablishedBeforeBoundary(t *testing.T) {
	window := 5 * time.Second
	tracker, err := newUpgradeHaltTracker("v2.3.0", 55, window)
	require.NoError(t, err)
	startedAt := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	earlier := validUpgradeHaltObservations("v2.3.0", 55)
	for node, observation := range earlier {
		observation.ConsensusHeight = 53
		observation.ApplicationHeight = 53
		earlier[node] = observation
	}
	complete, err := tracker.Observe(startedAt, earlier, nil)
	require.NoError(t, err)
	require.False(t, complete)

	incomplete := validUpgradeHaltObservations("v2.3.0", 55)
	delete(incomplete, "validator-2")
	complete, err = tracker.Observe(startedAt.Add(time.Second), incomplete, nil)
	require.NoError(t, err)
	require.False(t, complete)
	complete, err = tracker.Observe(startedAt.Add(time.Second+window), incomplete, nil)
	require.NoError(t, err)
	require.False(t, complete, "a missing node must not become the expected set after pre-boundary observations")
	require.Nil(t, tracker.Evidence())
}

func TestUpgradeHaltTrackerRejectsBackwardObservationTime(t *testing.T) {
	tracker, err := newUpgradeHaltTracker("v2.3.0", 55, 5*time.Second)
	require.NoError(t, err)
	observations := validUpgradeHaltObservations("v2.3.0", 55)
	startedAt := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	complete, err := tracker.Observe(startedAt, observations, nil)
	require.NoError(t, err)
	require.False(t, complete)

	complete, err = tracker.Observe(startedAt.Add(-time.Nanosecond), observations, nil)

	require.False(t, complete)
	require.ErrorContains(t, err, "upgrade halt observation time moved backward")
	require.Equal(t, observations, tracker.Evidence())
}

func TestUpgradeHaltTrackerSemanticErrorBreaksUninterruptedWindow(t *testing.T) {
	window := 5 * time.Second
	tracker, err := newUpgradeHaltTracker("v2.3.0", 55, window)
	require.NoError(t, err)
	valid := validUpgradeHaltObservations("v2.3.0", 55)
	startedAt := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	complete, err := tracker.Observe(startedAt, valid, nil)
	require.NoError(t, err)
	require.False(t, complete)

	wrongPlan := cloneUpgradeHaltObservations(valid)
	observation := wrongPlan["validator-0"]
	observation.PlanName = "v2.4.0"
	wrongPlan[observation.Node] = observation
	complete, err = tracker.Observe(startedAt.Add(4*time.Second), wrongPlan, nil)
	require.Error(t, err)
	require.False(t, complete)
	require.Equal(t, valid, tracker.Evidence())

	complete, err = tracker.Observe(startedAt.Add(window), valid, nil)
	require.NoError(t, err)
	require.False(t, complete, "a semantic error must break the uninterrupted window")
	complete, err = tracker.Observe(startedAt.Add(2*window), valid, nil)
	require.NoError(t, err)
	require.True(t, complete)
}

func TestUpgradeHaltTrackerEvidenceIsDefensiveCopy(t *testing.T) {
	tracker, err := newUpgradeHaltTracker("v2.3.0", 55, 5*time.Second)
	require.NoError(t, err)
	observations := validUpgradeHaltObservations("v2.3.0", 55)
	want := cloneUpgradeHaltObservations(observations)
	complete, err := tracker.Observe(time.Now(), observations, nil)
	require.NoError(t, err)
	require.False(t, complete)

	mutatedInput := observations["validator-0"]
	mutatedInput.ApplicationHash = "MUTATED-INPUT"
	observations[mutatedInput.Node] = mutatedInput
	firstRead := tracker.Evidence()
	require.Equal(t, want, firstRead)

	mutatedEvidence := firstRead["validator-1"]
	mutatedEvidence.ApplicationHash = "MUTATED-EVIDENCE"
	firstRead[mutatedEvidence.Node] = mutatedEvidence
	delete(firstRead, "fullnode-0")
	require.Equal(t, want, tracker.Evidence())
}

func TestNewUpgradeHaltTrackerRejectsInvalidContract(t *testing.T) {
	for _, test := range []struct {
		name   string
		plan   string
		height int64
		window time.Duration
		want   string
	}{
		{name: "missing plan", height: 55, window: 5 * time.Second, want: "plan name"},
		{name: "nonpositive height", plan: "v2.3.0", window: 5 * time.Second, want: "height"},
		{name: "nonpositive window", plan: "v2.3.0", height: 55, want: "stability window"},
	} {
		t.Run(test.name, func(t *testing.T) {
			tracker, err := newUpgradeHaltTracker(test.plan, test.height, test.window)
			require.Nil(t, tracker)
			require.ErrorContains(t, err, test.want)
		})
	}
}

func TestDecodeObservedOldBinaryUpgradePlanRetriesIncompleteFile(t *testing.T) {
	for _, contents := range [][]byte{
		nil,
		[]byte(`{"name":"v2.3.0","height":`),
	} {
		plan, retryErr, validationErr := decodeObservedOldBinaryUpgradePlan(
			contents,
			"v2.3.0",
			55,
		)

		require.Empty(t, plan)
		require.Error(t, retryErr)
		require.NoError(t, validationErr)
	}
}

func TestDecodeObservedOldBinaryUpgradePlanRejectsCompleteInvalidFile(t *testing.T) {
	for _, test := range []struct {
		name     string
		contents string
	}{
		{name: "malformed JSON", contents: `{"name":"v2.3.0","height":"55"} trailing`},
		{name: "missing plan name", contents: `{"height":"55"}`},
		{name: "invalid plan height", contents: `{"name":"v2.3.0","height":"invalid"}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			plan, retryErr, validationErr := decodeObservedOldBinaryUpgradePlan(
				[]byte(test.contents),
				"v2.3.0",
				55,
			)

			require.Empty(t, plan)
			require.NoError(t, retryErr)
			require.Error(t, validationErr)
		})
	}
}

func TestDecodeObservedOldBinaryUpgradePlanRejectsCompleteWrongPlan(t *testing.T) {
	plan, retryErr, validationErr := decodeObservedOldBinaryUpgradePlan(
		[]byte(`{"name":"v2.4.0","height":"56"}`),
		"v2.3.0",
		55,
	)

	require.Equal(t, oldBinaryUpgradePlan{Name: "v2.4.0", Height: 56}, plan)
	require.NoError(t, retryErr)
	require.ErrorContains(t, validationErr, `plan "v2.4.0" at 56, want "v2.3.0" at 55`)
}

func TestDecodeObservedOldBinaryUpgradePlanAcceptsExactPlan(t *testing.T) {
	plan, retryErr, validationErr := decodeObservedOldBinaryUpgradePlan(
		[]byte(`{"name":"v2.3.0","height":"55"}`),
		"v2.3.0",
		55,
	)

	require.Equal(t, oldBinaryUpgradePlan{Name: "v2.3.0", Height: 55}, plan)
	require.NoError(t, retryErr)
	require.NoError(t, validationErr)
}

func validUpgradeHaltObservations(planName string, upgradeHeight int64) map[string]upgradeHaltObservation {
	observations := make(map[string]upgradeHaltObservation, 5)
	for _, node := range []string{"validator-0", "validator-1", "validator-2", "validator-3", "fullnode-0"} {
		observations[node] = upgradeHaltObservation{
			Node:                node,
			ConsensusHeight:     upgradeHeight,
			ConsensusBlockID:    "11223344",
			ConsensusAppHash:    "AABBCCDD",
			ApplicationHeight:   upgradeHeight - 1,
			ApplicationHash:     "AABBCCDD",
			PlanName:            planName,
			PlanHeight:          upgradeHeight,
			ContainerRunning:    true,
			ContainerStatus:     "running",
			ContainerPaused:     false,
			ContainerRestarting: false,
			ContainerOOMKilled:  false,
			ContainerDead:       false,
		}
	}
	return observations
}
