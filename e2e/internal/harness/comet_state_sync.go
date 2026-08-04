package harness

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"go.uber.org/zap"
)

var detachedCometStateSyncNodeIndex atomic.Int64

const (
	cometStateSyncProviderPollInterval = 500 * time.Millisecond
	cometStateSyncProviderProbeTimeout = 5 * time.Second
)

type cometStateSyncProviderPollHooks struct {
	readLogs     func(context.Context, *cosmos.ChainNode, time.Time) ([]byte, error)
	probeTimeout time.Duration
	pollInterval time.Duration
}

type cometStateSyncProviderProbeResult struct {
	index    int
	contents []byte
	err      error
}

// CometStateSyncRequest describes an actual online CometBFT state-sync
// bootstrap. RPCSources must be two distinct, healthy nodes from the running
// post-upgrade network. ExpectedImage prevents an old chain-spec image from
// silently being used after the existing nodes have been switched in place.
type CometStateSyncRequest struct {
	Step                  string
	RPCSources            []*cosmos.ChainNode
	ExpectedImage         ImageRef
	QueryCommand          []string
	ProviderSnapshotSince time.Time
	TrustHeightLag        int64
	TrustPeriod           time.Duration
	DiscoveryTime         time.Duration
	ChunkRequestTimeout   time.Duration
	ChunkFetchers         int32
	ProviderWaitTimeout   time.Duration
	CompletionTimeout     time.Duration
}

// CometStateSyncSourceEvidence records the independent RPC/light-client and
// P2P snapshot-provider inputs used by the detached node.
type CometStateSyncSourceEvidence struct {
	Node         string   `json:"node"`
	RPCServer    string   `json:"rpc_server"`
	ChainID      string   `json:"chain_id"`
	LatestHeight int64    `json:"latest_height"`
	Image        ImageRef `json:"image"`
}

// CometStateSyncAgreementEvidence is an exact block ID and application hash
// comparison between the reference and restored nodes, plus the restored
// node's catching_up and peer observations.
type CometStateSyncAgreementEvidence struct {
	TargetHeight int64           `json:"target_height"`
	Reference    QuorumNodeState `json:"reference"`
	Restored     QuorumNodeState `json:"restored"`
	Agreement    QuorumAgreement `json:"agreement"`
}

// CometStateSyncQueryEvidence is canonical JSON returned by the fresh node.
// Height zero means the current application height; a positive value is the
// exact historical height supplied to the Cosmos SDK query boundary.
type CometStateSyncQueryEvidence struct {
	Phase           string          `json:"phase"`
	RequestedHeight int64           `json:"requested_height"`
	Response        json.RawMessage `json:"response"`
}

// CometStateSyncQueryContinuityEvidence proves both current and historical
// application queries before and after restarting the same restored node.
type CometStateSyncQueryContinuityEvidence struct {
	Command          []string                    `json:"command"`
	HistoricalHeight int64                       `json:"historical_height"`
	CurrentBefore    CometStateSyncQueryEvidence `json:"current_before_restart"`
	HistoricalBefore CometStateSyncQueryEvidence `json:"historical_before_restart"`
	CurrentAfter     CometStateSyncQueryEvidence `json:"current_after_restart"`
	HistoricalAfter  CometStateSyncQueryEvidence `json:"historical_after_restart"`
}

// CometStateSyncArtifactPaths names the raw evidence retained outside the
// structured result.
type CometStateSyncArtifactPaths struct {
	Config           string `json:"config"`
	BeforeRestartLog string `json:"before_restart_log"`
	FinalLog         string `json:"final_log"`
	Evidence         string `json:"evidence"`
}

// CometStateSyncEvidence proves a real snapshot discovery/fetch/apply path,
// not application DB import and not genesis-to-head block sync.
type CometStateSyncEvidence struct {
	SchemaVersion           int                                   `json:"schema_version"`
	Mode                    string                                `json:"mode"`
	RecordedAt              time.Time                             `json:"recorded_at"`
	CompletedAt             time.Time                             `json:"completed_at"`
	Step                    string                                `json:"step"`
	Sources                 []CometStateSyncSourceEvidence        `json:"sources"`
	Providers               []CometStateSyncProviderEvidence      `json:"providers"`
	TrustHistory            []BlockEvidence                       `json:"trust_history"`
	Config                  CometStateSyncConfigEvidence          `json:"config"`
	Node                    string                                `json:"node"`
	NodeImage               ImageRef                              `json:"node_image"`
	Volume                  string                                `json:"volume"`
	FreshDataInventory      []string                              `json:"fresh_data_inventory"`
	StateSyncStartedAt      time.Time                             `json:"state_sync_started_at"`
	StateSyncLogs           CometStateSyncLogEvidence             `json:"state_sync_logs"`
	BeforeRestart           CometStateSyncAgreementEvidence       `json:"before_restart"`
	AfterRestart            CometStateSyncAgreementEvidence       `json:"after_restart"`
	ReferenceGenesisBlock   BlockEvidence                         `json:"reference_genesis_block"`
	GenesisBlockUnavailable bool                                  `json:"genesis_block_unavailable"`
	GenesisBlockQueryError  string                                `json:"genesis_block_query_error"`
	RestartSkippedStateSync bool                                  `json:"restart_skipped_state_sync"`
	Queries                 CometStateSyncQueryContinuityEvidence `json:"queries"`
	Artifacts               CometStateSyncArtifactPaths           `json:"artifacts"`
	NodeStopped             bool                                  `json:"node_stopped"`
	Error                   string                                `json:"error,omitempty"`
}

