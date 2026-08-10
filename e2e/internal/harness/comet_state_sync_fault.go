package harness

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

const (
	defaultCometStateSyncFaultTimeout     = 30 * time.Second
	maximumCometStateSyncFaultTimeout     = 60 * time.Second
	cometStateSyncFaultDeadlineSlack      = 6 * time.Second
	minimumCorruptChunkSnapshotInterval   = uint64(20)
	minimumCorruptChunkSnapshotHeadroom   = uint64(12)
	unavailableCometStateSyncRPCServerOne = "http://127.0.0.1:1"
	unavailableCometStateSyncRPCServerTwo = "http://127.0.0.1:2"
)

// CometStateSyncUnavailableProviderEvidence proves that a fresh state-sync
// node fails explicitly and within a deadline when both configured light RPC
// providers are unreachable. The original healthy sources are still retained
// as provenance for the trust height/hash and provider snapshot.
type CometStateSyncUnavailableProviderEvidence struct {
	SchemaVersion      int                                          `json:"schema_version"`
	Mode               string                                       `json:"mode"`
	RecordedAt         time.Time                                    `json:"recorded_at"`
	CompletedAt        time.Time                                    `json:"completed_at"`
	Step               string                                       `json:"step"`
	Sources            []CometStateSyncSourceEvidence               `json:"sources"`
	Providers          []CometStateSyncProviderEvidence             `json:"providers"`
	TrustHistory       []BlockEvidence                              `json:"trust_history"`
	Config             CometStateSyncConfigEvidence                 `json:"config"`
	InjectedRPCServers []string                                     `json:"injected_rpc_servers"`
	Node               string                                       `json:"node"`
	NodeImage          ImageRef                                     `json:"node_image"`
	Volume             string                                       `json:"volume"`
	FreshDataInventory []string                                     `json:"fresh_data_inventory"`
	FailureTimeout     string                                       `json:"failure_timeout"`
	Elapsed            string                                       `json:"elapsed"`
	StartError         string                                       `json:"start_error,omitempty"`
	Logs               CometStateSyncUnavailableProviderLogEvidence `json:"logs"`
	ConfigArtifact     string                                       `json:"config_artifact"`
	LogArtifact        string                                       `json:"log_artifact"`
	EvidenceArtifact   string                                       `json:"evidence_artifact"`
	Rejected           bool                                         `json:"rejected"`
	NodeStopped        bool                                         `json:"node_stopped"`
	Error              string                                       `json:"error,omitempty"`
}

// CometStateSyncCorruptedChunkRequest adds the actual source snapshot cadence
// to the base request. The harness verifies this value from both app.toml
// files and requires enough block headroom before mutating chunk zero.
type CometStateSyncCorruptedChunkRequest struct {
	StateSync        CometStateSyncRequest
	SnapshotInterval uint64
	FailureTimeout   time.Duration
}

// CometStateSyncChunkMutationEvidence records the exact provider file mutated
// and proves byte-for-byte restoration before the fault method returns.
type CometStateSyncChunkMutationEvidence struct {
	Node            string `json:"node"`
	Height          uint64 `json:"height"`
	Format          uint32 `json:"format"`
	Chunk           uint32 `json:"chunk"`
	RelativePath    string `json:"relative_path"`
	Size            int    `json:"size"`
	OriginalSHA256  string `json:"original_sha256"`
	CorruptedSHA256 string `json:"corrupted_sha256"`
	RestoredSHA256  string `json:"restored_sha256,omitempty"`
	Restored        bool   `json:"restored"`
	RestoreError    string `json:"restore_error,omitempty"`
}

