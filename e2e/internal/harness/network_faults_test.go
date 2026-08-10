package harness

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/require"
)

func TestRecordNetworkFaultCleanupPreservesFailureInArtifactAndManifest(t *testing.T) {
	store, err := newArtifactStore(
		"test",
		"run-network-cleanup",
		artifactTestConfig(filepath.Join(trustedArtifactTempDir(t), "artifacts")),
	)
	require.NoError(t, err)
	network := &Network{artifacts: store}
	cleanupErr := errors.New("restore proxy route failed")

	err = network.RecordNetworkFaultCleanup("proxy-route", cleanupErr)
	require.ErrorIs(t, err, cleanupErr)

	contents, err := os.ReadFile(filepath.Join(store.dir, "network-faults", "cleanup.jsonl"))
	require.NoError(t, err)
	var evidence NetworkFaultCleanupEvidence
	require.NoError(t, json.Unmarshal(contents, &evidence))
	require.Equal(t, "proxy-route", evidence.Phase)
	require.Equal(t, "failed", evidence.Result)
	require.Equal(t, cleanupErr.Error(), evidence.Error)
	require.Len(t, store.failures, 1)
	require.Equal(t, "network-fault-cleanup-proxy-route", store.failures[0].Stage)
}

func TestRecordNetworkFaultCategorySeparatesEnvironmentFailureFromExpectedP2PRuntimeFault(t *testing.T) {
	store, err := newArtifactStore(
		"test",
		"run-network-category",
		artifactTestConfig(filepath.Join(trustedArtifactTempDir(t), "artifacts")),
	)
	require.NoError(t, err)
	network := &Network{artifacts: store}

	require.NoError(t, network.RecordNetworkFaultCategory(NetworkFaultCategoryEvidence{
		Category: NetworkFaultCategoryEnvironmentPreflight,
		Phase:    "docker-socket",
		Outcome:  NetworkFaultOutcomeFailed,
		Scope:    NetworkFaultScopeLocalEnvironment,
		Error:    "permission denied",
	}))
	require.NoError(t, network.RecordNetworkFaultCategory(NetworkFaultCategoryEvidence{
		Category: NetworkFaultCategoryChainP2PRuntime,
		Phase:    "full-node-partition",
		Outcome:  NetworkFaultOutcomeExpectedFaultObserved,
		Scope:    NetworkFaultScopeRunOwnedDockerP2P,
	}))

	environment, err := os.ReadFile(filepath.Join(store.dir, "environment", "network-failure-categories.jsonl"))
	require.NoError(t, err)
	require.Contains(t, string(environment), `"category":"environment-preflight"`)
	runtime, err := os.ReadFile(filepath.Join(store.dir, "network-faults", "failure-categories.jsonl"))
	require.NoError(t, err)
	require.Contains(t, string(runtime), `"category":"chain-p2p-runtime"`)
	require.Len(t, store.failures, 1)
	require.Equal(t, "environment-preflight-docker-socket", store.failures[0].Stage)
}

func TestP2PFaultProxyConfigValidate(t *testing.T) {
	t.Parallel()
	valid := P2PFaultProxyConfig{
		Name:          "latency-loss",
		Alias:         "p2p-fault-proxy",
		Image:         ImageRef{Repository: "panacea-e2e-current", Version: "local"},
		TargetAddress: "validator:26656",
		Delay:         50 * time.Millisecond,
		Jitter:        20 * time.Millisecond,
		DropEvery:     10,
		Seed:          20260804,
	}
	require.NoError(t, valid.Validate())

	tests := map[string]func(*P2PFaultProxyConfig){
		"unsafe name":      func(config *P2PFaultProxyConfig) { config.Name = "../proxy" },
		"unsafe alias":     func(config *P2PFaultProxyConfig) { config.Alias = "proxy.local" },
		"missing image":    func(config *P2PFaultProxyConfig) { config.Image = ImageRef{} },
		"missing target":   func(config *P2PFaultProxyConfig) { config.TargetAddress = "" },
		"invalid IPv4":     func(config *P2PFaultProxyConfig) { config.IPv4Address = "not-an-ip" },
		"negative delay":   func(config *P2PFaultProxyConfig) { config.Delay = -time.Millisecond },
		"negative jitter":  func(config *P2PFaultProxyConfig) { config.Jitter = -time.Millisecond },
		"zero random seed": func(config *P2PFaultProxyConfig) { config.Seed = 0 },
	}
	for name, mutate := range tests {
		name, mutate := name, mutate
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			config := valid
			mutate(&config)
			require.Error(t, config.Validate())
		})
	}
}

func TestResolveNetworkFaultEndpointAcceptsDockerNetworkIDOrName(t *testing.T) {
	t.Parallel()
	endpoint := &dockernetwork.EndpointSettings{
		NetworkID: "network-sha256-id",
		IPAddress: "172.30.0.4",
	}
	networks := map[string]*dockernetwork.EndpointSettings{
		"interchaintest-abcd": endpoint,
	}

	resolvedByID, ok := resolveNetworkFaultEndpoint(networks, "network-sha256-id")
	require.True(t, ok)
	require.Same(t, endpoint, resolvedByID)
	resolvedByName, ok := resolveNetworkFaultEndpoint(networks, "interchaintest-abcd")
	require.True(t, ok)
	require.Same(t, endpoint, resolvedByName)
	_, ok = resolveNetworkFaultEndpoint(networks, "missing")
	require.False(t, ok)
	_, ok = resolveNetworkFaultEndpoint(networks, "")
	require.False(t, ok)
}

func TestValidateNetworkFaultAttachmentTransitionRequiresObservedIP(t *testing.T) {
	t.Parallel()

	require.NoError(t, validateNetworkFaultAttachmentTransition("disconnect", true, false, "172.30.0.4", ""))
	require.NoError(t, validateNetworkFaultAttachmentTransition("reconnect", false, true, "", "172.30.0.5"))
	require.Error(t, validateNetworkFaultAttachmentTransition("disconnect", true, false, "", ""))
	require.Error(t, validateNetworkFaultAttachmentTransition("disconnect", true, true, "172.30.0.4", "172.30.0.4"))
	require.Error(t, validateNetworkFaultAttachmentTransition("reconnect", false, true, "", ""))
	require.Error(t, validateNetworkFaultAttachmentTransition("reconnect", true, true, "172.30.0.4", "172.30.0.4"))
}
