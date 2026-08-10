package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	upgradeGroupVestingDenom             = "umed"
	upgradeGroupAdminKey                 = "upgrade-group-admin"
	upgradeGroupInitialMemberKey         = "upgrade-group-initial-member"
	upgradeGroupAddedMemberKey           = "upgrade-group-added-member"
	upgradeVestingAccountKey             = "upgrade-vesting-account"
	upgradeVestingSpendRecipientKey      = "upgrade-vesting-spend-recipient"
	upgradeGroupInitialMetadata          = "v2.2.1-group"
	upgradeGroupUpdatedMetadata          = "v2.3.0-group"
	upgradeGroupInitialMemberMetadata    = "initial-member"
	upgradeGroupAddedMemberMetadata      = "post-upgrade-member"
	upgradeGroupInitialMemberWeight      = "1"
	upgradeGroupAddedMemberWeight        = "2"
	upgradeGroupAdminFundAmount          = "100000000"
	upgradeVestingOriginalAmount         = "1000000000"
	upgradeVestingFreeAmount             = "100000000"
	upgradeVestingPostUpgradeSpendAmount = "10000000"
	upgradeVestingPostRestartSpendAmount = "5000000"
	upgradeGroupInitialMembersPath       = "upgrade/group-vesting/initial-members.json"
	upgradeGroupAddedMembersPath         = "upgrade/group-vesting/added-members.json"
)

type upgradeGroupVestingFixture struct {
	GroupID               uint64 `json:"group_id"`
	GroupAdminKeyName     string `json:"group_admin_key_name"`
	GroupAdminAddress     string `json:"group_admin_address"`
	InitialMemberAddress  string `json:"initial_member_address"`
	AddedMemberAddress    string `json:"added_member_address"`
	VestingAccountKeyName string `json:"vesting_account_key_name"`
	VestingAccountAddress string `json:"vesting_account_address"`
	SpendRecipientAddress string `json:"spend_recipient_address"`
	VestingOriginalAmount string `json:"vesting_original_amount"`
	VestingEndTimeUnix    int64  `json:"vesting_end_time_unix"`
}

type upgradeGroupMemberState struct {
	GroupID  uint64 `json:"group_id"`
	Address  string `json:"address"`
	Weight   string `json:"weight"`
	Metadata string `json:"metadata"`
	AddedAt  string `json:"added_at"`
}

type upgradeGroupState struct {
	ID          uint64                    `json:"id"`
	Admin       string                    `json:"admin"`
	Metadata    string                    `json:"metadata"`
	Version     uint64                    `json:"version"`
	TotalWeight string                    `json:"total_weight"`
	CreatedAt   string                    `json:"created_at"`
	Members     []upgradeGroupMemberState `json:"members"`
}

type upgradeVestingState struct {
	AccountType      string `json:"account_type"`
	Address          string `json:"address"`
	AccountNumber    uint64 `json:"account_number"`
	Sequence         uint64 `json:"sequence"`
	OriginalVesting  string `json:"original_vesting"`
	DelegatedFree    string `json:"delegated_free"`
	DelegatedVesting string `json:"delegated_vesting"`
	EndTimeUnix      int64  `json:"end_time_unix"`
	BankBalance      string `json:"bank_balance"`
	LockedBalance    string `json:"locked_balance"`
	FreeBalance      string `json:"free_balance"`
	SpendableBalance string `json:"spendable_balance"`
	RecipientBalance string `json:"recipient_balance"`
}

type upgradeGroupVestingCheckpoint struct {
	Phase       string                               `json:"phase"`
	RecordedAt  time.Time                            `json:"recorded_at"`
	Height      int64                                `json:"height"`
	Observation harness.UpgradeCheckpointObservation `json:"observation"`
	BlockTime   time.Time                            `json:"block_time"`
	Group       upgradeGroupState                    `json:"group"`
	Vesting     upgradeVestingState                  `json:"vesting"`
	TxHashes    []string                             `json:"tx_hashes,omitempty"`
}

type upgradeGroupVestingPreparationEvidence struct {
	Fixture      upgradeGroupVestingFixture               `json:"fixture"`
	Transactions []upgradeGroupVestingTransactionEvidence `json:"transactions"`
	Checkpoint   upgradeGroupVestingCheckpoint            `json:"checkpoint"`
}

func (e upgradeGroupVestingPreparationEvidence) TxHashes() []string {
	txHashes := make([]string, 0, len(e.Transactions))
	for _, transaction := range e.Transactions {
		txHashes = append(txHashes, transaction.TxHash)
	}
	return txHashes
}

type upgradeGroupVestingTransactionEvidence struct {
	Operation string `json:"operation"`
	TxHash    string `json:"tx_hash"`
	Height    int64  `json:"height"`
}

type upgradeGroupVestingMutationEvidence struct {
	UpdatedGroupMetadata string                                   `json:"updated_group_metadata"`
	AddedMemberAddress   string                                   `json:"added_member_address"`
	AddedMemberWeight    string                                   `json:"added_member_weight"`
	AddedMemberMetadata  string                                   `json:"added_member_metadata"`
	SpendAmount          string                                   `json:"spend_amount"`
	GroupMetadataTxHash  string                                   `json:"group_metadata_tx_hash"`
	GroupMemberTxHash    string                                   `json:"group_member_tx_hash"`
	VestingSpendTxHash   string                                   `json:"vesting_spend_tx_hash"`
	Transactions         []upgradeGroupVestingTransactionEvidence `json:"transactions"`
	Before               upgradeGroupVestingCheckpoint            `json:"before"`
	After                upgradeGroupVestingCheckpoint            `json:"after"`
}

type upgradeGroupVestingPostRestartEvidence struct {
	SpendAmount string                                 `json:"spend_amount"`
	SpendTxHash string                                 `json:"spend_tx_hash"`
	Transaction upgradeGroupVestingTransactionEvidence `json:"transaction"`
	BeforeSpend upgradeGroupVestingCheckpoint          `json:"before_spend"`
	AfterSpend  upgradeGroupVestingCheckpoint          `json:"after_spend"`
}

func upgradeGroupVestingFundingArgs(sender, recipient, amount string) []string {
	return []string{
		"bank", "send",
		sender,
		recipient,
		amount,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	}
}

