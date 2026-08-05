package e2e_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestUpgradeP0RedelegationQueryArgsUseCrossVersionSingularCommand(t *testing.T) {
	t.Parallel()

	got := upgradeP0RedelegationQueryArgs(upgradeP0StakingQueueFixture{
		RedelegateAddress:    "panacea1delegator",
		SourceValidator:      "panaceavaloper1source",
		DestinationValidator: "panaceavaloper1destination",
	})
	want := []string{
		"staking", "redelegation",
		"panacea1delegator",
		"panaceavaloper1source",
		"panaceavaloper1destination",
	}
	require.Equal(t, want, got)
}

func TestUpgradeP0RedelegationPhaseRequiresExactPresenceContract(t *testing.T) {
	t.Parallel()

	for _, phase := range []string{"before", "completed", "post-restart"} {
		requiresEntry, err := upgradeP0RedelegationPhaseRequiresEntry(phase)
		require.NoError(t, err)
		require.False(t, requiresEntry, phase)
	}
	for _, phase := range []string{"queued", "post-upgrade-pending"} {
		requiresEntry, err := upgradeP0RedelegationPhaseRequiresEntry(phase)
		require.NoError(t, err)
		require.True(t, requiresEntry, phase)
	}
	_, err := upgradeP0RedelegationPhaseRequiresEntry("unexpected")
	require.ErrorContains(t, err, "unsupported staking queue phase")
}

func TestDecodeUpgradeP0UnbondingEntriesAcceptsProtoJSONNullAsEmpty(t *testing.T) {
	t.Parallel()

	entries, err := decodeUpgradeP0UnbondingEntries(
		[]byte(`{"unbonding_responses":null,"pagination":{}}`),
		"panacea1delegator",
		"panaceavaloper1validator",
	)
	require.NoError(t, err)
	require.Empty(t, entries)

	_, err = decodeUpgradeP0UnbondingEntries(
		[]byte(`{"pagination":{}}`),
		"panacea1delegator",
		"panaceavaloper1validator",
	)
	require.ErrorContains(t, err, "no unbonding_responses")

	_, err = decodeUpgradeP0UnbondingEntries(
		[]byte(`{"unbonding_responses":{},"pagination":{}}`),
		"panacea1delegator",
		"panaceavaloper1validator",
	)
	require.ErrorContains(t, err, "malformed unbonding_responses")
}