// CometStateSyncCorruptedChunkEvidence proves an actual Cosmos SDK checksum
// rejection and exhaustion of the two deliberately corrupted P2P providers.
type CometStateSyncCorruptedChunkEvidence struct {
	SchemaVersion           int                                     `json:"schema_version"`
	Mode                    string                                  `json:"mode"`
	RecordedAt              time.Time                               `json:"recorded_at"`
	CompletedAt             time.Time                               `json:"completed_at"`
	Step                    string                                  `json:"step"`
	Sources                 []CometStateSyncSourceEvidence          `json:"sources"`
	Providers               []CometStateSyncProviderEvidence        `json:"providers"`
	TrustHistory            []BlockEvidence                         `json:"trust_history"`
	Config                  CometStateSyncConfigEvidence            `json:"config"`
	SnapshotHeight          uint64                                  `json:"snapshot_height"`
	SnapshotFormat          uint32                                  `json:"snapshot_format"`
	SnapshotInterval        uint64                                  `json:"snapshot_interval"`
	SnapshotHeadroomBlocks  uint64                                  `json:"snapshot_headroom_blocks"`
	Mutations               []CometStateSyncChunkMutationEvidence   `json:"mutations"`
	Node                    string                                  `json:"node"`
	NodeImage               ImageRef                                `json:"node_image"`
	Volume                  string                                  `json:"volume"`
	FreshDataInventory      []string                                `json:"fresh_data_inventory"`
	FailureTimeout          string                                  `json:"failure_timeout"`
	Elapsed                 string                                  `json:"elapsed"`
	StartError              string                                  `json:"start_error,omitempty"`
	Logs                    CometStateSyncCorruptedChunkLogEvidence `json:"logs"`
	ProviderConfigArtifacts []string                                `json:"provider_config_artifacts"`
	ChunkListArtifacts      []string                                `json:"chunk_list_artifacts"`
	ConfigArtifact          string                                  `json:"config_artifact"`
	LogArtifact             string                                  `json:"log_artifact"`
	EvidenceArtifact        string                                  `json:"evidence_artifact"`
	Rejected                bool                                    `json:"rejected"`
	NodeStopped             bool                                    `json:"node_stopped"`
	Error                   string                                  `json:"error,omitempty"`
}

type cometStateSyncChunkPath struct {
	Format       uint32
	Chunk        uint32
	RelativePath string
}

type cometStateSyncOriginalChunk struct {
	Node          *cosmos.ChainNode
	EvidenceIndex int
	Contents      []byte
}

