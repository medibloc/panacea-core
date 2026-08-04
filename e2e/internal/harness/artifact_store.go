package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

const (
	artifactCollectionBaseTimeout    = 30 * time.Second
	artifactCollectionPerNodeTimeout = 15 * time.Second
	artifactCollectionMaxTimeout     = 2 * time.Minute
	artifactLogTail                  = "5000"
	artifactLogMaxBytes              = 16 << 20
)

type artifactStore struct {
	dir      string
	testName string
	runID    string
	config   Config
	started  time.Time

	mu                        sync.Mutex
	chain                     *cosmos.CosmosChain
	client                    *dockerclient.Client
	networkID                 string
	state                     string
	buildError                string
	failures                  []artifactFailure
	failed                    bool
	cleanup                   artifactCleanup
	nodeHeights               map[string]int64
	intentionallyStoppedNodes map[string]struct{}
	cleanupRegistered         bool
}

type artifactFailure struct {
	RecordedAt time.Time `json:"recorded_at"`
	Stage      string    `json:"stage"`
	Error      string    `json:"error"`
}

type artifactCleanup struct {
	State                   string     `json:"state"`
	Result                  string     `json:"result"`
	RecordedAt              *time.Time `json:"recorded_at,omitempty"`
	InterchainCloseError    string     `json:"interchain_close_error,omitempty"`
	ArtifactCollectionError string     `json:"artifact_collection_error,omitempty"`
	DockerCleanupError      string     `json:"docker_cleanup_error,omitempty"`
}

type artifactStatus struct {
	OK             bool      `json:"ok"`
	RecordedAt     time.Time `json:"recorded_at"`
	LastHeight     int64     `json:"last_height,omitempty"`
	Peers          *int      `json:"peers,omitempty"`
	Error          string    `json:"error,omitempty"`
	PeerError      string    `json:"peer_error,omitempty"`
	CometBFTStatus any       `json:"cometbft_status,omitempty"`
}

func newArtifactStore(testName, runID string, cfg Config) (*artifactStore, error) {
	// Validate before resolving or creating any filesystem path. In particular,
	// filepath.Join must never receive an unchecked caller-provided run ID.
	if !runIDPattern.MatchString(runID) {
		return nil, fmt.Errorf("run ID %q must match %s", runID, runIDPattern)
	}

	root := cfg.ArtifactRoot
	if root == "" {
		root = os.Getenv("PANACEA_E2E_ROOT")
	}
	allowedRepoRoot, err := defaultArtifactRoot()
	if err != nil {
		return nil, err
	}
	lexicalRepoRoot := filepath.Dir(filepath.Dir(allowedRepoRoot))
	resolvedRepoRoot, err := resolveThroughExistingAncestor(lexicalRepoRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve repository root: %w", err)
	}
	expectedRepoArtifactRoot := filepath.Join(resolvedRepoRoot, ".local", "e2e")
	if root == "" {
		root = allowedRepoRoot
	}
	root, err = resolveThroughExistingAncestor(root)
	if err != nil {
		return nil, fmt.Errorf("resolve artifact root: %w", err)
	}
	allowedRepoRoot, err = resolveThroughExistingAncestor(allowedRepoRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve repository artifact root: %w", err)
	}
	if !pathWithin(expectedRepoArtifactRoot, allowedRepoRoot) {
		return nil, fmt.Errorf("repository artifact root resolves outside %s: %s", expectedRepoArtifactRoot, allowedRepoRoot)
	}
	// Do not use os.TempDir here: TMPDIR is caller-controlled and TMPDIR=/
	// would turn the containment check into an allow-all. /tmp is the explicit
	// external root permitted by the E2E isolation contract; symlinks (such as
	// macOS /tmp -> /private/tmp) are resolved before comparison.
	tempRoot, err := resolveThroughExistingAncestor("/tmp")
	if err != nil {
		return nil, fmt.Errorf("resolve temporary artifact root: %w", err)
	}
	if !pathWithin(allowedRepoRoot, root) && !pathWithin(tempRoot, root) {
		return nil, fmt.Errorf("artifact root %s must be under %s or %s", root, allowedRepoRoot, tempRoot)
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("create artifact root: %w", err)
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("resolve artifact root symlinks: %w", err)
	}
	if !pathWithin(allowedRepoRoot, resolvedRoot) && !pathWithin(tempRoot, resolvedRoot) {
		return nil, fmt.Errorf("artifact root resolves outside allowed roots: %s", resolvedRoot)
	}

	dir := filepath.Join(root, runID)
	if err := os.Mkdir(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create run artifact directory: %w", err)
	}
	for _, subdir := range []string{"nodes", "tx", "queries", "metrics"} {
		if err := os.Mkdir(filepath.Join(dir, subdir), 0o700); err != nil {
			return nil, fmt.Errorf("create artifact directory %s: %w", subdir, err)
		}
	}
	store := &artifactStore{
		dir:                       dir,
		testName:                  testName,
		runID:                     runID,
		config:                    cfg,
		started:                   time.Now().UTC(),
		state:                     "initializing",
		cleanup:                   artifactCleanup{State: "not-registered", Result: "pending"},
		nodeHeights:               make(map[string]int64),
		intentionallyStoppedNodes: make(map[string]struct{}),
	}
	if err := store.writeManifest(false); err != nil {
		return nil, err
	}
	return store, nil
}

