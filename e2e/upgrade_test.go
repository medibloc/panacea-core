package e2e_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	upgradeName                      = "v2.3.0"
	upgradeBinaryVersion             = "2.3.0"
	upgradeLegacyPNFTUnsignedPath    = "upgrade/legacy-pnft-unsigned.json"
	upgradeLegacyPNFTSignedPath      = "upgrade/legacy-pnft-signed.json"
	legacyPNFTDisabledMessage        = "legacy PNFT messages are disabled"
	upgradeV221Commit                = "a1b342939ba6ac3092aeebbee6a2fa741a34d47f"
	upgradeP0InvariantValidatorIndex = 1
)

type upgradePreservedState struct {
	Address           string
	Bank              upgradeBankAccountState
	PreBankTxHash     string
	PreBankRecipient  string
	DID               upgradeDIDFixtures
	AOLFixture        harness.AOLUpgradeFixture
	AOLPreparation    harness.AOLUpgradePreparationEvidence
	AOLPreCheckpoint  harness.AOLUpgradeCheckpoint
	IBCClientParams   json.RawMessage
	IBCConnectParams  json.RawMessage
	IBCTransferParams json.RawMessage
}

type upgradeBinaryIdentity struct {
	Name             string `json:"name"`
	Version          string `json:"version"`
	Commit           string `json:"commit"`
	CosmosSDKVersion string `json:"cosmos_sdk_version"`
}

type upgradeRunScenario struct {
	Name                         string
	LegacyPNFTAdversarialFixture bool
	RunPostUpgradeStateSync      bool
	RunP0BoundaryMatrix          bool
}

type legacyPNFTUpgradeRunState struct {
	Creator   ibc.Wallet
	Prepared  preparedLegacyPNFTFixture
	Isolation *legacyPNFTIsolationEvidence
}

func TestDecodeUpgradeModuleVersions(t *testing.T) {
	t.Parallel()

	versions, err := decodeUpgradeModuleVersions([]byte(`{
		"module_versions": [
			{"name":"nft","version":"1"},
			{"name":"pnft","version":1},
			{"name":"ibc","version":"6"}
		]
	}`))
	require.NoError(t, err)
	require.Equal(t, uint64(1), versions["nft"])
	require.Equal(t, uint64(1), versions["pnft"])
	require.Equal(t, uint64(6), versions["ibc"])

	_, err = decodeUpgradeModuleVersions([]byte(`{"module_versions":[{"name":"nft","version":"invalid"}]}`))
	require.ErrorContains(t, err, "nft")
	_, err = decodeUpgradeModuleVersions([]byte(`{"module_versions":[{"name":"nft","version":"1"},{"name":"nft","version":"1"}]}`))
	require.ErrorContains(t, err, "duplicate")
}

