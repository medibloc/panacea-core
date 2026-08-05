package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestDecodeAndValidateUpgradeP0GenesisContract(t *testing.T) {
	t.Parallel()

	staking, slashing, err := decodeUpgradeP0GenesisContract(
		[]byte(`{"params":{"unbonding_time":"600s"}}`),
		[]byte(`{"params":{"signed_blocks_window":"100","min_signed_per_window":"0.800000000000000000","downtime_jail_duration":"300s","slash_fraction_downtime":"0.010000000000000000"}}`),
	)
	require.NoError(t, err)
	require.Equal(t, upgradeP0GenesisContractParams{
		StakingUnbondingTime:          "10m0s",
		SlashingSignedBlocksWindow:    100,
		SlashingMinSignedPerWindow:    "0.800000000000000000",
		SlashingDowntimeJailDuration:  "5m0s",
		SlashingSlashFractionDowntime: "0.010000000000000000",
	}, staking)
	require.JSONEq(t, `{"params":{"unbonding_time":"600s"}}`, string(slashing.StakingRaw))
	require.JSONEq(t, `{"params":{"signed_blocks_window":"100","min_signed_per_window":"0.800000000000000000","downtime_jail_duration":"300s","slash_fraction_downtime":"0.010000000000000000"}}`, string(slashing.SlashingRaw))

	expected := harness.Config{
		StakingUnbondingTime:          "600s",
		SlashingSignedBlocksWindow:    100,
		SlashingMinSignedPerWindow:    "0.800000000000000000",
		SlashingDowntimeJailDuration:  "300s",
		SlashingSlashFractionDowntime: "0.010000000000000000",
	}
	require.NoError(t, validateUpgradeP0GenesisContract(staking, expected))

	mismatch := expected
	mismatch.SlashingSignedBlocksWindow = 101
	require.ErrorContains(t, validateUpgradeP0GenesisContract(staking, mismatch), "signed_blocks_window")
}

func TestDecodeUpgradeP0GenesisContractAcceptsNumericWindowAndRejectsMissingFields(t *testing.T) {
	t.Parallel()

	params, _, err := decodeUpgradeP0GenesisContract(
		[]byte(`{"params":{"unbonding_time":"180s"}}`),
		[]byte(`{"params":{"signed_blocks_window":30,"min_signed_per_window":"0.8","downtime_jail_duration":"120s","slash_fraction_downtime":"0.01"}}`),
	)
	require.NoError(t, err)
	require.Equal(t, int64(30), params.SlashingSignedBlocksWindow)

	_, _, err = decodeUpgradeP0GenesisContract(
		[]byte(`{"params":{}}`),
		[]byte(`{"params":{"signed_blocks_window":"30"}}`),
	)
	require.Error(t, err)
}

func TestDecodeUpgradeP0GenesisContractAcceptsV221FlatCLIResponse(t *testing.T) {
	t.Parallel()

	params, _, err := decodeUpgradeP0GenesisContract(
		[]byte(`{"unbonding_time":"600s","max_validators":50,"bond_denom":"umed"}`),
		[]byte(`{"signed_blocks_window":"100","min_signed_per_window":"0.800000000000000000","downtime_jail_duration":"300s","slash_fraction_double_sign":"0.050000000000000000","slash_fraction_downtime":"0.010000000000000000"}`),
	)
	require.NoError(t, err)
	require.Equal(t, upgradeP0GenesisContractParams{
		StakingUnbondingTime:          "10m0s",
		SlashingSignedBlocksWindow:    100,
		SlashingMinSignedPerWindow:    "0.800000000000000000",
		SlashingDowntimeJailDuration:  "5m0s",
		SlashingSlashFractionDowntime: "0.010000000000000000",
	}, params)
}

func TestDecodeUpgradeP0GenesisContractCanonicalizesEquivalentCurrentDurationsAndDecimals(t *testing.T) {
	t.Parallel()

	legacy, _, err := decodeUpgradeP0GenesisContract(
		[]byte(`{"unbonding_time":"600s"}`),
		[]byte(`{"signed_blocks_window":"100","min_signed_per_window":"0.800000000000000000","downtime_jail_duration":"300s","slash_fraction_downtime":"0.010000000000000000"}`),
	)
	require.NoError(t, err)
	current, _, err := decodeUpgradeP0GenesisContract(
		[]byte(`{"params":{"unbonding_time":"10m0s"}}`),
		[]byte(`{"params":{"signed_blocks_window":100,"min_signed_per_window":"0.8","downtime_jail_duration":"5m0s","slash_fraction_downtime":"0.01"}}`),
	)
	require.NoError(t, err)
	require.Equal(t, legacy, current)
	require.Equal(t, "10m0s", current.StakingUnbondingTime)
	require.Equal(t, "5m0s", current.SlashingDowntimeJailDuration)
	require.Equal(t, "0.800000000000000000", current.SlashingMinSignedPerWindow)
	require.Equal(t, "0.010000000000000000", current.SlashingSlashFractionDowntime)

	require.NoError(t, validateUpgradeP0GenesisContract(current, harness.Config{
		StakingUnbondingTime:          "600s",
		SlashingSignedBlocksWindow:    100,
		SlashingMinSignedPerWindow:    "0.800000000000000000",
		SlashingDowntimeJailDuration:  "300s",
		SlashingSlashFractionDowntime: "0.010000000000000000",
	}))
}
