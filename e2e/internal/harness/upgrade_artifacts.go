package harness

import (
	"context"
	"errors"
	"fmt"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

// CaptureNodeContainerLog snapshots a soon-to-be-replaced container before
// Interchaintest removes it during a binary image switch.
func (n *Network) CaptureNodeContainerLog(
	ctx context.Context,
	node *cosmos.ChainNode,
	relativePath string,
) error {
	if node == nil {
		return errors.New("node is required")
	}
	if node.DockerClient == nil || node.ContainerID() == "" {
		return fmt.Errorf("node %s has no Docker container", node.Name())
	}
	return n.artifacts.collectLogs(ctx, node.DockerClient, node.ContainerID(), relativePath)
}
