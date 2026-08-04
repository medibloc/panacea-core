package harness

import (
	"context"
	"errors"
	"testing"

	"github.com/BurntSushi/toml"
	dockertypes "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/strangelove-ventures/interchaintest/v8/dockerutil"
	"github.com/stretchr/testify/require"
)

func TestRewriteCometRPCListenerForNetworkFaultChangesOnlyListener(t *testing.T) {
	t.Parallel()
	input := []byte(`moniker = "full-node"
[rpc]
laddr = "tcp://0.0.0.0:26657"
max_open_connections = 900
unsafe = false
[p2p]
laddr = "tcp://0.0.0.0:26656"
persistent_peers = "validator@validator:26656"
`)

	output, originalListener, err := rewriteCometRPCListenerForNetworkFault(input)
	require.NoError(t, err)
	require.Equal(t, "tcp://0.0.0.0:26657", originalListener)

	var before, after map[string]any
	_, err = toml.Decode(string(input), &before)
	require.NoError(t, err)
	_, err = toml.Decode(string(output), &after)
	require.NoError(t, err)
	afterRPC := after["rpc"].(map[string]any)
	require.Equal(t, NetworkFaultLoopbackRPCListener, afterRPC["laddr"])
	afterRPC["laddr"] = originalListener
	require.Equal(t, before, after, "RPC boundary fault must preserve every other semantic config value")
}

func TestShouldRetryRPCFaultNormalStartRequiresOwnedLiveCatchingUpState(t *testing.T) {
	t.Parallel()
	startErr := errors.New("still catching up")
	require.True(t, shouldRetryRPCFaultNormalStart(startErr, nil, true, true))
	require.False(t, shouldRetryRPCFaultNormalStart(nil, nil, true, true))
	require.False(t, shouldRetryRPCFaultNormalStart(startErr, context.Canceled, true, true))
	require.False(t, shouldRetryRPCFaultNormalStart(startErr, nil, false, true))
	require.False(t, shouldRetryRPCFaultNormalStart(startErr, nil, true, false))
}

func TestValidateRunOwnedRPCFaultContainerRequiresExactLabelAndVolume(t *testing.T) {
	t.Parallel()
	valid := dockertypes.ContainerJSON{
		ContainerJSONBase: &dockertypes.ContainerJSONBase{
			ID:    "container-id",
			State: &dockertypes.ContainerState{Running: true},
		},
		Config: &dockercontainer.Config{Labels: map[string]string{
			dockerutil.CleanupLabel: "run-owned",
		}},
		Mounts: []dockertypes.MountPoint{{
			Name:        "node-volume",
			Destination: "/var/cosmos-chain/panacea",
			RW:          true,
		}},
	}

	identity, err := validateRunOwnedRPCFaultContainer(
		valid,
		"container-id",
		"run-owned",
		"node-volume",
		"/var/cosmos-chain/panacea",
	)
	require.NoError(t, err)
	require.Equal(t, "container-id", identity.ContainerID)
	require.Equal(t, "run-owned", identity.CleanupLabel)
	require.Equal(t, "node-volume", identity.VolumeName)
	require.True(t, identity.VolumeWritable)

	wrongID := valid
	wrongID.ContainerJSONBase = &dockertypes.ContainerJSONBase{ID: "other", State: valid.State}
	_, err = validateRunOwnedRPCFaultContainer(wrongID, "container-id", "run-owned", "node-volume", "/var/cosmos-chain/panacea")
	require.ErrorContains(t, err, "does not match expected")

	wrongLabel := valid
	wrongLabel.Config = &dockercontainer.Config{Labels: map[string]string{dockerutil.CleanupLabel: "another-run"}}
	_, err = validateRunOwnedRPCFaultContainer(wrongLabel, "container-id", "run-owned", "node-volume", "/var/cosmos-chain/panacea")
	require.ErrorContains(t, err, "does not match run")

	wrongVolume := valid
	wrongVolume.Mounts = []dockertypes.MountPoint{{Name: "other-volume", Destination: "/var/cosmos-chain/panacea", RW: true}}
	_, err = validateRunOwnedRPCFaultContainer(wrongVolume, "container-id", "run-owned", "node-volume", "/var/cosmos-chain/panacea")
	require.ErrorContains(t, err, "has no expected volume")

	readOnly := valid
	readOnly.Mounts = []dockertypes.MountPoint{{Name: "node-volume", Destination: "/var/cosmos-chain/panacea", RW: false}}
	_, err = validateRunOwnedRPCFaultContainer(readOnly, "container-id", "run-owned", "node-volume", "/var/cosmos-chain/panacea")
	require.ErrorContains(t, err, "not writable")
}