func defaultArtifactRoot() (string, error) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("locate harness source")
	}
	repositoryRoot, err := filepath.Abs(filepath.Join(filepath.Dir(source), "..", "..", ".."))
	if err != nil {
		return "", err
	}
	return filepath.Join(repositoryRoot, ".local", "e2e"), nil
}

func pathWithin(base, target string) bool {
	base, err := filepath.Abs(base)
	if err != nil {
		return false
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return false
	}
	relative, err := filepath.Rel(base, target)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

// resolveThroughExistingAncestor resolves every existing symlink before any
// missing descendant is created. Resolving only after MkdirAll would allow an
// existing symlink below an allowed-looking path to create directories outside
// the approved roots before the escape was detected.
func resolveThroughExistingAncestor(candidate string) (string, error) {
	abs, err := filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	abs = filepath.Clean(abs)

	ancestor := abs
	var suffix []string
	for {
		_, statErr := os.Lstat(ancestor)
		if statErr == nil {
			break
		}
		if !os.IsNotExist(statErr) {
			return "", statErr
		}
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			return "", fmt.Errorf("no existing ancestor for %s", abs)
		}
		suffix = append([]string{filepath.Base(ancestor)}, suffix...)
		ancestor = parent
	}

	resolved, err := filepath.EvalSymlinks(ancestor)
	if err != nil {
		return "", err
	}
	parts := append([]string{resolved}, suffix...)
	return filepath.Clean(filepath.Join(parts...)), nil
}

func (s *artifactStore) attach(chain *cosmos.CosmosChain, client *dockerclient.Client, networkID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chain = chain
	s.client = client
	s.networkID = networkID
}

func (s *artifactStore) setBuildError(err error) {
	if err == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failed = true
	s.state = "build-failed"
	s.buildError = err.Error()
}

func (s *artifactStore) markRunning() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = "running"
}

func (s *artifactStore) markCleanupRegistered() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupRegistered = true
	s.cleanup.State = "registered"
}

func (s *artifactStore) cleanupIsRegistered() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cleanupRegistered
}

func (s *artifactStore) recordFailure(stage string, err error) {
	if err == nil {
		return
	}
	failure := artifactFailure{RecordedAt: time.Now().UTC(), Stage: stage, Error: err.Error()}
	s.mu.Lock()
	s.failures = append(s.failures, failure)
	s.mu.Unlock()
	_ = s.appendJSONLine("queries/failures.jsonl", failure)
}

func (s *artifactStore) recordTestPanic(recovered any) {
	failure := artifactFailure{
		RecordedAt: time.Now().UTC(),
		Stage:      "test-panic",
		Error:      fmt.Sprint(recovered),
	}
	s.mu.Lock()
	s.failures = append(s.failures, failure)
	s.failed = true
	s.state = "failed"
	s.mu.Unlock()
	_ = s.appendJSONLine("queries/failures.jsonl", failure)
}

