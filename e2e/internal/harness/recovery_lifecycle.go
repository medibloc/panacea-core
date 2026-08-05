package harness

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

type recoveryAction struct {
	RecordedAt time.Time `json:"recorded_at"`
	Step       string    `json:"step"`
	Action     string    `json:"action"`
	Node       string    `json:"node"`
	Stdout     string    `json:"stdout,omitempty"`
	Stderr     string    `json:"stderr,omitempty"`
	Error      string    `json:"error,omitempty"`
}

func runRecoveryStoppedOperation(stop, operation, start func() error) error {
	stopErr := stop()
	if stopErr != nil {
		// A stop request can fail after changing container state. Always make a
		// bounded best effort to return the disposable test node to running.
		return errors.Join(stopErr, start())
	}
	var operationErr error
	if operation != nil {
		operationErr = operation()
	}
	return errors.Join(operationErr, start())
}

func (n *Network) runRecoveryAction(
	step string,
	action string,
	node *cosmos.ChainNode,
	operation func() ([]byte, []byte, error),
) ([]byte, error) {
	started := time.Now().UTC()
	stdout, stderr, operationErr := operation()
	record := recoveryAction{
		RecordedAt: started,
		Step:       step,
		Action:     action,
		Node:       node.Name(),
		Stdout:     boundedArtifactText(stdout),
		Stderr:     boundedArtifactText(stderr),
		Error:      errorString(operationErr),
	}
	artifactErr := n.artifacts.appendJSONLine("recovery/actions.jsonl", record)
	if operationErr != nil {
		n.artifacts.recordFailure("recovery-"+action, operationErr)
	}
	if artifactErr != nil {
		n.artifacts.recordFailure("recovery-action-artifact", artifactErr)
	}
	return stdout, errors.Join(operationErr, artifactErr)
}

func (n *Network) stopNodeForRecovery(ctx context.Context, step string, node *cosmos.ChainNode) error {
	_, err := n.runRecoveryAction(step, "graceful-stop", node, func() ([]byte, []byte, error) {
		return nil, nil, node.StopContainer(ctx)
	})
	return err
}

func (n *Network) startNodeAfterRecovery(ctx context.Context, step string, node *cosmos.ChainNode) error {
	_, err := n.runRecoveryAction(step, "start", node, func() ([]byte, []byte, error) {
		return nil, nil, node.StartContainer(ctx)
	})
	return err
}

func (n *Network) killNodeForRecovery(ctx context.Context, step string, node *cosmos.ChainNode) error {
	containerID := node.ContainerID()
	_, err := n.runRecoveryAction(step, "sigkill", node, func() ([]byte, []byte, error) {
		if err := node.DockerClient.ContainerKill(ctx, containerID, "SIGKILL"); err != nil {
			return nil, nil, err
		}
		waited, waitErrors := node.DockerClient.ContainerWait(
			ctx,
			containerID,
			dockercontainer.WaitConditionNotRunning,
		)
		select {
		case response := <-waited:
			if response.Error != nil {
				return nil, nil, fmt.Errorf("wait for SIGKILL exit: %s", response.Error.Message)
			}
			return []byte(fmt.Sprintf("exit status %d", response.StatusCode)), nil, nil
		case waitErr := <-waitErrors:
			return nil, nil, fmt.Errorf("wait for SIGKILL exit: %w", waitErr)
		case <-ctx.Done():
			return nil, nil, fmt.Errorf("wait for SIGKILL exit: %w", ctx.Err())
		}
	})
	return err
}

func (n *Network) restartNodeAbruptly(
	ctx context.Context,
	step string,
	role string,
	node *cosmos.ChainNode,
	afterRestart func() error,
) error {
	if node.DockerClient == nil || node.ContainerID() == "" {
		return fmt.Errorf("%s %s has no running container identity", role, node.Name())
	}

	n.txMu.Lock()
	defer n.txMu.Unlock()
	if err := runRecoveryStoppedOperation(
		func() error { return n.killNodeForRecovery(ctx, step, node) },
		nil,
		func() error { return n.startNodeAfterRecovery(ctx, step, node) },
	); err != nil {
		return fmt.Errorf("abrupt %s restart on %s: %w", role, node.Name(), err)
	}
	if afterRestart != nil {
		if err := afterRestart(); err != nil {
			return fmt.Errorf("abrupt %s restart on %s: %w", role, node.Name(), err)
		}
	}
	return nil
}

