package e2e_test

import (
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUpgradeGroupVestingFundingArgsUseDeterministicGasLimit(t *testing.T) {
	args := upgradeGroupVestingFundingArgs("faucet", "panacea1recipient", "100umed")
	require.Equal(t, []string{
		"bank", "send", "faucet", "panacea1recipient", "100umed",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	}, args)
	require.True(t, slices.Contains(args, "--gas"))
}

func TestDecodeUpgradeGroupAndVestingStateAcrossSDKShapes(t *testing.T) {
	fixture := upgradeGroupVestingFixture{
		GroupID:               7,
		GroupAdminAddress:     "panacea1admin",
		InitialMemberAddress:  "panacea1member",
		VestingAccountAddress: "panacea1vesting",
		SpendRecipientAddress: "panacea1recipient",
		VestingOriginalAmount: "1000000000",
		VestingEndTimeUnix:    4102444800,
	}
	groupState, err := decodeUpgradeGroupState(
		[]byte(`{"info":{"id":"7","admin":"panacea1admin","metadata":"v2.2.1-group","version":"1","total_weight":"1","created_at":"2026-08-04T00:00:00.123456780Z"}}`),
		[]byte(`{"members":[{"group_id":"7","member":{"address":"panacea1member","weight":"1","metadata":"initial-member","added_at":"2026-08-04T00:00:00.123456780Z"}}]}`),
		fixture,
	)
	require.NoError(t, err)
	require.Equal(t, uint64(7), groupState.ID)
	require.Equal(t, uint64(1), groupState.Version)
	require.Equal(t, "1", groupState.TotalWeight)
	require.Len(t, groupState.Members, 1)
	require.Equal(t, fixture.InitialMemberAddress, groupState.Members[0].Address)
	currentGroupState, err := decodeUpgradeGroupState(
		[]byte(`{"response":{"info":{"groupId":7,"admin":"panacea1admin","metadata":"v2.2.1-group","version":1,"totalWeight":"1","createdAt":"2026-08-04T00:00:00.12345678Z"}}}`),
		[]byte(`{"response":{"members":[{"groupId":7,"member":{"address":"panacea1member","weight":"1","metadata":"initial-member","addedAt":"2026-08-04T00:00:00.12345678Z"}}]}}`),
		fixture,
	)
	require.NoError(t, err)
	require.Equal(t, groupState, currentGroupState)
	require.Equal(t, "2026-08-04T00:00:00.12345678Z", currentGroupState.CreatedAt)
	require.Equal(t, "2026-08-04T00:00:00.12345678Z", currentGroupState.Members[0].AddedAt)

	vestingState, err := decodeUpgradeVestingState(
		[]byte(`{"account":{"@type":"/cosmos.vesting.v1beta1.DelayedVestingAccount","base_vesting_account":{"base_account":{"address":"panacea1vesting","account_number":17,"sequence":"2"},"original_vesting":[{"denom":"umed","amount":"1000000000"}],"delegated_free":[],"delegated_vesting":[],"end_time":"4102444800"}}}`),
		[]byte(`{"balances":[{"denom":"umed","amount":"1100000000"}]}`),
		[]byte(`{"balances":[{"denom":"umed","amount":"100000000"}]}`),
		[]byte(`{"balances":[{"denom":"umed","amount":"42"}]}`),
		fixture,
	)
	require.NoError(t, err)
	require.Equal(t, "/cosmos.vesting.v1beta1.DelayedVestingAccount", vestingState.AccountType)
	require.Equal(t, uint64(17), vestingState.AccountNumber)
	require.Equal(t, uint64(2), vestingState.Sequence)
	require.Equal(t, "1000000000", vestingState.OriginalVesting)
	require.Equal(t, "100000000", vestingState.SpendableBalance)
	require.Equal(t, "1000000000", vestingState.LockedBalance)
	require.Equal(t, "42", vestingState.RecipientBalance)
	currentVestingState, err := decodeUpgradeVestingState(
		[]byte(`{"account":{"type":"/cosmos.vesting.v1beta1.DelayedVestingAccount","value":{"base_vesting_account":{"base_account":{"address":"panacea1vesting","account_number":"17"},"original_vesting":[{"denom":"umed","amount":"1000000000"}],"delegated_free":[],"delegated_vesting":[],"end_time":"4102444800"}}}}`),
		[]byte(`{"response":{"balances":[{"denom":"umed","amount":"1100000000"}]}}`),
		[]byte(`{"response":{"balances":[{"denom":"umed","amount":"100000000"}]}}`),
		[]byte(`{"response":{"balances":[{"denom":"umed","amount":"42"}]}}`),
		fixture,
	)
	require.NoError(t, err)
	expectedCurrentVestingState := vestingState
	expectedCurrentVestingState.Sequence = 0
	require.Equal(t, expectedCurrentVestingState, currentVestingState)

	_, err = decodeUpgradeVestingState(
		[]byte(`{"account":{"@type":"/cosmos.vesting.v1beta1.DelayedVestingAccount","type":"/cosmos.vesting.v1beta1.DelayedVestingAccount","base_vesting_account":{}}}`),
		[]byte(`{"balances":[]}`), []byte(`{"balances":[]}`), []byte(`{"balances":[]}`), fixture,
	)
	require.ErrorContains(t, err, "exactly one")
}

func TestDecodeUpgradeGroupStateRejectsMalformedTimestamps(t *testing.T) {
	t.Parallel()

	fixture := upgradeGroupVestingFixture{
		GroupID:              7,
		GroupAdminAddress:    "panacea1admin",
		InitialMemberAddress: "panacea1member",
	}
	validInfo := []byte(`{"info":{"id":"7","admin":"panacea1admin","metadata":"v2.2.1-group","version":"1","total_weight":"1","created_at":"2026-08-04T00:00:00Z"}}`)
	validMembers := []byte(`{"members":[{"group_id":"7","member":{"address":"panacea1member","weight":"1","metadata":"initial-member","added_at":"2026-08-04T00:00:00Z"}}]}`)

	_, err := decodeUpgradeGroupState(
		[]byte(`{"info":{"id":"7","admin":"panacea1admin","metadata":"v2.2.1-group","version":"1","total_weight":"1","created_at":"not-a-time"}}`),
		validMembers,
		fixture,
	)
	require.ErrorContains(t, err, "created_at")

	_, err = decodeUpgradeGroupState(
		validInfo,
		[]byte(`{"members":[{"group_id":"7","member":{"address":"panacea1member","weight":"1","metadata":"initial-member","added_at":false}}]}`),
		fixture,
	)
	require.ErrorContains(t, err, "added_at")
}

func TestDecodeUpgradeVestingStateHandlesOmittedBaseAccountProtoDefaults(t *testing.T) {
	t.Parallel()

	fixture := upgradeGroupVestingFixture{
		VestingAccountAddress: "panacea1vesting",
		SpendRecipientAddress: "panacea1recipient",
		VestingOriginalAmount: "1000000000",
		VestingEndTimeUnix:    4102444800,
	}
	decode := func(account string) (upgradeVestingState, error) {
		return decodeUpgradeVestingState(
			[]byte(account),
			[]byte(`{"balances":[{"denom":"umed","amount":"1100000000"}]}`),
			[]byte(`{"balances":[{"denom":"umed","amount":"100000000"}]}`),
			[]byte(`{"balances":[]}`),
			fixture,
		)
	}

	state, err := decode(`{"account":{"type":"/cosmos.vesting.v1beta1.DelayedVestingAccount","value":{"base_vesting_account":{"base_account":{"address":"panacea1vesting"},"original_vesting":[{"denom":"umed","amount":"1000000000"}],"end_time":"4102444800"}}}}`)
	require.NoError(t, err)
	require.Zero(t, state.AccountNumber)
	require.Zero(t, state.Sequence)

	for _, test := range []struct {
		name     string
		account  string
		contains string
	}{
		{
			name:     "account number wrong type",
			account:  `{"account":{"type":"/cosmos.vesting.v1beta1.DelayedVestingAccount","value":{"base_vesting_account":{"base_account":{"address":"panacea1vesting","account_number":false},"original_vesting":[{"denom":"umed","amount":"1000000000"}],"end_time":"4102444800"}}}}`,
			contains: "account_number",
		},
		{
			name:     "sequence null",
			account:  `{"account":{"type":"/cosmos.vesting.v1beta1.DelayedVestingAccount","value":{"base_vesting_account":{"base_account":{"address":"panacea1vesting","sequence":null},"original_vesting":[{"denom":"umed","amount":"1000000000"}],"end_time":"4102444800"}}}}`,
			contains: "sequence",
		},
		{
			name:     "sequence wrong type",
			account:  `{"account":{"type":"/cosmos.vesting.v1beta1.DelayedVestingAccount","value":{"base_vesting_account":{"address":"panacea1vesting","base_account":{"address":"panacea1vesting","sequence":[]},"original_vesting":[{"denom":"umed","amount":"1000000000"}],"end_time":"4102444800"}}}}`,
			contains: "sequence",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := decode(test.account)
			require.ErrorContains(t, err, test.contains)
		})
	}
}

