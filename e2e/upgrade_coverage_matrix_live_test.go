package e2e_test

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

type connectedUpgradeCoverageInput struct {
	Scenario              upgradeRunScenario
	Preserved             upgradePreservedState
	ProposalID            uint64
	PostUpgradeProposalID uint64
	Staking               upgradeStakingFixture
	AuthzFeegrant         upgradeAuthzFeegrantFixture
	GroupVesting          upgradeGroupVestingFixture
	NFT                   nftLifecycleEvidence
	LegacyPNFT            *legacyPNFTUpgradeRunState
	CompatibleSignedBank  upgradeCompatibleSignedTxFixture
	HaltMempool           *upgradeHaltMempoolFixture
	StakingTimeQueues     *upgradeP0StakingQueueEvidence
	SlashingJail          *upgradeP0SlashingEvidence
	LegacyAminoCustom     *upgradeV221LegacyAminoCustomTxsFixture
}

func buildConnectedUpgradeCoverageMatrix(input connectedUpgradeCoverageInput) harness.UpgradeCoverageMatrix {
	passed := func(name harness.UpgradeCoveragePhaseName, paths ...string) harness.UpgradeCoveragePhase {
		return harness.UpgradeCoveragePhase{
			Name:          name,
			Status:        harness.UpgradeCoverageStatusPassed,
			ArtifactPaths: paths,
		}
	}
	passedRow := func(
		area harness.UpgradeCoverageArea,
		priority harness.UpgradeCoveragePriority,
		stateObjectIDs []string,
		lowerLevelTests []string,
		paths [5][]string,
	) harness.UpgradeCoverageRow {
		phases := []harness.UpgradeCoveragePhase{
			passed(harness.UpgradeCoveragePhaseV221Preparation, paths[0]...),
			passed(harness.UpgradeCoveragePhasePreUpgradeCheckpoint, paths[1]...),
			passed(harness.UpgradeCoveragePhasePostUpgradePreservation, paths[2]...),
			passed(harness.UpgradeCoveragePhasePostUpgradeMutation, paths[3]...),
			passed(harness.UpgradeCoveragePhasePostRestart, paths[4]...),
		}
		return harness.UpgradeCoverageRow{
			Area:            area,
			Priority:        priority,
			Status:          harness.UpgradeCoverageStatusPassed,
			StateObjectIDs:  stateObjectIDs,
			LowerLevelTests: lowerLevelTests,
			QueryCoverage:   connectedUpgradeQueryCoverage(area, input.LegacyPNFT != nil),
			Phases:          phases,
		}
	}

	authBankIDs := []string{
		"account:" + input.Preserved.Address,
		"old-signed-account:" + input.CompatibleSignedBank.SignerAddress,
		"old-signed-tx:" + input.CompatibleSignedBank.TxHash,
	}
	authBankPaths := [5][]string{
		{"upgrade/pre-state.json", "upgrade/auth-bank/pre-signed-compatible-preparation.json", "upgrade/auth-bank/pre-signed-compatible-signed.json", "upgrade/auth-bank/pre-signed-compatible-signature-validation.txt", "tx/committed-results.jsonl"},
		{"state-checkpoints/pre-upgrade.json", "upgrade/auth-bank/pre-signed-compatible-pre-upgrade.json"},
		{"state-checkpoints/post-upgrade-preservation.json", "upgrade/auth-bank/pre-signed-compatible-post-upgrade-preservation.json", "tx/historical-lookups.jsonl"},
		{"state-checkpoints/post-upgrade.json", "upgrade/auth-bank/pre-signed-compatible-post-upgrade-mutation.json", "tx/committed-results.jsonl"},
		{"state-checkpoints/post-restart.json", "upgrade/auth-bank/pre-signed-compatible-post-restart.json", "tx/historical-lookups.jsonl"},
	}
	stakingIDs := []string{
		"delegation:" + input.Staking.DelegatorAddress + "/" + input.Staking.ValidatorOperator,
		"validator:" + input.Staking.ValidatorOperator,
	}
	stakingPaths := [5][]string{
		{"upgrade/staking/preparation.json"},
		{"upgrade/staking/checkpoints/pre-upgrade-checkpoint.json"},
		{"upgrade/staking/checkpoints/post-upgrade-preservation.json"},
		{"upgrade/staking/mutation.json", "upgrade/staking/validator-liveness.json"},
		{"upgrade/staking/checkpoints/post-restart.json", "upgrade/staking/post-restart-validation.json"},
	}
	slashingIDs := []string{"signing-info:" + input.Staking.ValidatorConsensusAddr}
	slashingPaths := [5][]string{
		{"upgrade/staking/preparation.json"},
		{"upgrade/staking/checkpoints/pre-upgrade-checkpoint.json"},
		{"upgrade/staking/checkpoints/post-upgrade-preservation.json"},
		{"upgrade/staking/validator-liveness.json"},
		{"upgrade/staking/post-restart-validation.json"},
	}
	didIDs := []string{
		"did:" + input.Preserved.DID.Updated.DID,
		"did:" + input.Preserved.DID.Deactivated.DID,
	}
	didPaths := [5][]string{
		{"upgrade/did/pre-upgrade.json"},
		{"upgrade/did/pre-upgrade.json", "queries/results.jsonl"},
		{"queries/results.jsonl", "tx/historical-lookups.jsonl"},
		{"upgrade/did/post-upgrade.json"},
		{"upgrade/did/post-restart.json", "tx/historical-lookups.jsonl"},
	}
	aolIDs := []string{"aol-topic:" + input.Preserved.AOLFixture.TopicName}
	aolPaths := [5][]string{
		{harness.AOLUpgradePreparationArtifactPath},
		{harness.AOLUpgradePreCheckpointArtifactPath},
		{harness.AOLUpgradePostPreservationArtifactPath},
		{harness.AOLUpgradePostMutationArtifactPath},
		{harness.AOLUpgradePostRestartArtifactPath},
	}
	if input.HaltMempool != nil {
		if signer := strings.TrimSpace(input.HaltMempool.SignerAddress); signer != "" {
			authBankIDs = append(authBankIDs, "halt-mempool-signer:"+signer)
		}
		for _, transaction := range input.HaltMempool.Transactions {
			if transaction.CheckTx != nil && strings.TrimSpace(transaction.CheckTx.TxHash) != "" {
				authBankIDs = append(authBankIDs, "halt-mempool-tx:"+transaction.CheckTx.TxHash)
			}
		}
		authBankPaths[0] = append(authBankPaths[0],
			"upgrade/halt-mempool-carrier/preparation.json",
			"upgrade/halt-mempool-carrier/tx-0-signed.json",
			"upgrade/halt-mempool-carrier/tx-1-signed.json",
		)
		authBankPaths[1] = append(authBankPaths[1],
			"upgrade/halt-mempool-carrier/pre-upgrade.json",
			"upgrade/halt-mempool-carrier/halt-checktx.json",
		)
		authBankPaths[2] = append(authBankPaths[2], "upgrade/halt-mempool-carrier/halt-checktx.json")
		authBankPaths[3] = append(authBankPaths[3], "upgrade/halt-mempool-carrier/carrier-commit.json")
		authBankPaths[4] = append(authBankPaths[4], "upgrade/halt-mempool-carrier/post-restart.json")
	}
	if input.StakingTimeQueues != nil {
		stakingIDs = append(stakingIDs,
			"unbonding-time-queue:"+input.StakingTimeQueues.SourceValidator,
			"redelegation-time-queue:"+input.StakingTimeQueues.SourceValidator+"/"+input.StakingTimeQueues.DestinationValidator,
		)
		stakingPaths[0] = append(stakingPaths[0], "upgrade/staking-time-queue/preparation.json")
		stakingPaths[1] = append(stakingPaths[1], "upgrade/staking-time-queue/queued.json")
		stakingPaths[2] = append(stakingPaths[2], "upgrade/staking-time-queue/post-upgrade-pending.json")
		stakingPaths[3] = append(stakingPaths[3], "upgrade/staking-time-queue/completed.json")
		stakingPaths[4] = append(stakingPaths[4], "upgrade/staking-time-queue/post-restart.json")
	}
	if input.SlashingJail != nil {
		slashingIDs = append(slashingIDs,
			"downtime-validator:"+input.SlashingJail.Before.Validator.OperatorAddress,
			"downtime-signing-info:"+input.SlashingJail.Before.SigningInfo.Address,
		)
		slashingPaths[0] = append(slashingPaths[0], "upgrade/slashing-jail/preparation.json")
		slashingPaths[1] = append(slashingPaths[1], "upgrade/slashing-jail/stopped-at-halt.json")
		slashingPaths[2] = append(slashingPaths[2],
			"upgrade/slashing-jail/jailed.json",
			"upgrade/system-modules/p0-slashing-supply-accounting.json",
		)
		slashingPaths[3] = append(slashingPaths[3], "upgrade/slashing-jail/unjail-and-rejoin.json")
		slashingPaths[4] = append(slashingPaths[4], "upgrade/slashing-jail/post-restart.json")
	}
	if input.LegacyAminoCustom != nil {
		for _, item := range []upgradeV221LegacyAminoCustomTxFixture{
			input.LegacyAminoCustom.AOL,
			input.LegacyAminoCustom.DID,
		} {
			if signer := strings.TrimSpace(item.SignerAddress); signer != "" {
				authBankIDs = append(authBankIDs, "legacy-amino-account:"+signer)
			}
			if hash := strings.TrimSpace(item.TxHash); hash != "" {
				authBankIDs = append(authBankIDs, "legacy-amino-tx:"+hash)
			}
		}
		didIDs = append(didIDs, "legacy-amino-did:"+input.LegacyAminoCustom.DID.StateObjectID)
		aolIDs = append(aolIDs, "legacy-amino-aol:"+input.LegacyAminoCustom.AOL.StateObjectID)
		legacyPreparationPaths := []string{
			"upgrade/legacy-amino-custom/preparation.json",
			"upgrade/legacy-amino-custom/aol-create-topic/signed.json",
			"upgrade/legacy-amino-custom/did-update/signed.json",
		}
		legacyCheckpointPaths := []string{"upgrade/legacy-amino-custom/pre-upgrade.json"}
		legacyPreservationPaths := []string{"upgrade/legacy-amino-custom/post-upgrade.json"}
		legacyMutationPaths := []string{
			"upgrade/legacy-amino-custom/post-upgrade.json",
			"upgrade/legacy-amino-custom/aol-create-topic/tampered.json",
			"upgrade/legacy-amino-custom/did-update/tampered.json",
		}
		legacyRestartPaths := []string{"upgrade/legacy-amino-custom/post-restart.json"}
		for _, paths := range []*[5][]string{&authBankPaths, &didPaths, &aolPaths} {
			paths[0] = append(paths[0], legacyPreparationPaths...)
			paths[1] = append(paths[1], legacyCheckpointPaths...)
			paths[2] = append(paths[2], legacyPreservationPaths...)
			paths[3] = append(paths[3], legacyMutationPaths...)
			paths[4] = append(paths[4], legacyRestartPaths...)
		}
	}
	legacyIDs := []string{"legacy-pnft:" + input.Scenario.Name}
	legacyPaths := [5][]string{
		{"upgrade/legacy-pnft-normal-empty.json"},
		{"upgrade/legacy-pnft-normal-empty.json", "upgrade/raw-upgrade-info.json"},
		{"upgrade/nft-empty-post-upgrade-preservation.json", "upgrade/module-versions.json"},
		{"upgrade/legacy-pnft-signed.json", "upgrade/legacy-pnft-new-raw-rejection.json", "tx/committed-results.jsonl", "tx/raw-requests.jsonl"},
		{"upgrade/module-versions-post-restart.json", "upgrade/nft-after-restart.json"},
	}
	if input.LegacyPNFT != nil {
		legacyIDs = []string{
			"legacy-denom:" + input.LegacyPNFT.Prepared.Fixture.DenomID,
			"legacy-nft:" + input.LegacyPNFT.Prepared.Fixture.PNFTID,
		}
		legacyPaths = [5][]string{
			{"upgrade/legacy-pnft-v221-fixture.json"},
			{"upgrade/legacy-pnft-v221-fixture.json", "upgrade/legacy-pnft-old-signature-validation.txt"},
			{"upgrade/legacy-pnft-old-signed-rejection.json", "upgrade/nft-empty-post-upgrade-preservation.json"},
			{"upgrade/legacy-pnft-standard-isolation-post-upgrade.json", "upgrade/legacy-pnft-new-raw-rejection.json", "tx/committed-results.jsonl", "tx/raw-requests.jsonl"},
			{"upgrade/legacy-pnft-standard-isolation-post-restart.json", "upgrade/module-versions-post-restart.json"},
		}
	}

	rows := []harness.UpgradeCoverageRow{
		passedRow(
			harness.UpgradeCoverageAreaAuthBank,
			harness.UpgradeCoveragePriorityP0,
			authBankIDs,
			[]string{"upgrade_deep_helpers_test.go", "upgrade_compatible_signed_tx_test.go", "upgrade_halt_mempool_carrier_test.go", "upgrade_v221_legacy_amino_custom_test.go", "upgrade_test.go"},
			authBankPaths,
		),
		passedRow(
			harness.UpgradeCoverageAreaGov,
			harness.UpgradeCoveragePriorityP0,
			[]string{
				"proposal:" + strconv.FormatUint(input.ProposalID, 10),
				"proposal:" + strconv.FormatUint(input.PostUpgradeProposalID, 10),
			},
			[]string{"upgrade_gov_matrix_test.go"},
			[5][]string{
				{"upgrade/timeline.jsonl"},
				{"state-checkpoints/gov-pre-upgrade.json"},
				{"state-checkpoints/gov-post-upgrade.json", "upgrade/module-versions.json"},
				{"upgrade/gov-expedited-proposal-input.json", "upgrade/post-governance-expedited-proposal.json"},
				{"state-checkpoints/gov-post-restart.json", "state-checkpoints/gov-expedited-post-restart.json", "queries/results.jsonl"},
			},
		),
		passedRow(
			harness.UpgradeCoverageAreaStaking,
			harness.UpgradeCoveragePriorityP0,
			stakingIDs,
			[]string{"upgrade_staking_matrix_test.go", "upgrade_staking_time_queue_test.go", "upgrade_p0_staking_test.go"},
			stakingPaths,
		),
		passedRow(
			harness.UpgradeCoverageAreaDistribution,
			harness.UpgradeCoveragePriorityP0,
			[]string{"distribution-rewards:" + input.Staking.DelegatorAddress},
			[]string{"upgrade_staking_matrix_test.go"},
			[5][]string{
				{"upgrade/staking/preparation.json"},
				{"upgrade/staking/checkpoints/pre-upgrade-checkpoint.json"},
				{"upgrade/staking/checkpoints/post-upgrade-preservation.json"},
				{"upgrade/staking/mutation.json"},
				{"upgrade/staking/checkpoints/post-restart.json"},
			},
		),
		passedRow(
			harness.UpgradeCoverageAreaSlashing,
			harness.UpgradeCoveragePriorityP0,
			slashingIDs,
			[]string{"upgrade_staking_matrix_test.go", "upgrade_slashing_jail_test.go", "upgrade_p0_staking_test.go"},
			slashingPaths,
		),
		passedRow(
			harness.UpgradeCoverageAreaDID,
			harness.UpgradeCoveragePriorityP0,
			didIDs,
			[]string{"upgrade_did_matrix_test.go", "upgrade_v221_legacy_amino_custom_test.go"},
			didPaths,
		),
		passedRow(
			harness.UpgradeCoverageAreaAOL,
			harness.UpgradeCoveragePriorityP0,
			aolIDs,
			[]string{"internal/harness/upgrade_aol_test.go", "upgrade_v221_legacy_amino_custom_test.go"},
			aolPaths,
		),
		passedRow(
			harness.UpgradeCoverageAreaNFT,
			harness.UpgradeCoveragePriorityP0,
			[]string{"nft-class:" + input.NFT.ClassID, "nft:" + input.NFT.ClassID + "/" + input.NFT.NFTID},
			[]string{"nft_lifecycle_test.go", "nft_pagination_test.go"},
			[5][]string{
				{"upgrade/scenario.json"},
				{"upgrade/pre-state.json", "upgrade/raw-upgrade-info.json"},
				{"upgrade/nft-empty-post-upgrade-preservation.json", "upgrade/module-versions.json"},
				{"upgrade/nft-before-restart.json", "tx/committed-results.jsonl"},
				{"upgrade/nft-after-restart.json"},
			},
		),
		passedRow(
			harness.UpgradeCoverageAreaLegacyPNFT,
			harness.UpgradeCoveragePriorityP1,
			legacyIDs,
			[]string{"upgrade_legacy_pnft_fixture_test.go", "x/pnft/legacy"},
			legacyPaths,
		),
		passedRow(
			harness.UpgradeCoverageAreaAuthzFeegrant,
			harness.UpgradeCoveragePriorityP1,
			input.AuthzFeegrant.StateObjectIDs(),
			[]string{"upgrade_authz_feegrant_matrix_test.go"},
			[5][]string{
				{"upgrade/authz-feegrant/preparation.json"},
				{"upgrade/authz-feegrant/checkpoints/pre-upgrade-checkpoint.json"},
				{"upgrade/authz-feegrant/checkpoints/post-upgrade-preservation.json"},
				{"upgrade/authz-feegrant/mutation.json", "upgrade/authz-feegrant/checkpoints/post-upgrade-mutation.json"},
				{"upgrade/authz-feegrant/checkpoints/post-restart.json"},
			},
		),
		passedRow(
			harness.UpgradeCoverageAreaGroupVesting,
			harness.UpgradeCoveragePriorityP1,
			[]string{
				"group:" + strconv.FormatUint(input.GroupVesting.GroupID, 10),
				"vesting-account:" + input.GroupVesting.VestingAccountAddress,
			},
			[]string{"upgrade_group_vesting_matrix_test.go"},
			[5][]string{
				{"upgrade/group-vesting/preparation.json"},
				{"upgrade/group-vesting/checkpoints/pre-upgrade-checkpoint.json"},
				{"upgrade/group-vesting/checkpoints/post-upgrade-preservation.json"},
				{"upgrade/group-vesting/mutation.json", "upgrade/group-vesting/checkpoints/post-upgrade-mutation.json"},
				{"upgrade/group-vesting/post-restart.json"},
			},
		),
		passedRow(
			harness.UpgradeCoverageAreaSystemModules,
			harness.UpgradeCoveragePriorityP1,
			[]string{"mint/burn/consensus/params/capability/crisis", "bank-supply:umed"},
			[]string{"upgrade_system_modules_test.go", "internal/harness/system_invariant_test.go"},
			[5][]string{
				{"upgrade/system-modules/preparation.json"},
				{"upgrade/system-modules/checkpoints/pre-upgrade-checkpoint.json"},
				{"upgrade/system-modules/checkpoints/post-upgrade-preservation.json", "upgrade/system-modules/invariants/post-upgrade-evidence.json"},
				{"upgrade/system-modules/mutation.json", "upgrade/system-modules/checkpoints/post-upgrade-mutation.json"},
				{"upgrade/system-modules/checkpoints/post-restart.json", "upgrade/system-modules/invariants/post-restart-evidence.json"},
			},
		),
	}
	if input.LegacyPNFT == nil {
		const adversarialReason = "non-empty legacy PNFT coverage requires the separate adversarial upgrade-deep scenario"
		for rowIndex := range rows {
			if rows[rowIndex].Area != harness.UpgradeCoverageAreaLegacyPNFT {
				continue
			}
			rows[rowIndex].Status = harness.UpgradeCoverageStatusNotRun
			for phaseIndex := range rows[rowIndex].Phases {
				rows[rowIndex].Phases[phaseIndex].Status = harness.UpgradeCoverageStatusNotRun
				rows[rowIndex].Phases[phaseIndex].ArtifactPaths = nil
				rows[rowIndex].Phases[phaseIndex].Reason = adversarialReason
			}
		}
	}
	if input.Scenario.LegacyPNFTAdversarialFixture {
		const p0BoundaryReason = "strengthened auth/bank, staking, slashing, DID, and AOL evidence runs only in the normal connected P0 boundary lane"
		p0BoundaryAreas := map[harness.UpgradeCoverageArea]struct{}{
			harness.UpgradeCoverageAreaAuthBank: {},
			harness.UpgradeCoverageAreaStaking:  {},
			harness.UpgradeCoverageAreaSlashing: {},
			harness.UpgradeCoverageAreaDID:      {},
			harness.UpgradeCoverageAreaAOL:      {},
		}
		for rowIndex := range rows {
			if _, ok := p0BoundaryAreas[rows[rowIndex].Area]; !ok {
				continue
			}
			rows[rowIndex].Status = harness.UpgradeCoverageStatusNotRun
			rows[rowIndex].QueryCoverage = unsupportedUpgradeQueryCoverage(p0BoundaryReason)
			for phaseIndex := range rows[rowIndex].Phases {
				rows[rowIndex].Phases[phaseIndex].Status = harness.UpgradeCoverageStatusNotRun
				rows[rowIndex].Phases[phaseIndex].ArtifactPaths = nil
				rows[rowIndex].Phases[phaseIndex].Reason = p0BoundaryReason
			}
		}
	}

	const ibcReason = "executed only by the isolated Hermes/Osmosis two-chain ibc-upgrade runner lane"
	ibcPhases := make([]harness.UpgradeCoveragePhase, 0, 5)
	for _, name := range []harness.UpgradeCoveragePhaseName{
		harness.UpgradeCoveragePhaseV221Preparation,
		harness.UpgradeCoveragePhasePreUpgradeCheckpoint,
		harness.UpgradeCoveragePhasePostUpgradePreservation,
		harness.UpgradeCoveragePhasePostUpgradeMutation,
		harness.UpgradeCoveragePhasePostRestart,
	} {
		ibcPhases = append(ibcPhases, harness.UpgradeCoveragePhase{
			Name:   name,
			Status: harness.UpgradeCoverageStatusNotRun,
			Reason: ibcReason,
		})
	}
	rows = append(rows, harness.UpgradeCoverageRow{
		Area:            harness.UpgradeCoverageAreaIBCTransfer,
		Priority:        harness.UpgradeCoveragePriorityP1,
		Status:          harness.UpgradeCoverageStatusNotRun,
		StateObjectIDs:  []string{"ibc:panacea-osmosis-primary-lane"},
		LowerLevelTests: []string{"internal/harness/ibc_upgrade_continuity_test.go"},
		QueryCoverage:   unsupportedUpgradeQueryCoverage(ibcReason),
		Phases:          ibcPhases,
	})

	return harness.UpgradeCoverageMatrix{
		SchemaVersion: harness.UpgradeCoverageMatrixSchemaVersion,
		RecordedAt:    time.Now().UTC(),
		UpgradeName:   upgradeName,
		SourceVersion: "2.2.1",
		TargetVersion: upgradeBinaryVersion,
		Rows:          rows,
	}
}

