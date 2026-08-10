package harness

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

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
	return n.restartNode(ctx, node)
}

func (n *Network) restartNode(
	ctx context.Context,
	node *cosmos.ChainNode,
) (NodeRestartEvidence, error) {
	if node == nil {
		return NodeRestartEvidence{}, errors.New("node is required")
	}
	before, err := n.NodeLatestBlock(ctx, node)
	if err != nil {
		return NodeRestartEvidence{}, err
	}
	const mode = "graceful"
	err = node.StopContainer(ctx)
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
