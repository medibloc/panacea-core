package e2e_test

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	upgradeP0QueueUnbondAmount     = "100000000"
	upgradeP0QueueRedelegateAmount = "100000000"
	upgradeP0QueueFee              = "2500000"
)

type upgradeP0QueueEntry struct {
	CreationHeight    int64     `json:"creation_height"`
	CompletionTime    time.Time `json:"completion_time"`
	InitialBalance    string    `json:"initial_balance"`
	Balance           string    `json:"balance"`
	SharesDestination string    `json:"shares_destination,omitempty"`
}

type upgradeP0StakingQueueCheckpoint struct {
	Phase                        string                  `json:"phase"`
	Height                       int64                   `json:"height"`
	RecordedAt                   time.Time               `json:"recorded_at"`
	BankBalance                  string                  `json:"bank_balance"`
	SourceDelegationBalance      string                  `json:"source_delegation_balance"`
	DestinationDelegationBalance string                  `json:"destination_delegation_balance"`
	Pool                         upgradeStakingPoolState `json:"pool"`
	Unbonding                    []upgradeP0QueueEntry   `json:"unbonding,omitempty"`
	Redelegation                 []upgradeP0QueueEntry   `json:"redelegation,omitempty"`
}

type upgradeP0StakingQueueEvidence struct {
	SourceValidator           string                          `json:"source_validator"`
	DestinationValidator      string                          `json:"destination_validator"`
	UnbondAmount              string                          `json:"unbond_amount"`
	RedelegateAmount          string                          `json:"redelegate_amount"`
	FeePerTx                  string                          `json:"fee_per_tx"`
	UnbondWithdrawnReward     string                          `json:"unbond_withdrawn_reward"`
	RedelegateWithdrawnReward string                          `json:"redelegate_withdrawn_reward"`
	UnbondTxHash              string                          `json:"unbond_tx_hash"`
	RedelegateTxHash          string                          `json:"redelegate_tx_hash"`
	Before                    upgradeP0StakingQueueCheckpoint `json:"before"`
	Queued                    upgradeP0StakingQueueCheckpoint `json:"queued"`
	PostUpgradePending        upgradeP0StakingQueueCheckpoint `json:"post_upgrade_pending"`
	Completed                 upgradeP0StakingQueueCheckpoint `json:"completed"`
	PostRestart               upgradeP0StakingQueueCheckpoint `json:"post_restart"`
}

type upgradeP0SlashingCheckpoint struct {
	Phase          string                  `json:"phase"`
	Height         int64                   `json:"height"`
	RecordedAt     time.Time               `json:"recorded_at"`
	ValidatorPower int64                   `json:"validator_power"`
	Validator      upgradeValidatorState   `json:"validator"`
	SigningInfo    upgradeSigningInfoState `json:"signing_info"`
}

type upgradeP0SlashingEvidence struct {
	ValidatorIndex       int                         `json:"validator_index"`
	UpgradeHeight        int64                       `json:"upgrade_height"`
	StoppedAt            int64                       `json:"stopped_at_height"`
	OutageBlocksObserved int64                       `json:"outage_blocks_observed"`
	MissedBlocksObserved int64                       `json:"missed_blocks_observed"`
	Before               upgradeP0SlashingCheckpoint `json:"before"`
	Jailed               upgradeP0SlashingCheckpoint `json:"jailed"`
	EarlyUnjail          harness.TxResult            `json:"early_unjail"`
	UnjailTxHash         string                      `json:"unjail_tx_hash"`
	Unjailed             upgradeP0SlashingCheckpoint `json:"unjailed"`
	Rejoined             upgradeP0SlashingCheckpoint `json:"rejoined"`
	SignedCommitHeight   int64                       `json:"signed_commit_height"`
	PostRestart          upgradeP0SlashingCheckpoint `json:"post_restart"`
}

