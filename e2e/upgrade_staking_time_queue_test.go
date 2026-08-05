package e2e_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	upgradeP0QueueInitialDelegation = "300000000"
	upgradeP0QueueFundAmount        = "500000000"
	upgradeP0RedelegationNotFound   = "redelegation not found for delegator address"
)

type upgradeP0StakingQueueFixture struct {
	UnbondKeyName        string   `json:"unbond_key_name"`
	UnbondAddress        string   `json:"unbond_address"`
	RedelegateKeyName    string   `json:"redelegate_key_name"`
	RedelegateAddress    string   `json:"redelegate_address"`
	SourceValidator      string   `json:"source_validator"`
	DestinationValidator string   `json:"destination_validator"`
	PreparationTxHashes  []string `json:"preparation_tx_hashes"`
}

func prepareUpgradeP0StakingQueue(
	ctx context.Context,
	network *harness.Network,
) (upgradeP0StakingQueueFixture, upgradeP0StakingQueueCheckpoint, error) {
	if err := validateUpgradeStakingNetwork(network, 1); err != nil {
		return upgradeP0StakingQueueFixture{}, upgradeP0StakingQueueCheckpoint{}, err
	}
	unbondWallet, err := network.BuildWallet(ctx, "upgrade-p0-unbond-delegator", "")
	if err != nil {
		return upgradeP0StakingQueueFixture{}, upgradeP0StakingQueueCheckpoint{}, err
	}
	redelegateWallet, err := network.BuildWallet(ctx, "upgrade-p0-redelegate-delegator", "")
	if err != nil {
		return upgradeP0StakingQueueFixture{}, upgradeP0StakingQueueCheckpoint{}, err
	}
	sourceValidator, err := queryUpgradeValidatorOperator(ctx, network.Chain.Validators[0])
	if err != nil {
		return upgradeP0StakingQueueFixture{}, upgradeP0StakingQueueCheckpoint{}, err
	}
	destinationValidator, err := queryUpgradeValidatorOperator(ctx, network.Chain.Validators[1])
	if err != nil {
		return upgradeP0StakingQueueFixture{}, upgradeP0StakingQueueCheckpoint{}, err
	}
	if sourceValidator == destinationValidator {
		return upgradeP0StakingQueueFixture{}, upgradeP0StakingQueueCheckpoint{}, errors.New("staking queue validators must be distinct")
	}
	fixture := upgradeP0StakingQueueFixture{
		UnbondKeyName:        unbondWallet.KeyName(),
		UnbondAddress:        unbondWallet.FormattedAddress(),
		RedelegateKeyName:    redelegateWallet.KeyName(),
		RedelegateAddress:    redelegateWallet.FormattedAddress(),
		SourceValidator:      sourceValidator,
		DestinationValidator: destinationValidator,
	}
	txNode := network.Chain.Validators[0]
	for index, wallet := range []struct {
		key     string
		address string
	}{
		{key: fixture.UnbondKeyName, address: fixture.UnbondAddress},
		{key: fixture.RedelegateKeyName, address: fixture.RedelegateAddress},
	} {
		funded, fundErr := network.BroadcastAndWaitTx(
			ctx,
			"upgrade-p0-queue-fund-"+strconv.Itoa(index),
			txNode,
			interchaintest.FaucetAccountKeyName,
			"bank", "send", interchaintest.FaucetAccountKeyName, wallet.address,
			upgradeP0QueueFundAmount+upgradeStakingDenom,
			"--gas", "500000",
			"--broadcast-mode", "sync",
		)
		if fundErr != nil {
			return fixture, upgradeP0StakingQueueCheckpoint{}, fmt.Errorf("fund P0 queue delegator %d: %w", index, fundErr)
		}
		delegated, delegateErr := network.BroadcastAndWaitTx(
			ctx,
			"upgrade-p0-queue-delegate-"+strconv.Itoa(index),
			txNode,
			wallet.key,
			"staking", "delegate", fixture.SourceValidator,
			upgradeP0QueueInitialDelegation+upgradeStakingDenom,
			"--fees", upgradeP0QueueFee+upgradeStakingDenom,
			"--gas", "500000",
			"--broadcast-mode", "sync",
		)
		if delegateErr != nil {
			return fixture, upgradeP0StakingQueueCheckpoint{}, fmt.Errorf("prepare P0 queue delegation %d: %w", index, delegateErr)
		}
		fixture.PreparationTxHashes = append(fixture.PreparationTxHashes, funded.TxHash, delegated.TxHash)
	}
	if err := waitUpgradeStakingBlocks(ctx, network, 2); err != nil {
		return fixture, upgradeP0StakingQueueCheckpoint{}, err
	}
	before, err := captureUpgradeP0StakingQueueCheckpoint(ctx, network, fixture, "before")
	if err != nil {
		return fixture, upgradeP0StakingQueueCheckpoint{}, err
	}
	if len(before.Unbonding) != 0 || len(before.Redelegation) != 0 {
		return fixture, before, errors.New("dedicated P0 staking queue accounts unexpectedly have pending entries")
	}
	if err := network.WriteArtifactJSON("upgrade/staking-time-queue/preparation.json", map[string]any{
		"fixture": fixture,
		"before":  before,
	}); err != nil {
		return fixture, before, err
	}
	return fixture, before, nil
}

