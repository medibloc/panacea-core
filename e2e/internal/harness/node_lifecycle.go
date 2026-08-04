package harness

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

func runStoppedNodeOperation(stop, operation, start func() error) error {
	if stop == nil || operation == nil || start == nil {
		return errors.New("stop, operation, and start callbacks are required")
	}
	if err := stop(); err != nil {
		return err
	}
	operationErr := operation()
	startErr := start()
	return errors.Join(operationErr, startErr)
}

// NodeRestartEvidence proves that one concrete node reopened its persisted
// database, retained the pre-restart history point, and advanced afterwards.
type NodeRestartEvidence struct {
	RecordedAt time.Time     `json:"recorded_at"`
	Mode       string        `json:"mode"`
	Node       string        `json:"node"`
	Before     BlockEvidence `json:"before"`
	After      BlockEvidence `json:"after"`
}

func (n *Network) GracefulRestartNode(ctx context.Context, node *cosmos.ChainNode) (NodeRestartEvidence, error) {
	return n.restartNode(ctx, node, false)
}

func (n *Network) ForceKillAndRestartNode(ctx context.Context, node *cosmos.ChainNode) (NodeRestartEvidence, error) {
	return n.restartNode(ctx, node, true)
}

func (n *Network) restartNode(
	ctx context.Context,
	node *cosmos.ChainNode,
	force bool,
) (NodeRestartEvidence, error) {
	if node == nil {
		return NodeRestartEvidence{}, errors.New("node is required")
	}
	before, err := n.NodeLatestBlock(ctx, node)
	if err != nil {
		return NodeRestartEvidence{}, err
	}
	mode := "graceful"
	if force {
		mode = "sigkill"
		err = node.DockerClient.ContainerKill(ctx, node.ContainerID(), "SIGKILL")
	} else {
		err = node.StopContainer(ctx)
	}
	if err != nil {
		return NodeRestartEvidence{}, fmt.Errorf("%s stop %s: %w", mode, node.Name(), err)
	}
	if err := n.artifacts.appendJSONLine("nodes/restarts.jsonl", map[string]any{
		"recorded_at": time.Now().UTC(),
		"mode":        mode,
		"node":        node.Name(),
		"phase":       "stopped",
		"before":      before,
	}); err != nil {
		return NodeRestartEvidence{}, err
	}
	if err := node.StartContainer(ctx); err != nil {
		return NodeRestartEvidence{}, fmt.Errorf("restart %s after %s stop: %w", node.Name(), mode, err)
	}
	if err := n.WaitForNodeHeight(ctx, node, before.Height+1); err != nil {
		return NodeRestartEvidence{}, fmt.Errorf("wait for %s after %s restart: %w", node.Name(), mode, err)
	}
	retained, err := n.NodeBlock(ctx, node, before.Height)
	if err != nil {
		return NodeRestartEvidence{}, fmt.Errorf("verify retained history on %s: %w", node.Name(), err)
	}
	if err := validateSameHistory([]BlockEvidence{before, retained}); err != nil {
		return NodeRestartEvidence{}, fmt.Errorf("persisted history changed across %s restart: %w", mode, err)
	}
	after, err := n.NodeLatestBlock(ctx, node)
	if err != nil {
		return NodeRestartEvidence{}, err
	}
	evidence := NodeRestartEvidence{
		RecordedAt: time.Now().UTC(),
		Mode:       mode,
		Node:       node.Name(),
		Before:     before,
		After:      after,
	}
	if err := n.artifacts.appendJSONLine("nodes/restarts.jsonl", evidence); err != nil {
		return NodeRestartEvidence{}, err
	}
	return evidence, nil
}

// ExportStateTwiceAtHeight must be called while all long-running node
// containers are stopped. It executes two independent exports against the
// preserved full-node volume and requires byte-for-byte determinism.
func (n *Network) ExportStateTwiceAtHeight(ctx context.Context, height int64) ([]byte, error) {
	if height <= 0 {
		return nil, fmt.Errorf("export height must be positive, got %d", height)
	}
	first, err := n.Chain.ExportState(ctx, height)
	if err != nil {
		return nil, fmt.Errorf("first state export at height %d: %w", height, err)
	}
	second, err := n.Chain.ExportState(ctx, height)
	if err != nil {
		return nil, fmt.Errorf("second state export at height %d: %w", height, err)
	}
	if !bytes.Equal([]byte(first), []byte(second)) {
		_ = n.artifacts.write("recovery/export-first.json", []byte(first))
		_ = n.artifacts.write("recovery/export-second.json", []byte(second))
		return nil, fmt.Errorf("state exports at height %d are not byte-identical", height)
	}
	if err := n.artifacts.write("recovery/export.json", []byte(first)); err != nil {
		return nil, fmt.Errorf("record state export: %w", err)
	}
	if err := n.artifacts.writeJSON("recovery/export-evidence.json", map[string]any{
		"recorded_at": time.Now().UTC(),
		"height":      height,
		"bytes":       len(first),
		"equal":       true,
	}); err != nil {
		return nil, err
	}
	return []byte(first), nil
}