func validateUpgradeP0StakingQueueEvidence(evidence upgradeP0StakingQueueEvidence) error {
	unbond, err := positiveUpgradeP0Amount("unbond", evidence.UnbondAmount)
	if err != nil {
		return err
	}
	redelegate, err := positiveUpgradeP0Amount("redelegate", evidence.RedelegateAmount)
	if err != nil {
		return err
	}
	fee, err := positiveUpgradeP0Amount("fee", evidence.FeePerTx)
	if err != nil {
		return err
	}
	unbondReward, err := nonNegativeUpgradeP0Amount("unbond withdrawn reward", evidence.UnbondWithdrawnReward)
	if err != nil {
		return err
	}
	redelegateReward, err := nonNegativeUpgradeP0Amount("redelegate withdrawn reward", evidence.RedelegateWithdrawnReward)
	if err != nil {
		return err
	}
	checkpoints := []upgradeP0StakingQueueCheckpoint{
		evidence.Before,
		evidence.Queued,
		evidence.PostUpgradePending,
		evidence.Completed,
		evidence.PostRestart,
	}
	for index, checkpoint := range checkpoints {
		if checkpoint.Height <= 0 || checkpoint.RecordedAt.IsZero() || strings.TrimSpace(checkpoint.Phase) == "" {
			return fmt.Errorf("staking queue checkpoint %d is incomplete", index)
		}
		if index > 0 && checkpoint.Height <= checkpoints[index-1].Height {
			return fmt.Errorf("staking queue checkpoint heights do not advance at %s", checkpoint.Phase)
		}
		for label, amount := range map[string]string{
			"bank":                   checkpoint.BankBalance,
			"source delegation":      checkpoint.SourceDelegationBalance,
			"destination delegation": checkpoint.DestinationDelegationBalance,
			"bonded pool":            checkpoint.Pool.BondedTokens,
			"not-bonded pool":        checkpoint.Pool.NotBondedTokens,
		} {
			if _, parseErr := nonNegativeUpgradeP0Amount(label, amount); parseErr != nil {
				return fmt.Errorf("%s checkpoint: %w", checkpoint.Phase, parseErr)
			}
		}
	}
	if len(evidence.Queued.Unbonding) != 1 || len(evidence.Queued.Redelegation) != 1 ||
		len(evidence.PostUpgradePending.Unbonding) != 1 || len(evidence.PostUpgradePending.Redelegation) != 1 {
		return errors.New("staking unbonding and redelegation must remain pending across the upgrade")
	}
	if len(evidence.Completed.Unbonding) != 0 || len(evidence.Completed.Redelegation) != 0 ||
		len(evidence.PostRestart.Unbonding) != 0 || len(evidence.PostRestart.Redelegation) != 0 {
		return errors.New("completed staking queues must remain absent after completion and restart")
	}
	if !reflect.DeepEqual(evidence.Queued.Unbonding, evidence.PostUpgradePending.Unbonding) ||
		!reflect.DeepEqual(evidence.Queued.Redelegation, evidence.PostUpgradePending.Redelegation) {
		return errors.New("pending staking queue entries changed across the upgrade")
	}
	unbondEntry := evidence.Queued.Unbonding[0]
	redelegationEntry := evidence.Queued.Redelegation[0]
	if unbondEntry.CompletionTime.IsZero() || redelegationEntry.CompletionTime.IsZero() {
		return errors.New("staking queue entries require non-zero completion times")
	}
	if !evidence.PostUpgradePending.RecordedAt.Before(unbondEntry.CompletionTime) ||
		!evidence.PostUpgradePending.RecordedAt.Before(redelegationEntry.CompletionTime) {
		return errors.New("staking queues did not remain pending across the upgrade before completion time")
	}
	latestCompletion := unbondEntry.CompletionTime
	if redelegationEntry.CompletionTime.After(latestCompletion) {
		latestCompletion = redelegationEntry.CompletionTime
	}
	if evidence.Completed.RecordedAt.Before(latestCompletion) {
		return errors.New("staking queue completion was observed before completion time")
	}
	if unbondEntry.InitialBalance != evidence.UnbondAmount || unbondEntry.Balance != evidence.UnbondAmount ||
		redelegationEntry.InitialBalance != evidence.RedelegateAmount || redelegationEntry.Balance != evidence.RedelegateAmount ||
		strings.TrimSpace(redelegationEntry.SharesDestination) == "" {
		return errors.New("staking queue entry amounts do not match the requested transitions")
	}

	beforeBank, _ := nonNegativeUpgradeP0Amount("bank", evidence.Before.BankBalance)
	queuedBank, _ := nonNegativeUpgradeP0Amount("bank", evidence.Queued.BankBalance)
	wantQueuedBank := new(big.Int).Sub(beforeBank, new(big.Int).Mul(fee, big.NewInt(2)))
	wantQueuedBank.Add(wantQueuedBank, unbondReward)
	wantQueuedBank.Add(wantQueuedBank, redelegateReward)
	if queuedBank.Cmp(wantQueuedBank) != 0 {
		return fmt.Errorf("queued bank balance %s, want %s after fees and withdrawn rewards", queuedBank, wantQueuedBank)
	}
	if err := requireUpgradeP0CheckpointAmountsEqual("post-upgrade pending", evidence.Queued, evidence.PostUpgradePending); err != nil {
		return err
	}

	beforeSource, _ := nonNegativeUpgradeP0Amount("source delegation", evidence.Before.SourceDelegationBalance)
	wantSource := new(big.Int).Sub(new(big.Int).Sub(beforeSource, unbond), redelegate)
	queuedSource, _ := nonNegativeUpgradeP0Amount("source delegation", evidence.Queued.SourceDelegationBalance)
	if queuedSource.Cmp(wantSource) != 0 {
		return fmt.Errorf("queued source delegation %s, want %s", queuedSource, wantSource)
	}
	beforeDestination, _ := nonNegativeUpgradeP0Amount("destination delegation", evidence.Before.DestinationDelegationBalance)
	wantDestination := new(big.Int).Add(beforeDestination, redelegate)
	queuedDestination, _ := nonNegativeUpgradeP0Amount("destination delegation", evidence.Queued.DestinationDelegationBalance)
	if queuedDestination.Cmp(wantDestination) != 0 {
		return fmt.Errorf("queued destination delegation %s, want %s", queuedDestination, wantDestination)
	}

	beforeBonded, _ := nonNegativeUpgradeP0Amount("bonded pool", evidence.Before.Pool.BondedTokens)
	queuedBonded, _ := nonNegativeUpgradeP0Amount("bonded pool", evidence.Queued.Pool.BondedTokens)
	if queuedBonded.Cmp(new(big.Int).Sub(beforeBonded, unbond)) != 0 {
		return errors.New("queued bonded pool does not reflect the unbond amount")
	}
	beforeNotBonded, _ := nonNegativeUpgradeP0Amount("not-bonded pool", evidence.Before.Pool.NotBondedTokens)
	queuedNotBonded, _ := nonNegativeUpgradeP0Amount("not-bonded pool", evidence.Queued.Pool.NotBondedTokens)
	if queuedNotBonded.Cmp(new(big.Int).Add(beforeNotBonded, unbond)) != 0 {
		return errors.New("queued not-bonded pool does not reflect the unbond amount")
	}
	completedBank, _ := nonNegativeUpgradeP0Amount("bank", evidence.Completed.BankBalance)
	if completedBank.Cmp(new(big.Int).Add(queuedBank, unbond)) != 0 {
		return errors.New("completed bank balance does not receive the unbond amount exactly once")
	}
	completedNotBonded, _ := nonNegativeUpgradeP0Amount("not-bonded pool", evidence.Completed.Pool.NotBondedTokens)
	if completedNotBonded.Cmp(new(big.Int).Sub(queuedNotBonded, unbond)) != 0 {
		return errors.New("completed not-bonded pool does not release the unbond amount")
	}
	if evidence.Completed.SourceDelegationBalance != evidence.Queued.SourceDelegationBalance ||
		evidence.Completed.DestinationDelegationBalance != evidence.Queued.DestinationDelegationBalance {
		return errors.New("staking delegation changed while time queues completed")
	}
	if evidence.Completed.BankBalance != evidence.PostRestart.BankBalance ||
		evidence.Completed.SourceDelegationBalance != evidence.PostRestart.SourceDelegationBalance ||
		evidence.Completed.DestinationDelegationBalance != evidence.PostRestart.DestinationDelegationBalance {
		return errors.New("post-restart staking queue account state changed")
	}
	return nil
}

