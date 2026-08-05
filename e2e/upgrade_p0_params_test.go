package e2e_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

type upgradeP0GenesisContractParams struct {
	StakingUnbondingTime          string `json:"staking_unbonding_time"`
	SlashingSignedBlocksWindow    int64  `json:"slashing_signed_blocks_window"`
	SlashingMinSignedPerWindow    string `json:"slashing_min_signed_per_window"`
	SlashingDowntimeJailDuration  string `json:"slashing_downtime_jail_duration"`
	SlashingSlashFractionDowntime string `json:"slashing_slash_fraction_downtime"`
}

type upgradeP0GenesisContractEvidence struct {
	Phase       string                         `json:"phase"`
	Height      int64                          `json:"height"`
	RecordedAt  time.Time                      `json:"recorded_at"`
	Params      upgradeP0GenesisContractParams `json:"params"`
	StakingRaw  json.RawMessage                `json:"staking_raw"`
	SlashingRaw json.RawMessage                `json:"slashing_raw"`
}

func captureUpgradeP0GenesisContract(
	ctx context.Context,
	network *harness.Network,
	phase string,
	expected harness.Config,
) (upgradeP0GenesisContractEvidence, error) {
	if network == nil || network.Chain == nil || len(network.Chain.FullNodes) == 0 {
		return upgradeP0GenesisContractEvidence{}, errors.New("P0 genesis contract requires a full node")
	}
	phase = strings.TrimSpace(phase)
	if phase == "" {
		return upgradeP0GenesisContractEvidence{}, errors.New("P0 genesis contract phase is required")
	}
	height, err := network.Chain.FullNodes[0].Height(ctx)
	if err != nil {
		return upgradeP0GenesisContractEvidence{}, err
	}
	stakingRaw, err := network.FullNodeCLIQuery(
		ctx,
		"upgrade-p0-params-"+phase+"-staking",
		"staking", "params",
	)
	if err != nil {
		return upgradeP0GenesisContractEvidence{}, err
	}
	slashingRaw, err := network.FullNodeCLIQuery(
		ctx,
		"upgrade-p0-params-"+phase+"-slashing",
		"slashing", "params",
	)
	if err != nil {
		return upgradeP0GenesisContractEvidence{}, err
	}
	params, raw, err := decodeUpgradeP0GenesisContract(stakingRaw, slashingRaw)
	if err != nil {
		return upgradeP0GenesisContractEvidence{}, err
	}
	if err := validateUpgradeP0GenesisContract(params, expected); err != nil {
		return upgradeP0GenesisContractEvidence{}, fmt.Errorf("%s P0 genesis contract: %w", phase, err)
	}
	raw.Phase = phase
	raw.Height = height
	raw.RecordedAt = time.Now().UTC()
	raw.Params = params
	if err := network.WriteArtifactJSON("upgrade/p0/genesis-contract-"+phase+".json", raw); err != nil {
		return upgradeP0GenesisContractEvidence{}, err
	}
	return raw, nil
}

func decodeUpgradeP0GenesisContract(
	stakingRaw []byte,
	slashingRaw []byte,
) (upgradeP0GenesisContractParams, upgradeP0GenesisContractEvidence, error) {
	type stakingParams struct {
		UnbondingTime string `json:"unbonding_time"`
	}
	var staking stakingParams
	if err := decodeUpgradeModuleParams(stakingRaw, &staking); err != nil {
		return upgradeP0GenesisContractParams{}, upgradeP0GenesisContractEvidence{}, fmt.Errorf("decode staking params: %w", err)
	}
	type slashingParams struct {
		SignedBlocksWindow    json.RawMessage `json:"signed_blocks_window"`
		MinSignedPerWindow    string          `json:"min_signed_per_window"`
		DowntimeJailDuration  string          `json:"downtime_jail_duration"`
		SlashFractionDowntime string          `json:"slash_fraction_downtime"`
	}
	var slashing slashingParams
	if err := decodeUpgradeModuleParams(slashingRaw, &slashing); err != nil {
		return upgradeP0GenesisContractParams{}, upgradeP0GenesisContractEvidence{}, fmt.Errorf("decode slashing params: %w", err)
	}
	windowText := strings.Trim(strings.TrimSpace(string(slashing.SignedBlocksWindow)), `"`)
	window, err := strconv.ParseInt(windowText, 10, 64)
	if err != nil || window <= 0 {
		return upgradeP0GenesisContractParams{}, upgradeP0GenesisContractEvidence{}, fmt.Errorf(
			"decode positive signed_blocks_window %q", windowText,
		)
	}
	params, err := canonicalUpgradeP0GenesisContractParams(upgradeP0GenesisContractParams{
		StakingUnbondingTime:          strings.TrimSpace(staking.UnbondingTime),
		SlashingSignedBlocksWindow:    window,
		SlashingMinSignedPerWindow:    strings.TrimSpace(slashing.MinSignedPerWindow),
		SlashingDowntimeJailDuration:  strings.TrimSpace(slashing.DowntimeJailDuration),
		SlashingSlashFractionDowntime: strings.TrimSpace(slashing.SlashFractionDowntime),
	})
	if err != nil {
		return upgradeP0GenesisContractParams{}, upgradeP0GenesisContractEvidence{}, err
	}
	return params, upgradeP0GenesisContractEvidence{
		StakingRaw:  append(json.RawMessage(nil), stakingRaw...),
		SlashingRaw: append(json.RawMessage(nil), slashingRaw...),
	}, nil
}

