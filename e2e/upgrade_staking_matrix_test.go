package e2e_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	cmttypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	upgradeStakingDenom                    = "umed"
	upgradeStakingDelegatorKey             = "upgrade-staking-delegator"
	upgradeStakingRewardRecipientKey       = "upgrade-staking-reward-recipient"
	upgradeStakingFundAmount               = "2000000000"
	upgradeStakingInitialDelegationAmount  = "1000000000"
	upgradeStakingAdditionalDelegateAmount = "100000000"
	upgradeStakingRewardAccrualBlocks      = int64(5)
)

type upgradeStakingFixture struct {
	DelegatorKeyName         string          `json:"delegator_key_name"`
	DelegatorAddress         string          `json:"delegator_address"`
	RewardRecipientKeyName   string          `json:"reward_recipient_key_name"`
	RewardRecipientAddress   string          `json:"reward_recipient_address"`
	ValidatorIndex           int             `json:"validator_index"`
	ValidatorOperator        string          `json:"validator_operator"`
	ValidatorConsensusAddr   string          `json:"validator_consensus_address"`
	ValidatorConsensusPubKey json.RawMessage `json:"validator_consensus_pubkey"`
	InitialDelegationAmount  string          `json:"initial_delegation_amount"`
}

type upgradeStakingQueryResponses struct {
	Delegation         []byte
	Validator          []byte
	Pool               []byte
	DelegatorRewards   []byte
	OutstandingRewards []byte
	Commission         []byte
	SigningInfo        []byte
}

type upgradeStakingCoin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type upgradeStakingDecCoins []upgradeStakingCoin

func (coins upgradeStakingDecCoins) AmountOf(denom string) string {
	for _, coin := range coins {
		if coin.Denom == denom {
			return coin.Amount
		}
	}
	return "0"
}

type upgradeDelegationState struct {
	DelegatorAddress  string             `json:"delegator_address"`
	ValidatorOperator string             `json:"validator_operator"`
	Shares            string             `json:"shares"`
	Balance           upgradeStakingCoin `json:"balance"`
}

type upgradeValidatorState struct {
	OperatorAddress string          `json:"operator_address"`
	ConsensusPubKey json.RawMessage `json:"consensus_pubkey"`
	Jailed          bool            `json:"jailed"`
	Status          string          `json:"status"`
	Tokens          string          `json:"tokens"`
	DelegatorShares string          `json:"delegator_shares"`
}

type upgradeStakingPoolState struct {
	NotBondedTokens string `json:"not_bonded_tokens"`
	BondedTokens    string `json:"bonded_tokens"`
}

type upgradeSigningInfoState struct {
	Address             string `json:"address"`
	StartHeight         int64  `json:"start_height"`
	IndexOffset         int64  `json:"index_offset"`
	JailedUntil         string `json:"jailed_until"`
	Tombstoned          bool   `json:"tombstoned"`
	MissedBlocksCounter int64  `json:"missed_blocks_counter"`
}

type upgradeStakingState struct {
	Delegation          upgradeDelegationState  `json:"delegation"`
	Validator           upgradeValidatorState   `json:"validator"`
	Pool                upgradeStakingPoolState `json:"pool"`
	DelegatorRewards    upgradeStakingDecCoins  `json:"delegator_rewards"`
	OutstandingRewards  upgradeStakingDecCoins  `json:"outstanding_rewards"`
	ValidatorCommission upgradeStakingDecCoins  `json:"validator_commission"`
	SigningInfo         upgradeSigningInfoState `json:"signing_info"`
}

type upgradeStakingCheckpoint struct {
	Phase                string                               `json:"phase"`
	RecordedAt           time.Time                            `json:"recorded_at"`
	Height               int64                                `json:"height"`
	Observation          harness.UpgradeCheckpointObservation `json:"observation"`
	DelegatorBankBalance string                               `json:"delegator_bank_balance"`
	RecipientBankBalance string                               `json:"reward_recipient_bank_balance"`
	State                upgradeStakingState                  `json:"state"`
	TxHashes             []string                             `json:"tx_hashes,omitempty"`
}

type upgradeStakingMutationEvidence struct {
	AdditionalDelegationAmount string                   `json:"additional_delegation_amount"`
	DelegateTxHash             string                   `json:"delegate_tx_hash"`
	WithdrawRewardTxHash       string                   `json:"withdraw_reward_tx_hash"`
	Before                     upgradeStakingCheckpoint `json:"before"`
	BeforeRewardWithdraw       upgradeStakingCheckpoint `json:"before_reward_withdraw"`
	After                      upgradeStakingCheckpoint `json:"after"`
}

type upgradeStakingPreparationEvidence struct {
	Fixture               upgradeStakingFixture    `json:"fixture"`
	FundTxHash            string                   `json:"fund_tx_hash"`
	SetWithdrawAddressTx  string                   `json:"set_withdraw_address_tx_hash"`
	InitialDelegateTxHash string                   `json:"initial_delegate_tx_hash"`
	Checkpoint            upgradeStakingCheckpoint `json:"checkpoint"`
}

func (e upgradeStakingPreparationEvidence) TxHashes() []string {
	return []string{e.FundTxHash, e.SetWithdrawAddressTx, e.InitialDelegateTxHash}
}

// prepareUpgradeStakingMatrix creates one dedicated v2.2.1 delegator, routes
// its rewards to a fee-isolated recipient, delegates to the selected
// validator, and captures the first of the five connected checkpoints.
func prepareUpgradeStakingMatrix(
	ctx context.Context,
	network *harness.Network,
	validatorIndex int,
) (upgradeStakingPreparationEvidence, error) {
	if err := validateUpgradeStakingNetwork(network, validatorIndex); err != nil {
		return upgradeStakingPreparationEvidence{}, err
	}
	targetValidator := network.Chain.Validators[validatorIndex]
	validatorOperator, err := queryUpgradeValidatorOperator(ctx, targetValidator)
	if err != nil {
		return upgradeStakingPreparationEvidence{}, err
	}
	delegator, err := network.BuildWallet(ctx, upgradeStakingDelegatorKey, "")
	if err != nil {
		return upgradeStakingPreparationEvidence{}, fmt.Errorf("build upgrade staking delegator: %w", err)
	}
	rewardRecipient, err := network.BuildWallet(ctx, upgradeStakingRewardRecipientKey, "")
	if err != nil {
		return upgradeStakingPreparationEvidence{}, fmt.Errorf("build upgrade staking reward recipient: %w", err)
	}
	fixture := upgradeStakingFixture{
		DelegatorKeyName:        delegator.KeyName(),
		DelegatorAddress:        delegator.FormattedAddress(),
		RewardRecipientKeyName:  rewardRecipient.KeyName(),
		RewardRecipientAddress:  rewardRecipient.FormattedAddress(),
		ValidatorIndex:          validatorIndex,
		ValidatorOperator:       validatorOperator,
		InitialDelegationAmount: upgradeStakingInitialDelegationAmount,
	}
	txNode := network.Chain.Validators[0]
	funded, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-staking-fund-delegator",
		txNode,
		interchaintest.FaucetAccountKeyName,
		"bank", "send",
		interchaintest.FaucetAccountKeyName,
		fixture.DelegatorAddress,
		upgradeStakingFundAmount+upgradeStakingDenom,
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return upgradeStakingPreparationEvidence{}, fmt.Errorf("fund upgrade staking delegator: %w", err)
	}
	setWithdrawAddress, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-staking-set-withdraw-address",
		txNode,
		fixture.DelegatorKeyName,
		"distribution", "set-withdraw-addr", fixture.RewardRecipientAddress,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return upgradeStakingPreparationEvidence{}, fmt.Errorf("set upgrade staking reward recipient: %w", err)
	}
	delegated, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-staking-initial-delegate",
		txNode,
		fixture.DelegatorKeyName,
		"staking", "delegate", fixture.ValidatorOperator,
		upgradeStakingInitialDelegationAmount+upgradeStakingDenom,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return upgradeStakingPreparationEvidence{}, fmt.Errorf("create v2.2.1 staking delegation: %w", err)
	}
	if err := waitUpgradeStakingBlocks(ctx, network, upgradeStakingRewardAccrualBlocks); err != nil {
		return upgradeStakingPreparationEvidence{}, err
	}
	txHashes := []string{funded.TxHash, setWithdrawAddress.TxHash, delegated.TxHash}
	checkpoint, err := captureUpgradeStakingCheckpoint(
		ctx,
		network,
		fixture,
		"v2.2.1-preparation",
		txHashes,
	)
	if err != nil {
		return upgradeStakingPreparationEvidence{}, err
	}
	if checkpoint.State.Delegation.Balance.Amount != upgradeStakingInitialDelegationAmount {
		return upgradeStakingPreparationEvidence{}, fmt.Errorf(
			"initial delegation balance %s, want %s",
			checkpoint.State.Delegation.Balance.Amount,
			upgradeStakingInitialDelegationAmount,
		)
	}
	reward, err := sdkmath.LegacyNewDecFromStr(checkpoint.State.DelegatorRewards.AmountOf(upgradeStakingDenom))
	if err != nil || !reward.IsPositive() {
		return upgradeStakingPreparationEvidence{}, fmt.Errorf(
			"v2.2.1 delegation reward did not accrue after %d blocks: %s",
			upgradeStakingRewardAccrualBlocks,
			checkpoint.State.DelegatorRewards.AmountOf(upgradeStakingDenom),
		)
	}
	if checkpoint.State.Validator.Jailed || checkpoint.State.SigningInfo.Tombstoned {
		return upgradeStakingPreparationEvidence{}, errors.New("v2.2.1 target validator is jailed or tombstoned")
	}
	fixture.ValidatorConsensusAddr = checkpoint.State.SigningInfo.Address
	fixture.ValidatorConsensusPubKey = append(json.RawMessage(nil), checkpoint.State.Validator.ConsensusPubKey...)
	evidence := upgradeStakingPreparationEvidence{
		Fixture:               fixture,
		FundTxHash:            funded.TxHash,
		SetWithdrawAddressTx:  setWithdrawAddress.TxHash,
		InitialDelegateTxHash: delegated.TxHash,
		Checkpoint:            checkpoint,
	}
	if err := network.WriteArtifactJSON("upgrade/staking/preparation.json", evidence); err != nil {
		return upgradeStakingPreparationEvidence{}, fmt.Errorf("record staking preparation: %w", err)
	}
	return evidence, nil
}

