package e2e_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
	"github.com/stretchr/testify/require"
)

func TestUpgradeChaosAfterHandlerRecoveryRestartsKilledValidatorBeforeReplacingThird(t *testing.T) {
	t.Parallel()

	var operations []string
	stall, switched, err := runUpgradeChaosAfterHandlerRecovery(
		func() error {
			operations = append(operations, "restart-killed")
			return nil
		},
		func() (harness.QuorumHeightWindow, error) {
			operations = append(operations, "observe-two-current-stall")
			return harness.QuorumHeightWindow{StartHeight: 61, EndHeight: 61}, nil
		},
		func() (harness.UpgradeNodeSwitchEvidence, error) {
			operations = append(operations, "switch-third")
			return harness.UpgradeNodeSwitchEvidence{Phase: "third-quorum"}, nil
		},
	)
	require.NoError(t, err)
	require.Equal(t, []string{"restart-killed", "observe-two-current-stall", "switch-third"}, operations)
	require.Equal(t, int64(61), stall.EndHeight)
	require.Equal(t, "third-quorum", switched.Phase)
}

func TestUpgradeChaosAfterHandlerRecoveryStopsWhenKilledValidatorCannotRestart(t *testing.T) {
	t.Parallel()

	want := errors.New("restart failed")
	stallCalled := false
	switchCalled := false
	_, _, err := runUpgradeChaosAfterHandlerRecovery(
		func() error { return want },
		func() (harness.QuorumHeightWindow, error) {
			stallCalled = true
			return harness.QuorumHeightWindow{}, nil
		},
		func() (harness.UpgradeNodeSwitchEvidence, error) {
			switchCalled = true
			return harness.UpgradeNodeSwitchEvidence{}, nil
		},
	)
	require.ErrorIs(t, err, want)
	require.False(t, stallCalled)
	require.False(t, switchCalled)
}

func TestDecodeUpgradeChaosPlanRequiresPermutationAndQuorumKill(t *testing.T) {
	t.Parallel()

	beforePlan, err := decodeUpgradeChaosPlan([]byte(`{"seed":101,"switch_order":[1,2,0,3],"kill_index":0,"boundary":"before-upgrade-handler-commit"}`))
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 0, 3}, beforePlan.SwitchOrder)

	afterPlan, err := decodeUpgradeChaosPlan([]byte(`{"seed":202,"switch_order":[2,0,3,1],"kill_index":2,"boundary":"after-first-post-upgrade-commit"}`))
	require.NoError(t, err)
	require.Equal(t, []int{2, 0, 3, 1}, afterPlan.SwitchOrder)

	_, err = decodeUpgradeChaosPlan([]byte(`{"seed":101,"switch_order":[0,0,2,3],"kill_index":0,"boundary":"before-upgrade-handler-commit"}`))
	require.ErrorContains(t, err, "duplicated")

	_, err = decodeUpgradeChaosPlan([]byte(`{"seed":101,"switch_order":[0,1,2,3],"kill_index":0,"boundary":"before-upgrade-handler-commit"}`))
	require.ErrorContains(t, err, "third activated validator")

	_, err = decodeUpgradeChaosPlan([]byte(`{"seed":202,"switch_order":[0,1,2,3],"kill_index":2,"boundary":"after-first-post-upgrade-commit"}`))
	require.ErrorContains(t, err, "first two switched")

	_, err = decodeUpgradeChaosPlan([]byte(`{"seed":0,"switch_order":[1,2,0,3],"kill_index":0,"boundary":"before-upgrade-handler-commit"}`))
	require.ErrorContains(t, err, "non-zero")

	_, err = decodeUpgradeChaosPlan([]byte(`{"seed":101,"switch_order":[1,2,0,3],"kill_index":0}`))
	require.ErrorContains(t, err, "boundary")

	_, err = decodeUpgradeChaosPlan([]byte(`{"seed":101,"switch_order":[1,2,0,3],"kill_index":0,"boundary":"after-h-plus-three"}`))
	require.ErrorContains(t, err, "boundary")
}

func TestUpgradeChaosBoundaryWindowPinsApplicationHeightOnBothSidesOfTheHandlerCommit(t *testing.T) {
	t.Parallel()

	before, err := upgradeChaosBoundaryWindowFor(50, upgradeChaosBoundaryBeforeUpgradeHandlerCommit)
	require.NoError(t, err)
	require.Equal(t, upgradeChaosBoundaryWindow{TargetHeight: 49, MinimumHeight: 49, MaximumHeight: 49}, before)

	after, err := upgradeChaosBoundaryWindowFor(50, upgradeChaosBoundaryAfterFirstPostUpgradeCommit)
	require.NoError(t, err)
	require.Equal(t, upgradeChaosBoundaryWindow{TargetHeight: 50, MinimumHeight: 50, MaximumHeight: 50}, after)
}