// ExpectCometStateSyncUnavailableProviders replaces both otherwise valid RPC
// URLs with deterministic loopback closed ports. Passing means the light
// client reports provider transport/setup failure without restoring a state.
func (n *Network) ExpectCometStateSyncUnavailableProviders(
	ctx context.Context,
	request CometStateSyncRequest,
	failureTimeout time.Duration,
) (evidence CometStateSyncUnavailableProviderEvidence, retErr error) {
	evidence = CometStateSyncUnavailableProviderEvidence{
		SchemaVersion:      1,
		Mode:               "actual-cometbft-state-sync-unavailable-providers",
		RecordedAt:         time.Now().UTC(),
		Step:               request.Step,
		InjectedRPCServers: []string{unavailableCometStateSyncRPCServerOne, unavailableCometStateSyncRPCServerTwo},
		ConfigArtifact:     "recovery/state-sync/unavailable-providers-config.toml",
		LogArtifact:        "recovery/state-sync/unavailable-providers.log",
		EvidenceArtifact:   "recovery/state-sync/unavailable-providers.json",
	}
	if n == nil || n.Chain == nil || n.artifacts == nil {
		return evidence, errors.New("state-sync network and artifact store are required")
	}
	defer func() {
		evidence.CompletedAt = time.Now().UTC()
		evidence.Error = errorString(retErr)
		artifactErr := n.artifacts.writeJSON(evidence.EvidenceArtifact, evidence)
		retErr = errors.Join(retErr, artifactErr)
		if retErr != nil {
			n.artifacts.recordFailure("comet-state-sync-unavailable-providers", retErr)
		}
	}()

	normalized, err := n.normalizeCometStateSyncRequest(request, false)
	if err != nil {
		return evidence, err
	}
	evidence.Step = normalized.Step
	failureTimeout, err = normalizeCometStateSyncFaultTimeout(failureTimeout)
	if err != nil {
		return evidence, err
	}
	evidence.FailureTimeout = failureTimeout.String()

	inputs, err := n.resolveCometStateSyncInputs(
		ctx,
		normalized,
		"recovery/state-sync/unavailable-provider-source",
	)
	evidence.Sources = inputs.Sources
	evidence.Providers = inputs.Providers
	evidence.TrustHistory = inputs.TrustHistory
	if err != nil {
		return evidence, err
	}
	inputs.Plan.RPCServers = append([]string(nil), evidence.InjectedRPCServers...)
	evidence.Config = inputs.Plan.configEvidence()

	node, err := n.newDetachedCometStateSyncNode(ctx, normalized.RPCSources[0])
	if err != nil {
		return evidence, err
	}
	evidence.Node = node.Name()
	evidence.NodeImage = imageRefFromNode(node)
	evidence.Volume = node.VolumeName
	defer func() {
		logs, stopped, cleanupErr := n.captureAndStopDetachedCometStateSyncNode(node, evidence.LogArtifact)
		evidence.NodeStopped = stopped
		if len(logs) > 0 {
			evidence.Logs = parseCometStateSyncUnavailableProviderLogs(logs)
		}
		retErr = errors.Join(retErr, cleanupErr)
	}()

	configContents, inventory, err := n.initializeDetachedCometStateSyncNodeWithOptions(
		ctx,
		node,
		normalized.RPCSources,
		inputs.Plan,
		detachedCometStateSyncNodeOptions{DisablePEX: true},
	)
	if err != nil {
		return evidence, err
	}
	evidence.FreshDataInventory = inventory
	if err := n.artifacts.write(evidence.ConfigArtifact, configContents); err != nil {
		return evidence, fmt.Errorf("record unavailable-provider config.toml: %w", err)
	}
	if err := node.CreateNodeContainer(ctx); err != nil {
		return evidence, fmt.Errorf("create unavailable-provider state-sync container: %w", err)
	}

	startedAt := time.Now()
	failureCtx, cancelFailure := context.WithTimeout(ctx, failureTimeout)
	startResult := make(chan error, 1)
	go func() {
		startResult <- startDetachedCometStateSyncNode(failureCtx, node)
	}()
	logs, logEvidence := waitForCometStateSyncUnavailableProviderFailure(failureCtx, node, startedAt.UTC())
	cancelFailure()
	startErr := <-startResult
	evidence.StartError = errorString(startErr)
	elapsed := time.Since(startedAt)
	evidence.Elapsed = elapsed.String()
	evidence.Logs = logEvidence
	if len(logs) > 0 {
		if err := n.artifacts.write(evidence.LogArtifact, logs); err != nil {
			return evidence, fmt.Errorf("record unavailable-provider logs: %w", err)
		}
	}
	if err := validateCometStateSyncUnavailableProviderFailure(evidence.Logs, elapsed, failureTimeout); err != nil {
		return evidence, err
	}
	evidence.Rejected = true
	return evidence, nil
}