// captureUpgradeStakingCheckpoint pins all CLI state queries to one full-node
// height. Bank balances are queried through the full-node gRPC boundary and
// included alongside the pinned staking, distribution, and slashing state.
func captureUpgradeStakingCheckpoint(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeStakingFixture,
	phase string,
	txHashes []string,
) (upgradeStakingCheckpoint, error) {
	if err := validateUpgradeStakingNetwork(network, fixture.ValidatorIndex); err != nil {
		return upgradeStakingCheckpoint{}, err
	}
	if err := validateUpgradeStakingPhase(phase); err != nil {
		return upgradeStakingCheckpoint{}, err
	}
	if strings.TrimSpace(fixture.DelegatorAddress) == "" || strings.TrimSpace(fixture.RewardRecipientAddress) == "" ||
		strings.TrimSpace(fixture.ValidatorOperator) == "" {
		return upgradeStakingCheckpoint{}, errors.New("staking checkpoint fixture addresses are required")
	}
	fullNode := network.Chain.FullNodes[0]
	height, err := fullNode.Height(ctx)
	if err != nil {
		return upgradeStakingCheckpoint{}, fmt.Errorf("query staking checkpoint height for %s: %w", phase, err)
	}
	observation, err := network.CaptureUpgradeCheckpointObservation(ctx, "upgrade-staking-"+phase, fullNode, height)
	if err != nil {
		return upgradeStakingCheckpoint{}, err
	}
	query := func(suffix string, command ...string) ([]byte, error) {
		command = append(append([]string(nil), command...), "--height", strconv.FormatInt(height, 10))
		return network.FullNodeCLIQuery(ctx, "upgrade-staking-"+phase+"-"+suffix, command...)
	}
	delegationRaw, err := query("delegation", "staking", "delegation", fixture.DelegatorAddress, fixture.ValidatorOperator)
	if err != nil {
		return upgradeStakingCheckpoint{}, err
	}
	validatorRaw, err := query("validator", "staking", "validator", fixture.ValidatorOperator)
	if err != nil {
		return upgradeStakingCheckpoint{}, err
	}
	validator, err := decodeUpgradeValidator(validatorRaw)
	if err != nil {
		return upgradeStakingCheckpoint{}, err
	}
	consensusPubKey := strings.TrimSpace(string(validator.ConsensusPubKey))
	if consensusPubKey == "" {
		return upgradeStakingCheckpoint{}, errors.New("staking validator has no consensus public key")
	}
	poolRaw, err := query("pool", "staking", "pool")
	if err != nil {
		return upgradeStakingCheckpoint{}, err
	}
	delegatorRewardsRaw, err := query(
		"delegator-rewards",
		"distribution", "rewards", fixture.DelegatorAddress,
	)
	if err != nil {
		return upgradeStakingCheckpoint{}, err
	}
	outstandingRewardsRaw, err := query(
		"outstanding-rewards",
		"distribution", "validator-outstanding-rewards", fixture.ValidatorOperator,
	)
	if err != nil {
		return upgradeStakingCheckpoint{}, err
	}
	commissionRaw, err := query("commission", "distribution", "commission", fixture.ValidatorOperator)
	if err != nil {
		return upgradeStakingCheckpoint{}, err
	}
	signingInfoRaw, err := query("signing-info", "slashing", "signing-info", consensusPubKey)
	if err != nil {
		return upgradeStakingCheckpoint{}, err
	}
	state, err := decodeUpgradeStakingQueries(upgradeStakingQueryResponses{
		Delegation:         delegationRaw,
		Validator:          validatorRaw,
		Pool:               poolRaw,
		DelegatorRewards:   delegatorRewardsRaw,
		OutstandingRewards: outstandingRewardsRaw,
		Commission:         commissionRaw,
		SigningInfo:        signingInfoRaw,
	}, fixture)
	if err != nil {
		return upgradeStakingCheckpoint{}, fmt.Errorf("decode %s staking checkpoint: %w", phase, err)
	}
	if fixture.ValidatorConsensusAddr != "" && state.SigningInfo.Address != fixture.ValidatorConsensusAddr {
		return upgradeStakingCheckpoint{}, fmt.Errorf(
			"validator consensus address %q, want %q",
			state.SigningInfo.Address,
			fixture.ValidatorConsensusAddr,
		)
	}
	if len(fixture.ValidatorConsensusPubKey) > 0 {
		equal, err := equalUpgradeConsensusPubKeys(state.Validator.ConsensusPubKey, fixture.ValidatorConsensusPubKey)
		if err != nil {
			return upgradeStakingCheckpoint{}, fmt.Errorf("compare validator consensus public key with fixture: %w", err)
		}
		if !equal {
			return upgradeStakingCheckpoint{}, errors.New("validator consensus public key changed from fixture")
		}
	}
	pinnedCtx := harness.ContextAtHeight(ctx, height)
	delegatorBalance, err := network.QueryFullNodeBalance(pinnedCtx, fixture.DelegatorAddress, upgradeStakingDenom)
	if err != nil {
		return upgradeStakingCheckpoint{}, err
	}
	recipientBalance, err := network.QueryFullNodeBalance(pinnedCtx, fixture.RewardRecipientAddress, upgradeStakingDenom)
	if err != nil {
		return upgradeStakingCheckpoint{}, err
	}
	checkpoint := upgradeStakingCheckpoint{
		Phase:                phase,
		RecordedAt:           observation.ObservedAt,
		Height:               height,
		Observation:          observation,
		DelegatorBankBalance: delegatorBalance.String(),
		RecipientBankBalance: recipientBalance.String(),
		State:                state,
		TxHashes:             append([]string(nil), txHashes...),
	}
	artifactPath := "upgrade/staking/checkpoints/" + phase + ".json"
	if err := network.WriteArtifactJSON(artifactPath, checkpoint); err != nil {
		return upgradeStakingCheckpoint{}, fmt.Errorf("record %s staking checkpoint: %w", phase, err)
	}
	if err := network.AppendArtifactJSON("upgrade/staking/phases.jsonl", checkpoint); err != nil {
		return upgradeStakingCheckpoint{}, fmt.Errorf("record %s staking phase: %w", phase, err)
	}
	return checkpoint, nil
}

func validateUpgradeStakingNetwork(network *harness.Network, validatorIndex int) error {
	if network == nil || network.Chain == nil {
		return errors.New("upgrade staking network is required")
	}
	if len(network.Chain.FullNodes) == 0 {
		return errors.New("upgrade staking network requires a full node")
	}
	if validatorIndex < 0 || validatorIndex >= len(network.Chain.Validators) {
		return fmt.Errorf("upgrade staking validator index %d outside [0,%d)", validatorIndex, len(network.Chain.Validators))
	}
	return nil
}

func validateUpgradeStakingPhase(phase string) error {
	if phase == "" || phase != strings.TrimSpace(phase) {
		return errors.New("staking checkpoint phase is required without surrounding whitespace")
	}
	for _, character := range phase {
		if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || character == '-' || character == '.' {
			continue
		}
		return fmt.Errorf("staking checkpoint phase %q contains unsupported character %q", phase, character)
	}
	return nil
}

func queryUpgradeValidatorOperator(ctx context.Context, node *cosmos.ChainNode) (string, error) {
	if node == nil {
		return "", errors.New("validator node is required")
	}
	stdout, stderr, err := node.Exec(ctx, node.BinCommand(
		"keys", "show", "validator",
		"--bech", "val",
		"--address",
		"--keyring-backend", "test",
	), node.Chain.Config().Env)
	if err != nil {
		return "", fmt.Errorf("query validator operator address from %s: %w: %s", node.Name(), err, strings.TrimSpace(string(stderr)))
	}
	operator := strings.TrimSpace(string(stdout))
	hrp, _, err := bech32.DecodeAndConvert(operator)
	if err != nil || hrp != "panaceavaloper" {
		return "", fmt.Errorf("validator operator address from %s is invalid: %q", node.Name(), operator)
	}
	return operator, nil
}