func TestObserveExactUpgradeChaosBoundaryAcceptsOnlyTheTargetHeight(t *testing.T) {
	t.Parallel()

	heights := []int64{49, 49, 50}
	index := 0
	observed, err := observeExactUpgradeChaosBoundary(
		context.Background(),
		upgradeChaosBoundaryWindow{TargetHeight: 50, MinimumHeight: 50, MaximumHeight: 50},
		time.Microsecond,
		func(context.Context) (int64, error) {
			height := heights[index]
			if index < len(heights)-1 {
				index++
			}
			return height, nil
		},
	)
	require.NoError(t, err)
	require.Equal(t, int64(50), observed.Height)
	require.False(t, observed.ObservedAt.IsZero())
}

func TestObserveExactUpgradeChaosBoundaryFailsClosedOnOvershoot(t *testing.T) {
	t.Parallel()

	_, err := observeExactUpgradeChaosBoundary(
		context.Background(),
		upgradeChaosBoundaryWindow{TargetHeight: 50, MinimumHeight: 50, MaximumHeight: 50},
		time.Microsecond,
		func(context.Context) (int64, error) { return 51, nil },
	)
	require.ErrorContains(t, err, "outside exact boundary window")
}

func TestObserveExactUpgradeChaosBoundaryPropagatesHeightErrors(t *testing.T) {
	t.Parallel()

	want := errors.New("rpc unavailable")
	_, err := observeExactUpgradeChaosBoundary(
		context.Background(),
		upgradeChaosBoundaryWindow{TargetHeight: 50, MinimumHeight: 50, MaximumHeight: 50},
		time.Microsecond,
		func(context.Context) (int64, error) { return 0, want },
	)
	require.ErrorIs(t, err, want)
}

func TestValidateUpgradeChaosBoundaryEvidenceRejectsLateKillAndIncompleteRecovery(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	evidence := validUpgradeChaosBoundaryEvidence(now)
	require.NoError(t, evidence.Validate())

	late := evidence
	late.KillObservedApplicationHeight = 51
	require.ErrorContains(t, late.Validate(), "kill application height")

	lateObservation := evidence
	lateObservation.FirstPostUpgradeObservedAt = now.Add(2 * time.Millisecond)
	require.ErrorContains(t, lateObservation.Validate(), "before SIGKILL")

	wrongHeightKind := evidence
	wrongHeightKind.HeightKind = "consensus"
	require.ErrorContains(t, wrongHeightKind.Validate(), "height kind")

	wrongObserverApplicationHeight := evidence
	wrongObserverApplicationHeight.ObserverApplicationHeightAfterKill = 49
	require.ErrorContains(t, wrongObserverApplicationHeight.Validate(), "observer application height")

	consensusAdvancedAfterKill := evidence
	consensusAdvancedAfterKill.ConsensusHeightAfterKill = 51
	require.ErrorContains(t, consensusAdvancedAfterKill.Validate(), "consensus height escaped")

	badCarrier := evidence
	badCarrier.CarrierHeaderAppHash = strings.Repeat("CD", 32)
	require.ErrorContains(t, badCarrier.Validate(), "carrier")

	shortMigrationHash := evidence
	shortMigrationHash.MigrationResultAppHash = "AABBCCDD"
	require.ErrorContains(t, shortMigrationHash.Validate(), "64 hexadecimal")

	nonHexCarrierHash := evidence
	nonHexCarrierHash.CarrierHeaderAppHash = strings.Repeat("GG", 32)
	require.ErrorContains(t, nonHexCarrierHash.Validate(), "not hexadecimal")

	failedInjection := evidence
	failedInjection.FaultInjectionSucceeded = false
	failedInjection.FaultInjectionError = "SIGKILL failed"
	require.ErrorContains(t, failedInjection.Validate(), "fault injection failed")

	unrecordedInjection := evidence
	unrecordedInjection.FaultInjectionSucceeded = false
	require.ErrorContains(t, unrecordedInjection.Validate(), "not recorded as successful")

	contradictoryInjection := evidence
	contradictoryInjection.FaultInjectionError = "unexpected error"
	require.ErrorContains(t, contradictoryInjection.Validate(), "succeeded with an error")

	incomplete := evidence
	incomplete.RecoveryEndHeight = 50
	require.ErrorContains(t, incomplete.Validate(), "recovery")

	before := evidence
	before.Boundary = upgradeChaosBoundaryBeforeUpgradeHandlerCommit
	before.TargetKillHeight = 49
	before.MinimumKillHeight = 49
	before.MaximumKillHeight = 49
	before.KillObservedApplicationHeight = 49
	before.ObserverApplicationHeightAfterKill = 49
	before.ConsensusHeightAtKill = 50
	before.ConsensusHeightAfterKill = 50
	before.RecoveryStartHeight = 50
	before.RecoveryEndHeight = 53
	before.FirstPostUpgradeObservedAt = before.RestartCompletedAt.Add(time.Millisecond)
	before.RecoveryCompletedAt = before.FirstPostUpgradeObservedAt.Add(time.Millisecond)
	require.NoError(t, before.Validate())

	applicationHeightUsedForRecovery := before
	applicationHeightUsedForRecovery.RecoveryStartHeight = 49
	require.ErrorContains(t, applicationHeightUsedForRecovery.Validate(), "recovery")

	before.FirstPostUpgradeObservedAt = before.RestartRequestedAt.Add(-time.Millisecond)
	require.ErrorContains(t, before.Validate(), "before validator restart")
}

