package harness

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/dockerutil"
)

const (
	// NetworkFaultLoopbackRPCListener leaves CometBFT RPC alive inside the
	// container while making the host-published eth0:26657 boundary unreachable.
	NetworkFaultLoopbackRPCListener = "tcp://127.0.0.1:26657"
	rpcFaultRollbackTimeout         = 90 * time.Second
	rpcFaultFirstRestoreTimeout     = 20 * time.Second
)

type rpcFaultContainerIdentity struct {
	ContainerID       string `json:"container_id"`
	CleanupLabel      string `json:"cleanup_label"`
	VolumeName        string `json:"volume_name"`
	VolumeDestination string `json:"volume_destination"`
	VolumeWritable    bool   `json:"volume_writable"`
}

// FullNodeRPCBoundaryFault is the in-memory rollback token and JSON-safe
// evidence for one host-RPC listener isolation fault.
type FullNodeRPCBoundaryFault struct {
	NodeName         string                    `json:"node_name"`
	Container        rpcFaultContainerIdentity `json:"container"`
	OriginalListener string                    `json:"original_listener"`
	FaultListener    string                    `json:"fault_listener"`
	OriginalSHA256   string                    `json:"original_sha256"`
	ModifiedSHA256   string                    `json:"modified_sha256"`
	AppliedAt        time.Time                 `json:"applied_at"`
	originalConfig   []byte
}

func (e FullNodeRPCBoundaryFault) Validate() error {
	var validationErrors []error
	for name, value := range map[string]string{
		"node name":          e.NodeName,
		"container ID":       e.Container.ContainerID,
		"cleanup label":      e.Container.CleanupLabel,
		"volume name":        e.Container.VolumeName,
		"volume destination": e.Container.VolumeDestination,
		"original listener":  e.OriginalListener,
		"fault listener":     e.FaultListener,
		"original SHA-256":   e.OriginalSHA256,
		"modified SHA-256":   e.ModifiedSHA256,
	} {
		if strings.TrimSpace(value) == "" {
			validationErrors = append(validationErrors, fmt.Errorf("RPC fault %s is required", name))
		}
	}
	if !e.Container.VolumeWritable {
		validationErrors = append(validationErrors, errors.New("RPC fault node volume must be writable"))
	}
	if e.FaultListener != NetworkFaultLoopbackRPCListener {
		validationErrors = append(validationErrors, fmt.Errorf("RPC fault listener %q must equal %q", e.FaultListener, NetworkFaultLoopbackRPCListener))
	}
	if e.OriginalListener == e.FaultListener {
		validationErrors = append(validationErrors, errors.New("RPC fault listener did not change"))
	}
	if e.OriginalSHA256 == e.ModifiedSHA256 {
		validationErrors = append(validationErrors, errors.New("RPC fault config hash did not change"))
	}
	if e.AppliedAt.IsZero() {
		validationErrors = append(validationErrors, errors.New("RPC fault applied_at is required"))
	}
	if len(e.originalConfig) == 0 {
		validationErrors = append(validationErrors, errors.New("RPC fault rollback config is required"))
	} else if networkFaultSHA256(e.originalConfig) != e.OriginalSHA256 {
		validationErrors = append(validationErrors, errors.New("RPC fault rollback config hash does not match evidence"))
	}
	return errors.Join(validationErrors...)
}

