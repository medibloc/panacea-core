package e2e_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestUpgradeMatrixPhaseValidatorsAcceptCanonicalV221Phase(t *testing.T) {
	validators := map[string]func(string) error{
		"staking":       validateUpgradeStakingPhase,
		"group-vesting": validateUpgradeGroupVestingPhase,
	}
	for name, validate := range validators {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, validate("v2.2.1-preparation"))
			require.Error(t, validate("../outside"))
			require.Error(t, validate(" post-upgrade"))
		})
	}
}

func TestDecodeUpgradeStakingQueriesAcceptsV221CLIShapes(t *testing.T) {
	fixture := upgradeStakingFixture{
		DelegatorAddress:       "panacea1delegator",
		RewardRecipientAddress: "panacea1recipient",
		ValidatorOperator:      "panaceavaloper1validator",
	}
	state, err := decodeUpgradeStakingQueries(upgradeStakingQueryResponses{
		Delegation:         []byte(`{"delegation":{"delegator_address":"panacea1delegator","validator_address":"panaceavaloper1validator","shares":"100.000000000000000000"},"balance":{"denom":"umed","amount":"100"}}`),
		Validator:          []byte(`{"operator_address":"panaceavaloper1validator","consensus_pubkey":{"@type":"/cosmos.crypto.ed25519.PubKey","key":"AQID"},"jailed":false,"status":"BOND_STATUS_BONDED","tokens":"5000100","delegator_shares":"5000100.000000000000000000"}`),
		Pool:               []byte(`{"not_bonded_tokens":"7","bonded_tokens":"20000100"}`),
		DelegatorRewards:   []byte(`{"rewards":[{"validator_address":"panaceavaloper1validator","reward":[{"denom":"umed","amount":"12.500000000000000000"}]}],"total":[{"denom":"umed","amount":"12.500000000000000000"}]}`),
		OutstandingRewards: []byte(`{"rewards":[{"denom":"umed","amount":"55.250000000000000000"}]}`),
		Commission:         []byte(`{"commission":[{"denom":"umed","amount":"4.125000000000000000"}]}`),
		SigningInfo:        []byte(`{"address":"panaceavalcons1validator","start_height":"1","index_offset":"88","jailed_until":"1970-01-01T00:00:00Z","tombstoned":false,"missed_blocks_counter":"0"}`),
	}, fixture)
	require.NoError(t, err)
	require.Equal(t, "100", state.Delegation.Balance.Amount)
	require.Equal(t, "100.000000000000000000", state.Delegation.Shares)
	require.Equal(t, "5000100", state.Validator.Tokens)
	require.Equal(t, "20000100", state.Pool.BondedTokens)
	require.Equal(t, "12.500000000000000000", state.DelegatorRewards.AmountOf("umed"))
	require.Equal(t, "55.250000000000000000", state.OutstandingRewards.AmountOf("umed"))
	require.Equal(t, "4.125000000000000000", state.ValidatorCommission.AmountOf("umed"))
	require.Equal(t, int64(88), state.SigningInfo.IndexOffset)
	require.False(t, state.SigningInfo.Tombstoned)
}