func TestValidateUpgradeP0StakingQueueEvidenceRequiresPendingAcrossUpgradeAndExactlyOnceCompletion(t *testing.T) {
	t.Parallel()

	completion := time.Date(2026, 8, 5, 12, 3, 0, 0, time.UTC)
	valid := upgradeP0StakingQueueEvidence{
		UnbondAmount:              "100",
		RedelegateAmount:          "200",
		FeePerTx:                  "5",
		UnbondWithdrawnReward:     "3",
		RedelegateWithdrawnReward: "2",
		Before: upgradeP0StakingQueueCheckpoint{
			Phase: "before", Height: 90, RecordedAt: completion.Add(-3 * time.Minute),
			BankBalance: "1000", SourceDelegationBalance: "1000", DestinationDelegationBalance: "0",
			Pool: upgradeStakingPoolState{BondedTokens: "4000", NotBondedTokens: "0"},
		},
		Queued: upgradeP0StakingQueueCheckpoint{
			Phase: "queued", Height: 95, RecordedAt: completion.Add(-2 * time.Minute),
			BankBalance: "995", SourceDelegationBalance: "700", DestinationDelegationBalance: "200",
			Pool:      upgradeStakingPoolState{BondedTokens: "3900", NotBondedTokens: "100"},
			Unbonding: []upgradeP0QueueEntry{{CompletionTime: completion, InitialBalance: "100", Balance: "100"}},
			Redelegation: []upgradeP0QueueEntry{{
				CompletionTime: completion, InitialBalance: "200", Balance: "200", SharesDestination: "200.000000000000000000",
			}},
		},
		PostUpgradePending: upgradeP0StakingQueueCheckpoint{
			Phase: "post-upgrade-pending", Height: 104, RecordedAt: completion.Add(-time.Minute),
			BankBalance: "995", SourceDelegationBalance: "700", DestinationDelegationBalance: "200",
			Pool:      upgradeStakingPoolState{BondedTokens: "3900", NotBondedTokens: "100"},
			Unbonding: []upgradeP0QueueEntry{{CompletionTime: completion, InitialBalance: "100", Balance: "100"}},
			Redelegation: []upgradeP0QueueEntry{{
				CompletionTime: completion, InitialBalance: "200", Balance: "200", SharesDestination: "200.000000000000000000",
			}},
		},
		Completed: upgradeP0StakingQueueCheckpoint{
			Phase: "completed", Height: 190, RecordedAt: completion.Add(time.Second),
			BankBalance: "1095", SourceDelegationBalance: "700", DestinationDelegationBalance: "200",
			Pool: upgradeStakingPoolState{BondedTokens: "3900", NotBondedTokens: "0"},
		},
		PostRestart: upgradeP0StakingQueueCheckpoint{
			Phase: "post-restart", Height: 200, RecordedAt: completion.Add(time.Minute),
			BankBalance: "1095", SourceDelegationBalance: "700", DestinationDelegationBalance: "200",
			Pool: upgradeStakingPoolState{BondedTokens: "3900", NotBondedTokens: "0"},
		},
	}
	require.NoError(t, validateUpgradeP0StakingQueueEvidence(valid))

	completedBeforeUpgrade := valid
	completedBeforeUpgrade.PostUpgradePending.Unbonding = nil
	require.ErrorContains(t, validateUpgradeP0StakingQueueEvidence(completedBeforeUpgrade), "pending across the upgrade")

	doubleCompletion := valid
	doubleCompletion.PostRestart.BankBalance = "1195"
	require.ErrorContains(t, validateUpgradeP0StakingQueueEvidence(doubleCompletion), "post-restart")

	poolDrift := valid
	poolDrift.Completed.Pool.NotBondedTokens = "100"
	require.ErrorContains(t, validateUpgradeP0StakingQueueEvidence(poolDrift), "not-bonded pool")

	// The connected upgrade lane also slashes and later delegates through
	// unrelated validators. Those global bonded-pool changes must not make the
	// dedicated queue accounts look as though they completed twice.
	unrelatedPoolMutation := valid
	unrelatedPoolMutation.PostUpgradePending.Pool.BondedTokens = "8900"
	unrelatedPoolMutation.Completed.Pool.BondedTokens = "8900"
	unrelatedPoolMutation.PostRestart.Pool.BondedTokens = "9100"
	require.NoError(t, validateUpgradeP0StakingQueueEvidence(unrelatedPoolMutation))

	wrongRewardAccounting := valid
	wrongRewardAccounting.RedelegateWithdrawnReward = "1"
	require.ErrorContains(t, validateUpgradeP0StakingQueueEvidence(wrongRewardAccounting), "after fees and withdrawn rewards")
}

func TestDecodeUpgradeP0WithdrawnRewardRequiresExactEventIdentityAndDenom(t *testing.T) {
	t.Parallel()

	result := &harness.TxResult{Events: []harness.TxEvent{{
		Type: "withdraw_rewards",
		Attributes: []harness.TxEventAttribute{
			{Key: "amount", Value: "3067umed"},
			{Key: "validator", Value: "panaceavaloper1source"},
			{Key: "delegator", Value: "panacea1delegator"},
		},
	}}}
	reward, err := decodeUpgradeP0WithdrawnReward(
		result,
		"panacea1delegator",
		"panaceavaloper1source",
	)
	require.NoError(t, err)
	require.Equal(t, "3067", reward)

	wrongDenom := *result
	wrongDenom.Events = append([]harness.TxEvent(nil), result.Events...)
	wrongDenom.Events[0].Attributes = append([]harness.TxEventAttribute(nil), result.Events[0].Attributes...)
	wrongDenom.Events[0].Attributes[0].Value = "3067uatom"
	_, err = decodeUpgradeP0WithdrawnReward(&wrongDenom, "panacea1delegator", "panaceavaloper1source")
	require.ErrorContains(t, err, "denom")

	_, err = decodeUpgradeP0WithdrawnReward(result, "panacea1other", "panaceavaloper1source")
	require.ErrorContains(t, err, "exactly one")

	duplicate := *result
	duplicate.Events = append(append([]harness.TxEvent(nil), result.Events...), result.Events[0])
	_, err = decodeUpgradeP0WithdrawnReward(&duplicate, "panacea1delegator", "panaceavaloper1source")
	require.ErrorContains(t, err, "exactly one")
}

