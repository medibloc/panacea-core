package e2e_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

type upgradeChaosPlan struct {
	Seed        int64                `json:"seed"`
	SwitchOrder []int                `json:"switch_order"`
	KillIndex   int                  `json:"kill_index"`
	Boundary    upgradeChaosBoundary `json:"boundary"`
}

type upgradeChaosBoundary string

const (
	upgradeChaosBoundaryBeforeUpgradeHandlerCommit  upgradeChaosBoundary = "before-upgrade-handler-commit"
	upgradeChaosBoundaryAfterFirstPostUpgradeCommit upgradeChaosBoundary = "after-first-post-upgrade-commit"
	upgradeChaosBoundaryPollInterval                                     = 25 * time.Millisecond
	upgradeChaosBoundaryKillTimeout                                      = 90 * time.Second
	upgradeChaosQuorumRecoveryTimeout                                    = 2 * time.Minute
	upgradeChaosLogCaptureTimeout                                        = 15 * time.Second
	upgradeChaosPreRestartLogDirectory                                   = "chaos/upgraded-container-logs/pre-coordinated-restart"
	upgradeChaosPreRestartLogManifest                                    = upgradeChaosPreRestartLogDirectory + "/manifest.json"
)

type upgradeChaosBoundaryWindow struct {
	TargetHeight  int64 `json:"target_height"`
	MinimumHeight int64 `json:"minimum_height"`
	MaximumHeight int64 `json:"maximum_height"`
}

type upgradeChaosBoundaryObservation struct {
	Height     int64     `json:"height"`
	ObservedAt time.Time `json:"observed_at"`
}

type upgradeChaosBoundaryEvidence struct {
	Boundary                           upgradeChaosBoundary `json:"boundary"`
	HandlerHeight                      int64                `json:"handler_height"`
	HeightKind                         string               `json:"height_kind"`
	TargetKillHeight                   int64                `json:"target_kill_height"`
	MinimumKillHeight                  int64                `json:"minimum_kill_height"`
	MaximumKillHeight                  int64                `json:"maximum_kill_height"`
	KillValidatorIndex                 int                  `json:"kill_validator_index"`
	KillNode                           string               `json:"kill_node"`
	ObserverNode                       string               `json:"observer_node"`
	KillObservedApplicationHeight      int64                `json:"kill_observed_application_height"`
	ObserverApplicationHeightAfterKill int64                `json:"observer_application_height_after_kill"`
	ConsensusHeightAtKill              int64                `json:"consensus_height_at_kill"`
	ConsensusHeightAfterKill           int64                `json:"consensus_height_after_kill"`
	KillRequestedAt                    time.Time            `json:"kill_requested_at"`
	KillCompletedAt                    time.Time            `json:"kill_completed_at"`
	FaultInjectionSucceeded            bool                 `json:"fault_injection_succeeded"`
	FaultInjectionError                string               `json:"fault_injection_error,omitempty"`
	FirstPostUpgradeHeight             int64                `json:"first_post_upgrade_height"`
	FirstPostUpgradeObservedAt         time.Time            `json:"first_post_upgrade_observed_at"`
	FirstPostUpgradeBlockTime          time.Time            `json:"first_post_upgrade_block_time"`
	MigrationResultAppHash             string               `json:"migration_result_app_hash"`
	CarrierHeaderAppHash               string               `json:"carrier_header_app_hash"`
	RestartRequestedAt                 time.Time            `json:"restart_requested_at"`
	RestartCompletedAt                 time.Time            `json:"restart_completed_at"`
	RestartSucceeded                   bool                 `json:"restart_succeeded"`
	RestartError                       string               `json:"restart_error,omitempty"`
	RecoveryStartHeight                int64                `json:"recovery_start_height"`
	RecoveryEndHeight                  int64                `json:"recovery_end_height"`
	RecoveryCompletedAt                time.Time            `json:"recovery_completed_at"`
	RecoverySucceeded                  bool                 `json:"recovery_succeeded"`
	RecoveryError                      string               `json:"recovery_error,omitempty"`
}

type upgradeChaosEvidence struct {
	Plan                   upgradeChaosPlan                    `json:"plan"`
	UpgradeHeight          int64                               `json:"upgrade_height"`
	TwoNewBinaryStall      harness.QuorumHeightWindow          `json:"two_new_binary_stall"`
	FirstQuorumProgress    harness.QuorumHeightWindow          `json:"first_quorum_progress"`
	ForcedStopStall        harness.QuorumHeightWindow          `json:"forced_stop_stall"`
	ForcedRecoveryProgress harness.QuorumHeightWindow          `json:"forced_recovery_progress"`
	BoundaryFault          upgradeChaosBoundaryEvidence        `json:"boundary_fault"`
	PreRestartAgreement    harness.QuorumAgreement             `json:"pre_restart_agreement"`
	FinalAgreement         harness.QuorumAgreement             `json:"final_agreement"`
	UpgradedContainerLogs  []upgradeChaosContainerLogEvidence  `json:"upgraded_container_logs_before_restart"`
	Restart                upgradeRestartEvidence              `json:"restart"`
	ModuleVersions         map[string]uint64                   `json:"module_versions"`
	NodeVersionsBefore     []upgradeChaosNodeVersionEvidence   `json:"node_versions_before_restart"`
	NodeVersionsAfter      []upgradeChaosNodeVersionEvidence   `json:"node_versions_after_restart"`
	AppliedPlanBefore      []upgradeChaosAppliedPlanEvidence   `json:"applied_plan_before_restart"`
	AppliedPlanAfter       []upgradeChaosAppliedPlanEvidence   `json:"applied_plan_after_restart"`
	Switches               []harness.UpgradeNodeSwitchEvidence `json:"switches"`
}

type upgradeChaosNodeVersionEvidence struct {
	Node     string            `json:"node"`
	Versions map[string]uint64 `json:"versions"`
}

type upgradeChaosAppliedPlanEvidence struct {
	Node     string          `json:"node"`
	Response json.RawMessage `json:"response"`
}

type upgradeChaosContainerLogEvidence struct {
	NodeIndex        int       `json:"node_index"`
	Node             string    `json:"node"`
	ArtifactPath     string    `json:"artifact_path"`
	CaptureStartedAt time.Time `json:"capture_started_at"`
	CapturedAt       time.Time `json:"captured_at"`
	CaptureSucceeded bool      `json:"capture_succeeded"`
	CaptureError     string    `json:"capture_error,omitempty"`
}

func newUpgradeChaosContainerLogEvidence(index int, node string) (upgradeChaosContainerLogEvidence, error) {
	if index < 0 {
		return upgradeChaosContainerLogEvidence{}, fmt.Errorf("upgrade chaos log node index must not be negative: %d", index)
	}
	node = strings.TrimSpace(node)
	if node == "" {
		return upgradeChaosContainerLogEvidence{}, errors.New("upgrade chaos log node name is required")
	}
	return upgradeChaosContainerLogEvidence{
		NodeIndex:    index,
		Node:         node,
		ArtifactPath: fmt.Sprintf("%s/node-%d.log", upgradeChaosPreRestartLogDirectory, index),
	}, nil
}