// CometStateSyncBadTrustHashEvidence is an expected-failure record. Error is
// empty when the intentionally mutated hash was rejected within the deadline;
// StartError remains as diagnostic evidence and does not fail the suite.
type CometStateSyncBadTrustHashEvidence struct {
	SchemaVersion      int                               `json:"schema_version"`
	Mode               string                            `json:"mode"`
	RecordedAt         time.Time                         `json:"recorded_at"`
	CompletedAt        time.Time                         `json:"completed_at"`
	Step               string                            `json:"step"`
	Sources            []CometStateSyncSourceEvidence    `json:"sources"`
	Providers          []CometStateSyncProviderEvidence  `json:"providers"`
	TrustHistory       []BlockEvidence                   `json:"trust_history"`
	Config             CometStateSyncConfigEvidence      `json:"config"`
	OriginalTrustHash  string                            `json:"original_trust_hash"`
	MutatedTrustHash   string                            `json:"mutated_trust_hash"`
	Node               string                            `json:"node"`
	NodeImage          ImageRef                          `json:"node_image"`
	Volume             string                            `json:"volume"`
	FreshDataInventory []string                          `json:"fresh_data_inventory"`
	FailureTimeout     string                            `json:"failure_timeout"`
	Elapsed            string                            `json:"elapsed"`
	StartError         string                            `json:"start_error,omitempty"`
	Logs               CometStateSyncBadTrustLogEvidence `json:"logs"`
	ConfigArtifact     string                            `json:"config_artifact"`
	LogArtifact        string                            `json:"log_artifact"`
	EvidenceArtifact   string                            `json:"evidence_artifact"`
	Rejected           bool                              `json:"rejected"`
	NodeStopped        bool                              `json:"node_stopped"`
	Error              string                            `json:"error,omitempty"`
}

type resolvedCometStateSyncInputs struct {
	Sources      []CometStateSyncSourceEvidence
	Providers    []CometStateSyncProviderEvidence
	TrustHistory []BlockEvidence
	Plan         cometStateSyncPlan
}