func connectedUpgradeQueryCoverage(area harness.UpgradeCoverageArea, adversarialLegacy bool) []harness.UpgradeQueryCoverage {
	const unexercised = "this connected live row does not exercise this query boundary"
	coverage := unsupportedUpgradeQueryCoverage(unexercised)
	set := func(
		boundary harness.UpgradeQueryBoundary,
		step string,
		historical bool,
		reason string,
	) {
		for index := range coverage {
			if coverage[index].Boundary != boundary {
				continue
			}
			coverage[index].Supported = true
			coverage[index].Exercised = true
			coverage[index].Reason = reason
			coverage[index].EvidencePaths = []string{"queries/results.jsonl"}
			reference := harness.UpgradeQueryEvidenceReference{
				ArtifactPath:     "queries/results.jsonl",
				Boundary:         boundary,
				Step:             step,
				HistoricalHeight: historical,
			}
			coverage[index].Evidence = []harness.UpgradeQueryEvidenceReference{reference}
			coverage[index].HistoricalHeightSupported = historical
			if historical {
				if boundary == harness.UpgradeQueryBoundaryREST {
					coverage[index].HistoricalHeightExercised = true
					coverage[index].HistoricalHeightReason = "the live REST response returned the requested checkpoint height in server metadata"
					coverage[index].HistoricalHeightEvidencePaths = []string{"queries/results.jsonl"}
					coverage[index].HistoricalHeightEvidence = []harness.UpgradeQueryEvidenceReference{reference}
				} else {
					coverage[index].HistoricalHeightReason = "a height-pinned request succeeded, but this artifact has no server-returned CLI/gRPC response height; historical support is declared but not exercised"
					coverage[index].HistoricalHeightEvidencePaths = []string{harness.UpgradeQueryCoverageArtifactPath}
				}
			} else {
				coverage[index].HistoricalHeightReason = "this lane exercises only latest-state queries for this boundary"
			}
		}
	}
	appendEvidence := func(
		boundary harness.UpgradeQueryBoundary,
		step string,
		historical bool,
	) {
		for index := range coverage {
			if coverage[index].Boundary != boundary {
				continue
			}
			coverage[index].Evidence = append(
				coverage[index].Evidence,
				harness.UpgradeQueryEvidenceReference{
					ArtifactPath:     "queries/results.jsonl",
					Boundary:         boundary,
					Step:             step,
					HistoricalHeight: historical,
				},
			)
		}
	}
	declaredSupported := func(boundary harness.UpgradeQueryBoundary, reason string) {
		for index := range coverage {
			if coverage[index].Boundary != boundary {
				continue
			}
			coverage[index].Supported = true
			coverage[index].Reason = reason + "; this lane has no structured live transport record, so exercised=false"
			coverage[index].HistoricalHeightReason = reason + "; no server-validated historical response is claimed"
		}
	}
	switch area {
	case harness.UpgradeCoverageAreaAuthBank:
		set(harness.UpgradeQueryBoundaryCLI, "upgrade-compatible-old-signer-post-restart", false, "auth account state is queried through the CLI")
		declaredSupported(harness.UpgradeQueryBoundaryGRPC, "typed bank gRPC is available and used by the scenario")
	case harness.UpgradeCoverageAreaGov:
		set(harness.UpgradeQueryBoundaryCLI, "upgrade-expedited-post-restart", false, "governance params, tally, and proposals are queried through the CLI")
	case harness.UpgradeCoverageAreaStaking:
		set(harness.UpgradeQueryBoundaryCLI, "upgrade-staking-post-restart-delegation", true, "staking state is queried through the CLI")
		declaredSupported(harness.UpgradeQueryBoundaryGRPC, "typed bank gRPC is available and used as supplementary state")
	case harness.UpgradeCoverageAreaDistribution:
		set(harness.UpgradeQueryBoundaryCLI, "upgrade-staking-post-restart-delegator-rewards", true, "distribution state is queried through a height-pinned CLI command")
	case harness.UpgradeCoverageAreaSlashing:
		set(harness.UpgradeQueryBoundaryCLI, "upgrade-staking-post-restart-signing-info", true, "slashing state is queried through a height-pinned CLI command")
	case harness.UpgradeCoverageAreaDID:
		set(harness.UpgradeQueryBoundaryCLI, "upgrade-post-restart-updated-did", true, "DID documents are queried through a height-pinned CLI command")
	case harness.UpgradeCoverageAreaAOL:
		set(harness.UpgradeQueryBoundaryCLI, "upgrade-aol-post-restart-topic", true, "AOL state is queried through the CLI")
		set(harness.UpgradeQueryBoundaryREST, "upgrade-aol-post-restart-owner-topics", true, "AOL pagination is queried through REST")
	case harness.UpgradeCoverageAreaNFT:
		set(harness.UpgradeQueryBoundaryCLI, "upgrade-class-after-restart-panacea-cli", true, "NFT query parity exercises the CLI")
		set(harness.UpgradeQueryBoundaryGRPC, "upgrade-class-after-restart-panacea-grpc", true, "NFT query parity exercises gRPC")
		set(harness.UpgradeQueryBoundaryREST, "upgrade-class-after-restart-panacea-class-raw-rest", true, "NFT query parity exercises REST")
	case harness.UpgradeCoverageAreaLegacyPNFT:
		if adversarialLegacy {
			set(harness.UpgradeQueryBoundaryCLI, "v221-legacy-pnft-denom", false, "v2.2.1 legacy state is queried through its CLI before the upgrade")
			set(harness.UpgradeQueryBoundaryGRPC, "upgrade-post-upgrade-preservation-empty-panacea-nft-records", true, "the current Panacea NFT service proves that no standard or Panacea NFT state was migrated at the legacy canonical class ID")
			appendEvidence(harness.UpgradeQueryBoundaryGRPC, "upgrade-post-upgrade-preservation-panacea-class-record-absent", true)
			appendEvidence(harness.UpgradeQueryBoundaryGRPC, "upgrade-post-upgrade-preservation-panacea-nft-record-absent", true)
		} else {
			set(harness.UpgradeQueryBoundaryREST, "v221-legacy-pnft-empty", true, "the complete legacy denom page is queried through height-pinned REST")
		}
	case harness.UpgradeCoverageAreaAuthzFeegrant:
		declaredSupported(harness.UpgradeQueryBoundaryGRPC, "typed authz and feegrant gRPC calls are available and used")
	case harness.UpgradeCoverageAreaGroupVesting:
		set(harness.UpgradeQueryBoundaryCLI, "upgrade-group-vesting-post-restart-query-group-info", true, "group and vesting state is queried through height-pinned CLI commands")
	case harness.UpgradeCoverageAreaSystemModules:
		declaredSupported(harness.UpgradeQueryBoundaryGRPC, "typed system-module gRPC calls are available and used")
		set(harness.UpgradeQueryBoundaryREST, "upgrade-system-post-restart-mint-params", true, "mint state is queried through height-pinned REST")
	}
	return coverage
}