func TestValidateUpgradeP0StakingPreservationAccountsForExactConcurrentSlash(t *testing.T) {
	t.Parallel()

	before := upgradeStakingCheckpoint{
		Phase: "pre-upgrade-checkpoint", Height: 40,
		DelegatorBankBalance: "900", RecipientBankBalance: "0",
		State: upgradeStakingState{
			Delegation: upgradeDelegationState{
				DelegatorAddress: "panacea1delegator", ValidatorOperator: "panaceavaloper1validator",
				Shares: "100.000000000000000000", Balance: upgradeStakingCoin{Denom: "umed", Amount: "100"},
			},
			Validator: upgradeValidatorState{
				OperatorAddress: "panaceavaloper1validator",
				ConsensusPubKey: json.RawMessage(`{"@type":"/cosmos.crypto.ed25519.PubKey","key":"AQID"}`),
				Status:          "BOND_STATUS_BONDED", Tokens: "5000100", DelegatorShares: "5000100.000000000000000000",
			},
			Pool:                upgradeStakingPoolState{BondedTokens: "20000100", NotBondedTokens: "107"},
			DelegatorRewards:    upgradeStakingDecCoins{{Denom: "umed", Amount: "12.5"}},
			OutstandingRewards:  upgradeStakingDecCoins{{Denom: "umed", Amount: "55.25"}},
			ValidatorCommission: upgradeStakingDecCoins{{Denom: "umed", Amount: "4.125"}},
			SigningInfo: upgradeSigningInfoState{
				Address: "panaceavalcons1validator", StartHeight: 1, IndexOffset: 88,
				JailedUntil: "1970-01-01T00:00:00Z",
			},
		},
	}
	after := before
	after.Phase = "post-upgrade-preservation"
	after.Height = 50
	after.State.Validator.ConsensusPubKey = json.RawMessage(`{"type":"/cosmos.crypto.ed25519.PubKey","value":"AQID"}`)
	after.State.Pool.BondedTokens = "19999100"
	after.State.Pool.NotBondedTokens = "7"
	after.State.SigningInfo.IndexOffset = 98
	after.State.DelegatorRewards[0].Amount = "14"
	after.State.OutstandingRewards[0].Amount = "60"
	after.State.ValidatorCommission[0].Amount = "5"
	slashing := upgradeP0SlashingEvidence{
		Before: upgradeP0SlashingCheckpoint{Validator: upgradeValidatorState{Tokens: "100000"}},
		Jailed: upgradeP0SlashingCheckpoint{Validator: upgradeValidatorState{Tokens: "99000"}},
	}
	queues := upgradeP0StakingQueueEvidence{
		UnbondAmount:       "100",
		PostUpgradePending: upgradeP0StakingQueueCheckpoint{Height: 45},
		Completed:          upgradeP0StakingQueueCheckpoint{Height: 49},
	}

	require.NoError(t, validateUpgradeP0StakingPreservationWithSlashing(before, after, slashing, queues))
	wrongPool := after
	wrongPool.State.Pool.BondedTokens = "19999101"
	require.ErrorContains(t, validateUpgradeP0StakingPreservationWithSlashing(before, wrongPool, slashing, queues), "slash delta")
}

func TestValidateUpgradeSystemModulePreservationWithSlashingAccountsForExactSupplyBurn(t *testing.T) {
	t.Parallel()

	before := validUpgradeSystemModuleCheckpoint(t)
	after := before
	after.Height++
	after.RecordedAt = after.RecordedAt.Add(time.Second)
	after.Observation.Height = after.Height
	after.Observation.ObservedAt = after.RecordedAt
	after.Export.Height = after.Height
	after.MintQueryHeight = after.Height
	after.Supply = "70"
	after.Export.Modules = make(map[string]json.RawMessage, len(before.Export.Modules))
	for name, value := range before.Export.Modules {
		after.Export.Modules[name] = append(json.RawMessage(nil), value...)
	}
	after.Export.Modules["bank"] = json.RawMessage(`{"supply":[{"denom":"umed","amount":"70"}]}`)

	slashing := upgradeP0SlashingEvidence{
		Before: upgradeP0SlashingCheckpoint{Validator: upgradeValidatorState{Tokens: "100"}},
		Jailed: upgradeP0SlashingCheckpoint{Validator: upgradeValidatorState{Tokens: "60"}},
	}
	accounting, err := validateUpgradeSystemModulePreservationWithSlashing(before, after, slashing)
	require.NoError(t, err)
	require.Equal(t, upgradeP0SystemSupplyAccounting{
		BeforeSupply: "100",
		AfterSupply:  "70",
		SlashBurn:    "40",
		MintAccrual:  "10",
	}, accounting)
	require.ErrorContains(t, validateUpgradeSystemModulePreservation(before, after), "supply decreased")

	after.Supply = "50"
	after.Export.Modules["bank"] = json.RawMessage(`{"supply":[{"denom":"umed","amount":"50"}]}`)
	_, err = validateUpgradeSystemModulePreservationWithSlashing(before, after, slashing)
	require.ErrorContains(t, err, "exceeds observed slash burn")
}