// Cosmos SDK CLI versions differ here: v2.2.1 returns the params object
// directly, while other clients may wrap it in {"params": ...}.
func decodeUpgradeModuleParams(raw []byte, target any) error {
	var envelope struct {
		Params json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return err
	}
	payload := raw
	if len(envelope.Params) != 0 && string(envelope.Params) != "null" {
		payload = envelope.Params
	}
	return json.Unmarshal(payload, target)
}

func validateUpgradeP0GenesisContract(actual upgradeP0GenesisContractParams, expected harness.Config) error {
	actual, err := canonicalUpgradeP0GenesisContractParams(actual)
	if err != nil {
		return fmt.Errorf("actual P0 params: %w", err)
	}
	want, err := canonicalUpgradeP0GenesisContractParams(upgradeP0GenesisContractParams{
		StakingUnbondingTime:          expected.StakingUnbondingTime,
		SlashingSignedBlocksWindow:    expected.SlashingSignedBlocksWindow,
		SlashingMinSignedPerWindow:    expected.SlashingMinSignedPerWindow,
		SlashingDowntimeJailDuration:  expected.SlashingDowntimeJailDuration,
		SlashingSlashFractionDowntime: expected.SlashingSlashFractionDowntime,
	})
	if err != nil {
		return fmt.Errorf("expected P0 override: %w", err)
	}
	checks := []struct {
		name string
		got  string
		want string
	}{
		{"staking.unbonding_time", actual.StakingUnbondingTime, want.StakingUnbondingTime},
		{"slashing.signed_blocks_window", strconv.FormatInt(actual.SlashingSignedBlocksWindow, 10), strconv.FormatInt(want.SlashingSignedBlocksWindow, 10)},
		{"slashing.min_signed_per_window", actual.SlashingMinSignedPerWindow, want.SlashingMinSignedPerWindow},
		{"slashing.downtime_jail_duration", actual.SlashingDowntimeJailDuration, want.SlashingDowntimeJailDuration},
		{"slashing.slash_fraction_downtime", actual.SlashingSlashFractionDowntime, want.SlashingSlashFractionDowntime},
	}
	for _, check := range checks {
		if check.got != check.want {
			return fmt.Errorf("%s=%q, want %q", check.name, check.got, check.want)
		}
	}
	return nil
}

func canonicalUpgradeP0GenesisContractParams(
	params upgradeP0GenesisContractParams,
) (upgradeP0GenesisContractParams, error) {
	if params.SlashingSignedBlocksWindow <= 0 {
		return upgradeP0GenesisContractParams{}, fmt.Errorf(
			"slashing.signed_blocks_window must be positive, got %d",
			params.SlashingSignedBlocksWindow,
		)
	}
	canonicalDuration := func(name, value string) (string, error) {
		duration, err := time.ParseDuration(strings.TrimSpace(value))
		if err != nil || duration <= 0 {
			return "", fmt.Errorf("%s=%q must be a positive duration", name, value)
		}
		return duration.String(), nil
	}
	canonicalFraction := func(name, value string) (string, error) {
		fraction, err := sdkmath.LegacyNewDecFromStr(strings.TrimSpace(value))
		if err != nil || !fraction.IsPositive() || fraction.GT(sdkmath.LegacyOneDec()) {
			return "", fmt.Errorf("%s=%q must be in (0,1]", name, value)
		}
		return fraction.String(), nil
	}

	var err error
	params.StakingUnbondingTime, err = canonicalDuration("staking.unbonding_time", params.StakingUnbondingTime)
	if err != nil {
		return upgradeP0GenesisContractParams{}, err
	}
	params.SlashingDowntimeJailDuration, err = canonicalDuration(
		"slashing.downtime_jail_duration",
		params.SlashingDowntimeJailDuration,
	)
	if err != nil {
		return upgradeP0GenesisContractParams{}, err
	}
	params.SlashingMinSignedPerWindow, err = canonicalFraction(
		"slashing.min_signed_per_window",
		params.SlashingMinSignedPerWindow,
	)
	if err != nil {
		return upgradeP0GenesisContractParams{}, err
	}
	params.SlashingSlashFractionDowntime, err = canonicalFraction(
		"slashing.slash_fraction_downtime",
		params.SlashingSlashFractionDowntime,
	)
	if err != nil {
		return upgradeP0GenesisContractParams{}, err
	}
	return params, nil
}