func waitUpgradeStakingBlocks(ctx context.Context, network *harness.Network, blocks int64) error {
	if blocks <= 0 {
		return fmt.Errorf("staking block wait must be positive: %d", blocks)
	}
	if network == nil || network.Chain == nil || len(network.Chain.FullNodes) == 0 {
		return errors.New("staking block wait requires a full node")
	}
	fullNode := network.Chain.FullNodes[0]
	startHeight, err := fullNode.Height(ctx)
	if err != nil {
		return fmt.Errorf("query staking wait start height: %w", err)
	}
	if err := network.WaitForNodeHeight(ctx, fullNode, startHeight+blocks); err != nil {
		return fmt.Errorf("wait %d staking reward blocks after %d: %w", blocks, startHeight, err)
	}
	return nil
}

// mutateUpgradeStakingMatrix exercises the post-upgrade write paths. An
// additional delegation first realizes previously accrued rewards to the
// fee-isolated recipient; a later explicit withdrawal proves the remaining
// reward state can also be consumed by the upgraded binary.
func mutateUpgradeStakingMatrix(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeStakingFixture,
	before upgradeStakingCheckpoint,
) (upgradeStakingMutationEvidence, error) {
	if err := validateUpgradeStakingNetwork(network, fixture.ValidatorIndex); err != nil {
		return upgradeStakingMutationEvidence{}, err
	}
	if before.State.Delegation.DelegatorAddress != fixture.DelegatorAddress ||
		before.State.Delegation.ValidatorOperator != fixture.ValidatorOperator {
		return upgradeStakingMutationEvidence{}, errors.New("pre-mutation checkpoint does not match staking fixture")
	}
	txNode := network.Chain.Validators[0]
	delegated, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-staking-additional-delegate",
		txNode,
		fixture.DelegatorKeyName,
		"staking", "delegate", fixture.ValidatorOperator,
		upgradeStakingAdditionalDelegateAmount+upgradeStakingDenom,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return upgradeStakingMutationEvidence{}, fmt.Errorf("post-upgrade additional delegation: %w", err)
	}
	if err := waitUpgradeStakingBlocks(ctx, network, upgradeStakingRewardAccrualBlocks); err != nil {
		return upgradeStakingMutationEvidence{}, err
	}
	beforeWithdrawHashes := append(append([]string(nil), before.TxHashes...), delegated.TxHash)
	beforeWithdraw, err := captureUpgradeStakingCheckpoint(
		ctx,
		network,
		fixture,
		"post-upgrade-before-reward-withdraw",
		beforeWithdrawHashes,
	)
	if err != nil {
		return upgradeStakingMutationEvidence{}, err
	}
	withdrawn, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-staking-withdraw-reward",
		txNode,
		fixture.DelegatorKeyName,
		"distribution", "withdraw-rewards", fixture.ValidatorOperator,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return upgradeStakingMutationEvidence{}, fmt.Errorf("post-upgrade reward withdrawal: %w", err)
	}
	if err := waitUpgradeStakingBlocks(ctx, network, 2); err != nil {
		return upgradeStakingMutationEvidence{}, err
	}
	afterHashes := append(append([]string(nil), beforeWithdrawHashes...), withdrawn.TxHash)
	after, err := captureUpgradeStakingCheckpoint(
		ctx,
		network,
		fixture,
		"post-upgrade-mutation",
		afterHashes,
	)
	if err != nil {
		return upgradeStakingMutationEvidence{}, err
	}
	evidence := upgradeStakingMutationEvidence{
		AdditionalDelegationAmount: upgradeStakingAdditionalDelegateAmount,
		DelegateTxHash:             delegated.TxHash,
		WithdrawRewardTxHash:       withdrawn.TxHash,
		Before:                     before,
		BeforeRewardWithdraw:       beforeWithdraw,
		After:                      after,
	}
	if err := validateUpgradeStakingMutation(evidence); err != nil {
		return upgradeStakingMutationEvidence{}, err
	}
	if err := network.WriteArtifactJSON("upgrade/staking/mutation.json", evidence); err != nil {
		return upgradeStakingMutationEvidence{}, fmt.Errorf("record upgrade staking mutation: %w", err)
	}
	return evidence, nil
}

type upgradeValidatorLivenessEvidence struct {
	ValidatorIndex         int                        `json:"validator_index"`
	TargetConsensusAddress string                     `json:"target_consensus_address"`
	ValidatorSet           []harness.ValidatorPower   `json:"validator_set"`
	MinimumStoppedBlocks   int64                      `json:"minimum_stopped_blocks"`
	FaultWindow            harness.QuorumHeightWindow `json:"fault_window"`
	BeforeStop             upgradeStakingCheckpoint   `json:"before_stop"`
	WhileStopped           upgradeStakingCheckpoint   `json:"while_stopped"`
	AfterRejoin            upgradeStakingCheckpoint   `json:"after_rejoin"`
	SignedCommitHeight     int64                      `json:"signed_commit_height"`
	RejoinHistory          []harness.BlockEvidence    `json:"rejoin_history"`
}

// captureAndValidateUpgradeStakingPreservation proves that the v2.2.1
// delegation, validator, pool, reward, commission, and signing state survived
// the binary upgrade before any post-upgrade transaction mutates that state.
func captureAndValidateUpgradeStakingPreservation(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeStakingFixture,
	before upgradeStakingCheckpoint,
) (upgradeStakingCheckpoint, error) {
	after, err := captureUpgradeStakingCheckpoint(
		ctx,
		network,
		fixture,
		"post-upgrade-preservation",
		before.TxHashes,
	)
	if err != nil {
		return upgradeStakingCheckpoint{}, err
	}
	if err := validateUpgradeStakingPreservation(before, after); err != nil {
		return upgradeStakingCheckpoint{}, fmt.Errorf("validate staking state across upgrade: %w", err)
	}
	return after, nil
}