// prepareUpgradeGroupVestingMatrix creates both sides of the group/vesting
// coverage row on v2.2.1: a mutable group with one member and a delayed
// vesting account that has a separately funded free balance.
func prepareUpgradeGroupVestingMatrix(
	ctx context.Context,
	network *harness.Network,
) (upgradeGroupVestingPreparationEvidence, error) {
	if err := validateUpgradeGroupVestingNetwork(network); err != nil {
		return upgradeGroupVestingPreparationEvidence{}, err
	}
	groupAdmin, err := network.BuildWallet(ctx, upgradeGroupAdminKey, "")
	if err != nil {
		return upgradeGroupVestingPreparationEvidence{}, fmt.Errorf("build upgrade group admin: %w", err)
	}
	initialMember, err := network.BuildWallet(ctx, upgradeGroupInitialMemberKey, "")
	if err != nil {
		return upgradeGroupVestingPreparationEvidence{}, fmt.Errorf("build upgrade group initial member: %w", err)
	}
	addedMember, err := network.BuildWallet(ctx, upgradeGroupAddedMemberKey, "")
	if err != nil {
		return upgradeGroupVestingPreparationEvidence{}, fmt.Errorf("build upgrade group added member: %w", err)
	}
	vestingAccount, err := network.BuildWallet(ctx, upgradeVestingAccountKey, "")
	if err != nil {
		return upgradeGroupVestingPreparationEvidence{}, fmt.Errorf("build upgrade vesting account: %w", err)
	}
	spendRecipient, err := network.BuildWallet(ctx, upgradeVestingSpendRecipientKey, "")
	if err != nil {
		return upgradeGroupVestingPreparationEvidence{}, fmt.Errorf("build upgrade vesting spend recipient: %w", err)
	}
	node := network.Chain.Validators[0]
	fullNode := network.Chain.FullNodes[0]
	height, err := fullNode.Height(ctx)
	if err != nil {
		return upgradeGroupVestingPreparationEvidence{}, fmt.Errorf("query group/vesting preparation height: %w", err)
	}
	block, err := fullNode.Client.Block(ctx, &height)
	if err != nil {
		return upgradeGroupVestingPreparationEvidence{}, fmt.Errorf("query group/vesting preparation block %d: %w", height, err)
	}
	if block == nil || block.Block == nil {
		return upgradeGroupVestingPreparationEvidence{}, fmt.Errorf("query group/vesting preparation block %d returned no block", height)
	}
	fixture := upgradeGroupVestingFixture{
		GroupAdminKeyName:     groupAdmin.KeyName(),
		GroupAdminAddress:     groupAdmin.FormattedAddress(),
		InitialMemberAddress:  initialMember.FormattedAddress(),
		AddedMemberAddress:    addedMember.FormattedAddress(),
		VestingAccountKeyName: vestingAccount.KeyName(),
		VestingAccountAddress: vestingAccount.FormattedAddress(),
		SpendRecipientAddress: spendRecipient.FormattedAddress(),
		VestingOriginalAmount: upgradeVestingOriginalAmount,
		VestingEndTimeUnix:    block.Block.Time.Add(24 * time.Hour).Unix(),
	}
	if err := writeUpgradeGroupMembersFile(
		ctx,
		network,
		node,
		upgradeGroupInitialMembersPath,
		fixture.InitialMemberAddress,
		upgradeGroupInitialMemberWeight,
		upgradeGroupInitialMemberMetadata,
	); err != nil {
		return upgradeGroupVestingPreparationEvidence{}, err
	}

	adminFund, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-group-admin-fund",
		node,
		interchaintest.FaucetAccountKeyName,
		upgradeGroupVestingFundingArgs(
			interchaintest.FaucetAccountKeyName,
			fixture.GroupAdminAddress,
			upgradeGroupAdminFundAmount+upgradeGroupVestingDenom,
		)...,
	)
	if err != nil {
		return upgradeGroupVestingPreparationEvidence{}, fmt.Errorf("fund upgrade group admin: %w", err)
	}
	groupCreated, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-group-create-v221",
		node,
		fixture.GroupAdminKeyName,
		"group", "create-group",
		fixture.GroupAdminAddress,
		upgradeGroupInitialMetadata,
		path.Join(node.HomeDir(), upgradeGroupInitialMembersPath),
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return upgradeGroupVestingPreparationEvidence{}, fmt.Errorf("create v2.2.1 group: %w", err)
	}
	fixture.GroupID, err = upgradeGroupIDFromCommittedTx(groupCreated)
	if err != nil {
		return upgradeGroupVestingPreparationEvidence{}, err
	}
	vestingCreated, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-vesting-create-v221",
		node,
		interchaintest.FaucetAccountKeyName,
		"vesting", "create-vesting-account",
		fixture.VestingAccountAddress,
		upgradeVestingOriginalAmount+upgradeGroupVestingDenom,
		strconv.FormatInt(fixture.VestingEndTimeUnix, 10),
		"--delayed",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return upgradeGroupVestingPreparationEvidence{}, fmt.Errorf("create v2.2.1 delayed vesting account: %w", err)
	}
	freeFunded, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-vesting-fund-free-balance",
		node,
		interchaintest.FaucetAccountKeyName,
		upgradeGroupVestingFundingArgs(
			interchaintest.FaucetAccountKeyName,
			fixture.VestingAccountAddress,
			upgradeVestingFreeAmount+upgradeGroupVestingDenom,
		)...,
	)
	if err != nil {
		return upgradeGroupVestingPreparationEvidence{}, fmt.Errorf("fund upgrade vesting free balance: %w", err)
	}
	transactions, err := upgradeGroupVestingTransactionsFromResults([]struct {
		operation string
		result    *harness.TxResult
	}{
		{operation: "fund-group-admin", result: adminFund},
		{operation: "create-group", result: groupCreated},
		{operation: "create-delayed-vesting-account", result: vestingCreated},
		{operation: "fund-vesting-free-balance", result: freeFunded},
	})
	if err != nil {
		return upgradeGroupVestingPreparationEvidence{}, err
	}
	txHashes := make([]string, 0, len(transactions))
	for _, transaction := range transactions {
		txHashes = append(txHashes, transaction.TxHash)
	}
	checkpoint, err := captureUpgradeGroupVestingCheckpoint(
		ctx,
		network,
		fixture,
		"v2.2.1-preparation",
		txHashes,
	)
	if err != nil {
		return upgradeGroupVestingPreparationEvidence{}, err
	}
	if err := validateUpgradeGroupVestingTransactions(
		transactions,
		[]upgradeGroupVestingTransactionEvidence{
			{Operation: "fund-group-admin", TxHash: adminFund.TxHash},
			{Operation: "create-group", TxHash: groupCreated.TxHash},
			{Operation: "create-delayed-vesting-account", TxHash: vestingCreated.TxHash},
			{Operation: "fund-vesting-free-balance", TxHash: freeFunded.TxHash},
		},
		0,
		checkpoint.Height,
	); err != nil {
		return upgradeGroupVestingPreparationEvidence{}, err
	}
	if checkpoint.Group.Metadata != upgradeGroupInitialMetadata || checkpoint.Group.Version != 1 ||
		checkpoint.Group.TotalWeight != upgradeGroupInitialMemberWeight || len(checkpoint.Group.Members) != 1 {
		return upgradeGroupVestingPreparationEvidence{}, fmt.Errorf("unexpected initial group state: %+v", checkpoint.Group)
	}
	initialState := checkpoint.Group.Members[0]
	if initialState.Address != fixture.InitialMemberAddress || initialState.Weight != upgradeGroupInitialMemberWeight ||
		initialState.Metadata != upgradeGroupInitialMemberMetadata {
		return upgradeGroupVestingPreparationEvidence{}, fmt.Errorf("unexpected initial group member: %+v", initialState)
	}
	wantBankBalance := mustUpgradeGroupVestingInt(upgradeVestingOriginalAmount).Add(mustUpgradeGroupVestingInt(upgradeVestingFreeAmount))
	if checkpoint.Vesting.OriginalVesting != upgradeVestingOriginalAmount || checkpoint.Vesting.DelegatedFree != "0" ||
		checkpoint.Vesting.DelegatedVesting != "0" || checkpoint.Vesting.LockedBalance != upgradeVestingOriginalAmount ||
		checkpoint.Vesting.BankBalance != wantBankBalance.String() || checkpoint.Vesting.FreeBalance != upgradeVestingFreeAmount ||
		checkpoint.Vesting.SpendableBalance != upgradeVestingFreeAmount || checkpoint.Vesting.RecipientBalance != "0" {
		return upgradeGroupVestingPreparationEvidence{}, fmt.Errorf("unexpected initial delayed vesting state: %+v", checkpoint.Vesting)
	}
	evidence := upgradeGroupVestingPreparationEvidence{
		Fixture:      fixture,
		Transactions: transactions,
		Checkpoint:   checkpoint,
	}
	if err := network.WriteArtifactJSON("upgrade/group-vesting/preparation.json", evidence); err != nil {
		return upgradeGroupVestingPreparationEvidence{}, fmt.Errorf("record group/vesting preparation: %w", err)
	}
	return evidence, nil
}

