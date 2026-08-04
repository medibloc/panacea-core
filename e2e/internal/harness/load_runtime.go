package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	cmttypes "github.com/cometbft/cometbft/types"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

// LoadNodeRuntimeSample is a best-effort operational snapshot. Missing
// readings remain nil and are explained in Unavailable; callers decide which
// readings are required for their baseline.
type LoadNodeRuntimeSample struct {
	Label                 string            `json:"label"`
	RecordedAt            time.Time         `json:"recorded_at"`
	Node                  string            `json:"node"`
	Role                  string            `json:"role"`
	ContainerID           string            `json:"container_id"`
	ProcessID             *int              `json:"process_id,omitempty"`
	Height                *int64            `json:"height,omitempty"`
	CatchingUp            *bool             `json:"catching_up,omitempty"`
	Peers                 *int              `json:"peers,omitempty"`
	MempoolTransactions   *int              `json:"mempool_transactions,omitempty"`
	MempoolBytes          *int64            `json:"mempool_bytes,omitempty"`
	CPUPercent            *float64          `json:"cpu_percent,omitempty"`
	MemoryUsageBytes      *uint64           `json:"memory_usage_bytes,omitempty"`
	RSSBytes              *uint64           `json:"rss_bytes,omitempty"`
	PIDCount              *uint64           `json:"pid_count,omitempty"`
	OpenFiles             *int              `json:"open_files,omitempty"`
	DBSizeBytes           *uint64           `json:"db_size_bytes,omitempty"`
	BlockDeviceWriteBytes *uint64           `json:"block_device_write_bytes,omitempty"`
	Goroutines            *uint64           `json:"goroutines,omitempty"`
	Unavailable           map[string]string `json:"unavailable,omitempty"`
}

// QueryGasOverrideEvidence records each bounded configuration transition.
type QueryGasOverrideEvidence struct {
	RecordedAt    time.Time `json:"recorded_at"`
	Phase         string    `json:"phase"`
	Node          string    `json:"node"`
	PreviousLimit uint64    `json:"previous_limit"`
	CurrentLimit  uint64    `json:"current_limit"`
	Height        int64     `json:"height"`
}

// FullNodeQueryGasOverride owns a reversible full-node-only app.toml change.
type FullNodeQueryGasOverride struct {
	network       *Network
	node          *cosmos.ChainNode
	original      []byte
	originalLimit uint64
	appliedLimit  uint64

	mu       sync.Mutex
	restored bool
}

// ApplyFullNodeQueryGasLimit restarts only the query full node with a bounded
// query gas limit. Validators and their application configuration are never
// touched.
func (n *Network) ApplyFullNodeQueryGasLimit(
	ctx context.Context,
	limit uint64,
) (*FullNodeQueryGasOverride, error) {
	if len(n.Chain.FullNodes) == 0 {
		return nil, errors.New("network has no full node")
	}
	node := n.Chain.FullNodes[0]
	original, err := node.ReadFile(ctx, "config/app.toml")
	if err != nil {
		return nil, fmt.Errorf("read %s app.toml: %w", node.Name(), err)
	}
	rewritten, previous, err := RewriteQueryGasLimit(original, limit)
	if err != nil {
		return nil, fmt.Errorf("rewrite %s app.toml: %w", node.Name(), err)
	}
	restartHeight, err := node.Height(ctx)
	if err != nil {
		return nil, fmt.Errorf("read %s height before query gas override: %w", node.Name(), err)
	}
	override := &FullNodeQueryGasOverride{
		network:       n,
		node:          node,
		original:      append([]byte(nil), original...),
		originalLimit: previous,
		appliedLimit:  limit,
	}
	if err := replaceRunningNodeAppConfig(ctx, node, rewritten, restartHeight); err != nil {
		return nil, recoverOriginalAppConfig(node, original, restartHeight, fmt.Errorf("apply full-node query gas limit: %w", err))
	}
	height, err := node.Height(ctx)
	if err != nil {
		return nil, recoverOriginalAppConfig(node, original, restartHeight, fmt.Errorf("read full-node height after query gas override: %w", err))
	}
	evidence := QueryGasOverrideEvidence{
		RecordedAt:    time.Now().UTC(),
		Phase:         "applied",
		Node:          node.Name(),
		PreviousLimit: previous,
		CurrentLimit:  limit,
		Height:        height,
	}
	if err := n.artifacts.appendJSONLine("metrics/query-gas-overrides.jsonl", evidence); err != nil {
		return nil, recoverOriginalAppConfig(node, original, restartHeight, fmt.Errorf("record query gas override: %w", err))
	}
	return override, nil
}

