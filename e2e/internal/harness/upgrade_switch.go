package harness

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

type upgradeSwitchPlan struct {
	Node string          `json:"node"`
	From ibc.DockerImage `json:"from"`
	To   ibc.DockerImage `json:"to"`
}

type upgradeSwitchOperations struct {
	capture func() error
	stop    func() error
	remove  func() error
	create  func() error
	start   func() error
}

type upgradeSwitchOperationEvent struct {
	RecordedAt time.Time `json:"recorded_at"`
	Phase      string    `json:"phase"`
	Node       string    `json:"node"`
	Operation  string    `json:"operation"`
	State      string    `json:"state"`
	Error      string    `json:"error,omitempty"`
}

type upgradeBatchSwitch struct {
	wasRunning bool
	operations upgradeSwitchOperations
}

func runUpgradePrepareOperations(wasRunning bool, operations upgradeSwitchOperations) error {
	if operations.capture == nil || operations.remove == nil || operations.create == nil {
		return errors.New("upgrade switch operations are incomplete")
	}
	if wasRunning && operations.stop == nil {
		return errors.New("upgrade switch stop operation is required for a running container")
	}
	if err := operations.capture(); err != nil {
		return fmt.Errorf("capture old container log: %w", err)
	}
	if wasRunning {
		if err := operations.stop(); err != nil {
			return fmt.Errorf("stop old container: %w", err)
		}
	}
	if err := operations.remove(); err != nil {
		return fmt.Errorf("remove old container: %w", err)
	}
	if err := operations.create(); err != nil {
		return fmt.Errorf("create upgraded container: %w", err)
	}
	return nil
}

func runUpgradeSwitchOperations(wasRunning bool, operations upgradeSwitchOperations) error {
	if operations.start == nil {
		return errors.New("upgrade switch start operation is incomplete")
	}
	if err := runUpgradePrepareOperations(wasRunning, operations); err != nil {
		return err
	}
	if err := operations.start(); err != nil {
		return fmt.Errorf("start upgraded container: %w", err)
	}
	return nil
}

// runUpgradeBatchSwitchOperations creates every replacement container before
// starting any of them. Starts run concurrently so equal-power validators can
// discover peers and form the post-upgrade quorum while StartContainer waits
// for catching_up=false.
func runUpgradeBatchSwitchOperations(switches []upgradeBatchSwitch) []error {
	results := make([]error, len(switches))
	for index, planned := range switches {
		if planned.operations.start == nil {
			results[index] = errors.New("upgrade switch start operation is incomplete")
		}
	}
	if errors.Join(results...) != nil {
		return results
	}
	preparationFailed := false
	for index, planned := range switches {
		if preparationFailed {
			results[index] = errors.New("upgrade batch preparation skipped after an earlier failure")
			continue
		}
		if err := runUpgradePrepareOperations(planned.wasRunning, planned.operations); err != nil {
			results[index] = err
			preparationFailed = true
		}
	}
	if preparationFailed {
		return results
	}

	var wait sync.WaitGroup
	wait.Add(len(switches))
	for index := range switches {
		go func(index int) {
			defer wait.Done()
			if switches[index].operations.start == nil {
				results[index] = errors.New("upgrade switch start operation is incomplete")
				return
			}
			if err := switches[index].operations.start(); err != nil {
				results[index] = fmt.Errorf("start upgraded container: %w", err)
			}
		}(index)
	}
	wait.Wait()
	return results
}

// UpgradeNodeSwitchEvidence is the durable record of one container image
// replacement over an unchanged Interchaintest node volume and identity.
type UpgradeNodeSwitchEvidence struct {
	RecordedAt  time.Time         `json:"recorded_at"`
	CompletedAt time.Time         `json:"completed_at"`
	Phase       string            `json:"phase"`
	Plan        upgradeSwitchPlan `json:"plan"`
	WasRunning  bool              `json:"was_running"`
	OldImageID  string            `json:"old_image_id"`
	NewImageID  string            `json:"new_image_id"`
	Error       string            `json:"error,omitempty"`
}

type plannedUpgradeNodeSwitch struct {
	evidence   UpgradeNodeSwitchEvidence
	operations upgradeSwitchOperations
}

