package e2e_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/types/bech32"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	upgradeP0SlashingValidatorIndex = 3
	upgradeP0SlashingFee            = "2500000"
	upgradeP0SlashingMinimumMisses  = 21
)

type upgradeP0SlashingFixture struct {
	ValidatorIndex   int             `json:"validator_index"`
	OperatorAddress  string          `json:"operator_address"`
	ConsensusAddress string          `json:"consensus_address"`
	ConsensusPubKey  json.RawMessage `json:"consensus_pubkey"`
}

func prepareUpgradeP0Slashing(
	ctx context.Context,
	network *harness.Network,
) (upgradeP0SlashingFixture, upgradeP0SlashingEvidence, error) {
	if err := validateUpgradeStakingNetwork(network, upgradeP0SlashingValidatorIndex); err != nil {
		return upgradeP0SlashingFixture{}, upgradeP0SlashingEvidence{}, err
	}
	operator, err := queryUpgradeValidatorOperator(ctx, network.Chain.Validators[upgradeP0SlashingValidatorIndex])
	if err != nil {
		return upgradeP0SlashingFixture{}, upgradeP0SlashingEvidence{}, err
	}
	validatorRaw, err := network.FullNodeCLIQuery(
		ctx,
		"upgrade-p0-slashing-prepare-validator",
		"staking", "validator", operator,
	)
	if err != nil {
		return upgradeP0SlashingFixture{}, upgradeP0SlashingEvidence{}, err
	}
	validator, err := decodeUpgradeValidator(validatorRaw)
	if err != nil {
		return upgradeP0SlashingFixture{}, upgradeP0SlashingEvidence{}, err
	}
	signingRaw, err := network.FullNodeCLIQuery(
		ctx,
		"upgrade-p0-slashing-prepare-signing-info",
		"slashing", "signing-info", strings.TrimSpace(string(validator.ConsensusPubKey)),
	)
	if err != nil {
		return upgradeP0SlashingFixture{}, upgradeP0SlashingEvidence{}, err
	}
	signing, err := decodeUpgradeSigningInfo(signingRaw)
	if err != nil {
		return upgradeP0SlashingFixture{}, upgradeP0SlashingEvidence{}, err
	}
	fixture := upgradeP0SlashingFixture{
		ValidatorIndex:   upgradeP0SlashingValidatorIndex,
		OperatorAddress:  operator,
		ConsensusAddress: signing.Address,
		ConsensusPubKey:  append(json.RawMessage(nil), validator.ConsensusPubKey...),
	}
	before, err := captureUpgradeP0SlashingCheckpoint(ctx, network, fixture, "before")
	if err != nil {
		return fixture, upgradeP0SlashingEvidence{}, err
	}
	if before.Validator.Jailed || before.SigningInfo.Tombstoned || before.ValidatorPower <= 0 {
		return fixture, upgradeP0SlashingEvidence{}, errors.New("P0 slashing target is not a healthy active validator")
	}
	evidence := upgradeP0SlashingEvidence{
		ValidatorIndex: fixture.ValidatorIndex,
		Before:         before,
	}
	if err := network.WriteArtifactJSON("upgrade/slashing-jail/preparation.json", map[string]any{
		"fixture":  fixture,
		"evidence": evidence,
	}); err != nil {
		return fixture, evidence, err
	}
	return fixture, evidence, nil
}

func stopUpgradeP0SlashingTargetAtHalt(
	ctx context.Context,
	network *harness.Network,
	upgradeHeight int64,
	evidence *upgradeP0SlashingEvidence,
) error {
	if evidence == nil {
		return errors.New("P0 slashing evidence is required")
	}
	if err := network.StopQuorumValidator(
		ctx,
		"upgrade-p0-slashing-stop-at-halt",
		evidence.ValidatorIndex,
	); err != nil {
		return err
	}
	evidence.UpgradeHeight = upgradeHeight
	evidence.StoppedAt = upgradeHeight
	return network.WriteArtifactJSON("upgrade/slashing-jail/stopped-at-halt.json", evidence)
}