func captureUpgradeGroupVestingCheckpoint(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeGroupVestingFixture,
	phase string,
	txHashes []string,
) (upgradeGroupVestingCheckpoint, error) {
	if err := validateUpgradeGroupVestingNetwork(network); err != nil {
		return upgradeGroupVestingCheckpoint{}, err
	}
	if err := validateUpgradeGroupVestingPhase(phase); err != nil {
		return upgradeGroupVestingCheckpoint{}, err
	}
	if fixture.GroupID == 0 || fixture.GroupAdminAddress == "" || fixture.InitialMemberAddress == "" ||
		fixture.VestingAccountAddress == "" || fixture.SpendRecipientAddress == "" {
		return upgradeGroupVestingCheckpoint{}, errors.New("group/vesting checkpoint fixture is incomplete")
	}
	fullNode := network.Chain.FullNodes[0]
	height, err := fullNode.Height(ctx)
	if err != nil {
		return upgradeGroupVestingCheckpoint{}, fmt.Errorf("query %s group/vesting height: %w", phase, err)
	}
	block, err := fullNode.Client.Block(ctx, &height)
	if err != nil {
		return upgradeGroupVestingCheckpoint{}, fmt.Errorf("query %s group/vesting block %d: %w", phase, height, err)
	}
	if block == nil || block.Block == nil {
		return upgradeGroupVestingCheckpoint{}, fmt.Errorf("query %s group/vesting block %d returned no block", phase, height)
	}
	observation, err := network.CaptureUpgradeCheckpointObservation(ctx, "upgrade-group-vesting-"+phase, fullNode, height)
	if err != nil {
		return upgradeGroupVestingCheckpoint{}, err
	}
	query := func(suffix string, command ...string) ([]byte, error) {
		command = append(append([]string(nil), command...), "--height", strconv.FormatInt(height, 10))
		return network.FullNodeCLIQuery(ctx, "upgrade-group-vesting-"+phase+"-"+suffix, command...)
	}
	groupInfoRaw, err := query("group-info", "group", "group-info", strconv.FormatUint(fixture.GroupID, 10))
	if err != nil {
		return upgradeGroupVestingCheckpoint{}, err
	}
	groupMembersRaw, err := query("group-members", "group", "group-members", strconv.FormatUint(fixture.GroupID, 10))
	if err != nil {
		return upgradeGroupVestingCheckpoint{}, err
	}
	vestingAccountRaw, err := query("vesting-account", "auth", "account", fixture.VestingAccountAddress)
	if err != nil {
		return upgradeGroupVestingCheckpoint{}, err
	}
	vestingBalancesRaw, err := query("vesting-balances", "bank", "balances", fixture.VestingAccountAddress)
	if err != nil {
		return upgradeGroupVestingCheckpoint{}, err
	}
	spendableRaw, err := query("vesting-spendable", "bank", "spendable-balances", fixture.VestingAccountAddress)
	if err != nil {
		return upgradeGroupVestingCheckpoint{}, err
	}
	recipientBalancesRaw, err := query("recipient-balances", "bank", "balances", fixture.SpendRecipientAddress)
	if err != nil {
		return upgradeGroupVestingCheckpoint{}, err
	}
	groupState, err := decodeUpgradeGroupState(groupInfoRaw, groupMembersRaw, fixture)
	if err != nil {
		return upgradeGroupVestingCheckpoint{}, fmt.Errorf("decode %s group state: %w", phase, err)
	}
	vestingState, err := decodeUpgradeVestingState(
		vestingAccountRaw,
		vestingBalancesRaw,
		spendableRaw,
		recipientBalancesRaw,
		fixture,
	)
	if err != nil {
		return upgradeGroupVestingCheckpoint{}, fmt.Errorf("decode %s vesting state: %w", phase, err)
	}
	checkpoint := upgradeGroupVestingCheckpoint{
		Phase:       phase,
		RecordedAt:  observation.ObservedAt,
		Height:      height,
		Observation: observation,
		BlockTime:   block.Block.Time.UTC(),
		Group:       groupState,
		Vesting:     vestingState,
		TxHashes:    append([]string(nil), txHashes...),
	}
	if err := validateUpgradeGroupVestingCheckpoint(checkpoint); err != nil {
		return upgradeGroupVestingCheckpoint{}, err
	}
	if err := network.WriteArtifactJSON("upgrade/group-vesting/checkpoints/"+phase+".json", checkpoint); err != nil {
		return upgradeGroupVestingCheckpoint{}, fmt.Errorf("record %s group/vesting checkpoint: %w", phase, err)
	}
	if err := network.AppendArtifactJSON("upgrade/group-vesting/phases.jsonl", checkpoint); err != nil {
		return upgradeGroupVestingCheckpoint{}, fmt.Errorf("record %s group/vesting phase: %w", phase, err)
	}
	return checkpoint, nil
}

func captureAndValidateUpgradeGroupVestingPreservation(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeGroupVestingFixture,
	before upgradeGroupVestingCheckpoint,
) (upgradeGroupVestingCheckpoint, error) {
	after, err := captureUpgradeGroupVestingCheckpoint(
		ctx,
		network,
		fixture,
		"post-upgrade-preservation",
		before.TxHashes,
	)
	if err != nil {
		return upgradeGroupVestingCheckpoint{}, err
	}
	if err := validateUpgradeGroupVestingPreservation(before, after); err != nil {
		return upgradeGroupVestingCheckpoint{}, fmt.Errorf("validate group/vesting state across upgrade: %w", err)
	}
	return after, nil
}

func mutateUpgradeGroupVestingMatrix(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeGroupVestingFixture,
	before upgradeGroupVestingCheckpoint,
) (upgradeGroupVestingMutationEvidence, error) {
	if err := validateUpgradeGroupVestingNetwork(network); err != nil {
		return upgradeGroupVestingMutationEvidence{}, err
	}
	if err := validateUpgradeGroupVestingCheckpoint(before); err != nil {
		return upgradeGroupVestingMutationEvidence{}, err
	}
	if before.Group.ID != fixture.GroupID || before.Vesting.Address != fixture.VestingAccountAddress {
		return upgradeGroupVestingMutationEvidence{}, errors.New("pre-mutation checkpoint does not match group/vesting fixture")
	}
	node := network.Chain.Validators[0]
	if err := writeUpgradeGroupMembersFile(
		ctx,
		network,
		node,
		upgradeGroupAddedMembersPath,
		fixture.AddedMemberAddress,
		upgradeGroupAddedMemberWeight,
		upgradeGroupAddedMemberMetadata,
	); err != nil {
		return upgradeGroupVestingMutationEvidence{}, err
	}
	metadataUpdated, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-group-update-metadata",
		node,
		fixture.GroupAdminKeyName,
		"group", "update-group-metadata",
		fixture.GroupAdminAddress,
		strconv.FormatUint(fixture.GroupID, 10),
		upgradeGroupUpdatedMetadata,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return upgradeGroupVestingMutationEvidence{}, fmt.Errorf("update group metadata after upgrade: %w", err)
	}
	membersUpdated, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-group-add-member",
		node,
		fixture.GroupAdminKeyName,
		"group", "update-group-members",
		fixture.GroupAdminAddress,
		strconv.FormatUint(fixture.GroupID, 10),
		path.Join(node.HomeDir(), upgradeGroupAddedMembersPath),
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return upgradeGroupVestingMutationEvidence{}, fmt.Errorf("add group member after upgrade: %w", err)
	}
	spent, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-vesting-spend-free-balance",
		node,
		fixture.VestingAccountKeyName,
		"bank", "send",
		fixture.VestingAccountAddress,
		fixture.SpendRecipientAddress,
		upgradeVestingPostUpgradeSpendAmount+upgradeGroupVestingDenom,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return upgradeGroupVestingMutationEvidence{}, fmt.Errorf("spend delayed vesting free balance after upgrade: %w", err)
	}
	transactions, err := upgradeGroupVestingTransactionsFromResults([]struct {
		operation string
		result    *harness.TxResult
	}{
		{operation: "update-group-metadata", result: metadataUpdated},
		{operation: "update-group-members", result: membersUpdated},
		{operation: "vesting-spend", result: spent},
	})
	if err != nil {
		return upgradeGroupVestingMutationEvidence{}, err
	}
	afterHashes := append(append([]string(nil), before.TxHashes...), metadataUpdated.TxHash, membersUpdated.TxHash, spent.TxHash)
	after, err := captureUpgradeGroupVestingCheckpoint(
		ctx,
		network,
		fixture,
		"post-upgrade-mutation",
		afterHashes,
	)
	if err != nil {
		return upgradeGroupVestingMutationEvidence{}, err
	}
	evidence := upgradeGroupVestingMutationEvidence{
		UpdatedGroupMetadata: upgradeGroupUpdatedMetadata,
		AddedMemberAddress:   fixture.AddedMemberAddress,
		AddedMemberWeight:    upgradeGroupAddedMemberWeight,
		AddedMemberMetadata:  upgradeGroupAddedMemberMetadata,
		SpendAmount:          upgradeVestingPostUpgradeSpendAmount,
		GroupMetadataTxHash:  metadataUpdated.TxHash,
		GroupMemberTxHash:    membersUpdated.TxHash,
		VestingSpendTxHash:   spent.TxHash,
		Transactions:         transactions,
		Before:               before,
		After:                after,
	}
	if err := validateUpgradeGroupVestingMutation(evidence); err != nil {
		return upgradeGroupVestingMutationEvidence{}, err
	}
	if err := network.WriteArtifactJSON("upgrade/group-vesting/mutation.json", evidence); err != nil {
		return upgradeGroupVestingMutationEvidence{}, fmt.Errorf("record group/vesting mutation: %w", err)
	}
	return evidence, nil
}