func (e upgradeChaosContainerLogEvidence) Validate() error {
	planned, err := newUpgradeChaosContainerLogEvidence(e.NodeIndex, e.Node)
	if err != nil {
		return err
	}
	if e.ArtifactPath != planned.ArtifactPath {
		return fmt.Errorf("upgrade chaos log artifact path %q, want %q", e.ArtifactPath, planned.ArtifactPath)
	}
	if e.CaptureStartedAt.IsZero() || e.CapturedAt.Before(e.CaptureStartedAt) {
		return fmt.Errorf("upgrade chaos log capture timestamps for %s are missing or out of order", e.Node)
	}
	if !e.CaptureSucceeded {
		if strings.TrimSpace(e.CaptureError) == "" {
			return fmt.Errorf("upgrade chaos log capture for %s was not recorded as successful", e.Node)
		}
		return fmt.Errorf("upgrade chaos log capture for %s failed: %s", e.Node, e.CaptureError)
	}
	if strings.TrimSpace(e.CaptureError) != "" {
		return fmt.Errorf("upgrade chaos log capture for %s succeeded with an error: %s", e.Node, e.CaptureError)
	}
	return nil
}

func (e upgradeChaosEvidence) Validate() error {
	var validationErrors []error
	if err := e.BoundaryFault.Validate(); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("validate upgrade chaos boundary fault: %w", err))
	}
	if e.Restart.TargetHeight <= 0 {
		validationErrors = append(validationErrors, fmt.Errorf("upgrade chaos restart target height must be positive: %d", e.Restart.TargetHeight))
	}
	if e.PreRestartAgreement.Height <= 0 {
		validationErrors = append(validationErrors, errors.New("upgrade chaos pre-restart agreement is missing"))
	}
	if e.FinalAgreement.Height != e.Restart.TargetHeight {
		validationErrors = append(validationErrors, fmt.Errorf(
			"upgrade chaos final agreement height %d is not post-restart target height %d",
			e.FinalAgreement.Height,
			e.Restart.TargetHeight,
		))
	}
	if e.FinalAgreement.Height <= e.PreRestartAgreement.Height {
		validationErrors = append(validationErrors, fmt.Errorf(
			"upgrade chaos final agreement height %d did not advance beyond pre-restart agreement height %d",
			e.FinalAgreement.Height,
			e.PreRestartAgreement.Height,
		))
	}
	if strings.TrimSpace(e.FinalAgreement.BlockHash) == "" || strings.TrimSpace(e.FinalAgreement.AppHash) == "" {
		validationErrors = append(validationErrors, errors.New("upgrade chaos final agreement hashes are incomplete"))
	}
	if len(e.UpgradedContainerLogs) == 0 {
		validationErrors = append(validationErrors, errors.New("upgrade chaos pre-restart upgraded-container logs are missing"))
	}

	logNodes := make(map[string]struct{}, len(e.UpgradedContainerLogs))
	for _, logEvidence := range e.UpgradedContainerLogs {
		if err := logEvidence.Validate(); err != nil {
			validationErrors = append(validationErrors, err)
		}
		if _, duplicate := logNodes[logEvidence.Node]; duplicate {
			validationErrors = append(validationErrors, fmt.Errorf("upgrade chaos log node %q is duplicated", logEvidence.Node))
		}
		logNodes[logEvidence.Node] = struct{}{}
	}
	if len(e.FinalAgreement.Nodes) != len(logNodes) {
		validationErrors = append(validationErrors, fmt.Errorf(
			"upgrade chaos final agreement has %d nodes, want %d captured upgraded containers",
			len(e.FinalAgreement.Nodes),
			len(logNodes),
		))
	}
	seenAgreementNodes := make(map[string]struct{}, len(e.FinalAgreement.Nodes))
	for _, node := range e.FinalAgreement.Nodes {
		if strings.TrimSpace(node) == "" {
			validationErrors = append(validationErrors, errors.New("upgrade chaos final agreement has an empty node name"))
			continue
		}
		if _, duplicate := seenAgreementNodes[node]; duplicate {
			validationErrors = append(validationErrors, fmt.Errorf("upgrade chaos final agreement node %q is duplicated", node))
		}
		seenAgreementNodes[node] = struct{}{}
		if _, captured := logNodes[node]; !captured {
			validationErrors = append(validationErrors, fmt.Errorf("upgrade chaos final agreement node %q has no pre-restart upgraded-container log", node))
		}
	}
	if len(e.PreRestartAgreement.Nodes) != len(logNodes) {
		validationErrors = append(validationErrors, fmt.Errorf(
			"upgrade chaos pre-restart agreement has %d nodes, want %d captured upgraded containers",
			len(e.PreRestartAgreement.Nodes),
			len(logNodes),
		))
	}
	seenPreRestartNodes := make(map[string]struct{}, len(e.PreRestartAgreement.Nodes))
	for _, node := range e.PreRestartAgreement.Nodes {
		if strings.TrimSpace(node) == "" {
			validationErrors = append(validationErrors, errors.New("upgrade chaos pre-restart agreement has an empty node name"))
			continue
		}
		if _, duplicate := seenPreRestartNodes[node]; duplicate {
			validationErrors = append(validationErrors, fmt.Errorf("upgrade chaos pre-restart agreement node %q is duplicated", node))
		}
		seenPreRestartNodes[node] = struct{}{}
		if _, captured := logNodes[node]; !captured {
			validationErrors = append(validationErrors, fmt.Errorf("upgrade chaos pre-restart agreement node %q has no pre-restart upgraded-container log", node))
		}
		if _, final := seenAgreementNodes[node]; !final {
			validationErrors = append(validationErrors, fmt.Errorf("upgrade chaos pre-restart agreement node %q is missing from final agreement", node))
		}
	}
	return errors.Join(validationErrors...)
}

func validateUpgradeChaosAppHash(label string, value string) error {
	if len(value) != 64 {
		return fmt.Errorf("%s app hash must contain exactly 64 hexadecimal characters, got %d", label, len(value))
	}
	if _, err := hex.DecodeString(value); err != nil {
		return fmt.Errorf("%s app hash is not hexadecimal: %w", label, err)
	}
	return nil
}

func (p upgradeChaosPlan) Validate(validatorCount int) error {
	if p.Seed == 0 {
		return errors.New("upgrade chaos seed must be non-zero")
	}
	if validatorCount < 4 {
		return fmt.Errorf("upgrade chaos requires at least four validators, got %d", validatorCount)
	}
	if len(p.SwitchOrder) != validatorCount {
		return fmt.Errorf("upgrade chaos switch order has %d entries, want %d", len(p.SwitchOrder), validatorCount)
	}
	seen := make(map[int]struct{}, validatorCount)
	for _, index := range p.SwitchOrder {
		if index < 0 || index >= validatorCount {
			return fmt.Errorf("upgrade chaos switch index %d outside [0,%d)", index, validatorCount)
		}
		if _, duplicate := seen[index]; duplicate {
			return fmt.Errorf("upgrade chaos switch index %d is duplicated", index)
		}
		seen[index] = struct{}{}
	}
	switch p.Boundary {
	case upgradeChaosBoundaryBeforeUpgradeHandlerCommit:
		if p.KillIndex != p.SwitchOrder[2] {
			return fmt.Errorf(
				"pre-handler upgrade chaos kill index %d must be the third activated validator %d",
				p.KillIndex,
				p.SwitchOrder[2],
			)
		}
	case upgradeChaosBoundaryAfterFirstPostUpgradeCommit:
		if p.KillIndex != p.SwitchOrder[0] && p.KillIndex != p.SwitchOrder[1] {
			return fmt.Errorf("post-handler upgrade chaos kill index %d is not one of the first two switched validators", p.KillIndex)
		}
	default:
		return fmt.Errorf("upgrade chaos boundary %q is unsupported", p.Boundary)
	}
	return nil
}