func waitForUpgradeP0SlashingJail(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeP0SlashingFixture,
	evidence *upgradeP0SlashingEvidence,
) error {
	if evidence == nil || evidence.UpgradeHeight <= 0 {
		return errors.New("P0 slashing halt evidence is required")
	}
	fullNode := network.Chain.FullNodes[0]
	nextHeight, err := fullNode.Height(ctx)
	if err != nil {
		return err
	}
	// The configured threshold is 21 missed blocks. Leave enough polling
	// headroom for delayed validator-set power updates on slower Docker hosts.
	maximumHeight := nextHeight + 40
	maxMissed := evidence.Before.SigningInfo.MissedBlocksCounter
	for nextHeight <= maximumHeight {
		if err := network.WaitForNodeHeight(ctx, fullNode, nextHeight); err != nil {
			return err
		}
		checkpoint, captureErr := captureUpgradeP0SlashingCheckpoint(
			ctx,
			network,
			fixture,
			"jail-poll-"+strconv.FormatInt(nextHeight, 10),
		)
		if captureErr != nil {
			return captureErr
		}
		if checkpoint.SigningInfo.MissedBlocksCounter > maxMissed {
			maxMissed = checkpoint.SigningInfo.MissedBlocksCounter
		}
		if checkpoint.Validator.Jailed && checkpoint.ValidatorPower == 0 {
			evidence.OutageBlocksObserved = checkpoint.Height - evidence.StoppedAt
			evidence.MissedBlocksObserved = maxMissed
			evidence.Jailed = checkpoint
			if checkpoint.Height <= evidence.UpgradeHeight || checkpoint.SigningInfo.Tombstoned {
				return errors.New("P0 downtime jail did not occur safely after the upgrade boundary")
			}
			return network.WriteArtifactJSON("upgrade/slashing-jail/jailed.json", evidence)
		}
		nextHeight = checkpoint.Height + 1
	}
	return fmt.Errorf("validator %s was not jailed through height %d", fixture.OperatorAddress, maximumHeight)
}

func exerciseUpgradeP0Unjail(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeP0SlashingFixture,
	evidence *upgradeP0SlashingEvidence,
) error {
	if evidence == nil || !evidence.Jailed.Validator.Jailed {
		return errors.New("P0 jailed checkpoint is required")
	}
	validatorNode := network.Chain.Validators[fixture.ValidatorIndex]
	early, err := network.BroadcastAndWaitTxExpectDeliverFailure(
		ctx,
		"upgrade-p0-slashing-early-unjail",
		validatorNode,
		"validator",
		"slashing",
		4,
		"slashing", "unjail",
		"--fees", upgradeP0SlashingFee+upgradeStakingDenom,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return err
	}
	evidence.EarlyUnjail = *early
	jailedUntil, err := time.Parse(time.RFC3339Nano, evidence.Jailed.SigningInfo.JailedUntil)
	if err != nil {
		return fmt.Errorf("decode P0 jailed_until: %w", err)
	}
	wait := time.Until(jailedUntil.Add(time.Second))
	if wait > 0 {
		timer := time.NewTimer(wait)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for P0 jailed_until %s: %w", jailedUntil, ctx.Err())
		case <-timer.C:
		}
	}
	unjail, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-p0-slashing-unjail",
		validatorNode,
		"validator",
		"slashing", "unjail",
		"--fees", upgradeP0SlashingFee+upgradeStakingDenom,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return err
	}
	evidence.UnjailTxHash = unjail.TxHash
	unjailCheckpoint, err := captureUpgradeP0SlashingCheckpoint(ctx, network, fixture, "unjailed")
	if err != nil {
		return err
	}
	evidence.Unjailed = unjailCheckpoint
	_, consensusBytes, err := bech32.DecodeAndConvert(fixture.ConsensusAddress)
	if err != nil {
		return fmt.Errorf("decode P0 consensus address: %w", err)
	}
	signedHeight, err := waitForUpgradeValidatorCommitSignature(
		ctx,
		network,
		network.Chain.FullNodes[0],
		consensusBytes,
		unjailCheckpoint.Height,
		16,
	)
	if err != nil {
		return err
	}
	evidence.SignedCommitHeight = signedHeight
	rejoined, err := captureUpgradeP0SlashingCheckpoint(ctx, network, fixture, "rejoined")
	if err != nil {
		return err
	}
	evidence.Rejoined = rejoined
	return network.WriteArtifactJSON("upgrade/slashing-jail/unjail-and-rejoin.json", evidence)
}