// exerciseUpgradeGroupVestingPostRestart first proves that both objects are
// queryable unchanged after the all-node restart, then performs a second spend
// from the same still-locked vesting account to prove free coins remain usable.
func exerciseUpgradeGroupVestingPostRestart(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeGroupVestingFixture,
	mutation upgradeGroupVestingMutationEvidence,
) (upgradeGroupVestingPostRestartEvidence, error) {
	if err := validateUpgradeGroupVestingMutation(mutation); err != nil {
		return upgradeGroupVestingPostRestartEvidence{}, fmt.Errorf("validate pre-restart group/vesting mutation: %w", err)
	}
	beforeSpend, err := captureUpgradeGroupVestingCheckpoint(
		ctx,
		network,
		fixture,
		"post-restart-query",
		mutation.After.TxHashes,
	)
	if err != nil {
		return upgradeGroupVestingPostRestartEvidence{}, err
	}
	if err := validateUpgradeGroupVestingPreservation(mutation.After, beforeSpend); err != nil {
		return upgradeGroupVestingPostRestartEvidence{}, fmt.Errorf("validate group/vesting state after restart: %w", err)
	}
	node := network.Chain.Validators[0]
	spent, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-post-restart-vesting-spend",
		node,
		fixture.VestingAccountKeyName,
		"bank", "send",
		fixture.VestingAccountAddress,
		fixture.SpendRecipientAddress,
		upgradeVestingPostRestartSpendAmount+upgradeGroupVestingDenom,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return upgradeGroupVestingPostRestartEvidence{}, fmt.Errorf("spend delayed vesting free balance after restart: %w", err)
	}
	transactions, err := upgradeGroupVestingTransactionsFromResults([]struct {
		operation string
		result    *harness.TxResult
	}{{operation: "post-restart-vesting-spend", result: spent}})
	if err != nil {
		return upgradeGroupVestingPostRestartEvidence{}, err
	}
	afterHashes := append(append([]string(nil), beforeSpend.TxHashes...), spent.TxHash)
	afterSpend, err := captureUpgradeGroupVestingCheckpoint(
		ctx,
		network,
		fixture,
		"post-restart-spend",
		afterHashes,
	)
	if err != nil {
		return upgradeGroupVestingPostRestartEvidence{}, err
	}
	evidence := upgradeGroupVestingPostRestartEvidence{
		SpendAmount: upgradeVestingPostRestartSpendAmount,
		SpendTxHash: spent.TxHash,
		Transaction: transactions[0],
		BeforeSpend: beforeSpend,
		AfterSpend:  afterSpend,
	}
	if err := validateUpgradeGroupVestingPostRestart(mutation.After, evidence); err != nil {
		return upgradeGroupVestingPostRestartEvidence{}, err
	}
	if err := network.WriteArtifactJSON("upgrade/group-vesting/post-restart.json", evidence); err != nil {
		return upgradeGroupVestingPostRestartEvidence{}, fmt.Errorf("record post-restart group/vesting evidence: %w", err)
	}
	return evidence, nil
}

func validateUpgradeGroupVestingNetwork(network *harness.Network) error {
	if network == nil || network.Chain == nil {
		return errors.New("group/vesting network is required")
	}
	if len(network.Chain.Validators) == 0 || len(network.Chain.FullNodes) == 0 {
		return errors.New("group/vesting network requires a validator and full node")
	}
	if network.Chain.FullNodes[0].Client == nil {
		return errors.New("group/vesting full node has no RPC client")
	}
	return nil
}

func validateUpgradeGroupVestingPhase(phase string) error {
	if phase == "" || phase != strings.TrimSpace(phase) {
		return errors.New("group/vesting checkpoint phase is required without surrounding whitespace")
	}
	for _, character := range phase {
		if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || character == '-' || character == '.' {
			continue
		}
		return fmt.Errorf("group/vesting checkpoint phase %q contains unsupported character %q", phase, character)
	}
	return nil
}