func (s *artifactStore) markIntentionallyStopped(nodeNames ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, nodeName := range nodeNames {
		if strings.TrimSpace(nodeName) != "" {
			s.intentionallyStoppedNodes[nodeName] = struct{}{}
		}
	}
}

func (s *artifactStore) nodeIntentionallyStopped(nodeName string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.intentionallyStoppedNodes[nodeName]
	return ok
}

func validateIntentionallyStoppedContainer(expected, inspected, running bool) (bool, error) {
	if !expected {
		return false, nil
	}
	if !inspected {
		return true, errors.New("intentionally stopped node has no inspectable container")
	}
	if running {
		return true, errors.New("node expected to remain stopped is running")
	}
	return true, nil
}

func (s *artifactStore) collect(failed bool) error {
	s.mu.Lock()
	s.failed = s.failed || failed
	effectiveFailure := s.failed
	if effectiveFailure {
		s.state = "failed"
	} else {
		s.state = "passed"
	}
	chain := s.chain
	client := s.client
	s.mu.Unlock()

	nodeCount := 0
	if chain != nil {
		nodeCount = len(chain.Nodes())
	}
	ctx, cancel := context.WithTimeout(context.Background(), artifactCollectionDeadline(nodeCount))
	defer cancel()

	var collectionErrors []error
	if err := s.writeManifest(effectiveFailure); err != nil {
		collectionErrors = append(collectionErrors, err)
	}
	if chain != nil {
		if err := s.collectChain(ctx, chain, client); err != nil {
			collectionErrors = append(collectionErrors, err)
		}
	}
	for _, err := range collectionErrors {
		s.recordFailure("artifact-collection", err)
	}
	if len(collectionErrors) > 0 {
		s.mu.Lock()
		s.failed = true
		s.state = "artifact-failed"
		s.mu.Unlock()
	}
	if effectiveFailure || len(collectionErrors) > 0 {
		if err := s.writeFailureSummary(ctx, chain, collectionErrors); err != nil {
			collectionErrors = append(collectionErrors, err)
		}
	}
	if err := s.writeManifest(effectiveFailure || len(collectionErrors) > 0); err != nil {
		collectionErrors = append(collectionErrors, err)
	}
	return errors.Join(collectionErrors...)
}

func artifactCollectionDeadline(nodeCount int) time.Duration {
	if nodeCount < 0 {
		nodeCount = 0
	}
	nodesUntilCap := int((artifactCollectionMaxTimeout - artifactCollectionBaseTimeout + artifactCollectionPerNodeTimeout - 1) / artifactCollectionPerNodeTimeout)
	if nodeCount >= nodesUntilCap {
		return artifactCollectionMaxTimeout
	}
	deadline := artifactCollectionBaseTimeout + time.Duration(nodeCount)*artifactCollectionPerNodeTimeout
	return deadline
}

func (s *artifactStore) collectChain(ctx context.Context, chain *cosmos.CosmosChain, client *dockerclient.Client) error {
	var collectionErrors []error
	if len(chain.Validators) > 0 {
		genesis, err := chain.Validators[0].ReadFile(ctx, "config/genesis.json")
		if err != nil {
			collectionErrors = append(collectionErrors, fmt.Errorf("read genesis: %w", err))
		} else if err := s.write("genesis.json", genesis); err != nil {
			collectionErrors = append(collectionErrors, err)
		}
		stdout, stderr, err := chain.Validators[0].ExecBin(ctx, "version", "--long")
		if err != nil {
			collectionErrors = append(collectionErrors, fmt.Errorf("read binary version: %w: %s", err, strings.TrimSpace(string(stderr))))
		} else if err := s.write("versions.txt", stdout); err != nil {
			collectionErrors = append(collectionErrors, err)
		}
	}
	for _, node := range chain.Nodes() {
		if err := s.collectNode(ctx, node, client); err != nil {
			collectionErrors = append(collectionErrors, fmt.Errorf("node %s: %w", node.Name(), err))
		}
	}
	return errors.Join(collectionErrors...)
}