func unsupportedUpgradeQueryCoverage(reason string) []harness.UpgradeQueryCoverage {
	result := make([]harness.UpgradeQueryCoverage, 0, 3)
	for _, boundary := range []harness.UpgradeQueryBoundary{
		harness.UpgradeQueryBoundaryCLI,
		harness.UpgradeQueryBoundaryGRPC,
		harness.UpgradeQueryBoundaryREST,
	} {
		result = append(result, harness.UpgradeQueryCoverage{
			Boundary:                      boundary,
			Reason:                        reason,
			EvidencePaths:                 []string{harness.UpgradeQueryCoverageArtifactPath},
			HistoricalHeightReason:        reason + "; no height-pinned evidence was produced",
			HistoricalHeightEvidencePaths: []string{harness.UpgradeQueryCoverageArtifactPath},
		})
	}
	return result
}

func recordConnectedUpgradeCoverage(
	network *harness.Network,
	input connectedUpgradeCoverageInput,
) error {
	if err := validateConnectedUpgradeCoverageInput(input); err != nil {
		return fmt.Errorf("validate connected upgrade coverage input: %w", err)
	}
	matrix := buildConnectedUpgradeCoverageMatrix(input)
	if err := matrix.Validate(); err != nil {
		return fmt.Errorf("validate connected upgrade coverage matrix: %w", err)
	}
	return network.RecordUpgradeCoverageMatrix(matrix)
}