// RunCometStateSync creates a detached, empty-volume full node using the image
// of the supplied running sources, performs online CometBFT state sync, proves
// truncated block history and canonical application state, restarts the same
// node, and repeats current/historical query and commitment checks.
func (n *Network) RunCometStateSync(
	ctx context.Context,
	request CometStateSyncRequest,
) (evidence CometStateSyncEvidence, retErr error) {
	evidence = CometStateSyncEvidence{
		SchemaVersion: 1,
		Mode:          "actual-cometbft-state-sync",
		RecordedAt:    time.Now().UTC(),
		Step:          request.Step,
		Artifacts: CometStateSyncArtifactPaths{
			Config:           "recovery/state-sync/config.toml",
			BeforeRestartLog: "recovery/state-sync/node-before-restart.log",
			FinalLog:         "recovery/state-sync/node.log",
			Evidence:         "recovery/state-sync/evidence.json",
		},
	}
	if n == nil || n.Chain == nil || n.artifacts == nil {
		return evidence, errors.New("state-sync network and artifact store are required")
	}
	defer func() {
		evidence.CompletedAt = time.Now().UTC()
		evidence.Error = errorString(retErr)
		artifactErr := n.artifacts.writeJSON(evidence.Artifacts.Evidence, evidence)
		retErr = errors.Join(retErr, artifactErr)
		if retErr != nil {
			n.artifacts.recordFailure("comet-state-sync", retErr)
		}
	}()

	normalized, err := n.normalizeCometStateSyncRequest(request, true)
	if err != nil {
		return evidence, err
	}
	evidence.Step = normalized.Step
	inputs, err := n.resolveCometStateSyncInputs(ctx, normalized, "recovery/state-sync/provider")
	evidence.Sources = inputs.Sources
	evidence.Providers = inputs.Providers
	evidence.TrustHistory = inputs.TrustHistory
	if inputs.Plan.TrustHash != "" {
		evidence.Config = inputs.Plan.configEvidence()
	}
	if err != nil {
		return evidence, err
	}

	node, err := n.newDetachedCometStateSyncNode(ctx, normalized.RPCSources[0])
	if err != nil {
		return evidence, err
	}
	evidence.Node = node.Name()
	evidence.NodeImage = imageRefFromNode(node)
	evidence.Volume = node.VolumeName
	defer func() {
		logs, stopped, cleanupErr := n.captureAndStopDetachedCometStateSyncNode(node, evidence.Artifacts.FinalLog)
		evidence.NodeStopped = stopped
		if len(logs) > 0 && evidence.StateSyncLogs.SnapshotHeight == 0 {
			evidence.StateSyncLogs = parseCometStateSyncLogs(logs)
		}
		retErr = errors.Join(retErr, cleanupErr)
	}()

	configContents, inventory, err := n.initializeDetachedCometStateSyncNode(ctx, node, normalized.RPCSources, inputs.Plan)
	if err != nil {
		return evidence, err
	}
	evidence.FreshDataInventory = inventory
	if err := n.artifacts.write(evidence.Artifacts.Config, configContents); err != nil {
		return evidence, fmt.Errorf("record state-sync config.toml: %w", err)
	}
	if err := node.CreateNodeContainer(ctx); err != nil {
		return evidence, fmt.Errorf("create detached state-sync node container: %w", err)
	}

	evidence.StateSyncStartedAt = time.Now().UTC()
	startCtx, cancelStart := context.WithTimeout(ctx, normalized.CompletionTimeout)
	startErr := node.StartContainer(startCtx)
	if startErr != nil {
		cancelStart()
		return evidence, fmt.Errorf("start detached state-sync node: %w", startErr)
	}
	stateSyncLogs, stateSyncLogEvidence, err := waitForCometStateSyncCompletion(startCtx, node, evidence.StateSyncStartedAt)
	cancelStart()
	if len(stateSyncLogs) > 0 {
		if artifactErr := n.artifacts.write(evidence.Artifacts.BeforeRestartLog, stateSyncLogs); artifactErr != nil {
			return evidence, fmt.Errorf("record pre-restart state-sync log: %w", artifactErr)
		}
	}
	evidence.StateSyncLogs = stateSyncLogEvidence
	if err != nil {
		return evidence, err
	}

	evidence.BeforeRestart, err = captureCometStateSyncAgreement(ctx, normalized.RPCSources[0], node)
	if err != nil {
		return evidence, fmt.Errorf("verify restored node before restart: %w", err)
	}
	evidence.ReferenceGenesisBlock, err = n.NodeBlock(ctx, normalized.RPCSources[0], 1)
	if err != nil {
		return evidence, fmt.Errorf("query reference genesis block: %w", err)
	}
	one := int64(1)
	genesisResult, genesisErr := node.Client.Block(ctx, &one)
	evidence.GenesisBlockQueryError = errorString(genesisErr)
	evidence.GenesisBlockUnavailable = genesisErr != nil
	if genesisErr == nil {
		return evidence, fmt.Errorf("state-sync node unexpectedly retained block 1: %+v", genesisResult)
	}

	evidence.Queries.Command = append([]string(nil), normalized.QueryCommand...)
	evidence.Queries.HistoricalHeight = evidence.StateSyncLogs.SnapshotHeight
	evidence.Queries.CurrentBefore, err = n.queryDetachedCometStateSyncNode(
		ctx, normalized.Step+"-current-before-restart", "before-restart", node, normalized.QueryCommand, 0,
	)
	if err != nil {
		return evidence, err
	}
	evidence.Queries.HistoricalBefore, err = n.queryDetachedCometStateSyncNode(
		ctx,
		normalized.Step+"-historical-before-restart",
		"before-restart",
		node,
		normalized.QueryCommand,
		evidence.Queries.HistoricalHeight,
	)
	if err != nil {
		return evidence, err
	}

	if err := node.StopContainer(ctx); err != nil {
		return evidence, fmt.Errorf("stop restored node before restart: %w", err)
	}
	restartStartedAt := time.Now().UTC()
	if err := node.StartContainer(ctx); err != nil {
		return evidence, fmt.Errorf("restart restored state-sync node: %w", err)
	}
	evidence.AfterRestart, err = captureCometStateSyncAgreement(ctx, normalized.RPCSources[0], node)
	if err != nil {
		return evidence, fmt.Errorf("verify restored node after restart: %w", err)
	}
	restartLogs, err := recoveryContainerLogs(ctx, node, restartStartedAt)
	if err != nil {
		return evidence, fmt.Errorf("read restored-node restart logs: %w", err)
	}
	evidence.RestartSkippedStateSync = strings.Contains(
		string(restartLogs),
		"Found local state with non-zero height, skipping state sync",
	)
	if !evidence.RestartSkippedStateSync {
		return evidence, errors.New("restored-node restart logs do not prove that non-zero local state skipped a second state sync")
	}

	evidence.Queries.CurrentAfter, err = n.queryDetachedCometStateSyncNode(
		ctx, normalized.Step+"-current-after-restart", "after-restart", node, normalized.QueryCommand, 0,
	)
	if err != nil {
		return evidence, err
	}
	evidence.Queries.HistoricalAfter, err = n.queryDetachedCometStateSyncNode(
		ctx,
		normalized.Step+"-historical-after-restart",
		"after-restart",
		node,
		normalized.QueryCommand,
		evidence.Queries.HistoricalHeight,
	)
	if err != nil {
		return evidence, err
	}
	if err := validateCometStateSyncQueryContinuity(
		evidence.Queries.CurrentBefore.Response,
		evidence.Queries.CurrentAfter.Response,
		"current",
	); err != nil {
		return evidence, err
	}
	if err := validateCometStateSyncQueryContinuity(
		evidence.Queries.HistoricalBefore.Response,
		evidence.Queries.HistoricalAfter.Response,
		"historical",
	); err != nil {
		return evidence, err
	}
	return evidence, nil
}