func TestValidateUpgradeGroupVestingPreservationRejectsStateLoss(t *testing.T) {
	before := upgradeGroupVestingCheckpoint{
		Phase:     "pre-upgrade-checkpoint",
		Height:    40,
		BlockTime: time.Unix(1_800_000_000, 0).UTC(),
		TxHashes:  []string{"FUND", "GROUP", "VESTING", "FREE"},
		Group: upgradeGroupState{
			ID: 7, Admin: "panacea1admin", Metadata: "v2.2.1-group", Version: 1,
			TotalWeight: "1", CreatedAt: "2026-08-04T00:00:00Z",
			Members: []upgradeGroupMemberState{{
				GroupID: 7, Address: "panacea1member", Weight: "1",
				Metadata: "initial-member", AddedAt: "2026-08-04T00:00:00Z",
			}},
		},
		Vesting: upgradeVestingState{
			AccountType: "/cosmos.vesting.v1beta1.DelayedVestingAccount", Address: "panacea1vesting",
			AccountNumber: 17, Sequence: 0, OriginalVesting: "1000000000",
			DelegatedFree: "0", DelegatedVesting: "0", EndTimeUnix: 4_102_444_800,
			BankBalance: "1100000000", LockedBalance: "1000000000", FreeBalance: "100000000",
			SpendableBalance: "100000000", RecipientBalance: "0",
		},
	}
	after := before
	after.Phase = "post-upgrade-preservation"
	after.Height = 50
	after.BlockTime = before.BlockTime.Add(10 * time.Second)
	require.NoError(t, validateUpgradeGroupVestingPreservation(before, after))

	lostMember := after
	lostMember.Group.Members = nil
	require.ErrorContains(t, validateUpgradeGroupVestingPreservation(before, lostMember), "group state")

	unlocked := after
	unlocked.Vesting.SpendableBalance = "1100000000"
	require.ErrorContains(t, validateUpgradeGroupVestingPreservation(before, unlocked), "vesting state")
}