func captureUpgradeP0SlashingAfterRestart(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeP0SlashingFixture,
	evidence *upgradeP0SlashingEvidence,
) error {
	checkpoint, err := captureUpgradeP0SlashingCheckpoint(ctx, network, fixture, "post-restart")
	if err != nil {
		return err
	}
	evidence.PostRestart = checkpoint
	if err := validateUpgradeP0SlashingEvidence(*evidence); err != nil {
		return err
	}
	return network.WriteArtifactJSON("upgrade/slashing-jail/post-restart.json", evidence)
}

func captureUpgradeP0SlashingCheckpoint(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeP0SlashingFixture,
	phase string,
) (upgradeP0SlashingCheckpoint, error) {
	if network == nil || network.Chain == nil || len(network.Chain.FullNodes) == 0 {
		return upgradeP0SlashingCheckpoint{}, errors.New("P0 slashing checkpoint requires a full node")
	}
	fullNode := network.Chain.FullNodes[0]
	height, err := fullNode.Height(ctx)
	if err != nil {
		return upgradeP0SlashingCheckpoint{}, err
	}
	query := func(suffix string, command ...string) (json.RawMessage, error) {
		command = append(command, "--height", strconv.FormatInt(height, 10))
		return network.FullNodeCLIQuery(ctx, "upgrade-p0-slashing-"+phase+"-"+suffix, command...)
	}
	validatorRaw, err := query("validator", "staking", "validator", fixture.OperatorAddress)
	if err != nil {
		return upgradeP0SlashingCheckpoint{}, err
	}
	validator, err := decodeUpgradeValidator(validatorRaw)
	if err != nil {
		return upgradeP0SlashingCheckpoint{}, err
	}
	equalKey, err := equalUpgradeConsensusPubKeys(validator.ConsensusPubKey, fixture.ConsensusPubKey)
	if err != nil || !equalKey {
		return upgradeP0SlashingCheckpoint{}, errors.New("P0 slashing validator consensus public key changed")
	}
	signingRaw, err := query("signing-info", "slashing", "signing-info", strings.TrimSpace(string(validator.ConsensusPubKey)))
	if err != nil {
		return upgradeP0SlashingCheckpoint{}, err
	}
	signing, err := decodeUpgradeSigningInfo(signingRaw)
	if err != nil {
		return upgradeP0SlashingCheckpoint{}, err
	}
	if signing.Address != fixture.ConsensusAddress {
		return upgradeP0SlashingCheckpoint{}, fmt.Errorf("P0 signing address %s, want %s", signing.Address, fixture.ConsensusAddress)
	}
	power, err := upgradeP0ValidatorPower(ctx, network, fullNode, height, signing.Address)
	if err != nil {
		return upgradeP0SlashingCheckpoint{}, err
	}
	checkpoint := upgradeP0SlashingCheckpoint{
		Phase:          phase,
		Height:         height,
		RecordedAt:     time.Now().UTC(),
		ValidatorPower: power,
		Validator:      validator,
		SigningInfo:    signing,
	}
	if err := network.WriteArtifactJSON("upgrade/slashing-jail/checkpoints/"+phase+".json", checkpoint); err != nil {
		return checkpoint, err
	}
	return checkpoint, nil
}

func upgradeP0ValidatorPower(
	ctx context.Context,
	network *harness.Network,
	observer interface {
		Height(context.Context) (int64, error)
	},
	height int64,
	consensusAddress string,
) (int64, error) {
	fullNode := network.Chain.FullNodes[0]
	if observer == nil || fullNode == nil {
		return 0, errors.New("P0 validator power observer is required")
	}
	hrp, consensusBytes, err := bech32.DecodeAndConvert(consensusAddress)
	if err != nil || hrp != "panaceavalcons" {
		return 0, fmt.Errorf("decode P0 consensus address %q", consensusAddress)
	}
	want := strings.ToUpper(hex.EncodeToString(consensusBytes))
	validators, err := network.ValidatorSet(ctx, fullNode, height)
	if err != nil {
		return 0, err
	}
	for _, validator := range validators {
		if strings.EqualFold(validator.Address, want) {
			return validator.Power, nil
		}
	}
	return 0, nil
}
