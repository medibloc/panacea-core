package harness

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

const panicCleanupTimingProbeEnv = "PANACEA_E2E_PANIC_CLEANUP_TIMING_PROBE"

func TestTestingCleanupDoesNotObserveActivePanicAsFailure(t *testing.T) {
	if os.Getenv(panicCleanupTimingProbeEnv) == "1" {
		t.Cleanup(func() {
			fmt.Printf("panic-cleanup-t-failed=%t\n", t.Failed())
		})
		panic("intentional panic cleanup timing probe")
	}

	command := exec.Command(os.Args[0], "-test.run=^TestTestingCleanupDoesNotObserveActivePanicAsFailure$", "-test.v")
	command.Env = append(os.Environ(), panicCleanupTimingProbeEnv+"=1")
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatalf("panic timing child unexpectedly passed:\n%s", output)
	}
	if !strings.Contains(string(output), "panic-cleanup-t-failed=false") {
		t.Fatalf("testing cleanup failure timing changed; child output:\n%s", output)
	}
}

func TestNetworkRecordTestPanicPersistsFailureAndRepanics(t *testing.T) {
	store, err := newArtifactStore(
		"test",
		"run-test-panic",
		artifactTestConfig(filepath.Join(trustedArtifactTempDir(t), "artifacts")),
	)
	if err != nil {
		t.Fatalf("newArtifactStore: %v", err)
	}
	network := &Network{artifacts: store}

	recovered := invokeNetworkTestPanicRecorder(network, "legacy dec reflection")
	if recovered != "legacy dec reflection" {
		t.Fatalf("recovered panic = %#v, want original panic value", recovered)
	}
	if err := store.collect(false); err != nil {
		t.Fatalf("collect: %v", err)
	}
	if err := store.recordCleanup(nil, nil, nil); err != nil {
		t.Fatalf("recordCleanup: %v", err)
	}

	manifestBytes, err := os.ReadFile(filepath.Join(store.dir, "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var manifest struct {
		State    string            `json:"state"`
		Failed   bool              `json:"failed"`
		Failures []artifactFailure `json:"failures"`
	}
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if !manifest.Failed || manifest.State != "failed-cleaned" {
		t.Fatalf("panic manifest = failed:%t state:%q", manifest.Failed, manifest.State)
	}
	if len(manifest.Failures) != 1 || manifest.Failures[0].Stage != "test-panic" ||
		!strings.Contains(manifest.Failures[0].Error, "legacy dec reflection") {
		t.Fatalf("panic failures = %#v", manifest.Failures)
	}
	summary, err := os.ReadFile(filepath.Join(store.dir, "failure-summary.txt"))
	if err != nil {
		t.Fatalf("read failure summary: %v", err)
	}
	if !strings.Contains(string(summary), "failure[test-panic]: legacy dec reflection") {
		t.Fatalf("failure summary does not identify the test panic:\n%s", summary)
	}
}

func invokeNetworkTestPanicRecorder(network *Network, panicValue any) (recovered any) {
	defer func() {
		recovered = recover()
	}()
	func() {
		defer network.RecordTestPanic()
		panic(panicValue)
	}()
	return nil
}

func TestNewArtifactStoreRejectsInvalidRunIDBeforeMutation(t *testing.T) {
	tests := map[string]string{
		"path traversal": "../escaped",
		"too long":       strings.Repeat("a", 43),
	}
	for name, runID := range tests {
		t.Run(name, func(t *testing.T) {
			sandbox := t.TempDir()
			root := filepath.Join(sandbox, "artifacts")
			if _, err := newArtifactStore("test", runID, artifactTestConfig(root)); err == nil {
				t.Fatalf("newArtifactStore(%q) unexpectedly succeeded", runID)
			}
			if _, err := os.Stat(root); !os.IsNotExist(err) {
				t.Fatalf("invalid run ID mutated artifact root %s: %v", root, err)
			}
			if _, err := os.Stat(filepath.Join(sandbox, "escaped")); !os.IsNotExist(err) {
				t.Fatalf("invalid run ID escaped artifact root: %v", err)
			}
		})
	}
}

func TestNewArtifactStoreAllowsNonExistingDescendantOfAllowedRoot(t *testing.T) {
	root := filepath.Join(trustedArtifactTempDir(t), "one", "two", "three")
	store, err := newArtifactStore("test", "run-safe", artifactTestConfig(root))
	if err != nil {
		t.Fatalf("newArtifactStore: %v", err)
	}
	if info, err := os.Stat(store.dir); err != nil {
		t.Fatalf("stat run directory: %v", err)
	} else if !info.IsDir() {
		t.Fatalf("run path %s is not a directory", store.dir)
	}
}

func TestNewArtifactStoreDoesNotTrustHostileTMPDIR(t *testing.T) {
	t.Setenv("TMPDIR", string(filepath.Separator))
	allowedRepoRoot, err := defaultArtifactRoot()
	if err != nil {
		t.Fatalf("defaultArtifactRoot: %v", err)
	}
	repositoryRoot := filepath.Dir(filepath.Dir(allowedRepoRoot))
	outsideAllowedRoots := filepath.Join(repositoryRoot, "docs")

	_, err = newArtifactStore("test", "adr", artifactTestConfig(outsideAllowedRoots))
	if err == nil {
		t.Fatal("newArtifactStore unexpectedly trusted TMPDIR=/")
	}
	if !strings.Contains(err.Error(), "must be under") {
		t.Fatalf("newArtifactStore returned non-safety error: %v", err)
	}
}

func TestNewArtifactStoreRejectsSymlinkEscapeBeforeMutation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation is not reliably available on Windows test hosts")
	}
	allowedRepoRoot, err := defaultArtifactRoot()
	if err != nil {
		t.Fatalf("defaultArtifactRoot: %v", err)
	}
	repositoryRoot := filepath.Dir(filepath.Dir(allowedRepoRoot))
	targetName := fmt.Sprintf(".artifact-store-symlink-escape-%d-%d", os.Getpid(), time.Now().UnixNano())
	targetRoot := filepath.Join(repositoryRoot, targetName)
	if _, err := os.Stat(targetRoot); !os.IsNotExist(err) {
		t.Fatalf("test target already exists: %s", targetRoot)
	}
	t.Cleanup(func() { _ = os.RemoveAll(targetRoot) })

	sandbox := t.TempDir()
	link := filepath.Join(sandbox, "outside")
	if err := os.Symlink(repositoryRoot, link); err != nil {
		t.Fatalf("create symlink: %v", err)
	}
	configuredRoot := filepath.Join(link, targetName, "nested")
	if _, err := newArtifactStore("test", "run-safe", artifactTestConfig(configuredRoot)); err == nil {
		t.Fatal("newArtifactStore unexpectedly accepted a symlink escape")
	}
	if _, err := os.Stat(targetRoot); !os.IsNotExist(err) {
		t.Fatalf("symlink escape mutated %s: %v", targetRoot, err)
	}
}