func (s *artifactStore) collectNode(ctx context.Context, node *cosmos.ChainNode, client *dockerclient.Client) error {
	var collectionErrors []error
	base := filepath.Join("nodes", node.Name())
	for _, name := range []string{"app.toml", "config.toml", "client.toml"} {
		contents, err := node.ReadFile(ctx, filepath.Join("config", name))
		if err != nil {
			collectionErrors = append(collectionErrors, fmt.Errorf("read %s: %w", name, err))
			continue
		}
		if err := s.write(filepath.Join(base, "config", name), contents); err != nil {
			collectionErrors = append(collectionErrors, err)
		}
	}
	containerInspected := false
	containerRunning := false
	if client != nil && node.ContainerID() != "" {
		inspect, err := client.ContainerInspect(ctx, node.ContainerID())
		if err != nil {
			collectionErrors = append(collectionErrors, fmt.Errorf("inspect container: %w", err))
		} else {
			containerInspected = true
			containerRunning = inspect.State != nil && inspect.State.Running
			var ports any
			if inspect.NetworkSettings != nil {
				ports = inspect.NetworkSettings.Ports
			}
			safeInspect := struct {
				ID      string `json:"id"`
				Name    string `json:"name"`
				Image   string `json:"image"`
				Created string `json:"created"`
				State   any    `json:"state"`
				Ports   any    `json:"ports"`
			}{
				ID: inspect.ID, Name: inspect.Name, Image: inspect.Image,
				Created: inspect.Created, State: inspect.State, Ports: ports,
			}
			if err := s.writeJSON(filepath.Join(base, "container-state.json"), safeInspect); err != nil {
				collectionErrors = append(collectionErrors, err)
			}
		}
		if err := s.collectLogs(ctx, client, node.ContainerID(), filepath.Join(base, "logs", "container.log")); err != nil {
			collectionErrors = append(collectionErrors, err)
		}
	}

	statusArtifact := artifactStatus{RecordedAt: time.Now().UTC()}
	handledStopped, stoppedErr := validateIntentionallyStoppedContainer(
		s.nodeIntentionallyStopped(node.Name()),
		containerInspected,
		containerRunning,
	)
	if handledStopped {
		if stoppedErr != nil {
			statusArtifact.Error = stoppedErr.Error()
			collectionErrors = append(collectionErrors, fmt.Errorf("status: %w", stoppedErr))
		} else {
			statusArtifact.Error = "node intentionally stopped after terminal export"
		}
	} else {
		var statusErr error
		if node.Client == nil {
			statusErr = errors.New("CometBFT RPC client is not initialized")
		} else {
			status, err := node.Client.Status(ctx)
			if err != nil {
				statusErr = err
			} else if status == nil {
				statusErr = errors.New("CometBFT RPC returned an empty status")
			} else {
				statusArtifact.OK = true
				statusArtifact.LastHeight = status.SyncInfo.LatestBlockHeight
				statusArtifact.CometBFTStatus = status
				s.recordNodeHeight(node.Name(), statusArtifact.LastHeight)
				netInfo, netInfoErr := node.Client.NetInfo(ctx)
				switch {
				case netInfoErr != nil:
					statusArtifact.PeerError = netInfoErr.Error()
					collectionErrors = append(collectionErrors, fmt.Errorf("peer count: %w", netInfoErr))
				case netInfo == nil:
					statusArtifact.PeerError = "CometBFT RPC returned empty net info"
					collectionErrors = append(collectionErrors, errors.New("peer count: CometBFT RPC returned empty net info"))
				default:
					peers := netInfo.NPeers
					statusArtifact.Peers = &peers
				}
			}
		}
		if statusErr != nil {
			statusArtifact.Error = statusErr.Error()
			collectionErrors = append(collectionErrors, fmt.Errorf("status: %w", statusErr))
		}
	}
	if err := s.writeJSON(filepath.Join(base, "status.json"), statusArtifact); err != nil {
		collectionErrors = append(collectionErrors, err)
	}
	return errors.Join(collectionErrors...)
}

func (s *artifactStore) recordNodeHeight(nodeName string, height int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if height > s.nodeHeights[nodeName] {
		s.nodeHeights[nodeName] = height
	}
}

