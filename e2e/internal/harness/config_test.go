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

func TestPanaceaGenesisModifierAppliesExplicitStakingAndSlashingOverrides(t *testing.T) {
	t.Parallel()

	spec, err := NewPanaceaChainSpec(
		"fault-genesis-a1b2c3",
		ImageRef{Repository: "panacea-e2e", Version: "current"},
		Topology{
			Validators:                    4,
			FullNodes:                     1,
			StakingUnbondingTime:          "8s",
			SlashingSignedBlocksWindow:    10,
			SlashingMinSignedPerWindow:    "0.500000000000000000",
			SlashingDowntimeJailDuration:  "7s",
			SlashingSlashFractionDowntime: "0.010000000000000000",
		},
	)
	require.NoError(t, err)
	cfg, err := spec.Config(zap.NewNop())
	require.NoError(t, err)
	require.NotNil(t, cfg.ModifyGenesis)

	input := []byte(`{
		"app_state": {
			"gov": {"params": {
				"voting_period": "259200s",
				"max_deposit_period": "1209600s",
				"min_deposit": [{"denom":"umed","amount":"100000000000"}]
			}},
			"staking": {"params": {
				"unbonding_time": "1814400s",
				"max_validators": 100
			}},
			"slashing": {"params": {
				"signed_blocks_window": "100",
				"min_signed_per_window": "0.500000000000000000",
				"downtime_jail_duration": "600s",
				"slash_fraction_downtime": "0.000100000000000000",
				"slash_fraction_double_sign": "0.050000000000000000"
			}}
		}
	}`)
	modified, err := cfg.ModifyGenesis(*cfg, input)
	require.NoError(t, err)

	var output map[string]any
	require.NoError(t, json.Unmarshal(modified, &output))
	appState := output["app_state"].(map[string]any)
	govParams := appState["gov"].(map[string]any)["params"].(map[string]any)
	stakingParams := appState["staking"].(map[string]any)["params"].(map[string]any)
	slashingParams := appState["slashing"].(map[string]any)["params"].(map[string]any)

	require.Equal(t, "20s", govParams["voting_period"])
	require.Equal(t, "8s", stakingParams["unbonding_time"])
	require.EqualValues(t, 100, stakingParams["max_validators"], "unrelated staking genesis must be preserved")
	require.Equal(t, "10", slashingParams["signed_blocks_window"])
	require.Equal(t, "0.500000000000000000", slashingParams["min_signed_per_window"])
	require.Equal(t, "7s", slashingParams["downtime_jail_duration"])
	require.Equal(t, "0.010000000000000000", slashingParams["slash_fraction_downtime"])
	require.Equal(t, "0.050000000000000000", slashingParams["slash_fraction_double_sign"], "unrelated slashing genesis must be preserved")
}

func TestTopologyFromConfigCarriesTestGenesisOverrides(t *testing.T) {
	t.Parallel()

	topology := topologyFromConfig(Config{
		NumValidators:                 4,
		NumFullNodes:                  1,
		StakingUnbondingTime:          "8s",
		SlashingSignedBlocksWindow:    10,
		SlashingMinSignedPerWindow:    "0.500000000000000000",
		SlashingDowntimeJailDuration:  "7s",
		SlashingSlashFractionDowntime: "0.010000000000000000",
	})

	require.Equal(t, 4, topology.Validators)
	require.Equal(t, 1, topology.FullNodes)
	require.Equal(t, "8s", topology.StakingUnbondingTime)
	require.Equal(t, int64(10), topology.SlashingSignedBlocksWindow)
	require.Equal(t, "0.500000000000000000", topology.SlashingMinSignedPerWindow)
	require.Equal(t, "7s", topology.SlashingDowntimeJailDuration)
	require.Equal(t, "0.010000000000000000", topology.SlashingSlashFractionDowntime)
}

func TestPanaceaGenesisModifierLeavesStakingAndSlashingDefaultsWhenOverridesAreZero(t *testing.T) {
	t.Parallel()

	spec, err := NewPanaceaChainSpec(
		"default-genesis-a1b2c3",
		ImageRef{Repository: "panacea-e2e", Version: "current"},
		Topology{Validators: 1},
	)
	require.NoError(t, err)
	cfg, err := spec.Config(zap.NewNop())
	require.NoError(t, err)

	input := []byte(`{
		"app_state": {
			"gov": {"params": {
				"voting_period": "259200s",
				"max_deposit_period": "1209600s",
				"min_deposit": [{"denom":"umed","amount":"100000000000"}]
			}},
			"staking": {"params": {"unbonding_time": "1814400s"}},
			"slashing": {"params": {
				"signed_blocks_window": "100",
				"min_signed_per_window": "0.500000000000000000",
				"downtime_jail_duration": "600s",
				"slash_fraction_downtime": "0.000100000000000000"
			}}
		}
	}`)
	modified, err := cfg.ModifyGenesis(*cfg, input)
	require.NoError(t, err)

	var output map[string]any
	require.NoError(t, json.Unmarshal(modified, &output))
	appState := output["app_state"].(map[string]any)
	stakingParams := appState["staking"].(map[string]any)["params"].(map[string]any)
	slashingParams := appState["slashing"].(map[string]any)["params"].(map[string]any)
	require.Equal(t, "1814400s", stakingParams["unbonding_time"])
	require.Equal(t, "100", slashingParams["signed_blocks_window"])
	require.Equal(t, "0.500000000000000000", slashingParams["min_signed_per_window"])
	require.Equal(t, "600s", slashingParams["downtime_jail_duration"])
	require.Equal(t, "0.000100000000000000", slashingParams["slash_fraction_downtime"])
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