// ExpectCometStateSyncBadTrustHash starts a second empty-volume node with one
// bit of the correct trust hash changed. Success means the light client emits
// an explicit hash-mismatch rejection within failureTimeout and never restores
// a snapshot.
func (n *Network) ExpectCometStateSyncBadTrustHash(
	ctx context.Context,
	request CometStateSyncRequest,
	failureTimeout time.Duration,
) (evidence CometStateSyncBadTrustHashEvidence, retErr error) {
	evidence = CometStateSyncBadTrustHashEvidence{
		SchemaVersion:    1,
		Mode:             "actual-cometbft-state-sync-bad-trust-hash",
		RecordedAt:       time.Now().UTC(),
		Step:             request.Step,
		ConfigArtifact:   "recovery/state-sync/bad-trust-hash-config.toml",
		LogArtifact:      "recovery/state-sync/bad-trust-hash.log",
		EvidenceArtifact: "recovery/state-sync/bad-trust-hash.json",
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
			n.artifacts.recordFailure("comet-state-sync-bad-trust-hash", retErr)
		}
	}()

	normalized, err := n.normalizeCometStateSyncRequest(request, false)
	if err != nil {
		return evidence, err
	}
	evidence.Step = normalized.Step
	if failureTimeout == 0 {
		failureTimeout = defaultCometStateSyncBadTrustHashTimeout
	}
	if failureTimeout < 5*time.Second || failureTimeout > maximumCometStateSyncBadTrustHashTimeout {
		return evidence, fmt.Errorf("bad trust-hash failure timeout must be within [5s,%s], got %s", maximumCometStateSyncBadTrustHashTimeout, failureTimeout)
	}
	evidence.FailureTimeout = failureTimeout.String()

	inputs, err := n.resolveCometStateSyncInputs(ctx, normalized, "recovery/state-sync/bad-trust-hash-provider")
	evidence.Sources = inputs.Sources
	evidence.Providers = inputs.Providers
	evidence.TrustHistory = inputs.TrustHistory
	if err != nil {
		return evidence, err
	}
	evidence.OriginalTrustHash = inputs.Plan.TrustHash
	mutatedHash, err := mutateCometStateSyncTrustHash(inputs.Plan.TrustHash)
	if err != nil {
		return evidence, err
	}
	inputs.Plan.TrustHash = mutatedHash
	evidence.MutatedTrustHash = mutatedHash
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
			evidence.Logs = parseCometStateSyncBadTrustLogs(logs)
		}
		retErr = errors.Join(retErr, cleanupErr)
	}()

	configContents, inventory, err := n.initializeDetachedCometStateSyncNode(ctx, node, normalized.RPCSources, inputs.Plan)
	if err != nil {
		return evidence, err
	}
	evidence.FreshDataInventory = inventory
	if err := n.artifacts.write(evidence.ConfigArtifact, configContents); err != nil {
		return evidence, fmt.Errorf("record bad trust-hash config.toml: %w", err)
	}
	if err := node.CreateNodeContainer(ctx); err != nil {
		return evidence, fmt.Errorf("create bad trust-hash state-sync container: %w", err)
	}

	startedAt := time.Now()
	failureCtx, cancelFailure := context.WithTimeout(ctx, failureTimeout)
	startResult := make(chan error, 1)
	go func() {
		startResult <- node.StartContainer(failureCtx)
	}()
	logs, logEvidence := waitForCometStateSyncBadTrustRejection(failureCtx, node, startedAt.UTC())
	cancelFailure()
	startErr := <-startResult
	evidence.StartError = errorString(startErr)
	elapsed := time.Since(startedAt)
	evidence.Elapsed = elapsed.String()
	evidence.Logs = logEvidence
	if len(logs) > 0 {
		if err := n.artifacts.write(evidence.LogArtifact, logs); err != nil {
			return evidence, fmt.Errorf("record bad trust-hash node logs: %w", err)
		}
	}
	if err := validateCometStateSyncBadTrustFailure(evidence.Logs, elapsed, failureTimeout); err != nil {
		return evidence, err
	}
	evidence.Rejected = true
	return evidence, nil
}

func (n *Network) normalizeCometStateSyncRequest(
	request CometStateSyncRequest,
	requireQuery bool,
) (CometStateSyncRequest, error) {
	request.Step = strings.TrimSpace(request.Step)
	if request.Step == "" {
		return request, errors.New("state-sync step is required")
	}
	if len(request.RPCSources) != 2 {
		return request, fmt.Errorf("state sync requires exactly two source nodes, got %d", len(request.RPCSources))
	}
	if request.RPCSources[0] == nil || request.RPCSources[1] == nil {
		return request, errors.New("state-sync source nodes must not be nil")
	}
	if request.RPCSources[0] == request.RPCSources[1] || request.RPCSources[0].Name() == request.RPCSources[1].Name() {
		return request, errors.New("state-sync source nodes must be distinct")
	}
	if strings.TrimSpace(request.ExpectedImage.Repository) == "" || strings.TrimSpace(request.ExpectedImage.Version) == "" {
		return request, errors.New("state-sync expected post-upgrade image is required")
	}
	if requireQuery {
		if err := validateCometStateSyncQueryCommand(request.QueryCommand); err != nil {
			return request, err
		}
	}
	request.QueryCommand = append([]string(nil), request.QueryCommand...)
	if request.ProviderSnapshotSince.IsZero() {
		request.ProviderSnapshotSince = n.artifacts.started
	}
	if request.TrustHeightLag == 0 {
		request.TrustHeightLag = defaultCometStateSyncTrustHeightLag
	}
	if request.TrustHeightLag < 1 {
		return request, errors.New("state-sync trust height lag must be positive")
	}
	if request.ProviderWaitTimeout == 0 {
		request.ProviderWaitTimeout = defaultCometStateSyncProviderWaitTimeout
	}
	if request.ProviderWaitTimeout < time.Second || request.ProviderWaitTimeout > 2*time.Minute {
		return request, fmt.Errorf("state-sync provider wait timeout must be within [1s,2m], got %s", request.ProviderWaitTimeout)
	}
	if request.CompletionTimeout == 0 {
		request.CompletionTimeout = defaultCometStateSyncCompletionTimeout
	}
	if request.CompletionTimeout < 10*time.Second || request.CompletionTimeout > 5*time.Minute {
		return request, fmt.Errorf("state-sync completion timeout must be within [10s,5m], got %s", request.CompletionTimeout)
	}
	return request, nil
}