// ExpectCometStateSyncCorruptedChunks flips one byte in chunk zero of the
// newest common provider snapshot, starts a fresh node against only those two
// peers, requires checksum rejection plus provider exhaustion, and restores
// the exact source bytes before returning.
func (n *Network) ExpectCometStateSyncCorruptedChunks(
	ctx context.Context,
	request CometStateSyncCorruptedChunkRequest,
) (evidence CometStateSyncCorruptedChunkEvidence, retErr error) {
	evidence = CometStateSyncCorruptedChunkEvidence{
		SchemaVersion:    1,
		Mode:             "actual-cometbft-state-sync-corrupted-chunks",
		RecordedAt:       time.Now().UTC(),
		Step:             request.StateSync.Step,
		SnapshotInterval: request.SnapshotInterval,
		ConfigArtifact:   "recovery/state-sync/corrupted-chunks-config.toml",
		LogArtifact:      "recovery/state-sync/corrupted-chunks.log",
		EvidenceArtifact: "recovery/state-sync/corrupted-chunks.json",
	}
	if n == nil || n.Chain == nil || n.artifacts == nil {
		return evidence, errors.New("state-sync network and artifact store are required")
	}
	defer func() {
		evidence.CompletedAt = time.Now().UTC()
		evidence.Error = errorString(retErr)
		artifactErr := n.artifacts.writeJSON(evidence.EvidenceArtifact, evidence)
		retErr = errors.Join(retErr, artifactErr)
		if retErr != nil {
			n.artifacts.recordFailure("comet-state-sync-corrupted-chunks", retErr)
		}
	}()

	normalized, err := n.normalizeCometStateSyncRequest(request.StateSync, false)
	if err != nil {
		return evidence, err
	}
	evidence.Step = normalized.Step
	if request.SnapshotInterval < minimumCorruptChunkSnapshotInterval {
		return evidence, fmt.Errorf(
			"corrupted-chunk snapshot interval must be at least %d blocks, got %d",
			minimumCorruptChunkSnapshotInterval,
			request.SnapshotInterval,
		)
	}
	failureTimeout, err := normalizeCometStateSyncFaultTimeout(request.FailureTimeout)
	if err != nil {
		return evidence, err
	}
	evidence.FailureTimeout = failureTimeout.String()

	inputs, err := n.resolveCometStateSyncInputs(
		ctx,
		normalized,
		"recovery/state-sync/corrupted-chunk-source",
	)
	evidence.Sources = inputs.Sources
	evidence.Providers = inputs.Providers
	evidence.TrustHistory = inputs.TrustHistory
	if err != nil {
		return evidence, err
	}
	inputs.Plan.ChunkFetchers = 1
	evidence.Config = inputs.Plan.configEvidence()

	targetHeight, ok := latestCommonProviderSnapshotHeight(inputs.Providers)
	if !ok || targetHeight <= 1 {
		return evidence, errors.New("corrupted-chunk fault requires a common completed provider snapshot")
	}
	evidence.SnapshotHeight = uint64(targetHeight)
	commonLatest := minimumCometStateSyncSourceHeight(inputs.Sources)
	if commonLatest < targetHeight+2 {
		return evidence, fmt.Errorf("provider snapshot %d does not retain verifiable H+2 at %d", targetHeight, commonLatest)
	}
	nextSnapshotHeight := evidence.SnapshotHeight + request.SnapshotInterval
	if uint64(commonLatest) >= nextSnapshotHeight {
		return evidence, fmt.Errorf(
			"provider snapshot %d is stale for interval %d at common height %d",
			targetHeight,
			request.SnapshotInterval,
			commonLatest,
		)
	}
	evidence.SnapshotHeadroomBlocks = nextSnapshotHeight - uint64(commonLatest)
	if evidence.SnapshotHeadroomBlocks < minimumCorruptChunkSnapshotHeadroom {
		return evidence, fmt.Errorf(
			"corrupted-chunk fault has only %d blocks before the next snapshot; require at least %d",
			evidence.SnapshotHeadroomBlocks,
			minimumCorruptChunkSnapshotHeadroom,
		)
	}

	chunkPaths := make([][]cometStateSyncChunkPath, len(normalized.RPCSources))
	for index, source := range normalized.RPCSources {
		appConfig, err := source.ReadFile(ctx, "config/app.toml")
		if err != nil {
			return evidence, fmt.Errorf("read provider %s app.toml: %w", source.Name(), err)
		}
		actualInterval, err := parseCometStateSyncProviderSnapshotInterval(appConfig)
		if err != nil {
			return evidence, fmt.Errorf("parse provider %s snapshot interval: %w", source.Name(), err)
		}
		if actualInterval != request.SnapshotInterval {
			return evidence, fmt.Errorf(
				"provider %s snapshot interval=%d, want %d",
				source.Name(),
				actualInterval,
				request.SnapshotInterval,
			)
		}
		configArtifact := fmt.Sprintf("recovery/state-sync/corrupted-chunk-provider-%d-app.toml", index)
		evidence.ProviderConfigArtifacts = append(evidence.ProviderConfigArtifacts, configArtifact)
		if err := n.artifacts.write(configArtifact, appConfig); err != nil {
			return evidence, fmt.Errorf("record provider %d app.toml: %w", index, err)
		}

		heightRoot := path.Join(
			source.HomeDir(),
			"data",
			"snapshots",
			strconv.FormatUint(evidence.SnapshotHeight, 10),
		)
		stdout, stderr, err := source.Exec(
			ctx,
			[]string{"find", heightRoot, "-mindepth", "2", "-maxdepth", "2", "-type", "f", "-print"},
			n.Chain.Config().Env,
		)
		listArtifact := fmt.Sprintf("recovery/state-sync/corrupted-chunk-provider-%d-files.txt", index)
		evidence.ChunkListArtifacts = append(evidence.ChunkListArtifacts, listArtifact)
		if artifactErr := n.artifacts.write(listArtifact, stdout); artifactErr != nil {
			return evidence, fmt.Errorf("record provider %d snapshot chunk list: %w", index, artifactErr)
		}
		if err != nil {
			return evidence, fmt.Errorf(
				"list provider %s snapshot %d chunks: %w: %s",
				source.Name(),
				targetHeight,
				err,
				boundedString(stderr, txStderrMaxBytes),
			)
		}
		chunkPaths[index], err = parseCometStateSyncSnapshotChunkPaths(source.HomeDir(), evidence.SnapshotHeight, stdout)
		if err != nil {
			return evidence, fmt.Errorf("parse provider %s snapshot chunks: %w", source.Name(), err)
		}
	}

	format, pathsByProvider, err := selectCommonCometStateSyncChunkZero(chunkPaths)
	if err != nil {
		return evidence, err
	}
	evidence.SnapshotFormat = format
	var originals []cometStateSyncOriginalChunk
	defer func() {
		retErr = errors.Join(retErr, restoreCometStateSyncProviderChunks(n, &evidence, originals))
	}()
	for index, source := range normalized.RPCSources {
		selected := pathsByProvider[index]
		original, err := source.ReadFile(ctx, selected.RelativePath)
		if err != nil {
			return evidence, fmt.Errorf("read provider %s snapshot chunk: %w", source.Name(), err)
		}
		if len(original) == 0 {
			return evidence, fmt.Errorf("provider %s snapshot chunk is empty", source.Name())
		}
		corrupted := append([]byte(nil), original...)
		corrupted[0] ^= 0xff
		mutation := CometStateSyncChunkMutationEvidence{
			Node:            source.Name(),
			Height:          evidence.SnapshotHeight,
			Format:          selected.Format,
			Chunk:           selected.Chunk,
			RelativePath:    selected.RelativePath,
			Size:            len(original),
			OriginalSHA256:  cometStateSyncSHA256(original),
			CorruptedSHA256: cometStateSyncSHA256(corrupted),
		}
		if mutation.OriginalSHA256 == mutation.CorruptedSHA256 {
			return evidence, fmt.Errorf("provider %s chunk mutation did not change SHA-256", source.Name())
		}
		evidence.Mutations = append(evidence.Mutations, mutation)
		originals = append(originals, cometStateSyncOriginalChunk{
			Node:          source,
			EvidenceIndex: len(evidence.Mutations) - 1,
			Contents:      append([]byte(nil), original...),
		})
		if err := source.WriteFile(ctx, corrupted, selected.RelativePath); err != nil {
			return evidence, fmt.Errorf("corrupt provider %s snapshot chunk: %w", source.Name(), err)
		}
		written, err := source.ReadFile(ctx, selected.RelativePath)
		if err != nil {
			return evidence, fmt.Errorf("verify corrupted provider %s snapshot chunk: %w", source.Name(), err)
		}
		if got := cometStateSyncSHA256(written); got != mutation.CorruptedSHA256 {
			return evidence, fmt.Errorf("provider %s corrupted chunk SHA-256=%s, want %s", source.Name(), got, mutation.CorruptedSHA256)
		}
	}

	node, err := n.newDetachedCometStateSyncNode(ctx, normalized.RPCSources[0])
	if err != nil {
		return evidence, err
	}
	evidence.Node = node.Name()
	evidence.NodeImage = imageRefFromNode(node)
	evidence.Volume = node.VolumeName
	defer func() {
		logs, stopped, cleanupErr := n.captureAndStopDetachedCometStateSyncNode(node, evidence.LogArtifact)
		evidence.NodeStopped = stopped
		if len(logs) > 0 {
			evidence.Logs = parseCometStateSyncCorruptedChunkLogs(logs)
		}
		retErr = errors.Join(retErr, cleanupErr)
	}()

	configContents, inventory, err := n.initializeDetachedCometStateSyncNodeWithOptions(
		ctx,
		node,
		normalized.RPCSources,
		inputs.Plan,
		detachedCometStateSyncNodeOptions{DisablePEX: true},
	)
	if err != nil {
		return evidence, err
	}
	evidence.FreshDataInventory = inventory
	if err := n.artifacts.write(evidence.ConfigArtifact, configContents); err != nil {
		return evidence, fmt.Errorf("record corrupted-chunk config.toml: %w", err)
	}
	if err := node.CreateNodeContainer(ctx); err != nil {
		return evidence, fmt.Errorf("create corrupted-chunk state-sync container: %w", err)
	}

	startedAt := time.Now()
	failureCtx, cancelFailure := context.WithTimeout(ctx, failureTimeout)
	startResult := make(chan error, 1)
	go func() {
		startResult <- startDetachedCometStateSyncNode(failureCtx, node)
	}()
	logs, logEvidence := waitForCometStateSyncCorruptedChunkFailure(failureCtx, node, startedAt.UTC())
	cancelFailure()
	startErr := <-startResult
	evidence.StartError = errorString(startErr)
	elapsed := time.Since(startedAt)
	evidence.Elapsed = elapsed.String()
	evidence.Logs = logEvidence
	if len(logs) > 0 {
		if err := n.artifacts.write(evidence.LogArtifact, logs); err != nil {
			return evidence, fmt.Errorf("record corrupted-chunk logs: %w", err)
		}
	}
	if err := validateCometStateSyncCorruptedChunkFailure(evidence.Logs, elapsed, failureTimeout); err != nil {
		return evidence, err
	}
	evidence.Rejected = true
	return evidence, nil
}