func upgradeChaosBoundaryWindowFor(upgradeHeight int64, boundary upgradeChaosBoundary) (upgradeChaosBoundaryWindow, error) {
	if upgradeHeight <= 1 {
		return upgradeChaosBoundaryWindow{}, fmt.Errorf("upgrade handler height must be greater than one: %d", upgradeHeight)
	}
	var target int64
	switch boundary {
	case upgradeChaosBoundaryBeforeUpgradeHandlerCommit:
		target = upgradeHeight - 1
	case upgradeChaosBoundaryAfterFirstPostUpgradeCommit:
		target = upgradeHeight
	default:
		return upgradeChaosBoundaryWindow{}, fmt.Errorf("upgrade chaos boundary %q is unsupported", boundary)
	}
	return upgradeChaosBoundaryWindow{TargetHeight: target, MinimumHeight: target, MaximumHeight: target}, nil
}

func observeExactUpgradeChaosBoundary(
	ctx context.Context,
	window upgradeChaosBoundaryWindow,
	pollInterval time.Duration,
	height func(context.Context) (int64, error),
) (upgradeChaosBoundaryObservation, error) {
	if window.TargetHeight <= 0 || window.MinimumHeight != window.TargetHeight || window.MaximumHeight != window.TargetHeight {
		return upgradeChaosBoundaryObservation{}, fmt.Errorf("upgrade chaos boundary window must contain one positive exact target: %+v", window)
	}
	if pollInterval <= 0 {
		return upgradeChaosBoundaryObservation{}, errors.New("upgrade chaos boundary poll interval must be positive")
	}
	if height == nil {
		return upgradeChaosBoundaryObservation{}, errors.New("upgrade chaos boundary height observer is required")
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		observedHeight, err := height(ctx)
		if err != nil {
			return upgradeChaosBoundaryObservation{}, fmt.Errorf("observe upgrade chaos boundary height: %w", err)
		}
		if observedHeight > window.MaximumHeight {
			return upgradeChaosBoundaryObservation{}, fmt.Errorf(
				"observed height %d outside exact boundary window [%d,%d]",
				observedHeight,
				window.MinimumHeight,
				window.MaximumHeight,
			)
		}
		if observedHeight == window.TargetHeight {
			return upgradeChaosBoundaryObservation{Height: observedHeight, ObservedAt: time.Now().UTC()}, nil
		}
		select {
		case <-ctx.Done():
			return upgradeChaosBoundaryObservation{}, fmt.Errorf(
				"wait for exact upgrade chaos boundary height %d: %w",
				window.TargetHeight,
				ctx.Err(),
			)
		case <-ticker.C:
		}
	}
}

func (e upgradeChaosBoundaryEvidence) Validate() error {
	window, err := upgradeChaosBoundaryWindowFor(e.HandlerHeight, e.Boundary)
	if err != nil {
		return err
	}
	if e.TargetKillHeight != window.TargetHeight || e.MinimumKillHeight != window.MinimumHeight || e.MaximumKillHeight != window.MaximumHeight {
		return fmt.Errorf("boundary evidence kill window %+v does not match planned window %+v", upgradeChaosBoundaryWindow{
			TargetHeight: e.TargetKillHeight, MinimumHeight: e.MinimumKillHeight, MaximumHeight: e.MaximumKillHeight,
		}, window)
	}
	if e.HeightKind != "application" {
		return fmt.Errorf("boundary evidence height kind %q is not application", e.HeightKind)
	}
	if e.KillObservedApplicationHeight < e.MinimumKillHeight || e.KillObservedApplicationHeight > e.MaximumKillHeight {
		return fmt.Errorf(
			"kill application height %d is outside exact boundary window [%d,%d]",
			e.KillObservedApplicationHeight,
			e.MinimumKillHeight,
			e.MaximumKillHeight,
		)
	}
	if e.KillValidatorIndex < 0 || strings.TrimSpace(e.KillNode) == "" || strings.TrimSpace(e.ObserverNode) == "" || e.KillNode == e.ObserverNode {
		return errors.New("kill validator and independent observer identities are incomplete")
	}
	if e.ObserverApplicationHeightAfterKill < e.MinimumKillHeight || e.ObserverApplicationHeightAfterKill > e.MaximumKillHeight {
		return fmt.Errorf(
			"independent observer application height after kill %d escaped boundary window [%d,%d]",
			e.ObserverApplicationHeightAfterKill,
			e.MinimumKillHeight,
			e.MaximumKillHeight,
		)
	}
	if e.ConsensusHeightAtKill != e.HandlerHeight || e.ConsensusHeightAfterKill != e.HandlerHeight {
		return fmt.Errorf(
			"consensus height escaped handler boundary: at-kill=%d after-kill=%d want=%d",
			e.ConsensusHeightAtKill,
			e.ConsensusHeightAfterKill,
			e.HandlerHeight,
		)
	}
	if e.KillRequestedAt.IsZero() || e.KillCompletedAt.Before(e.KillRequestedAt) {
		return errors.New("kill timestamps are missing or out of order")
	}
	if !e.FaultInjectionSucceeded {
		if strings.TrimSpace(e.FaultInjectionError) == "" {
			return errors.New("upgrade boundary fault injection was not recorded as successful")
		}
		return fmt.Errorf("upgrade boundary fault injection failed: %s", e.FaultInjectionError)
	}
	if strings.TrimSpace(e.FaultInjectionError) != "" {
		return fmt.Errorf("upgrade boundary fault injection succeeded with an error: %s", e.FaultInjectionError)
	}
	if e.FirstPostUpgradeHeight != e.HandlerHeight || e.FirstPostUpgradeObservedAt.IsZero() || e.FirstPostUpgradeBlockTime.IsZero() {
		return fmt.Errorf("first post-upgrade block evidence is incomplete at height %d", e.FirstPostUpgradeHeight)
	}
	if err := validateUpgradeChaosAppHash("migration result", e.MigrationResultAppHash); err != nil {
		return err
	}
	if err := validateUpgradeChaosAppHash("H+1 carrier header", e.CarrierHeaderAppHash); err != nil {
		return err
	}
	if !strings.EqualFold(e.MigrationResultAppHash, e.CarrierHeaderAppHash) {
		return fmt.Errorf(
			"migration result app hash %q does not match H+1 carrier header app hash %q",
			e.MigrationResultAppHash,
			e.CarrierHeaderAppHash,
		)
	}
	if e.Boundary == upgradeChaosBoundaryAfterFirstPostUpgradeCommit && e.FirstPostUpgradeObservedAt.After(e.KillRequestedAt) {
		return errors.New("post-handler boundary did not observe the first post-upgrade commit before SIGKILL")
	}
	if e.RestartRequestedAt.Before(e.KillCompletedAt) || e.RestartCompletedAt.Before(e.RestartRequestedAt) {
		return errors.New("restart timestamps are missing or out of order")
	}
	if !e.RestartSucceeded || e.RestartError != "" {
		return fmt.Errorf("validator restart did not succeed: %s", e.RestartError)
	}
	if e.Boundary == upgradeChaosBoundaryBeforeUpgradeHandlerCommit && e.FirstPostUpgradeObservedAt.Before(e.RestartCompletedAt) {
		return errors.New("pre-handler boundary observed the first post-upgrade commit before validator restart completed")
	}
	if e.RecoveryStartHeight != e.HandlerHeight || e.RecoveryEndHeight <= e.RecoveryStartHeight || e.RecoveryCompletedAt.Before(e.RestartCompletedAt) {
		return fmt.Errorf("recovery evidence is incomplete: start=%d end=%d", e.RecoveryStartHeight, e.RecoveryEndHeight)
	}
	if !e.RecoverySucceeded || e.RecoveryError != "" {
		return fmt.Errorf("validator recovery did not succeed: %s", e.RecoveryError)
	}
	return nil
}