func (n *Network) resolveCometStateSyncInputs(
	ctx context.Context,
	request CometStateSyncRequest,
	providerArtifactPrefix string,
) (resolvedCometStateSyncInputs, error) {
	var resolved resolvedCometStateSyncInputs
	sources, err := n.validateCometStateSyncSources(ctx, request)
	if err != nil {
		return resolved, err
	}
	providers, err := n.waitForCometStateSyncProviders(
		ctx,
		request.RPCSources,
		request.ProviderSnapshotSince,
		request.ProviderWaitTimeout,
		providerArtifactPrefix,
	)
	resolved.Sources = sources
	resolved.Providers = providers
	if err != nil {
		return resolved, err
	}
	latestSnapshot, ok := latestCommonProviderSnapshotHeight(providers)
	if !ok || latestSnapshot <= 1 {
		return resolved, errors.New("source logs do not identify a common completed state snapshot above height 1")
	}
	for _, source := range request.RPCSources {
		if err := n.WaitForNodeHeight(ctx, source, latestSnapshot+2); err != nil {
			return resolved, fmt.Errorf("wait for state-sync source %s to retain H+2 for snapshot %d: %w", source.Name(), latestSnapshot, err)
		}
	}

	commonLatest := int64(^uint64(0) >> 1)
	chainID := ""
	for index, source := range request.RPCSources {
		status, err := source.Client.Status(ctx)
		if err != nil {
			return resolved, fmt.Errorf("query state-sync source %s status: %w", source.Name(), err)
		}
		if status == nil {
			return resolved, fmt.Errorf("state-sync source %s returned empty status", source.Name())
		}
		if status.SyncInfo.CatchingUp {
			return resolved, fmt.Errorf("state-sync source %s is still catching up", source.Name())
		}
		if chainID == "" {
			chainID = status.NodeInfo.Network
		} else if status.NodeInfo.Network != chainID {
			return resolved, fmt.Errorf("state-sync source chain IDs differ: %s != %s", chainID, status.NodeInfo.Network)
		}
		resolved.Sources[index].ChainID = status.NodeInfo.Network
		resolved.Sources[index].LatestHeight = status.SyncInfo.LatestBlockHeight
		if status.SyncInfo.LatestBlockHeight < commonLatest {
			commonLatest = status.SyncInfo.LatestBlockHeight
		}
	}
	if _, usable := usableProviderSnapshotHeight(providers, commonLatest); !usable {
		return resolved, fmt.Errorf("no completed provider snapshot has verifiable H+2 at common source height %d", commonLatest)
	}
	trustHeight := commonLatest - request.TrustHeightLag
	if trustHeight <= 0 {
		return resolved, fmt.Errorf("common source height %d is too low for trust lag %d", commonLatest, request.TrustHeightLag)
	}
	trustHistory, err := n.RequireSameHistoryAtHeight(ctx, trustHeight, request.RPCSources...)
	resolved.TrustHistory = trustHistory
	if err != nil {
		return resolved, fmt.Errorf("verify independent RPC sources at trust height %d: %w", trustHeight, err)
	}
	plan, err := newCometStateSyncPlan(
		[]string{resolved.Sources[0].RPCServer, resolved.Sources[1].RPCServer},
		trustHeight,
		trustHistory[0].BlockID,
		cometStateSyncPlanOptions{
			TrustPeriod:         request.TrustPeriod,
			DiscoveryTime:       request.DiscoveryTime,
			ChunkRequestTimeout: request.ChunkRequestTimeout,
			ChunkFetchers:       request.ChunkFetchers,
		},
	)
	if err != nil {
		return resolved, err
	}
	resolved.Plan = plan
	return resolved, nil
}

func (n *Network) validateCometStateSyncSources(
	ctx context.Context,
	request CometStateSyncRequest,
) ([]CometStateSyncSourceEvidence, error) {
	wantImage := request.ExpectedImage
	evidence := make([]CometStateSyncSourceEvidence, len(request.RPCSources))
	for index, source := range request.RPCSources {
		if source.Chain != n.Chain {
			return evidence, fmt.Errorf("state-sync source %s does not belong to the supplied Panacea network", source.Name())
		}
		if source.DockerClient == nil || source.DockerClient != n.artifacts.client {
			return evidence, fmt.Errorf("state-sync source %s does not use the network Docker client", source.Name())
		}
		if source.NetworkID == "" || source.NetworkID != n.artifacts.networkID {
			return evidence, fmt.Errorf("state-sync source %s does not use the network Docker network", source.Name())
		}
		if source.ContainerID() == "" {
			return evidence, fmt.Errorf("state-sync source %s has no running container", source.Name())
		}
		if source.Client == nil {
			return evidence, fmt.Errorf("state-sync source %s has no RPC client", source.Name())
		}
		if source.Image.Repository != wantImage.Repository || source.Image.Version != wantImage.Version {
			return evidence, fmt.Errorf(
				"state-sync source %s image %s:%s is not expected post-upgrade image %s:%s",
				source.Name(),
				source.Image.Repository,
				source.Image.Version,
				wantImage.Repository,
				wantImage.Version,
			)
		}
		status, err := source.Client.Status(ctx)
		if err != nil {
			return evidence, fmt.Errorf("query initial state-sync source %s status: %w", source.Name(), err)
		}
		if status == nil || status.SyncInfo.LatestBlockHeight <= 0 || status.SyncInfo.CatchingUp {
			return evidence, fmt.Errorf("state-sync source %s is not healthy and caught up", source.Name())
		}
		evidence[index] = CometStateSyncSourceEvidence{
			Node:         source.Name(),
			RPCServer:    "http://" + source.HostName() + ":26657",
			ChainID:      status.NodeInfo.Network,
			LatestHeight: status.SyncInfo.LatestBlockHeight,
			Image:        imageRefFromNode(source),
		}
	}
	if evidence[0].ChainID != evidence[1].ChainID {
		return evidence, fmt.Errorf("state-sync source chain IDs differ: %s != %s", evidence[0].ChainID, evidence[1].ChainID)
	}
	return evidence, nil
}

