package e2e_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestUpgradeSystemModuleExportDoesNotPerturbP0SlashingTarget(t *testing.T) {
	t.Parallel()

	require.NotEqual(t, upgradeP0SlashingValidatorIndex, upgradeSystemModuleExportValidatorIndex)
	require.NotEqual(t, upgradeP0SlashingValidatorIndex, upgradeP0InvariantValidatorIndex)
	require.NotEqual(t, upgradeSystemModuleExportValidatorIndex, upgradeP0InvariantValidatorIndex)
}

func TestDecodeUpgradeSystemModuleExportRequiresAndCanonicalizesEveryModule(t *testing.T) {
	t.Parallel()

	exported := []byte(`{
		"initial_height":"42",
		"app_state":{
			"bank":{"supply":[{"denom":"umed","amount":"11"}]},
			"burn":{},
			"capability":{"owners":[],"index":"7"},
			"consensus":{"params":{"block":{"max_bytes":"1"}}},
			"crisis":{"constant_fee":{"amount":"100","denom":"umed"}},
			"mint":{"params":{"mint_denom":"umed"},"minter":{"inflation":"0.1"}},
			"params":{"params":[]}
		}
	}`)

	got, err := decodeUpgradeSystemModuleExport(exported)
	require.NoError(t, err)
	require.Equal(t, int64(41), got.Height)
	require.Equal(t, []string{"bank", "burn", "capability", "consensus", "crisis", "mint"}, got.ModuleNames)
	require.JSONEq(t, `{"index":"7","owners":[]}`, string(got.Modules["capability"]))
	require.Len(t, got.ModuleDigests, len(got.ModuleNames))
	for _, name := range got.ModuleNames {
		require.NotEmpty(t, got.ModuleDigests[name], name)
	}
}

func TestDecodeUpgradeSystemModuleExportReadsCurrentTopLevelConsensus(t *testing.T) {
	t.Parallel()

	exported := []byte(`{
		"initial_height":"5",
		"consensus":{"params":{"block":{"max_bytes":"2"}},"validators":[]},
		"app_state":{
			"bank":{},"burn":{},"capability":{},"crisis":{},"mint":{}
		}
	}`)
	got, err := decodeUpgradeSystemModuleExport(exported)
	require.NoError(t, err)
	require.JSONEq(t, `{"block":{"max_bytes":"2"}}`, string(got.Modules["consensus"]))
}

func TestDecodeUpgradeSystemModuleExportRejectsMissingOrMalformedState(t *testing.T) {
	t.Parallel()

	_, err := decodeUpgradeSystemModuleExport([]byte(`{"initial_height":"1","app_state":{}}`))
	require.ErrorContains(t, err, "initial_height")

	_, err = decodeUpgradeSystemModuleExport([]byte(`{"initial_height":"2","app_state":{"bank":{}}}`))
	require.ErrorContains(t, err, "burn")

	_, err = decodeUpgradeSystemModuleExport([]byte(`not-json`))
	require.ErrorContains(t, err, "decode")
}

func TestDecodeUpgradeSystemMintRESTResponsesUsesVersionNeutralJSON(t *testing.T) {
	want := minttypes.Params{
		MintDenom:           upgradeSystemDenom,
		InflationRateChange: sdkmath.LegacyMustNewDecFromStr("0.03"),
		InflationMax:        sdkmath.LegacyMustNewDecFromStr("0.10"),
		InflationMin:        sdkmath.LegacyMustNewDecFromStr("0.07"),
		GoalBonded:          sdkmath.LegacyMustNewDecFromStr("0.67"),
		BlocksPerYear:       5256000,
	}
	fixtures := map[string]struct {
		params    string
		inflation string
		annual    string
	}{
		"v2.2.1 proto names and string uint64": {
			params:    `{"params":{"mint_denom":"umed","inflation_rate_change":"0.03","inflation_max":"0.10","inflation_min":"0.07","goal_bonded":"0.67","blocks_per_year":"5256000"}}`,
			inflation: `{"inflation":"0.070000000000000000"}`,
			annual:    `{"annual_provisions":"2.000000000000000000"}`,
		},
		"current JSON names and numeric uint64": {
			params:    `{"params":{"mintDenom":"umed","inflationRateChange":"0.03","inflationMax":"0.10","inflationMin":"0.07","goalBonded":"0.67","blocksPerYear":5256000}}`,
			inflation: `{"inflation":"0.070000000000000000"}`,
			annual:    `{"annualProvisions":"2.000000000000000000"}`,
		},
	}
	for name, fixture := range fixtures {
		fixture := fixture
		t.Run(name, func(t *testing.T) {
			state, err := decodeUpgradeSystemMintRESTResponses(
				[]byte(fixture.params),
				[]byte(fixture.inflation),
				[]byte(fixture.annual),
			)
			require.NoError(t, err)
			require.Equal(t, want, state.Params)
			require.NoError(t, state.Params.Validate())
			require.Equal(t, "0.070000000000000000", state.Inflation)
			require.Equal(t, "2.000000000000000000", state.AnnualProvisions)
		})
	}

	_, err := decodeUpgradeSystemMintRESTResponses(
		[]byte(`{"params":null}`),
		[]byte(`{"inflation":""}`),
		[]byte(`{"annual_provisions":"2"}`),
	)
	require.Error(t, err)

	for _, blocksPerYear := range []string{`"-1"`, `-1`, `"18446744073709551616"`} {
		_, err := decodeUpgradeSystemMintRESTResponses(
			[]byte(fmt.Sprintf(`{"params":{"mint_denom":"umed","inflation_rate_change":"0.03","inflation_max":"0.10","inflation_min":"0.07","goal_bonded":"0.67","blocks_per_year":%s}}`, blocksPerYear)),
			[]byte(`{"inflation":"0.07"}`),
			[]byte(`{"annual_provisions":"2"}`),
		)
		require.ErrorContains(t, err, "blocks_per_year")
	}
}

