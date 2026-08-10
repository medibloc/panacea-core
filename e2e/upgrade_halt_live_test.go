package e2e_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	upgradeHaltWaitTimeout     = 4 * time.Minute
	upgradeHaltProbeTimeout    = 5 * time.Second
	upgradeHaltPollInterval    = 500 * time.Millisecond
	upgradeHaltStabilityWindow = 5 * time.Second
)

type oldBinaryUpgradePlan struct {
	Name   string `json:"name"`
	Height int64  `json:"height"`
}

type oldBinaryUpgradeHaltEvidence struct {
	UpgradeHeight             int64                             `json:"upgrade_height"`
	ExpectedApplicationHeight int64                             `json:"expected_application_height"`
	StabilityWindow           string                            `json:"stability_window"`
	BoundaryObservedAt        time.Time                         `json:"boundary_observed_at,omitempty"`
	RecordedAt                time.Time                         `json:"recorded_at"`
	Nodes                     map[string]upgradeHaltObservation `json:"nodes,omitempty"`
	LastTransientProbeError   string                            `json:"last_transient_probe_error,omitempty"`
}

func waitForOldBinaryUpgradeHalt(
	ctx context.Context,
	network *harness.Network,
	upgradeHeight int64,
) (oldBinaryUpgradeHaltEvidence, error) {
	evidence := oldBinaryUpgradeHaltEvidence{
		UpgradeHeight:             upgradeHeight,
		ExpectedApplicationHeight: upgradeHeight - 1,
		StabilityWindow:           upgradeHaltStabilityWindow.String(),
	}
	if network == nil || network.Chain == nil {
		return evidence, errors.New("upgrade halt network is unavailable")
	}
	nodes := network.Chain.Nodes()
	if len(nodes) == 0 {
		return evidence, errors.New("upgrade halt requires at least one node")
	}
	nodeNames := make(map[string]struct{}, len(nodes))
	for index, node := range nodes {
		if node == nil {
			return evidence, fmt.Errorf("upgrade halt node %d is nil", index)
		}
		name := strings.TrimSpace(node.Name())
		if name == "" {
			return evidence, fmt.Errorf("upgrade halt node %d has no name", index)
		}
		if _, exists := nodeNames[name]; exists {
			return evidence, fmt.Errorf("upgrade halt node name %q is duplicated", name)
		}
		nodeNames[name] = struct{}{}
		if node.Client == nil {
			return evidence, fmt.Errorf("upgrade halt node %s has no RPC client", name)
		}
		if node.DockerClient == nil {
			return evidence, fmt.Errorf("upgrade halt node %s has no Docker client", name)
		}
	}

	waitCtx, cancel := context.WithTimeout(ctx, upgradeHaltWaitTimeout)
	defer cancel()
	if err := waitForOldBinaryConsensusBoundary(waitCtx, nodes, upgradeHeight); err != nil {
		return evidence, err
	}
	plans, err := waitForOldBinaryUpgradePlans(waitCtx, nodes, upgradeName, upgradeHeight)
	if err != nil {
		return evidence, err
	}
	if err := recordRawOldBinaryUpgradeInfo(waitCtx, network, nodes); err != nil {
		return evidence, err
	}
	evidence.BoundaryObservedAt = time.Now().UTC()

	tracker, err := newUpgradeHaltTracker(upgradeName, upgradeHeight, upgradeHaltStabilityWindow)
	if err != nil {
		return evidence, err
	}
	ticker := time.NewTicker(upgradeHaltPollInterval)
	defer ticker.Stop()
	for {
		observations, probeErr := probeOldBinaryUpgradeHalt(waitCtx, nodes, plans)
		now := time.Now().UTC()
		complete, trackErr := tracker.Observe(now, observations, probeErr)
		evidence.RecordedAt = now
		evidence.Nodes = tracker.Evidence()
		if probeErr != nil {
			evidence.LastTransientProbeError = probeErr.Error()
		}
		pollEvidence := map[string]any{
			"recorded_at":  now,
			"complete":     complete,
			"observations": observations,
			"stable_nodes": evidence.Nodes,
		}
		if probeErr != nil {
			pollEvidence["probe_error"] = probeErr.Error()
		}
		if trackErr != nil {
			pollEvidence["validation_error"] = trackErr.Error()
		}
		if err := network.AppendArtifactJSON("upgrade/halt-stability.jsonl", pollEvidence); err != nil {
			return evidence, fmt.Errorf("record old binary halt stability poll: %w", err)
		}
		if trackErr != nil {
			return evidence, fmt.Errorf("validate old binary upgrade halt: %w", trackErr)
		}
		if complete {
			return evidence, nil
		}

		select {
		case <-waitCtx.Done():
			return evidence, fmt.Errorf(
				"old binary upgrade halt at height %d was not stable for %s: %w",
				upgradeHeight,
				upgradeHaltStabilityWindow,
				waitCtx.Err(),
			)
		case <-ticker.C:
		}
	}
}