func (n *Network) waitForCometStateSyncProviders(
	ctx context.Context,
	sources []*cosmos.ChainNode,
	since time.Time,
	timeout time.Duration,
	artifactPrefix string,
) ([]CometStateSyncProviderEvidence, error) {
	return n.waitForCometStateSyncProvidersWithHooks(
		ctx,
		sources,
		since,
		timeout,
		artifactPrefix,
		cometStateSyncProviderPollHooks{
			readLogs:     recoveryContainerLogs,
			probeTimeout: cometStateSyncProviderProbeTimeout,
			pollInterval: cometStateSyncProviderPollInterval,
		},
	)
}

func (n *Network) waitForCometStateSyncProvidersWithHooks(
	ctx context.Context,
	sources []*cosmos.ChainNode,
	since time.Time,
	timeout time.Duration,
	artifactPrefix string,
	hooks cometStateSyncProviderPollHooks,
) ([]CometStateSyncProviderEvidence, error) {
	if n == nil || n.artifacts == nil {
		return nil, errors.New("state-sync network artifact store is required")
	}
	if len(sources) == 0 {
		return nil, errors.New("at least one state-sync provider source is required")
	}
	if hooks.readLogs == nil {
		return nil, errors.New("state-sync provider log reader is required")
	}
	if hooks.probeTimeout <= 0 {
		return nil, errors.New("state-sync provider probe timeout must be positive")
	}
	if hooks.pollInterval <= 0 {
		return nil, errors.New("state-sync provider poll interval must be positive")
	}

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ticker := time.NewTicker(hooks.pollInterval)
	defer ticker.Stop()

	evidence := make([]CometStateSyncProviderEvidence, len(sources))
	logs := make([][]byte, len(sources))
	satisfied := make([]bool, len(sources))
	lastProbeErr := make([]error, len(sources))
	for index, source := range sources {
		evidence[index] = CometStateSyncProviderEvidence{
			Node:           source.Name(),
			Since:          since,
			RawLogArtifact: fmt.Sprintf("%s-%d.log", artifactPrefix, index),
		}
	}

	writeLogs := func() error {
		var writeErr error
		for index := range logs {
			if err := n.artifacts.write(evidence[index].RawLogArtifact, logs[index]); err != nil {
				writeErr = errors.Join(writeErr, fmt.Errorf("record state-sync provider %d logs: %w", index, err))
			}
		}
		return writeErr
	}
	timeoutError := func() error {
		operationErr := fmt.Errorf("wait for completed Cosmos SDK provider snapshot markers: %w", waitCtx.Err())
		for index, probeErr := range lastProbeErr {
			if probeErr != nil {
				operationErr = errors.Join(
					operationErr,
					fmt.Errorf("provider %s last log probe: %w", sources[index].Name(), probeErr),
				)
			}
		}
		return errors.Join(operationErr, writeLogs())
	}

	for {
		active := make([]int, 0, len(sources))
		for index := range sources {
			if !satisfied[index] {
				active = append(active, index)
			}
		}

		results := make(chan cometStateSyncProviderProbeResult, len(active))
		for _, index := range active {
			index := index
			go func() {
				probeCtx, probeCancel := context.WithTimeout(waitCtx, hooks.probeTimeout)
				defer probeCancel()
				contents, err := hooks.readLogs(probeCtx, sources[index], since)
				results <- cometStateSyncProviderProbeResult{index: index, contents: contents, err: err}
			}()
		}

		for pending := len(active); pending > 0; pending-- {
			select {
			case result := <-results:
				index := result.index
				if result.err != nil {
					lastProbeErr[index] = result.err
					evidence[index].LogError = result.err.Error()
					continue
				}
				logs[index] = result.contents
				evidence[index] = parseCometStateSyncProviderLogs(
					sources[index].Name(),
					since,
					evidence[index].RawLogArtifact,
					result.contents,
				)
				lastProbeErr[index] = nil
				satisfied[index] = len(evidence[index].CompletedSnapshotHeights) > 0
			case <-waitCtx.Done():
				return evidence, timeoutError()
			}
		}

		allFound := true
		for index := range satisfied {
			if !satisfied[index] {
				allFound = false
				break
			}
		}
		if allFound {
			if err := writeLogs(); err != nil {
				return evidence, err
			}
			return evidence, nil
		}
		select {
		case <-waitCtx.Done():
			return evidence, timeoutError()
		case <-ticker.C:
		}
	}
}

func latestProviderSnapshotHeight(providers []CometStateSyncProviderEvidence) (int64, bool) {
	var latest int64
	for _, provider := range providers {
		for _, height := range provider.CompletedSnapshotHeights {
			if height > latest {
				latest = height
			}
		}
	}
	return latest, latest > 0
}

func latestCommonProviderSnapshotHeight(providers []CometStateSyncProviderEvidence) (int64, bool) {
	if len(providers) == 0 {
		return 0, false
	}
	common := make(map[int64]int, len(providers[0].CompletedSnapshotHeights))
	for _, height := range providers[0].CompletedSnapshotHeights {
		if height > 0 {
			common[height] = 1
		}
	}
	for providerIndex := 1; providerIndex < len(providers); providerIndex++ {
		seen := make(map[int64]struct{}, len(providers[providerIndex].CompletedSnapshotHeights))
		for _, height := range providers[providerIndex].CompletedSnapshotHeights {
			seen[height] = struct{}{}
		}
		for height, count := range common {
			if count != providerIndex {
				delete(common, height)
				continue
			}
			if _, ok := seen[height]; !ok {
				delete(common, height)
				continue
			}
			common[height] = providerIndex + 1
		}
	}
	var latest int64
	for height, count := range common {
		if count == len(providers) && height > latest {
			latest = height
		}
	}
	return latest, latest > 0
}

