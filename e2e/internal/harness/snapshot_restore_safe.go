package harness

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
)

const interchaintestNodeHomeRoot = "/var/cosmos-chain"

type applicationSnapshotRestorePlan struct {
	ApplicationDB string
	BackupDB      string
	Archive       string
}

func newApplicationSnapshotRestorePlan(
	nodeHome string,
	height uint64,
	format uint32,
) (applicationSnapshotRestorePlan, error) {
	if height == 0 {
		return applicationSnapshotRestorePlan{}, errors.New("snapshot height must be positive")
	}
	if format == 0 {
		return applicationSnapshotRestorePlan{}, errors.New("snapshot format must be positive")
	}
	cleanHome := path.Clean(nodeHome)
	if nodeHome != cleanHome || !path.IsAbs(cleanHome) {
		return applicationSnapshotRestorePlan{}, fmt.Errorf("node home must be a clean absolute path: %q", nodeHome)
	}
	rootPrefix := interchaintestNodeHomeRoot + "/"
	if !strings.HasPrefix(cleanHome, rootPrefix) || cleanHome == interchaintestNodeHomeRoot {
		return applicationSnapshotRestorePlan{}, fmt.Errorf(
			"node home %q must be a child of %s",
			nodeHome,
			interchaintestNodeHomeRoot,
		)
	}
	suffix := fmt.Sprintf("%d-%d", height, format)
	return applicationSnapshotRestorePlan{
		ApplicationDB: path.Join(cleanHome, "data", "application.db"),
		BackupDB:      path.Join(cleanHome, "data", "application.db.snapshot-backup-"+suffix),
		Archive:       path.Join(cleanHome, "snapshot-"+suffix+".tar.gz"),
	}, nil
}

type applicationDBSwapOperations struct {
	MoveOriginalAside func() error
	RestoreSnapshot   func() error
	ValidateRestored  func() error
	RemoveRestored    func() error
	MoveOriginalBack  func() error
}

// executeApplicationDBSwap moves the closed application DB aside before any
// restore is attempted. A failed restore is recoverable: the partial target is
// removed and the original DB is moved back. On success it returns that same
// rollback operation for callers to use if node start or verification fails.
func executeApplicationDBSwap(operations applicationDBSwapOperations) (func() error, error) {
	if operations.MoveOriginalAside == nil ||
		operations.RestoreSnapshot == nil ||
		operations.ValidateRestored == nil ||
		operations.RemoveRestored == nil ||
		operations.MoveOriginalBack == nil {
		return nil, errors.New("application DB swap operations are incomplete")
	}
	if err := operations.MoveOriginalAside(); err != nil {
		return nil, fmt.Errorf("move original application DB aside: %w", err)
	}
	rollback := func() error {
		return errors.Join(operations.RemoveRestored(), operations.MoveOriginalBack())
	}
	if err := operations.RestoreSnapshot(); err != nil {
		return nil, errors.Join(fmt.Errorf("restore application snapshot: %w", err), rollback())
	}
	if err := operations.ValidateRestored(); err != nil {
		return nil, errors.Join(fmt.Errorf("validate restored application DB: %w", err), rollback())
	}
	return rollback, nil
}

func selectLocalSnapshotAtHeight(
	snapshots []LocalSnapshot,
	height uint64,
) (LocalSnapshot, bool, error) {
	var selected LocalSnapshot
	found := false
	for _, snapshot := range snapshots {
		if snapshot.Height != height {
			continue
		}
		if found {
			return LocalSnapshot{}, false, fmt.Errorf("multiple local snapshots exist at height %d", height)
		}
		selected = snapshot
		found = true
	}
	return selected, found, nil
}

// ApplicationSnapshotRestoreEvidence records a portable dump/load round trip,
// the application-only DB swap, and the restored full node's return to the
// validator's canonical history. BackupDB intentionally remains in the
// disposable node volume, making the successful operation recoverable until
// the E2E run's labelled-volume cleanup.
type ApplicationSnapshotRestoreEvidence struct {
	RecordedAt      time.Time       `json:"recorded_at"`
	Step            string          `json:"step"`
	Node            string          `json:"node"`
	Snapshot        LocalSnapshot   `json:"snapshot"`
	Archive         string          `json:"archive"`
	BackupDB        string          `json:"backup_db"`
	Before          []BlockEvidence `json:"before"`
	RestoredHistory []BlockEvidence `json:"restored_history"`
	CaughtUpHistory []BlockEvidence `json:"caught_up_history"`
	Application     struct {
		Height  int64  `json:"height"`
		AppHash string `json:"app_hash"`
	} `json:"application"`
}

