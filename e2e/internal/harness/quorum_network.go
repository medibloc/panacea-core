package harness

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

const quorumPollInterval = 500 * time.Millisecond

// QuorumNodeState records the live sync and peer state plus the exact
// commitment used in an agreement proof.
type QuorumNodeState struct {
	QuorumCommitment
	ChainID      string `json:"chain_id"`
	LatestHeight int64  `json:"latest_height"`
	CatchingUp   bool   `json:"catching_up"`
	Peers        int    `json:"peers"`
}

// WaitForQuorumProgress proves that one explicit observation node committed
// at least minimumBlocks after startHeight.
func (n *Network) WaitForQuorumProgress(
	ctx context.Context,
	phase string,
	node *cosmos.ChainNode,
	startHeight int64,
	minimumBlocks int64,
) (QuorumHeightWindow, error) {
	if node == nil {
		return QuorumHeightWindow{}, errors.New("quorum progress node is required")
	}
	observer, _ := NewQuorumObserver(quorumPollInterval)
	window, operationErr := observer.WaitForProgress(ctx, startHeight, minimumBlocks, node.Height)
	recordErr := n.recordQuorumEvidence(phase, "progress", map[string]any{
		"node":   node.Name(),
		"window": window,
	}, operationErr)
	return window, errors.Join(operationErr, recordErr)
}

// ObserveQuorumStall absorbs any block already in flight and then proves that
// the explicit observation node's committed height stays fixed.
func (n *Network) ObserveQuorumStall(
	ctx context.Context,
	phase string,
	node *cosmos.ChainNode,
	quietPeriod time.Duration,
	observationPeriod time.Duration,
) (QuorumHeightWindow, error) {
	if node == nil {
		return QuorumHeightWindow{}, errors.New("quorum stall node is required")
	}
	observer, _ := NewQuorumObserver(quorumPollInterval)
	window, operationErr := observer.ObserveStall(ctx, quietPeriod, observationPeriod, node.Height)
	recordErr := n.recordQuorumEvidence(phase, "stall", map[string]any{
		"node":               node.Name(),
		"quiet_period":       quietPeriod.String(),
		"observation_period": observationPeriod.String(),
		"window":             window,
	}, operationErr)
	return window, errors.Join(operationErr, recordErr)
}

// WaitForQuorumAgreement waits for every selected node to finish catching up
// to targetHeight with at least one connected peer, reads that exact block,
// and verifies a single block and application commitment across all nodes.
func (n *Network) WaitForQuorumAgreement(
	ctx context.Context,
	phase string,
	targetHeight int64,
	nodes ...*cosmos.ChainNode,
) (QuorumAgreement, error) {
	if strings.TrimSpace(phase) == "" {
		return QuorumAgreement{}, errors.New("quorum agreement phase is required")
	}
	if targetHeight <= 0 {
		return QuorumAgreement{}, fmt.Errorf("quorum agreement height must be positive: %d", targetHeight)
	}
	if len(nodes) < 2 {
		return QuorumAgreement{}, errors.New("quorum agreement requires at least two nodes")
	}

	states := make([]QuorumNodeState, 0, len(nodes))
	var operationErr error
	for _, node := range nodes {
		state, err := waitForQuorumNodeState(ctx, targetHeight, node)
		if err != nil {
			operationErr = err
			break
		}
		states = append(states, state)
	}

	commitments := make([]QuorumCommitment, 0, len(states))
	for _, state := range states {
		commitments = append(commitments, state.QuorumCommitment)
	}
	var agreement QuorumAgreement
	if operationErr == nil {
		agreement, operationErr = VerifyCommonCommitment(targetHeight, commitments)
	}
	recordErr := n.recordQuorumEvidence(phase, "agreement", map[string]any{
		"target_height": targetHeight,
		"states":        states,
		"agreement":     agreement,
	}, operationErr)
	return agreement, errors.Join(operationErr, recordErr)
}