func TestDecodeUpgradeStakingQueriesAcceptsOmittedFalseProtoJSONBooleans(t *testing.T) {
	fixture := upgradeStakingFixture{
		DelegatorAddress:       "panacea1delegator",
		RewardRecipientAddress: "panacea1recipient",
		ValidatorOperator:      "panaceavaloper1validator",
	}
	responses := upgradeStakingQueryResponses{
		Delegation:         []byte(`{"delegation_response":{"delegation":{"delegator_address":"panacea1delegator","validator_address":"panaceavaloper1validator","shares":"100.000000000000000000"},"balance":{"denom":"umed","amount":"100"}}}`),
		Validator:          []byte(`{"validator":{"operator_address":"panaceavaloper1validator","consensus_pubkey":{"type":"/cosmos.crypto.ed25519.PubKey","value":"AQID"},"status":"BOND_STATUS_BONDED","tokens":"5000100","delegator_shares":"5000100.000000000000000000"}}`),
		Pool:               []byte(`{"pool":{"not_bonded_tokens":"7","bonded_tokens":"20000100"}}`),
		DelegatorRewards:   []byte(`{"rewards":[{"validator_address":"panaceavaloper1validator","reward":["12.500000000000000000umed"]}],"total":["12.500000000000000000umed"]}`),
		OutstandingRewards: []byte(`{"rewards":{"rewards":["55.250000000000000000umed"]}}`),
		Commission:         []byte(`{"commission":{"commission":["4.125000000000000000umed"]}}`),
		SigningInfo:        []byte(`{"val_signing_info":{"address":"panaceavalcons1validator","index_offset":"88","jailed_until":"1970-01-01T00:00:00Z"}}`),
	}

	state, err := decodeUpgradeStakingQueries(responses, fixture)
	require.NoError(t, err)
	require.False(t, state.Validator.Jailed)
	require.JSONEq(t, `{"@type":"/cosmos.crypto.ed25519.PubKey","key":"AQID"}`, string(state.Validator.ConsensusPubKey))
	require.False(t, state.SigningInfo.Tombstoned)
	require.Zero(t, state.SigningInfo.StartHeight)
	require.Equal(t, int64(88), state.SigningInfo.IndexOffset)
	require.Zero(t, state.SigningInfo.MissedBlocksCounter)

	responses.Validator = []byte(`{"validator":{"operator_address":"panaceavaloper1validator","consensus_pubkey":{"type":"/cosmos.crypto.ed25519.PubKey","value":"AQID"},"jailed":"false","status":"BOND_STATUS_BONDED","tokens":"5000100","delegator_shares":"5000100.000000000000000000"}}`)
	_, err = decodeUpgradeStakingQueries(responses, fixture)
	require.ErrorContains(t, err, "jailed is not a boolean")

	responses.Validator = []byte(`{"validator":{"operator_address":"panaceavaloper1validator","consensus_pubkey":{"type":"/cosmos.crypto.ed25519.PubKey","value":"AQID"},"status":"BOND_STATUS_BONDED","tokens":"5000100","delegator_shares":"5000100.000000000000000000"}}`)
	responses.SigningInfo = []byte(`{"val_signing_info":{"address":"panaceavalcons1validator","index_offset":"88","jailed_until":"1970-01-01T00:00:00Z","missed_blocks_counter":"not-an-int"}}`)
	_, err = decodeUpgradeStakingQueries(responses, fixture)
	require.ErrorContains(t, err, "missed_blocks_counter is not a valid int64")
}

func TestDecodeUpgradeDecCoinsAcceptsLegacyObjectsAndCurrentCompactStrings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		raw   string
		field string
		want  upgradeStakingDecCoins
	}{
		{
			name:  "v2.2.1 object coins",
			raw:   `{"rewards":[{"denom":"umed","amount":"55.250000000000000000"}]}`,
			field: "rewards",
			want:  upgradeStakingDecCoins{{Denom: "umed", Amount: "55.250000000000000000"}},
		},
		{
			name:  "v2.3.0 nested compact coins",
			raw:   `{"rewards":{"rewards":["55.250000000000000000umed"]}}`,
			field: "rewards",
			want:  upgradeStakingDecCoins{{Denom: "umed", Amount: "55.250000000000000000"}},
		},
		{
			name:  "mixed encodings sorted by denom",
			raw:   `{"coins":[{"denom":"umed","amount":"2.000000000000000000"},"1.000000000000000000uatom"]}`,
			field: "coins",
			want: upgradeStakingDecCoins{
				{Denom: "uatom", Amount: "1.000000000000000000"},
				{Denom: "umed", Amount: "2.000000000000000000"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			coins, err := decodeUpgradeDecCoins([]byte(test.raw), test.field, test.name)
			require.NoError(t, err)
			require.Equal(t, test.want, coins)
		})
	}

	empty, err := decodeUpgradeDecCoins([]byte(`{"coins":[]}`), "coins", "empty rewards")
	require.NoError(t, err)
	require.Empty(t, empty)
}

func TestDecodeUpgradeDecCoinsRejectsMalformedCoinsAndDenoms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		raw      string
		contains string
	}{
		{name: "missing list", raw: `{"coins":{}}`, contains: "has no coins coin list"},
		{name: "invalid compact coin", raw: `{"coins":["not-a-coin"]}`, contains: "invalid compact encoding"},
		{name: "wrong element type", raw: `{"coins":[7]}`, contains: "neither an object nor a compact string"},
		{name: "missing denom", raw: `{"coins":[{"amount":"1"}]}`, contains: "missing denom or amount"},
		{name: "missing amount", raw: `{"coins":[{"denom":"umed"}]}`, contains: "missing denom or amount"},
		{name: "invalid denom", raw: `{"coins":[{"denom":"bad denom","amount":"1"}]}`, contains: "invalid denom"},
		{name: "invalid amount", raw: `{"coins":[{"denom":"umed","amount":"not-a-decimal"}]}`, contains: "invalid amount"},
		{name: "negative amount", raw: `{"coins":[{"denom":"umed","amount":"-1"}]}`, contains: "invalid amount"},
		{name: "duplicate mixed denom", raw: `{"coins":[{"denom":"umed","amount":"1"},"2.000000000000000000umed"]}`, contains: "duplicate denom"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := decodeUpgradeDecCoins([]byte(test.raw), "coins", "test coins")
			require.ErrorContains(t, err, test.contains)
		})
	}
}