// FreshFullNodeSyncEvidence proves that Interchaintest v8.8.1 dynamically
// created a new node volume/container, block-synced it from the existing
// peers, and observed the same block ID and app hash at an exact height.
type FreshFullNodeSyncEvidence struct {
	RecordedAt        time.Time       `json:"recorded_at"`
	Step              string          `json:"step"`
	Node              string          `json:"node"`
	PreviousNodeCount int             `json:"previous_full_node_count"`
	TargetHeight      int64           `json:"target_height"`
	History           []BlockEvidence `json:"history"`
}

type snapshotRestoreAction struct {
	RecordedAt time.Time `json:"recorded_at"`
	Step       string    `json:"step"`
	Node       string    `json:"node"`
	Action     string    `json:"action"`
	Command    []string  `json:"command,omitempty"`
	Stdout     string    `json:"stdout,omitempty"`
	Stderr     string    `json:"stderr,omitempty"`
	Error      string    `json:"error,omitempty"`
}

// RestoreFullNodeApplicationFromPortableSnapshot exercises the Cosmos SDK
// v0.50 snapshot export -> dump -> delete -> load -> restore contract against
// one real, stopped full node. Only data/application.db is moved; CometBFT's
// blockstore, consensus state, tx index, WAL, node keys, and snapshot store are
// preserved. The original application DB is restored automatically if the
// snapshot import, node restart, catch-up, or history verification fails.
func (n *Network) RestoreFullNodeApplicationFromPortableSnapshot(
	ctx context.Context,
	step string,
	fullNodeIndex int,
) (evidence ApplicationSnapshotRestoreEvidence, retErr error) {
	if strings.TrimSpace(step) == "" {
		return evidence, errors.New("snapshot restore step is required")
	}
	if len(n.Chain.Validators) == 0 {
		return evidence, errors.New("snapshot restore requires a validator")
	}
	if fullNodeIndex < 0 || fullNodeIndex >= len(n.Chain.FullNodes) {
		return evidence, fmt.Errorf("full-node index %d is out of range", fullNodeIndex)
	}

	n.txMu.Lock()
	defer n.txMu.Unlock()

	validator := n.Chain.Validators[0]
	node := n.Chain.FullNodes[fullNodeIndex]
	height, err := node.Height(ctx)
	if err != nil {
		return evidence, fmt.Errorf("read snapshot source height from %s: %w", node.Name(), err)
	}
	if height <= 0 {
		return evidence, fmt.Errorf("snapshot source height must be positive, got %d", height)
	}
	if err := n.WaitForNodeHeight(ctx, validator, height); err != nil {
		return evidence, err
	}
	before, err := n.RequireSameHistoryAtHeight(ctx, height, validator, node)
	if err != nil {
		return evidence, fmt.Errorf("verify source history at height %d: %w", height, err)
	}

	if err := n.runSnapshotRestoreNodeAction(ctx, step, node, "stop", nil, func() ([]byte, []byte, error) {
		return nil, nil, node.StopContainer(ctx)
	}); err != nil {
		startErr := n.runSnapshotRestoreNodeAction(ctx, step, node, "restart-after-stop-error", nil, func() ([]byte, []byte, error) {
			return nil, nil, node.StartContainer(ctx)
		})
		return evidence, errors.Join(fmt.Errorf("stop %s for snapshot restore: %w", node.Name(), err), startErr)
	}

	recoveryNeeded := true
	containerMayBeRunning := false
	var rollbackApplicationDB func() error
	defer func() {
		if retErr == nil || !recoveryNeeded {
			return
		}
		var recoveryErrors []error
		if containerMayBeRunning {
			recoveryErrors = append(recoveryErrors, n.runSnapshotRestoreNodeAction(
				ctx, step, node, "stop-before-rollback", nil,
				func() ([]byte, []byte, error) { return nil, nil, node.StopContainer(ctx) },
			))
		}
		if rollbackApplicationDB != nil {
			recoveryErrors = append(recoveryErrors, rollbackApplicationDB())
		}
		recoveryErrors = append(recoveryErrors, n.runSnapshotRestoreNodeAction(
			ctx, step, node, "restart-original", nil,
			func() ([]byte, []byte, error) { return nil, nil, node.StartContainer(ctx) },
		))
		retErr = errors.Join(append([]error{retErr}, recoveryErrors...)...)
	}()

	snapshot, err := n.ensureLocalApplicationSnapshot(ctx, step, node, uint64(height))
	if err != nil {
		return evidence, err
	}
	plan, err := newApplicationSnapshotRestorePlan(node.HomeDir(), snapshot.Height, snapshot.Format)
	if err != nil {
		return evidence, err
	}
	if err := n.requireSnapshotPathAbsent(ctx, step, node, "require-archive-absent", plan.Archive); err != nil {
		return evidence, err
	}
	if err := n.requireSnapshotPathAbsent(ctx, step, node, "require-backup-absent", plan.BackupDB); err != nil {
		return evidence, err
	}
	if err := n.requireSnapshotDirectory(ctx, step, node, "require-application-db", plan.ApplicationDB); err != nil {
		return evidence, err
	}

	heightText := strconv.FormatUint(snapshot.Height, 10)
	formatText := strconv.FormatUint(uint64(snapshot.Format), 10)
	if _, err := n.runSnapshotRestoreBin(
		ctx, step, node, "snapshot-dump",
		"snapshots", "dump", heightText, formatText, "--output", plan.Archive,
	); err != nil {
		return evidence, fmt.Errorf("dump portable snapshot archive: %w", err)
	}
	if _, err := n.runSnapshotRestoreBin(
		ctx, step, node, "snapshot-delete-before-load",
		"snapshots", "delete", heightText, formatText,
	); err != nil {
		return evidence, fmt.Errorf("delete local snapshot before portable load: %w", err)
	}
	if _, err := n.runSnapshotRestoreBin(
		ctx, step, node, "snapshot-load", "snapshots", "load", plan.Archive,
	); err != nil {
		return evidence, fmt.Errorf("load portable snapshot archive: %w", err)
	}
	loadedSnapshots, err := n.listLocalApplicationSnapshots(ctx, step, node, "snapshot-list-after-load")
	if err != nil {
		return evidence, err
	}
	loaded, found, err := selectLocalSnapshotAtHeight(loadedSnapshots, snapshot.Height)
	if err != nil {
		return evidence, err
	}
	if !found || loaded != snapshot {
		return evidence, fmt.Errorf("portable snapshot metadata changed: before=%+v after=%+v found=%t", snapshot, loaded, found)
	}

	rollbackApplicationDB, err = executeApplicationDBSwap(applicationDBSwapOperations{
		MoveOriginalAside: func() error {
			return n.runSnapshotRestoreExec(ctx, step, node, "move-application-db-aside", "mv", "--", plan.ApplicationDB, plan.BackupDB)
		},
		RestoreSnapshot: func() error {
			_, restoreErr := n.runSnapshotRestoreBin(
				ctx, step, node, "snapshot-restore", "snapshots", "restore", heightText, formatText,
			)
			return restoreErr
		},
		ValidateRestored: func() error {
			return n.requireSnapshotDirectory(ctx, step, node, "validate-restored-application-db", plan.ApplicationDB)
		},
		RemoveRestored: func() error {
			return n.runSnapshotRestoreExec(ctx, step, node, "remove-restored-application-db", "rm", "-rf", "--", plan.ApplicationDB)
		},
		MoveOriginalBack: func() error {
			return n.runSnapshotRestoreExec(ctx, step, node, "move-original-application-db-back", "mv", "--", plan.BackupDB, plan.ApplicationDB)
		},
	})
	if err != nil {
		return evidence, err
	}

	containerMayBeRunning = true
	if err := n.runSnapshotRestoreNodeAction(ctx, step, node, "start-restored", nil, func() ([]byte, []byte, error) {
		return nil, nil, node.StartContainer(ctx)
	}); err != nil {
		return evidence, fmt.Errorf("start restored full node %s: %w", node.Name(), err)
	}
	if err := n.WaitForNodeHeight(ctx, node, height+1); err != nil {
		return evidence, fmt.Errorf("wait for restored full node progress: %w", err)
	}
	restoredHistory, err := n.RequireSameHistoryAtHeight(ctx, height, validator, node)
	if err != nil {
		return evidence, fmt.Errorf("restored history differs at snapshot height %d: %w", height, err)
	}
	targetHeight, err := validator.Height(ctx)
	if err != nil {
		return evidence, fmt.Errorf("read validator catch-up target: %w", err)
	}
	if err := n.WaitForNodeHeight(ctx, node, targetHeight); err != nil {
		return evidence, fmt.Errorf("wait for restored full node catch-up to %d: %w", targetHeight, err)
	}
	caughtUpHistory, err := n.RequireSameHistoryAtHeight(ctx, targetHeight, validator, node)
	if err != nil {
		return evidence, fmt.Errorf("restored node history differs at catch-up height %d: %w", targetHeight, err)
	}
	applicationInfo, err := node.Client.ABCIInfo(ctx)
	if err != nil {
		return evidence, fmt.Errorf("query restored application info: %w", err)
	}
	if applicationInfo == nil || applicationInfo.Response.LastBlockHeight < targetHeight {
		return evidence, fmt.Errorf(
			"restored application height is behind: got %d want at least %d",
			applicationInfo.Response.LastBlockHeight,
			targetHeight,
		)
	}
	applicationHash := strings.ToUpper(fmt.Sprintf("%X", applicationInfo.Response.LastBlockAppHash))
	if applicationHash == "" {
		return evidence, errors.New("restored application returned an empty app hash")
	}

	evidence = ApplicationSnapshotRestoreEvidence{
		RecordedAt:      time.Now().UTC(),
		Step:            step,
		Node:            node.Name(),
		Snapshot:        snapshot,
		Archive:         plan.Archive,
		BackupDB:        plan.BackupDB,
		Before:          before,
		RestoredHistory: restoredHistory,
		CaughtUpHistory: caughtUpHistory,
	}
	evidence.Application.Height = applicationInfo.Response.LastBlockHeight
	evidence.Application.AppHash = applicationHash
	if err := n.artifacts.writeJSON("recovery/snapshot-restore.json", evidence); err != nil {
		return evidence, fmt.Errorf("record snapshot restore evidence: %w", err)
	}
	recoveryNeeded = false
	return evidence, nil
}