func normalizeCometStateSyncFaultTimeout(timeout time.Duration) (time.Duration, error) {
	if timeout == 0 {
		timeout = defaultCometStateSyncFaultTimeout
	}
	if timeout < 10*time.Second || timeout > maximumCometStateSyncFaultTimeout {
		return 0, fmt.Errorf("state-sync fault timeout must be within [10s,%s], got %s", maximumCometStateSyncFaultTimeout, timeout)
	}
	return timeout, nil
}

func minimumCometStateSyncSourceHeight(sources []CometStateSyncSourceEvidence) int64 {
	if len(sources) == 0 {
		return 0
	}
	minimum := sources[0].LatestHeight
	for _, source := range sources[1:] {
		if source.LatestHeight < minimum {
			minimum = source.LatestHeight
		}
	}
	return minimum
}

func parseCometStateSyncProviderSnapshotInterval(contents []byte) (uint64, error) {
	tree, err := toml.LoadBytes(contents)
	if err != nil {
		return 0, fmt.Errorf("parse app.toml: %w", err)
	}
	raw := tree.Get("state-sync.snapshot-interval")
	interval, err := strconv.ParseUint(fmt.Sprint(raw), 10, 64)
	if err != nil || interval == 0 {
		return 0, fmt.Errorf("invalid state-sync.snapshot-interval %v", raw)
	}
	return interval, nil
}