func beginUpgradeP0StakingQueues(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeP0StakingQueueFixture,
	before upgradeP0StakingQueueCheckpoint,
) (upgradeP0StakingQueueEvidence, error) {
	evidence := upgradeP0StakingQueueEvidence{
		SourceValidator:      fixture.SourceValidator,
		DestinationValidator: fixture.DestinationValidator,
		UnbondAmount:         upgradeP0QueueUnbondAmount,
		RedelegateAmount:     upgradeP0QueueRedelegateAmount,
		FeePerTx:             upgradeP0QueueFee,
		Before:               before,
	}
	txNode := network.Chain.Validators[0]
	unbonded, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-p0-queue-unbond",
		txNode,
		fixture.UnbondKeyName,
		"staking", "unbond", fixture.SourceValidator,
		upgradeP0QueueUnbondAmount+upgradeStakingDenom,
		"--fees", upgradeP0QueueFee+upgradeStakingDenom,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return evidence, fmt.Errorf("begin P0 unbonding queue: %w", err)
	}
	unbondReward, err := decodeUpgradeP0WithdrawnReward(
		unbonded,
		fixture.UnbondAddress,
		fixture.SourceValidator,
	)
	if err != nil {
		return evidence, fmt.Errorf("capture P0 unbonding automatic reward withdrawal: %w", err)
	}
	redelegated, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-p0-queue-redelegate",
		txNode,
		fixture.RedelegateKeyName,
		"staking", "redelegate", fixture.SourceValidator, fixture.DestinationValidator,
		upgradeP0QueueRedelegateAmount+upgradeStakingDenom,
		"--fees", upgradeP0QueueFee+upgradeStakingDenom,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return evidence, fmt.Errorf("begin P0 redelegation queue: %w", err)
	}
	redelegateReward, err := decodeUpgradeP0WithdrawnReward(
		redelegated,
		fixture.RedelegateAddress,
		fixture.SourceValidator,
	)
	if err != nil {
		return evidence, fmt.Errorf("capture P0 redelegation automatic reward withdrawal: %w", err)
	}
	evidence.UnbondWithdrawnReward = unbondReward
	evidence.RedelegateWithdrawnReward = redelegateReward
	evidence.UnbondTxHash = unbonded.TxHash
	evidence.RedelegateTxHash = redelegated.TxHash
	queued, err := captureUpgradeP0StakingQueueCheckpoint(ctx, network, fixture, "queued")
	if err != nil {
		return evidence, err
	}
	evidence.Queued = queued
	if len(queued.Unbonding) != 1 || len(queued.Redelegation) != 1 {
		return evidence, errors.New("staking queue transactions did not create exactly one entry each")
	}
	if err := network.WriteArtifactJSON("upgrade/staking-time-queue/queued.json", evidence); err != nil {
		return evidence, err
	}
	return evidence, nil
}