func (s *artifactStore) collectLogs(ctx context.Context, client *dockerclient.Client, containerID, destination string) error {
	reader, err := client.ContainerLogs(ctx, containerID, dockertypes.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Tail:       artifactLogTail,
	})
	if err != nil {
		return fmt.Errorf("read container logs: %w", err)
	}
	defer reader.Close()
	const (
		stdoutHeader = "[stdout]\n"
		stderrHeader = "\n[stderr]\n"
		truncatedLog = "\n[truncated: combined Docker logs exceeded 16 MiB]\n"
	)
	payloadBudget := artifactLogMaxBytes - len(stdoutHeader) - len(stderrHeader) - len(truncatedLog)
	budget := &artifactLogBudget{remaining: payloadBudget}
	stdout := artifactLogBuffer{budget: budget}
	stderr := artifactLogBuffer{budget: budget}
	if _, err := stdcopy.StdCopy(&stdout, &stderr, reader); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("demultiplex container logs: %w", err)
	}
	combined := make([]byte, 0, artifactLogMaxBytes)
	combined = append(combined, stdoutHeader...)
	combined = append(combined, stdout.Bytes()...)
	combined = append(combined, stderrHeader...)
	combined = append(combined, stderr.Bytes()...)
	if budget.truncated {
		combined = append(combined, truncatedLog...)
	}
	if len(combined) > artifactLogMaxBytes {
		return fmt.Errorf("internal error: bounded Docker log is %d bytes", len(combined))
	}
	return s.write(destination, combined)
}

type artifactLogBudget struct {
	remaining int
	truncated bool
}

type artifactLogBuffer struct {
	bytes.Buffer
	budget *artifactLogBudget
}

func (w *artifactLogBuffer) Write(contents []byte) (int, error) {
	originalLength := len(contents)
	writable := originalLength
	if writable > w.budget.remaining {
		writable = w.budget.remaining
		w.budget.truncated = true
	}
	if writable > 0 {
		_, _ = w.Buffer.Write(contents[:writable])
		w.budget.remaining -= writable
	}
	// Report the full input as consumed so stdcopy can keep draining the
	// bounded 5,000-line Docker response without allocating more memory.
	return originalLength, nil
}

func (s *artifactStore) writeManifest(failed bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failed = s.failed || failed
	failures := append([]artifactFailure(nil), s.failures...)
	heights := make(map[string]int64, len(s.nodeHeights))
	for nodeName, height := range s.nodeHeights {
		heights[nodeName] = height
	}
	manifest := struct {
		RunID         string            `json:"run_id"`
		TestName      string            `json:"test_name"`
		State         string            `json:"state"`
		Failed        bool              `json:"failed"`
		StartedAt     time.Time         `json:"started_at"`
		CollectedAt   time.Time         `json:"collected_at"`
		Image         string            `json:"image"`
		ChainID       string            `json:"chain_id,omitempty"`
		NumValidators int               `json:"num_validators"`
		NumFullNodes  int               `json:"num_full_nodes"`
		NetworkID     string            `json:"docker_network_id,omitempty"`
		BuildError    string            `json:"build_error,omitempty"`
		Failures      []artifactFailure `json:"failures,omitempty"`
		Cleanup       artifactCleanup   `json:"cleanup"`
		NodeHeights   map[string]int64  `json:"node_heights,omitempty"`
	}{
		RunID: s.runID, TestName: s.testName, State: s.state, Failed: s.failed,
		StartedAt: s.started, CollectedAt: time.Now().UTC(),
		Image:         s.config.Image.Repository + ":" + s.config.Image.Version,
		NumValidators: s.config.NumValidators, NumFullNodes: s.config.NumFullNodes,
		NetworkID: s.networkID, BuildError: s.buildError, Failures: failures,
		Cleanup: s.cleanup, NodeHeights: heights,
	}
	if s.chain != nil {
		manifest.ChainID = s.chain.Config().ChainID
	}
	return s.writeJSON("manifest.json", manifest)
}