func upgradeChaosBoundaryObserver(plan upgradeChaosPlan, validators []*cosmos.ChainNode) *cosmos.ChainNode {
	if len(plan.SwitchOrder) < 3 {
		return nil
	}
	for _, index := range plan.SwitchOrder[:3] {
		if index >= 0 && index < len(validators) && index != plan.KillIndex {
			return validators[index]
		}
	}
	return nil
}

func killUpgradeChaosValidatorAtBoundary(
	ctx context.Context,
	network *harness.Network,
	plan upgradeChaosPlan,
	upgradeHeight int64,
	observer *cosmos.ChainNode,
) (evidence upgradeChaosBoundaryEvidence, retErr error) {
	window, err := upgradeChaosBoundaryWindowFor(upgradeHeight, plan.Boundary)
	evidence = upgradeChaosBoundaryEvidence{
		Boundary:          plan.Boundary,
		HandlerHeight:     upgradeHeight,
		HeightKind:        "application",
		TargetKillHeight:  window.TargetHeight,
		MinimumKillHeight: window.MinimumHeight,
		MaximumKillHeight: window.MaximumHeight,
	}
	defer func() {
		evidence.FaultInjectionSucceeded = retErr == nil
		if retErr != nil {
			evidence.FaultInjectionError = retErr.Error()
		} else {
			evidence.FaultInjectionError = ""
		}
	}()
	if err != nil {
		return evidence, err
	}
	if network == nil || network.Chain == nil {
		return evidence, errors.New("upgrade chaos network is required")
	}
	if observer == nil {
		return evidence, errors.New("upgrade chaos boundary observer is required")
	}
	if plan.KillIndex < 0 || plan.KillIndex >= len(network.Chain.Validators) {
		return evidence, fmt.Errorf("upgrade chaos kill validator index %d is out of range", plan.KillIndex)
	}
	evidence.KillValidatorIndex = plan.KillIndex
	killNode := network.Chain.Validators[plan.KillIndex]
	evidence.KillNode = killNode.Name()
	evidence.ObserverNode = observer.Name()

	observation, err := observeExactUpgradeChaosBoundary(
		ctx,
		window,
		upgradeChaosBoundaryPollInterval,
		func(observationCtx context.Context) (int64, error) {
			_, applicationHeight, heightErr := upgradeChaosNodeHeights(observationCtx, killNode)
			return applicationHeight, heightErr
		},
	)
	if err != nil {
		return evidence, err
	}
	consensusHeight, applicationHeight, err := upgradeChaosNodeHeights(ctx, killNode)
	if err != nil {
		return evidence, fmt.Errorf("capture kill-target boundary heights: %w", err)
	}
	if applicationHeight != observation.Height {
		return evidence, fmt.Errorf(
			"kill-target application height changed from observed %d to %d before SIGKILL",
			observation.Height,
			applicationHeight,
		)
	}
	evidence.KillObservedApplicationHeight = observation.Height
	evidence.ConsensusHeightAtKill = consensusHeight
	evidence.KillRequestedAt = time.Now().UTC()
	if plan.Boundary == upgradeChaosBoundaryAfterFirstPostUpgradeCommit {
		evidence.FirstPostUpgradeHeight = upgradeHeight
		evidence.FirstPostUpgradeObservedAt = observation.ObservedAt
	}

	killErr := network.KillQuorumValidator(
		ctx,
		fmt.Sprintf("%s-sigkill", plan.Boundary),
		plan.KillIndex,
	)
	evidence.KillCompletedAt = time.Now().UTC()
	consensusHeightAfterKill, applicationHeightAfterKill, heightErr := upgradeChaosNodeHeights(ctx, observer)
	if heightErr != nil {
		heightErr = fmt.Errorf("observe independent node after upgrade boundary SIGKILL: %w", heightErr)
	} else {
		evidence.ConsensusHeightAfterKill = consensusHeightAfterKill
		evidence.ObserverApplicationHeightAfterKill = applicationHeightAfterKill
		if applicationHeightAfterKill < window.MinimumHeight || applicationHeightAfterKill > window.MaximumHeight {
			heightErr = fmt.Errorf(
				"independent observer application height after SIGKILL %d is outside boundary window [%d,%d]",
				applicationHeightAfterKill,
				window.MinimumHeight,
				window.MaximumHeight,
			)
		} else if consensusHeightAfterKill != upgradeHeight {
			heightErr = fmt.Errorf(
				"consensus height after SIGKILL %d escaped handler height %d",
				consensusHeightAfterKill,
				upgradeHeight,
			)
		}
	}
	return evidence, errors.Join(killErr, heightErr)
}

func upgradeChaosNodeHeights(ctx context.Context, node *cosmos.ChainNode) (int64, int64, error) {
	if node == nil || node.Client == nil {
		return 0, 0, errors.New("upgrade chaos height observer is unavailable")
	}
	status, err := node.Client.Status(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("query consensus status: %w", err)
	}
	if status == nil || status.SyncInfo.LatestBlockHeight <= 0 {
		return 0, 0, errors.New("query consensus status returned no positive height")
	}
	application, err := node.Client.ABCIInfo(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("query ABCI application info: %w", err)
	}
	if application == nil || application.Response.LastBlockHeight < 0 {
		return 0, 0, errors.New("query ABCI application info returned no valid height")
	}
	return status.SyncInfo.LatestBlockHeight, application.Response.LastBlockHeight, nil
}