func decodeUpgradeP0WithdrawnReward(
	result *harness.TxResult,
	expectedDelegator string,
	expectedValidator string,
) (string, error) {
	if result == nil {
		return "", errors.New("withdrawn reward requires a committed transaction result")
	}
	matching := make([]harness.TxEvent, 0, 1)
	for _, event := range result.Events {
		if event.Type != "withdraw_rewards" ||
			event.Attribute("delegator") != expectedDelegator ||
			event.Attribute("validator") != expectedValidator {
			continue
		}
		matching = append(matching, event)
	}
	if len(matching) != 1 {
		return "", fmt.Errorf(
			"withdrawn reward requires exactly one event for delegator %s and validator %s, got %d",
			expectedDelegator,
			expectedValidator,
			len(matching),
		)
	}
	coin, err := sdk.ParseCoinNormalized(matching[0].Attribute("amount"))
	if err != nil {
		return "", fmt.Errorf("decode withdrawn reward amount: %w", err)
	}
	if coin.Denom != upgradeStakingDenom {
		return "", fmt.Errorf("withdrawn reward denom %q, want %q", coin.Denom, upgradeStakingDenom)
	}
	if coin.Amount.IsNegative() {
		return "", fmt.Errorf("withdrawn reward amount must not be negative: %s", coin.Amount)
	}
	return coin.Amount.String(), nil
}

func captureUpgradeP0StakingQueuesPendingAfterUpgrade(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeP0StakingQueueFixture,
	evidence *upgradeP0StakingQueueEvidence,
) error {
	checkpoint, err := captureUpgradeP0StakingQueueCheckpoint(ctx, network, fixture, "post-upgrade-pending")
	if err != nil {
		return err
	}
	evidence.PostUpgradePending = checkpoint
	if len(checkpoint.Unbonding) != 1 || len(checkpoint.Redelegation) != 1 ||
		!reflectUpgradeP0QueueEntriesEqual(evidence.Queued, checkpoint) {
		return errors.New("staking time queues were not preserved pending across the binary upgrade")
	}
	return network.WriteArtifactJSON("upgrade/staking-time-queue/post-upgrade-pending.json", evidence)
}

func completeUpgradeP0StakingQueues(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeP0StakingQueueFixture,
	evidence *upgradeP0StakingQueueEvidence,
) error {
	if len(evidence.PostUpgradePending.Unbonding) != 1 || len(evidence.PostUpgradePending.Redelegation) != 1 {
		return errors.New("cannot wait for unobserved post-upgrade staking queues")
	}
	completion := evidence.PostUpgradePending.Unbonding[0].CompletionTime
	if candidate := evidence.PostUpgradePending.Redelegation[0].CompletionTime; candidate.After(completion) {
		completion = candidate
	}
	wait := time.Until(completion.Add(2 * time.Second))
	if wait > 0 {
		timer := time.NewTimer(wait)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for staking queue completion time %s: %w", completion, ctx.Err())
		case <-timer.C:
		}
	}
	if err := waitUpgradeStakingBlocks(ctx, network, 2); err != nil {
		return err
	}
	completed, err := captureUpgradeP0StakingQueueCheckpoint(ctx, network, fixture, "completed")
	if err != nil {
		return err
	}
	evidence.Completed = completed
	if len(completed.Unbonding) != 0 || len(completed.Redelegation) != 0 {
		return errors.New("staking time queue entries remain after completion time")
	}
	return network.WriteArtifactJSON("upgrade/staking-time-queue/completed.json", evidence)
}

func captureUpgradeP0StakingQueuesAfterRestart(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeP0StakingQueueFixture,
	evidence *upgradeP0StakingQueueEvidence,
) error {
	checkpoint, err := captureUpgradeP0StakingQueueCheckpoint(ctx, network, fixture, "post-restart")
	if err != nil {
		return err
	}
	evidence.PostRestart = checkpoint
	if err := validateUpgradeP0StakingQueueEvidence(*evidence); err != nil {
		return err
	}
	return network.WriteArtifactJSON("upgrade/staking-time-queue/post-restart.json", evidence)
}