func recordRawOldBinaryUpgradeInfo(
	ctx context.Context,
	network *harness.Network,
	nodes []*cosmos.ChainNode,
) error {
	type rawUpgradeInfoEvidence struct {
		Node   string `json:"node"`
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
		Bytes  int    `json:"bytes"`
	}
	records := make([]rawUpgradeInfoEvidence, 0, len(nodes))
	for index, node := range nodes {
		contents, err := node.ReadFile(ctx, "data/upgrade-info.json")
		if err != nil {
			return fmt.Errorf("read raw upgrade-info.json for %s: %w", node.Name(), err)
		}
		if !json.Valid(contents) {
			return fmt.Errorf("raw upgrade-info.json for %s is invalid JSON", node.Name())
		}
		relativePath := fmt.Sprintf("upgrade/raw-upgrade-info/node-%d.json", index)
		if err := network.WriteArtifact(relativePath, contents); err != nil {
			return fmt.Errorf("record raw upgrade-info.json for %s: %w", node.Name(), err)
		}
		digest := sha256.Sum256(contents)
		records = append(records, rawUpgradeInfoEvidence{
			Node:   node.Name(),
			Path:   relativePath,
			SHA256: fmt.Sprintf("%x", digest[:]),
			Bytes:  len(contents),
		})
	}
	if err := network.WriteArtifactJSON("upgrade/raw-upgrade-info.json", records); err != nil {
		return fmt.Errorf("record raw upgrade-info index: %w", err)
	}
	return nil
}

func waitForOldBinaryConsensusBoundary(
	ctx context.Context,
	nodes []*cosmos.ChainNode,
	upgradeHeight int64,
) error {
	ticker := time.NewTicker(upgradeHaltPollInterval)
	defer ticker.Stop()
	var lastHeights map[string]int64
	var lastProbeErr error
	for {
		heights, probeErr := probeOldBinaryConsensusHeights(ctx, nodes)
		lastHeights = heights
		lastProbeErr = probeErr
		allAtBoundary := probeErr == nil && len(heights) == len(nodes)
		for node, height := range heights {
			if height > upgradeHeight {
				return fmt.Errorf(
					"old binary %s advanced to consensus height %d past upgrade height %d",
					node,
					height,
					upgradeHeight,
				)
			}
			allAtBoundary = allAtBoundary && height == upgradeHeight
		}
		if allAtBoundary {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf(
				"wait for old binaries at consensus height %d; last heights %v; last probe error %v: %w",
				upgradeHeight,
				lastHeights,
				lastProbeErr,
				ctx.Err(),
			)
		case <-ticker.C:
		}
	}
}

func probeOldBinaryConsensusHeights(
	ctx context.Context,
	nodes []*cosmos.ChainNode,
) (map[string]int64, error) {
	type result struct {
		node   string
		height int64
		err    error
	}
	results := make(chan result, len(nodes))
	for _, node := range nodes {
		go func(node *cosmos.ChainNode) {
			probeCtx, cancel := context.WithTimeout(ctx, upgradeHaltProbeTimeout)
			defer cancel()
			height, err := node.Height(probeCtx)
			results <- result{node: node.Name(), height: height, err: err}
		}(node)
	}

	heights := make(map[string]int64, len(nodes))
	var probeErrors []error
	for range nodes {
		observed := <-results
		if observed.err != nil {
			probeErrors = append(probeErrors, fmt.Errorf("probe %s consensus height: %w", observed.node, observed.err))
			continue
		}
		heights[observed.node] = observed.height
	}
	return heights, errors.Join(probeErrors...)
}