func captureUpgradeChaosFirstPostUpgradeBlock(
	ctx context.Context,
	node *cosmos.ChainNode,
	upgradeHeight int64,
) (time.Time, time.Time, string, string, error) {
	if node == nil || node.Client == nil {
		return time.Time{}, time.Time{}, "", "", errors.New("first post-upgrade block observer is unavailable")
	}
	if upgradeHeight <= 0 {
		return time.Time{}, time.Time{}, "", "", fmt.Errorf("first post-upgrade block height must be positive: %d", upgradeHeight)
	}
	_, applicationHeight, err := upgradeChaosNodeHeights(ctx, node)
	observedAt := time.Now().UTC()
	if err != nil {
		return time.Time{}, time.Time{}, "", "", fmt.Errorf("query first post-upgrade application height %d: %w", upgradeHeight, err)
	}
	if applicationHeight < upgradeHeight {
		return time.Time{}, time.Time{}, "", "", fmt.Errorf(
			"first post-upgrade application height %d has not reached %d",
			applicationHeight,
			upgradeHeight,
		)
	}
	result, err := node.Client.Block(ctx, &upgradeHeight)
	if err != nil {
		return time.Time{}, time.Time{}, "", "", fmt.Errorf("query first post-upgrade block %d: %w", upgradeHeight, err)
	}
	if result == nil || result.Block == nil {
		return time.Time{}, time.Time{}, "", "", fmt.Errorf("query first post-upgrade block %d returned no block", upgradeHeight)
	}
	if result.Block.Height != upgradeHeight {
		return time.Time{}, time.Time{}, "", "", fmt.Errorf(
			"first post-upgrade block response height %d, want %d",
			result.Block.Height,
			upgradeHeight,
		)
	}
	if result.Block.Time.IsZero() {
		return time.Time{}, time.Time{}, "", "", fmt.Errorf("first post-upgrade block %d has no consensus timestamp", upgradeHeight)
	}
	results, err := node.Client.BlockResults(ctx, &upgradeHeight)
	if err != nil {
		return time.Time{}, time.Time{}, "", "", fmt.Errorf("query migration block results %d: %w", upgradeHeight, err)
	}
	if results == nil || len(results.AppHash) == 0 {
		return time.Time{}, time.Time{}, "", "", fmt.Errorf("migration block results %d have no app hash", upgradeHeight)
	}
	carrierHeight := upgradeHeight + 1
	carrier, err := node.Client.Block(ctx, &carrierHeight)
	if err != nil {
		return time.Time{}, time.Time{}, "", "", fmt.Errorf("query migration app-hash carrier block %d: %w", carrierHeight, err)
	}
	if carrier == nil || carrier.Block == nil || carrier.Block.Height != carrierHeight || len(carrier.Block.AppHash) == 0 {
		return time.Time{}, time.Time{}, "", "", fmt.Errorf("migration app-hash carrier block %d is incomplete", carrierHeight)
	}
	migrationHash := strings.ToUpper(fmt.Sprintf("%X", results.AppHash))
	carrierHash := strings.ToUpper(fmt.Sprintf("%X", []byte(carrier.Block.AppHash)))
	if err := validateUpgradeChaosAppHash("migration result", migrationHash); err != nil {
		return time.Time{}, time.Time{}, migrationHash, carrierHash, err
	}
	if err := validateUpgradeChaosAppHash("H+1 carrier header", carrierHash); err != nil {
		return time.Time{}, time.Time{}, migrationHash, carrierHash, err
	}
	if migrationHash != carrierHash {
		return time.Time{}, time.Time{}, migrationHash, carrierHash, fmt.Errorf(
			"migration result app hash %s does not match carrier header app hash %s",
			migrationHash,
			carrierHash,
		)
	}
	return observedAt, result.Block.Time.UTC(), migrationHash, carrierHash, nil
}

func restoreUpgradeChaosNodeForCleanup(
	t *testing.T,
	node *cosmos.ChainNode,
	needsRestore *bool,
	phase string,
	restore func(context.Context) error,
) {
	t.Helper()
	if needsRestore == nil || !*needsRestore || node == nil {
		return
	}
	if restore == nil {
		t.Errorf("restore %s node %s before artifact collection: restore callback is nil", phase, node.Name())
		return
	}
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cleanupCancel()
	if node.DockerClient != nil && node.ContainerID() != "" {
		inspect, err := node.DockerClient.ContainerInspect(cleanupCtx, node.ContainerID())
		if err == nil && inspect.State != nil && inspect.State.Running {
			*needsRestore = false
			return
		}
	}
	if err := restore(cleanupCtx); err != nil {
		t.Errorf("restore %s node %s before artifact collection: %v", phase, node.Name(), err)
		return
	}
	*needsRestore = false
}

func captureUpgradeChaosContainerLogs(
	ctx context.Context,
	network *harness.Network,
	nodes []*cosmos.ChainNode,
) ([]upgradeChaosContainerLogEvidence, error) {
	if network == nil {
		return nil, errors.New("upgrade chaos network is required for container log capture")
	}
	if len(nodes) == 0 {
		return nil, errors.New("upgrade chaos container log capture requires nodes")
	}

	evidence := make([]upgradeChaosContainerLogEvidence, len(nodes))
	seenNodes := make(map[string]struct{}, len(nodes))
	for index, node := range nodes {
		if node == nil {
			return nil, fmt.Errorf("upgrade chaos container log node %d is nil", index)
		}
		planned, err := newUpgradeChaosContainerLogEvidence(index, node.Name())
		if err != nil {
			return nil, err
		}
		if _, duplicate := seenNodes[planned.Node]; duplicate {
			return nil, fmt.Errorf("upgrade chaos container log node %q is duplicated", planned.Node)
		}
		seenNodes[planned.Node] = struct{}{}
		evidence[index] = planned
	}

	var captureErrors []error
	for index, node := range nodes {
		evidence[index].CaptureStartedAt = time.Now().UTC()
		captureCtx, captureCancel := context.WithTimeout(ctx, upgradeChaosLogCaptureTimeout)
		captureErr := network.CaptureNodeContainerLog(captureCtx, node, evidence[index].ArtifactPath)
		captureCancel()
		evidence[index].CapturedAt = time.Now().UTC()
		evidence[index].CaptureSucceeded = captureErr == nil
		if captureErr != nil {
			evidence[index].CaptureError = captureErr.Error()
			captureErrors = append(captureErrors, fmt.Errorf("capture upgraded container log for %s: %w", node.Name(), captureErr))
		}
	}
	if err := network.WriteArtifactJSON(upgradeChaosPreRestartLogManifest, evidence); err != nil {
		captureErrors = append(captureErrors, fmt.Errorf("record upgraded container log manifest: %w", err))
	}
	return evidence, errors.Join(captureErrors...)
}

func runUpgradeChaosAfterHandlerRecovery(
	restartKilled func() error,
	observeTwoCurrentStall func() (harness.QuorumHeightWindow, error),
	switchThird func() (harness.UpgradeNodeSwitchEvidence, error),
) (harness.QuorumHeightWindow, harness.UpgradeNodeSwitchEvidence, error) {
	if restartKilled == nil || observeTwoCurrentStall == nil || switchThird == nil {
		return harness.QuorumHeightWindow{}, harness.UpgradeNodeSwitchEvidence{}, errors.New(
			"post-handler upgrade chaos recovery callbacks are incomplete",
		)
	}
	if err := restartKilled(); err != nil {
		return harness.QuorumHeightWindow{}, harness.UpgradeNodeSwitchEvidence{}, fmt.Errorf(
			"restart killed validator before replacing third validator: %w",
			err,
		)
	}
	stall, err := observeTwoCurrentStall()
	if err != nil {
		return stall, harness.UpgradeNodeSwitchEvidence{}, fmt.Errorf(
			"observe two-current-validator stall after killed validator restart: %w",
			err,
		)
	}
	switched, err := switchThird()
	if err != nil {
		return stall, switched, fmt.Errorf("replace third validator after boundary restart: %w", err)
	}
	return stall, switched, nil
}

func TestV221UpgradeBoundaryChaos(t *testing.T) {
	if os.Getenv("PANACEA_E2E_UPGRADE_CHAOS") != "1" {
		t.Skip("use ./scripts/e2e/run.sh upgrade-chaos")
	}
	plans := []upgradeChaosPlan{
		{Seed: 101, SwitchOrder: []int{1, 2, 0, 3}, KillIndex: 0, Boundary: upgradeChaosBoundaryBeforeUpgradeHandlerCommit},
		{Seed: 202, SwitchOrder: []int{2, 0, 3, 1}, KillIndex: 2, Boundary: upgradeChaosBoundaryAfterFirstPostUpgradeCommit},
		{Seed: 303, SwitchOrder: []int{3, 2, 1, 0}, KillIndex: 2, Boundary: upgradeChaosBoundaryAfterFirstPostUpgradeCommit},
	}
	for _, plan := range plans {
		plan := plan
		t.Run("seed-"+strconv.FormatInt(plan.Seed, 10), func(t *testing.T) {
			require.NoError(t, plan.Validate(4))
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			runUpgradeBoundaryChaosPlan(t, ctx, plan)
		})
	}
}