func decodeUpgradeModuleVersions(raw []byte) (map[string]uint64, error) {
	var response struct {
		ModuleVersions []struct {
			Name    string          `json:"name"`
			Version json.RawMessage `json:"version"`
		} `json:"module_versions"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, fmt.Errorf("decode upgrade module versions: %w", err)
	}
	if len(response.ModuleVersions) == 0 {
		return nil, errors.New("upgrade module versions response is empty")
	}
	versions := make(map[string]uint64, len(response.ModuleVersions))
	for _, module := range response.ModuleVersions {
		name := strings.TrimSpace(module.Name)
		if name == "" {
			return nil, errors.New("upgrade module version has no name")
		}
		if _, exists := versions[name]; exists {
			return nil, fmt.Errorf("duplicate upgrade module version %q", name)
		}
		versionText := strings.TrimSpace(string(module.Version))
		if len(versionText) >= 2 && versionText[0] == '"' && versionText[len(versionText)-1] == '"' {
			if err := json.Unmarshal(module.Version, &versionText); err != nil {
				return nil, fmt.Errorf("decode upgrade module version %q: %w", name, err)
			}
		}
		version, err := strconv.ParseUint(versionText, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("decode upgrade module version %q value %q: %w", name, versionText, err)
		}
		versions[name] = version
	}
	return versions, nil
}

func TestV221ToCurrentMultiValidatorUpgrade(t *testing.T) {
	runV221ToCurrentMultiValidatorUpgrade(t, upgradeRunScenario{
		Name:                    "normal-empty-legacy-pnft",
		RunPostUpgradeStateSync: true,
		RunP0BoundaryMatrix:     true,
	})
}

func TestV221ToCurrentLegacyPNFTAdversarialUpgrade(t *testing.T) {
	runV221ToCurrentMultiValidatorUpgrade(t, upgradeRunScenario{
		Name:                         "adversarial-non-empty-legacy-pnft",
		LegacyPNFTAdversarialFixture: true,
	})
}

func runV221ToCurrentMultiValidatorUpgrade(t *testing.T, scenario upgradeRunScenario) {
	t.Helper()
	if os.Getenv("PANACEA_E2E_UPGRADE") != "1" {
		t.Skip("use ./scripts/e2e/run.sh upgrade to provide the required binary identity")
	}
	require.NotEmpty(t, scenario.Name)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	networkConfig := harness.Config{
		Image:              harness.V221Image(),
		NumValidators:      4,
		NumFullNodes:       1,
		TimeoutCommit:      "1s",
		SnapshotInterval:   5,
		SnapshotKeepRecent: 3,
	}
	if scenario.RunP0BoundaryMatrix {
		networkConfig.StakingUnbondingTime = "600s"
		networkConfig.SlashingSignedBlocksWindow = 100
		networkConfig.SlashingMinSignedPerWindow = "0.800000000000000000"
		networkConfig.SlashingDowntimeJailDuration = "300s"
		networkConfig.SlashingSlashFractionDowntime = "0.010000000000000000"
	}
	network, err := harness.Start(ctx, t, networkConfig)
	require.NoError(t, err)
	defer network.RecordTestPanic()
	require.Len(t, network.Chain.Validators, 4)
	require.Len(t, network.Chain.FullNodes, 1)
	toolchainEvidence, err := network.CaptureAndRecordGoToolchainEvidence(ctx)
	require.NoError(t, err)
	require.Equal(t, "go1.23.12", toolchainEvidence.GOVersion)
	require.Contains(t, toolchainEvidence.CompilerVersion, "go1.23.12")
	require.NoError(t, network.WriteArtifactJSON("upgrade/scenario.json", map[string]any{
		"name":                            scenario.Name,
		"legacy_pnft_adversarial_fixture": scenario.LegacyPNFTAdversarialFixture,
		"run_post_upgrade_state_sync":     scenario.RunPostUpgradeStateSync,
		"run_p0_boundary_matrix":          scenario.RunP0BoundaryMatrix,
	}))

	startHeight, err := network.Chain.Height(ctx)
	require.NoError(t, err)
	preUpgradeHeight := startHeight + 3
	require.NoError(t, network.WaitForHeight(ctx, preUpgradeHeight))
	require.NoError(t, network.WaitForFullNode(ctx, preUpgradeHeight))
	validatorSet, err := network.ValidatorSet(ctx, network.Chain.FullNodes[0], preUpgradeHeight)
	require.NoError(t, err)
	require.Len(t, validatorSet, 4)
	require.Positive(t, validatorSet[0].Power)
	for _, validator := range validatorSet[1:] {
		require.Equal(t, validatorSet[0].Power, validator.Power, "upgrade validators must have equal voting power")
	}

	for _, node := range network.Chain.Nodes() {
		identity, err := captureUpgradeNodeVersion(ctx, network, node, "old")
		require.NoError(t, err)
		require.Equal(t, "panacea-core", identity.Name)
		require.Equal(t, "2.2.1", identity.Version)
		require.Equal(t, upgradeV221Commit, identity.Commit)
		require.Equal(t, "v0.47.10", identity.CosmosSDKVersion)
	}
	_, err = captureExactUpgradeModuleVersions(
		ctx,
		network,
		"pre-upgrade",
		"upgrade/module-versions-pre-upgrade.json",
		"2.2.1",
		upgradeV221ExpectedModuleVersions,
	)
	require.NoError(t, err)
	var p0GenesisBefore upgradeP0GenesisContractEvidence
	if scenario.RunP0BoundaryMatrix {
		p0GenesisBefore, err = captureUpgradeP0GenesisContract(ctx, network, "pre-upgrade", networkConfig)
		require.NoError(t, err)
	}
	preserved := prepareUpgradeState(t, ctx, network)
	preUpgradeCheckpoint, err := captureUpgradeStateCheckpoint(
		ctx,
		network,
		"pre-upgrade",
		preserved.Address,
		[]string{preserved.PreBankTxHash},
	)
	require.NoError(t, err)
	requireUpgradeBankAccountEqual(t, preserved.Bank, preUpgradeCheckpoint.Bank)
	require.NoError(t, network.WriteArtifactJSON("upgrade/pre-state.json", map[string]any{
		"address":               preserved.Address,
		"bank":                  preserved.Bank,
		"pre_bank_tx_hash":      preserved.PreBankTxHash,
		"pre_bank_recipient":    preserved.PreBankRecipient,
		"did":                   preserved.DID,
		"aol_fixture":           preserved.AOLFixture,
		"aol_preparation":       preserved.AOLPreparation,
		"aol_pre_checkpoint":    preserved.AOLPreCheckpoint,
		"ibc_client_params":     json.RawMessage(preserved.IBCClientParams),
		"ibc_connection_params": json.RawMessage(preserved.IBCConnectParams),
		"ibc_transfer_params":   json.RawMessage(preserved.IBCTransferParams),
	}))
	proposer, err := buildAndFundUpgradeRawTxSigner(ctx, network, "upgrade-proposer")
	require.NoError(t, err)
	compatibleSignedBankTx, err := prepareV221CompatibleSignedBankTx(ctx, network)
	require.NoError(t, err)
	var legacyAminoCustomTxs upgradeV221LegacyAminoCustomTxsFixture
	var haltMempoolTxs upgradeHaltMempoolFixture
	if scenario.RunP0BoundaryMatrix {
		legacyAminoCustomTxs, err = prepareV221LegacyAminoCustomTxs(ctx, network)
		require.NoError(t, err)
		haltMempoolTxs, err = prepareV221UpgradeHaltMempoolTxs(ctx, network)
		require.NoError(t, err)
	}
	var legacyPNFTRun *legacyPNFTUpgradeRunState
	if scenario.LegacyPNFTAdversarialFixture {
		legacyCreator := buildAndFundNFTWallet(t, ctx, network, "upgrade-legacy-pnft-adversarial-creator")
		legacyRecipient, buildErr := network.BuildWallet(ctx, "upgrade-legacy-pnft-adversarial-recipient", "")
		require.NoError(t, buildErr)
		legacyPrepared, prepareErr := prepareV221LegacyPNFTFixture(
			ctx,
			network,
			legacyCreator,
			legacyRecipient.FormattedAddress(),
		)
		require.NoError(t, prepareErr)
		legacyPNFTRun = &legacyPNFTUpgradeRunState{
			Creator:  legacyCreator,
			Prepared: legacyPrepared,
		}
	} else {
		require.NoError(t, assertV221LegacyPNFTStoreEmpty(ctx, network))
	}
	stakingPreparation, err := prepareUpgradeStakingMatrix(ctx, network, 0)
	require.NoError(t, err)
	var p0QueueFixture upgradeP0StakingQueueFixture
	var p0QueueEvidence upgradeP0StakingQueueEvidence
	var p0SlashingFixture upgradeP0SlashingFixture
	var p0SlashingEvidence upgradeP0SlashingEvidence
	if scenario.RunP0BoundaryMatrix {
		var p0QueueBefore upgradeP0StakingQueueCheckpoint
		p0QueueFixture, p0QueueBefore, err = prepareUpgradeP0StakingQueue(ctx, network)
		require.NoError(t, err)
		p0QueueEvidence.Before = p0QueueBefore
		p0SlashingFixture, p0SlashingEvidence, err = prepareUpgradeP0Slashing(ctx, network)
		require.NoError(t, err)
	}
	authzFeegrantPreparation := prepareUpgradeAuthzFeegrant(t, ctx, network)
	groupVestingPreparation, err := prepareUpgradeGroupVestingMatrix(ctx, network)
	require.NoError(t, err)
	systemPreparation, err := prepareUpgradeSystemModules(ctx, network)
	require.NoError(t, err)

	proposalBaseHeight, err := network.Chain.Height(ctx)
	require.NoError(t, err)
	// Voting alone consumes roughly twenty one-second blocks. The deep matrix
	// then takes height-pinned checkpoints and deterministic exports on the old
	// binary, so retain enough headroom for slower CI Docker hosts.
	upgradeHeightOffset := int64(90)
	if scenario.RunP0BoundaryMatrix {
		upgradeHeightOffset = 120
	}
	upgradeHeight := proposalBaseHeight + upgradeHeightOffset
	proposalTx, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-submit-proposal",
		network.Chain.Validators[0],
		proposer.KeyName(),
		"gov", "submit-legacy-proposal", "software-upgrade", upgradeName,
		"--title", "Panacea v2.3.0 E2E upgrade",
		"--description", "Interchaintest multi-validator state migration",
		"--deposit", "1umed",
		"--upgrade-height", strconv.FormatInt(upgradeHeight, 10),
		"--upgrade-info", "{}",
		"--no-validate",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	// Generate the historical message after the proposal has consumed the
	// proposer's sequence; the signed transaction must remain CheckTx-valid
	// until it is broadcast by the upgraded binary.
	prepareLegacyPNFTUnsignedTransaction(t, ctx, network, proposer.KeyName())
	proposalID, err := proposalIDFromCommittedTx(proposalTx)
	require.NoError(t, err)
	require.NoError(t, network.AppendArtifactJSON("upgrade/timeline.jsonl", map[string]any{
		"event":          "proposal-submitted",
		"recorded_at":    time.Now().UTC(),
		"proposal_id":    proposalID,
		"upgrade_height": upgradeHeight,
		"tx_hash":        proposalTx.TxHash,
	}))

	for index, validator := range network.Chain.Validators {
		voteTx, voteErr := network.BroadcastAndWaitTx(
			ctx,
			"upgrade-vote-validator-"+strconv.Itoa(index),
			validator,
			"validator",
			"gov", "vote", strconv.FormatUint(proposalID, 10), "yes",
			"--gas", "500000",
			"--broadcast-mode", "sync",
		)
		require.NoError(t, voteErr)
		require.NoError(t, network.AppendArtifactJSON("upgrade/timeline.jsonl", map[string]any{
			"event":       "validator-voted",
			"recorded_at": time.Now().UTC(),
			"validator":   validator.Name(),
			"tx_hash":     voteTx.TxHash,
		}))
	}
	require.NoError(t, waitForProposalPassed(ctx, network, proposalID))
	if scenario.RunP0BoundaryMatrix {
		queueStartHeight := upgradeHeight - 45
		require.NoError(t, network.WaitForHeight(ctx, queueStartHeight))
		require.NoError(t, network.WaitForFullNode(ctx, queueStartHeight))
		observedQueueStartHeight, heightErr := network.Chain.Height(ctx)
		require.NoError(t, heightErr)
		require.Less(t, observedQueueStartHeight, upgradeHeight-10,
			"P0 staking queues require at least ten committing blocks before the upgrade halt")
		p0QueueEvidence, err = beginUpgradeP0StakingQueues(
			ctx,
			network,
			p0QueueFixture,
			p0QueueEvidence.Before,
		)
		require.NoError(t, err)
		require.NoError(t, network.AppendArtifactJSON("upgrade/timeline.jsonl", map[string]any{
			"event":           "p0-staking-time-queues-started",
			"recorded_at":     time.Now().UTC(),
			"observed_height": observedQueueStartHeight,
			"upgrade_height":  upgradeHeight,
		}))
	}
	preUpgradeGov, err := captureUpgradeGovCheckpoint(ctx, network, "pre-upgrade", proposalID)
	require.NoError(t, err)
	preUpgradeStaking, err := captureUpgradeStakingCheckpoint(
		ctx,
		network,
		stakingPreparation.Fixture,
		"pre-upgrade-checkpoint",
		stakingPreparation.TxHashes(),
	)
	require.NoError(t, err)
	preUpgradeAuthzFeegrant, err := captureUpgradeAuthzFeegrantCheckpoint(
		ctx,
		network,
		authzFeegrantPreparation.Fixture,
		"pre-upgrade-checkpoint",
		false,
	)
	require.NoError(t, err)
	preUpgradeGroupVesting, err := captureUpgradeGroupVestingCheckpoint(
		ctx,
		network,
		groupVestingPreparation.Fixture,
		"pre-upgrade-checkpoint",
		groupVestingPreparation.TxHashes(),
	)
	require.NoError(t, err)
	preUpgradeSystemModules, err := captureUpgradeSystemModuleCheckpoint(
		ctx,
		network,
		"pre-upgrade-checkpoint",
	)
	require.NoError(t, err)
	require.NoError(t, captureV221CompatibleSignedBankTxPreUpgrade(ctx, network, &compatibleSignedBankTx))
	if scenario.RunP0BoundaryMatrix {
		require.NoError(t, captureV221LegacyAminoCustomTxsPreUpgrade(ctx, network, &legacyAminoCustomTxs))
		require.NoError(t, captureV221UpgradeHaltMempoolTxsPreUpgrade(ctx, network, &haltMempoolTxs))
	}

	preHaltBlock := captureOldPreHaltBlock(t, ctx, network, upgradeHeight-1)
	require.NoError(t, network.WriteArtifactJSON("upgrade/pre-halt-block.json", map[string]any{
		"upgrade_height":                       upgradeHeight,
		"previous_consensus_block_height":      upgradeHeight - 1,
		"previous_block_app_hash_state_height": upgradeHeight - 2,
		"nodes":                                preHaltBlock,
	}))
	haltEvidence, haltErr := waitForOldBinaryUpgradeHalt(ctx, network, upgradeHeight)
	require.NoError(t, network.WriteArtifactJSON("upgrade/old-binary-halt.json", haltEvidence))
	require.NoError(t, haltErr)
	fullNode := network.Chain.FullNodes[0]
	if scenario.RunP0BoundaryMatrix {
		require.NoError(t, stopUpgradeP0SlashingTargetAtHalt(
			ctx,
			network,
			upgradeHeight,
			&p0SlashingEvidence,
		))
		require.NoError(t, submitV221UpgradeHaltMempoolTxsAtHalt(
			ctx,
			network,
			fullNode,
			&haltMempoolTxs,
		))
	}

	currentImage := harness.CurrentImage()
	quorumValidators := []*cosmos.ChainNode{
		network.Chain.Validators[0],
		network.Chain.Validators[1],
		network.Chain.Validators[2],
	}
	quorumNodes := append([]*cosmos.ChainNode{}, quorumValidators...)
	if !scenario.RunP0BoundaryMatrix {
		quorumNodes = append(quorumNodes, fullNode)
	}
	quorumSwitchCtx, quorumSwitchCancel := context.WithTimeout(ctx, 3*time.Minute)
	quorumSwitches, err := network.SwitchNodeImagesTogether(
		quorumSwitchCtx,
		"quorum-batch",
		quorumNodes,
		currentImage,
	)
	quorumSwitchCancel()
	require.NoError(t, err)
	require.Len(t, quorumSwitches, len(quorumNodes))
	for index, node := range quorumNodes {
		identity, identityErr := captureUpgradeNodeVersion(ctx, network, node, "quorum-current")
		require.NoError(t, identityErr)
		assertCurrentUpgradeBinaryIdentity(t, identity)
		require.NoError(t, network.AppendArtifactJSON("upgrade/timeline.jsonl", map[string]any{
			"event":  "quorum-node-switched",
			"node":   node.Name(),
			"switch": quorumSwitches[index],
		}))
	}

	postUpgradeTarget := upgradeHeight + 3
	postCtx, postCancel := context.WithTimeout(ctx, 90*time.Second)
	require.NoError(t, network.WaitForNodeHeight(postCtx, network.Chain.Validators[0], postUpgradeTarget))
	postCancel()
	postUpgradeHistoryNodes := append([]*cosmos.ChainNode{}, quorumNodes...)
	if scenario.RunP0BoundaryMatrix {
		carrierCommitCtx, carrierCommitCancel := context.WithTimeout(ctx, 2*time.Minute)
		require.NoError(t, assertV221UpgradeHaltMempoolTxsCommittedOnNode(
			carrierCommitCtx,
			network,
			network.Chain.Validators[0],
			&haltMempoolTxs,
		))
		carrierCommitCancel()

		fullNodeSwitch, switchErr := network.SwitchNodeImage(ctx, "mempool-carrier-full-node", fullNode, currentImage)
		require.NoError(t, switchErr)
		fullNodeIdentity, identityErr := captureUpgradeNodeVersion(ctx, network, fullNode, "full-node-current")
		require.NoError(t, identityErr)
		assertCurrentUpgradeBinaryIdentity(t, fullNodeIdentity)
		fullNodeCtx, fullNodeCancel := context.WithTimeout(ctx, 90*time.Second)
		require.NoError(t, network.WaitForFullNode(fullNodeCtx, postUpgradeTarget))
		fullNodeCancel()
		require.NoError(t, network.AppendArtifactJSON("upgrade/timeline.jsonl", map[string]any{
			"event":  "mempool-carrier-full-node-switched",
			"node":   fullNode.Name(),
			"switch": fullNodeSwitch,
		}))
		postUpgradeHistoryNodes = append(postUpgradeHistoryNodes, fullNode)
		p0GenesisAfter, paramsErr := captureUpgradeP0GenesisContract(ctx, network, "post-upgrade", networkConfig)
		require.NoError(t, paramsErr)
		require.Equal(t, p0GenesisBefore.Params, p0GenesisAfter.Params)
		require.NoError(t, network.WriteArtifactJSON("upgrade/p0/genesis-contract.json", map[string]any{
			"configured_overrides": networkConfig,
			"pre_upgrade":          p0GenesisBefore,
			"post_upgrade":         p0GenesisAfter,
		}))
		require.NoError(t, captureUpgradeP0StakingQueuesPendingAfterUpgrade(
			ctx,
			network,
			p0QueueFixture,
			&p0QueueEvidence,
		))
	} else {
		postCtx, postCancel = context.WithTimeout(ctx, 90*time.Second)
		require.NoError(t, network.WaitForFullNode(postCtx, postUpgradeTarget))
		postCancel()
	}
	firstObservedHeight, err := network.Chain.Validators[0].Height(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, firstObservedHeight, postUpgradeTarget)
	firstPostBlockEvidence, err := network.RequireSameHistoryAtHeight(ctx, upgradeHeight, postUpgradeHistoryNodes...)
	require.NoError(t, err)
	migratedStateEvidence, err := network.RequireSameHistoryAtHeight(ctx, upgradeHeight+1, postUpgradeHistoryNodes...)
	require.NoError(t, err)
	require.NoError(t, network.WriteArtifactJSON("upgrade/first-post-upgrade-block.json", map[string]any{
		"first_post_upgrade_block_height": upgradeHeight,
		"first_post_upgrade_block":        firstPostBlockEvidence,
		"migrated_state_height":           upgradeHeight,
		"app_hash_carrier_block_height":   upgradeHeight + 1,
		"migrated_state_app_hash":         migratedStateEvidence,
	}))
	require.NoError(t, network.AppendArtifactJSON("upgrade/timeline.jsonl", map[string]any{
		"event":                 "quorum-resumed",
		"recorded_at":           time.Now().UTC(),
		"first_observed_height": firstObservedHeight,
		"first_post_height":     upgradeHeight,
		"upgrade_height":        upgradeHeight,
	}))

	delayed := network.Chain.Validators[3]
	if scenario.RunP0BoundaryMatrix {
		require.NoError(t, waitForUpgradeP0SlashingJail(
			ctx,
			network,
			p0SlashingFixture,
			&p0SlashingEvidence,
		))
	}
	delayedSwitch, err := network.SwitchNodeImage(ctx, "delayed-validator", delayed, currentImage)
	require.NoError(t, err)
	delayedIdentity, err := captureUpgradeNodeVersion(ctx, network, delayed, "delayed-current")
	require.NoError(t, err)
	assertCurrentUpgradeBinaryIdentity(t, delayedIdentity)
	delayedCtx, delayedCancel := context.WithTimeout(ctx, 90*time.Second)
	require.NoError(t, network.WaitForNodeHeight(delayedCtx, delayed, postUpgradeTarget))
	delayedCancel()
	require.NoError(t, network.AppendArtifactJSON("upgrade/timeline.jsonl", map[string]any{
		"event":       "delayed-validator-caught-up",
		"recorded_at": time.Now().UTC(),
		"node":        delayed.Name(),
		"height":      postUpgradeTarget,
		"switch":      delayedSwitch,
	}))
	if scenario.RunP0BoundaryMatrix {
		require.NoError(t, exerciseUpgradeP0Unjail(
			ctx,
			network,
			p0SlashingFixture,
			&p0SlashingEvidence,
		))
		require.NoError(t, completeUpgradeP0StakingQueues(
			ctx,
			network,
			p0QueueFixture,
			&p0QueueEvidence,
		))
	}

	allNodes := append([]*cosmos.ChainNode{}, network.Chain.Validators...)
	allNodes = append(allNodes, network.Chain.FullNodes[0])
	for _, node := range allNodes {
		require.NoError(t, network.WaitForNodeHeight(ctx, node, postUpgradeTarget))
	}
	_, err = network.RequireSameHistoryAtHeight(ctx, upgradeHeight, allNodes...)
	require.NoError(t, err)
	_, err = network.RequireSameHistoryAtHeight(ctx, postUpgradeTarget, allNodes...)
	require.NoError(t, err)

	appliedRaw, err := network.FullNodeCLIQuery(ctx, "upgrade-applied-plan", "upgrade", "applied", upgradeName)
	require.NoError(t, err)
	var applied struct {
		Height string `json:"height"`
	}
	require.NoError(t, json.Unmarshal(appliedRaw, &applied))
	require.Equal(t, strconv.FormatInt(upgradeHeight, 10), applied.Height)

	moduleVersions, err := captureExactUpgradeModuleVersions(
		ctx,
		network,
		"post-upgrade",
		"upgrade/module-versions.json",
		upgradeBinaryVersion,
		upgradeCurrentExpectedModuleVersions,
	)
	require.NoError(t, err)
	emptyCurrentNFTClassID := proposer.FormattedAddress() + ":" + legacyPNFTLocalClassID
	if legacyPNFTRun != nil {
		emptyCurrentNFTClassID = legacyPNFTRun.Prepared.Fixture.DenomID
	}
	require.NoError(t, assertCurrentNFTStoreEmpty(
		ctx,
		network,
		"post-upgrade-preservation",
		emptyCurrentNFTClassID,
	))
	postUpgradeGov, err := captureUpgradeGovCheckpoint(ctx, network, "post-upgrade", proposalID)
	require.NoError(t, err)
	assertUpgradeGovMigration(t, preUpgradeGov, postUpgradeGov)
	var postUpgradeStaking upgradeStakingCheckpoint
	if scenario.RunP0BoundaryMatrix {
		postUpgradeStaking, err = captureUpgradeStakingCheckpoint(
			ctx,
			network,
			stakingPreparation.Fixture,
			"post-upgrade-preservation",
			preUpgradeStaking.TxHashes,
		)
		require.NoError(t, err)
		require.NoError(t, validateUpgradeP0StakingPreservationWithSlashing(
			preUpgradeStaking,
			postUpgradeStaking,
			p0SlashingEvidence,
			p0QueueEvidence,
		))
		require.NoError(t, network.WriteArtifactJSON("upgrade/staking/p0-preservation-validation.json", map[string]any{
			"pre_upgrade":  preUpgradeStaking,
			"post_upgrade": postUpgradeStaking,
			"slashing":     p0SlashingEvidence,
			"time_queues":  p0QueueEvidence,
		}))
	} else {
		postUpgradeStaking, err = captureAndValidateUpgradeStakingPreservation(
			ctx,
			network,
			stakingPreparation.Fixture,
			preUpgradeStaking,
		)
		require.NoError(t, err)
	}
	_, err = captureAndValidateUpgradeAuthzFeegrantPreserved(
		ctx,
		network,
		authzFeegrantPreparation.Fixture,
		preUpgradeAuthzFeegrant,
	)
	require.NoError(t, err)
	postUpgradeGroupVesting, err := captureAndValidateUpgradeGroupVestingPreservation(
		ctx,
		network,
		groupVestingPreparation.Fixture,
		preUpgradeGroupVesting,
	)
	require.NoError(t, err)
	postUpgradeSystemModules, err := captureUpgradeSystemModuleCheckpoint(
		ctx,
		network,
		"post-upgrade-preservation",
	)
	require.NoError(t, err)
	if scenario.RunP0BoundaryMatrix {
		supplyAccounting, accountingErr := validateUpgradeSystemModulePreservationWithSlashing(
			preUpgradeSystemModules,
			postUpgradeSystemModules,
			p0SlashingEvidence,
		)
		require.NoError(t, accountingErr)
		require.NoError(t, network.WriteArtifactJSON(
			"upgrade/system-modules/p0-slashing-supply-accounting.json",
			supplyAccounting,
		))
	} else {
		require.NoError(t, validateUpgradeSystemModulePreservation(preUpgradeSystemModules, postUpgradeSystemModules))
	}
	postUpgradeInvariantValidator := 3
	if scenario.RunP0BoundaryMatrix {
		// Keep independent stop/start probes off both validator 3, which owns the
		// jail/unjail proof, and validator 2, which owns deterministic exports.
		postUpgradeInvariantValidator = upgradeP0InvariantValidatorIndex
	}
	postUpgradeInvariant, err := network.AssertValidatorInvariantsAtHeight(
		ctx,
		"post-upgrade",
		postUpgradeInvariantValidator,
		postUpgradeSystemModules.Height,
	)
	require.NoError(t, err)
	require.True(t, postUpgradeInvariant.AllInvariantsPassed)

	assertUpgradeStatePreserved(t, ctx, network, preserved, preserved.Bank)
	assertUpgradeDIDFixturesPreserved(t, ctx, network, preserved.DID)
	postUpgradeAOL, err := network.CaptureAndAssertAOLUpgradePreserved(
		ctx,
		preserved.AOLFixture,
		preserved.AOLPreCheckpoint,
	)
	require.NoError(t, err)
	postPreservationCheckpoint, err := captureUpgradeStateCheckpoint(
		ctx,
		network,
		"post-upgrade-preservation",
		preserved.Address,
		[]string{preserved.PreBankTxHash},
	)
	require.NoError(t, err)
	requireUpgradeBankAccountEqual(t, preserved.Bank, postPreservationCheckpoint.Bank)
	require.NoError(t, recordUpgradeHistoricalTx(ctx, network, "post-upgrade", preserved.PreBankTxHash))
	for _, didCreateTxHash := range []string{
		preserved.DID.Updated.CreateTxHash,
		preserved.DID.Deactivated.CreateTxHash,
	} {
		require.NoError(t, recordUpgradeHistoricalTx(ctx, network, "post-upgrade", didCreateTxHash))
	}
	for _, aolTx := range preserved.AOLPreparation.Transactions {
		require.NoError(t, recordUpgradeHistoricalTx(ctx, network, "post-upgrade", aolTx.TxHash))
	}
	for _, stakingTxHash := range stakingPreparation.TxHashes() {
		require.NoError(t, recordUpgradeHistoricalTx(ctx, network, "post-upgrade", stakingTxHash))
	}
	for _, authzFeegrantTxHash := range authzFeegrantPreparation.TxHashes() {
		require.NoError(t, recordUpgradeHistoricalTx(ctx, network, "post-upgrade", authzFeegrantTxHash))
	}
	for _, groupVestingTxHash := range groupVestingPreparation.TxHashes() {
		require.NoError(t, recordUpgradeHistoricalTx(ctx, network, "post-upgrade", groupVestingTxHash))
	}
	require.NoError(t, recordUpgradeHistoricalTx(ctx, network, "post-upgrade", systemPreparation.BurnTxHash))
	require.NoError(t, broadcastV221CompatibleSignedBankTxAfterUpgrade(ctx, network, &compatibleSignedBankTx))
	if scenario.RunP0BoundaryMatrix {
		require.NoError(t, broadcastV221LegacyAminoCustomTxsAfterUpgrade(
			ctx,
			network,
			&legacyAminoCustomTxs,
		))
	}
	if legacyPNFTRun != nil {
		require.NoError(t, assertOldBinarySignedLegacyPNFTDisabled(ctx, network, legacyPNFTRun.Prepared))
		isolation, isolationErr := createStandardNFTAtLegacyIDs(
			ctx,
			network,
			legacyPNFTRun.Creator,
			legacyPNFTRun.Prepared,
		)
		require.NoError(t, isolationErr)
		legacyPNFTRun.Isolation = &isolation
	}
	assertLegacyPNFTDisabled(t, ctx, network, proposer.KeyName())
	require.NoError(t, assertNewRawLegacyPNFTDisabled(ctx, network, proposer, preserved.PreBankRecipient))
	postBankTxHashes := runPostUpgradeBankTransactions(t, ctx, network, preservedAddressKey)
	allBankTxHashes := append([]string{preserved.PreBankTxHash}, postBankTxHashes...)
	postUpgradeCheckpoint, err := captureUpgradeStateCheckpoint(
		ctx,
		network,
		"post-upgrade",
		preserved.Address,
		allBankTxHashes,
	)
	require.NoError(t, err)
	require.Equal(t, preserved.Bank.AccountNumber, postUpgradeCheckpoint.Bank.AccountNumber)
	require.Equal(t, preserved.Bank.Sequence+uint64(len(postBankTxHashes)), postUpgradeCheckpoint.Bank.Sequence)
	postUpgradeProposal := submitPostUpgradeGovProposal(t, ctx, network, proposer.KeyName())
	postUpgradeProposalID := postUpgradeProposal.ProposalID
	didMutation := runPostUpgradeDIDMutations(t, ctx, network, preserved.DID)
	authzFeegrantMutation, err := mutateUpgradeAuthzFeegrant(
		ctx,
		network,
		authzFeegrantPreparation.Fixture,
	)
	require.NoError(t, err)
	groupVestingMutation, err := mutateUpgradeGroupVestingMatrix(
		ctx,
		network,
		groupVestingPreparation.Fixture,
		postUpgradeGroupVesting,
	)
	require.NoError(t, err)
	aolMutation, err := network.MutateAOLAfterUpgrade(
		ctx,
		network.Chain.Validators[0],
		preserved.AOLFixture,
		postUpgradeAOL,
	)
	require.NoError(t, err)
	stakingMutation, err := mutateUpgradeStakingMatrix(
		ctx,
		network,
		stakingPreparation.Fixture,
		postUpgradeStaking,
	)
	require.NoError(t, err)
	stakingLiveness, err := exerciseUpgradeValidatorLiveness(
		ctx,
		network,
		stakingPreparation.Fixture,
		stakingMutation,
		3,
	)
	require.NoError(t, err)
	systemMutation, err := mutateUpgradeSystemModules(ctx, network, systemPreparation)
	require.NoError(t, err)
	var expectedPreexistingNFTClassIDs []string
	if legacyPNFTRun != nil {
		require.NotNil(t, legacyPNFTRun.Isolation)
		expectedPreexistingNFTClassIDs = append(
			expectedPreexistingNFTClassIDs,
			legacyPNFTRun.Isolation.ClassID,
		)
	}
	nftBeforeRestart := runNFTLifecycle(t, ctx, network, expectedPreexistingNFTClassIDs...)
	require.NoError(t, network.WriteArtifactJSON("upgrade/nft-before-restart.json", nftBeforeRestart))

	for _, node := range allNodes {
		require.NoError(t, network.CaptureNodeContainerLog(
			ctx,
			node,
			"upgrade/current-logs-before-restart/"+node.Name()+".log",
		))
	}
	restartEvidence, err := restartUpgradeNetworkWithEvidence(ctx, network, "post-upgrade-all-node", 3)
	require.NoError(t, err)
	postRestartSnapshotSince := time.Now().UTC()
	assertUpgradeStatePreserved(t, ctx, network, preserved, postUpgradeCheckpoint.Bank)
	assertUpgradeDIDMutationsPersisted(t, ctx, network, preserved.DID, didMutation)
	_, err = network.CaptureAndAssertAOLAfterRestart(ctx, preserved.AOLFixture, aolMutation.Final)
	require.NoError(t, err)
	_, err = captureAndValidateUpgradeStakingPostRestart(
		ctx,
		network,
		stakingPreparation.Fixture,
		stakingLiveness,
	)
	require.NoError(t, err)
	_, err = captureAndValidateUpgradeAuthzFeegrantPostRestart(
		ctx,
		network,
		authzFeegrantPreparation.Fixture,
		authzFeegrantMutation,
	)
	require.NoError(t, err)
	groupVestingPostRestart, err := exerciseUpgradeGroupVestingPostRestart(
		ctx,
		network,
		groupVestingPreparation.Fixture,
		groupVestingMutation,
	)
	require.NoError(t, err)
	postRestartSystemModules, err := captureUpgradeSystemModuleCheckpoint(
		ctx,
		network,
		"post-restart",
	)
	require.NoError(t, err)
	require.NoError(t, validateUpgradeSystemModulePreservation(systemMutation.Checkpoint, postRestartSystemModules))
	postRestartInvariantValidator := 3
	if scenario.RunP0BoundaryMatrix {
		// Use a second healthy validator so neither the P0 jail target nor the
		// earlier export target accumulates seven misses in the short window.
		postRestartInvariantValidator = 1
	}
	postRestartInvariant, err := network.AssertValidatorInvariantsAtHeight(
		ctx,
		"post-restart",
		postRestartInvariantValidator,
		postRestartSystemModules.Height,
	)
	require.NoError(t, err)
	require.True(t, postRestartInvariant.AllInvariantsPassed)
	require.NoError(t, assertV221CompatibleSignedBankTxAfterRestart(ctx, network, compatibleSignedBankTx))
	if scenario.RunP0BoundaryMatrix {
		require.NoError(t, assertV221LegacyAminoCustomTxsAfterRestart(ctx, network, legacyAminoCustomTxs))
		require.NoError(t, assertV221UpgradeHaltMempoolTxsAfterRestart(ctx, network, haltMempoolTxs))
		require.NoError(t, captureUpgradeP0StakingQueuesAfterRestart(
			ctx,
			network,
			p0QueueFixture,
			&p0QueueEvidence,
		))
		require.NoError(t, captureUpgradeP0SlashingAfterRestart(
			ctx,
			network,
			p0SlashingFixture,
			&p0SlashingEvidence,
		))
	}
	if legacyPNFTRun != nil {
		require.NotNil(t, legacyPNFTRun.Isolation)
		require.NoError(t, assertStandardNFTAtLegacyIDsPersisted(ctx, network, *legacyPNFTRun.Isolation))
	}
	postRestartCheckpoint, err := captureUpgradeStateCheckpoint(
		ctx,
		network,
		"post-restart",
		preserved.Address,
		allBankTxHashes,
	)
	require.NoError(t, err)
	requireUpgradeBankAccountEqual(t, postUpgradeCheckpoint.Bank, postRestartCheckpoint.Bank)
	for _, txHash := range allBankTxHashes {
		require.NoError(t, recordUpgradeHistoricalTx(ctx, network, "post-restart", txHash))
	}
	for _, didTxHash := range []string{
		preserved.DID.Updated.CreateTxHash,
		preserved.DID.Deactivated.CreateTxHash,
		didMutation.UpdateTxHash,
		didMutation.DeactivateTxHash,
	} {
		require.NoError(t, recordUpgradeHistoricalTx(ctx, network, "post-restart", didTxHash))
	}
	for _, aolTx := range append(
		append([]harness.AOLUpgradeTransactionEvidence(nil), preserved.AOLPreparation.Transactions...),
		aolMutation.Transactions...,
	) {
		require.NoError(t, recordUpgradeHistoricalTx(ctx, network, "post-restart", aolTx.TxHash))
	}
	for _, stakingTxHash := range append(
		stakingPreparation.TxHashes(),
		stakingMutation.DelegateTxHash,
		stakingMutation.WithdrawRewardTxHash,
	) {
		require.NoError(t, recordUpgradeHistoricalTx(ctx, network, "post-restart", stakingTxHash))
	}
	for _, authzFeegrantTxHash := range append(
		authzFeegrantPreparation.TxHashes(),
		authzFeegrantMutation.TxHashes()...,
	) {
		require.NoError(t, recordUpgradeHistoricalTx(ctx, network, "post-restart", authzFeegrantTxHash))
	}
	groupVestingTxHashes := append([]string(nil), groupVestingPreparation.TxHashes()...)
	for _, transaction := range groupVestingMutation.Transactions {
		groupVestingTxHashes = append(groupVestingTxHashes, transaction.TxHash)
	}
	groupVestingTxHashes = append(groupVestingTxHashes, groupVestingPostRestart.SpendTxHash)
	for _, groupVestingTxHash := range groupVestingTxHashes {
		require.NoError(t, recordUpgradeHistoricalTx(ctx, network, "post-restart", groupVestingTxHash))
	}
	for _, systemTxHash := range []string{systemPreparation.BurnTxHash, systemMutation.BurnTxHash} {
		require.NoError(t, recordUpgradeHistoricalTx(ctx, network, "post-restart", systemTxHash))
	}
	if legacyPNFTRun != nil {
		for _, txHash := range []string{
			legacyPNFTRun.Prepared.CreateTxHash,
			legacyPNFTRun.Prepared.MintTxHash,
			legacyPNFTRun.Isolation.CreateTxHash,
			legacyPNFTRun.Isolation.MintTxHash,
		} {
			require.NoError(t, recordUpgradeHistoricalTx(ctx, network, "post-restart", txHash))
		}
	}
	postRestartGov, err := captureUpgradeGovCheckpoint(ctx, network, "post-restart", proposalID)
	require.NoError(t, err)
	require.Equal(t, postUpgradeGov.Params, postRestartGov.Params)
	require.Equal(t, postUpgradeGov.Tally, postRestartGov.Tally)
	require.NoError(t, capturePostRestartExpeditedGovProposal(ctx, network, &postUpgradeProposal))
	for _, govProposalID := range []uint64{proposalID, postUpgradeProposalID} {
		_, queryErr := network.FullNodeCLIQuery(
			ctx,
			"upgrade-post-restart-proposal-"+strconv.FormatUint(govProposalID, 10),
			"gov", "proposal", strconv.FormatUint(govProposalID, 10),
		)
		require.NoError(t, queryErr)
	}
	postRestartConsensusHeight, err := network.Chain.Height(ctx)
	require.NoError(t, err)
	require.Greater(t, postRestartConsensusHeight, int64(1))
	postRestartQueryHeight := postRestartConsensusHeight - 1
	nftAfterRestart := nftLifecycleEvidence{
		ClassID: nftBeforeRestart.ClassID,
		NFTID:   nftBeforeRestart.NFTID,
		Class: assertClassQueryBoundaryParityAtHeight(
			t,
			ctx,
			network,
			"upgrade-class-after-restart",
			postRestartQueryHeight,
			nftBeforeRestart.ClassID,
		),
		Tombstone: assertNFTQueryBoundaryParityAtHeight(
			t,
			ctx,
			network,
			"upgrade-tombstone-after-restart",
			postRestartQueryHeight,
			nftBeforeRestart.ClassID,
			nftBeforeRestart.NFTID,
		),
		Pagination: captureNFTRestartPaginationEvidence(
			t,
			ctx,
			network,
			"upgrade-pagination-after-restart",
			nftBeforeRestart.Pagination.Owner,
			nftBeforeRestart.Pagination.Cursor,
		),
	}
	require.Equal(t, nftBeforeRestart.Class, nftAfterRestart.Class)
	require.Equal(t, nftBeforeRestart.Tombstone, nftAfterRestart.Tombstone)
	require.Equal(t, nftBeforeRestart.Pagination.Owner, nftAfterRestart.Pagination.Owner)
	require.Equal(t, nftBeforeRestart.Pagination.Cursor, nftAfterRestart.Pagination.Cursor)
	require.Equal(t, nftBeforeRestart.Pagination.FirstPage, nftAfterRestart.Pagination.FirstPage)
	require.Equal(t, nftBeforeRestart.Pagination.CursorPage, nftAfterRestart.Pagination.CursorPage)
	require.NoError(t, network.WriteArtifactJSON("upgrade/nft-after-restart.json", nftAfterRestart))
	for _, node := range allNodes {
		require.NoError(t, network.WaitForNodeHeight(ctx, node, restartEvidence.TargetHeight))
	}
	_, err = network.RequireSameHistoryAtHeight(ctx, restartEvidence.TargetHeight, allNodes...)
	require.NoError(t, err)
	postRestartModuleVersions, err := captureExactUpgradeModuleVersions(
		ctx,
		network,
		"post-restart",
		"upgrade/module-versions-post-restart.json",
		upgradeBinaryVersion,
		upgradeCurrentExpectedModuleVersions,
	)
	require.NoError(t, err)
	require.Equal(t, moduleVersions, postRestartModuleVersions)
	if scenario.RunPostUpgradeStateSync {
		stateSyncEvidence, stateSyncErr := network.RunCometStateSync(ctx, harness.CometStateSyncRequest{
			Step:                  "v2.2.1-to-current-post-restart-state-sync",
			RPCSources:            network.Chain.Validators[:2],
			ExpectedImage:         currentImage,
			QueryCommand:          []string{"nft", "nft-record", nftAfterRestart.ClassID, nftAfterRestart.NFTID},
			ProviderSnapshotSince: postRestartSnapshotSince,
			ProviderWaitTimeout:   time.Minute,
			CompletionTimeout:     3 * time.Minute,
		})
		require.NoError(t, stateSyncErr)
		require.Equal(t, "actual-cometbft-state-sync", stateSyncEvidence.Mode)
		require.True(t, stateSyncEvidence.StateSyncLogs.RestoredSnapshot)
		require.True(t, stateSyncEvidence.RestartSkippedStateSync)
		require.True(t, stateSyncEvidence.GenesisBlockUnavailable)
		require.True(t, stateSyncEvidence.NodeStopped)
		for phase, query := range map[string]harness.CometStateSyncQueryEvidence{
			"current-before-restart":    stateSyncEvidence.Queries.CurrentBefore,
			"historical-before-restart": stateSyncEvidence.Queries.HistoricalBefore,
			"current-after-restart":     stateSyncEvidence.Queries.CurrentAfter,
			"historical-after-restart":  stateSyncEvidence.Queries.HistoricalAfter,
		} {
			var restored nftRecordQueryResponse
			require.NoError(t, json.Unmarshal(query.Response, &restored), phase)
			require.Equal(t, nftAfterRestart.Tombstone, restored, phase)
			require.NotNil(t, restored.NFTRecord.BurnTombstone, phase)
		}
	}
	coverageInput := connectedUpgradeCoverageInput{
		Scenario:              scenario,
		Preserved:             preserved,
		ProposalID:            proposalID,
		PostUpgradeProposalID: postUpgradeProposalID,
		Staking:               stakingPreparation.Fixture,
		AuthzFeegrant:         authzFeegrantPreparation.Fixture,
		GroupVesting:          groupVestingPreparation.Fixture,
		NFT:                   nftAfterRestart,
		LegacyPNFT:            legacyPNFTRun,
		CompatibleSignedBank:  compatibleSignedBankTx,
	}
	if scenario.RunP0BoundaryMatrix {
		coverageInput.HaltMempool = &haltMempoolTxs
		coverageInput.StakingTimeQueues = &p0QueueEvidence
		coverageInput.SlashingJail = &p0SlashingEvidence
		coverageInput.LegacyAminoCustom = &legacyAminoCustomTxs
	}
	require.NoError(t, recordConnectedUpgradeCoverage(network, coverageInput))
}

const preservedAddressKey = "upgrade-preserved"

func prepareUpgradeState(t *testing.T, ctx context.Context, network *harness.Network) upgradePreservedState {
	t.Helper()
	wallet := buildAndFundNFTWallet(t, ctx, network, preservedAddressKey)
	did := prepareUpgradeDIDFixtures(t, ctx, network)
	aolOwner := buildAndFundNFTWallet(t, ctx, network, "upgrade-aol-owner")
	aolWriter := buildAndFundNFTWallet(t, ctx, network, "upgrade-aol-writer")
	aolMutationWriter, err := network.BuildWallet(ctx, "upgrade-aol-mutation-writer", "")
	require.NoError(t, err)
	aolFixture := harness.AOLUpgradeFixture{
		OwnerAddress:              aolOwner.FormattedAddress(),
		OwnerKeyName:              aolOwner.KeyName(),
		WriterAddress:             aolWriter.FormattedAddress(),
		WriterKeyName:             aolWriter.KeyName(),
		MutationWriterAddress:     aolMutationWriter.FormattedAddress(),
		TopicName:                 "upgrade-aol-matrix",
		TopicDescription:          "v2.2.1 AOL state retained across v2.3.0 migration",
		WriterMoniker:             "upgrade-persistent-writer",
		WriterDescription:         "writer created before the v2.3.0 upgrade",
		MutationWriterMoniker:     "upgrade-transient-writer",
		MutationWriterDescription: "writer added and deleted after the v2.3.0 upgrade",
		InitialRecord: harness.AOLUpgradeRecordInput{
			Key:   "pre-upgrade-record",
			Value: "created-by-v2.2.1",
		},
		MutationRecord: harness.AOLUpgradeRecordInput{
			Key:   "post-upgrade-record",
			Value: "created-by-v2.3.0",
		},
	}
	aolPreparation, err := network.PrepareAOLUpgradeFixture(
		ctx,
		network.Chain.Validators[0],
		aolFixture,
	)
	require.NoError(t, err)
	aolPreCheckpoint, err := network.CaptureAOLUpgradePreUpgradeCheckpoint(ctx, aolFixture)
	require.NoError(t, err)
	ibcClientRaw, err := network.FullNodeCLIQuery(ctx, "upgrade-pre-ibc-client-params", "ibc", "client", "params")
	require.NoError(t, err)
	ibcConnectionRaw, err := network.FullNodeCLIQuery(ctx, "upgrade-pre-ibc-connection-params", "ibc", "connection", "params")
	require.NoError(t, err)
	ibcTransferRaw, err := network.FullNodeCLIQuery(ctx, "upgrade-pre-ibc-transfer-params", "ibc-transfer", "params")
	require.NoError(t, err)
	preBankRecipient, err := network.BuildWallet(ctx, "upgrade-pre-bank-recipient", "")
	require.NoError(t, err)
	preBankTx, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-pre-bank-send",
		network.Chain.Validators[0],
		wallet.KeyName(),
		"bank", "send", wallet.KeyName(), preBankRecipient.FormattedAddress(), "7umed",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	preBankRecipientBalance, err := network.QueryFullNodeBalance(ctx, preBankRecipient.FormattedAddress(), "umed")
	require.NoError(t, err)
	require.Equal(t, "7", preBankRecipientBalance.String())
	bank, err := queryUpgradeBankAccount(ctx, network, "upgrade-pre-bank-account", wallet.FormattedAddress())
	require.NoError(t, err)

	return upgradePreservedState{
		Address:           wallet.FormattedAddress(),
		Bank:              bank,
		PreBankTxHash:     preBankTx.TxHash,
		PreBankRecipient:  preBankRecipient.FormattedAddress(),
		DID:               did,
		AOLFixture:        aolFixture,
		AOLPreparation:    aolPreparation,
		AOLPreCheckpoint:  aolPreCheckpoint,
		IBCClientParams:   append(json.RawMessage(nil), ibcClientRaw...),
		IBCConnectParams:  append(json.RawMessage(nil), ibcConnectionRaw...),
		IBCTransferParams: append(json.RawMessage(nil), ibcTransferRaw...),
	}
}

func assertUpgradeStatePreserved(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	want upgradePreservedState,
	wantBank upgradeBankAccountState,
) {
	t.Helper()
	bank, err := queryUpgradeBankAccount(ctx, network, "upgrade-preserved-bank-account", want.Address)
	require.NoError(t, err)
	requireUpgradeBankAccountEqual(t, wantBank, bank)

	ibcClientRaw, err := network.FullNodeCLIQuery(ctx, "upgrade-post-ibc-client-params", "ibc", "client", "params")
	require.NoError(t, err)
	require.JSONEq(t, string(want.IBCClientParams), string(ibcClientRaw))
	ibcConnectionRaw, err := network.FullNodeCLIQuery(ctx, "upgrade-post-ibc-connection-params", "ibc", "connection", "params")
	require.NoError(t, err)
	require.JSONEq(t, string(want.IBCConnectParams), string(ibcConnectionRaw))
	ibcTransferRaw, err := network.FullNodeCLIQuery(ctx, "upgrade-post-ibc-transfer-params", "ibc-transfer", "params")
	require.NoError(t, err)
	require.JSONEq(t, string(want.IBCTransferParams), string(ibcTransferRaw))
}

func waitForProposalPassed(ctx context.Context, network *harness.Network, proposalID uint64) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		raw, err := network.FullNodeCLIQuery(ctx, "upgrade-proposal-status", "gov", "proposal", strconv.FormatUint(proposalID, 10))
		if err == nil {
			var response struct {
				Proposal struct {
					Status string `json:"status"`
				} `json:"proposal"`
				Status string `json:"status"`
			}
			if json.Unmarshal(raw, &response) == nil {
				proposalStatus := response.Proposal.Status
				if proposalStatus == "" {
					proposalStatus = response.Status
				}
				switch strings.ToUpper(proposalStatus) {
				case "PROPOSAL_STATUS_PASSED", "PASSED":
					return nil
				case "PROPOSAL_STATUS_REJECTED", "PROPOSAL_STATUS_FAILED", "REJECTED", "FAILED":
					return fmt.Errorf("upgrade proposal %d ended with %s", proposalID, proposalStatus)
				}
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for upgrade proposal %d: %w", proposalID, ctx.Err())
		case <-ticker.C:
		}
	}
}

func captureOldPreHaltBlock(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	lastCommittedHeight int64,
) []harness.BlockEvidence {
	t.Helper()
	type observation struct {
		index    int
		evidence harness.BlockEvidence
		err      error
	}
	nodes := network.Chain.Nodes()
	observations := make(chan observation, len(nodes))
	waitCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	for index, node := range nodes {
		go func(index int, node *cosmos.ChainNode) {
			if err := network.WaitForNodeHeight(waitCtx, node, lastCommittedHeight); err != nil {
				observations <- observation{index: index, err: err}
				return
			}
			evidence, err := network.NodeBlock(waitCtx, node, lastCommittedHeight)
			observations <- observation{index: index, evidence: evidence, err: err}
		}(index, node)
	}

	evidence := make([]harness.BlockEvidence, len(nodes))
	for range nodes {
		observed := <-observations
		require.NoError(t, observed.err)
		evidence[observed.index] = observed.evidence
	}
	require.NotEmpty(t, evidence)
	for _, observed := range evidence[1:] {
		require.Equal(t, evidence[0].BlockID, observed.BlockID, "old binaries diverged on the block before the upgrade halt")
		require.Equal(t, evidence[0].AppHash, observed.AppHash, "old binaries have different app hashes before the upgrade halt")
	}
	return evidence
}

func decodeUpgradeInfo(contents []byte) (string, int64, error) {
	var response struct {
		Name   string          `json:"name"`
		Height json.RawMessage `json:"height"`
	}
	if err := json.Unmarshal(contents, &response); err != nil {
		return "", 0, err
	}
	if strings.TrimSpace(response.Name) == "" {
		return "", 0, errors.New("upgrade info has no plan name")
	}
	heightText := strings.Trim(strings.TrimSpace(string(response.Height)), `"`)
	height, err := strconv.ParseInt(heightText, 10, 64)
	if err != nil || height <= 0 {
		return "", 0, fmt.Errorf("upgrade info has invalid height %q", heightText)
	}
	return response.Name, height, nil
}

func prepareLegacyPNFTUnsignedTransaction(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	keyName string,
) {
	t.Helper()
	node := network.Chain.Validators[0]
	stdout, stderr, err := node.Exec(ctx, node.TxCommand(
		keyName,
		"pnft", "create-denom",
		"--denom-id", "upgrade-disabled-probe",
		"--denom-symbol", "UPGRADE",
		"--denom-name", "Upgrade Disabled Probe",
		"--generate-only",
		"--gas", "500000",
	), node.Chain.Config().Env)
	require.NoError(t, err, strings.TrimSpace(string(stderr)))
	require.True(t, json.Valid(stdout), "old binary returned invalid unsigned legacy PNFT JSON: %s", stdout)
	require.NoError(t, node.WriteFile(ctx, stdout, upgradeLegacyPNFTUnsignedPath))
	require.NoError(t, network.WriteArtifact("upgrade/legacy-pnft-unsigned.json", stdout))
}

func assertLegacyPNFTDisabled(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	keyName string,
) {
	t.Helper()
	node := network.Chain.Validators[0]
	unsignedPath := path.Join(node.HomeDir(), upgradeLegacyPNFTUnsignedPath)
	stdout, stderr, err := node.Exec(ctx, node.NodeCommand(
		"tx", "sign", unsignedPath,
		"--from", keyName,
		"--keyring-backend", "test",
		"--chain-id", node.Chain.Config().ChainID,
		"--output", "json",
	), node.Chain.Config().Env)
	require.NoError(t, err, strings.TrimSpace(string(stderr)))
	require.True(t, json.Valid(stdout), "upgraded binary returned invalid signed legacy PNFT JSON: %s", stdout)
	require.NoError(t, node.WriteFile(ctx, stdout, upgradeLegacyPNFTSignedPath))
	require.NoError(t, network.WriteArtifact("upgrade/legacy-pnft-signed.json", stdout))

	result, err := network.BroadcastSignedTxFileAndWaitDeliverFailure(
		ctx,
		"upgrade-legacy-pnft-disabled",
		node,
		upgradeLegacyPNFTSignedPath,
		"sdk",
		18,
	)
	require.NoError(t, err)
	require.Contains(t, result.RawLog, legacyPNFTDisabledMessage)
}

func runPostUpgradeBankTransactions(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	keyName string,
) []string {
	t.Helper()
	recipient, err := network.BuildWallet(ctx, "upgrade-bank-recipient", "")
	require.NoError(t, err)
	sendTx, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-post-bank-send",
		network.Chain.Validators[0],
		keyName,
		"bank", "send", keyName, recipient.FormattedAddress(), "1umed",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	balance, err := network.QueryFullNodeBalance(ctx, recipient.FormattedAddress(), "umed")
	require.NoError(t, err)
	require.Equal(t, "1", balance.String())

	firstMultiRecipient, err := network.BuildWallet(ctx, "upgrade-bank-multi-recipient-first", "")
	require.NoError(t, err)
	secondMultiRecipient, err := network.BuildWallet(ctx, "upgrade-bank-multi-recipient-second", "")
	require.NoError(t, err)
	multiSendTx, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-post-bank-multi-send",
		network.Chain.Validators[0],
		keyName,
		"bank", "multi-send",
		keyName,
		firstMultiRecipient.FormattedAddress(),
		secondMultiRecipient.FormattedAddress(),
		"2umed",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	for _, multiRecipient := range []string{
		firstMultiRecipient.FormattedAddress(),
		secondMultiRecipient.FormattedAddress(),
	} {
		multiBalance, queryErr := network.QueryFullNodeBalance(ctx, multiRecipient, "umed")
		require.NoError(t, queryErr)
		require.Equal(t, "2", multiBalance.String())
	}
	return []string{sendTx.TxHash, multiSendTx.TxHash}
}

func captureUpgradeNodeVersion(
	ctx context.Context,
	network *harness.Network,
	node *cosmos.ChainNode,
	phase string,
) (upgradeBinaryIdentity, error) {
	stdout, stderr, err := node.ExecBin(ctx, "version", "--long")
	if err != nil {
		return upgradeBinaryIdentity{}, fmt.Errorf("read %s version for %s: %w: %s", phase, node.Name(), err, strings.TrimSpace(string(stderr)))
	}
	if err := network.WriteArtifact("upgrade/versions/"+phase+"-"+node.Name()+".txt", stdout); err != nil {
		return upgradeBinaryIdentity{}, err
	}
	identity, err := decodeUpgradeBinaryIdentity(stdout)
	if err != nil {
		return upgradeBinaryIdentity{}, fmt.Errorf("decode %s version for %s: %w", phase, node.Name(), err)
	}
	return identity, nil
}

func decodeUpgradeBinaryIdentity(contents []byte) (upgradeBinaryIdentity, error) {
	fields := make(map[string]string)
	for _, line := range strings.Split(string(contents), "\n") {
		key, value, found := strings.Cut(line, ":")
		if !found || strings.HasPrefix(strings.TrimSpace(key), "-") {
			continue
		}
		key = strings.TrimSpace(key)
		switch key {
		case "name", "version", "commit", "cosmos_sdk_version":
			fields[key] = strings.TrimSpace(value)
		}
	}
	identity := upgradeBinaryIdentity{
		Name:             fields["name"],
		Version:          fields["version"],
		Commit:           fields["commit"],
		CosmosSDKVersion: fields["cosmos_sdk_version"],
	}
	if identity.Name == "" || identity.Version == "" || identity.Commit == "" || identity.CosmosSDKVersion == "" {
		return upgradeBinaryIdentity{}, fmt.Errorf("incomplete binary identity: %+v", identity)
	}
	return identity, nil
}

func assertCurrentUpgradeBinaryIdentity(t *testing.T, identity upgradeBinaryIdentity) {
	t.Helper()
	require.NoError(t, validateExpectedCurrentUpgradeBinaryIdentity(
		identity,
		os.Getenv("PANACEA_E2E_CURRENT_BINARY_VERSION"),
		os.Getenv("PANACEA_E2E_CURRENT_COMMIT"),
	))
}

func validateExpectedCurrentUpgradeBinaryIdentity(
	identity upgradeBinaryIdentity,
	expectedVersion string,
	expectedCommit string,
) error {
	if err := validateCurrentUpgradeBinaryIdentity(identity); err != nil {
		return err
	}
	expectedVersion = strings.TrimSpace(expectedVersion)
	if expectedVersion == "" {
		return errors.New("PANACEA_E2E_CURRENT_BINARY_VERSION is required")
	}
	if identity.Version != expectedVersion {
		return fmt.Errorf("current binary version %q does not match expected version %q", identity.Version, expectedVersion)
	}
	expectedCommit = strings.TrimSpace(expectedCommit)
	if expectedCommit == "" {
		return errors.New("PANACEA_E2E_CURRENT_COMMIT is required")
	}
	if identity.Commit != expectedCommit {
		return fmt.Errorf("current binary commit %q does not match expected worktree commit %q", identity.Commit, expectedCommit)
	}
	return nil
}

func validateCurrentUpgradeBinaryIdentity(identity upgradeBinaryIdentity) error {
	if identity.Name != "panacea-core" {
		return fmt.Errorf("current binary name must be panacea-core: %q", identity.Name)
	}
	if identity.Version != upgradeBinaryVersion {
		return fmt.Errorf("current binary version must be %s: %q", upgradeBinaryVersion, identity.Version)
	}
	if identity.Commit == upgradeV221Commit {
		return fmt.Errorf("current image resolved to the old binary commit %s", identity.Commit)
	}
	if identity.CosmosSDKVersion != "v0.50.15" {
		return fmt.Errorf("current binary Cosmos SDK version must be v0.50.15: %q", identity.CosmosSDKVersion)
	}
	return nil
}