func waitForOldBinaryUpgradePlans(
	ctx context.Context,
	nodes []*cosmos.ChainNode,
	expectedName string,
	expectedHeight int64,
) (map[string]oldBinaryUpgradePlan, error) {
	ticker := time.NewTicker(upgradeHaltPollInterval)
	defer ticker.Stop()
	var lastPlans map[string]oldBinaryUpgradePlan
	var lastProbeErr error
	for {
		plans, retryErr, validationErr := probeOldBinaryUpgradePlans(ctx, nodes, expectedName, expectedHeight)
		lastPlans = plans
		lastProbeErr = retryErr
		if validationErr != nil {
			return plans, validationErr
		}
		if retryErr == nil && len(plans) == len(nodes) {
			return plans, nil
		}

		select {
		case <-ctx.Done():
			return lastPlans, fmt.Errorf(
				"wait for old binary upgrade-info %q at height %d; last plans %v; last probe error %v: %w",
				expectedName,
				expectedHeight,
				lastPlans,
				lastProbeErr,
				ctx.Err(),
			)
		case <-ticker.C:
		}
	}
}

func probeOldBinaryUpgradePlans(
	ctx context.Context,
	nodes []*cosmos.ChainNode,
	expectedName string,
	expectedHeight int64,
) (map[string]oldBinaryUpgradePlan, error, error) {
	type result struct {
		node        string
		plan        oldBinaryUpgradePlan
		retryErr    error
		validateErr error
	}
	results := make(chan result, len(nodes))
	for _, node := range nodes {
		go func(node *cosmos.ChainNode) {
			probeCtx, cancel := context.WithTimeout(ctx, upgradeHaltProbeTimeout)
			defer cancel()
			contents, err := node.ReadFile(probeCtx, "data/upgrade-info.json")
			if err != nil {
				results <- result{node: node.Name(), retryErr: err}
				return
			}
			plan, retryErr, validationErr := decodeObservedOldBinaryUpgradePlan(
				contents,
				expectedName,
				expectedHeight,
			)
			results <- result{
				node:        node.Name(),
				plan:        plan,
				retryErr:    retryErr,
				validateErr: validationErr,
			}
		}(node)
	}

	plans := make(map[string]oldBinaryUpgradePlan, len(nodes))
	var retryErrors []error
	var validationErrors []error
	for range nodes {
		observed := <-results
		if observed.retryErr != nil {
			retryErrors = append(retryErrors, fmt.Errorf("read %s upgrade-info.json: %w", observed.node, observed.retryErr))
			continue
		}
		if observed.plan.Name != "" {
			plans[observed.node] = observed.plan
		}
		if observed.validateErr != nil {
			validationErrors = append(validationErrors, fmt.Errorf("validate %s upgrade-info.json: %w", observed.node, observed.validateErr))
		}
	}
	return plans, errors.Join(retryErrors...), errors.Join(validationErrors...)
}

func decodeObservedOldBinaryUpgradePlan(
	contents []byte,
	expectedName string,
	expectedHeight int64,
) (oldBinaryUpgradePlan, error, error) {
	name, height, err := decodeUpgradeInfo(contents)
	if err != nil {
		if isIncompleteUpgradeInfo(contents, err) {
			return oldBinaryUpgradePlan{}, fmt.Errorf("decode incomplete upgrade-info.json: %w", err), nil
		}
		return oldBinaryUpgradePlan{}, nil, fmt.Errorf("decode invalid upgrade-info.json: %w", err)
	}
	plan := oldBinaryUpgradePlan{Name: name, Height: height}
	if name != expectedName || height != expectedHeight {
		return plan, nil, fmt.Errorf(
			"plan %q at %d, want %q at %d",
			name,
			height,
			expectedName,
			expectedHeight,
		)
	}
	return plan, nil, nil
}

func isIncompleteUpgradeInfo(contents []byte, err error) bool {
	if strings.TrimSpace(string(contents)) == "" {
		return true
	}
	var syntaxErr *json.SyntaxError
	return errors.As(err, &syntaxErr) && syntaxErr.Error() == "unexpected end of JSON input"
}