// exerciseUpgradeValidatorLiveness stops exactly one validator only after the
// live validator set proves the remaining voting power can still commit. It
// then proves missed-signature accounting, restarts that same validator state,
// observes a new commit signature, and compares its history with the full node.
func exerciseUpgradeValidatorLiveness(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeStakingFixture,
	mutation upgradeStakingMutationEvidence,
	minimumStoppedBlocks int64,
) (_ upgradeValidatorLivenessEvidence, returnErr error) {
	if err := validateUpgradeStakingNetwork(network, fixture.ValidatorIndex); err != nil {
		return upgradeValidatorLivenessEvidence{}, err
	}
	if minimumStoppedBlocks <= 0 {
		return upgradeValidatorLivenessEvidence{}, fmt.Errorf("minimum stopped blocks must be positive: %d", minimumStoppedBlocks)
	}
	if mutation.After.State.Delegation.DelegatorAddress != fixture.DelegatorAddress ||
		mutation.After.State.Delegation.ValidatorOperator != fixture.ValidatorOperator {
		return upgradeValidatorLivenessEvidence{}, errors.New("post-mutation checkpoint does not match staking fixture")
	}

	consensusHRP, consensusAddress, err := bech32.DecodeAndConvert(fixture.ValidatorConsensusAddr)
	if err != nil {
		return upgradeValidatorLivenessEvidence{}, fmt.Errorf("decode target validator consensus address: %w", err)
	}
	if consensusHRP != "panaceavalcons" || len(consensusAddress) == 0 {
		return upgradeValidatorLivenessEvidence{}, fmt.Errorf(
			"target validator consensus address has HRP %q and %d bytes, want panaceavalcons with bytes",
			consensusHRP,
			len(consensusAddress),
		)
	}
	targetConsensusHex := strings.ToUpper(hex.EncodeToString(consensusAddress))
	fullNode := network.Chain.FullNodes[0]
	targetValidator := network.Chain.Validators[fixture.ValidatorIndex]

	beforeStop, err := captureUpgradeStakingCheckpoint(
		ctx,
		network,
		fixture,
		"post-upgrade-before-validator-stop",
		mutation.After.TxHashes,
	)
	if err != nil {
		return upgradeValidatorLivenessEvidence{}, err
	}
	validatorSet, err := network.ValidatorSet(ctx, fullNode, beforeStop.Height)
	if err != nil {
		return upgradeValidatorLivenessEvidence{}, err
	}
	if err := validateSingleValidatorStopSafety(validatorSet, targetConsensusHex); err != nil {
		return upgradeValidatorLivenessEvidence{}, fmt.Errorf("refuse unsafe validator stop: %w", err)
	}

	validatorStopped := false
	defer func() {
		if !validatorStopped {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 90*time.Second)
		defer cancel()
		cleanupErr := network.StartQuorumValidator(
			cleanupCtx,
			"post-upgrade-validator-stop-cleanup",
			fixture.ValidatorIndex,
		)
		returnErr = errors.Join(returnErr, cleanupErr)
	}()

	if err := network.StopQuorumValidator(
		ctx,
		"post-upgrade-validator-stop",
		fixture.ValidatorIndex,
	); err != nil {
		return upgradeValidatorLivenessEvidence{}, err
	}
	validatorStopped = true
	faultStartHeight, err := fullNode.Height(ctx)
	if err != nil {
		return upgradeValidatorLivenessEvidence{}, fmt.Errorf("query height after validator stop: %w", err)
	}
	// One extra carrier block absorbs a commit already in flight when the
	// container stops, leaving at least minimumStoppedBlocks definite misses.
	faultWindow, err := network.WaitForQuorumProgress(
		ctx,
		"post-upgrade-validator-stopped-progress",
		fullNode,
		faultStartHeight,
		minimumStoppedBlocks+1,
	)
	if err != nil {
		return upgradeValidatorLivenessEvidence{}, err
	}
	whileStopped, err := captureUpgradeStakingCheckpoint(
		ctx,
		network,
		fixture,
		"post-upgrade-validator-stopped",
		mutation.After.TxHashes,
	)
	if err != nil {
		return upgradeValidatorLivenessEvidence{}, err
	}

	if err := network.StartQuorumValidator(
		ctx,
		"post-upgrade-validator-rejoin",
		fixture.ValidatorIndex,
	); err != nil {
		return upgradeValidatorLivenessEvidence{}, err
	}
	validatorStopped = false
	if err := network.WaitForNodeHeight(ctx, targetValidator, faultWindow.EndHeight); err != nil {
		return upgradeValidatorLivenessEvidence{}, fmt.Errorf("wait for rejoined validator to catch up: %w", err)
	}
	signedCommitHeight, err := waitForUpgradeValidatorCommitSignature(
		ctx,
		network,
		fullNode,
		consensusAddress,
		faultWindow.EndHeight,
		8,
	)
	if err != nil {
		return upgradeValidatorLivenessEvidence{}, err
	}
	afterRejoin, err := captureUpgradeStakingCheckpoint(
		ctx,
		network,
		fixture,
		"post-upgrade-validator-rejoined",
		mutation.After.TxHashes,
	)
	if err != nil {
		return upgradeValidatorLivenessEvidence{}, err
	}
	if err := network.WaitForNodeHeight(ctx, targetValidator, afterRejoin.Height); err != nil {
		return upgradeValidatorLivenessEvidence{}, fmt.Errorf("wait for rejoined validator at checkpoint height: %w", err)
	}
	rejoinHistory, err := network.RequireSameHistoryAtHeight(
		ctx,
		afterRejoin.Height,
		targetValidator,
		fullNode,
	)
	if err != nil {
		return upgradeValidatorLivenessEvidence{}, fmt.Errorf("compare rejoined validator history: %w", err)
	}

	evidence := upgradeValidatorLivenessEvidence{
		ValidatorIndex:         fixture.ValidatorIndex,
		TargetConsensusAddress: targetConsensusHex,
		ValidatorSet:           validatorSet,
		MinimumStoppedBlocks:   minimumStoppedBlocks,
		FaultWindow:            faultWindow,
		BeforeStop:             beforeStop,
		WhileStopped:           whileStopped,
		AfterRejoin:            afterRejoin,
		SignedCommitHeight:     signedCommitHeight,
		RejoinHistory:          rejoinHistory,
	}
	if err := validateUpgradeValidatorLiveness(evidence); err != nil {
		return upgradeValidatorLivenessEvidence{}, err
	}
	if err := network.WriteArtifactJSON("upgrade/staking/validator-liveness.json", evidence); err != nil {
		return upgradeValidatorLivenessEvidence{}, fmt.Errorf("record validator liveness evidence: %w", err)
	}
	return evidence, nil
}

// waitForUpgradeValidatorCommitSignature scans historical commits already
// produced during startup, then waits a bounded number of additional blocks.
// A block at height H carries signatures for LastCommit.Height H-1.
func waitForUpgradeValidatorCommitSignature(
	ctx context.Context,
	network *harness.Network,
	observer *cosmos.ChainNode,
	consensusAddress []byte,
	minimumSignedHeight int64,
	maximumFutureBlocks int64,
) (int64, error) {
	if network == nil || observer == nil {
		return 0, errors.New("commit signature observer network and node are required")
	}
	if len(consensusAddress) == 0 {
		return 0, errors.New("commit signature consensus address is required")
	}
	if minimumSignedHeight < 0 || maximumFutureBlocks <= 0 {
		return 0, fmt.Errorf(
			"commit signature bounds are invalid: minimum=%d future_blocks=%d",
			minimumSignedHeight,
			maximumFutureBlocks,
		)
	}
	latestHeight, err := observer.Height(ctx)
	if err != nil {
		return 0, fmt.Errorf("query commit signature observer height: %w", err)
	}
	firstCarrierHeight := minimumSignedHeight + 2
	lastCarrierHeight := latestHeight + maximumFutureBlocks
	if lastCarrierHeight < firstCarrierHeight+maximumFutureBlocks-1 {
		lastCarrierHeight = firstCarrierHeight + maximumFutureBlocks - 1
	}
	for carrierHeight := firstCarrierHeight; carrierHeight <= lastCarrierHeight; carrierHeight++ {
		if err := network.WaitForNodeHeight(ctx, observer, carrierHeight); err != nil {
			return 0, fmt.Errorf("wait for commit carrier block %d: %w", carrierHeight, err)
		}
		result, err := observer.Client.Block(ctx, &carrierHeight)
		if err != nil {
			return 0, fmt.Errorf("query commit carrier block %d: %w", carrierHeight, err)
		}
		if result == nil || result.Block == nil || result.Block.LastCommit == nil {
			return 0, fmt.Errorf("commit carrier block %d returned no last commit", carrierHeight)
		}
		lastCommit := result.Block.LastCommit
		if lastCommit.Height <= minimumSignedHeight {
			continue
		}
		for _, signature := range lastCommit.Signatures {
			if signature.BlockIDFlag == cmttypes.BlockIDFlagCommit &&
				bytes.Equal(signature.ValidatorAddress, consensusAddress) {
				return lastCommit.Height, nil
			}
		}
	}
	return 0, fmt.Errorf(
		"validator %s did not sign a commit after height %d through carrier block %d",
		strings.ToUpper(hex.EncodeToString(consensusAddress)),
		minimumSignedHeight,
		lastCarrierHeight,
	)
}

// captureAndValidateUpgradeStakingPostRestart is the fifth and final
// checkpoint. The caller invokes it only after the suite's all-node restart.
func captureAndValidateUpgradeStakingPostRestart(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeStakingFixture,
	liveness upgradeValidatorLivenessEvidence,
) (upgradeStakingCheckpoint, error) {
	if err := validateUpgradeValidatorLiveness(liveness); err != nil {
		return upgradeStakingCheckpoint{}, fmt.Errorf("validate pre-restart validator liveness: %w", err)
	}
	afterRestart, err := captureUpgradeStakingCheckpoint(
		ctx,
		network,
		fixture,
		"post-restart",
		liveness.AfterRejoin.TxHashes,
	)
	if err != nil {
		return upgradeStakingCheckpoint{}, err
	}
	if err := validateUpgradeStakingPreservation(liveness.AfterRejoin, afterRestart); err != nil {
		return upgradeStakingCheckpoint{}, fmt.Errorf("validate staking state across all-node restart: %w", err)
	}
	if err := network.WriteArtifactJSON("upgrade/staking/post-restart-validation.json", map[string]any{
		"before_restart": liveness.AfterRejoin,
		"after_restart":  afterRestart,
	}); err != nil {
		return upgradeStakingCheckpoint{}, fmt.Errorf("record post-restart staking validation: %w", err)
	}
	return afterRestart, nil
}