func TestValidateUpgradeStakingPreservationAllowsAccrualButRejectsStateLoss(t *testing.T) {
	before := upgradeStakingCheckpoint{
		Phase:                "pre-upgrade-checkpoint",
		Height:               40,
		DelegatorBankBalance: "900",
		RecipientBankBalance: "0",
		State: upgradeStakingState{
			Delegation: upgradeDelegationState{
				DelegatorAddress: "panacea1delegator", ValidatorOperator: "panaceavaloper1validator",
				Shares: "100.000000000000000000", Balance: upgradeStakingCoin{Denom: "umed", Amount: "100"},
			},
			Validator: upgradeValidatorState{
				OperatorAddress: "panaceavaloper1validator", ConsensusPubKey: json.RawMessage(`{"@type":"/cosmos.crypto.ed25519.PubKey","key":"AQID"}`),
				Status: "BOND_STATUS_BONDED", Tokens: "5000100", DelegatorShares: "5000100.000000000000000000",
			},
			Pool:                upgradeStakingPoolState{BondedTokens: "20000100", NotBondedTokens: "7"},
			DelegatorRewards:    upgradeStakingDecCoins{{Denom: "umed", Amount: "12.5"}},
			OutstandingRewards:  upgradeStakingDecCoins{{Denom: "umed", Amount: "55.25"}},
			ValidatorCommission: upgradeStakingDecCoins{{Denom: "umed", Amount: "4.125"}},
			SigningInfo: upgradeSigningInfoState{
				Address: "panaceavalcons1validator", StartHeight: 1, IndexOffset: 88,
				JailedUntil: "1970-01-01T00:00:00Z", MissedBlocksCounter: 0,
			},
		},
	}
	after := before
	after.Phase = "post-upgrade-preservation"
	after.Height = 50
	after.State.Validator.ConsensusPubKey = json.RawMessage(`{"type":"/cosmos.crypto.ed25519.PubKey","value":"AQID"}`)
	after.State.DelegatorRewards = upgradeStakingDecCoins{{Denom: "umed", Amount: "14.0"}}
	after.State.OutstandingRewards = upgradeStakingDecCoins{{Denom: "umed", Amount: "60.0"}}
	after.State.ValidatorCommission = upgradeStakingDecCoins{{Denom: "umed", Amount: "5.0"}}
	after.State.SigningInfo.IndexOffset = 98

	require.NoError(t, validateUpgradeStakingPreservation(before, after))

	// missed_blocks_counter is a sliding-window cardinality, not a monotonic
	// sequence. Old misses legitimately expire while identity and index_offset
	// continue forward.
	expiredMiss := before
	expiredMiss.State.SigningInfo.MissedBlocksCounter = 4
	require.NoError(t, validateUpgradeStakingPreservation(expiredMiss, after))

	lost := after
	lost.State.Delegation.Balance.Amount = "99"
	require.ErrorContains(t, validateUpgradeStakingPreservation(before, lost), "delegation balance")

	tombstoned := after
	tombstoned.State.SigningInfo.Tombstoned = true
	require.ErrorContains(t, validateUpgradeStakingPreservation(before, tombstoned), "tombstoned")

	changedKey := after
	changedKey.State.Validator.ConsensusPubKey = json.RawMessage(`{"type":"/cosmos.crypto.ed25519.PubKey","value":"AQIE"}`)
	require.ErrorContains(t, validateUpgradeStakingPreservation(before, changedKey), "public key changed")
}