func reflectUpgradeP0QueueEntriesEqual(left, right upgradeP0StakingQueueCheckpoint) bool {
	leftJSON, leftErr := json.Marshal(struct {
		Unbonding    []upgradeP0QueueEntry `json:"unbonding"`
		Redelegation []upgradeP0QueueEntry `json:"redelegation"`
	}{left.Unbonding, left.Redelegation})
	rightJSON, rightErr := json.Marshal(struct {
		Unbonding    []upgradeP0QueueEntry `json:"unbonding"`
		Redelegation []upgradeP0QueueEntry `json:"redelegation"`
	}{right.Unbonding, right.Redelegation})
	return leftErr == nil && rightErr == nil && string(leftJSON) == string(rightJSON)
}

func captureUpgradeP0StakingQueueCheckpoint(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeP0StakingQueueFixture,
	phase string,
) (upgradeP0StakingQueueCheckpoint, error) {
	if network == nil || network.Chain == nil || len(network.Chain.FullNodes) == 0 {
		return upgradeP0StakingQueueCheckpoint{}, errors.New("staking queue checkpoint requires a full node")
	}
	fullNode := network.Chain.FullNodes[0]
	height, err := fullNode.Height(ctx)
	if err != nil {
		return upgradeP0StakingQueueCheckpoint{}, err
	}
	query := func(suffix string, command ...string) (json.RawMessage, error) {
		command = append(command, "--height", strconv.FormatInt(height, 10))
		return network.FullNodeCLIQuery(ctx, "upgrade-p0-queue-"+phase+"-"+suffix, command...)
	}
	unbondDelegationsRaw, err := query("unbond-delegations", "staking", "delegations", fixture.UnbondAddress)
	if err != nil {
		return upgradeP0StakingQueueCheckpoint{}, err
	}
	redelegateDelegationsRaw, err := query("redelegate-delegations", "staking", "delegations", fixture.RedelegateAddress)
	if err != nil {
		return upgradeP0StakingQueueCheckpoint{}, err
	}
	unbondingRaw, err := query("unbonding", "staking", "unbonding-delegations", fixture.UnbondAddress)
	if err != nil {
		return upgradeP0StakingQueueCheckpoint{}, err
	}
	redelegationRequiresEntry, err := upgradeP0RedelegationPhaseRequiresEntry(phase)
	if err != nil {
		return upgradeP0StakingQueueCheckpoint{}, err
	}
	var redelegationRaw json.RawMessage
	redelegationArgs := upgradeP0RedelegationQueryArgs(fixture)
	if redelegationRequiresEntry {
		redelegationRaw, err = query("redelegation", redelegationArgs...)
	} else {
		redelegationArgs = append(redelegationArgs, "--height", strconv.FormatInt(height, 10))
		err = network.FullNodeGRPCQueryExpectedError(
			ctx,
			"upgrade-p0-queue-"+phase+"-redelegation-absent",
			upgradeP0RedelegationNotFound,
			redelegationArgs...,
		)
		redelegationRaw = json.RawMessage(`{"redelegation_responses":[]}`)
	}
	if err != nil {
		return upgradeP0StakingQueueCheckpoint{}, err
	}
	poolRaw, err := query("pool", "staking", "pool")
	if err != nil {
		return upgradeP0StakingQueueCheckpoint{}, err
	}
	unbondDelegations, err := decodeUpgradeP0DelegationBalances(unbondDelegationsRaw)
	if err != nil {
		return upgradeP0StakingQueueCheckpoint{}, err
	}
	redelegateDelegations, err := decodeUpgradeP0DelegationBalances(redelegateDelegationsRaw)
	if err != nil {
		return upgradeP0StakingQueueCheckpoint{}, err
	}
	sourceBalance := new(big.Int)
	destinationBalance := new(big.Int)
	for _, delegation := range append(unbondDelegations, redelegateDelegations...) {
		amount, parseErr := nonNegativeUpgradeP0Amount("delegation balance", delegation.Balance.Amount)
		if parseErr != nil {
			return upgradeP0StakingQueueCheckpoint{}, parseErr
		}
		switch delegation.ValidatorOperator {
		case fixture.SourceValidator:
			sourceBalance.Add(sourceBalance, amount)
		case fixture.DestinationValidator:
			destinationBalance.Add(destinationBalance, amount)
		}
	}
	unbondEntries, err := decodeUpgradeP0UnbondingEntries(unbondingRaw, fixture.UnbondAddress, fixture.SourceValidator)
	if err != nil {
		return upgradeP0StakingQueueCheckpoint{}, err
	}
	redelegationEntries, err := decodeUpgradeP0RedelegationEntries(
		redelegationRaw,
		fixture.RedelegateAddress,
		fixture.SourceValidator,
		fixture.DestinationValidator,
	)
	if err != nil {
		return upgradeP0StakingQueueCheckpoint{}, err
	}
	pool, err := decodeUpgradeStakingPool(poolRaw)
	if err != nil {
		return upgradeP0StakingQueueCheckpoint{}, err
	}
	pinned := harness.ContextAtHeight(ctx, height)
	unbondBank, err := network.QueryFullNodeBalance(pinned, fixture.UnbondAddress, upgradeStakingDenom)
	if err != nil {
		return upgradeP0StakingQueueCheckpoint{}, err
	}
	redelegateBank, err := network.QueryFullNodeBalance(pinned, fixture.RedelegateAddress, upgradeStakingDenom)
	if err != nil {
		return upgradeP0StakingQueueCheckpoint{}, err
	}
	checkpoint := upgradeP0StakingQueueCheckpoint{
		Phase:                        phase,
		Height:                       height,
		RecordedAt:                   time.Now().UTC(),
		BankBalance:                  new(big.Int).Add(unbondBank.BigInt(), redelegateBank.BigInt()).String(),
		SourceDelegationBalance:      sourceBalance.String(),
		DestinationDelegationBalance: destinationBalance.String(),
		Pool:                         pool,
		Unbonding:                    unbondEntries,
		Redelegation:                 redelegationEntries,
	}
	if err := network.WriteArtifactJSON("upgrade/staking-time-queue/checkpoints/"+phase+".json", checkpoint); err != nil {
		return checkpoint, err
	}
	return checkpoint, nil
}