func validateUpgradeStakingPreservation(before, after upgradeStakingCheckpoint) error {
	if before.Height <= 0 || after.Height <= before.Height {
		return fmt.Errorf("staking preservation heights must advance: before=%d after=%d", before.Height, after.Height)
	}
	if before.DelegatorBankBalance != after.DelegatorBankBalance {
		return fmt.Errorf("delegator bank balance changed from %s to %s", before.DelegatorBankBalance, after.DelegatorBankBalance)
	}
	if before.RecipientBankBalance != after.RecipientBankBalance {
		return fmt.Errorf("reward recipient bank balance changed from %s to %s", before.RecipientBankBalance, after.RecipientBankBalance)
	}

	beforeDelegation := before.State.Delegation
	afterDelegation := after.State.Delegation
	if beforeDelegation.DelegatorAddress != afterDelegation.DelegatorAddress ||
		beforeDelegation.ValidatorOperator != afterDelegation.ValidatorOperator {
		return errors.New("delegation identity changed across upgrade")
	}
	if beforeDelegation.Balance != afterDelegation.Balance {
		return fmt.Errorf("delegation balance changed from %+v to %+v", beforeDelegation.Balance, afterDelegation.Balance)
	}
	if beforeDelegation.Shares != afterDelegation.Shares {
		return fmt.Errorf("delegation shares changed from %s to %s", beforeDelegation.Shares, afterDelegation.Shares)
	}

	beforeValidator := before.State.Validator
	afterValidator := after.State.Validator
	if beforeValidator.OperatorAddress != afterValidator.OperatorAddress {
		return errors.New("validator operator changed across upgrade")
	}
	consensusPubKeyEqual, err := equalUpgradeConsensusPubKeys(beforeValidator.ConsensusPubKey, afterValidator.ConsensusPubKey)
	if err != nil {
		return fmt.Errorf("compare validator consensus public key across upgrade: %w", err)
	}
	if !consensusPubKeyEqual {
		return errors.New("validator consensus public key changed across upgrade")
	}
	if beforeValidator.Jailed || afterValidator.Jailed {
		return errors.New("validator became or remained jailed across upgrade")
	}
	if beforeValidator.Status != afterValidator.Status {
		return fmt.Errorf("validator status changed from %s to %s", beforeValidator.Status, afterValidator.Status)
	}
	if beforeValidator.Tokens != afterValidator.Tokens {
		return fmt.Errorf("validator tokens changed from %s to %s", beforeValidator.Tokens, afterValidator.Tokens)
	}
	if beforeValidator.DelegatorShares != afterValidator.DelegatorShares {
		return fmt.Errorf("validator delegator shares changed from %s to %s", beforeValidator.DelegatorShares, afterValidator.DelegatorShares)
	}
	if before.State.Pool != after.State.Pool {
		return fmt.Errorf("staking pool changed from %+v to %+v", before.State.Pool, after.State.Pool)
	}
	if err := validateDecCoinsNonDecreasing("delegator rewards", before.State.DelegatorRewards, after.State.DelegatorRewards); err != nil {
		return err
	}
	if err := validateDecCoinsNonDecreasing("validator outstanding rewards", before.State.OutstandingRewards, after.State.OutstandingRewards); err != nil {
		return err
	}
	if err := validateDecCoinsNonDecreasing("validator commission", before.State.ValidatorCommission, after.State.ValidatorCommission); err != nil {
		return err
	}

	beforeSigning := before.State.SigningInfo
	afterSigning := after.State.SigningInfo
	if beforeSigning.Address != afterSigning.Address || beforeSigning.StartHeight != afterSigning.StartHeight {
		return errors.New("validator signing identity or start height changed across upgrade")
	}
	if beforeSigning.JailedUntil != afterSigning.JailedUntil {
		return errors.New("validator signing jailed_until changed across upgrade")
	}
	if beforeSigning.Tombstoned || afterSigning.Tombstoned {
		return errors.New("validator signing info is tombstoned")
	}
	if afterSigning.IndexOffset < beforeSigning.IndexOffset {
		return fmt.Errorf("validator signing index offset regressed from %d to %d", beforeSigning.IndexOffset, afterSigning.IndexOffset)
	}
	return nil
}

func validateSingleValidatorStopSafety(validators []harness.ValidatorPower, targetAddress string) error {
	targetAddress = strings.ToUpper(strings.TrimSpace(targetAddress))
	if targetAddress == "" {
		return errors.New("target consensus address is required")
	}
	if len(validators) < 2 {
		return errors.New("validator stop safety requires at least two validators")
	}
	totalPower := new(big.Int)
	targetPower := new(big.Int)
	foundTarget := false
	seen := make(map[string]struct{}, len(validators))
	for _, validator := range validators {
		address := strings.ToUpper(strings.TrimSpace(validator.Address))
		if address == "" || validator.Power <= 0 {
			return fmt.Errorf("validator set contains invalid address or power: %+v", validator)
		}
		if _, duplicate := seen[address]; duplicate {
			return fmt.Errorf("validator set contains duplicate address %s", address)
		}
		seen[address] = struct{}{}
		power := big.NewInt(validator.Power)
		totalPower.Add(totalPower, power)
		if address == targetAddress {
			targetPower.Set(power)
			foundTarget = true
		}
	}
	if !foundTarget {
		return fmt.Errorf("target consensus address %s is not present in validator set", targetAddress)
	}
	remainingPower := new(big.Int).Sub(new(big.Int).Set(totalPower), targetPower)
	threeRemaining := new(big.Int).Mul(new(big.Int).Set(remainingPower), big.NewInt(3))
	twoTotal := new(big.Int).Mul(new(big.Int).Set(totalPower), big.NewInt(2))
	if threeRemaining.Cmp(twoTotal) <= 0 {
		return fmt.Errorf(
			"stopping %s leaves power %s of %s, not more than the required two-thirds",
			targetAddress,
			remainingPower.String(),
			totalPower.String(),
		)
	}
	return nil
}

func validateUpgradeStakingMutation(evidence upgradeStakingMutationEvidence) error {
	additional, ok := sdkmath.NewIntFromString(evidence.AdditionalDelegationAmount)
	if !ok || !additional.IsPositive() {
		return fmt.Errorf("additional delegation amount must be positive: %q", evidence.AdditionalDelegationAmount)
	}
	if strings.TrimSpace(evidence.DelegateTxHash) == "" || strings.TrimSpace(evidence.WithdrawRewardTxHash) == "" {
		return errors.New("staking mutation requires delegate and reward-withdraw transaction hashes")
	}
	if strings.EqualFold(evidence.DelegateTxHash, evidence.WithdrawRewardTxHash) {
		return errors.New("staking mutation transaction hashes must be distinct")
	}
	if evidence.BeforeRewardWithdraw.Height <= evidence.Before.Height || evidence.After.Height <= evidence.BeforeRewardWithdraw.Height {
		return fmt.Errorf(
			"staking mutation checkpoint heights must advance: before=%d before_withdraw=%d after=%d",
			evidence.Before.Height,
			evidence.BeforeRewardWithdraw.Height,
			evidence.After.Height,
		)
	}

	before := evidence.Before.State
	beforeWithdraw := evidence.BeforeRewardWithdraw.State
	after := evidence.After.State
	if before.Delegation.DelegatorAddress != beforeWithdraw.Delegation.DelegatorAddress ||
		before.Delegation.ValidatorOperator != beforeWithdraw.Delegation.ValidatorOperator ||
		beforeWithdraw.Delegation.DelegatorAddress != after.Delegation.DelegatorAddress ||
		beforeWithdraw.Delegation.ValidatorOperator != after.Delegation.ValidatorOperator {
		return errors.New("staking mutation changed delegation identity")
	}
	if before.Delegation.Balance.Denom != upgradeStakingDenom ||
		beforeWithdraw.Delegation.Balance.Denom != upgradeStakingDenom ||
		after.Delegation.Balance.Denom != upgradeStakingDenom {
		return errors.New("staking mutation delegation denom changed")
	}
	beforeBalance, ok := sdkmath.NewIntFromString(before.Delegation.Balance.Amount)
	if !ok {
		return fmt.Errorf("invalid pre-mutation delegation balance %q", before.Delegation.Balance.Amount)
	}
	wantBalance := beforeBalance.Add(additional)
	beforeWithdrawBalance, ok := sdkmath.NewIntFromString(beforeWithdraw.Delegation.Balance.Amount)
	if !ok || !beforeWithdrawBalance.Equal(wantBalance) {
		return fmt.Errorf("delegation balance after delegate is %s, want %s", beforeWithdraw.Delegation.Balance.Amount, wantBalance)
	}
	afterBalance, ok := sdkmath.NewIntFromString(after.Delegation.Balance.Amount)
	if !ok || !afterBalance.Equal(wantBalance) {
		return fmt.Errorf("delegation balance after reward withdraw is %s, want %s", after.Delegation.Balance.Amount, wantBalance)
	}
	beforeShares, err := sdkmath.LegacyNewDecFromStr(before.Delegation.Shares)
	if err != nil {
		return fmt.Errorf("decode pre-mutation delegation shares: %w", err)
	}
	beforeWithdrawShares, err := sdkmath.LegacyNewDecFromStr(beforeWithdraw.Delegation.Shares)
	if err != nil || !beforeWithdrawShares.GT(beforeShares) {
		return fmt.Errorf("delegation shares did not increase from %s to %s", before.Delegation.Shares, beforeWithdraw.Delegation.Shares)
	}
	if after.Delegation.Shares != beforeWithdraw.Delegation.Shares {
		return fmt.Errorf("reward withdraw changed delegation shares from %s to %s", beforeWithdraw.Delegation.Shares, after.Delegation.Shares)
	}

	for _, comparison := range []struct {
		label  string
		before string
		after  string
	}{
		{label: "validator tokens", before: before.Validator.Tokens, after: beforeWithdraw.Validator.Tokens},
		{label: "bonded pool tokens", before: before.Pool.BondedTokens, after: beforeWithdraw.Pool.BondedTokens},
	} {
		beforeAmount, parsedBefore := sdkmath.NewIntFromString(comparison.before)
		afterAmount, parsedAfter := sdkmath.NewIntFromString(comparison.after)
		if !parsedBefore || !parsedAfter || !afterAmount.Sub(beforeAmount).Equal(additional) {
			return fmt.Errorf(
				"%s did not increase by %s: before=%s after=%s",
				comparison.label,
				additional,
				comparison.before,
				comparison.after,
			)
		}
	}
	if before.Pool.NotBondedTokens != beforeWithdraw.Pool.NotBondedTokens {
		return errors.New("additional bonded delegation changed not-bonded pool tokens")
	}
	if beforeWithdraw.Validator.Tokens != after.Validator.Tokens || beforeWithdraw.Pool != after.Pool {
		return errors.New("reward withdrawal changed staking validator or pool tokens")
	}

	rewardBefore, err := sdkmath.LegacyNewDecFromStr(beforeWithdraw.DelegatorRewards.AmountOf(upgradeStakingDenom))
	if err != nil || !rewardBefore.IsPositive() {
		return fmt.Errorf("delegator reward before withdrawal must be positive: %s", beforeWithdraw.DelegatorRewards.AmountOf(upgradeStakingDenom))
	}
	rewardAfter, err := sdkmath.LegacyNewDecFromStr(after.DelegatorRewards.AmountOf(upgradeStakingDenom))
	if err != nil {
		return fmt.Errorf("decode delegator reward after withdrawal: %w", err)
	}
	if !rewardAfter.LT(rewardBefore) {
		return fmt.Errorf("delegator reward did not decrease after withdrawal: before=%s after=%s", rewardBefore, rewardAfter)
	}
	recipientBefore, ok := sdkmath.NewIntFromString(evidence.BeforeRewardWithdraw.RecipientBankBalance)
	if !ok {
		return fmt.Errorf("invalid reward recipient balance before withdrawal %q", evidence.BeforeRewardWithdraw.RecipientBankBalance)
	}
	recipientAfter, ok := sdkmath.NewIntFromString(evidence.After.RecipientBankBalance)
	if !ok || !recipientAfter.GT(recipientBefore) {
		return fmt.Errorf("reward recipient balance did not increase on withdrawal: before=%s after=%s", recipientBefore, evidence.After.RecipientBankBalance)
	}
	if beforeWithdraw.Validator.Jailed || after.Validator.Jailed ||
		beforeWithdraw.SigningInfo.Tombstoned || after.SigningInfo.Tombstoned {
		return errors.New("staking mutation left validator jailed or tombstoned")
	}
	return nil
}

