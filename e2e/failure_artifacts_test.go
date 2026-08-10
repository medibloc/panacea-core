package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	volumetypes "github.com/docker/docker/api/types/volume"
	dockerclient "github.com/docker/docker/client"
	"github.com/strangelove-ventures/interchaintest/v8/dockerutil"
	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	failureProbeOperationTimeout = 2 * time.Minute
	failureProbeChildTimeout     = 5 * time.Minute
)

func TestFailureArtifactsAndCleanup(t *testing.T) {
	if os.Getenv("PANACEA_E2E_FAILURE_PROBE") != "1" {
		t.Skip("set PANACEA_E2E_FAILURE_PROBE=1 or use ./scripts/e2e/run.sh smoke")
	}
	if os.Getenv("PANACEA_E2E_FAILURE_PROBE_CHILD") == "1" {
		runFailureProbeChild(t)
		return
	}

	runID := fmt.Sprintf("failure-%x", time.Now().UnixNano())
	artifactRoot := trustedE2ETempDir(t)
	command := exec.Command(
		os.Args[0],
		"-test.run=^TestFailureArtifactsAndCleanup$",
		"-test.timeout="+failureProbeChildTimeout.String(),
		"-test.v",
	)
	command.Env = append(os.Environ(),
		"PANACEA_E2E_FAILURE_PROBE_CHILD=1",
		"PANACEA_E2E_FAILURE_PROBE_RUN_ID="+runID,
		"PANACEA_E2E_FAILURE_PROBE_ROOT="+artifactRoot,
	)
	output, err := command.CombinedOutput()
	require.Error(t, err, "failure-probe child unexpectedly passed:\n%s", output)
	require.Contains(t, string(output), "intentional artifact probe failure")

	runDir := filepath.Join(artifactRoot, runID)
	requireFailureArtifactContract(t, runDir)
	requireNoLabeledDockerResources(t, runID)
}

func trustedE2ETempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "panacea-e2e-live-test-")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(dir))
	})
	return dir
}

func runFailureProbeChild(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), failureProbeOperationTimeout)
	defer cancel()

	network, err := harness.Start(ctx, t, harness.Config{
		Image:         harness.CurrentImage(),
		NumValidators: 1,
		NumFullNodes:  1,
		RunID:         os.Getenv("PANACEA_E2E_FAILURE_PROBE_RUN_ID"),
		ArtifactRoot:  os.Getenv("PANACEA_E2E_FAILURE_PROBE_ROOT"),
	})
	require.NoError(t, err)
	defer network.RecordTestPanic()

	height, err := network.Chain.Height(ctx)
	require.NoError(t, err)
	require.NoError(t, network.WaitForFullNode(ctx, height))

	probeCtx, probeCancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer probeCancel()
	err = network.WaitForHeight(probeCtx, height+1_000_000)
	require.Error(t, err)
	t.Errorf("intentional artifact probe failure: %v", err)
}

func requireFailureArtifactContract(t *testing.T, runDir string) {
	t.Helper()
	for _, name := range []string{"manifest.json", "cleanup.json", "genesis.json", "versions.txt", "failure-summary.txt"} {
		require.FileExists(t, filepath.Join(runDir, name))
	}

	manifestBytes, err := os.ReadFile(filepath.Join(runDir, "manifest.json"))
	require.NoError(t, err)
	var manifest struct {
		State    string `json:"state"`
		Failed   bool   `json:"failed"`
		Failures []struct {
			Stage string `json:"stage"`
			Error string `json:"error"`
		} `json:"failures"`
		Cleanup struct {
			State  string `json:"state"`
			Result string `json:"result"`
		} `json:"cleanup"`
	}
	require.NoError(t, json.Unmarshal(manifestBytes, &manifest))
	require.True(t, manifest.Failed)
	require.Equal(t, "failed-cleaned", manifest.State)
	require.Equal(t, "completed", manifest.Cleanup.State)
	require.Equal(t, "succeeded", manifest.Cleanup.Result)
	require.NotEmpty(t, manifest.Failures)
	require.Equal(t, "wait-validator-height", manifest.Failures[0].Stage)

	summary, err := os.ReadFile(filepath.Join(runDir, "failure-summary.txt"))
	require.NoError(t, err)
	require.Contains(t, string(summary), "failure[wait-validator-height]")
	require.Contains(t, string(summary), "last_height:")

	nodes, err := os.ReadDir(filepath.Join(runDir, "nodes"))
	require.NoError(t, err)
	require.Len(t, nodes, 2)
	for _, node := range nodes {
		require.True(t, node.IsDir())
		base := filepath.Join(runDir, "nodes", node.Name())
		for _, name := range []string{"app.toml", "config.toml", "client.toml"} {
			require.FileExists(t, filepath.Join(base, "config", name))
		}
		require.FileExists(t, filepath.Join(base, "status.json"))
		require.FileExists(t, filepath.Join(base, "container-state.json"))
		require.FileExists(t, filepath.Join(base, "logs", "container.log"))
	}

	forbiddenNames := map[string]struct{}{
		"priv_validator_key.json":   {},
		"priv_validator_state.json": {},
		"node_key.json":             {},
	}
	require.NoError(t, filepath.WalkDir(runDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if _, forbidden := forbiddenNames[entry.Name()]; forbidden {
			return fmt.Errorf("secret-bearing file was collected: %s", path)
		}
		if strings.HasPrefix(entry.Name(), "keyring-") {
			return fmt.Errorf("keyring was collected: %s", path)
		}
		return nil
	}))
}

func requireNoLabeledDockerResources(t *testing.T, runID string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	client, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	require.NoError(t, err)
	defer client.Close()

	label := filters.NewArgs(filters.Arg("label", dockerutil.CleanupLabel+"="+runID))
	containers, err := client.ContainerList(ctx, dockertypes.ContainerListOptions{All: true, Filters: label})
	require.NoError(t, err)
	require.Empty(t, containers)
	volumes, err := client.VolumeList(ctx, volumetypes.ListOptions{Filters: label})
	require.NoError(t, err)
	require.Empty(t, volumes.Volumes)
	networks, err := client.NetworkList(ctx, dockertypes.NetworkListOptions{Filters: label})
	require.NoError(t, err)
	require.Empty(t, networks)
}