func newUpgradeSwitchPlan(nodeName string, current ibc.DockerImage, target ImageRef) (upgradeSwitchPlan, error) {
	if strings.TrimSpace(nodeName) == "" {
		return upgradeSwitchPlan{}, errors.New("upgrade node name is required")
	}
	if strings.TrimSpace(target.Repository) == "" || strings.TrimSpace(target.Version) == "" {
		return upgradeSwitchPlan{}, errors.New("upgrade target image repository and version are required")
	}
	if current.Repository == target.Repository && current.Version == target.Version {
		return upgradeSwitchPlan{}, fmt.Errorf("node %s already uses upgrade image %s:%s", nodeName, target.Repository, target.Version)
	}
	next := current
	next.Repository = target.Repository
	next.Version = target.Version
	return upgradeSwitchPlan{Node: nodeName, From: current, To: next}, nil
}

// SwitchNodeImage replaces one stopped or running node container while
// retaining its volume, validator key, node key, and CometBFT state. It never
// duplicates the validator process: the old container is removed before the
// new image is created and started.
func (n *Network) SwitchNodeImage(
	ctx context.Context,
	phase string,
	node *cosmos.ChainNode,
	target ImageRef,
) (UpgradeNodeSwitchEvidence, error) {
	planned, err := n.planUpgradeNodeSwitch(ctx, phase, node, target)
	if err != nil {
		return UpgradeNodeSwitchEvidence{}, err
	}
	err = runUpgradeSwitchOperations(planned.evidence.WasRunning, planned.operations)
	if err != nil {
		err = fmt.Errorf("switch image for %s: %w", node.Name(), err)
	}
	return n.finishUpgradeNodeSwitch(planned.evidence, err)
}

// SwitchNodeImagesTogether replaces a quorum set in two phases: all old
// containers are removed and all new containers are created first, then all
// replacements start concurrently. No old and new process for one validator
// key can overlap.
func (n *Network) SwitchNodeImagesTogether(
	ctx context.Context,
	phase string,
	nodes []*cosmos.ChainNode,
	target ImageRef,
) ([]UpgradeNodeSwitchEvidence, error) {
	if len(nodes) == 0 {
		return nil, errors.New("upgrade switch batch requires at least one node")
	}
	planned := make([]plannedUpgradeNodeSwitch, len(nodes))
	seen := make(map[string]struct{}, len(nodes))
	for index, node := range nodes {
		if node == nil {
			return nil, fmt.Errorf("upgrade switch batch node %d is nil", index)
		}
		if _, duplicate := seen[node.Name()]; duplicate {
			return nil, fmt.Errorf("upgrade switch batch contains duplicate node %s", node.Name())
		}
		seen[node.Name()] = struct{}{}
		item, err := n.planUpgradeNodeSwitch(ctx, phase, node, target)
		if err != nil {
			return nil, err
		}
		planned[index] = item
	}

	batch := make([]upgradeBatchSwitch, len(planned))
	for index, item := range planned {
		batch[index] = upgradeBatchSwitch{
			wasRunning: item.evidence.WasRunning,
			operations: item.operations,
		}
	}
	operationErrors := runUpgradeBatchSwitchOperations(batch)
	evidence := make([]UpgradeNodeSwitchEvidence, len(planned))
	var switchErrors []error
	for index, item := range planned {
		operationErr := operationErrors[index]
		if operationErr != nil {
			operationErr = fmt.Errorf("switch image for %s: %w", nodes[index].Name(), operationErr)
		}
		finished, finishErr := n.finishUpgradeNodeSwitch(item.evidence, operationErr)
		evidence[index] = finished
		if finishErr != nil {
			switchErrors = append(switchErrors, finishErr)
		}
	}
	return evidence, errors.Join(switchErrors...)
}