func validateUpgradeValidatorLiveness(evidence upgradeValidatorLivenessEvidence) error {
	if evidence.ValidatorIndex < 0 {
		return fmt.Errorf("validator liveness index must not be negative: %d", evidence.ValidatorIndex)
	}
	if err := validateSingleValidatorStopSafety(evidence.ValidatorSet, evidence.TargetConsensusAddress); err != nil {
		return err
	}
	if evidence.MinimumStoppedBlocks <= 0 {
		return errors.New("minimum stopped blocks must be positive")
	}
	if evidence.FaultWindow.EndHeight-evidence.FaultWindow.StartHeight < evidence.MinimumStoppedBlocks ||
		evidence.FaultWindow.TargetHeight < evidence.FaultWindow.StartHeight+evidence.MinimumStoppedBlocks {
		return fmt.Errorf(
			"validator stop window advanced %d blocks, want at least %d",
			evidence.FaultWindow.EndHeight-evidence.FaultWindow.StartHeight,
			evidence.MinimumStoppedBlocks,
		)
	}
	if evidence.BeforeStop.Height > evidence.FaultWindow.StartHeight ||
		evidence.WhileStopped.Height < evidence.FaultWindow.EndHeight ||
		evidence.AfterRejoin.Height <= evidence.WhileStopped.Height {
		return fmt.Errorf(
			"validator liveness checkpoint heights are inconsistent: before=%d window=%d..%d stopped=%d rejoined=%d",
			evidence.BeforeStop.Height,
			evidence.FaultWindow.StartHeight,
			evidence.FaultWindow.EndHeight,
			evidence.WhileStopped.Height,
			evidence.AfterRejoin.Height,
		)
	}

	before := evidence.BeforeStop.State
	stopped := evidence.WhileStopped.State
	rejoined := evidence.AfterRejoin.State
	if before.Validator.OperatorAddress == "" ||
		before.Validator.OperatorAddress != stopped.Validator.OperatorAddress ||
		stopped.Validator.OperatorAddress != rejoined.Validator.OperatorAddress {
		return errors.New("validator operator changed during stop/rejoin")
	}
	if before.Validator.Jailed || stopped.Validator.Jailed || rejoined.Validator.Jailed {
		return errors.New("validator became jailed during safe stop/rejoin")
	}
	if rejoined.Validator.Status != "BOND_STATUS_BONDED" {
		return fmt.Errorf("rejoined validator status is %q, want BOND_STATUS_BONDED", rejoined.Validator.Status)
	}
	beforeSigning := before.SigningInfo
	stoppedSigning := stopped.SigningInfo
	rejoinedSigning := rejoined.SigningInfo
	if beforeSigning.Address == "" ||
		beforeSigning.Address != stoppedSigning.Address ||
		stoppedSigning.Address != rejoinedSigning.Address ||
		beforeSigning.StartHeight != stoppedSigning.StartHeight ||
		stoppedSigning.StartHeight != rejoinedSigning.StartHeight {
		return errors.New("validator signing identity changed during stop/rejoin")
	}
	if beforeSigning.JailedUntil != stoppedSigning.JailedUntil || stoppedSigning.JailedUntil != rejoinedSigning.JailedUntil {
		return errors.New("validator signing jailed_until changed during safe stop/rejoin")
	}
	if beforeSigning.Tombstoned || stoppedSigning.Tombstoned || rejoinedSigning.Tombstoned {
		return errors.New("validator signing info became tombstoned during safe stop/rejoin")
	}
	if stoppedSigning.IndexOffset-beforeSigning.IndexOffset < evidence.MinimumStoppedBlocks {
		return fmt.Errorf(
			"validator signing index advanced %d while stopped, want at least %d",
			stoppedSigning.IndexOffset-beforeSigning.IndexOffset,
			evidence.MinimumStoppedBlocks,
		)
	}
	if stoppedSigning.MissedBlocksCounter-beforeSigning.MissedBlocksCounter < evidence.MinimumStoppedBlocks {
		return fmt.Errorf(
			"validator missed blocks counter advanced %d while stopped, want at least %d",
			stoppedSigning.MissedBlocksCounter-beforeSigning.MissedBlocksCounter,
			evidence.MinimumStoppedBlocks,
		)
	}
	if rejoinedSigning.IndexOffset < stoppedSigning.IndexOffset ||
		rejoinedSigning.MissedBlocksCounter < stoppedSigning.MissedBlocksCounter {
		return errors.New("validator signing state regressed after rejoin")
	}
	if evidence.SignedCommitHeight <= evidence.FaultWindow.EndHeight ||
		evidence.SignedCommitHeight > evidence.AfterRejoin.Height {
		return fmt.Errorf(
			"rejoined validator commit signature height %d is outside (%d, %d]",
			evidence.SignedCommitHeight,
			evidence.FaultWindow.EndHeight,
			evidence.AfterRejoin.Height,
		)
	}
	if len(evidence.RejoinHistory) < 2 {
		return fmt.Errorf("validator rejoin history has %d observations, want at least 2", len(evidence.RejoinHistory))
	}
	wantHistory := evidence.RejoinHistory[0]
	if wantHistory.Height != evidence.AfterRejoin.Height || wantHistory.BlockID == "" || wantHistory.AppHash == "" {
		return fmt.Errorf(
			"validator rejoin history is incomplete at checkpoint height %d: %+v",
			evidence.AfterRejoin.Height,
			wantHistory,
		)
	}
	seenHistoryNodes := make(map[string]struct{}, len(evidence.RejoinHistory))
	for _, observed := range evidence.RejoinHistory {
		if strings.TrimSpace(observed.Node) == "" {
			return errors.New("validator rejoin history contains an unnamed node")
		}
		if _, duplicate := seenHistoryNodes[observed.Node]; duplicate {
			return fmt.Errorf("validator rejoin history contains duplicate node %q", observed.Node)
		}
		seenHistoryNodes[observed.Node] = struct{}{}
		if observed.Height != wantHistory.Height ||
			!strings.EqualFold(observed.BlockID, wantHistory.BlockID) ||
			!strings.EqualFold(observed.AppHash, wantHistory.AppHash) {
			return fmt.Errorf(
				"validator rejoin history diverged at node %s: want height=%d block=%s app=%s, got height=%d block=%s app=%s",
				observed.Node,
				wantHistory.Height,
				wantHistory.BlockID,
				wantHistory.AppHash,
				observed.Height,
				observed.BlockID,
				observed.AppHash,
			)
		}
	}
	return nil
}

func validateDecCoinsNonDecreasing(label string, before, after upgradeStakingDecCoins) error {
	for _, expected := range before {
		beforeAmount, err := sdkmath.LegacyNewDecFromStr(expected.Amount)
		if err != nil {
			return fmt.Errorf("%s before amount for %s: %w", label, expected.Denom, err)
		}
		afterText := after.AmountOf(expected.Denom)
		afterAmount, err := sdkmath.LegacyNewDecFromStr(afterText)
		if err != nil {
			return fmt.Errorf("%s after amount for %s: %w", label, expected.Denom, err)
		}
		if afterAmount.LT(beforeAmount) {
			return fmt.Errorf("%s for %s decreased from %s to %s", label, expected.Denom, expected.Amount, afterText)
		}
	}
	return nil
}