func writeUpgradeGroupMembersFile(
	ctx context.Context,
	network *harness.Network,
	node *cosmos.ChainNode,
	relativePath string,
	address string,
	weight string,
	metadata string,
) error {
	if node == nil || strings.TrimSpace(relativePath) == "" || strings.TrimSpace(address) == "" {
		return errors.New("group member file node, path, and address are required")
	}
	document, err := json.MarshalIndent(map[string]any{
		"members": []map[string]string{{
			"address":  address,
			"weight":   weight,
			"metadata": metadata,
		}},
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode group member file: %w", err)
	}
	if err := node.WriteFile(ctx, document, relativePath); err != nil {
		return fmt.Errorf("write group member file %s: %w", relativePath, err)
	}
	if err := network.WriteArtifact(relativePath, document); err != nil {
		return fmt.Errorf("record group member file %s: %w", relativePath, err)
	}
	return nil
}

func upgradeGroupIDFromCommittedTx(result *harness.TxResult) (uint64, error) {
	if result == nil {
		return 0, errors.New("group creation has no committed transaction")
	}
	event, ok := result.FindEvent("cosmos.group.v1.EventCreateGroup")
	if !ok {
		return 0, errors.New("group creation transaction has no EventCreateGroup")
	}
	groupID, err := strconv.ParseUint(strings.TrimSpace(event.Attribute("group_id")), 10, 64)
	if err != nil || groupID == 0 {
		return 0, fmt.Errorf("group creation event has invalid group_id %q", event.Attribute("group_id"))
	}
	return groupID, nil
}

func upgradeGroupVestingTransactionsFromResults(results []struct {
	operation string
	result    *harness.TxResult
}) ([]upgradeGroupVestingTransactionEvidence, error) {
	evidence := make([]upgradeGroupVestingTransactionEvidence, 0, len(results))
	for _, result := range results {
		if strings.TrimSpace(result.operation) == "" || result.result == nil || strings.TrimSpace(result.result.TxHash) == "" ||
			result.result.HeightInt64() <= 0 {
			return nil, fmt.Errorf("group/vesting transaction result is incomplete for operation %q", result.operation)
		}
		evidence = append(evidence, upgradeGroupVestingTransactionEvidence{
			Operation: result.operation,
			TxHash:    result.result.TxHash,
			Height:    result.result.HeightInt64(),
		})
	}
	return evidence, nil
}

func validateUpgradeGroupVestingPostRestart(
	afterMutation upgradeGroupVestingCheckpoint,
	evidence upgradeGroupVestingPostRestartEvidence,
) error {
	if err := validateUpgradeGroupVestingPreservation(afterMutation, evidence.BeforeSpend); err != nil {
		return fmt.Errorf("validate group/vesting post-restart query: %w", err)
	}
	if err := validateUpgradeGroupVestingCheckpoint(evidence.AfterSpend); err != nil {
		return fmt.Errorf("validate group/vesting post-restart spend checkpoint: %w", err)
	}
	if evidence.AfterSpend.Height <= evidence.BeforeSpend.Height || evidence.AfterSpend.BlockTime.Before(evidence.BeforeSpend.BlockTime) {
		return fmt.Errorf("post-restart vesting spend checkpoints did not advance from %d to %d", evidence.BeforeSpend.Height, evidence.AfterSpend.Height)
	}
	if !reflect.DeepEqual(evidence.BeforeSpend.Group, evidence.AfterSpend.Group) {
		return fmt.Errorf("group state changed during post-restart vesting spend: before=%+v after=%+v", evidence.BeforeSpend.Group, evidence.AfterSpend.Group)
	}
	if strings.TrimSpace(evidence.SpendTxHash) == "" {
		return errors.New("post-restart vesting spend transaction hash is required")
	}
	if err := validateUpgradeGroupVestingTransactions(
		[]upgradeGroupVestingTransactionEvidence{evidence.Transaction},
		[]upgradeGroupVestingTransactionEvidence{{Operation: "post-restart-vesting-spend", TxHash: evidence.SpendTxHash}},
		evidence.BeforeSpend.Height,
		evidence.AfterSpend.Height,
	); err != nil {
		return err
	}
	for _, previous := range evidence.BeforeSpend.TxHashes {
		if strings.EqualFold(previous, evidence.SpendTxHash) {
			return fmt.Errorf("post-restart vesting spend transaction hash %s is not new", evidence.SpendTxHash)
		}
	}
	wantLineage := append(append([]string(nil), evidence.BeforeSpend.TxHashes...), evidence.SpendTxHash)
	if !reflect.DeepEqual(wantLineage, evidence.AfterSpend.TxHashes) {
		return fmt.Errorf("post-restart group/vesting transaction lineage is %v, want %v", evidence.AfterSpend.TxHashes, wantLineage)
	}
	if err := validateUpgradeVestingSpend(evidence.BeforeSpend.Vesting, evidence.AfterSpend.Vesting, evidence.SpendAmount); err != nil {
		return fmt.Errorf("validate post-restart vesting spendability: %w", err)
	}
	return nil
}

func validateUpgradeGroupVestingMutation(evidence upgradeGroupVestingMutationEvidence) error {
	if err := validateUpgradeGroupVestingCheckpoint(evidence.Before); err != nil {
		return fmt.Errorf("validate group/vesting mutation before checkpoint: %w", err)
	}
	if err := validateUpgradeGroupVestingCheckpoint(evidence.After); err != nil {
		return fmt.Errorf("validate group/vesting mutation after checkpoint: %w", err)
	}
	if evidence.After.Height <= evidence.Before.Height || evidence.After.BlockTime.Before(evidence.Before.BlockTime) {
		return fmt.Errorf("group/vesting mutation checkpoints did not advance from %d to %d", evidence.Before.Height, evidence.After.Height)
	}
	txHashes := []string{evidence.GroupMetadataTxHash, evidence.GroupMemberTxHash, evidence.VestingSpendTxHash}
	seenTxHashes := make(map[string]struct{}, len(txHashes))
	for _, txHash := range txHashes {
		if strings.TrimSpace(txHash) == "" {
			return errors.New("group/vesting mutation transaction hashes are required")
		}
		normalized := strings.ToUpper(txHash)
		if _, duplicate := seenTxHashes[normalized]; duplicate {
			return fmt.Errorf("group/vesting mutation transaction hash %s is duplicated", txHash)
		}
		seenTxHashes[normalized] = struct{}{}
	}
	if err := validateUpgradeGroupVestingTransactions(
		evidence.Transactions,
		[]upgradeGroupVestingTransactionEvidence{
			{Operation: "update-group-metadata", TxHash: evidence.GroupMetadataTxHash},
			{Operation: "update-group-members", TxHash: evidence.GroupMemberTxHash},
			{Operation: "vesting-spend", TxHash: evidence.VestingSpendTxHash},
		},
		evidence.Before.Height,
		evidence.After.Height,
	); err != nil {
		return err
	}
	wantLineage := append(append([]string(nil), evidence.Before.TxHashes...), txHashes...)
	if !reflect.DeepEqual(wantLineage, evidence.After.TxHashes) {
		return fmt.Errorf("group/vesting mutation transaction lineage is %v, want %v", evidence.After.TxHashes, wantLineage)
	}

	beforeGroup := evidence.Before.Group
	afterGroup := evidence.After.Group
	if beforeGroup.ID != afterGroup.ID || beforeGroup.Admin != afterGroup.Admin || beforeGroup.CreatedAt != afterGroup.CreatedAt {
		return errors.New("group identity changed during supported mutations")
	}
	if afterGroup.Metadata != evidence.UpdatedGroupMetadata {
		return fmt.Errorf("updated group metadata is %q, want %q", afterGroup.Metadata, evidence.UpdatedGroupMetadata)
	}
	if afterGroup.Version != beforeGroup.Version+2 {
		return fmt.Errorf("group version advanced from %d to %d, want two writes", beforeGroup.Version, afterGroup.Version)
	}
	addedWeight, err := sdkmath.LegacyNewDecFromStr(evidence.AddedMemberWeight)
	if err != nil || !addedWeight.IsPositive() {
		return fmt.Errorf("added member weight is invalid: %q", evidence.AddedMemberWeight)
	}
	beforeWeight, err := sdkmath.LegacyNewDecFromStr(beforeGroup.TotalWeight)
	if err != nil {
		return fmt.Errorf("decode pre-mutation group weight: %w", err)
	}
	afterWeight, err := sdkmath.LegacyNewDecFromStr(afterGroup.TotalWeight)
	if err != nil || !afterWeight.Equal(beforeWeight.Add(addedWeight)) {
		return fmt.Errorf("group total weight after mutation is %s, want %s", afterGroup.TotalWeight, beforeWeight.Add(addedWeight))
	}
	beforeMembers := make(map[string]upgradeGroupMemberState, len(beforeGroup.Members))
	for _, member := range beforeGroup.Members {
		beforeMembers[member.Address] = member
	}
	var addedMember *upgradeGroupMemberState
	for index := range afterGroup.Members {
		member := afterGroup.Members[index]
		if member.Address == evidence.AddedMemberAddress {
			copyOfMember := member
			addedMember = &copyOfMember
			continue
		}
		beforeMember, found := beforeMembers[member.Address]
		if !found || !reflect.DeepEqual(beforeMember, member) {
			return fmt.Errorf("existing group member %s changed during mutation", member.Address)
		}
		delete(beforeMembers, member.Address)
	}
	if len(beforeMembers) != 0 {
		return fmt.Errorf("group mutation lost existing members: %v", beforeMembers)
	}
	if addedMember == nil {
		return fmt.Errorf("added member %s is missing", evidence.AddedMemberAddress)
	}
	if addedMember.GroupID != beforeGroup.ID || addedMember.Weight != evidence.AddedMemberWeight ||
		addedMember.Metadata != evidence.AddedMemberMetadata || strings.TrimSpace(addedMember.AddedAt) == "" {
		return fmt.Errorf("added member state is invalid: %+v", *addedMember)
	}

	if err := validateUpgradeVestingSpend(evidence.Before.Vesting, evidence.After.Vesting, evidence.SpendAmount); err != nil {
		return fmt.Errorf("validate post-upgrade vesting spendability: %w", err)
	}
	return nil
}

func validateUpgradeGroupVestingTransactions(
	actual []upgradeGroupVestingTransactionEvidence,
	expected []upgradeGroupVestingTransactionEvidence,
	minimumHeight int64,
	maximumHeight int64,
) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("group/vesting transaction evidence has %d entries, want %d", len(actual), len(expected))
	}
	previousHeight := minimumHeight
	for index := range expected {
		observed := actual[index]
		want := expected[index]
		if observed.Operation != want.Operation || !strings.EqualFold(observed.TxHash, want.TxHash) {
			return fmt.Errorf("group/vesting transaction %d is %+v, want operation=%s hash=%s", index, observed, want.Operation, want.TxHash)
		}
		if observed.Height <= previousHeight || observed.Height > maximumHeight {
			return fmt.Errorf(
				"group/vesting transaction %s height %d is outside (%d,%d]",
				observed.Operation,
				observed.Height,
				previousHeight,
				maximumHeight,
			)
		}
		previousHeight = observed.Height
	}
	return nil
}