func TestValidateUpgradeGroupVestingMutationProvesGroupWritesAndFreeCoinSpend(t *testing.T) {
	before := upgradeGroupVestingCheckpoint{
		Phase: "post-upgrade-preservation", Height: 50, BlockTime: time.Unix(1_800_000_000, 0).UTC(),
		TxHashes: []string{"FUND", "GROUP", "VESTING", "FREE"},
		Group: upgradeGroupState{
			ID: 7, Admin: "panacea1admin", Metadata: "v2.2.1-group", Version: 1,
			TotalWeight: "1", CreatedAt: "2026-08-04T00:00:00Z",
			Members: []upgradeGroupMemberState{{
				GroupID: 7, Address: "panacea1member", Weight: "1",
				Metadata: "initial-member", AddedAt: "2026-08-04T00:00:00Z",
			}},
		},
		Vesting: upgradeVestingState{
			AccountType: "/cosmos.vesting.v1beta1.DelayedVestingAccount", Address: "panacea1vesting",
			AccountNumber: 17, OriginalVesting: "1000000000", DelegatedFree: "0", DelegatedVesting: "0",
			EndTimeUnix: 4_102_444_800, BankBalance: "1100000000", LockedBalance: "1000000000",
			FreeBalance: "100000000", SpendableBalance: "100000000", RecipientBalance: "0",
		},
	}
	after := before
	after.Phase = "post-upgrade-mutation"
	after.Height = 60
	after.BlockTime = before.BlockTime.Add(10 * time.Second)
	after.TxHashes = append(append([]string(nil), before.TxHashes...), "METADATA", "MEMBER", "SPEND")
	after.Group.Metadata = "v2.3.0-group"
	after.Group.Version = 3
	after.Group.TotalWeight = "3"
	after.Group.Members = append(after.Group.Members, upgradeGroupMemberState{
		GroupID: 7, Address: "panacea1added", Weight: "2", Metadata: "post-upgrade-member",
		AddedAt: "2026-08-04T00:01:00Z",
	})
	after.Vesting.Sequence = 1
	after.Vesting.BankBalance = "1087500000"
	after.Vesting.FreeBalance = "87500000"
	after.Vesting.SpendableBalance = "87500000"
	after.Vesting.RecipientBalance = "10000000"

	evidence := upgradeGroupVestingMutationEvidence{
		UpdatedGroupMetadata: "v2.3.0-group",
		AddedMemberAddress:   "panacea1added",
		AddedMemberWeight:    "2",
		AddedMemberMetadata:  "post-upgrade-member",
		SpendAmount:          "10000000",
		GroupMetadataTxHash:  "METADATA",
		GroupMemberTxHash:    "MEMBER",
		VestingSpendTxHash:   "SPEND",
		Before:               before,
		After:                after,
		Transactions: []upgradeGroupVestingTransactionEvidence{
			{Operation: "update-group-metadata", TxHash: "METADATA", Height: 52},
			{Operation: "update-group-members", TxHash: "MEMBER", Height: 54},
			{Operation: "vesting-spend", TxHash: "SPEND", Height: 56},
		},
	}
	require.NoError(t, validateUpgradeGroupVestingMutation(evidence))

	lockedSpend := evidence
	lockedSpend.After.Vesting.LockedBalance = "999000000"
	lockedSpend.After.Vesting.FreeBalance = "88500000"
	require.ErrorContains(t, validateUpgradeGroupVestingMutation(lockedSpend), "locked")

	missingMember := evidence
	missingMember.After.Group.Members = missingMember.After.Group.Members[:1]
	require.ErrorContains(t, validateUpgradeGroupVestingMutation(missingMember), "added member")

	missingHeight := evidence
	missingHeight.Transactions[2].Height = 0
	require.ErrorContains(t, validateUpgradeGroupVestingMutation(missingHeight), "height")
}