func TestArtifactCollectionDeadline(t *testing.T) {
	tests := []struct {
		nodes int
		want  time.Duration
	}{
		{nodes: -1, want: 30 * time.Second},
		{nodes: 0, want: 30 * time.Second},
		{nodes: 1, want: 45 * time.Second},
		{nodes: 2, want: 60 * time.Second},
		{nodes: 3, want: 75 * time.Second},
		{nodes: 5, want: 105 * time.Second},
		{nodes: 6, want: 2 * time.Minute},
		{nodes: 100, want: 2 * time.Minute},
		{nodes: int(^uint(0) >> 1), want: 2 * time.Minute},
	}
	for _, test := range tests {
		if got := artifactCollectionDeadline(test.nodes); got != test.want {
			t.Errorf("artifactCollectionDeadline(%d) = %s, want %s", test.nodes, got, test.want)
		}
	}
}

func TestValidateIntentionallyStoppedContainer(t *testing.T) {
	tests := []struct {
		name      string
		expected  bool
		inspected bool
		running   bool
		handled   bool
		wantError string
	}{
		{name: "ordinary node", handled: false},
		{name: "stopped as expected", expected: true, inspected: true, handled: true},
		{name: "container missing", expected: true, handled: true, wantError: "no inspectable container"},
		{name: "unexpectedly running", expected: true, inspected: true, running: true, handled: true, wantError: "expected to remain stopped"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handled, err := validateIntentionallyStoppedContainer(test.expected, test.inspected, test.running)
			if handled != test.handled {
				t.Fatalf("handled = %t, want %t", handled, test.handled)
			}
			if test.wantError == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("error = %v, want substring %q", err, test.wantError)
			}
		})
	}
}

func TestArtifactLogBufferUsesSharedBound(t *testing.T) {
	budget := &artifactLogBudget{remaining: 10}
	stdout := &artifactLogBuffer{budget: budget}
	stderr := &artifactLogBuffer{budget: budget}
	if n, err := stdout.Write([]byte("12345678")); err != nil || n != 8 {
		t.Fatalf("stdout.Write = (%d, %v), want (8, nil)", n, err)
	}
	if n, err := stderr.Write([]byte("abcdefgh")); err != nil || n != 8 {
		t.Fatalf("stderr.Write = (%d, %v), want (8, nil)", n, err)
	}
	if got := stdout.Len() + stderr.Len(); got != 10 {
		t.Fatalf("captured %d bytes, want shared maximum 10", got)
	}
	if !budget.truncated {
		t.Fatal("shared log budget did not report truncation")
	}
}