func TestUpgradeSystemCrossBoundaryComparisonNormalizesProtoIntegerEncoding(t *testing.T) {
	t.Parallel()

	legacyMint := json.RawMessage(`{
		"mint_denom":"umed",
		"inflation_rate_change":"0.030000000000000000",
		"inflation_max":"0.100000000000000000",
		"inflation_min":"0.070000000000000000",
		"goal_bonded":"0.670000000000000000",
		"blocks_per_year":"5256000"
	}`)
	currentMint := json.RawMessage(`{
		"mintDenom":"umed",
		"inflationRateChange":"0.03",
		"inflationMax":"0.10",
		"inflationMin":"0.07",
		"goalBonded":"0.67",
		"blocksPerYear":5256000
	}`)
	legacyMintParams, err := decodeUpgradeSystemMintParams(legacyMint)
	require.NoError(t, err)
	currentMintParams, err := decodeUpgradeSystemMintParams(currentMint)
	require.NoError(t, err)
	require.True(t, canonicalUpgradeSystemValuesEqual(legacyMintParams, currentMintParams))

	legacyConsensus := json.RawMessage(`{
		"block":{"max_bytes":"22020096","max_gas":"-1"},
		"evidence":{
			"max_age_num_blocks":"302400",
			"max_age_duration":"1814400000000000",
			"max_bytes":"1048576"
		},
		"validator":{"pub_key_types":["ed25519"]},
		"version":{"app":"0"}
	}`)
	currentConsensus := json.RawMessage(`{
		"block":{"max_bytes":22020096,"max_gas":-1},
		"evidence":{
			"max_age_num_blocks":302400,
			"max_age_duration":1814400000000000,
			"max_bytes":1048576
		},
		"validator":{"pub_key_types":["ed25519"]},
		"version":{},
		"abci":{}
	}`)
	equal, err := upgradeSystemConsensusStatesEqual(legacyConsensus, currentConsensus)
	require.NoError(t, err)
	require.True(t, equal)

	mutatedConsensus := json.RawMessage(`{
		"block":{"max_bytes":22020096,"max_gas":1},
		"evidence":{
			"max_age_num_blocks":302400,
			"max_age_duration":1814400000000000,
			"max_bytes":1048576
		},
		"validator":{"pub_key_types":["ed25519"]},
		"version":{}
	}`)
	equal, err = upgradeSystemConsensusStatesEqual(legacyConsensus, mutatedConsensus)
	require.NoError(t, err)
	require.False(t, equal, "a real consensus parameter mutation must not be normalized away")

	_, err = decodeUpgradeSystemConsensusState([]byte(`{
		"block":{"max_bytes":"1","max_gas":"-1"},
		"evidence":{"max_age_num_blocks":"1","max_age_duration":"1","max_bytes":"1"},
		"validator":{"pub_key_types":["ed25519"]},
		"version":{},
		"future_field":{}
	}`))
	require.ErrorContains(t, err, "unsupported field")
}