func TestUpgradeChaosContainerLogEvidenceRequiresExplicitSuccessfulCapture(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	evidence, err := newUpgradeChaosContainerLogEvidence(1, "panacea-val-1")
	require.NoError(t, err)
	evidence.CaptureStartedAt = now
	evidence.CapturedAt = now.Add(time.Millisecond)
	evidence.CaptureSucceeded = true
	require.Equal(t, "chaos/upgraded-container-logs/pre-coordinated-restart/node-1.log", evidence.ArtifactPath)
	require.NoError(t, evidence.Validate())

	wrongPath := evidence
	wrongPath.ArtifactPath = "chaos/node-1.log"
	require.ErrorContains(t, wrongPath.Validate(), "artifact path")

	failed := evidence
	failed.CaptureSucceeded = false
	failed.CaptureError = "Docker logs unavailable"
	require.ErrorContains(t, failed.Validate(), "capture for panacea-val-1 failed")

	contradictory := evidence
	contradictory.CaptureError = "unexpected error"
	require.ErrorContains(t, contradictory.Validate(), "succeeded with an error")
}

func TestUpgradeChaosRecoverySummaryRequiresPostRestartFinalAgreement(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	logs := make([]upgradeChaosContainerLogEvidence, 2)
	for index, node := range []string{"panacea-val-0", "panacea-val-1"} {
		logEvidence, err := newUpgradeChaosContainerLogEvidence(index, node)
		require.NoError(t, err)
		logEvidence.CaptureStartedAt = now
		logEvidence.CapturedAt = now.Add(time.Millisecond)
		logEvidence.CaptureSucceeded = true
		logs[index] = logEvidence
	}

	evidence := upgradeChaosEvidence{
		BoundaryFault: validUpgradeChaosBoundaryEvidence(now),
		PreRestartAgreement: harness.QuorumAgreement{
			Height: 60, BlockHash: "PRE-BLOCK", AppHash: "PRE-APP", Nodes: []string{"panacea-val-0", "panacea-val-1"},
		},
		FinalAgreement: harness.QuorumAgreement{
			Height: 63, BlockHash: "POST-BLOCK", AppHash: "POST-APP", Nodes: []string{"panacea-val-0", "panacea-val-1"},
		},
		UpgradedContainerLogs: logs,
		Restart:               upgradeRestartEvidence{BeforeHeight: 60, TargetHeight: 63},
	}
	require.NoError(t, evidence.Validate())

	stale := evidence
	stale.FinalAgreement = stale.PreRestartAgreement
	require.ErrorContains(t, stale.Validate(), "not post-restart target height")

	missingLog := evidence
	missingLog.UpgradedContainerLogs = missingLog.UpgradedContainerLogs[:1]
	require.ErrorContains(t, missingLog.Validate(), "want 1 captured upgraded containers")

	wrongPreRestartNodes := evidence
	wrongPreRestartNodes.PreRestartAgreement.Nodes = []string{"panacea-val-0", "panacea-val-2"}
	require.ErrorContains(t, wrongPreRestartNodes.Validate(), "pre-restart agreement node \"panacea-val-2\"")

	invalidBoundary := evidence
	invalidBoundary.BoundaryFault.FaultInjectionSucceeded = false
	require.ErrorContains(t, invalidBoundary.Validate(), "validate upgrade chaos boundary fault")
}

func validUpgradeChaosBoundaryEvidence(now time.Time) upgradeChaosBoundaryEvidence {
	validHash := strings.Repeat("AB", 32)
	return upgradeChaosBoundaryEvidence{
		Boundary:                           upgradeChaosBoundaryAfterFirstPostUpgradeCommit,
		HandlerHeight:                      50,
		HeightKind:                         "application",
		TargetKillHeight:                   50,
		MinimumKillHeight:                  50,
		MaximumKillHeight:                  50,
		KillValidatorIndex:                 0,
		KillNode:                           "validator-0",
		ObserverNode:                       "validator-1",
		KillObservedApplicationHeight:      50,
		ObserverApplicationHeightAfterKill: 50,
		ConsensusHeightAtKill:              50,
		ConsensusHeightAfterKill:           50,
		KillRequestedAt:                    now.Add(time.Millisecond),
		KillCompletedAt:                    now.Add(2 * time.Millisecond),
		FaultInjectionSucceeded:            true,
		FirstPostUpgradeHeight:             50,
		FirstPostUpgradeObservedAt:         now,
		FirstPostUpgradeBlockTime:          now.Add(-time.Second),
		MigrationResultAppHash:             validHash,
		CarrierHeaderAppHash:               validHash,
		RestartRequestedAt:                 now.Add(3 * time.Millisecond),
		RestartCompletedAt:                 now.Add(4 * time.Millisecond),
		RestartSucceeded:                   true,
		RecoveryStartHeight:                50,
		RecoveryEndHeight:                  53,
		RecoveryCompletedAt:                now.Add(5 * time.Millisecond),
		RecoverySucceeded:                  true,
	}
}