// Restore returns the full node to its byte-for-byte original app.toml and is
// idempotent so tests can call it both explicitly and from cleanup.
func (o *FullNodeQueryGasOverride) Restore(ctx context.Context) error {
	if o == nil {
		return nil
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.restored {
		return nil
	}
	restartHeight, err := o.node.Height(ctx)
	if err != nil {
		return fmt.Errorf("read %s height before query gas restore: %w", o.node.Name(), err)
	}
	if err := replaceRunningNodeAppConfig(ctx, o.node, o.original, restartHeight); err != nil {
		return fmt.Errorf("restore full-node query gas limit: %w", err)
	}
	o.restored = true
	height, err := o.node.Height(ctx)
	if err != nil {
		return fmt.Errorf("read full-node height after query gas restore: %w", err)
	}
	evidence := QueryGasOverrideEvidence{
		RecordedAt:    time.Now().UTC(),
		Phase:         "restored",
		Node:          o.node.Name(),
		PreviousLimit: o.appliedLimit,
		CurrentLimit:  o.originalLimit,
		Height:        height,
	}
	if err := o.network.artifacts.appendJSONLine("metrics/query-gas-overrides.jsonl", evidence); err != nil {
		return fmt.Errorf("record query gas restore: %w", err)
	}
	return nil
}

func recoverOriginalAppConfig(node *cosmos.ChainNode, original []byte, restartHeight int64, cause error) error {
	recoveryCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	recoveryErr := replaceRunningNodeAppConfig(recoveryCtx, node, original, restartHeight)
	if recoveryErr != nil {
		recoveryErr = fmt.Errorf("recover original %s app.toml: %w", node.Name(), recoveryErr)
	}
	return errors.Join(cause, recoveryErr)
}

func replaceRunningNodeAppConfig(ctx context.Context, node *cosmos.ChainNode, contents []byte, restartHeight int64) error {
	if restartHeight <= 0 {
		return errors.New("node restart readiness height must be positive")
	}
	restartCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err := node.StopContainer(restartCtx); err != nil {
		return fmt.Errorf("stop %s: %w", node.Name(), err)
	}
	if err := node.WriteFile(restartCtx, contents, "config/app.toml"); err != nil {
		startErr := node.StartContainer(restartCtx)
		return errors.Join(fmt.Errorf("write %s app.toml: %w", node.Name(), err), startErr)
	}
	if err := node.StartContainer(restartCtx); err != nil {
		return fmt.Errorf("start %s: %w", node.Name(), err)
	}
	if err := waitForHeight(restartCtx, restartHeight, node.Height); err != nil {
		return fmt.Errorf("wait for %s readiness after app.toml change: %w", node.Name(), err)
	}
	return nil
}

// CaptureLoadRuntimeMetrics snapshots validators and full nodes and writes one
// JSONL record per node. Capture failures stay in each sample's Unavailable map
// so the artifact is written before callers enforce their baseline contract.
func (n *Network) CaptureLoadRuntimeMetrics(
	ctx context.Context,
	label string,
) ([]LoadNodeRuntimeSample, error) {
	samples := make([]LoadNodeRuntimeSample, 0, len(n.Chain.Validators)+len(n.Chain.FullNodes))
	for _, node := range n.Chain.Validators {
		samples = append(samples, captureLoadNodeRuntime(ctx, label, "validator", node))
	}
	for _, node := range n.Chain.FullNodes {
		samples = append(samples, captureLoadNodeRuntime(ctx, label, "full_node", node))
	}
	for _, sample := range samples {
		if err := n.artifacts.appendJSONLine("metrics/node-runtime.jsonl", sample); err != nil {
			return samples, fmt.Errorf("record runtime metrics for %s: %w", sample.Node, err)
		}
	}
	return samples, nil
}

func captureLoadNodeRuntime(
	ctx context.Context,
	label string,
	role string,
	node *cosmos.ChainNode,
) LoadNodeRuntimeSample {
	sample := LoadNodeRuntimeSample{
		Label:       label,
		RecordedAt:  time.Now().UTC(),
		Node:        node.Name(),
		Role:        role,
		ContainerID: node.ContainerID(),
		Unavailable: make(map[string]string),
	}
	status, err := node.Client.Status(ctx)
	if err != nil {
		sample.Unavailable["status"] = err.Error()
	} else if status == nil {
		sample.Unavailable["status"] = "RPC returned an empty result"
	} else {
		height := status.SyncInfo.LatestBlockHeight
		catchingUp := status.SyncInfo.CatchingUp
		sample.Height = &height
		sample.CatchingUp = &catchingUp
	}
	if netInfo, netErr := node.Client.NetInfo(ctx); netErr != nil {
		sample.Unavailable["peers"] = netErr.Error()
	} else if netInfo == nil {
		sample.Unavailable["peers"] = "RPC returned an empty result"
	} else {
		peers := netInfo.NPeers
		sample.Peers = &peers
	}
	if mempool, mempoolErr := node.Client.NumUnconfirmedTxs(ctx); mempoolErr != nil {
		sample.Unavailable["mempool"] = mempoolErr.Error()
	} else if mempool == nil {
		sample.Unavailable["mempool"] = "RPC returned an empty result"
	} else {
		count := mempool.Count
		mempoolBytes := mempool.TotalBytes
		sample.MempoolTransactions = &count
		sample.MempoolBytes = &mempoolBytes
	}

	inspect, inspectErr := node.DockerClient.ContainerInspect(ctx, node.ContainerID())
	if inspectErr != nil {
		sample.Unavailable["process_id"] = inspectErr.Error()
	} else if inspect.State != nil {
		pid := inspect.State.Pid
		sample.ProcessID = &pid
	}
	statsResponse, statsErr := node.DockerClient.ContainerStatsOneShot(ctx, node.ContainerID())
	if statsErr != nil {
		sample.Unavailable["docker_stats"] = statsErr.Error()
	} else {
		var stats dockertypes.StatsJSON
		decodeErr := json.NewDecoder(statsResponse.Body).Decode(&stats)
		closeErr := statsResponse.Body.Close()
		if decodeErr != nil || closeErr != nil {
			sample.Unavailable["docker_stats"] = errors.Join(decodeErr, closeErr).Error()
		} else {
			cpu := dockerCPUPercent(stats)
			sample.CPUPercent = &cpu
			memory := stats.MemoryStats.Usage
			sample.MemoryUsageBytes = &memory
			rss := dockerRSS(stats)
			sample.RSSBytes = &rss
			pids := stats.PidsStats.Current
			sample.PIDCount = &pids
			writes := dockerBlockWriteBytes(stats)
			sample.BlockDeviceWriteBytes = &writes
		}
	}

	if stdout, stderr, execErr := execInRunningNode(ctx, node, []string{"sh", "-c", "ls -1 /proc/1/fd"}); execErr != nil {
		sample.Unavailable["open_files"] = boundedExecError(execErr, stderr)
	} else {
		openFiles := countNonEmptyLines(stdout)
		sample.OpenFiles = &openFiles
	}
	if stdout, stderr, execErr := execInRunningNode(ctx, node, []string{"du", "-sk", node.HomeDir() + "/data"}); execErr != nil {
		sample.Unavailable["db_size_bytes"] = boundedExecError(execErr, stderr)
	} else {
		fields := strings.Fields(string(stdout))
		kilobytes, parseErr := parseFirstUint64(fields)
		if parseErr != nil {
			sample.Unavailable["db_size_bytes"] = parseErr.Error()
		} else {
			bytes := kilobytes * 1024
			sample.DBSizeBytes = &bytes
		}
	}
	if goroutines, goroutineErr := loadNodeGoroutines(ctx, node); goroutineErr != nil {
		sample.Unavailable["goroutines"] = goroutineErr.Error()
	} else {
		sample.Goroutines = &goroutines
	}
	return sample
}

func loadNodeGoroutines(ctx context.Context, node *cosmos.ChainNode) (uint64, error) {
	address, err := node.GetHostAddress(ctx, "1317/tcp")
	if err != nil {
		return 0, fmt.Errorf("resolve API metrics address: %w", err)
	}
	address, err = normalizeHostAddress(address)
	if err != nil {
		return 0, fmt.Errorf("normalize API metrics address: %w", err)
	}
	body, err := getJSON(
		ctx,
		&http.Client{Timeout: 10 * time.Second},
		strings.TrimRight(address, "/")+"/metrics?format=prometheus",
	)
	if err != nil {
		return 0, fmt.Errorf("query API Prometheus metrics: %w", err)
	}
	return parsePrometheusGoroutines(body)
}

func parsePrometheusGoroutines(contents []byte) (uint64, error) {
	for _, line := range strings.Split(string(contents), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := strings.SplitN(fields[0], "{", 2)[0]
		if name != "go_goroutines" && name != "runtime_num_goroutines" &&
			!strings.HasSuffix(name, "_runtime_num_goroutines") {
			continue
		}
		value, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			return 0, fmt.Errorf("parse goroutine metric %q: %w", fields[1], err)
		}
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || math.Trunc(value) != value || value > math.MaxUint64 {
			return 0, fmt.Errorf("goroutine metric must be a non-negative integer, got %q", fields[1])
		}
		return uint64(value), nil
	}
	return 0, errors.New("Prometheus response does not contain a goroutine metric")
}

