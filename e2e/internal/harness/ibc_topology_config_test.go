package harness

import (
	"testing"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestIBCRelayerCleanupTimeoutOutlivesInterchaintestDockerStop(t *testing.T) {
	t.Parallel()

	require.Equal(t, 30*time.Second, interchaintestRelayerStopGrace)
	require.Equal(t, 45*time.Second, ibcRelayerCleanupTimeout)
	require.Greater(t, ibcRelayerCleanupTimeout, interchaintestRelayerStopGrace)
}

func TestNewOsmosisChainSpecPinsMainnetCounterpartyRelease(t *testing.T) {
	t.Parallel()

	spec, err := NewOsmosisChainSpec("ibc-a1b2c3")
	require.NoError(t, err)

	cfg, err := spec.Config(zap.NewNop())
	require.NoError(t, err)
	require.Equal(t, ibc.Cosmos, cfg.Type)
	require.Equal(t, "osmosis-ibc-a1b2c3", cfg.Name)
	require.Equal(t, "osmosis-ibc-a1b2c3", cfg.ChainID)
	require.Equal(t, "osmosisd", cfg.Bin)
	require.Equal(t, "osmo", cfg.Bech32Prefix)
	require.Equal(t, "uosmo", cfg.Denom)
	require.Equal(t, "118", cfg.CoinType)
	require.NotNil(t, cfg.CoinDecimals)
	require.EqualValues(t, 6, *cfg.CoinDecimals)
	// Osmosis v31's local feemarket starts at roughly 0.03uosmo.  The relayer
	// and test-user transactions must meet that floor or the handshake can
	// return a JSON ChainError while Hermes itself exits successfully.
	require.Equal(t, "0.03uosmo", cfg.GasPrices)
	require.Equal(t, 2.0, cfg.GasAdjustment)
	require.Equal(t, "auto", cfg.Gas)
	require.Equal(t, "336h", cfg.TrustingPeriod)
	require.Equal(t, 1, *spec.NumValidators)
	require.Equal(t, 1, *spec.NumFullNodes)

	wantImage := ibc.DockerImage{
		Repository: "docker.io/osmolabs/osmosis",
		Version:    "31.0.2@sha256:8de930072fef03ea034b5a38f3cf93e5f47b6ccb8b1776a34e402aa47c819e0e",
		UIDGID:     "0:0",
	}
	require.Equal(t, []ibc.DockerImage{wantImage}, cfg.Images)
	require.Equal(t, "docker.io/osmolabs/osmosis:31.0.2@sha256:8de930072fef03ea034b5a38f3cf93e5f47b6ccb8b1776a34e402aa47c819e0e", wantImage.Ref())

	cometOverrides, ok := cfg.ConfigFileOverrides["config/config.toml"].(testutil.Toml)
	require.True(t, ok)
	consensusOverrides, ok := cometOverrides["consensus"].(testutil.Toml)
	require.True(t, ok)
	require.Equal(t, "1s", consensusOverrides["timeout_commit"])
}

func TestPinnedIBCProvenanceIdentifiesExactOsmosisAndHermesSources(t *testing.T) {
	t.Parallel()

	provenance := PinnedIBCProvenance()

	require.Equal(t, "docker.io/osmolabs/osmosis:31.0.2@sha256:8de930072fef03ea034b5a38f3cf93e5f47b6ccb8b1776a34e402aa47c819e0e", provenance.Osmosis.Reference)
	require.Equal(t, "a56c05b0e83341b9a3c0e6e3508520f15e9f2e49", provenance.Osmosis.SourceCommit)
	require.Equal(t, "v0.50.14-v30-osmo", provenance.Osmosis.CosmosSDKVersion)
	require.Equal(t, "v0.38.22", provenance.Osmosis.CometBFTVersion)
	require.Equal(t, "v8.7.0", provenance.Osmosis.IBCGoVersion)

	require.Equal(t, ibc.DockerImage{
		Repository: "ghcr.io/informalsystems/hermes",
		Version:    "1.8.2@sha256:5422e8a26bf42db4a6223e999823df9269428bc936c9bc5826221632304d28b1",
		UIDGID:     "1000:1000",
	}, PinnedHermesImage())
	require.Equal(t, "ghcr.io/informalsystems/hermes:1.8.2@sha256:5422e8a26bf42db4a6223e999823df9269428bc936c9bc5826221632304d28b1", provenance.Hermes.Reference)
	require.Equal(t, "06dfbafb4893255a79043ec4032034a83ebd53df", provenance.Hermes.SourceCommit)
	require.Equal(t, "1.8.2+06dfbaf", provenance.Hermes.ReleaseIdentifier)
}

func TestIBCTopologyPlanGivesEveryComponentOneRunIdentity(t *testing.T) {
	t.Parallel()

	plan, err := newIBCTopologyPlan(IBCTopologyConfig{
		RunID:        "ibc-plan-a1b2c3",
		PanaceaImage: ImageRef{Repository: "panacea-e2e-v2.2.1", Version: "local"},
		ArtifactRoot: "/tmp/panacea-ibc-plan",
	})
	require.NoError(t, err)
	require.Equal(t, "ibc-plan-a1b2c3", plan.runID)
	require.Equal(t, "panacea-osmosis", plan.path)
	require.Equal(t, PinnedHermesImage(), plan.hermesImage)
	require.Equal(t, "panacea-ibc-plan-a1b2c3", plan.panaceaSpec.ChainConfig.ChainID)
	require.Equal(t, "osmosis-ibc-plan-a1b2c3", plan.osmosisSpec.ChainConfig.ChainID)
	require.Equal(t, IBCTopologyDescriptor{
		Path:              "panacea-osmosis",
		PanaceaChainID:    "panacea-ibc-plan-a1b2c3",
		OsmosisChainID:    "osmosis-ibc-plan-a1b2c3",
		PanaceaValidators: 1,
		PanaceaFullNodes:  1,
		OsmosisValidators: 1,
		OsmosisFullNodes:  1,
		SkipPathCreation:  true,
	}, plan.descriptor)
	require.Equal(t, "ibc-plan-a1b2c3", plan.artifactConfig.RunID)
	require.Equal(t, "/tmp/panacea-ibc-plan", plan.artifactConfig.ArtifactRoot)
}