func upgradeP0RedelegationQueryArgs(fixture upgradeP0StakingQueueFixture) []string {
	return []string{
		"staking",
		"redelegation",
		fixture.RedelegateAddress,
		fixture.SourceValidator,
		fixture.DestinationValidator,
	}
}

func upgradeP0RedelegationPhaseRequiresEntry(phase string) (bool, error) {
	switch phase {
	case "queued", "post-upgrade-pending":
		return true, nil
	case "before", "completed", "post-restart":
		return false, nil
	default:
		return false, fmt.Errorf("unsupported staking queue phase %q", phase)
	}
}

func decodeUpgradeP0DelegationBalances(raw []byte) ([]upgradeDelegationState, error) {
	document, err := decodeUpgradeJSON(raw, "staking delegations")
	if err != nil {
		return nil, err
	}
	root, ok := document.(map[string]any)
	if !ok {
		return nil, errors.New("staking delegations response is not an object")
	}
	responses, ok := root["delegation_responses"].([]any)
	if !ok {
		return nil, errors.New("staking delegations response has no delegation_responses array")
	}
	delegations := make([]upgradeDelegationState, 0, len(responses))
	for index, response := range responses {
		encoded, marshalErr := json.Marshal(response)
		if marshalErr != nil {
			return nil, marshalErr
		}
		delegation, decodeErr := decodeUpgradeDelegation(encoded)
		if decodeErr != nil {
			return nil, fmt.Errorf("decode delegation response %d: %w", index, decodeErr)
		}
		delegations = append(delegations, delegation)
	}
	return delegations, nil
}