// RestartFullNodeGracefully closes and reopens the selected full-node
// container while serializing against transaction broadcasts. The start phase
// is attempted even if stop-side artifact recording reports an error.
func (n *Network) RestartFullNodeGracefully(
	ctx context.Context,
	step string,
	fullNodeIndex int,
) error {
	if fullNodeIndex < 0 || fullNodeIndex >= len(n.Chain.FullNodes) {
		return fmt.Errorf("full-node index %d is out of range", fullNodeIndex)
	}
	node := n.Chain.FullNodes[fullNodeIndex]
	n.txMu.Lock()
	defer n.txMu.Unlock()
	if err := runRecoveryStoppedOperation(
		func() error { return n.stopNodeForRecovery(ctx, step, node) },
		nil,
		func() error { return n.startNodeAfterRecovery(ctx, step, node) },
	); err != nil {
		return fmt.Errorf("graceful full-node restart on %s: %w", node.Name(), err)
	}
	return nil
}

// RestartFullNodeAbruptly delivers SIGKILL to the selected full node, observes
// the container's non-running state, and starts the same container and mounted
// database again. The caller must subsequently prove history, state, and
// forward-progress continuity through public query boundaries.
func (n *Network) RestartFullNodeAbruptly(
	ctx context.Context,
	step string,
	fullNodeIndex int,
) error {
	if fullNodeIndex < 0 || fullNodeIndex >= len(n.Chain.FullNodes) {
		return fmt.Errorf("full-node index %d is out of range", fullNodeIndex)
	}
	return n.restartNodeAbruptly(ctx, step, "full-node", n.Chain.FullNodes[fullNodeIndex], nil)
}

// RestartValidatorAbruptly delivers SIGKILL to the selected validator,
// observes the container's non-running state, starts the same container and
// mounted database again, and requires CometBFT's post-crash consensus-WAL
// replay start and completion markers.
func (n *Network) RestartValidatorAbruptly(
	ctx context.Context,
	step string,
	validatorIndex int,
) error {
	if validatorIndex < 0 || validatorIndex >= len(n.Chain.Validators) {
		return fmt.Errorf("validator index %d is out of range", validatorIndex)
	}
	node := n.Chain.Validators[validatorIndex]
	replaySince := time.Now().UTC()

	return n.restartNodeAbruptly(ctx, step, "validator", node, func() error {
		replayCtx, replayCancel := context.WithTimeout(ctx, 30*time.Second)
		defer replayCancel()
		if err := n.waitForValidatorWALReplay(replayCtx, step, node, replaySince); err != nil {
			n.artifacts.recordFailure("recovery-wal-replay-"+step, err)
			return err
		}
		return nil
	})
}

func classifyWALReplayEvidence(logs []byte) error {
	text := string(logs)
	if !strings.Contains(text, "Catchup by replaying consensus messages") {
		return errors.New("post-crash logs have no consensus replay start marker")
	}
	if !strings.Contains(text, "Replay: Done") {
		return errors.New("post-crash logs have no consensus replay completion marker")
	}
	return nil
}

func (n *Network) waitForValidatorWALReplay(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	since time.Time,
) error {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	var (
		lastLogs []byte
		lastErr  error
	)
	for {
		logs, err := recoveryContainerLogs(ctx, node, since)
		if err == nil {
			lastLogs = logs
			err = classifyWALReplayEvidence(logs)
		}
		lastErr = err
		if err == nil {
			return n.artifacts.appendJSONLine("recovery/wal-replay.jsonl", map[string]any{
				"recorded_at": time.Now().UTC(),
				"step":        step,
				"node":        node.Name(),
				"since":       since,
				"markers": []string{
					"Catchup by replaying consensus messages",
					"Replay: Done",
				},
				"logs": boundedArtifactText(logs),
			})
		}

		select {
		case <-ctx.Done():
			artifactErr := n.artifacts.appendJSONLine("recovery/wal-replay.jsonl", map[string]any{
				"recorded_at": time.Now().UTC(),
				"step":        step,
				"node":        node.Name(),
				"since":       since,
				"error":       errorString(lastErr),
				"logs":        boundedArtifactText(lastLogs),
			})
			return errors.Join(
				fmt.Errorf("wait for post-crash consensus WAL replay: last error=%v: %w", lastErr, ctx.Err()),
				artifactErr,
			)
		case <-ticker.C:
		}
	}
}