func TestValidateUpgradeGroupVestingPostRestartProvesQueryAndSpendability(t *testing.T) {
	afterMutation := upgradeGroupVestingCheckpoint{
		Phase: "post-upgrade-mutation", Height: 60, BlockTime: time.Unix(1_800_000_000, 0).UTC(),
		TxHashes: []string{"FUND", "GROUP", "VESTING", "FREE", "METADATA", "MEMBER", "SPEND"},
		Group: upgradeGroupState{
			ID: 7, Admin: "panacea1admin", Metadata: "v2.3.0-group", Version: 3,
			TotalWeight: "3", CreatedAt: "2026-08-04T00:00:00Z",
			Members: []upgradeGroupMemberState{
				{GroupID: 7, Address: "panacea1added", Weight: "2", Metadata: "post-upgrade-member", AddedAt: "2026-08-04T00:01:00Z"},
				{GroupID: 7, Address: "panacea1member", Weight: "1", Metadata: "initial-member", AddedAt: "2026-08-04T00:00:00Z"},
			},
		},
		Vesting: upgradeVestingState{
			AccountType: "/cosmos.vesting.v1beta1.DelayedVestingAccount", Address: "panacea1vesting",
			AccountNumber: 17, Sequence: 1, OriginalVesting: "1000000000", DelegatedFree: "0", DelegatedVesting: "0",
			EndTimeUnix: 4_102_444_800, BankBalance: "1087500000", LockedBalance: "1000000000",
			FreeBalance: "87500000", SpendableBalance: "87500000", RecipientBalance: "10000000",
		},
	}
	beforeSpend := afterMutation
	beforeSpend.Phase = "post-restart-query"
	beforeSpend.Height = 70
	beforeSpend.BlockTime = afterMutation.BlockTime.Add(10 * time.Second)
	afterSpend := beforeSpend
	afterSpend.Phase = "post-restart-spend"
	afterSpend.Height = 73
	afterSpend.BlockTime = beforeSpend.BlockTime.Add(3 * time.Second)
	afterSpend.TxHashes = append(append([]string(nil), beforeSpend.TxHashes...), "RESTART-SPEND")
	afterSpend.Vesting.Sequence = 2
	afterSpend.Vesting.BankBalance = "1080000000"
	afterSpend.Vesting.FreeBalance = "80000000"
	afterSpend.Vesting.SpendableBalance = "80000000"
	afterSpend.Vesting.RecipientBalance = "15000000"
	evidence := upgradeGroupVestingPostRestartEvidence{
		SpendAmount: "5000000", SpendTxHash: "RESTART-SPEND",
		BeforeSpend: beforeSpend, AfterSpend: afterSpend,
		Transaction: upgradeGroupVestingTransactionEvidence{
			Operation: "post-restart-vesting-spend", TxHash: "RESTART-SPEND", Height: 72,
		},
	}
	require.NoError(t, validateUpgradeGroupVestingPostRestart(afterMutation, evidence))

	changedGroup := evidence
	changedGroup.AfterSpend.Group.Metadata = "lost"
	require.ErrorContains(t, validateUpgradeGroupVestingPostRestart(afterMutation, changedGroup), "group state")

	noRecipientCredit := evidence
	noRecipientCredit.AfterSpend.Vesting.RecipientBalance = beforeSpend.Vesting.RecipientBalance
	require.ErrorContains(t, validateUpgradeGroupVestingPostRestart(afterMutation, noRecipientCredit), "recipient")
}