func TestPrepareUpgradeSystemGenesisValidationInputProjectsOnlyV221PreUpgradeSentinel(t *testing.T) {
	t.Parallel()
	contents := []byte(`{
		"chain_id":"panacea-test",
		"app_state":{"ibc":{"connection_genesis":{"connections":[
			{
				"client_id":"09-localhost",
				"counterparty":{
					"client_id":"09-localhost",
					"connection_id":"connection-localhost",
					"prefix":{"key_prefix":"aWJj"}
				},
				"delay_period":"0",
				"id":"connection-localhost",
				"state":"STATE_OPEN",
				"versions":[{
					"features":["ORDER_ORDERED","ORDER_UNORDERED"],
					"identifier":"1"
				}]
			},
			{
				"client_id":"07-tendermint-0",
				"counterparty":{"client_id":"07-tendermint-1","connection_id":"connection-1"},
				"delay_period":"0",
				"id":"connection-0",
				"state":"STATE_OPEN"
			}
		]}}}
	}`)
	sourceCopy := append([]byte(nil), contents...)

	validationInput, evidence, err := prepareUpgradeSystemGenesisValidationInput(
		"pre-upgrade-checkpoint",
		contents,
	)
	require.NoError(t, err)
	require.True(t, evidence.Projected)
	require.Equal(t, "connection-localhost", evidence.RemovedConnectionID)
	require.JSONEq(t, `{
		"client_id":"09-localhost",
		"counterparty":{
			"client_id":"09-localhost",
			"connection_id":"connection-localhost",
			"prefix":{"key_prefix":"aWJj"}
		},
		"delay_period":"0",
		"id":"connection-localhost",
		"state":"STATE_OPEN",
		"versions":[{"features":["ORDER_ORDERED","ORDER_UNORDERED"],"identifier":"1"}]
	}`, string(evidence.RemovedConnection))
	require.Equal(t, "github.com/cosmos/ibc-go/v7@v7.3.2", evidence.UpstreamVersion)
	require.Contains(t, evidence.Reason, "InitGenesis")
	require.NotEmpty(t, evidence.SourceDigest)
	require.NotEmpty(t, evidence.ValidationInputDigest)
	require.NotEqual(t, evidence.SourceDigest, evidence.ValidationInputDigest)

	var projected struct {
		AppState struct {
			IBC struct {
				ConnectionGenesis struct {
					Connections []struct {
						ID string `json:"id"`
					} `json:"connections"`
				} `json:"connection_genesis"`
			} `json:"ibc"`
		} `json:"app_state"`
	}
	require.NoError(t, json.Unmarshal(validationInput, &projected))
	require.Equal(t, []string{"connection-0"}, []string{projected.AppState.IBC.ConnectionGenesis.Connections[0].ID})
	require.Equal(t, sourceCopy, contents, "source export bytes must remain immutable")

	directInput, directEvidence, err := prepareUpgradeSystemGenesisValidationInput(
		"post-upgrade-checkpoint",
		contents,
	)
	require.NoError(t, err)
	require.False(t, directEvidence.Projected)
	require.Empty(t, directEvidence.RemovedConnection)
	require.Equal(t, directEvidence.SourceDigest, directEvidence.ValidationInputDigest)
	require.Equal(t, contents, directInput)
}

func TestPrepareUpgradeSystemGenesisValidationInputRejectsMalformedV221LocalhostConnection(t *testing.T) {
	t.Parallel()
	contents := []byte(`{
		"app_state":{"ibc":{"connection_genesis":{"connections":[{
			"client_id":"07-tendermint-0",
			"counterparty":{
				"client_id":"09-localhost",
				"connection_id":"connection-localhost",
				"prefix":{"key_prefix":"aWJj"}
			},
			"delay_period":"0",
			"id":"connection-localhost",
			"state":"STATE_OPEN",
			"versions":[{"features":["ORDER_ORDERED","ORDER_UNORDERED"],"identifier":"1"}]
		}]}}}
	}`)

	_, _, err := prepareUpgradeSystemGenesisValidationInput("pre-upgrade-checkpoint", contents)
	require.ErrorContains(t, err, "localhost connection client_id")
}