func probeOldBinaryUpgradeHalt(
	ctx context.Context,
	nodes []*cosmos.ChainNode,
	plans map[string]oldBinaryUpgradePlan,
) (map[string]upgradeHaltObservation, error) {
	type result struct {
		node        string
		observation upgradeHaltObservation
		err         error
	}
	results := make(chan result, len(nodes))
	for _, node := range nodes {
		go func(node *cosmos.ChainNode) {
			name := node.Name()
			plan, ok := plans[name]
			if !ok {
				results <- result{node: name, err: errors.New("cached upgrade plan is missing")}
				return
			}

			statusCtx, statusCancel := context.WithTimeout(ctx, upgradeHaltProbeTimeout)
			statusResult, err := node.Client.Status(statusCtx)
			statusCancel()
			if err != nil {
				results <- result{node: name, err: fmt.Errorf("query consensus status: %w", err)}
				return
			}
			if statusResult == nil {
				results <- result{node: name, err: errors.New("consensus status response is nil")}
				return
			}
			consensusHeight := statusResult.SyncInfo.LatestBlockHeight
			if consensusHeight <= 0 {
				results <- result{node: name, err: fmt.Errorf("consensus height is invalid: %d", consensusHeight)}
				return
			}
			blockCtx, blockCancel := context.WithTimeout(ctx, upgradeHaltProbeTimeout)
			blockResult, err := node.Client.Block(blockCtx, &consensusHeight)
			blockCancel()
			if err != nil {
				results <- result{node: name, err: fmt.Errorf("query consensus block %d: %w", consensusHeight, err)}
				return
			}
			if blockResult == nil || blockResult.Block == nil {
				results <- result{node: name, err: fmt.Errorf("consensus block %d response is nil", consensusHeight)}
				return
			}
			if blockResult.Block.Height != consensusHeight {
				results <- result{node: name, err: fmt.Errorf(
					"consensus block response height %d, want %d",
					blockResult.Block.Height,
					consensusHeight,
				)}
				return
			}
			applicationCtx, applicationCancel := context.WithTimeout(ctx, upgradeHaltProbeTimeout)
			applicationResult, err := node.Client.ABCIInfo(applicationCtx)
			applicationCancel()
			if err != nil {
				results <- result{node: name, err: fmt.Errorf("query ABCI info: %w", err)}
				return
			}
			if applicationResult == nil {
				results <- result{node: name, err: errors.New("ABCI info response is nil")}
				return
			}

			inspectCtx, inspectCancel := context.WithTimeout(ctx, upgradeHaltProbeTimeout)
			inspect, err := node.DockerClient.ContainerInspect(inspectCtx, node.ContainerID())
			inspectCancel()
			if err != nil {
				results <- result{node: name, err: fmt.Errorf("inspect container: %w", err)}
				return
			}
			if inspect.State == nil {
				results <- result{node: name, err: errors.New("container state is nil")}
				return
			}

			results <- result{
				node: name,
				observation: upgradeHaltObservation{
					Node:                name,
					ConsensusHeight:     consensusHeight,
					ConsensusBlockID:    strings.ToUpper(fmt.Sprintf("%X", []byte(blockResult.BlockID.Hash))),
					ConsensusAppHash:    strings.ToUpper(fmt.Sprintf("%X", []byte(blockResult.Block.AppHash))),
					ApplicationHeight:   applicationResult.Response.LastBlockHeight,
					ApplicationHash:     strings.ToUpper(fmt.Sprintf("%X", applicationResult.Response.LastBlockAppHash)),
					PlanName:            plan.Name,
					PlanHeight:          plan.Height,
					ContainerRunning:    inspect.State.Running,
					ContainerStatus:     inspect.State.Status,
					ContainerPaused:     inspect.State.Paused,
					ContainerRestarting: inspect.State.Restarting,
					ContainerOOMKilled:  inspect.State.OOMKilled,
					ContainerDead:       inspect.State.Dead,
				},
			}
		}(node)
	}

	observations := make(map[string]upgradeHaltObservation, len(nodes))
	var probeErrors []error
	for range nodes {
		observed := <-results
		if observed.err != nil {
			probeErrors = append(probeErrors, fmt.Errorf("probe old binary halt for %s: %w", observed.node, observed.err))
			continue
		}
		observations[observed.node] = observed.observation
	}
	return observations, errors.Join(probeErrors...)
}