func (n *Network) planUpgradeNodeSwitch(
	ctx context.Context,
	phase string,
	node *cosmos.ChainNode,
	target ImageRef,
) (plannedUpgradeNodeSwitch, error) {
	if n == nil || n.artifacts == nil {
		return plannedUpgradeNodeSwitch{}, errors.New("upgrade network artifact store is unavailable")
	}
	if strings.TrimSpace(phase) == "" {
		return plannedUpgradeNodeSwitch{}, errors.New("upgrade switch phase is required")
	}
	if node == nil {
		return plannedUpgradeNodeSwitch{}, errors.New("upgrade node is required")
	}
	if node.DockerClient == nil || node.ContainerID() == "" {
		return plannedUpgradeNodeSwitch{}, fmt.Errorf("upgrade node %s has no Docker container", node.Name())
	}
	plan, err := newUpgradeSwitchPlan(node.Name(), node.Image, target)
	if err != nil {
		return plannedUpgradeNodeSwitch{}, err
	}
	evidence := UpgradeNodeSwitchEvidence{
		RecordedAt: time.Now().UTC(),
		Phase:      phase,
		Plan:       plan,
	}

	inspect, err := node.DockerClient.ContainerInspect(ctx, node.ContainerID())
	if err != nil {
		return plannedUpgradeNodeSwitch{}, fmt.Errorf("inspect old container for %s: %w", node.Name(), err)
	}
	evidence.WasRunning = inspect.State != nil && inspect.State.Running
	evidence.OldImageID = inspect.Image
	targetInspect, _, err := node.DockerClient.ImageInspectWithRaw(ctx, target.Repository+":"+target.Version)
	if err != nil {
		return plannedUpgradeNodeSwitch{}, fmt.Errorf("inspect target image for %s: %w", node.Name(), err)
	}
	evidence.NewImageID = targetInspect.ID
	if strings.TrimSpace(evidence.OldImageID) == "" || strings.TrimSpace(evidence.NewImageID) == "" {
		return plannedUpgradeNodeSwitch{}, fmt.Errorf("upgrade image IDs for %s must be non-empty", node.Name())
	}
	if evidence.OldImageID == evidence.NewImageID {
		return plannedUpgradeNodeSwitch{}, fmt.Errorf("upgrade images for %s resolve to the same Docker image %s", node.Name(), evidence.OldImageID)
	}
	operations := upgradeSwitchOperations{
		capture: func() error {
			return n.CaptureNodeContainerLog(ctx, node, "upgrade/old-logs/"+node.Name()+".log")
		},
		stop: func() error {
			return node.StopContainer(ctx)
		},
		remove: func() error {
			return node.RemoveContainer(ctx)
		},
		create: func() error {
			node.Image = plan.To
			return node.CreateNodeContainer(ctx)
		},
		start: func() error {
			return node.StartContainer(ctx)
		},
	}
	operations.capture = n.withUpgradeSwitchOperation(evidence.Phase, node.Name(), "capture-old-log", operations.capture)
	operations.stop = n.withUpgradeSwitchOperation(evidence.Phase, node.Name(), "stop-old-container", operations.stop)
	operations.remove = n.withUpgradeSwitchOperation(evidence.Phase, node.Name(), "remove-old-container", operations.remove)
	operations.create = n.withUpgradeSwitchOperation(evidence.Phase, node.Name(), "create-new-container", operations.create)
	operations.start = n.withUpgradeSwitchOperation(evidence.Phase, node.Name(), "start-new-container", operations.start)
	return plannedUpgradeNodeSwitch{evidence: evidence, operations: operations}, nil
}

func (n *Network) withUpgradeSwitchOperation(
	phase string,
	node string,
	operation string,
	run func() error,
) func() error {
	return func() error {
		startedAt := time.Now().UTC()
		startRecordErr := n.artifacts.appendJSONLine("upgrade/switch-order.jsonl", upgradeSwitchOperationEvent{
			RecordedAt: startedAt,
			Phase:      phase,
			Node:       node,
			Operation:  operation,
			State:      "started",
		})
		operationErr := run()
		completed := upgradeSwitchOperationEvent{
			RecordedAt: time.Now().UTC(),
			Phase:      phase,
			Node:       node,
			Operation:  operation,
			State:      "completed",
		}
		if operationErr != nil {
			completed.Error = operationErr.Error()
		}
		completeRecordErr := n.artifacts.appendJSONLine("upgrade/switch-order.jsonl", completed)
		return errors.Join(startRecordErr, operationErr, completeRecordErr)
	}
}

func (n *Network) finishUpgradeNodeSwitch(
	evidence UpgradeNodeSwitchEvidence,
	operationErr error,
) (UpgradeNodeSwitchEvidence, error) {
	evidence.CompletedAt = time.Now().UTC()
	if operationErr != nil {
		evidence.Error = operationErr.Error()
		n.artifacts.recordFailure("upgrade-node-switch-"+evidence.Plan.Node, operationErr)
	}
	recordErr := n.artifacts.appendJSONLine("upgrade/node-switches.jsonl", evidence)
	if recordErr != nil {
		n.artifacts.recordFailure("upgrade-node-switch-evidence", recordErr)
	}
	return evidence, errors.Join(operationErr, recordErr)
}