func requireUpgradeP0CheckpointAmountsEqual(label string, want, got upgradeP0StakingQueueCheckpoint) error {
	if want.BankBalance != got.BankBalance ||
		want.SourceDelegationBalance != got.SourceDelegationBalance ||
		want.DestinationDelegationBalance != got.DestinationDelegationBalance {
		return fmt.Errorf("%s staking queue account amounts changed: want=%+v got=%+v", label, want, got)
	}
	return nil
}

func validateUpgradeP0StakingPreservationWithSlashing(
	before upgradeStakingCheckpoint,
	after upgradeStakingCheckpoint,
	slashing upgradeP0SlashingEvidence,
	queues upgradeP0StakingQueueEvidence,
) error {
	if queues.Completed.Height <= queues.PostUpgradePending.Height ||
		len(queues.Completed.Unbonding) != 0 || len(queues.Completed.Redelegation) != 0 {
		return errors.New("staking preservation requires completed cross-upgrade time queues")
	}
	// The dedicated validator-0 fixture must be byte-for-byte preserved, while
	// validator 3 is deliberately slashed in the same connected upgrade lane.
	// Normalize only the global pool for the existing strict validator, then
	// account for that pool delta against the independently observed slash.
	normalizedBefore := before
	normalizedBefore.State.Pool = after.State.Pool
	if err := validateUpgradeStakingPreservation(normalizedBefore, after); err != nil {
		return err
	}
	beforeBonded, err := nonNegativeUpgradeP0Amount("pre-upgrade bonded pool", before.State.Pool.BondedTokens)
	if err != nil {
		return err
	}
	afterBonded, err := nonNegativeUpgradeP0Amount("post-upgrade bonded pool", after.State.Pool.BondedTokens)
	if err != nil {
		return err
	}
	beforeTokens, err := nonNegativeUpgradeP0Amount("pre-slash validator tokens", slashing.Before.Validator.Tokens)
	if err != nil {
		return err
	}
	jailedTokens, err := nonNegativeUpgradeP0Amount("jailed validator tokens", slashing.Jailed.Validator.Tokens)
	if err != nil {
		return err
	}
	poolDelta := new(big.Int).Sub(beforeBonded, afterBonded)
	slashDelta := new(big.Int).Sub(beforeTokens, jailedTokens)
	if slashDelta.Sign() <= 0 || poolDelta.Cmp(slashDelta) != 0 {
		return fmt.Errorf("bonded pool slash delta %s, want validator token slash delta %s", poolDelta, slashDelta)
	}
	beforeNotBonded, err := nonNegativeUpgradeP0Amount("pre-upgrade not-bonded pool", before.State.Pool.NotBondedTokens)
	if err != nil {
		return err
	}
	afterNotBonded, err := nonNegativeUpgradeP0Amount("post-upgrade not-bonded pool", after.State.Pool.NotBondedTokens)
	if err != nil {
		return err
	}
	unbondAmount, err := positiveUpgradeP0Amount("completed queue unbond", queues.UnbondAmount)
	if err != nil {
		return err
	}
	notBondedDelta := new(big.Int).Sub(beforeNotBonded, afterNotBonded)
	if notBondedDelta.Cmp(unbondAmount) != 0 {
		return fmt.Errorf(
			"not-bonded pool queue-completion delta %s, want unbond amount %s",
			notBondedDelta,
			unbondAmount,
		)
	}
	return nil
}