func validateConnectedUpgradeCoverageInput(input connectedUpgradeCoverageInput) error {
	if input.Scenario.LegacyPNFTAdversarialFixture {
		return nil
	}
	if !input.Scenario.RunP0BoundaryMatrix {
		return fmt.Errorf("normal connected upgrade coverage requires the P0 boundary matrix")
	}
	if input.HaltMempool == nil {
		return fmt.Errorf("normal connected upgrade coverage requires halt-mempool evidence")
	}
	if err := validateConnectedUpgradeHaltMempoolEvidence(*input.HaltMempool); err != nil {
		return fmt.Errorf("halt-mempool evidence: %w", err)
	}
	if input.StakingTimeQueues == nil {
		return fmt.Errorf("normal connected upgrade coverage requires staking time-queue evidence")
	}
	if err := validateUpgradeP0StakingQueueEvidence(*input.StakingTimeQueues); err != nil {
		return fmt.Errorf("staking time-queue evidence: %w", err)
	}
	if input.SlashingJail == nil {
		return fmt.Errorf("normal connected upgrade coverage requires slashing jail evidence")
	}
	if err := validateUpgradeP0SlashingEvidence(*input.SlashingJail); err != nil {
		return fmt.Errorf("slashing jail evidence: %w", err)
	}
	if input.LegacyAminoCustom == nil {
		return fmt.Errorf("normal connected upgrade coverage requires legacy-amino DID and AOL evidence")
	}
	for _, item := range []struct {
		label string
		kind  upgradeV221LegacyAminoCustomTxKind
		tx    upgradeV221LegacyAminoCustomTxFixture
	}{
		{label: "AOL", kind: upgradeV221LegacyAminoAOLCreateTopic, tx: input.LegacyAminoCustom.AOL},
		{label: "DID", kind: upgradeV221LegacyAminoDIDUpdate, tx: input.LegacyAminoCustom.DID},
	} {
		if item.tx.Kind != item.kind {
			return fmt.Errorf("legacy-amino %s evidence kind %q, want %q", item.label, item.tx.Kind, item.kind)
		}
		if item.tx.SignMode != upgradeV221LegacyAminoSignMode {
			return fmt.Errorf(
				"legacy-amino %s evidence sign mode %q, want %q",
				item.label,
				item.tx.SignMode,
				upgradeV221LegacyAminoSignMode,
			)
		}
		if item.tx.TamperedCheckTx == nil ||
			item.tx.TamperedCheckTx.Height != "0" ||
			item.tx.TamperedCheckTx.Codespace != upgradeV221LegacyAminoSDKSpace ||
			item.tx.TamperedCheckTx.Code != upgradeV221LegacyAminoSDKCode ||
			item.tx.TamperedCheckTx.TxHash == "" {
			return fmt.Errorf(
				"legacy-amino %s tamper evidence must prove exact sdk/4 CheckTx rejection: %+v",
				item.label,
				item.tx.TamperedCheckTx,
			)
		}
		if item.tx.TxHash == "" {
			return fmt.Errorf("legacy-amino %s evidence requires a committed tx hash", item.label)
		}
	}
	return nil
}