func TestValidateUpgradeSystemModuleCheckpointRequiresCrossBoundaryState(t *testing.T) {
	t.Parallel()

	checkpoint := validUpgradeSystemModuleCheckpoint(t)
	require.NoError(t, validateUpgradeSystemModuleCheckpoint(checkpoint))

	checkpoint.BurnAddressBalance = "1"
	require.ErrorContains(t, validateUpgradeSystemModuleCheckpoint(checkpoint), "burn-address")

	checkpoint = validUpgradeSystemModuleCheckpoint(t)
	checkpoint.Export.Modules["crisis"] = json.RawMessage(`{"constant_fee":{"denom":"umed","amount":"0"}}`)
	require.ErrorContains(t, validateUpgradeSystemModuleCheckpoint(checkpoint), "crisis constant fee")

	checkpoint = validUpgradeSystemModuleCheckpoint(t)
	checkpoint.LegacyParams = nil
	require.ErrorContains(t, validateUpgradeSystemModuleCheckpoint(checkpoint), "legacy params")

	checkpoint = validUpgradeSystemModuleCheckpoint(t)
	checkpoint.MintQueryHeight--
	require.ErrorContains(t, validateUpgradeSystemModuleCheckpoint(checkpoint), "checkpoint height")

	checkpoint = validUpgradeSystemModuleCheckpoint(t)
	checkpoint.MintQueryPaths[2] = "/cosmos/mint/v1beta1/not-annual-provisions"
	require.ErrorContains(t, validateUpgradeSystemModuleCheckpoint(checkpoint), "mint REST query paths")
}

func validUpgradeSystemModuleCheckpoint(t *testing.T) upgradeSystemModuleCheckpoint {
	t.Helper()
	params := minttypes.Params{
		MintDenom:           upgradeSystemDenom,
		InflationRateChange: sdkmath.LegacyMustNewDecFromStr("0.03"),
		InflationMax:        sdkmath.LegacyMustNewDecFromStr("0.10"),
		InflationMin:        sdkmath.LegacyMustNewDecFromStr("0.07"),
		GoalBonded:          sdkmath.LegacyMustNewDecFromStr("0.67"),
		BlocksPerYear:       5256000,
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)
	consensus := `{
		"block":{"max_bytes":"2","max_gas":"-1"},
		"evidence":{"max_age_num_blocks":"3","max_age_duration":"4","max_bytes":"5"},
		"validator":{"pub_key_types":["ed25519"]},
		"version":{"app":"0"}
	}`
	document := fmt.Sprintf(`{
		"initial_height":"10",
		"consensus":{"params":%s,"validators":[]},
		"app_state":{
			"bank":{"supply":[{"denom":"umed","amount":"100"}]},
			"burn":{},
			"capability":{"index":"1","owners":[]},
			"crisis":{"constant_fee":{"denom":"umed","amount":"1000000000000"}},
			"mint":{"minter":{"inflation":"0.070000000000000000","annual_provisions":"2.000000000000000000"},"params":%s}
		}
	}`, consensus, paramsJSON)
	exported, err := decodeUpgradeSystemModuleExport([]byte(document))
	require.NoError(t, err)
	return upgradeSystemModuleCheckpoint{
		Phase:      "pre-upgrade-checkpoint",
		RecordedAt: time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC),
		Height:     9,
		Observation: harness.UpgradeCheckpointObservation{
			ObservedAt:    time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC),
			Node:          "fullnode-0",
			QueryBoundary: harness.UpgradeCheckpointQueryBoundaryCometBFTRPC,
			Height:        9,
			BlockID:       "AABB",
			AppHash:       "CCDD",
		},
		MintParams:         params,
		Inflation:          "0.070000000000000000",
		AnnualProvisions:   "2.000000000000000000",
		ConsensusParams:    json.RawMessage(consensus),
		LegacyParams:       map[string]map[string]string{"crisis": {"ConstantFee": `{"denom":"umed","amount":"1000000000000"}`}},
		Supply:             "100",
		BurnAddressBalance: "0",
		Export:             exported,
		ExportDigest:       "ABC",
		GenesisValidation: upgradeSystemGenesisValidationEvidence{
			Projected:             true,
			SourceDigest:          "ABC",
			ValidationInputDigest: "DEF",
			RemovedConnectionID:   upgradeSystemLocalhostConnectionID,
			RemovedConnection:     json.RawMessage(`{"id":"connection-localhost"}`),
			Reason:                upgradeSystemLocalhostProjectionReason,
			UpstreamVersion:       upgradeSystemV221IBCGoVersion,
			PublicValidation: harness.GenesisValidationEvidence{
				CanonicalDigest: "DEF",
			},
		},
		ExportValidated:       true,
		HistoricalPublicQuery: true,
		MintQueryBoundary:     upgradeSystemMintRESTBoundary,
		MintQueryHeight:       9,
		MintQueryPaths:        upgradeSystemMintRESTQueryPaths(),
	}
}