// AddAndSyncFreshFullNode uses Interchaintest's supported AddFullNodes path.
// It validates ordinary block sync from an empty node volume. This is distinct
// from CometBFT state sync: v8.8.1 does not expose an atomic helper that wires
// trusted height/hash and RPC servers into a dynamically added node.
func (n *Network) AddAndSyncFreshFullNode(
	ctx context.Context,
	step string,
) (FreshFullNodeSyncEvidence, error) {
	if strings.TrimSpace(step) == "" {
		return FreshFullNodeSyncEvidence{}, errors.New("fresh full-node sync step is required")
	}
	if len(n.Chain.Validators) == 0 {
		return FreshFullNodeSyncEvidence{}, errors.New("fresh full-node sync requires a validator")
	}
	n.txMu.Lock()
	defer n.txMu.Unlock()

	validator := n.Chain.Validators[0]
	previousCount := len(n.Chain.FullNodes)
	startHeight, err := validator.Height(ctx)
	if err != nil {
		return FreshFullNodeSyncEvidence{}, fmt.Errorf("read fresh sync start height: %w", err)
	}
	overrides := map[string]any{
		"config/config.toml": testutil.Toml{"db_backend": "goleveldb"},
	}
	if err := n.Chain.AddFullNodes(ctx, overrides, 1); err != nil {
		err = fmt.Errorf("add fresh full node: %w", err)
		n.artifacts.recordFailure("fresh-full-node-add", err)
		return FreshFullNodeSyncEvidence{}, err
	}
	if len(n.Chain.FullNodes) != previousCount+1 {
		return FreshFullNodeSyncEvidence{}, fmt.Errorf(
			"full-node count after add is %d, want %d",
			len(n.Chain.FullNodes),
			previousCount+1,
		)
	}
	node := n.Chain.FullNodes[previousCount]
	if err := n.WaitForNodeHeight(ctx, node, startHeight); err != nil {
		return FreshFullNodeSyncEvidence{}, err
	}
	targetHeight, err := validator.Height(ctx)
	if err != nil {
		return FreshFullNodeSyncEvidence{}, fmt.Errorf("read fresh sync target height: %w", err)
	}
	if err := n.WaitForNodeHeight(ctx, node, targetHeight); err != nil {
		return FreshFullNodeSyncEvidence{}, err
	}
	history, err := n.RequireSameHistoryAtHeight(ctx, targetHeight, validator, node)
	if err != nil {
		return FreshFullNodeSyncEvidence{}, fmt.Errorf("fresh full node history differs at %d: %w", targetHeight, err)
	}
	evidence := FreshFullNodeSyncEvidence{
		RecordedAt:        time.Now().UTC(),
		Step:              step,
		Node:              node.Name(),
		PreviousNodeCount: previousCount,
		TargetHeight:      targetHeight,
		History:           history,
	}
	if err := n.artifacts.writeJSON("recovery/fresh-full-node-sync.json", evidence); err != nil {
		return evidence, fmt.Errorf("record fresh full-node sync evidence: %w", err)
	}
	return evidence, nil
}