func TestEqualUpgradeConsensusPubKeysNormalizesLegacyAndCurrentCLIShapes(t *testing.T) {
	legacy := json.RawMessage(`{"@type":"/cosmos.crypto.ed25519.PubKey","key":"AQID"}`)
	current := json.RawMessage(`{"type":"/cosmos.crypto.ed25519.PubKey","value":"AQID"}`)
	equal, err := equalUpgradeConsensusPubKeys(legacy, current)
	require.NoError(t, err)
	require.True(t, equal)

	changedType := json.RawMessage(`{"type":"/cosmos.crypto.secp256k1.PubKey","value":"AQID"}`)
	equal, err = equalUpgradeConsensusPubKeys(legacy, changedType)
	require.NoError(t, err)
	require.False(t, equal)

	_, err = equalUpgradeConsensusPubKeys(legacy, json.RawMessage(`{"type":"/cosmos.crypto.ed25519.PubKey","value":"not-base64"}`))
	require.ErrorContains(t, err, "canonical base64")
	_, err = equalUpgradeConsensusPubKeys(legacy, json.RawMessage(`{"@type":"/cosmos.crypto.ed25519.PubKey","key":"AQID","type":"/cosmos.crypto.ed25519.PubKey","value":"AQID"}`))
	require.ErrorContains(t, err, "exactly one")
}

func TestDecodeUpgradeDelegatorRewardsRequiresExactlyOneWellFormedTarget(t *testing.T) {
	raw := []byte(`{"rewards":[{"validator_address":"panaceavaloper1other","reward":[{"denom":"umed","amount":"2"}]},{"validator_address":"panaceavaloper1target","reward":[{"denom":"umed","amount":"3"}]}]}`)
	rewards, err := decodeUpgradeDelegatorRewards(raw, "panaceavaloper1target")
	require.NoError(t, err)
	require.Equal(t, "3", rewards.AmountOf("umed"))

	duplicate := []byte(`{"rewards":[{"validator_address":"panaceavaloper1target","reward":[]},{"validator_address":"panaceavaloper1target","reward":[]}]}`)
	_, err = decodeUpgradeDelegatorRewards(duplicate, "panaceavaloper1target")
	require.ErrorContains(t, err, "2 entries")

	missing := []byte(`{"rewards":[{"validator_address":"panaceavaloper1other","reward":[]}]}`)
	_, err = decodeUpgradeDelegatorRewards(missing, "panaceavaloper1target")
	require.ErrorContains(t, err, "0 entries")

	malformed := []byte(`{"rewards":[{"validator_address":"panaceavaloper1target","reward":"not-a-list"}]}`)
	_, err = decodeUpgradeDelegatorRewards(malformed, "panaceavaloper1target")
	require.ErrorContains(t, err, "missing validator_address or reward")

	legacyRestrictedShape := []byte(`{"rewards":[{"denom":"umed","amount":"3"}]}`)
	_, err = decodeUpgradeDelegatorRewards(legacyRestrictedShape, "panaceavaloper1target")
	require.ErrorContains(t, err, "missing validator_address or reward")

	invalidCompactCoin := []byte(`{"rewards":[{"validator_address":"panaceavaloper1target","reward":["not-a-coin"]}]}`)
	_, err = decodeUpgradeDelegatorRewards(invalidCompactCoin, "panaceavaloper1target")
	require.ErrorContains(t, err, "invalid compact encoding")
}

func TestValidateSingleValidatorStopSafetyRequiresMoreThanTwoThirdsRemainingPower(t *testing.T) {
	fourEqual := []harness.ValidatorPower{
		{Address: "AA", Power: 10},
		{Address: "BB", Power: 10},
		{Address: "CC", Power: 10},
		{Address: "DD", Power: 10},
	}
	require.NoError(t, validateSingleValidatorStopSafety(fourEqual, "AA"))

	threeEqual := fourEqual[:3]
	require.ErrorContains(t, validateSingleValidatorStopSafety(threeEqual, "AA"), "two-thirds")
	require.ErrorContains(t, validateSingleValidatorStopSafety(fourEqual, "EE"), "not present")
}

