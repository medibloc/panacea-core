package e2e_test

import (
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"
)

type upgradeHaltObservation struct {
	Node                string `json:"node"`
	ConsensusHeight     int64  `json:"consensus_height"`
	ConsensusBlockID    string `json:"consensus_block_id"`
	ConsensusAppHash    string `json:"consensus_app_hash"`
	ApplicationHeight   int64  `json:"application_height"`
	ApplicationHash     string `json:"application_hash"`
	PlanName            string `json:"plan_name"`
	PlanHeight          int64  `json:"plan_height"`
	ContainerRunning    bool   `json:"container_running"`
	ContainerStatus     string `json:"container_status"`
	ContainerPaused     bool   `json:"container_paused"`
	ContainerRestarting bool   `json:"container_restarting"`
	ContainerOOMKilled  bool   `json:"container_oom_killed"`
	ContainerDead       bool   `json:"container_dead"`
}

type upgradeHaltTracker struct {
	expectedPlan   string
	expectedHeight int64
	window         time.Duration
	stableSince    time.Time
	lastObservedAt time.Time
	expectedNodes  map[string]struct{}
	evidence       map[string]upgradeHaltObservation
}

func newUpgradeHaltTracker(
	expectedPlan string,
	expectedHeight int64,
	window time.Duration,
) (*upgradeHaltTracker, error) {
	if strings.TrimSpace(expectedPlan) == "" {
		return nil, errors.New("upgrade halt plan name is required")
	}
	if expectedHeight <= 0 {
		return nil, errors.New("upgrade halt height must be positive")
	}
	if window <= 0 {
		return nil, errors.New("upgrade halt stability window must be positive")
	}
	return &upgradeHaltTracker{
		expectedPlan:   expectedPlan,
		expectedHeight: expectedHeight,
		window:         window,
	}, nil
}