func runUpgradeBoundaryChaosPlan(t *testing.T, ctx context.Context, plan upgradeChaosPlan) {
	t.Helper()
	network, err := harness.Start(ctx, t, harness.Config{
		Image:         harness.V221Image(),
		NumValidators: 4,
		NumFullNodes:  1,
		TimeoutCommit: "2s",
	})
	require.NoError(t, err)
	defer network.RecordTestPanic()
	require.NoError(t, network.WriteArtifactJSON("chaos/plan.json", plan))

	oldIdentities := make(map[string]upgradeBinaryIdentity, len(network.Chain.Nodes()))
	for _, node := range network.Chain.Nodes() {
		identity, identityErr := captureUpgradeNodeVersion(ctx, network, node, "chaos-old")
		require.NoError(t, identityErr)
		require.Equal(t, "2.2.1", identity.Version)
		oldIdentities[node.Name()] = identity
	}
	proposer := buildAndFundNFTWallet(t, ctx, network, "upgrade-chaos-proposer")
	upgradeHeight := scheduleChaosUpgrade(t, ctx, network, proposer.KeyName())
	haltEvidence, haltErr := waitForOldBinaryUpgradeHalt(ctx, network, upgradeHeight)
	require.NoError(t, network.WriteArtifactJSON("chaos/old-binary-halt.json", haltEvidence))
	require.NoError(t, haltErr)

	fullNode := network.Chain.FullNodes[0]
	currentImage := harness.CurrentImage()
	fullNodeNeedsRestore := false
	defer restoreUpgradeChaosNodeForCleanup(
		t,
		fullNode,
		&fullNodeNeedsRestore,
		"halted full node",
		func(cleanupCtx context.Context) error {
			_, restoreErr := network.SwitchNodeImage(
				cleanupCtx,
				fmt.Sprintf("chaos-seed-%d-cleanup-halted-full-node", plan.Seed),
				fullNode,
				currentImage,
			)
			return restoreErr
		},
	)
	fullNodeStopStarted := time.Now().UTC()
	fullNodeStopErr := fullNode.StopContainer(ctx)
	fullNodeNeedsRestore = true
	require.NoError(t, network.AppendArtifactJSON("chaos/faults.jsonl", map[string]any{
		"recorded_at": fullNodeStopStarted,
		"fault":       "full-node-stopped-during-upgrade-halt",
		"node":        fullNode.Name(),
		"expected":    "validator consensus remains halted until three new validators run",
		"succeeded":   fullNodeStopErr == nil,
		"error":       errorText(fullNodeStopErr),
	}))
	require.NoError(t, fullNodeStopErr)

	switches := make([]harness.UpgradeNodeSwitchEvidence, 0, 5)
	boundaryObserver := upgradeChaosBoundaryObserver(plan, network.Chain.Validators)
	var boundaryEvidence upgradeChaosBoundaryEvidence
	defer func() {
		if artifactErr := network.WriteArtifactJSON("chaos/boundary-fault-final.json", boundaryEvidence); artifactErr != nil {
			t.Errorf("write final upgrade boundary fault evidence: %v", artifactErr)
		}
	}()
	killNode := network.Chain.Validators[plan.KillIndex]
	killNodeNeedsRestore := false
	defer restoreUpgradeChaosNodeForCleanup(
		t,
		killNode,
		&killNodeNeedsRestore,
		"SIGKILL validator",
		func(cleanupCtx context.Context) error {
			if killNode.Image.Repository == currentImage.Repository && killNode.Image.Version == currentImage.Version {
				return killNode.StartContainer(cleanupCtx)
			}
			_, restoreErr := network.SwitchNodeImage(
				cleanupCtx,
				fmt.Sprintf("chaos-seed-%d-cleanup-sigkill-validator", plan.Seed),
				killNode,
				currentImage,
			)
			return restoreErr
		},
	)
	if plan.Boundary == upgradeChaosBoundaryBeforeUpgradeHandlerCommit {
		killNodeNeedsRestore = true
		boundaryCtx, boundaryCancel := context.WithTimeout(ctx, upgradeChaosBoundaryKillTimeout)
		var boundaryErr error
		boundaryEvidence, boundaryErr = killUpgradeChaosValidatorAtBoundary(
			boundaryCtx,
			network,
			plan,
			upgradeHeight,
			boundaryObserver,
		)
		boundaryCancel()
		require.NoError(t, network.WriteArtifactJSON("chaos/boundary-fault-injection.json", boundaryEvidence))
		require.NoError(t, boundaryErr)
	}

	for position := 0; position < 2; position++ {
		index := plan.SwitchOrder[position]
		switchEvidence, switchErr := network.SwitchNodeImage(
			ctx,
			fmt.Sprintf("chaos-seed-%d-position-%d", plan.Seed, position),
			network.Chain.Validators[index],
			currentImage,
		)
		require.NoError(t, switchErr)
		switches = append(switches, switchEvidence)
	}
	applicationBoundary := make([]map[string]any, 0, 2)
	for _, index := range plan.SwitchOrder[:2] {
		node := network.Chain.Validators[index]
		consensusHeight, applicationHeight, heightErr := upgradeChaosNodeHeights(ctx, node)
		require.NoError(t, heightErr)
		require.Equal(t, upgradeHeight, consensusHeight)
		require.Equal(t, upgradeHeight, applicationHeight)
		applicationBoundary = append(applicationBoundary, map[string]any{
			"node":               node.Name(),
			"consensus_height":   consensusHeight,
			"application_height": applicationHeight,
			"recorded_at":        time.Now().UTC(),
		})
	}
	require.NoError(t, network.WriteArtifactJSON("chaos/application-boundary.json", applicationBoundary))

	observer := network.Chain.Validators[plan.SwitchOrder[0]]
	twoNewStall, err := network.ObserveQuorumStall(
		ctx,
		"two-new-binaries-no-quorum",
		observer,
		2*time.Second,
		3*time.Second,
	)
	require.NoError(t, err)
	require.Equal(t, twoNewStall.StartHeight, twoNewStall.EndHeight)
	require.Equal(t, upgradeHeight, twoNewStall.EndHeight)

	if plan.Boundary == upgradeChaosBoundaryAfterFirstPostUpgradeCommit {
		killNodeNeedsRestore = true
		boundaryCtx, boundaryCancel := context.WithTimeout(ctx, upgradeChaosBoundaryKillTimeout)
		var boundaryErr error
		boundaryEvidence, boundaryErr = killUpgradeChaosValidatorAtBoundary(
			boundaryCtx,
			network,
			plan,
			upgradeHeight,
			boundaryObserver,
		)
		boundaryCancel()
		require.NoError(t, network.WriteArtifactJSON("chaos/boundary-fault-injection.json", boundaryEvidence))
		require.NoError(t, boundaryErr)
	}

	thirdIndex := plan.SwitchOrder[2]
	var (
		thirdSwitch harness.UpgradeNodeSwitchEvidence
		forcedStall harness.QuorumHeightWindow
	)
	if plan.Boundary == upgradeChaosBoundaryBeforeUpgradeHandlerCommit {
		forcedStall, err = network.ObserveQuorumStall(
			ctx,
			"upgrade-boundary-validator-killed-no-quorum",
			observer,
			2*time.Second,
			3*time.Second,
		)
		require.NoError(t, err)
		require.Equal(t, forcedStall.StartHeight, forcedStall.EndHeight)
		require.Equal(t, upgradeHeight, forcedStall.StartHeight)
		boundaryEvidence.RestartRequestedAt = time.Now().UTC()
		var thirdSwitchErr error
		thirdSwitch, thirdSwitchErr = network.SwitchNodeImage(
			ctx,
			fmt.Sprintf("chaos-seed-%d-killed-third-quorum", plan.Seed),
			network.Chain.Validators[thirdIndex],
			currentImage,
		)
		boundaryEvidence.RestartCompletedAt = time.Now().UTC()
		boundaryEvidence.RestartSucceeded = thirdSwitchErr == nil
		if thirdSwitchErr != nil {
			boundaryEvidence.RestartError = thirdSwitchErr.Error()
		} else {
			killNodeNeedsRestore = false
		}
		require.NoError(t, thirdSwitchErr)
	} else {
		boundaryEvidence.RestartRequestedAt = time.Now().UTC()
		forcedStall, thirdSwitch, err = runUpgradeChaosAfterHandlerRecovery(
			func() error {
				restartErr := network.StartQuorumValidator(ctx, "upgrade-boundary-validator-rejoin", plan.KillIndex)
				boundaryEvidence.RestartCompletedAt = time.Now().UTC()
				boundaryEvidence.RestartSucceeded = restartErr == nil
				if restartErr != nil {
					boundaryEvidence.RestartError = restartErr.Error()
				} else {
					killNodeNeedsRestore = false
				}
				return restartErr
			},
			func() (harness.QuorumHeightWindow, error) {
				stall, stallErr := network.ObserveQuorumStall(
					ctx,
					"upgrade-boundary-validator-restarted-no-quorum",
					boundaryObserver,
					2*time.Second,
					3*time.Second,
				)
				if stallErr != nil {
					return stall, stallErr
				}
				if stall.StartHeight != stall.EndHeight || stall.StartHeight != upgradeHeight {
					return stall, fmt.Errorf(
						"two-current-validator recovery window advanced: start=%d end=%d want=%d",
						stall.StartHeight,
						stall.EndHeight,
						upgradeHeight,
					)
				}
				return stall, nil
			},
			func() (harness.UpgradeNodeSwitchEvidence, error) {
				return network.SwitchNodeImage(
					ctx,
					fmt.Sprintf("chaos-seed-%d-third-quorum", plan.Seed),
					network.Chain.Validators[thirdIndex],
					currentImage,
				)
			},
		)
		require.NoError(t, err)
	}
	switches = append(switches, thirdSwitch)

	firstAgreementNodes := []*cosmos.ChainNode{
		network.Chain.Validators[plan.SwitchOrder[0]],
		network.Chain.Validators[plan.SwitchOrder[1]],
		network.Chain.Validators[plan.SwitchOrder[2]],
	}

	delayedIndex := plan.SwitchOrder[3]
	require.Equal(t, "2.2.1", oldIdentities[network.Chain.Validators[delayedIndex].Name()].Version)
	require.NoError(t, network.AppendArtifactJSON("chaos/faults.jsonl", map[string]any{
		"recorded_at": time.Now().UTC(),
		"fault":       "old-binary-validator-left-out-of-post-upgrade-quorum",
		"node":        network.Chain.Validators[delayedIndex].Name(),
		"version":     oldIdentities[network.Chain.Validators[delayedIndex].Name()].Version,
		"expected":    "three correct binaries retain one history while incompatible validator is offline",
	}))

	forcedObserver := boundaryObserver
	boundaryEvidence.RecoveryStartHeight = upgradeHeight

	firstProgressCtx, firstProgressCancel := context.WithTimeout(ctx, upgradeChaosQuorumRecoveryTimeout)
	firstProgress, err := network.WaitForQuorumProgress(
		firstProgressCtx,
		"first-post-upgrade-commit-after-boundary-recovery",
		forcedObserver,
		forcedStall.EndHeight,
		1,
	)
	firstProgressCancel()
	if err != nil {
		boundaryEvidence.RecoveryError = err.Error()
	}
	require.NoError(t, err)
	recoveryCtx, recoveryCancel := context.WithTimeout(ctx, upgradeChaosQuorumRecoveryTimeout)
	forcedRecovery, err := network.WaitForQuorumProgress(
		recoveryCtx,
		"upgrade-boundary-validator-rejoin-progress",
		forcedObserver,
		forcedStall.EndHeight,
		3,
	)
	recoveryCancel()
	if err != nil {
		boundaryEvidence.RecoveryError = err.Error()
	}
	require.NoError(t, err)
	boundaryEvidence.RecoveryEndHeight = forcedRecovery.EndHeight
	boundaryEvidence.RecoveryCompletedAt = time.Now().UTC()
	firstPostObservedAt, firstPostBlockTime, migrationResultAppHash, carrierHeaderAppHash, err := captureUpgradeChaosFirstPostUpgradeBlock(
		ctx,
		killNode,
		upgradeHeight,
	)
	if err != nil {
		boundaryEvidence.RecoveryError = err.Error()
	}
	require.NoError(t, err)
	boundaryEvidence.FirstPostUpgradeHeight = upgradeHeight
	if boundaryEvidence.FirstPostUpgradeObservedAt.IsZero() {
		boundaryEvidence.FirstPostUpgradeObservedAt = firstPostObservedAt
	}
	boundaryEvidence.FirstPostUpgradeBlockTime = firstPostBlockTime
	boundaryEvidence.MigrationResultAppHash = migrationResultAppHash
	boundaryEvidence.CarrierHeaderAppHash = carrierHeaderAppHash
	boundaryEvidence.RecoverySucceeded = true
	require.NoError(t, boundaryEvidence.Validate())
	require.NoError(t, network.WriteArtifactJSON("chaos/boundary-fault-recovery.json", boundaryEvidence))
	_, err = network.WaitForQuorumAgreement(ctx, "first-post-upgrade-history", upgradeHeight, firstAgreementNodes...)
	require.NoError(t, err)
	_, err = network.WaitForQuorumAgreement(ctx, "migration-app-hash-carrier-history", upgradeHeight+1, firstAgreementNodes...)
	require.NoError(t, err)

	fullNodeSwitch, err := network.SwitchNodeImage(
		ctx,
		fmt.Sprintf("chaos-seed-%d-halted-full-node-catchup", plan.Seed),
		fullNode,
		currentImage,
	)
	require.NoError(t, err)
	fullNodeNeedsRestore = false
	switches = append(switches, fullNodeSwitch)
	require.NoError(t, network.WaitForFullNode(ctx, forcedRecovery.EndHeight))

	delayTarget := forcedRecovery.EndHeight + 5
	require.NoError(t, network.WaitForNodeHeight(ctx, forcedObserver, delayTarget))
	delayedSwitch, err := network.SwitchNodeImage(
		ctx,
		fmt.Sprintf("chaos-seed-%d-delayed-validator", plan.Seed),
		network.Chain.Validators[delayedIndex],
		currentImage,
	)
	require.NoError(t, err)
	switches = append(switches, delayedSwitch)
	require.NoError(t, network.WaitForNodeHeight(ctx, network.Chain.Validators[delayedIndex], delayTarget))

	allNodes := append([]*cosmos.ChainNode{}, network.Chain.Validators...)
	allNodes = append(allNodes, fullNode)
	finalHeight, err := network.Chain.Height(ctx)
	require.NoError(t, err)
	if finalHeight > 1 {
		finalHeight--
	}
	preRestartAgreement, err := network.WaitForQuorumAgreement(ctx, "all-upgraded-nodes-one-history", finalHeight, allNodes...)
	require.NoError(t, err)

	nodeVersionsBefore, err := captureUpgradeChaosNodeVersions(ctx, network, "before-restart", allNodes)
	require.NoError(t, err)
	moduleVersions := nodeVersionsBefore[0].Versions
	appliedPlanBefore, err := captureUpgradeChaosAppliedPlan(ctx, network, "before-restart", allNodes)
	require.NoError(t, err)
	upgradedContainerLogs, err := captureUpgradeChaosContainerLogs(ctx, network, allNodes)
	require.NoError(t, err)
	restartEvidence, err := restartUpgradeNetworkWithEvidence(ctx, network, "chaos-all-node", 3)
	require.NoError(t, err)
	nodeVersionsAfter, err := captureUpgradeChaosNodeVersions(ctx, network, "after-restart", allNodes)
	require.NoError(t, err)
	require.Equal(t, nodeVersionsBefore, nodeVersionsAfter)
	appliedPlanAfter, err := captureUpgradeChaosAppliedPlan(ctx, network, "after-restart", allNodes)
	require.NoError(t, err)
	require.Equal(t, appliedPlanBefore, appliedPlanAfter)
	postRestartAgreement, err := network.WaitForQuorumAgreement(ctx, "post-restart-one-history", restartEvidence.TargetHeight, allNodes...)
	require.NoError(t, err)

	evidence := upgradeChaosEvidence{
		Plan:                   plan,
		UpgradeHeight:          upgradeHeight,
		TwoNewBinaryStall:      twoNewStall,
		FirstQuorumProgress:    firstProgress,
		ForcedStopStall:        forcedStall,
		ForcedRecoveryProgress: forcedRecovery,
		BoundaryFault:          boundaryEvidence,
		PreRestartAgreement:    preRestartAgreement,
		FinalAgreement:         postRestartAgreement,
		UpgradedContainerLogs:  upgradedContainerLogs,
		Restart:                restartEvidence,
		ModuleVersions:         moduleVersions,
		NodeVersionsBefore:     nodeVersionsBefore,
		NodeVersionsAfter:      nodeVersionsAfter,
		AppliedPlanBefore:      appliedPlanBefore,
		AppliedPlanAfter:       appliedPlanAfter,
		Switches:               switches,
	}
	require.NoError(t, evidence.Validate())
	require.NoError(t, network.WriteArtifactJSON("chaos/recovery.json", evidence))
}