func recoveryContainerLogs(ctx context.Context, node *cosmos.ChainNode, since time.Time) ([]byte, error) {
	reader, err := node.DockerClient.ContainerLogs(ctx, node.ContainerID(), dockertypes.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Since:      strconv.FormatInt(since.Add(-time.Second).Unix(), 10),
	})
	if err != nil {
		return nil, fmt.Errorf("read post-crash container logs: %w", err)
	}
	defer reader.Close()

	budget := &artifactLogBudget{remaining: 1 << 20}
	stdout := artifactLogBuffer{budget: budget}
	stderr := artifactLogBuffer{budget: budget}
	if _, err := stdcopy.StdCopy(&stdout, &stderr, reader); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("demultiplex post-crash container logs: %w", err)
	}
	logs := make([]byte, 0, stdout.Len()+stderr.Len())
	logs = append(logs, stdout.Bytes()...)
	logs = append(logs, stderr.Bytes()...)
	return logs, nil
}

// RestoreFullNodeFromLocalSnapshot stops the selected full node, creates a
// local Cosmos SDK snapshot at an exact committed height, verifies that the
// snapshot store lists it, restores it into the closed application DB, and
// starts the same node again. The caller must subsequently prove catch-up and
// app-hash/state equality through the public query boundaries.
func (n *Network) RestoreFullNodeFromLocalSnapshot(
	ctx context.Context,
	step string,
	fullNodeIndex int,
	height int64,
) (LocalSnapshot, error) {
	if fullNodeIndex < 0 || fullNodeIndex >= len(n.Chain.FullNodes) {
		return LocalSnapshot{}, fmt.Errorf("full-node index %d is out of range", fullNodeIndex)
	}
	if height <= 0 {
		return LocalSnapshot{}, fmt.Errorf("snapshot height must be positive, got %d", height)
	}
	node := n.Chain.FullNodes[fullNodeIndex]
	n.txMu.Lock()
	defer n.txMu.Unlock()

	var snapshot LocalSnapshot
	if err := n.stopNodeForRecovery(ctx, step, node); err != nil {
		return snapshot, errors.Join(
			fmt.Errorf("stop %s for local snapshot recovery: %w", node.Name(), err),
			n.startNodeAfterRecovery(ctx, step+"-stop-failure", node),
		)
	}
	restartOriginal := func(cause error) (LocalSnapshot, error) {
		startErr := n.startNodeAfterRecovery(ctx, step+"-original", node)
		return snapshot, fmt.Errorf(
			"local snapshot recovery on %s: %w",
			node.Name(),
			errors.Join(cause, startErr),
		)
	}

	heightText := strconv.FormatInt(height, 10)
	if _, err := n.runRecoveryAction(step, "snapshot-export", node, func() ([]byte, []byte, error) {
		return node.ExecBin(ctx, "snapshots", "export", "--height", heightText)
	}); err != nil {
		return restartOriginal(fmt.Errorf("export local snapshot at height %d: %w", height, err))
	}
	listOutput, err := n.runRecoveryAction(step, "snapshot-list", node, func() ([]byte, []byte, error) {
		return node.ExecBin(ctx, "snapshots", "list")
	})
	if err != nil {
		return restartOriginal(fmt.Errorf("list local snapshots: %w", err))
	}
	snapshots, err := ParseLocalSnapshots(listOutput)
	if err != nil {
		return restartOriginal(err)
	}
	snapshot, err = FindLocalSnapshot(snapshots, uint64(height))
	if err != nil {
		return restartOriginal(err)
	}
	plan, err := newApplicationSnapshotRestorePlan(node.HomeDir(), snapshot.Height, snapshot.Format)
	if err != nil {
		return restartOriginal(err)
	}
	if _, err := n.runRecoveryAction(step, "snapshot-dump", node, func() ([]byte, []byte, error) {
		return node.ExecBin(
			ctx,
			"snapshots", "dump",
			heightText,
			strconv.FormatUint(uint64(snapshot.Format), 10),
			"--output", plan.Archive,
		)
	}); err != nil {
		return restartOriginal(fmt.Errorf("dump local snapshot archive: %w", err))
	}
	if err := n.artifacts.appendJSONLine("recovery/snapshot-plans.jsonl", map[string]any{
		"recorded_at": time.Now().UTC(),
		"step":        step,
		"node":        node.Name(),
		"snapshot":    snapshot,
		"plan":        plan,
	}); err != nil {
		return restartOriginal(fmt.Errorf("record local snapshot restore plan: %w", err))
	}

	var mutationArtifactErr error
	mutate := func(action string, operation func() ([]byte, []byte, error)) error {
		started := time.Now().UTC()
		stdout, stderr, operationErr := operation()
		artifactErr := n.artifacts.appendJSONLine("recovery/actions.jsonl", recoveryAction{
			RecordedAt: started,
			Step:       step,
			Action:     action,
			Node:       node.Name(),
			Stdout:     boundedArtifactText(stdout),
			Stderr:     boundedArtifactText(stderr),
			Error:      errorString(operationErr),
		})
		if operationErr != nil {
			n.artifacts.recordFailure("recovery-"+action, operationErr)
		}
		if artifactErr != nil {
			n.artifacts.recordFailure("recovery-action-artifact", artifactErr)
			mutationArtifactErr = errors.Join(mutationArtifactErr, artifactErr)
		}
		return operationErr
	}
	execUtility := func(command ...string) ([]byte, []byte, error) {
		return node.Exec(ctx, command, node.Chain.Config().Env)
	}

	rollback, swapErr := executeApplicationDBSwap(applicationDBSwapOperations{
		MoveOriginalAside: func() error {
			return mutate("snapshot-move-original", func() ([]byte, []byte, error) {
				return execUtility("mv", plan.ApplicationDB, plan.BackupDB)
			})
		},
		RestoreSnapshot: func() error {
			return mutate("snapshot-restore", func() ([]byte, []byte, error) {
				return node.ExecBin(
					ctx,
					"snapshots", "restore",
					heightText,
					strconv.FormatUint(uint64(snapshot.Format), 10),
				)
			})
		},
		ValidateRestored: func() error {
			return mutate("snapshot-validate-db", func() ([]byte, []byte, error) {
				return execUtility("test", "-e", plan.ApplicationDB)
			})
		},
		RemoveRestored: func() error {
			return mutate("snapshot-remove-failed-db", func() ([]byte, []byte, error) {
				return execUtility("rm", "-rf", plan.ApplicationDB)
			})
		},
		MoveOriginalBack: func() error {
			return mutate("snapshot-rollback-original", func() ([]byte, []byte, error) {
				return execUtility("mv", plan.BackupDB, plan.ApplicationDB)
			})
		},
	})
	if swapErr != nil {
		return restartOriginal(errors.Join(swapErr, mutationArtifactErr))
	}
	if mutationArtifactErr != nil {
		return restartOriginal(errors.Join(mutationArtifactErr, rollback()))
	}

	if startErr := n.startNodeAfterRecovery(ctx, step+"-restored", node); startErr != nil {
		stopErr := n.stopNodeForRecovery(ctx, step+"-restored-start-failure", node)
		if stopErr != nil {
			return snapshot, fmt.Errorf(
				"start restored full node %s and stop partial start: %w",
				node.Name(),
				errors.Join(startErr, stopErr),
			)
		}
		rollbackErr := rollback()
		_, recoveryErr := restartOriginal(errors.Join(startErr, rollbackErr, mutationArtifactErr))
		return snapshot, recoveryErr
	}
	return snapshot, nil
}