func dockerCPUPercent(stats dockertypes.StatsJSON) float64 {
	if stats.CPUStats.CPUUsage.TotalUsage < stats.PreCPUStats.CPUUsage.TotalUsage ||
		stats.CPUStats.SystemUsage <= stats.PreCPUStats.SystemUsage {
		return 0
	}
	cpuDelta := stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage
	systemDelta := stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage
	cpus := stats.CPUStats.OnlineCPUs
	if cpus == 0 {
		cpus = uint32(len(stats.CPUStats.CPUUsage.PercpuUsage))
	}
	if systemDelta == 0 || cpus == 0 {
		return 0
	}
	return float64(cpuDelta) / float64(systemDelta) * float64(cpus) * 100
}

func dockerRSS(stats dockertypes.StatsJSON) uint64 {
	if rss, ok := stats.MemoryStats.Stats["rss"]; ok {
		return rss
	}
	cache := stats.MemoryStats.Stats["cache"]
	if cache <= stats.MemoryStats.Usage {
		return stats.MemoryStats.Usage - cache
	}
	return stats.MemoryStats.Usage
}

func dockerBlockWriteBytes(stats dockertypes.StatsJSON) uint64 {
	var total uint64
	for _, entry := range stats.BlkioStats.IoServiceBytesRecursive {
		if strings.EqualFold(entry.Op, "write") {
			total += entry.Value
		}
	}
	return total
}