func validateUpgradeVestingSpend(beforeVesting, afterVesting upgradeVestingState, spendAmountText string) error {
	if beforeVesting.AccountType != afterVesting.AccountType || beforeVesting.Address != afterVesting.Address ||
		beforeVesting.AccountNumber != afterVesting.AccountNumber || beforeVesting.OriginalVesting != afterVesting.OriginalVesting ||
		beforeVesting.DelegatedFree != afterVesting.DelegatedFree || beforeVesting.DelegatedVesting != afterVesting.DelegatedVesting ||
		beforeVesting.EndTimeUnix != afterVesting.EndTimeUnix {
		return errors.New("vesting account identity or schedule changed during spend")
	}
	if beforeVesting.LockedBalance != afterVesting.LockedBalance {
		return fmt.Errorf("vesting locked balance changed from %s to %s", beforeVesting.LockedBalance, afterVesting.LockedBalance)
	}
	if afterVesting.Sequence != beforeVesting.Sequence+1 {
		return fmt.Errorf("vesting account sequence advanced from %d to %d, want one spend", beforeVesting.Sequence, afterVesting.Sequence)
	}
	spendAmount, ok := sdkmath.NewIntFromString(spendAmountText)
	if !ok || !spendAmount.IsPositive() {
		return fmt.Errorf("vesting spend amount is invalid: %q", spendAmountText)
	}
	bankDelta := mustUpgradeGroupVestingInt(beforeVesting.BankBalance).Sub(mustUpgradeGroupVestingInt(afterVesting.BankBalance))
	freeDelta := mustUpgradeGroupVestingInt(beforeVesting.FreeBalance).Sub(mustUpgradeGroupVestingInt(afterVesting.FreeBalance))
	spendableDelta := mustUpgradeGroupVestingInt(beforeVesting.SpendableBalance).Sub(mustUpgradeGroupVestingInt(afterVesting.SpendableBalance))
	recipientDelta := mustUpgradeGroupVestingInt(afterVesting.RecipientBalance).Sub(mustUpgradeGroupVestingInt(beforeVesting.RecipientBalance))
	if bankDelta.LT(spendAmount) || !bankDelta.Equal(freeDelta) || !bankDelta.Equal(spendableDelta) {
		return fmt.Errorf(
			"vesting free spend deltas disagree: amount=%s bank=%s free=%s spendable=%s",
			spendAmount,
			bankDelta,
			freeDelta,
			spendableDelta,
		)
	}
	if !recipientDelta.Equal(spendAmount) {
		return fmt.Errorf("vesting spend recipient gained %s, want %s", recipientDelta, spendAmount)
	}
	return nil
}

func validateUpgradeGroupVestingPreservation(before, after upgradeGroupVestingCheckpoint) error {
	if err := validateUpgradeGroupVestingCheckpoint(before); err != nil {
		return fmt.Errorf("validate before group/vesting checkpoint: %w", err)
	}
	if err := validateUpgradeGroupVestingCheckpoint(after); err != nil {
		return fmt.Errorf("validate after group/vesting checkpoint: %w", err)
	}
	if after.Height <= before.Height || after.BlockTime.Before(before.BlockTime) {
		return fmt.Errorf(
			"group/vesting checkpoint did not advance: before=%d/%s after=%d/%s",
			before.Height,
			before.BlockTime,
			after.Height,
			after.BlockTime,
		)
	}
	if !reflect.DeepEqual(before.Group, after.Group) {
		return fmt.Errorf("group state changed across preservation boundary: before=%+v after=%+v", before.Group, after.Group)
	}
	if !reflect.DeepEqual(before.Vesting, after.Vesting) {
		return fmt.Errorf("vesting state changed across preservation boundary: before=%+v after=%+v", before.Vesting, after.Vesting)
	}
	if !reflect.DeepEqual(before.TxHashes, after.TxHashes) {
		return fmt.Errorf("group/vesting transaction lineage changed: before=%v after=%v", before.TxHashes, after.TxHashes)
	}
	return nil
}

func validateUpgradeGroupVestingCheckpoint(checkpoint upgradeGroupVestingCheckpoint) error {
	if checkpoint.Height <= 0 || checkpoint.BlockTime.IsZero() {
		return fmt.Errorf("group/vesting checkpoint height or block time is invalid: %d/%s", checkpoint.Height, checkpoint.BlockTime)
	}
	if checkpoint.Group.ID == 0 || checkpoint.Group.Admin == "" || checkpoint.Group.Version == 0 ||
		checkpoint.Group.TotalWeight == "" || len(checkpoint.Group.Members) == 0 {
		return errors.New("group state checkpoint is incomplete")
	}
	if checkpoint.Vesting.AccountType != "/cosmos.vesting.v1beta1.DelayedVestingAccount" ||
		checkpoint.Vesting.Address == "" || checkpoint.Vesting.EndTimeUnix <= checkpoint.BlockTime.Unix() {
		return errors.New("vesting state checkpoint is not an active delayed vesting account")
	}
	for label, value := range map[string]string{
		"original vesting":  checkpoint.Vesting.OriginalVesting,
		"delegated free":    checkpoint.Vesting.DelegatedFree,
		"delegated vesting": checkpoint.Vesting.DelegatedVesting,
		"bank balance":      checkpoint.Vesting.BankBalance,
		"locked balance":    checkpoint.Vesting.LockedBalance,
		"free balance":      checkpoint.Vesting.FreeBalance,
		"spendable balance": checkpoint.Vesting.SpendableBalance,
		"recipient balance": checkpoint.Vesting.RecipientBalance,
	} {
		amount, ok := sdkmath.NewIntFromString(value)
		if !ok || amount.IsNegative() {
			return fmt.Errorf("vesting state checkpoint has invalid %s %q", label, value)
		}
	}
	bank := mustUpgradeGroupVestingInt(checkpoint.Vesting.BankBalance)
	locked := mustUpgradeGroupVestingInt(checkpoint.Vesting.LockedBalance)
	free := mustUpgradeGroupVestingInt(checkpoint.Vesting.FreeBalance)
	spendable := mustUpgradeGroupVestingInt(checkpoint.Vesting.SpendableBalance)
	if !locked.Add(free).Equal(bank) || !free.Equal(spendable) {
		return fmt.Errorf(
			"vesting state checkpoint balance partition is inconsistent: bank=%s locked=%s free=%s spendable=%s",
			bank,
			locked,
			free,
			spendable,
		)
	}
	return nil
}