func captureUpgradeChaosNodeVersions(
	ctx context.Context,
	network *harness.Network,
	phase string,
	nodes []*cosmos.ChainNode,
) ([]upgradeChaosNodeVersionEvidence, error) {
	if len(nodes) == 0 {
		return nil, errors.New("upgrade chaos module-version capture requires nodes")
	}
	result := make([]upgradeChaosNodeVersionEvidence, 0, len(nodes))
	var baseline map[string]uint64
	for _, node := range nodes {
		stdout, stderr, err := node.ExecQuery(ctx, "upgrade", "module-versions")
		if err != nil {
			return nil, fmt.Errorf("query %s module versions on %s: %w: %s", phase, node.Name(), err, string(stderr))
		}
		versions, err := decodeUpgradeModuleVersions(stdout)
		if err != nil {
			return nil, fmt.Errorf("decode %s module versions on %s: %w", phase, node.Name(), err)
		}
		if baseline == nil {
			baseline = versions
		} else if !mapsEqualUpgradeModuleVersions(baseline, versions) {
			return nil, fmt.Errorf("%s module versions on %s differ from the first upgraded node", phase, node.Name())
		}
		result = append(result, upgradeChaosNodeVersionEvidence{Node: node.Name(), Versions: versions})
	}
	if err := network.WriteArtifactJSON("chaos/module-versions-"+phase+".json", result); err != nil {
		return nil, err
	}
	return result, nil
}

