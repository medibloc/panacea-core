package harness

import (
	"encoding/json"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewPanaceaChainSpecBuildsIsolatedRealNodeTopology(t *testing.T) {
	t.Parallel()

	spec, err := NewPanaceaChainSpec(
		"smoke-a1b2c3",
		ImageRef{Repository: "panacea-e2e", Version: "current"},
		Topology{Validators: 1, FullNodes: 1},
	)
	require.NoError(t, err)

	cfg, err := spec.Config(zap.NewNop())
	require.NoError(t, err)
	require.Equal(t, ibc.Cosmos, cfg.Type)
	require.Equal(t, "panacea-smoke-a1b2c3", cfg.Name)
	require.Equal(t, "panacea-smoke-a1b2c3", cfg.ChainID)
	require.Equal(t, "panacead", cfg.Bin)
	require.Equal(t, "panacea", cfg.Bech32Prefix)
	require.Equal(t, "umed", cfg.Denom)
	require.Equal(t, "371", cfg.CoinType)
	require.NotNil(t, cfg.CoinDecimals)
	require.EqualValues(t, 6, *cfg.CoinDecimals)
	require.Equal(t, "5umed", cfg.GasPrices)
	require.Equal(t, 1.3, cfg.GasAdjustment)
	require.False(t, cfg.UsingChainIDFlagCLI)
	require.Equal(t, []ibc.DockerImage{{
		Repository: "panacea-e2e",
		Version:    "current",
		UIDGID:     "0:0",
	}}, cfg.Images)
	require.Equal(t, 1, *spec.NumValidators)
	require.Equal(t, 1, *spec.NumFullNodes)

	appOverrides, ok := cfg.ConfigFileOverrides["config/app.toml"].(testutil.Toml)
	require.True(t, ok)
	require.EqualValues(t, 10_000_000, appOverrides["query-gas-limit"])

	cometOverrides, ok := cfg.ConfigFileOverrides["config/config.toml"].(testutil.Toml)
	require.True(t, ok)
	require.Equal(t, "goleveldb", cometOverrides["db_backend"])
}

func TestNewPanaceaChainSpecAllowsExplicitDiagnosticDBBackendOverride(t *testing.T) {
	t.Parallel()

	spec, err := NewPanaceaChainSpec(
		"db-failure-a1b2c3",
		ImageRef{Repository: "panacea-e2e", Version: "current"},
		Topology{Validators: 1, DBBackend: "badgerdb"},
	)
	require.NoError(t, err)
	cfg, err := spec.Config(zap.NewNop())
	require.NoError(t, err)

	cometOverrides := cfg.ConfigFileOverrides["config/config.toml"].(testutil.Toml)
	require.Equal(t, "badgerdb", cometOverrides["db_backend"])
}

func TestPanaceaGenesisModifierUsesCrossGenerationGovernancePaths(t *testing.T) {
	t.Parallel()

	spec, err := NewPanaceaChainSpec(
		"genesis-a1b2c3",
		ImageRef{Repository: "panacea-e2e", Version: "current"},
		Topology{Validators: 1, FullNodes: 1},
	)
	require.NoError(t, err)
	cfg, err := spec.Config(zap.NewNop())
	require.NoError(t, err)
	require.NotNil(t, cfg.ModifyGenesis)

	input := []byte(`{
		"app_state": {
			"gov": {
				"params": {
					"voting_period": "259200s",
					"max_deposit_period": "1209600s",
					"min_deposit": [{"denom":"umed","amount":"100000000000"}]
				}
			}
		}
	}`)
	modified, err := cfg.ModifyGenesis(*cfg, input)
	require.NoError(t, err)

	var output struct {
		AppState struct {
			Gov struct {
				Params struct {
					VotingPeriod     string `json:"voting_period"`
					MaxDepositPeriod string `json:"max_deposit_period"`
					MinDeposit       []struct {
						Denom  string `json:"denom"`
						Amount string `json:"amount"`
					} `json:"min_deposit"`
				} `json:"params"`
			} `json:"gov"`
		} `json:"app_state"`
	}
	require.NoError(t, json.Unmarshal(modified, &output))
	require.Equal(t, "20s", output.AppState.Gov.Params.VotingPeriod)
	require.Equal(t, "20s", output.AppState.Gov.Params.MaxDepositPeriod)
	require.Equal(t, "umed", output.AppState.Gov.Params.MinDeposit[0].Denom)
	require.Equal(t, "1", output.AppState.Gov.Params.MinDeposit[0].Amount)
}

func TestNewPanaceaChainSpecAppliesSuiteSpecificRuntimeLimits(t *testing.T) {
	t.Parallel()

	spec, err := NewPanaceaChainSpec(
		"deep-a1b2c3",
		ImageRef{Repository: "panacea-e2e", Version: "current"},
		Topology{
			Validators:         4,
			FullNodes:          1,
			TimeoutCommit:      "1s",
			QueryGasLimit:      1_000_000,
			SnapshotInterval:   5,
			SnapshotKeepRecent: 2,
			EnableTelemetry:    true,
		},
	)
	require.NoError(t, err)
	cfg, err := spec.Config(zap.NewNop())
	require.NoError(t, err)

	appOverrides := cfg.ConfigFileOverrides["config/app.toml"].(testutil.Toml)
	require.EqualValues(t, 1_000_000, appOverrides["query-gas-limit"])
	stateSync := appOverrides["state-sync"].(testutil.Toml)
	require.EqualValues(t, 5, stateSync["snapshot-interval"])
	require.EqualValues(t, 2, stateSync["snapshot-keep-recent"])
	telemetry := appOverrides["telemetry"].(testutil.Toml)
	require.Equal(t, true, telemetry["enabled"])
	require.EqualValues(t, 60, telemetry["prometheus-retention-time"])

	cometOverrides := cfg.ConfigFileOverrides["config/config.toml"].(testutil.Toml)
	consensus := cometOverrides["consensus"].(testutil.Toml)
	require.Equal(t, "1s", consensus["timeout_commit"])
}