// ApplyFullNodeRPCBoundaryFault moves only rpc.laddr to container loopback and
// raw-starts the same labeled container. Interchaintest's normal StartContainer
// cannot be used here because its readiness contract intentionally requires the
// host RPC boundary that this fault disables.
func (n *Network) ApplyFullNodeRPCBoundaryFault(
	ctx context.Context,
	node *cosmos.ChainNode,
) (evidence FullNodeRPCBoundaryFault, retErr error) {
	if err := n.validateFullNodeRPCFaultTarget(node); err != nil {
		return evidence, err
	}
	n.txMu.Lock()
	defer n.txMu.Unlock()

	original, err := node.ReadFile(ctx, "config/config.toml")
	if err != nil {
		return evidence, fmt.Errorf("read full-node config before RPC fault: %w", err)
	}
	modified, originalListener, err := rewriteCometRPCListenerForNetworkFault(original)
	if err != nil {
		return evidence, err
	}
	beforeInspect, err := n.artifacts.client.ContainerInspect(ctx, node.ContainerID())
	if err != nil {
		return evidence, fmt.Errorf("inspect full node before RPC fault: %w", err)
	}
	identity, err := validateRunOwnedRPCFaultContainer(
		beforeInspect,
		node.ContainerID(),
		n.artifacts.runID,
		node.VolumeName,
		node.HomeDir(),
	)
	if err != nil {
		return evidence, err
	}
	if beforeInspect.State == nil || !beforeInspect.State.Running {
		return evidence, errors.New("RPC fault target container must be running before mutation")
	}
	evidence = FullNodeRPCBoundaryFault{
		NodeName:         node.Name(),
		Container:        identity,
		OriginalListener: originalListener,
		FaultListener:    NetworkFaultLoopbackRPCListener,
		OriginalSHA256:   networkFaultSHA256(original),
		ModifiedSHA256:   networkFaultSHA256(modified),
		AppliedAt:        time.Now().UTC(),
		originalConfig:   append([]byte(nil), original...),
	}
	if err := evidence.Validate(); err != nil {
		return evidence, err
	}

	configMayHaveChanged := false
	defer func() {
		if retErr == nil || !configMayHaveChanged {
			return
		}
		rollbackCtx, rollbackCancel := context.WithTimeout(context.Background(), rpcFaultRollbackTimeout)
		defer rollbackCancel()
		rollbackErr := n.restoreFullNodeRPCBoundaryFaultLocked(rollbackCtx, node, evidence, "apply-rollback")
		if rollbackErr != nil {
			retErr = errors.Join(retErr, fmt.Errorf("rollback failed RPC boundary fault: %w", rollbackErr))
		}
	}()

	if err := node.StopContainer(ctx); err != nil {
		return evidence, fmt.Errorf("stop full node before RPC fault: %w", err)
	}
	configMayHaveChanged = true
	if err := node.WriteFile(ctx, modified, "config/config.toml"); err != nil {
		return evidence, fmt.Errorf("write loopback RPC config: %w", err)
	}
	if err := n.artifacts.client.ContainerStart(ctx, node.ContainerID(), dockertypes.ContainerStartOptions{}); err != nil {
		return evidence, fmt.Errorf("raw-start run-owned RPC fault container: %w", err)
	}
	afterInspect, err := n.artifacts.client.ContainerInspect(ctx, node.ContainerID())
	if err != nil {
		return evidence, fmt.Errorf("inspect raw-started RPC fault container: %w", err)
	}
	afterIdentity, err := validateRunOwnedRPCFaultContainer(
		afterInspect,
		node.ContainerID(),
		n.artifacts.runID,
		node.VolumeName,
		node.HomeDir(),
	)
	if err != nil {
		return evidence, err
	}
	if afterIdentity != identity {
		return evidence, errors.New("raw RPC fault start changed container ownership or volume identity")
	}
	if afterInspect.State == nil || !afterInspect.State.Running {
		return evidence, errors.New("raw-started RPC fault container is not running")
	}
	if err := n.artifacts.write("network-faults/config/full-node-rpc-loopback-config.toml", modified); err != nil {
		return evidence, err
	}
	if err := n.artifacts.writeJSON("network-faults/rpc-listener-applied.json", evidence); err != nil {
		return evidence, err
	}
	return evidence, nil
}

// RestoreFullNodeRPCBoundaryFault restores the exact original config, then
// uses normal StartContainer so host RPC readiness and the ChainNode clients
// are re-established before returning.
func (n *Network) RestoreFullNodeRPCBoundaryFault(
	ctx context.Context,
	node *cosmos.ChainNode,
	evidence FullNodeRPCBoundaryFault,
) error {
	if err := n.validateFullNodeRPCFaultTarget(node); err != nil {
		return err
	}
	if err := evidence.Validate(); err != nil {
		return err
	}
	n.txMu.Lock()
	defer n.txMu.Unlock()
	return n.restoreFullNodeRPCBoundaryFaultLocked(ctx, node, evidence, "normal-restore")
}

