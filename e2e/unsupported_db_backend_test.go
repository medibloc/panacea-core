package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	unsupportedDBBackend          = "badgerdb"
	unsupportedDBOperationTimeout = time.Minute
	unsupportedDBChildTimeout     = 4 * time.Minute
	unsupportedDBParentTimeout    = 5 * time.Minute
)

// TestUnsupportedDBBackendFailsStartup proves the current image rejects a
// legacy Cosmos SDK database backend at the real panacead process boundary.
// The subprocess is intentionally failed so the normal harness failure path
// records diagnostics before label-scoped cleanup runs.
func TestUnsupportedDBBackendFailsStartup(t *testing.T) {
	if os.Getenv("PANACEA_E2E_FAILURE_PROBE") != "1" {
		t.Skip("set PANACEA_E2E_FAILURE_PROBE=1 or use ./scripts/e2e/run.sh smoke")
	}
	if os.Getenv("PANACEA_E2E_UNSUPPORTED_DB_CHILD") == "1" {
		runUnsupportedDBBackendChild(t)
		return
	}

	runID := fmt.Sprintf("db-failure-%x", time.Now().UnixNano())
	artifactRoot := trustedE2ETempDir(t)
	probeCtx, cancel := context.WithTimeout(context.Background(), unsupportedDBParentTimeout)
	defer cancel()
	command := exec.CommandContext(
		probeCtx,
		os.Args[0],
		"-test.run=^TestUnsupportedDBBackendFailsStartup$",
		"-test.timeout="+unsupportedDBChildTimeout.String(),
		"-test.v",
	)
	command.Env = append(os.Environ(),
		"PANACEA_E2E_UNSUPPORTED_DB_CHILD=1",
		"PANACEA_E2E_UNSUPPORTED_DB_RUN_ID="+runID,
		"PANACEA_E2E_UNSUPPORTED_DB_ROOT="+artifactRoot,
	)
	output, err := command.CombinedOutput()
	require.Error(t, err, "unsupported-DB child unexpectedly passed:\n%s", output)
	require.NoError(t, probeCtx.Err(), "unsupported-DB child exceeded its bounded startup window:\n%s", output)
	marker, clear := harness.ClassifyUnsupportedDBBackendStartup(unsupportedDBBackend, string(output))
	require.True(t, clear, "child output did not contain the clear SDK backend rejection %q:\n%s", marker, output)
	require.Contains(t, string(output), "intentional unsupported DB backend startup probe")

	runDir := filepath.Join(artifactRoot, runID)
	requireUnsupportedDBFailureArtifacts(t, runDir)
	requireNoLabeledDockerResources(t, runID)
}

func runUnsupportedDBBackendChild(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), unsupportedDBOperationTimeout)
	defer cancel()

	network, err := harness.Start(ctx, t, harness.Config{
		Image:         harness.CurrentImage(),
		NumValidators: 1,
		RunID:         os.Getenv("PANACEA_E2E_UNSUPPORTED_DB_RUN_ID"),
		ArtifactRoot:  os.Getenv("PANACEA_E2E_UNSUPPORTED_DB_ROOT"),
		DBBackend:     unsupportedDBBackend,
	})
	require.NotNil(t, network, "the harness must retain the failed network for diagnostic collection")
	defer network.RecordTestPanic()
	require.Error(t, err, "panacead unexpectedly started with %s", unsupportedDBBackend)
	t.Logf("unsupported DB startup failure: %v", err)
	marker, clear := harness.ClassifyUnsupportedDBBackendStartup(unsupportedDBBackend, err.Error())
	require.True(t, clear, "startup failed ambiguously instead of rejecting %s clearly: %v", unsupportedDBBackend, err)

	t.Errorf("intentional unsupported DB backend startup probe: %s", marker)
}

func requireUnsupportedDBFailureArtifacts(t *testing.T, runDir string) {
	t.Helper()
	for _, name := range []string{"manifest.json", "cleanup.json", "failure-summary.txt"} {
		require.FileExists(t, filepath.Join(runDir, name))
	}

	manifestBytes, err := os.ReadFile(filepath.Join(runDir, "manifest.json"))
	require.NoError(t, err)
	var manifest struct {
		Failed     bool   `json:"failed"`
		BuildError string `json:"build_error"`
		Cleanup    struct {
			State              string `json:"state"`
			DockerCleanupError string `json:"docker_cleanup_error"`
		} `json:"cleanup"`
	}
	require.NoError(t, json.Unmarshal(manifestBytes, &manifest))
	require.True(t, manifest.Failed)
	marker, clear := harness.ClassifyUnsupportedDBBackendStartup(unsupportedDBBackend, manifest.BuildError)
	require.True(t, clear, "manifest build_error did not preserve %q: %s", marker, manifest.BuildError)
	require.Equal(t, "completed", manifest.Cleanup.State)
	require.Empty(t, manifest.Cleanup.DockerCleanupError)

	nodes, err := os.ReadDir(filepath.Join(runDir, "nodes"))
	require.NoError(t, err)
	require.Len(t, nodes, 1)
	nodeDir := filepath.Join(runDir, "nodes", nodes[0].Name())
	configBytes, err := os.ReadFile(filepath.Join(nodeDir, "config", "config.toml"))
	require.NoError(t, err)
	require.Contains(t, string(configBytes), `db_backend = "badgerdb"`)
	containerStateBytes, err := os.ReadFile(filepath.Join(nodeDir, "container-state.json"))
	require.NoError(t, err)
	var containerState struct {
		State struct {
			Running  bool `json:"Running"`
			ExitCode int  `json:"ExitCode"`
		} `json:"state"`
	}
	require.NoError(t, json.Unmarshal(containerStateBytes, &containerState))
	require.False(t, containerState.State.Running)
	require.NotZero(t, containerState.State.ExitCode)
	logBytes, err := os.ReadFile(filepath.Join(nodeDir, "logs", "container.log"))
	require.NoError(t, err)
	_, clear = harness.ClassifyUnsupportedDBBackendStartup(unsupportedDBBackend, string(logBytes))
	require.True(t, clear, "container log did not preserve the process-level backend rejection")
}