func captureUpgradeChaosAppliedPlan(
	ctx context.Context,
	network *harness.Network,
	phase string,
	nodes []*cosmos.ChainNode,
) ([]upgradeChaosAppliedPlanEvidence, error) {
	if len(nodes) == 0 {
		return nil, errors.New("upgrade chaos applied-plan capture requires nodes")
	}
	result := make([]upgradeChaosAppliedPlanEvidence, 0, len(nodes))
	var baseline json.RawMessage
	for _, node := range nodes {
		stdout, stderr, err := node.ExecQuery(ctx, "upgrade", "applied", upgradeName)
		if err != nil {
			return nil, fmt.Errorf("query %s applied upgrade on %s: %w: %s", phase, node.Name(), err, string(stderr))
		}
		semantic, err := harness.NewSemanticJSON(stdout)
		if err != nil {
			return nil, fmt.Errorf("decode %s applied upgrade on %s: %w", phase, node.Name(), err)
		}
		if baseline == nil {
			baseline = json.RawMessage(semantic)
		} else if string(baseline) != string(semantic) {
			return nil, fmt.Errorf("%s applied upgrade result on %s differs from the first upgraded node", phase, node.Name())
		}
		result = append(result, upgradeChaosAppliedPlanEvidence{Node: node.Name(), Response: json.RawMessage(semantic)})
	}
	if err := network.WriteArtifactJSON("chaos/applied-plan-"+phase+".json", result); err != nil {
		return nil, err
	}
	return result, nil
}

func mapsEqualUpgradeModuleVersions(left, right map[string]uint64) bool {
	if len(left) != len(right) {
		return false
	}
	for name, version := range left {
		if right[name] != version {
			return false
		}
	}
	return true
}

func scheduleChaosUpgrade(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	proposerKey string,
) int64 {
	t.Helper()
	baseHeight, err := network.Chain.Height(ctx)
	require.NoError(t, err)
	upgradeHeight := baseHeight + 55
	proposalTx, err := network.BroadcastAndWaitTx(
		ctx,
		"chaos-upgrade-submit-proposal",
		network.Chain.Validators[0],
		proposerKey,
		"gov", "submit-legacy-proposal", "software-upgrade", upgradeName,
		"--title", "Panacea v2.3.0 chaos upgrade",
		"--description", "Upgrade boundary fault injection",
		"--deposit", "1umed",
		"--upgrade-height", strconv.FormatInt(upgradeHeight, 10),
		"--upgrade-info", "{}",
		"--no-validate",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	proposalID, err := proposalIDFromCommittedTx(proposalTx)
	require.NoError(t, err)
	for index, validator := range network.Chain.Validators {
		_, voteErr := network.BroadcastAndWaitTx(
			ctx,
			"chaos-upgrade-vote-"+strconv.Itoa(index),
			validator,
			"validator",
			"gov", "vote", strconv.FormatUint(proposalID, 10), "yes",
			"--gas", "500000",
			"--broadcast-mode", "sync",
		)
		require.NoError(t, voteErr)
	}
	require.NoError(t, waitForProposalPassed(ctx, network, proposalID))
	require.NoError(t, network.WriteArtifactJSON("chaos/upgrade-plan.json", map[string]any{
		"proposal_id":    proposalID,
		"upgrade_height": upgradeHeight,
		"tx_hash":        proposalTx.TxHash,
	}))
	return upgradeHeight
}

func decodeUpgradeChaosPlan(raw []byte) (upgradeChaosPlan, error) {
	var plan upgradeChaosPlan
	if err := json.Unmarshal(raw, &plan); err != nil {
		return upgradeChaosPlan{}, fmt.Errorf("decode upgrade chaos plan: %w", err)
	}
	if err := plan.Validate(4); err != nil {
		return upgradeChaosPlan{}, err
	}
	return plan, nil
}