func decodeUpgradeP0UnbondingEntries(raw []byte, delegator, validator string) ([]upgradeP0QueueEntry, error) {
	document, err := decodeUpgradeJSON(raw, "unbonding delegations")
	if err != nil {
		return nil, err
	}
	root, ok := document.(map[string]any)
	if !ok {
		return nil, errors.New("unbonding delegations response is not an object")
	}
	rawResponses, exists := root["unbonding_responses"]
	if !exists {
		return nil, errors.New("unbonding delegations response has no unbonding_responses array")
	}
	if rawResponses == nil {
		return nil, nil
	}
	responses, ok := rawResponses.([]any)
	if !ok {
		return nil, errors.New("unbonding delegations response has malformed unbonding_responses")
	}
	var entries []upgradeP0QueueEntry
	for _, rawResponse := range responses {
		response, ok := rawResponse.(map[string]any)
		if !ok || jsonText(response["delegator_address"]) != delegator || jsonText(response["validator_address"]) != validator {
			continue
		}
		rawEntries, ok := response["entries"].([]any)
		if !ok {
			return nil, errors.New("matching unbonding response has no entries")
		}
		for _, rawEntry := range rawEntries {
			entryObject, ok := rawEntry.(map[string]any)
			if !ok {
				return nil, errors.New("unbonding entry is not an object")
			}
			entry, decodeErr := decodeUpgradeP0QueueEntry(entryObject, "")
			if decodeErr != nil {
				return nil, decodeErr
			}
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func decodeUpgradeP0RedelegationEntries(raw []byte, delegator, source, destination string) ([]upgradeP0QueueEntry, error) {
	document, err := decodeUpgradeJSON(raw, "redelegations")
	if err != nil {
		return nil, err
	}
	root, ok := document.(map[string]any)
	if !ok {
		return nil, errors.New("redelegations response is not an object")
	}
	responses, ok := root["redelegation_responses"].([]any)
	if !ok {
		return nil, errors.New("redelegations response has no redelegation_responses array")
	}
	var entries []upgradeP0QueueEntry
	for _, rawResponse := range responses {
		response, ok := rawResponse.(map[string]any)
		if !ok {
			continue
		}
		redelegation, ok := response["redelegation"].(map[string]any)
		if !ok || jsonText(redelegation["delegator_address"]) != delegator ||
			jsonText(redelegation["validator_src_address"]) != source ||
			jsonText(redelegation["validator_dst_address"]) != destination {
			continue
		}
		rawEntries, ok := response["entries"].([]any)
		if !ok {
			return nil, errors.New("matching redelegation response has no entries")
		}
		for _, rawEntry := range rawEntries {
			entryWrapper, ok := rawEntry.(map[string]any)
			if !ok {
				return nil, errors.New("redelegation entry wrapper is not an object")
			}
			entryObject, ok := entryWrapper["redelegation_entry"].(map[string]any)
			if !ok {
				return nil, errors.New("redelegation entry wrapper has no redelegation_entry")
			}
			entry, decodeErr := decodeUpgradeP0QueueEntry(entryObject, jsonText(entryWrapper["balance"]))
			if decodeErr != nil {
				return nil, decodeErr
			}
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func decodeUpgradeP0QueueEntry(object map[string]any, balanceOverride string) (upgradeP0QueueEntry, error) {
	creationHeight, err := jsonInt64(object["creation_height"])
	if err != nil || creationHeight <= 0 {
		return upgradeP0QueueEntry{}, errors.New("staking queue entry has invalid creation_height")
	}
	completionTime, err := time.Parse(time.RFC3339Nano, jsonText(object["completion_time"]))
	if err != nil {
		return upgradeP0QueueEntry{}, fmt.Errorf("decode staking queue completion_time: %w", err)
	}
	entry := upgradeP0QueueEntry{
		CreationHeight:    creationHeight,
		CompletionTime:    completionTime,
		InitialBalance:    jsonText(object["initial_balance"]),
		Balance:           jsonText(object["balance"]),
		SharesDestination: jsonText(object["shares_dst"]),
	}
	if balanceOverride != "" {
		entry.Balance = balanceOverride
	}
	if _, err := positiveUpgradeP0Amount("staking queue initial balance", entry.InitialBalance); err != nil {
		return upgradeP0QueueEntry{}, err
	}
	if _, err := positiveUpgradeP0Amount("staking queue balance", entry.Balance); err != nil {
		return upgradeP0QueueEntry{}, err
	}
	return entry, nil
}