func validateUpgradeP0SlashingEvidence(evidence upgradeP0SlashingEvidence) error {
	if evidence.UpgradeHeight <= 0 || evidence.StoppedAt > evidence.UpgradeHeight || evidence.StoppedAt <= 0 {
		return errors.New("slashing outage must start no later than the upgrade boundary")
	}
	if evidence.Before.Height >= evidence.UpgradeHeight || evidence.Jailed.Height <= evidence.UpgradeHeight {
		return fmt.Errorf("validator must be healthy before and jailed after upgrade height %d", evidence.UpgradeHeight)
	}
	checkpoints := []upgradeP0SlashingCheckpoint{
		evidence.Before,
		evidence.Jailed,
		evidence.Unjailed,
		evidence.Rejoined,
		evidence.PostRestart,
	}
	operator := evidence.Before.Validator.OperatorAddress
	consensus := evidence.Before.SigningInfo.Address
	if strings.TrimSpace(operator) == "" || strings.TrimSpace(consensus) == "" {
		return errors.New("slashing validator identities are required")
	}
	for index, checkpoint := range checkpoints {
		if checkpoint.Height <= 0 || checkpoint.RecordedAt.IsZero() {
			return fmt.Errorf("slashing checkpoint %d is incomplete", index)
		}
		if checkpoint.Validator.OperatorAddress != operator || checkpoint.SigningInfo.Address != consensus {
			return errors.New("slashing validator identity changed")
		}
		if checkpoint.SigningInfo.Tombstoned {
			return fmt.Errorf("validator became tombstoned at %s", checkpoint.Phase)
		}
		if index > 0 && checkpoint.Height <= checkpoints[index-1].Height {
			return errors.New("slashing checkpoint heights do not advance")
		}
	}
	if evidence.Before.Validator.Jailed || evidence.Before.ValidatorPower <= 0 {
		return errors.New("validator must be active before the outage")
	}
	if !evidence.Jailed.Validator.Jailed || evidence.Jailed.ValidatorPower != 0 {
		return errors.New("downtime validator was not jailed and removed from active power")
	}
	beforeTokens, err := positiveUpgradeP0Amount("pre-slash validator tokens", evidence.Before.Validator.Tokens)
	if err != nil {
		return err
	}
	jailedTokens, err := positiveUpgradeP0Amount("jailed validator tokens", evidence.Jailed.Validator.Tokens)
	if err != nil {
		return err
	}
	if jailedTokens.Cmp(beforeTokens) >= 0 {
		return errors.New("downtime validator tokens were not slashed")
	}
	if evidence.OutageBlocksObserved < upgradeP0SlashingMinimumMisses ||
		evidence.Jailed.Height-evidence.StoppedAt != evidence.OutageBlocksObserved {
		return errors.New("downtime jail has no complete minimum outage window evidence")
	}
	jailedUntil, err := time.Parse(time.RFC3339Nano, evidence.Jailed.SigningInfo.JailedUntil)
	if err != nil || !jailedUntil.After(evidence.Jailed.RecordedAt) {
		return errors.New("downtime jail has no future jailed_until")
	}
	if evidence.EarlyUnjail.Codespace != "slashing" || evidence.EarlyUnjail.Code != 4 ||
		evidence.EarlyUnjail.HeightInt64() <= 0 {
		return fmt.Errorf("early unjail must commit an exact slashing/4 rejection: %+v", evidence.EarlyUnjail)
	}
	if evidence.Unjailed.RecordedAt.Before(jailedUntil) || evidence.Unjailed.Validator.Jailed ||
		evidence.Unjailed.Validator.Tokens != evidence.Jailed.Validator.Tokens {
		return errors.New("valid unjail did not occur after jailed_until without another slash")
	}
	if strings.TrimSpace(evidence.UnjailTxHash) == "" {
		return errors.New("successful unjail transaction hash is required")
	}
	if evidence.Rejoined.Validator.Jailed || evidence.Rejoined.ValidatorPower <= 0 ||
		evidence.SignedCommitHeight <= evidence.Unjailed.Height ||
		evidence.SignedCommitHeight > evidence.Rejoined.Height {
		return errors.New("unjail did not restore active power and a signed commit")
	}
	if evidence.PostRestart.Validator.Jailed || evidence.PostRestart.ValidatorPower <= 0 ||
		evidence.PostRestart.Validator.Tokens != evidence.Rejoined.Validator.Tokens {
		return errors.New("post-restart slashing state did not preserve the healthy rejoined validator")
	}
	return nil
}

func positiveUpgradeP0Amount(label, value string) (*big.Int, error) {
	amount, err := nonNegativeUpgradeP0Amount(label, value)
	if err != nil {
		return nil, err
	}
	if amount.Sign() <= 0 {
		return nil, fmt.Errorf("%s must be positive: %q", label, value)
	}
	return amount, nil
}

func nonNegativeUpgradeP0Amount(label, value string) (*big.Int, error) {
	amount, ok := new(big.Int).SetString(strings.TrimSpace(value), 10)
	if !ok || amount.Sign() < 0 {
		return nil, fmt.Errorf("%s is not a non-negative integer: %q", label, value)
	}
	return amount, nil
}