func decodeUpgradeGroupState(
	groupInfoRaw []byte,
	groupMembersRaw []byte,
	fixture upgradeGroupVestingFixture,
) (upgradeGroupState, error) {
	infoDocument, err := decodeUpgradeGroupVestingJSON(groupInfoRaw, "group info")
	if err != nil {
		return upgradeGroupState{}, err
	}
	info, ok := findUpgradeGroupVestingObject(infoDocument, func(candidate map[string]any) bool {
		return upgradeGroupVestingField(candidate, "id", "group_id", "groupId") != nil &&
			upgradeGroupVestingField(candidate, "admin") != nil &&
			upgradeGroupVestingField(candidate, "version") != nil &&
			upgradeGroupVestingField(candidate, "total_weight", "totalWeight") != nil
	})
	if !ok {
		return upgradeGroupState{}, errors.New("group info query has no group state")
	}
	id, err := upgradeGroupVestingUint64(upgradeGroupVestingField(info, "id", "group_id", "groupId"))
	if err != nil {
		return upgradeGroupState{}, fmt.Errorf("decode group id: %w", err)
	}
	version, err := upgradeGroupVestingUint64(upgradeGroupVestingField(info, "version"))
	if err != nil {
		return upgradeGroupState{}, fmt.Errorf("decode group version: %w", err)
	}
	createdAt, err := decodeUpgradeGroupVestingTimestamp(
		upgradeGroupVestingField(info, "created_at", "createdAt"),
	)
	if err != nil {
		return upgradeGroupState{}, fmt.Errorf("decode group created_at: %w", err)
	}
	state := upgradeGroupState{
		ID:          id,
		Admin:       upgradeGroupVestingText(upgradeGroupVestingField(info, "admin")),
		Metadata:    upgradeGroupVestingText(upgradeGroupVestingField(info, "metadata")),
		Version:     version,
		TotalWeight: upgradeGroupVestingText(upgradeGroupVestingField(info, "total_weight", "totalWeight")),
		CreatedAt:   createdAt,
	}
	if state.ID == 0 || state.Admin == "" || state.Metadata == "" || state.Version == 0 ||
		state.TotalWeight == "" || state.CreatedAt == "" {
		return upgradeGroupState{}, fmt.Errorf("group info is incomplete: %+v", state)
	}
	if fixture.GroupID != 0 && state.ID != fixture.GroupID {
		return upgradeGroupState{}, fmt.Errorf("group id %d, want %d", state.ID, fixture.GroupID)
	}
	if fixture.GroupAdminAddress != "" && state.Admin != fixture.GroupAdminAddress {
		return upgradeGroupState{}, fmt.Errorf("group admin %q, want %q", state.Admin, fixture.GroupAdminAddress)
	}
	if _, err := sdkmath.LegacyNewDecFromStr(state.TotalWeight); err != nil {
		return upgradeGroupState{}, fmt.Errorf("decode group total weight %q: %w", state.TotalWeight, err)
	}

	membersDocument, err := decodeUpgradeGroupVestingJSON(groupMembersRaw, "group members")
	if err != nil {
		return upgradeGroupState{}, err
	}
	membersObject, ok := findUpgradeGroupVestingObject(membersDocument, func(candidate map[string]any) bool {
		_, found := upgradeGroupVestingField(candidate, "members").([]any)
		return found
	})
	if !ok {
		return upgradeGroupState{}, errors.New("group members query has no member list")
	}
	memberValues := upgradeGroupVestingField(membersObject, "members").([]any)
	state.Members = make([]upgradeGroupMemberState, 0, len(memberValues))
	seen := make(map[string]struct{}, len(memberValues))
	totalWeight := sdkmath.LegacyZeroDec()
	for index, value := range memberValues {
		entry, ok := value.(map[string]any)
		if !ok {
			return upgradeGroupState{}, fmt.Errorf("group member %d is not an object", index)
		}
		memberObject, ok := upgradeGroupVestingField(entry, "member").(map[string]any)
		if !ok {
			return upgradeGroupState{}, fmt.Errorf("group member %d has no member state", index)
		}
		memberGroupID, err := upgradeGroupVestingUint64(upgradeGroupVestingField(entry, "group_id", "groupId"))
		if err != nil {
			return upgradeGroupState{}, fmt.Errorf("decode group member %d group id: %w", index, err)
		}
		addedAt, err := decodeUpgradeGroupVestingTimestamp(
			upgradeGroupVestingField(memberObject, "added_at", "addedAt"),
		)
		if err != nil {
			return upgradeGroupState{}, fmt.Errorf("decode group member %d added_at: %w", index, err)
		}
		member := upgradeGroupMemberState{
			GroupID:  memberGroupID,
			Address:  upgradeGroupVestingText(upgradeGroupVestingField(memberObject, "address")),
			Weight:   upgradeGroupVestingText(upgradeGroupVestingField(memberObject, "weight")),
			Metadata: upgradeGroupVestingText(upgradeGroupVestingField(memberObject, "metadata")),
			AddedAt:  addedAt,
		}
		if member.GroupID != state.ID || member.Address == "" || member.Weight == "" || member.AddedAt == "" {
			return upgradeGroupState{}, fmt.Errorf("group member %d is incomplete or belongs to another group: %+v", index, member)
		}
		if _, duplicate := seen[member.Address]; duplicate {
			return upgradeGroupState{}, fmt.Errorf("group members contain duplicate address %q", member.Address)
		}
		weight, err := sdkmath.LegacyNewDecFromStr(member.Weight)
		if err != nil || !weight.IsPositive() {
			return upgradeGroupState{}, fmt.Errorf("group member %s has invalid weight %q", member.Address, member.Weight)
		}
		totalWeight = totalWeight.Add(weight)
		seen[member.Address] = struct{}{}
		state.Members = append(state.Members, member)
	}
	if len(state.Members) == 0 {
		return upgradeGroupState{}, errors.New("group has no members")
	}
	declaredWeight, _ := sdkmath.LegacyNewDecFromStr(state.TotalWeight)
	if !declaredWeight.Equal(totalWeight) {
		return upgradeGroupState{}, fmt.Errorf("group total weight %s does not equal member weight %s", declaredWeight, totalWeight)
	}
	if fixture.InitialMemberAddress != "" {
		if _, found := seen[fixture.InitialMemberAddress]; !found {
			return upgradeGroupState{}, fmt.Errorf("initial group member %s is missing", fixture.InitialMemberAddress)
		}
	}
	sort.Slice(state.Members, func(i, j int) bool { return state.Members[i].Address < state.Members[j].Address })
	return state, nil
}

func decodeUpgradeVestingState(
	accountRaw []byte,
	balancesRaw []byte,
	spendableRaw []byte,
	recipientBalancesRaw []byte,
	fixture upgradeGroupVestingFixture,
) (upgradeVestingState, error) {
	accountDocument, err := decodeUpgradeGroupVestingJSON(accountRaw, "vesting account")
	if err != nil {
		return upgradeVestingState{}, err
	}
	accountEnvelope, ok := findUpgradeGroupVestingObject(accountDocument, func(candidate map[string]any) bool {
		legacyType := upgradeGroupVestingText(upgradeGroupVestingField(candidate, "@type"))
		currentType := upgradeGroupVestingText(upgradeGroupVestingField(candidate, "type"))
		return strings.HasSuffix(legacyType, ".DelayedVestingAccount") ||
			strings.HasSuffix(currentType, ".DelayedVestingAccount")
	})
	if !ok {
		return upgradeVestingState{}, errors.New("auth account query has no delayed vesting account")
	}
	legacyType := upgradeGroupVestingText(upgradeGroupVestingField(accountEnvelope, "@type"))
	currentType := upgradeGroupVestingText(upgradeGroupVestingField(accountEnvelope, "type"))
	if (legacyType == "") == (currentType == "") {
		return upgradeVestingState{}, errors.New("delayed vesting account must use exactly one @type or type encoding")
	}
	accountType := legacyType
	account := accountEnvelope
	if currentType != "" {
		accountType = currentType
		value, ok := upgradeGroupVestingField(accountEnvelope, "value").(map[string]any)
		if !ok {
			return upgradeVestingState{}, errors.New("type/value delayed vesting account has no value object")
		}
		account = value
	}
	baseVesting, ok := upgradeGroupVestingField(account, "base_vesting_account", "baseVestingAccount").(map[string]any)
	if !ok {
		return upgradeVestingState{}, errors.New("delayed vesting account has no base vesting account")
	}
	baseAccount, ok := upgradeGroupVestingField(baseVesting, "base_account", "baseAccount").(map[string]any)
	if !ok {
		return upgradeVestingState{}, errors.New("delayed vesting account has no base account")
	}
	accountNumber, err := decodeUpgradeGroupVestingProtoJSONUint64(baseAccount, "account_number", "accountNumber")
	if err != nil {
		return upgradeVestingState{}, fmt.Errorf("decode vesting account number: %w", err)
	}
	sequence, err := decodeUpgradeGroupVestingProtoJSONUint64(baseAccount, "sequence")
	if err != nil {
		return upgradeVestingState{}, fmt.Errorf("decode vesting account sequence: %w", err)
	}
	endTime, err := upgradeGroupVestingInt64(upgradeGroupVestingField(baseVesting, "end_time", "endTime"))
	if err != nil {
		return upgradeVestingState{}, fmt.Errorf("decode vesting end time: %w", err)
	}
	original, err := decodeUpgradeGroupVestingCoinAmount(
		upgradeGroupVestingField(baseVesting, "original_vesting", "originalVesting"),
		upgradeGroupVestingDenom,
		"original vesting",
	)
	if err != nil {
		return upgradeVestingState{}, err
	}
	delegatedFree, err := decodeUpgradeGroupVestingCoinAmount(
		upgradeGroupVestingField(baseVesting, "delegated_free", "delegatedFree"),
		upgradeGroupVestingDenom,
		"delegated free",
	)
	if err != nil {
		return upgradeVestingState{}, err
	}
	delegatedVesting, err := decodeUpgradeGroupVestingCoinAmount(
		upgradeGroupVestingField(baseVesting, "delegated_vesting", "delegatedVesting"),
		upgradeGroupVestingDenom,
		"delegated vesting",
	)
	if err != nil {
		return upgradeVestingState{}, err
	}
	bankBalance, err := decodeUpgradeGroupVestingBalance(balancesRaw, upgradeGroupVestingDenom, "vesting bank balance")
	if err != nil {
		return upgradeVestingState{}, err
	}
	spendableBalance, err := decodeUpgradeGroupVestingBalance(spendableRaw, upgradeGroupVestingDenom, "vesting spendable balance")
	if err != nil {
		return upgradeVestingState{}, err
	}
	recipientBalance, err := decodeUpgradeGroupVestingBalance(recipientBalancesRaw, upgradeGroupVestingDenom, "vesting recipient balance")
	if err != nil {
		return upgradeVestingState{}, err
	}
	originalAmount, _ := sdkmath.NewIntFromString(original)
	delegatedVestingAmount, _ := sdkmath.NewIntFromString(delegatedVesting)
	lockedAmount := originalAmount.Sub(delegatedVestingAmount)
	if lockedAmount.IsNegative() {
		return upgradeVestingState{}, fmt.Errorf("delegated vesting %s exceeds original vesting %s", delegatedVesting, original)
	}
	bankAmount, _ := sdkmath.NewIntFromString(bankBalance)
	if lockedAmount.GT(bankAmount) {
		lockedAmount = bankAmount
	}
	freeAmount := bankAmount.Sub(lockedAmount)
	state := upgradeVestingState{
		AccountType:      accountType,
		Address:          upgradeGroupVestingText(upgradeGroupVestingField(baseAccount, "address")),
		AccountNumber:    accountNumber,
		Sequence:         sequence,
		OriginalVesting:  original,
		DelegatedFree:    delegatedFree,
		DelegatedVesting: delegatedVesting,
		EndTimeUnix:      endTime,
		BankBalance:      bankBalance,
		LockedBalance:    lockedAmount.String(),
		FreeBalance:      freeAmount.String(),
		SpendableBalance: spendableBalance,
		RecipientBalance: recipientBalance,
	}
	if state.Address == "" || state.AccountType == "" || !strings.HasSuffix(state.AccountType, ".DelayedVestingAccount") {
		return upgradeVestingState{}, fmt.Errorf("vesting account identity is incomplete: %+v", state)
	}
	if fixture.VestingAccountAddress != "" && state.Address != fixture.VestingAccountAddress {
		return upgradeVestingState{}, fmt.Errorf("vesting account address %q, want %q", state.Address, fixture.VestingAccountAddress)
	}
	if fixture.VestingOriginalAmount != "" && state.OriginalVesting != fixture.VestingOriginalAmount {
		return upgradeVestingState{}, fmt.Errorf("original vesting amount %s, want %s", state.OriginalVesting, fixture.VestingOriginalAmount)
	}
	if fixture.VestingEndTimeUnix != 0 && state.EndTimeUnix != fixture.VestingEndTimeUnix {
		return upgradeVestingState{}, fmt.Errorf("vesting end time %d, want %d", state.EndTimeUnix, fixture.VestingEndTimeUnix)
	}
	if !freeAmount.Equal(mustUpgradeGroupVestingInt(spendableBalance)) {
		return upgradeVestingState{}, fmt.Errorf("vesting free balance %s does not equal spendable balance %s", freeAmount, spendableBalance)
	}
	return state, nil
}