func (t *upgradeHaltTracker) Observe(
	now time.Time,
	observations map[string]upgradeHaltObservation,
	probeErr error,
) (complete bool, err error) {
	defer func() {
		if err != nil {
			t.stableSince = time.Time{}
		}
	}()
	if !t.lastObservedAt.IsZero() && now.Before(t.lastObservedAt) {
		return false, fmt.Errorf(
			"upgrade halt observation time moved backward from %s to %s",
			t.lastObservedAt.Format(time.RFC3339Nano),
			now.Format(time.RFC3339Nano),
		)
	}
	t.lastObservedAt = now
	if probeErr != nil || len(observations) == 0 {
		t.stableSince = time.Time{}
		return false, nil
	}
	if t.expectedNodes != nil && !sameUpgradeHaltNodeSet(t.expectedNodes, observations) {
		t.stableSince = time.Time{}
		return false, nil
	}
	var (
		commitmentNode   string
		consensusBlockID string
		applicationHash  string
		beforeBoundary   bool
	)
	for node, observation := range observations {
		if observation.Node != node {
			return false, fmt.Errorf(
				"observation key %q does not match node %q",
				node,
				observation.Node,
			)
		}
		if !observation.ContainerRunning || observation.ContainerStatus != "running" ||
			observation.ContainerPaused || observation.ContainerRestarting ||
			observation.ContainerOOMKilled || observation.ContainerDead {
			return false, fmt.Errorf(
				"node %s has unhealthy Docker lifecycle: running=%t status=%q paused=%t restarting=%t oom_killed=%t dead=%t",
				node,
				observation.ContainerRunning,
				observation.ContainerStatus,
				observation.ContainerPaused,
				observation.ContainerRestarting,
				observation.ContainerOOMKilled,
				observation.ContainerDead,
			)
		}
		if observation.ConsensusHeight > t.expectedHeight {
			return false, fmt.Errorf(
				"node %s consensus height %d advanced past upgrade height %d",
				node,
				observation.ConsensusHeight,
				t.expectedHeight,
			)
		}
		if observation.ConsensusHeight < t.expectedHeight {
			beforeBoundary = true
		}
		expectedApplicationHeight := t.expectedHeight - 1
		if observation.ApplicationHeight > expectedApplicationHeight {
			return false, fmt.Errorf(
				"node %s application height %d advanced past pre-upgrade height %d",
				node,
				observation.ApplicationHeight,
				expectedApplicationHeight,
			)
		}
		if observation.ConsensusHeight == t.expectedHeight &&
			observation.ApplicationHeight != expectedApplicationHeight {
			return false, fmt.Errorf(
				"node %s application height %d, want exact pre-upgrade height %d",
				node,
				observation.ApplicationHeight,
				expectedApplicationHeight,
			)
		}
		if observation.PlanName != t.expectedPlan || observation.PlanHeight != t.expectedHeight {
			return false, fmt.Errorf(
				"node %s upgrade plan %q at %d, want %q at %d",
				node,
				observation.PlanName,
				observation.PlanHeight,
				t.expectedPlan,
				t.expectedHeight,
			)
		}
		if strings.TrimSpace(observation.ConsensusBlockID) == "" {
			return false, fmt.Errorf("node %s consensus block ID is empty", node)
		}
		if strings.TrimSpace(observation.ConsensusAppHash) == "" {
			return false, fmt.Errorf("node %s consensus block app hash is empty", node)
		}
		if strings.TrimSpace(observation.ApplicationHash) == "" {
			return false, fmt.Errorf("node %s application hash is empty", node)
		}
		if !strings.EqualFold(observation.ConsensusAppHash, observation.ApplicationHash) {
			return false, fmt.Errorf(
				"node %s consensus block %d app hash %s differs from application hash %s at height %d",
				node,
				t.expectedHeight,
				observation.ConsensusAppHash,
				observation.ApplicationHash,
				observation.ApplicationHeight,
			)
		}
		if applicationHash == "" {
			commitmentNode = node
			consensusBlockID = observation.ConsensusBlockID
			applicationHash = observation.ApplicationHash
		} else if !strings.EqualFold(consensusBlockID, observation.ConsensusBlockID) ||
			!strings.EqualFold(applicationHash, observation.ApplicationHash) {
			return false, fmt.Errorf(
				"node %s boundary block/app commitment %s/%s differs from node %s commitment %s/%s",
				node,
				observation.ConsensusBlockID,
				observation.ApplicationHash,
				commitmentNode,
				consensusBlockID,
				applicationHash,
			)
		}
	}
	if t.expectedNodes == nil {
		t.expectedNodes = make(map[string]struct{}, len(observations))
		for node := range observations {
			t.expectedNodes[node] = struct{}{}
		}
	}
	if beforeBoundary {
		t.stableSince = time.Time{}
		return false, nil
	}
	for node, observation := range observations {
		previous, ok := t.evidence[node]
		if !ok {
			continue
		}
		if !strings.EqualFold(previous.ConsensusBlockID, observation.ConsensusBlockID) ||
			!strings.EqualFold(previous.ConsensusAppHash, observation.ConsensusAppHash) ||
			!strings.EqualFold(previous.ApplicationHash, observation.ApplicationHash) {
			return false, fmt.Errorf(
				"node %s upgrade-boundary commitment changed during the stability window",
				node,
			)
		}
	}
	if t.stableSince.IsZero() || !maps.Equal(t.evidence, observations) {
		t.stableSince = now
	}
	t.evidence = cloneUpgradeHaltObservations(observations)
	return now.Sub(t.stableSince) >= t.window, nil
}

func sameUpgradeHaltNodeSet(
	expected map[string]struct{},
	observations map[string]upgradeHaltObservation,
) bool {
	if len(expected) != len(observations) {
		return false
	}
	for node := range expected {
		if _, ok := observations[node]; !ok {
			return false
		}
	}
	return true
}

func (t *upgradeHaltTracker) Evidence() map[string]upgradeHaltObservation {
	if t == nil {
		return nil
	}
	return cloneUpgradeHaltObservations(t.evidence)
}

func cloneUpgradeHaltObservations(
	observations map[string]upgradeHaltObservation,
) map[string]upgradeHaltObservation {
	if observations == nil {
		return nil
	}
	cloned := make(map[string]upgradeHaltObservation, len(observations))
	for node, observation := range observations {
		cloned[node] = observation
	}
	return cloned
}