func TestValidateUpgradeP0SlashingEvidenceRequiresCrossBoundaryJailSlashAndSignedRejoin(t *testing.T) {
	t.Parallel()

	jailedUntil := time.Date(2026, 8, 5, 12, 0, 45, 0, time.UTC)
	valid := upgradeP0SlashingEvidence{
		UpgradeHeight:        100,
		StoppedAt:            100,
		OutageBlocksObserved: upgradeP0SlashingMinimumMisses,
		MissedBlocksObserved: 7,
		Before: upgradeP0SlashingCheckpoint{
			Height: 99, RecordedAt: jailedUntil.Add(-time.Minute), ValidatorPower: 100,
			Validator:   upgradeValidatorState{OperatorAddress: "panaceavaloper1target", Tokens: "10000", Jailed: false},
			SigningInfo: upgradeSigningInfoState{Address: "panaceavalcons1target", IndexOffset: 90, MissedBlocksCounter: 1, Tombstoned: false},
		},
		Jailed: upgradeP0SlashingCheckpoint{
			Height: 100 + upgradeP0SlashingMinimumMisses, RecordedAt: jailedUntil.Add(-30 * time.Second), ValidatorPower: 0,
			Validator:   upgradeValidatorState{OperatorAddress: "panaceavaloper1target", Tokens: "9900", Jailed: true},
			SigningInfo: upgradeSigningInfoState{Address: "panaceavalcons1target", IndexOffset: 99, MissedBlocksCounter: 0, JailedUntil: jailedUntil.Format(time.RFC3339Nano), Tombstoned: false},
		},
		EarlyUnjail:  harness.TxResult{Height: "122", TxHash: "EARLY", Codespace: "slashing", Code: 4},
		UnjailTxHash: "UNJAIL",
		Unjailed: upgradeP0SlashingCheckpoint{
			Height: 140, RecordedAt: jailedUntil.Add(time.Second), ValidatorPower: 0,
			Validator:   upgradeValidatorState{OperatorAddress: "panaceavaloper1target", Tokens: "9900", Jailed: false},
			SigningInfo: upgradeSigningInfoState{Address: "panaceavalcons1target", IndexOffset: 99, MissedBlocksCounter: 0, JailedUntil: jailedUntil.Format(time.RFC3339Nano), Tombstoned: false},
		},
		Rejoined: upgradeP0SlashingCheckpoint{
			Height: 143, RecordedAt: jailedUntil.Add(4 * time.Second), ValidatorPower: 99,
			Validator:   upgradeValidatorState{OperatorAddress: "panaceavaloper1target", Tokens: "9900", Jailed: false},
			SigningInfo: upgradeSigningInfoState{Address: "panaceavalcons1target", IndexOffset: 102, MissedBlocksCounter: 0, JailedUntil: jailedUntil.Format(time.RFC3339Nano), Tombstoned: false},
		},
		SignedCommitHeight: 142,
		PostRestart: upgradeP0SlashingCheckpoint{
			Height: 150, RecordedAt: jailedUntil.Add(time.Minute), ValidatorPower: 99,
			Validator:   upgradeValidatorState{OperatorAddress: "panaceavaloper1target", Tokens: "9900", Jailed: false},
			SigningInfo: upgradeSigningInfoState{Address: "panaceavalcons1target", IndexOffset: 109, MissedBlocksCounter: 0, JailedUntil: jailedUntil.Format(time.RFC3339Nano), Tombstoned: false},
		},
	}
	require.NoError(t, validateUpgradeP0SlashingEvidence(valid))

	noCrossBoundary := valid
	noCrossBoundary.Jailed.Height = 100
	require.ErrorContains(t, validateUpgradeP0SlashingEvidence(noCrossBoundary), "after upgrade height")

	noSlash := valid
	noSlash.Jailed.Validator.Tokens = valid.Before.Validator.Tokens
	require.ErrorContains(t, validateUpgradeP0SlashingEvidence(noSlash), "slashed")

	tombstoned := valid
	tombstoned.PostRestart.SigningInfo.Tombstoned = true
	require.ErrorContains(t, validateUpgradeP0SlashingEvidence(tombstoned), "tombstoned")

	wrongEarlyFailure := valid
	wrongEarlyFailure.EarlyUnjail.Code = 3
	require.ErrorContains(t, validateUpgradeP0SlashingEvidence(wrongEarlyFailure), "early unjail")

	shortOutage := valid
	shortOutage.OutageBlocksObserved = upgradeP0SlashingMinimumMisses - 1
	require.ErrorContains(t, validateUpgradeP0SlashingEvidence(shortOutage), "minimum outage")
}