func decodeUpgradeGroupVestingBalance(raw []byte, denom, label string) (string, error) {
	document, err := decodeUpgradeGroupVestingJSON(raw, label)
	if err != nil {
		return "", err
	}
	object, ok := findUpgradeGroupVestingObject(document, func(candidate map[string]any) bool {
		_, found := upgradeGroupVestingField(candidate, "balances").([]any)
		return found
	})
	if !ok {
		return "", fmt.Errorf("%s query has no balances", label)
	}
	return decodeUpgradeGroupVestingCoinAmount(upgradeGroupVestingField(object, "balances"), denom, label)
}

func decodeUpgradeGroupVestingCoinAmount(value any, denom, label string) (string, error) {
	if value == nil {
		return "0", nil
	}
	coins, ok := value.([]any)
	if !ok {
		return "", fmt.Errorf("%s coin list has type %T", label, value)
	}
	amount := "0"
	found := false
	seen := make(map[string]struct{}, len(coins))
	for index, value := range coins {
		coin, ok := value.(map[string]any)
		if !ok {
			return "", fmt.Errorf("%s coin %d is not an object", label, index)
		}
		coinDenom := upgradeGroupVestingText(upgradeGroupVestingField(coin, "denom"))
		coinAmount := upgradeGroupVestingText(upgradeGroupVestingField(coin, "amount"))
		if coinDenom == "" || coinAmount == "" {
			return "", fmt.Errorf("%s coin %d is incomplete", label, index)
		}
		if _, duplicate := seen[coinDenom]; duplicate {
			return "", fmt.Errorf("%s contains duplicate denom %q", label, coinDenom)
		}
		parsed, parsedOK := sdkmath.NewIntFromString(coinAmount)
		if !parsedOK || parsed.IsNegative() {
			return "", fmt.Errorf("%s has invalid amount %q for %s", label, coinAmount, coinDenom)
		}
		seen[coinDenom] = struct{}{}
		if coinDenom == denom {
			amount = parsed.String()
			found = true
		}
	}
	if !found {
		return "0", nil
	}
	return amount, nil
}

func decodeUpgradeGroupVestingJSON(raw []byte, label string) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var document any
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("decode %s JSON: %w", label, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err == nil {
		return nil, fmt.Errorf("decode %s JSON: multiple values", label)
	} else if !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("decode %s JSON trailing data: %w", label, err)
	}
	return document, nil
}

func findUpgradeGroupVestingObject(value any, match func(map[string]any) bool) (map[string]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		if match(typed) {
			return typed, true
		}
		for _, child := range typed {
			if found, ok := findUpgradeGroupVestingObject(child, match); ok {
				return found, true
			}
		}
	case []any:
		for _, child := range typed {
			if found, ok := findUpgradeGroupVestingObject(child, match); ok {
				return found, true
			}
		}
	}
	return nil, false
}

func upgradeGroupVestingField(object map[string]any, names ...string) any {
	for _, name := range names {
		if value, ok := object[name]; ok {
			return value
		}
	}
	return nil
}

func upgradeGroupVestingText(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	default:
		return ""
	}
}

func decodeUpgradeGroupVestingTimestamp(value any) (string, error) {
	raw, ok := value.(string)
	if !ok || raw == "" {
		return "", fmt.Errorf("expected RFC3339 timestamp string, got %T", value)
	}
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return "", fmt.Errorf("parse RFC3339 timestamp %q: %w", raw, err)
	}
	return parsed.UTC().Format(time.RFC3339Nano), nil
}

func upgradeGroupVestingUint64(value any) (uint64, error) {
	text := upgradeGroupVestingText(value)
	if text == "" {
		return 0, fmt.Errorf("expected string or number, got %T", value)
	}
	return strconv.ParseUint(text, 10, 64)
}

// ProtoJSON omits uint64 fields whose value is zero. A missing base-account
// scalar therefore means zero, while a present null, boolean, object, or list
// is malformed evidence and must not be normalized into a healthy account.
func decodeUpgradeGroupVestingProtoJSONUint64(object map[string]any, names ...string) (uint64, error) {
	for _, name := range names {
		value, present := object[name]
		if !present {
			continue
		}
		decoded, err := upgradeGroupVestingUint64(value)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", name, err)
		}
		return decoded, nil
	}
	return 0, nil
}

func upgradeGroupVestingInt64(value any) (int64, error) {
	text := upgradeGroupVestingText(value)
	if text == "" {
		return 0, fmt.Errorf("expected string or number, got %T", value)
	}
	return strconv.ParseInt(text, 10, 64)
}

func mustUpgradeGroupVestingInt(value string) sdkmath.Int {
	parsed, ok := sdkmath.NewIntFromString(value)
	if !ok {
		panic("validated group/vesting integer became invalid")
	}
	return parsed
}