func TestRecordCleanupPersistsFailuresAndSummary(t *testing.T) {
	store, err := newArtifactStore("test", "run-cleanup", artifactTestConfig(filepath.Join(trustedArtifactTempDir(t), "artifacts")))
	if err != nil {
		t.Fatalf("newArtifactStore: %v", err)
	}
	store.recordFailure("wait-height", errors.New("height stalled at 7"))
	cleanupErr := errors.New("labelled volume remains")
	if err := store.recordCleanup(nil, nil, cleanupErr); err != nil {
		t.Fatalf("recordCleanup: %v", err)
	}

	manifestBytes, err := os.ReadFile(filepath.Join(store.dir, "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var manifest struct {
		State    string            `json:"state"`
		Failed   bool              `json:"failed"`
		Failures []artifactFailure `json:"failures"`
		Cleanup  artifactCleanup   `json:"cleanup"`
	}
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if !manifest.Failed || manifest.State != "cleanup-failed" {
		t.Fatalf("manifest outcome = failed:%t state:%q", manifest.Failed, manifest.State)
	}
	if manifest.Cleanup.State != "completed" || manifest.Cleanup.Result != "failed" {
		t.Fatalf("cleanup = %#v", manifest.Cleanup)
	}
	if manifest.Cleanup.DockerCleanupError != cleanupErr.Error() {
		t.Fatalf("docker cleanup error = %q, want %q", manifest.Cleanup.DockerCleanupError, cleanupErr)
	}
	if len(manifest.Failures) < 2 {
		t.Fatalf("manifest failures = %#v, want recorded wait and cleanup failures", manifest.Failures)
	}

	summary, err := os.ReadFile(filepath.Join(store.dir, "failure-summary.txt"))
	if err != nil {
		t.Fatalf("read failure summary: %v", err)
	}
	if !strings.Contains(string(summary), "docker_cleanup_error: "+cleanupErr.Error()) {
		t.Fatalf("failure summary does not contain cleanup error:\n%s", summary)
	}
}

func TestBuildErrorCannotBeOverwrittenBySuccessfulCollection(t *testing.T) {
	store, err := newArtifactStore("test", "run-build-failure", artifactTestConfig(filepath.Join(trustedArtifactTempDir(t), "artifacts")))
	if err != nil {
		t.Fatalf("newArtifactStore: %v", err)
	}
	buildErr := errors.New("node failed before readiness")
	store.setBuildError(buildErr)

	if err := store.collect(false); err != nil {
		t.Fatalf("collect: %v", err)
	}
	if err := store.recordCleanup(nil, nil, nil); err != nil {
		t.Fatalf("recordCleanup: %v", err)
	}

	manifestBytes, err := os.ReadFile(filepath.Join(store.dir, "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var manifest struct {
		State      string `json:"state"`
		Failed     bool   `json:"failed"`
		BuildError string `json:"build_error"`
	}
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if !manifest.Failed {
		t.Fatalf("build failure manifest unexpectedly passed: %#v", manifest)
	}
	if manifest.State != "failed-cleaned" {
		t.Fatalf("manifest state = %q, want failed-cleaned", manifest.State)
	}
	if manifest.BuildError != buildErr.Error() {
		t.Fatalf("build error = %q, want %q", manifest.BuildError, buildErr)
	}
	if _, err := os.Stat(filepath.Join(store.dir, "failure-summary.txt")); err != nil {
		t.Fatalf("failure summary was not preserved: %v", err)
	}
}

func TestWriteFailureSummarySkipsHeightProbeWithoutRPCClient(t *testing.T) {
	store, err := newArtifactStore(
		"test",
		"run-no-rpc",
		artifactTestConfig(filepath.Join(trustedArtifactTempDir(t), "artifacts")),
	)
	if err != nil {
		t.Fatalf("newArtifactStore: %v", err)
	}
	chain := &cosmos.CosmosChain{
		Validators: cosmos.ChainNodes{&cosmos.ChainNode{}},
	}

	if err := store.writeFailureSummary(context.Background(), chain, nil); err != nil {
		t.Fatalf("writeFailureSummary: %v", err)
	}
	summary, err := os.ReadFile(filepath.Join(store.dir, "failure-summary.txt"))
	if err != nil {
		t.Fatalf("read failure summary: %v", err)
	}
	if !strings.Contains(string(summary), "last_height: 0") {
		t.Fatalf("failure summary does not preserve zero height:\n%s", summary)
	}
}

func artifactTestConfig(root string) Config {
	return Config{
		Image:         ImageRef{Repository: "panacea-test", Version: "test"},
		NumValidators: 1,
		NumFullNodes:  1,
		ArtifactRoot:  root,
	}
}

func trustedArtifactTempDir(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("real-node E2E artifact isolation requires the Unix /tmp root")
	}
	dir, err := os.MkdirTemp("/tmp", "panacea-e2e-test-")
	if err != nil {
		t.Fatalf("create trusted artifact temp directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("remove trusted artifact temp directory %s: %v", dir, err)
		}
	})
	return dir
}