func (n *Network) newDetachedCometStateSyncNode(
	ctx context.Context,
	imageSource *cosmos.ChainNode,
) (*cosmos.ChainNode, error) {
	if imageSource == nil {
		return nil, errors.New("state-sync image source is required")
	}
	index := 10_000 + int(detachedCometStateSyncNodeIndex.Add(1))
	node, err := n.Chain.NewChainNode(
		ctx,
		n.artifacts.runID,
		n.artifacts.client,
		n.artifacts.networkID,
		imageSource.Image,
		false,
		index,
	)
	if err != nil {
		return nil, fmt.Errorf("create detached state-sync node volume: %w", err)
	}
	return node, nil
}

func (n *Network) initializeDetachedCometStateSyncNode(
	ctx context.Context,
	node *cosmos.ChainNode,
	sources []*cosmos.ChainNode,
	plan cometStateSyncPlan,
) ([]byte, []string, error) {
	return n.initializeDetachedCometStateSyncNodeWithOptions(
		ctx,
		node,
		sources,
		plan,
		detachedCometStateSyncNodeOptions{},
	)
}

type detachedCometStateSyncNodeOptions struct {
	DisablePEX bool
}

func (n *Network) initializeDetachedCometStateSyncNodeWithOptions(
	ctx context.Context,
	node *cosmos.ChainNode,
	sources []*cosmos.ChainNode,
	plan cometStateSyncPlan,
	options detachedCometStateSyncNodeOptions,
) ([]byte, []string, error) {
	if node == nil {
		return nil, nil, errors.New("detached state-sync node is required")
	}
	if err := node.InitFullNodeFiles(ctx); err != nil {
		return nil, nil, fmt.Errorf("initialize detached state-sync node home: %w", err)
	}
	peers, err := cometStateSyncPersistentPeers(ctx, sources)
	if err != nil {
		return nil, nil, err
	}
	if err := node.SetPeers(ctx, peers); err != nil {
		return nil, nil, fmt.Errorf("set detached state-sync P2P peers: %w", err)
	}
	genesis, err := sources[0].GenesisFileContent(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("read state-sync source genesis: %w", err)
	}
	if err := node.OverwriteGenesisFile(ctx, genesis); err != nil {
		return nil, nil, fmt.Errorf("write detached state-sync genesis: %w", err)
	}
	override := plan.tomlOverride()
	if options.DisablePEX {
		override["p2p"] = testutil.Toml{"pex": false}
	}
	if err := testutil.ModifyTomlConfigFile(
		ctx,
		zap.NewNop(),
		node.DockerClient,
		node.TestName,
		node.VolumeName,
		"config/config.toml",
		override,
	); err != nil {
		return nil, nil, fmt.Errorf("write detached state-sync config.toml: %w", err)
	}
	configContents, err := node.ReadFile(ctx, "config/config.toml")
	if err != nil {
		return nil, nil, fmt.Errorf("read detached state-sync config.toml: %w", err)
	}
	if err := validateRenderedCometStateSyncConfig(configContents, plan); err != nil {
		return nil, nil, err
	}
	if options.DisablePEX {
		if err := validateRenderedCometStateSyncPEXDisabled(configContents); err != nil {
			return nil, nil, err
		}
	}
	dataDirectory := path.Join(node.HomeDir(), "data")
	stdout, stderr, err := node.Exec(
		ctx,
		[]string{"find", dataDirectory, "-mindepth", "1", "-maxdepth", "4", "-print"},
		n.Chain.Config().Env,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("inspect fresh state-sync data directory: %w: %s", err, boundedString(stderr, txStderrMaxBytes))
	}
	inventory, err := parseStateSyncDataInventory(node.HomeDir(), stdout)
	if err != nil {
		return nil, inventory, err
	}
	return configContents, inventory, nil
}

func cometStateSyncPersistentPeers(ctx context.Context, sources []*cosmos.ChainNode) (string, error) {
	if len(sources) != 2 {
		return "", fmt.Errorf("state-sync persistent peers require exactly two nodes, got %d", len(sources))
	}
	peers := make([]string, len(sources))
	for index, source := range sources {
		if source == nil {
			return "", fmt.Errorf("state-sync peer %d is nil", index)
		}
		nodeID, err := source.NodeID(ctx)
		if err != nil {
			return "", fmt.Errorf("read state-sync peer %s node ID: %w", source.Name(), err)
		}
		if strings.TrimSpace(nodeID) == "" {
			return "", fmt.Errorf("state-sync peer %s returned an empty node ID", source.Name())
		}
		peers[index] = nodeID + "@" + source.HostName() + ":26656"
	}
	return strings.Join(peers, ","), nil
}

func waitForCometStateSyncCompletion(
	ctx context.Context,
	node *cosmos.ChainNode,
	since time.Time,
) ([]byte, CometStateSyncLogEvidence, error) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	var (
		lastLogs     []byte
		lastEvidence CometStateSyncLogEvidence
		lastErr      error
	)
	for {
		logs, err := recoveryContainerLogs(ctx, node, since)
		if err == nil {
			lastLogs = logs
			lastEvidence = parseCometStateSyncLogs(logs)
			lastErr = lastEvidence.validate()
			if lastErr == nil {
				return lastLogs, lastEvidence, nil
			}
		} else {
			lastErr = err
		}
		select {
		case <-ctx.Done():
			return lastLogs, lastEvidence, fmt.Errorf("wait for actual CometBFT state-sync completion markers: last error=%v: %w", lastErr, ctx.Err())
		case <-ticker.C:
		}
	}
}