func waitForQuorumNodeState(
	ctx context.Context,
	targetHeight int64,
	node *cosmos.ChainNode,
) (QuorumNodeState, error) {
	if node == nil {
		return QuorumNodeState{}, errors.New("quorum node is required")
	}
	if node.Client == nil {
		return QuorumNodeState{}, fmt.Errorf("quorum node %s has no RPC client", node.Name())
	}

	ticker := time.NewTicker(quorumPollInterval)
	defer ticker.Stop()
	var (
		latestHeight int64
		catchingUp   bool
		chainID      string
		peers        int
		lastErr      error
	)
	for {
		status, err := node.Client.Status(ctx)
		if err == nil && status != nil {
			latestHeight = status.SyncInfo.LatestBlockHeight
			catchingUp = status.SyncInfo.CatchingUp
			chainID = status.NodeInfo.Network
			if latestHeight >= targetHeight && !catchingUp {
				netInfo, netInfoErr := node.Client.NetInfo(ctx)
				switch {
				case netInfoErr != nil:
					lastErr = fmt.Errorf("query connected peers: %w", netInfoErr)
				case netInfo == nil:
					lastErr = errors.New("connected peers RPC returned an empty result")
				case netInfo.NPeers < 1:
					peers = netInfo.NPeers
					lastErr = errors.New("connected peers RPC reported no connected peers")
				default:
					peers = netInfo.NPeers
					lastErr = nil
				}
				if lastErr == nil {
					break
				}
			} else {
				lastErr = nil
			}
		} else if err != nil {
			lastErr = err
		} else {
			lastErr = errors.New("status returned an empty result")
		}

		select {
		case <-ctx.Done():
			return QuorumNodeState{}, fmt.Errorf(
				"wait for node %s at height %d: latest=%d catching_up=%t peers=%d last error=%v: %w",
				node.Name(),
				targetHeight,
				latestHeight,
				catchingUp,
				peers,
				lastErr,
				ctx.Err(),
			)
		case <-ticker.C:
		}
	}

	height := targetHeight
	result, err := node.Client.Block(ctx, &height)
	if err != nil {
		return QuorumNodeState{}, fmt.Errorf("query node %s block %d: %w", node.Name(), targetHeight, err)
	}
	if result == nil || result.Block == nil {
		return QuorumNodeState{}, fmt.Errorf("query node %s block %d returned an empty result", node.Name(), targetHeight)
	}
	if result.Block.Header.Height != targetHeight {
		return QuorumNodeState{}, fmt.Errorf(
			"query node %s block %d returned height %d",
			node.Name(),
			targetHeight,
			result.Block.Header.Height,
		)
	}
	blockResults, err := node.Client.BlockResults(ctx, &height)
	if err != nil {
		return QuorumNodeState{}, fmt.Errorf("query node %s block results %d: %w", node.Name(), targetHeight, err)
	}
	if blockResults == nil {
		return QuorumNodeState{}, fmt.Errorf("query node %s block results %d returned an empty result", node.Name(), targetHeight)
	}
	if blockResults.Height != targetHeight {
		return QuorumNodeState{}, fmt.Errorf(
			"query node %s block results %d returned height %d",
			node.Name(),
			targetHeight,
			blockResults.Height,
		)
	}
	return QuorumNodeState{
		QuorumCommitment: QuorumCommitment{
			Node:      node.Name(),
			Height:    targetHeight,
			BlockHash: fmt.Sprintf("%X", []byte(result.BlockID.Hash)),
			// BlockResults.AppHash is the application commitment after
			// FinalizeBlock at targetHeight. Header.AppHash commits the
			// preceding height and would be off by one for state mutations.
			AppHash: fmt.Sprintf("%X", blockResults.AppHash),
		},
		ChainID:      chainID,
		LatestHeight: latestHeight,
		CatchingUp:   catchingUp,
		Peers:        peers,
	}, nil
}

// StopQuorumValidator injects one validator fault without removing or cloning
// its container or validator state.
func (n *Network) StopQuorumValidator(ctx context.Context, phase string, index int) error {
	return n.changeQuorumValidatorState(ctx, phase, index, "stop")
}