func (n *Network) restoreFullNodeRPCBoundaryFaultLocked(
	ctx context.Context,
	node *cosmos.ChainNode,
	evidence FullNodeRPCBoundaryFault,
	reason string,
) error {
	inspect, err := n.artifacts.client.ContainerInspect(ctx, node.ContainerID())
	if err != nil {
		return fmt.Errorf("inspect RPC fault container before restore: %w", err)
	}
	identity, err := validateRunOwnedRPCFaultContainer(
		inspect,
		evidence.Container.ContainerID,
		evidence.Container.CleanupLabel,
		evidence.Container.VolumeName,
		evidence.Container.VolumeDestination,
	)
	if err != nil {
		return err
	}
	if identity != evidence.Container {
		return errors.New("RPC fault restore target identity differs from apply evidence")
	}
	if inspect.State == nil {
		return errors.New("RPC fault restore inspect has no container state")
	}
	if inspect.State.Running {
		if err := node.StopContainer(ctx); err != nil {
			return fmt.Errorf("stop RPC fault container for restore: %w", err)
		}
	}
	if err := node.WriteFile(ctx, evidence.originalConfig, "config/config.toml"); err != nil {
		return errors.Join(
			fmt.Errorf("restore original RPC config: %w", err),
			node.StartContainer(ctx),
		)
	}
	firstStartCtx, firstStartCancel := context.WithTimeout(ctx, rpcFaultFirstRestoreTimeout)
	firstStartErr := node.StartContainer(firstStartCtx)
	firstStartCancel()
	normalStartAttempts := 1
	readinessRecoveryRestart := false
	firstStartHeight := int64(0)
	if firstStartErr != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("normal-start restored RPC container: %w", errors.Join(firstStartErr, ctx.Err()))
		}
		startInspect, inspectErr := n.artifacts.client.ContainerInspect(ctx, node.ContainerID())
		statusCtx, statusCancel := context.WithTimeout(ctx, 3*time.Second)
		var catchingUp bool
		var statusErr error
		if node.Client == nil {
			statusErr = errors.New("restored RPC client is nil")
		} else {
			status, queryErr := node.Client.Status(statusCtx)
			statusErr = queryErr
			if status != nil {
				catchingUp = status.SyncInfo.CatchingUp
				firstStartHeight = status.SyncInfo.LatestBlockHeight
			}
		}
		statusCancel()
		containerRunning := inspectErr == nil && startInspect.State != nil && startInspect.State.Running
		if !shouldRetryRPCFaultNormalStart(firstStartErr, ctx.Err(), containerRunning, catchingUp) {
			return fmt.Errorf(
				"normal-start restored RPC container without a retryable catching-up state: %w",
				errors.Join(firstStartErr, inspectErr, statusErr),
			)
		}
		if err := n.artifacts.writeJSON("network-faults/rpc-listener-readiness-recovery.json", map[string]any{
			"node_name":                evidence.NodeName,
			"container_id":             evidence.Container.ContainerID,
			"first_start_error":        firstStartErr.Error(),
			"first_start_height":       firstStartHeight,
			"container_running":        true,
			"catching_up":              true,
			"recovery_restart_planned": true,
			"recorded_at":              time.Now().UTC(),
		}); err != nil {
			return err
		}
		if err := node.StopContainer(ctx); err != nil {
			return fmt.Errorf("stop catching-up RPC restore before bounded readiness restart: %w", err)
		}
		normalStartAttempts++
		readinessRecoveryRestart = true
		if err := node.StartContainer(ctx); err != nil {
			return fmt.Errorf(
				"normal-start restored RPC container after one bounded readiness restart: %w",
				errors.Join(firstStartErr, err),
			)
		}
	}
	restored, err := node.ReadFile(ctx, "config/config.toml")
	if err != nil {
		return fmt.Errorf("read restored RPC config: %w", err)
	}
	restoredHash := networkFaultSHA256(restored)
	if restoredHash != evidence.OriginalSHA256 {
		return fmt.Errorf("restored RPC config hash %s does not match original %s", restoredHash, evidence.OriginalSHA256)
	}
	restoredInspect, err := n.artifacts.client.ContainerInspect(ctx, node.ContainerID())
	if err != nil {
		return fmt.Errorf("inspect restored RPC container: %w", err)
	}
	restoredIdentity, err := validateRunOwnedRPCFaultContainer(
		restoredInspect,
		evidence.Container.ContainerID,
		evidence.Container.CleanupLabel,
		evidence.Container.VolumeName,
		evidence.Container.VolumeDestination,
	)
	if err != nil {
		return err
	}
	if restoredIdentity != evidence.Container {
		return errors.New("normal RPC restore changed container ownership or volume identity")
	}
	if restoredInspect.State == nil || !restoredInspect.State.Running {
		return errors.New("restored RPC container is not running")
	}
	return n.artifacts.writeJSON("network-faults/rpc-listener-restored.json", map[string]any{
		"node_name":                  evidence.NodeName,
		"container":                  restoredIdentity,
		"restored_listener":          evidence.OriginalListener,
		"restored_sha256":            restoredHash,
		"hash_matches":               true,
		"normal_start":               true,
		"normal_start_attempts":      normalStartAttempts,
		"readiness_recovery_restart": readinessRecoveryRestart,
		"first_start_error":          errorString(firstStartErr),
		"first_start_height":         firstStartHeight,
		"reason":                     reason,
		"restored_at":                time.Now().UTC(),
	})
}