func (n *Network) ensureLocalApplicationSnapshot(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	height uint64,
) (LocalSnapshot, error) {
	snapshots, err := n.listLocalApplicationSnapshots(ctx, step, node, "snapshot-list-before-export")
	if err != nil {
		return LocalSnapshot{}, err
	}
	selected, found, err := selectLocalSnapshotAtHeight(snapshots, height)
	if err != nil {
		return LocalSnapshot{}, err
	}
	if found {
		return selected, nil
	}
	if _, err := n.runSnapshotRestoreBin(
		ctx,
		step,
		node,
		"snapshot-export",
		"snapshots", "export", "--height", strconv.FormatUint(height, 10),
	); err != nil {
		return LocalSnapshot{}, fmt.Errorf("export local application snapshot at %d: %w", height, err)
	}
	snapshots, err = n.listLocalApplicationSnapshots(ctx, step, node, "snapshot-list-after-export")
	if err != nil {
		return LocalSnapshot{}, err
	}
	selected, found, err = selectLocalSnapshotAtHeight(snapshots, height)
	if err != nil {
		return LocalSnapshot{}, err
	}
	if !found {
		return LocalSnapshot{}, fmt.Errorf("exported snapshot at height %d was not listed", height)
	}
	return selected, nil
}

func (n *Network) listLocalApplicationSnapshots(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	action string,
) ([]LocalSnapshot, error) {
	stdout, err := n.runSnapshotRestoreBin(ctx, step, node, action, "snapshots", "list")
	if err != nil {
		return nil, fmt.Errorf("list local application snapshots: %w", err)
	}
	snapshots, err := ParseLocalSnapshots(stdout)
	if err != nil {
		return nil, err
	}
	return snapshots, nil
}