func parseFirstUint64(fields []string) (uint64, error) {
	if len(fields) == 0 {
		return 0, errors.New("command returned no numeric value")
	}
	value, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse command numeric value %q: %w", fields[0], err)
	}
	return value, nil
}

func countNonEmptyLines(contents []byte) int {
	count := 0
	for _, line := range strings.Split(strings.TrimSpace(string(contents)), "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

func boundedExecError(err error, stderr []byte) string {
	return fmt.Sprintf("%v: %s", err, boundedString(stderr, 4<<10))
}

func execInRunningNode(
	ctx context.Context,
	node *cosmos.ChainNode,
	command []string,
) ([]byte, []byte, error) {
	created, err := node.DockerClient.ContainerExecCreate(ctx, node.ContainerID(), dockertypes.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          append([]string(nil), command...),
	})
	if err != nil {
		return nil, nil, err
	}
	attached, err := node.DockerClient.ContainerExecAttach(ctx, created.ID, dockertypes.ExecStartCheck{})
	if err != nil {
		return nil, nil, err
	}
	defer attached.Close()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	_, copyErr := stdcopy.StdCopy(&stdout, &stderr, attached.Reader)
	if copyErr != nil && !errors.Is(copyErr, io.EOF) {
		return stdout.Bytes(), stderr.Bytes(), copyErr
	}
	inspected, err := node.DockerClient.ContainerExecInspect(ctx, created.ID)
	if err != nil {
		return stdout.Bytes(), stderr.Bytes(), err
	}
	if inspected.ExitCode != 0 {
		return stdout.Bytes(), stderr.Bytes(), fmt.Errorf("container exec exited with code %d", inspected.ExitCode)
	}
	return stdout.Bytes(), stderr.Bytes(), nil
}

// CollectLoadBlockMetrics reads exact validator blocks and FinalizeBlock gas
// results for (startExclusive, endInclusive].
func (n *Network) CollectLoadBlockMetrics(
	ctx context.Context,
	node *cosmos.ChainNode,
	startExclusive int64,
	endInclusive int64,
) ([]LoadBlockSample, error) {
	if node == nil {
		return nil, errors.New("block metrics node is required")
	}
	if startExclusive <= 0 || endInclusive <= startExclusive {
		return nil, fmt.Errorf("invalid block metric range (%d, %d]", startExclusive, endInclusive)
	}
	previous, err := node.Client.Block(ctx, &startExclusive)
	if err != nil || previous == nil || previous.Block == nil {
		return nil, fmt.Errorf("query starting block %d: %w", startExclusive, err)
	}
	previousTime := previous.Block.Time
	samples := make([]LoadBlockSample, 0, endInclusive-startExclusive)
	for height := startExclusive + 1; height <= endInclusive; height++ {
		block, blockErr := node.Client.Block(ctx, &height)
		if blockErr != nil || block == nil || block.Block == nil {
			return samples, fmt.Errorf("query block %d: %w", height, blockErr)
		}
		results, resultsErr := node.Client.BlockResults(ctx, &height)
		if resultsErr != nil || results == nil {
			return samples, fmt.Errorf("query block results %d: %w", height, resultsErr)
		}
		sample := LoadBlockSample{
			Height:               height,
			Time:                 block.Block.Time,
			IntervalMilliseconds: float64(block.Block.Time.Sub(previousTime)) / float64(time.Millisecond),
			Transactions:         len(block.Block.Data.Txs),
		}
		for _, txResult := range results.TxsResults {
			if txResult == nil {
				continue
			}
			sample.GasWanted += txResult.GasWanted
			sample.GasUsed += txResult.GasUsed
			if txResult.Code != 0 {
				sample.FailedTransactions++
			}
		}
		if block.Block.LastCommit != nil {
			sample.CommitSignatures = len(block.Block.LastCommit.Signatures)
			for _, signature := range block.Block.LastCommit.Signatures {
				if signature.BlockIDFlag != cmttypes.BlockIDFlagCommit {
					sample.MissedSignatures++
				}
			}
		}
		if err := n.artifacts.appendJSONLine("metrics/blocks.jsonl", sample); err != nil {
			return samples, fmt.Errorf("record block metrics %d: %w", height, err)
		}
		samples = append(samples, sample)
		previousTime = block.Block.Time
	}
	return samples, nil
}