func parseCometStateSyncSnapshotChunkPaths(
	nodeHome string,
	height uint64,
	contents []byte,
) ([]cometStateSyncChunkPath, error) {
	cleanHome := path.Clean(nodeHome)
	if !path.IsAbs(cleanHome) || cleanHome != nodeHome {
		return nil, fmt.Errorf("state-sync node home must be a clean absolute path: %q", nodeHome)
	}
	prefix := path.Join(cleanHome, "data", "snapshots", strconv.FormatUint(height, 10)) + "/"
	var chunks []cometStateSyncChunkPath
	seen := make(map[string]struct{})
	for _, line := range strings.Split(string(contents), "\n") {
		candidate := strings.TrimSpace(line)
		if candidate == "" {
			continue
		}
		cleaned := path.Clean(candidate)
		if !strings.HasPrefix(cleaned, prefix) {
			return nil, fmt.Errorf("snapshot chunk path escapes height directory: %q", candidate)
		}
		parts := strings.Split(strings.TrimPrefix(cleaned, prefix), "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("unexpected snapshot chunk path: %q", candidate)
		}
		format, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil || format == 0 {
			return nil, fmt.Errorf("invalid snapshot format path %q", parts[0])
		}
		chunk, err := strconv.ParseUint(parts[1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid snapshot chunk path %q", parts[1])
		}
		relativePath := strings.TrimPrefix(cleaned, cleanHome+"/")
		if _, duplicate := seen[relativePath]; duplicate {
			return nil, fmt.Errorf("duplicate snapshot chunk path %q", relativePath)
		}
		seen[relativePath] = struct{}{}
		chunks = append(chunks, cometStateSyncChunkPath{
			Format:       uint32(format),
			Chunk:        uint32(chunk),
			RelativePath: relativePath,
		})
	}
	if len(chunks) == 0 {
		return nil, fmt.Errorf("snapshot height %d has no chunk files", height)
	}
	sort.Slice(chunks, func(i, j int) bool {
		if chunks[i].Format != chunks[j].Format {
			return chunks[i].Format < chunks[j].Format
		}
		return chunks[i].Chunk < chunks[j].Chunk
	})
	return chunks, nil
}

func selectCommonCometStateSyncChunkZero(
	providers [][]cometStateSyncChunkPath,
) (uint32, []cometStateSyncChunkPath, error) {
	if len(providers) != 2 {
		return 0, nil, fmt.Errorf("corrupted-chunk fault requires exactly two provider inventories, got %d", len(providers))
	}
	byProvider := make([]map[uint32]cometStateSyncChunkPath, len(providers))
	for index, chunks := range providers {
		byProvider[index] = make(map[uint32]cometStateSyncChunkPath)
		for _, chunk := range chunks {
			if chunk.Chunk == 0 {
				byProvider[index][chunk.Format] = chunk
			}
		}
	}
	var selectedFormat uint32
	for format := range byProvider[0] {
		if _, ok := byProvider[1][format]; ok && format > selectedFormat {
			selectedFormat = format
		}
	}
	if selectedFormat == 0 {
		return 0, nil, errors.New("providers do not share snapshot chunk zero at a common format")
	}
	return selectedFormat, []cometStateSyncChunkPath{
		byProvider[0][selectedFormat],
		byProvider[1][selectedFormat],
	}, nil
}

func cometStateSyncSHA256(contents []byte) string {
	hash := sha256.Sum256(contents)
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

func restoreCometStateSyncProviderChunks(
	n *Network,
	evidence *CometStateSyncCorruptedChunkEvidence,
	originals []cometStateSyncOriginalChunk,
) error {
	if len(originals) == 0 {
		return nil
	}
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var restoreErr error
	for index := len(originals) - 1; index >= 0; index-- {
		original := originals[index]
		mutation := &evidence.Mutations[original.EvidenceIndex]
		if err := original.Node.WriteFile(cleanupCtx, original.Contents, mutation.RelativePath); err != nil {
			mutation.RestoreError = err.Error()
			restoreErr = errors.Join(restoreErr, fmt.Errorf("restore provider %s snapshot chunk: %w", original.Node.Name(), err))
			continue
		}
		contents, err := original.Node.ReadFile(cleanupCtx, mutation.RelativePath)
		if err != nil {
			mutation.RestoreError = err.Error()
			restoreErr = errors.Join(restoreErr, fmt.Errorf("verify restored provider %s snapshot chunk: %w", original.Node.Name(), err))
			continue
		}
		mutation.RestoredSHA256 = cometStateSyncSHA256(contents)
		mutation.Restored = mutation.RestoredSHA256 == mutation.OriginalSHA256
		if !mutation.Restored {
			mutation.RestoreError = fmt.Sprintf("restored SHA-256=%s, want %s", mutation.RestoredSHA256, mutation.OriginalSHA256)
			restoreErr = errors.Join(restoreErr, fmt.Errorf("provider %s %s", original.Node.Name(), mutation.RestoreError))
		}
	}
	return restoreErr
}

func waitForCometStateSyncUnavailableProviderFailure(
	ctx context.Context,
	node *cosmos.ChainNode,
	since time.Time,
) ([]byte, CometStateSyncUnavailableProviderLogEvidence) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	var lastLogs []byte
	var lastEvidence CometStateSyncUnavailableProviderLogEvidence
	for {
		logs, err := recoveryContainerLogs(ctx, node, since)
		if err == nil {
			lastLogs = logs
			lastEvidence = parseCometStateSyncUnavailableProviderLogs(logs)
			if lastEvidence.UnexpectedSuccess || lastEvidence.LightClientSetupFailed && lastEvidence.ProviderTransportFailure {
				return lastLogs, lastEvidence
			}
		}
		select {
		case <-ctx.Done():
			return finalCometStateSyncUnavailableProviderLogs(node, since, lastLogs, lastEvidence)
		case <-ticker.C:
		}
	}
}

func finalCometStateSyncUnavailableProviderLogs(
	node *cosmos.ChainNode,
	since time.Time,
	lastLogs []byte,
	lastEvidence CometStateSyncUnavailableProviderLogEvidence,
) ([]byte, CometStateSyncUnavailableProviderLogEvidence) {
	finalCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	logs, err := recoveryContainerLogs(finalCtx, node, since)
	if err == nil {
		return logs, parseCometStateSyncUnavailableProviderLogs(logs)
	}
	return lastLogs, lastEvidence
}

func waitForCometStateSyncCorruptedChunkFailure(
	ctx context.Context,
	node *cosmos.ChainNode,
	since time.Time,
) ([]byte, CometStateSyncCorruptedChunkLogEvidence) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	var lastLogs []byte
	var lastEvidence CometStateSyncCorruptedChunkLogEvidence
	for {
		logs, err := recoveryContainerLogs(ctx, node, since)
		if err == nil {
			lastLogs = logs
			lastEvidence = parseCometStateSyncCorruptedChunkLogs(logs)
			if lastEvidence.UnexpectedSuccess || lastEvidence.ChecksumMismatches >= 2 && lastEvidence.NoValidPeers {
				return lastLogs, lastEvidence
			}
		}
		select {
		case <-ctx.Done():
			finalCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			logs, err := recoveryContainerLogs(finalCtx, node, since)
			cancel()
			if err == nil {
				return logs, parseCometStateSyncCorruptedChunkLogs(logs)
			}
			return lastLogs, lastEvidence
		case <-ticker.C:
		}
	}
}