// KillQuorumValidator delivers SIGKILL to one exact validator container and
// waits until Docker confirms that process is no longer running. It never
// creates a replacement container or duplicates the validator key.
func (n *Network) KillQuorumValidator(ctx context.Context, phase string, index int) error {
	if n == nil || n.Chain == nil {
		return errors.New("quorum network is required")
	}
	if strings.TrimSpace(phase) == "" {
		return errors.New("quorum fault phase is required")
	}
	if index < 0 || index >= len(n.Chain.Validators) {
		return fmt.Errorf("validator index %d outside [0,%d)", index, len(n.Chain.Validators))
	}
	node := n.Chain.Validators[index]
	if node.DockerClient == nil || node.ContainerID() == "" {
		return fmt.Errorf("validator %d (%s) has no Docker container identity", index, node.Name())
	}
	startedAt := time.Now().UTC()
	operationErr := node.DockerClient.ContainerKill(ctx, node.ContainerID(), "SIGKILL")
	if operationErr == nil {
		waited, waitErrors := node.DockerClient.ContainerWait(ctx, node.ContainerID(), dockercontainer.WaitConditionNotRunning)
		select {
		case response := <-waited:
			if response.Error != nil {
				operationErr = fmt.Errorf("wait for SIGKILL exit: %s", response.Error.Message)
			}
		case waitErr := <-waitErrors:
			operationErr = fmt.Errorf("wait for SIGKILL exit: %w", waitErr)
		case <-ctx.Done():
			operationErr = fmt.Errorf("wait for SIGKILL exit: %w", ctx.Err())
		}
	}
	recordErr := n.recordQuorumEvidence(phase, "validator-lifecycle", map[string]any{
		"action":          "sigkill",
		"validator_index": index,
		"node":            node.Name(),
		"started_at":      startedAt,
		"completed_at":    time.Now().UTC(),
	}, operationErr)
	if operationErr != nil {
		operationErr = fmt.Errorf("SIGKILL quorum validator %d (%s): %w", index, node.Name(), operationErr)
	}
	return errors.Join(operationErr, recordErr)
}

// StartQuorumValidator restarts the exact stopped container and waits for
// Interchaintest's bounded RPC/catch-up readiness check.
func (n *Network) StartQuorumValidator(ctx context.Context, phase string, index int) error {
	return n.changeQuorumValidatorState(ctx, phase, index, "start")
}

func (n *Network) changeQuorumValidatorState(
	ctx context.Context,
	phase string,
	index int,
	action string,
) error {
	if n == nil || n.Chain == nil {
		return errors.New("quorum network is required")
	}
	if strings.TrimSpace(phase) == "" {
		return errors.New("quorum fault phase is required")
	}
	if index < 0 || index >= len(n.Chain.Validators) {
		return fmt.Errorf("validator index %d outside [0,%d)", index, len(n.Chain.Validators))
	}
	node := n.Chain.Validators[index]
	startedAt := time.Now().UTC()
	var operationErr error
	switch action {
	case "stop":
		operationErr = node.StopContainer(ctx)
	case "start":
		operationErr = node.StartContainer(ctx)
	default:
		operationErr = fmt.Errorf("unsupported validator lifecycle action %q", action)
	}
	recordErr := n.recordQuorumEvidence(phase, "validator-lifecycle", map[string]any{
		"action":          action,
		"validator_index": index,
		"node":            node.Name(),
		"started_at":      startedAt,
		"completed_at":    time.Now().UTC(),
	}, operationErr)
	if operationErr != nil {
		operationErr = fmt.Errorf("%s quorum validator %d (%s): %w", action, index, node.Name(), operationErr)
	}
	return errors.Join(operationErr, recordErr)
}

func (n *Network) recordQuorumEvidence(phase string, kind string, evidence any, operationErr error) error {
	if n == nil || n.artifacts == nil {
		return errors.New("quorum artifact store is unavailable")
	}
	record := map[string]any{
		"recorded_at": time.Now().UTC(),
		"phase":       phase,
		"kind":        kind,
		"evidence":    evidence,
		"error":       quorumErrorString(operationErr),
	}
	if err := n.artifacts.appendJSONLine("queries/quorum-observations.jsonl", record); err != nil {
		n.artifacts.recordFailure("quorum-artifact-"+kind, err)
		return fmt.Errorf("record quorum %s evidence: %w", kind, err)
	}
	if operationErr != nil {
		n.artifacts.recordFailure("quorum-"+kind+"-"+phase, operationErr)
	}
	return nil
}
