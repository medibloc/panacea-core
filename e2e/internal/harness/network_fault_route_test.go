package harness

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"
)

func TestRewriteCometP2PRouteRemovesEveryDirectBypass(t *testing.T) {
	t.Parallel()
	maxOutboundPeers := int64(0)
	input := []byte("\nmoniker = \"node\"\n[p2p]\nladdr = \"tcp://0.0.0.0:26656\"\nmax_num_outbound_peers = 10\npersistent_peers = \"id@direct-validator:26656,id2@direct-full-node:26656\"\nseeds = \"seed@seed:26656\"\npex = true\nunconditional_peer_ids = \"id\"\n")
	output, err := rewriteCometP2PRoute(input, cometP2PRouteMutation{
		persistentPeers:  "validator-id@fault-proxy:26656",
		peerExchange:     false,
		listenAddress:    "tcp://0.0.0.0:27656",
		maxOutboundPeers: &maxOutboundPeers,
	})
	require.NoError(t, err)
	var decoded map[string]any
	_, err = toml.Decode(string(output), &decoded)
	require.NoError(t, err)
	p2p := decoded["p2p"].(map[string]any)
	require.Equal(t, "validator-id@fault-proxy:26656", p2p["persistent_peers"])
	require.Equal(t, "", p2p["seeds"])
	require.Equal(t, false, p2p["pex"])
	require.Equal(t, "", p2p["unconditional_peer_ids"])
	require.Equal(t, "tcp://0.0.0.0:27656", p2p["laddr"])
	require.EqualValues(t, 0, p2p["max_num_outbound_peers"])
	require.NotContains(t, string(output), "direct-validator")
	require.NotContains(t, string(output), "direct-full-node")
}

func TestCometP2PProxyRouteMutationsDisableNonPersistentBypassOnBothNodes(t *testing.T) {
	t.Parallel()

	validator, fullNode := cometP2PProxyRouteMutations("validator-id", "fault-proxy")
	require.NotNil(t, validator.maxOutboundPeers)
	require.EqualValues(t, 0, *validator.maxOutboundPeers)
	require.False(t, validator.peerExchange)
	require.Equal(t, "tcp://0.0.0.0:27656", validator.listenAddress)

	require.NotNil(t, fullNode.maxOutboundPeers)
	require.EqualValues(t, 0, *fullNode.maxOutboundPeers)
	require.False(t, fullNode.peerExchange)
	require.Equal(t, "validator-id@fault-proxy:26656", fullNode.persistentPeers)
}

func TestDirectedNetworkFaultP2PMutationsUseOneWayPersistentDialing(t *testing.T) {
	t.Parallel()

	validator, fullNode, err := directedNetworkFaultP2PMutations(
		"validator-id",
		"172.30.0.2",
	)
	require.NoError(t, err)

	require.Empty(t, validator.persistentPeers)
	require.False(t, validator.peerExchange)
	require.NotNil(t, validator.maxOutboundPeers)
	require.Zero(t, *validator.maxOutboundPeers)

	require.Equal(t, "validator-id@172.30.0.2:26656", fullNode.persistentPeers)
	require.False(t, fullNode.peerExchange)
	require.NotNil(t, fullNode.maxOutboundPeers)
	require.Zero(t, *fullNode.maxOutboundPeers)
}

func TestDirectedNetworkFaultP2PMutationsRejectInvalidIdentity(t *testing.T) {
	t.Parallel()

	_, _, err := directedNetworkFaultP2PMutations("", "172.30.0.2")
	require.ErrorContains(t, err, "validator node ID")
	_, _, err = directedNetworkFaultP2PMutations("validator-id", "validator-host")
	require.ErrorContains(t, err, "validator IPv4 address")
}