func captureCometStateSyncAgreement(
	ctx context.Context,
	reference *cosmos.ChainNode,
	restored *cosmos.ChainNode,
) (CometStateSyncAgreementEvidence, error) {
	var evidence CometStateSyncAgreementEvidence
	if reference == nil || restored == nil {
		return evidence, errors.New("state-sync agreement requires reference and restored nodes")
	}
	referenceHeight, err := reference.Height(ctx)
	if err != nil {
		return evidence, fmt.Errorf("read state-sync reference height: %w", err)
	}
	restoredHeight, err := restored.Height(ctx)
	if err != nil {
		return evidence, fmt.Errorf("read restored state-sync node height: %w", err)
	}
	evidence.TargetHeight = referenceHeight
	if restoredHeight < evidence.TargetHeight {
		evidence.TargetHeight = restoredHeight
	}
	if evidence.TargetHeight <= 0 {
		return evidence, fmt.Errorf("state-sync agreement target must be positive: reference=%d restored=%d", referenceHeight, restoredHeight)
	}
	evidence.Reference, err = waitForQuorumNodeState(ctx, evidence.TargetHeight, reference)
	if err != nil {
		return evidence, err
	}
	evidence.Restored, err = waitForQuorumNodeState(ctx, evidence.TargetHeight, restored)
	if err != nil {
		return evidence, err
	}
	evidence.Agreement, err = VerifyCommonCommitment(evidence.TargetHeight, []QuorumCommitment{
		evidence.Reference.QuorumCommitment,
		evidence.Restored.QuorumCommitment,
	})
	if err != nil {
		return evidence, err
	}
	return evidence, nil
}

func (n *Network) queryDetachedCometStateSyncNode(
	ctx context.Context,
	step string,
	phase string,
	node *cosmos.ChainNode,
	command []string,
	height int64,
) (CometStateSyncQueryEvidence, error) {
	evidence := CometStateSyncQueryEvidence{Phase: phase, RequestedHeight: height}
	arguments, record := newDetachedCometStateSyncQueryRecord(step, phase, node.Name(), command, height)
	stdout, stderr, queryErr := node.ExecQuery(ctx, arguments...)
	semantic, semanticErr := NewSemanticJSON(stdout)
	if queryErr == nil && semanticErr == nil {
		evidence.Response = append(json.RawMessage(nil), semantic...)
	}
	record.Response = jsonOrString(stdout)
	record.Stderr = boundedString(stderr, txStderrMaxBytes)
	record.Error = errorString(errors.Join(queryErr, semanticErr))
	recordErr := n.recordQuery(record)
	if queryErr != nil {
		queryErr = fmt.Errorf("state-sync node query %s: %w: %s", step, queryErr, boundedString(stderr, txStderrMaxBytes))
	}
	if semanticErr != nil {
		semanticErr = fmt.Errorf("state-sync node query %s returned invalid JSON: %w", step, semanticErr)
	}
	return evidence, errors.Join(queryErr, semanticErr, recordErr)
}

func newDetachedCometStateSyncQueryRecord(
	step string,
	phase string,
	nodeName string,
	command []string,
	height int64,
) ([]string, queryRecord) {
	arguments := append([]string(nil), command...)
	if height > 0 {
		arguments = append(arguments, "--height", strconv.FormatInt(height, 10))
	}
	return arguments, queryRecord{
		Boundary:         "state-sync-cli",
		Step:             step,
		Height:           height,
		HistoricalHeight: height > 0,
		Request: map[string]any{
			"arguments": arguments,
			"height":    height,
		},
		Metadata: map[string]any{
			"node":  nodeName,
			"phase": phase,
		},
	}
}

func waitForCometStateSyncBadTrustRejection(
	ctx context.Context,
	node *cosmos.ChainNode,
	since time.Time,
) ([]byte, CometStateSyncBadTrustLogEvidence) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	var (
		lastLogs     []byte
		lastEvidence CometStateSyncBadTrustLogEvidence
	)
	for {
		logs, err := recoveryContainerLogs(ctx, node, since)
		if err == nil {
			lastLogs = logs
			lastEvidence = parseCometStateSyncBadTrustLogs(logs)
			if lastEvidence.RejectedTrustHash || lastEvidence.UnexpectedSuccess {
				return lastLogs, lastEvidence
			}
		}
		select {
		case <-ctx.Done():
			// The startup call may consume the deadline. A separate bounded read
			// retains the terminal log line without extending the failure test.
			finalCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			logs, err := recoveryContainerLogs(finalCtx, node, since)
			cancel()
			if err == nil {
				lastLogs = logs
				lastEvidence = parseCometStateSyncBadTrustLogs(logs)
			}
			return lastLogs, lastEvidence
		case <-ticker.C:
		}
	}
}

func (n *Network) captureAndStopDetachedCometStateSyncNode(
	node *cosmos.ChainNode,
	logArtifact string,
) ([]byte, bool, error) {
	if node == nil || node.ContainerID() == "" {
		return nil, false, nil
	}
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	logs, logErr := recoveryContainerLogs(cleanupCtx, node, n.artifacts.started)
	var artifactErr error
	if len(logs) > 0 {
		artifactErr = n.artifacts.write(logArtifact, logs)
	}
	inspect, inspectErr := node.DockerClient.ContainerInspect(cleanupCtx, node.ContainerID())
	stopped := false
	var stopErr error
	if inspectErr == nil {
		if inspect.State == nil || !inspect.State.Running {
			stopped = true
		} else {
			stopErr = node.StopContainer(cleanupCtx)
			stopped = stopErr == nil
		}
	}
	return logs, stopped, errors.Join(logErr, artifactErr, inspectErr, stopErr)
}

func imageRefFromNode(node *cosmos.ChainNode) ImageRef {
	if node == nil {
		return ImageRef{}
	}
	return ImageRef{Repository: node.Image.Repository, Version: node.Image.Version}
}