func validateConnectedUpgradeHaltMempoolEvidence(evidence upgradeHaltMempoolFixture) error {
	if len(evidence.Transactions) != upgradeHaltMempoolTxCount {
		return fmt.Errorf(
			"has %d transactions, want %d",
			len(evidence.Transactions),
			upgradeHaltMempoolTxCount,
		)
	}
	for index, transaction := range evidence.Transactions {
		if transaction.CheckTx == nil || transaction.CheckTx.Height != "0" ||
			transaction.CheckTx.Code != 0 || transaction.CheckTx.TxHash == "" {
			return fmt.Errorf("transaction %d has no successful CheckTx acceptance: %+v", index, transaction.CheckTx)
		}
		if transaction.Committed == nil || transaction.Committed.HeightInt64() <= 0 ||
			transaction.Committed.Code != 0 ||
			!strings.EqualFold(transaction.Committed.TxHash, transaction.CheckTx.TxHash) {
			return fmt.Errorf(
				"transaction %d has no exact CheckTx-to-commit hash lineage: check=%+v committed=%+v",
				index,
				transaction.CheckTx,
				transaction.Committed,
			)
		}
	}
	wantObserved := evidence.InitialSequence + uint64(len(evidence.Transactions))
	if evidence.Reconciliation.InitialSequence != evidence.InitialSequence ||
		evidence.Reconciliation.ObservedSequence != wantObserved ||
		evidence.Reconciliation.TransactionCount != len(evidence.Transactions) ||
		evidence.Reconciliation.CommittedPrefix != len(evidence.Transactions) ||
		len(evidence.Reconciliation.MissingSuffix) != 0 {
		return fmt.Errorf(
			"reconciliation does not prove the complete halt-mempool carrier set: %+v",
			evidence.Reconciliation,
		)
	}
	return nil
}