func shouldRetryRPCFaultNormalStart(
	startErr error,
	parentErr error,
	containerRunning bool,
	catchingUp bool,
) bool {
	return startErr != nil && parentErr == nil && containerRunning && catchingUp
}

func (n *Network) validateFullNodeRPCFaultTarget(node *cosmos.ChainNode) error {
	if n == nil || n.Chain == nil || n.artifacts == nil || n.artifacts.client == nil {
		return errors.New("RPC boundary fault network is unavailable")
	}
	if node == nil || strings.TrimSpace(node.ContainerID()) == "" {
		return errors.New("RPC boundary fault requires a concrete full node")
	}
	for _, fullNode := range n.Chain.FullNodes {
		if fullNode == node {
			return nil
		}
	}
	return errors.New("RPC boundary fault target is not owned by this network's full-node set")
}

func validateRunOwnedRPCFaultContainer(
	inspect dockertypes.ContainerJSON,
	expectedContainerID string,
	expectedCleanupLabel string,
	expectedVolumeName string,
	expectedVolumeDestination string,
) (rpcFaultContainerIdentity, error) {
	identity := rpcFaultContainerIdentity{}
	if inspect.ContainerJSONBase == nil || strings.TrimSpace(inspect.ID) == "" {
		return identity, errors.New("RPC fault Docker inspect has no container ID")
	}
	if inspect.ID != expectedContainerID {
		return identity, fmt.Errorf("RPC fault container ID %s does not match expected %s", inspect.ID, expectedContainerID)
	}
	if inspect.Config == nil {
		return identity, errors.New("RPC fault Docker inspect has no container config")
	}
	cleanupLabel := strings.TrimSpace(inspect.Config.Labels[dockerutil.CleanupLabel])
	if cleanupLabel == "" || cleanupLabel != expectedCleanupLabel {
		return identity, fmt.Errorf("RPC fault cleanup label %q does not match run %q", cleanupLabel, expectedCleanupLabel)
	}
	var matched *dockertypes.MountPoint
	for index := range inspect.Mounts {
		mount := &inspect.Mounts[index]
		if mount.Name == expectedVolumeName && mount.Destination == expectedVolumeDestination {
			matched = mount
			break
		}
	}
	if matched == nil {
		return identity, fmt.Errorf("RPC fault container has no expected volume %s at %s", expectedVolumeName, expectedVolumeDestination)
	}
	if !matched.RW {
		return identity, errors.New("RPC fault node volume is not writable")
	}
	identity = rpcFaultContainerIdentity{
		ContainerID:       inspect.ID,
		CleanupLabel:      cleanupLabel,
		VolumeName:        matched.Name,
		VolumeDestination: matched.Destination,
		VolumeWritable:    matched.RW,
	}
	return identity, nil
}

func rewriteCometRPCListenerForNetworkFault(contents []byte) ([]byte, string, error) {
	var document map[string]any
	if _, err := toml.Decode(string(contents), &document); err != nil {
		return nil, "", fmt.Errorf("decode config.toml for RPC fault: %w", err)
	}
	rpcValue, ok := document["rpc"]
	if !ok {
		return nil, "", errors.New("config.toml has no rpc section")
	}
	rpc, ok := rpcValue.(map[string]any)
	if !ok {
		return nil, "", fmt.Errorf("config.toml rpc section has type %T", rpcValue)
	}
	originalListener, ok := rpc["laddr"].(string)
	if !ok || strings.TrimSpace(originalListener) == "" {
		return nil, "", fmt.Errorf("config.toml rpc.laddr has invalid value %T", rpc["laddr"])
	}
	if originalListener == NetworkFaultLoopbackRPCListener {
		return nil, "", errors.New("config.toml RPC listener is already loopback-only")
	}
	rpc["laddr"] = NetworkFaultLoopbackRPCListener
	document["rpc"] = rpc
	var output bytes.Buffer
	if err := toml.NewEncoder(&output).Encode(document); err != nil {
		return nil, "", fmt.Errorf("encode RPC fault config.toml: %w", err)
	}
	return output.Bytes(), originalListener, nil
}