type upgradeConsensusPubKeyIdentity struct {
	TypeURL string
	Key     []byte
}

func equalUpgradeConsensusPubKeys(left, right json.RawMessage) (bool, error) {
	leftIdentity, err := decodeUpgradeConsensusPubKeyIdentity(left)
	if err != nil {
		return false, fmt.Errorf("decode left public key: %w", err)
	}
	rightIdentity, err := decodeUpgradeConsensusPubKeyIdentity(right)
	if err != nil {
		return false, fmt.Errorf("decode right public key: %w", err)
	}
	return leftIdentity.TypeURL == rightIdentity.TypeURL && bytes.Equal(leftIdentity.Key, rightIdentity.Key), nil
}

func decodeUpgradeConsensusPubKeyIdentity(raw json.RawMessage) (upgradeConsensusPubKeyIdentity, error) {
	document, err := decodeUpgradeJSON(raw, "validator consensus public key")
	if err != nil {
		return upgradeConsensusPubKeyIdentity{}, err
	}
	object, ok := document.(map[string]any)
	if !ok {
		return upgradeConsensusPubKeyIdentity{}, errors.New("expected a JSON object")
	}
	legacyType, legacyKey := jsonText(object["@type"]), jsonText(object["key"])
	currentType, currentKey := jsonText(object["type"]), jsonText(object["value"])
	legacyPresent := legacyType != "" || legacyKey != ""
	currentPresent := currentType != "" || currentKey != ""
	if legacyPresent == currentPresent {
		return upgradeConsensusPubKeyIdentity{}, errors.New("expected exactly one @type/key or type/value encoding")
	}
	typeURL, encodedKey := legacyType, legacyKey
	if currentPresent {
		typeURL, encodedKey = currentType, currentKey
	}
	if typeURL == "" || encodedKey == "" {
		return upgradeConsensusPubKeyIdentity{}, errors.New("public key encoding is missing its type or value")
	}
	key, err := base64.StdEncoding.Strict().DecodeString(encodedKey)
	if err != nil {
		return upgradeConsensusPubKeyIdentity{}, fmt.Errorf("public key value is not non-empty canonical base64: %w", err)
	}
	if len(key) == 0 {
		return upgradeConsensusPubKeyIdentity{}, errors.New("public key value is not non-empty canonical base64")
	}
	return upgradeConsensusPubKeyIdentity{TypeURL: typeURL, Key: key}, nil
}

func decodeUpgradeStakingQueries(
	responses upgradeStakingQueryResponses,
	fixture upgradeStakingFixture,
) (upgradeStakingState, error) {
	if strings.TrimSpace(fixture.DelegatorAddress) == "" {
		return upgradeStakingState{}, errors.New("staking fixture delegator address is required")
	}
	if strings.TrimSpace(fixture.ValidatorOperator) == "" {
		return upgradeStakingState{}, errors.New("staking fixture validator operator is required")
	}

	delegation, err := decodeUpgradeDelegation(responses.Delegation)
	if err != nil {
		return upgradeStakingState{}, err
	}
	if delegation.DelegatorAddress != fixture.DelegatorAddress {
		return upgradeStakingState{}, fmt.Errorf("delegation address %q, want %q", delegation.DelegatorAddress, fixture.DelegatorAddress)
	}
	if delegation.ValidatorOperator != fixture.ValidatorOperator {
		return upgradeStakingState{}, fmt.Errorf("delegation validator %q, want %q", delegation.ValidatorOperator, fixture.ValidatorOperator)
	}
	validator, err := decodeUpgradeValidator(responses.Validator)
	if err != nil {
		return upgradeStakingState{}, err
	}
	if validator.OperatorAddress != fixture.ValidatorOperator {
		return upgradeStakingState{}, fmt.Errorf("validator operator %q, want %q", validator.OperatorAddress, fixture.ValidatorOperator)
	}
	pool, err := decodeUpgradeStakingPool(responses.Pool)
	if err != nil {
		return upgradeStakingState{}, err
	}
	delegatorRewards, err := decodeUpgradeDelegatorRewards(responses.DelegatorRewards, fixture.ValidatorOperator)
	if err != nil {
		return upgradeStakingState{}, err
	}
	outstandingRewards, err := decodeUpgradeDecCoins(responses.OutstandingRewards, "rewards", "validator outstanding rewards")
	if err != nil {
		return upgradeStakingState{}, err
	}
	commission, err := decodeUpgradeDecCoins(responses.Commission, "commission", "validator commission")
	if err != nil {
		return upgradeStakingState{}, err
	}
	signingInfo, err := decodeUpgradeSigningInfo(responses.SigningInfo)
	if err != nil {
		return upgradeStakingState{}, err
	}
	return upgradeStakingState{
		Delegation:          delegation,
		Validator:           validator,
		Pool:                pool,
		DelegatorRewards:    delegatorRewards,
		OutstandingRewards:  outstandingRewards,
		ValidatorCommission: commission,
		SigningInfo:         signingInfo,
	}, nil
}

func decodeUpgradeDelegation(raw []byte) (upgradeDelegationState, error) {
	document, err := decodeUpgradeJSON(raw, "delegation")
	if err != nil {
		return upgradeDelegationState{}, err
	}
	object, ok := findUpgradeJSONObject(document, func(candidate map[string]any) bool {
		_, hasDelegation := candidate["delegation"].(map[string]any)
		_, hasBalance := candidate["balance"].(map[string]any)
		return hasDelegation && hasBalance
	})
	if !ok {
		return upgradeDelegationState{}, errors.New("delegation query has no delegation response")
	}
	delegation := object["delegation"].(map[string]any)
	balance := object["balance"].(map[string]any)
	state := upgradeDelegationState{
		DelegatorAddress:  jsonText(delegation["delegator_address"]),
		ValidatorOperator: jsonText(delegation["validator_address"]),
		Shares:            jsonText(delegation["shares"]),
		Balance: upgradeStakingCoin{
			Denom:  jsonText(balance["denom"]),
			Amount: jsonText(balance["amount"]),
		},
	}
	if state.DelegatorAddress == "" || state.ValidatorOperator == "" || state.Shares == "" {
		return upgradeDelegationState{}, errors.New("delegation query is missing addresses or shares")
	}
	if state.Balance.Denom != upgradeStakingDenom {
		return upgradeDelegationState{}, fmt.Errorf("delegation balance denom %q, want %q", state.Balance.Denom, upgradeStakingDenom)
	}
	if _, ok := sdkmath.NewIntFromString(state.Balance.Amount); !ok {
		return upgradeDelegationState{}, fmt.Errorf("invalid delegation balance %q", state.Balance.Amount)
	}
	if _, err := sdkmath.LegacyNewDecFromStr(state.Shares); err != nil {
		return upgradeDelegationState{}, fmt.Errorf("invalid delegation shares %q: %w", state.Shares, err)
	}
	return state, nil
}

func decodeUpgradeValidator(raw []byte) (upgradeValidatorState, error) {
	document, err := decodeUpgradeJSON(raw, "validator")
	if err != nil {
		return upgradeValidatorState{}, err
	}
	object, ok := findUpgradeJSONObject(document, func(candidate map[string]any) bool {
		_, hasOperator := candidate["operator_address"]
		_, hasPubKey := candidate["consensus_pubkey"].(map[string]any)
		_, hasTokens := candidate["tokens"]
		return hasOperator && hasPubKey && hasTokens
	})
	if !ok {
		return upgradeValidatorState{}, errors.New("validator query has no validator state")
	}
	consensusPubKey, err := json.Marshal(object["consensus_pubkey"])
	if err != nil {
		return upgradeValidatorState{}, fmt.Errorf("encode validator consensus pubkey: %w", err)
	}
	consensusPubKeyIdentity, err := decodeUpgradeConsensusPubKeyIdentity(consensusPubKey)
	if err != nil {
		return upgradeValidatorState{}, fmt.Errorf("decode validator consensus pubkey: %w", err)
	}
	consensusPubKey, err = json.Marshal(struct {
		TypeURL string `json:"@type"`
		Key     string `json:"key"`
	}{
		TypeURL: consensusPubKeyIdentity.TypeURL,
		Key:     base64.StdEncoding.EncodeToString(consensusPubKeyIdentity.Key),
	})
	if err != nil {
		return upgradeValidatorState{}, fmt.Errorf("encode canonical validator consensus pubkey: %w", err)
	}
	jailed, err := decodeUpgradeProtoJSONBool(object, "jailed")
	if err != nil {
		return upgradeValidatorState{}, fmt.Errorf("validator query: %w", err)
	}
	state := upgradeValidatorState{
		OperatorAddress: jsonText(object["operator_address"]),
		ConsensusPubKey: consensusPubKey,
		Jailed:          jailed,
		Status:          jsonText(object["status"]),
		Tokens:          jsonText(object["tokens"]),
		DelegatorShares: jsonText(object["delegator_shares"]),
	}
	if state.OperatorAddress == "" || state.Status == "" || state.Tokens == "" || state.DelegatorShares == "" {
		return upgradeValidatorState{}, errors.New("validator query is missing operator, status, tokens, or shares")
	}
	if _, ok := sdkmath.NewIntFromString(state.Tokens); !ok {
		return upgradeValidatorState{}, fmt.Errorf("invalid validator tokens %q", state.Tokens)
	}
	if _, err := sdkmath.LegacyNewDecFromStr(state.DelegatorShares); err != nil {
		return upgradeValidatorState{}, fmt.Errorf("invalid validator delegator shares %q: %w", state.DelegatorShares, err)
	}
	return state, nil
}