func TestValidateUpgradeStakingMutationProvesDelegationAndRewardWithdrawal(t *testing.T) {
	before := upgradeStakingCheckpoint{
		Phase: "post-upgrade-preservation", Height: 50,
		DelegatorBankBalance: "1000000", RecipientBankBalance: "0",
		State: upgradeStakingState{
			Delegation: upgradeDelegationState{
				DelegatorAddress: "panacea1delegator", ValidatorOperator: "panaceavaloper1validator",
				Shares: "100.0", Balance: upgradeStakingCoin{Denom: "umed", Amount: "100"},
			},
			Validator: upgradeValidatorState{OperatorAddress: "panaceavaloper1validator", Tokens: "5000100", DelegatorShares: "5000100.0"},
			Pool:      upgradeStakingPoolState{BondedTokens: "20000100", NotBondedTokens: "7"},
		},
	}
	beforeWithdraw := before
	beforeWithdraw.Phase = "post-upgrade-before-reward-withdraw"
	beforeWithdraw.Height = 55
	beforeWithdraw.RecipientBankBalance = "12"
	beforeWithdraw.State.Delegation.Balance.Amount = "125"
	beforeWithdraw.State.Delegation.Shares = "125.0"
	beforeWithdraw.State.Validator.Tokens = "5000125"
	beforeWithdraw.State.Validator.DelegatorShares = "5000125.0"
	beforeWithdraw.State.Pool.BondedTokens = "20000125"
	beforeWithdraw.State.DelegatorRewards = upgradeStakingDecCoins{{Denom: "umed", Amount: "9.75"}}
	after := beforeWithdraw
	after.Phase = "post-upgrade-mutation"
	after.Height = 57
	after.RecipientBankBalance = "22"
	after.State.DelegatorRewards = upgradeStakingDecCoins{{Denom: "umed", Amount: "0.25"}}

	evidence := upgradeStakingMutationEvidence{
		AdditionalDelegationAmount: "25",
		DelegateTxHash:             "DELEGATE",
		WithdrawRewardTxHash:       "WITHDRAW",
		Before:                     before,
		BeforeRewardWithdraw:       beforeWithdraw,
		After:                      after,
	}
	require.NoError(t, validateUpgradeStakingMutation(evidence))

	noRewardDrop := evidence
	noRewardDrop.After.State.DelegatorRewards = upgradeStakingDecCoins{{Denom: "umed", Amount: "10"}}
	require.ErrorContains(t, validateUpgradeStakingMutation(noRewardDrop), "did not decrease")

	noWithdrawCredit := evidence
	noWithdrawCredit.After.RecipientBankBalance = noWithdrawCredit.BeforeRewardWithdraw.RecipientBankBalance
	require.ErrorContains(t, validateUpgradeStakingMutation(noWithdrawCredit), "did not increase")
}

func TestValidateUpgradeValidatorLivenessProvesSafeMissAndSignedRejoin(t *testing.T) {
	checkpoint := func(phase string, height, offset, missed int64) upgradeStakingCheckpoint {
		return upgradeStakingCheckpoint{
			Phase: phase, Height: height,
			State: upgradeStakingState{
				Validator: upgradeValidatorState{OperatorAddress: "panaceavaloper1validator", Status: "BOND_STATUS_BONDED"},
				SigningInfo: upgradeSigningInfoState{
					Address: "panaceavalcons1validator", StartHeight: 1, IndexOffset: offset,
					JailedUntil: "1970-01-01T00:00:00Z", MissedBlocksCounter: missed,
				},
			},
		}
	}
	evidence := upgradeValidatorLivenessEvidence{
		ValidatorIndex:         0,
		TargetConsensusAddress: "AA",
		ValidatorSet: []harness.ValidatorPower{
			{Address: "AA", Power: 10}, {Address: "BB", Power: 10},
			{Address: "CC", Power: 10}, {Address: "DD", Power: 10},
		},
		MinimumStoppedBlocks: 3,
		FaultWindow: harness.QuorumHeightWindow{
			StartHeight: 60, EndHeight: 63, TargetHeight: 63,
		},
		BeforeStop:         checkpoint("post-upgrade-before-validator-stop", 60, 100, 0),
		WhileStopped:       checkpoint("post-upgrade-validator-stopped", 63, 103, 3),
		AfterRejoin:        checkpoint("post-upgrade-validator-rejoined", 67, 107, 3),
		SignedCommitHeight: 66,
		RejoinHistory: []harness.BlockEvidence{
			{Node: "validator-0", Height: 67, BlockID: "BLOCK", AppHash: "APP"},
			{Node: "full-node-0", Height: 67, BlockID: "BLOCK", AppHash: "APP"},
		},
	}
	require.NoError(t, validateUpgradeValidatorLiveness(evidence))

	tombstoned := evidence
	tombstoned.AfterRejoin.State.SigningInfo.Tombstoned = true
	require.ErrorContains(t, validateUpgradeValidatorLiveness(tombstoned), "tombstoned")

	noSignature := evidence
	noSignature.SignedCommitHeight = 0
	require.ErrorContains(t, validateUpgradeValidatorLiveness(noSignature), "commit signature")

	divergentHistory := evidence
	divergentHistory.RejoinHistory[1].AppHash = "OTHER"
	require.ErrorContains(t, validateUpgradeValidatorLiveness(divergentHistory), "history")
}