func (s *artifactStore) writeFailureSummary(ctx context.Context, chain *cosmos.CosmosChain, collectionErrors []error) error {
	s.mu.Lock()
	lastHeight := maxRecordedHeight(s.nodeHeights)
	s.mu.Unlock()
	// A chain can exist even when Build failed before Interchaintest initialized
	// its RPC clients. CosmosChain.Height dereferences the validator client, so
	// keep the already-recorded height instead of panicking during diagnostics.
	if chain != nil && len(chain.Validators) > 0 && chain.Validators[0] != nil && chain.Validators[0].Client != nil {
		if height, err := chain.Height(ctx); err == nil {
			if height > lastHeight {
				lastHeight = height
			}
		}
	}
	s.mu.Lock()
	buildError := s.buildError
	failures := append([]artifactFailure(nil), s.failures...)
	cleanup := s.cleanup
	s.mu.Unlock()
	summary := fmt.Sprintf("run_id: %s\nlast_height: %d\nbuild_error: %s\n", s.runID, lastHeight, buildError)
	for _, failure := range failures {
		summary += fmt.Sprintf("failure[%s]: %s\n", failure.Stage, failure.Error)
	}
	if joined := errors.Join(collectionErrors...); joined != nil {
		summary += "collection_error: " + joined.Error() + "\n"
	}
	if cleanup.Result == "failed" {
		summary += "cleanup_result: failed\n"
		if cleanup.InterchainCloseError != "" {
			summary += "interchain_close_error: " + cleanup.InterchainCloseError + "\n"
		}
		if cleanup.ArtifactCollectionError != "" {
			summary += "artifact_collection_error: " + cleanup.ArtifactCollectionError + "\n"
		}
		if cleanup.DockerCleanupError != "" {
			summary += "docker_cleanup_error: " + cleanup.DockerCleanupError + "\n"
		}
	}
	return s.write("failure-summary.txt", []byte(summary))
}

func maxRecordedHeight(heights map[string]int64) int64 {
	var maximum int64
	for _, height := range heights {
		if height > maximum {
			maximum = height
		}
	}
	return maximum
}

func (s *artifactStore) recordCleanup(closeErr, collectErr, cleanupErr error) error {
	joined := errors.Join(closeErr, collectErr, cleanupErr)
	failed := joined != nil
	if closeErr != nil {
		s.recordFailure("interchain-close", closeErr)
	}
	if collectErr != nil {
		s.recordFailure("artifact-collection", collectErr)
	}
	if cleanupErr != nil {
		s.recordFailure("docker-cleanup", cleanupErr)
	}
	now := time.Now().UTC()
	s.mu.Lock()
	s.failed = s.failed || failed
	s.cleanup = artifactCleanup{
		State:                   "completed",
		Result:                  "succeeded",
		RecordedAt:              &now,
		InterchainCloseError:    errorString(closeErr),
		ArtifactCollectionError: errorString(collectErr),
		DockerCleanupError:      errorString(cleanupErr),
	}
	if failed {
		s.state = "cleanup-failed"
		s.cleanup.Result = "failed"
	} else if s.failed {
		s.state = "failed-cleaned"
	} else {
		s.state = "cleaned"
	}
	cleanup := s.cleanup
	s.mu.Unlock()
	var finalizationErrors []error
	if err := s.writeJSON("cleanup.json", cleanup); err != nil {
		finalizationErrors = append(finalizationErrors, err)
	}
	if failed {
		if err := s.writeFailureSummary(context.Background(), nil, []error{joined}); err != nil {
			finalizationErrors = append(finalizationErrors, err)
		}
	}
	if err := s.writeManifest(failed); err != nil {
		finalizationErrors = append(finalizationErrors, err)
	}
	return errors.Join(finalizationErrors...)
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func (s *artifactStore) appendJSONLine(relativePath string, value any) error {
	contents, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	path, err := s.safePath(relativePath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(append(contents, '\n'))
	return err
}

func (s *artifactStore) writeJSON(relativePath string, value any) error {
	contents, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return s.write(relativePath, append(contents, '\n'))
}

func (s *artifactStore) write(relativePath string, contents []byte) error {
	path, err := s.safePath(relativePath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, contents, 0o600)
}

func (s *artifactStore) safePath(relativePath string) (string, error) {
	if filepath.IsAbs(relativePath) {
		return "", errors.New("artifact path must be relative")
	}
	path := filepath.Join(s.dir, filepath.Clean(relativePath))
	if !pathWithin(s.dir, path) {
		return "", fmt.Errorf("artifact path escapes run directory: %s", relativePath)
	}
	return path, nil
}