func decodeUpgradeStakingPool(raw []byte) (upgradeStakingPoolState, error) {
	document, err := decodeUpgradeJSON(raw, "staking pool")
	if err != nil {
		return upgradeStakingPoolState{}, err
	}
	object, ok := findUpgradeJSONObject(document, func(candidate map[string]any) bool {
		_, hasBonded := candidate["bonded_tokens"]
		_, hasNotBonded := candidate["not_bonded_tokens"]
		return hasBonded && hasNotBonded
	})
	if !ok {
		return upgradeStakingPoolState{}, errors.New("staking pool query has no pool state")
	}
	state := upgradeStakingPoolState{
		BondedTokens:    jsonText(object["bonded_tokens"]),
		NotBondedTokens: jsonText(object["not_bonded_tokens"]),
	}
	for label, value := range map[string]string{"bonded": state.BondedTokens, "not bonded": state.NotBondedTokens} {
		if _, ok := sdkmath.NewIntFromString(value); !ok {
			return upgradeStakingPoolState{}, fmt.Errorf("invalid %s pool tokens %q", label, value)
		}
	}
	return state, nil
}

func decodeUpgradeSigningInfo(raw []byte) (upgradeSigningInfoState, error) {
	document, err := decodeUpgradeJSON(raw, "slashing signing info")
	if err != nil {
		return upgradeSigningInfoState{}, err
	}
	object, ok := findUpgradeJSONObject(document, func(candidate map[string]any) bool {
		_, hasAddress := candidate["address"]
		_, hasJailedUntil := candidate["jailed_until"]
		return hasAddress && hasJailedUntil
	})
	if !ok {
		return upgradeSigningInfoState{}, errors.New("slashing query has no validator signing info")
	}
	tombstoned, err := decodeUpgradeProtoJSONBool(object, "tombstoned")
	if err != nil {
		return upgradeSigningInfoState{}, fmt.Errorf("slashing signing info: %w", err)
	}
	startHeight, err := decodeUpgradeProtoJSONInt64(object, "start_height")
	if err != nil {
		return upgradeSigningInfoState{}, fmt.Errorf("decode signing start height: %w", err)
	}
	indexOffset, err := decodeUpgradeProtoJSONInt64(object, "index_offset")
	if err != nil {
		return upgradeSigningInfoState{}, fmt.Errorf("decode signing index offset: %w", err)
	}
	missedBlocks, err := decodeUpgradeProtoJSONInt64(object, "missed_blocks_counter")
	if err != nil {
		return upgradeSigningInfoState{}, fmt.Errorf("decode missed blocks counter: %w", err)
	}
	state := upgradeSigningInfoState{
		Address:             jsonText(object["address"]),
		StartHeight:         startHeight,
		IndexOffset:         indexOffset,
		JailedUntil:         jsonText(object["jailed_until"]),
		Tombstoned:          tombstoned,
		MissedBlocksCounter: missedBlocks,
	}
	if state.Address == "" || state.JailedUntil == "" {
		return upgradeSigningInfoState{}, errors.New("slashing signing info is missing address or jailed_until")
	}
	return state, nil
}

// Protobuf JSON omits scalar fields whose value is the protobuf default. For a
// bool that makes an absent field semantically identical to false. Still reject
// a present value with the wrong JSON type so malformed query evidence cannot
// silently become a healthy state.
func decodeUpgradeProtoJSONBool(object map[string]any, field string) (bool, error) {
	value, present := object[field]
	if !present {
		return false, nil
	}
	decoded, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("%s is not a boolean", field)
	}
	return decoded, nil
}

func decodeUpgradeProtoJSONInt64(object map[string]any, field string) (int64, error) {
	value, present := object[field]
	if !present {
		return 0, nil
	}
	decoded, err := jsonInt64(value)
	if err != nil {
		return 0, fmt.Errorf("%s is not a valid int64: %w", field, err)
	}
	return decoded, nil
}

func decodeUpgradeDecCoins(raw []byte, field, label string) (upgradeStakingDecCoins, error) {
	document, err := decodeUpgradeJSON(raw, label)
	if err != nil {
		return nil, err
	}
	object, ok := findUpgradeJSONObject(document, func(candidate map[string]any) bool {
		_, found := candidate[field].([]any)
		return found
	})
	if !ok {
		return nil, fmt.Errorf("%s query has no %s coin list", label, field)
	}
	values := object[field].([]any)
	coins := make(upgradeStakingDecCoins, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for index, value := range values {
		var denom, amount string
		switch coin := value.(type) {
		case map[string]any:
			denom = jsonText(coin["denom"])
			amount = jsonText(coin["amount"])
		case string:
			parsed, err := sdk.ParseDecCoin(coin)
			if err != nil {
				return nil, fmt.Errorf("%s coin %d has invalid compact encoding %q: %w", label, index, coin, err)
			}
			denom = parsed.Denom
			amount = parsed.Amount.String()
		default:
			return nil, fmt.Errorf("%s coin %d is neither an object nor a compact string", label, index)
		}
		if denom == "" || amount == "" {
			return nil, fmt.Errorf("%s coin %d is missing denom or amount", label, index)
		}
		if err := sdk.ValidateDenom(denom); err != nil {
			return nil, fmt.Errorf("%s coin %d has invalid denom %q: %w", label, index, denom, err)
		}
		if _, duplicate := seen[denom]; duplicate {
			return nil, fmt.Errorf("%s contains duplicate denom %q", label, denom)
		}
		decimal, err := sdkmath.LegacyNewDecFromStr(amount)
		if err != nil || decimal.IsNegative() {
			return nil, fmt.Errorf("%s has invalid amount %q for %s", label, amount, denom)
		}
		seen[denom] = struct{}{}
		coins = append(coins, upgradeStakingCoin{Denom: denom, Amount: amount})
	}
	sort.Slice(coins, func(i, j int) bool { return coins[i].Denom < coins[j].Denom })
	return coins, nil
}

func decodeUpgradeDelegatorRewards(raw []byte, validatorOperator string) (upgradeStakingDecCoins, error) {
	if strings.TrimSpace(validatorOperator) == "" {
		return nil, errors.New("delegator rewards validator operator is required")
	}
	document, err := decodeUpgradeJSON(raw, "delegator rewards")
	if err != nil {
		return nil, err
	}
	object, ok := findUpgradeJSONObject(document, func(candidate map[string]any) bool {
		_, found := candidate["rewards"].([]any)
		return found
	})
	if !ok {
		return nil, errors.New("delegator rewards query has no rewards list")
	}
	values := object["rewards"].([]any)
	var selected upgradeStakingDecCoins
	matches := 0
	for index, value := range values {
		entry, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("delegator rewards entry %d is not an object", index)
		}
		validatorAddress := jsonText(entry["validator_address"])
		rewardValues, ok := entry["reward"].([]any)
		if validatorAddress == "" || !ok {
			return nil, fmt.Errorf("delegator rewards entry %d is missing validator_address or reward", index)
		}
		if validatorAddress != validatorOperator {
			continue
		}
		matches++
		encoded, err := json.Marshal(map[string]any{"reward": rewardValues})
		if err != nil {
			return nil, fmt.Errorf("encode delegator rewards for %s: %w", validatorOperator, err)
		}
		selected, err = decodeUpgradeDecCoins(encoded, "reward", "delegator validator rewards")
		if err != nil {
			return nil, err
		}
	}
	if matches != 1 {
		return nil, fmt.Errorf("delegator rewards has %d entries for validator %s, want 1", matches, validatorOperator)
	}
	return selected, nil
}

func decodeUpgradeJSON(raw []byte, label string) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var document any
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("decode %s JSON: %w", label, err)
	}
	if decoder.Decode(&struct{}{}) == nil {
		return nil, fmt.Errorf("decode %s JSON: multiple values", label)
	}
	return document, nil
}

func findUpgradeJSONObject(value any, match func(map[string]any) bool) (map[string]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		if match(typed) {
			return typed, true
		}
		for _, child := range typed {
			if found, ok := findUpgradeJSONObject(child, match); ok {
				return found, true
			}
		}
	case []any:
		for _, child := range typed {
			if found, ok := findUpgradeJSONObject(child, match); ok {
				return found, true
			}
		}
	}
	return nil, false
}

func jsonText(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	default:
		return ""
	}
}

func jsonInt64(value any) (int64, error) {
	text := jsonText(value)
	if text == "" {
		return 0, fmt.Errorf("expected string or number, got %T", value)
	}
	return strconv.ParseInt(text, 10, 64)
}