func (n *Network) requireSnapshotPathAbsent(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	action string,
	path string,
) error {
	return n.runSnapshotRestoreExec(ctx, step, node, action, "test", "!", "-e", path)
}

func (n *Network) requireSnapshotDirectory(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	action string,
	path string,
) error {
	return n.runSnapshotRestoreExec(ctx, step, node, action, "test", "-d", path)
}

func (n *Network) runSnapshotRestoreBin(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	action string,
	command ...string,
) ([]byte, error) {
	fullCommand := node.BinCommand(command...)
	var stdout []byte
	err := n.runSnapshotRestoreNodeAction(ctx, step, node, action, fullCommand, func() ([]byte, []byte, error) {
		var stderr []byte
		var operationErr error
		stdout, stderr, operationErr = node.Exec(ctx, fullCommand, node.Chain.Config().Env)
		return stdout, stderr, operationErr
	})
	return stdout, err
}

func (n *Network) runSnapshotRestoreExec(
	ctx context.Context,
	step string,
	node *cosmos.ChainNode,
	action string,
	command ...string,
) error {
	return n.runSnapshotRestoreNodeAction(ctx, step, node, action, command, func() ([]byte, []byte, error) {
		return node.Exec(ctx, command, node.Chain.Config().Env)
	})
}

func (n *Network) runSnapshotRestoreNodeAction(
	_ context.Context,
	step string,
	node *cosmos.ChainNode,
	action string,
	command []string,
	operation func() ([]byte, []byte, error),
) error {
	stdout, stderr, operationErr := operation()
	record := snapshotRestoreAction{
		RecordedAt: time.Now().UTC(),
		Step:       step,
		Node:       node.Name(),
		Action:     action,
		Command:    append([]string(nil), command...),
		Stdout:     boundedArtifactText(stdout),
		Stderr:     boundedArtifactText(stderr),
		Error:      errorString(operationErr),
	}
	artifactErr := n.artifacts.appendJSONLine("recovery/snapshot-restore-actions.jsonl", record)
	if operationErr != nil {
		n.artifacts.recordFailure("snapshot-restore-"+action, operationErr)
	}
	if artifactErr != nil {
		n.artifacts.recordFailure("snapshot-restore-action-artifact", artifactErr)
	}
	return errors.Join(operationErr, artifactErr)
}
