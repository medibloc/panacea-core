package e2e_test

import (
	"time"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

type ibcCoverageRequirement struct {
	area     harness.UpgradeCoverageArea
	priority harness.UpgradeCoveragePriority
}

var ibcCoverageRequirements = []ibcCoverageRequirement{
	{harness.UpgradeCoverageAreaAuthBank, harness.UpgradeCoveragePriorityP0},
	{harness.UpgradeCoverageAreaGov, harness.UpgradeCoveragePriorityP0},
	{harness.UpgradeCoverageAreaStaking, harness.UpgradeCoveragePriorityP0},
	{harness.UpgradeCoverageAreaDistribution, harness.UpgradeCoveragePriorityP0},
	{harness.UpgradeCoverageAreaSlashing, harness.UpgradeCoveragePriorityP0},
	{harness.UpgradeCoverageAreaDID, harness.UpgradeCoveragePriorityP0},
	{harness.UpgradeCoverageAreaAOL, harness.UpgradeCoveragePriorityP0},
	{harness.UpgradeCoverageAreaNFT, harness.UpgradeCoveragePriorityP0},
	{harness.UpgradeCoverageAreaLegacyPNFT, harness.UpgradeCoveragePriorityP1},
	{harness.UpgradeCoverageAreaAuthzFeegrant, harness.UpgradeCoveragePriorityP1},
	{harness.UpgradeCoverageAreaGroupVesting, harness.UpgradeCoveragePriorityP1},
	{harness.UpgradeCoverageAreaSystemModules, harness.UpgradeCoveragePriorityP1},
	{harness.UpgradeCoverageAreaIBCTransfer, harness.UpgradeCoveragePriorityP1},
}

func buildIBCUpgradeCoverageMatrix(evidence harness.IBCUpgradeContinuityEvidence) harness.UpgradeCoverageMatrix {
	phaseNames := []harness.UpgradeCoveragePhaseName{
		harness.UpgradeCoveragePhaseV221Preparation,
		harness.UpgradeCoveragePhasePreUpgradeCheckpoint,
		harness.UpgradeCoveragePhasePostUpgradePreservation,
		harness.UpgradeCoveragePhasePostUpgradeMutation,
		harness.UpgradeCoveragePhasePostRestart,
	}
	const separateLaneReason = "executed by the connected upgrade-deep runner lane, not the isolated IBC topology"
	rows := make([]harness.UpgradeCoverageRow, 0, len(ibcCoverageRequirements))
	for _, requirement := range ibcCoverageRequirements {
		phases := make([]harness.UpgradeCoveragePhase, 0, len(phaseNames))
		for _, name := range phaseNames {
			phases = append(phases, harness.UpgradeCoveragePhase{
				Name:   name,
				Status: harness.UpgradeCoverageStatusNotRun,
				Reason: separateLaneReason,
			})
		}
		row := harness.UpgradeCoverageRow{
			Area:            requirement.area,
			Priority:        requirement.priority,
			Status:          harness.UpgradeCoverageStatusNotRun,
			StateObjectIDs:  []string{"separate-connected-upgrade-lane:" + string(requirement.area)},
			LowerLevelTests: []string{"upgrade_coverage_matrix_live_unit_test.go"},
			QueryCoverage:   unsupportedUpgradeQueryCoverage(separateLaneReason),
			Phases:          phases,
		}
		if requirement.area == harness.UpgradeCoverageAreaIBCTransfer {
			channel := evidence.OriginalChannel
			row.Status = harness.UpgradeCoverageStatusPassed
			row.StateObjectIDs = []string{
				"ibc-client:" + channel.Panacea.ChainID + "/" + channel.Panacea.ClientID,
				"ibc-connection:" + channel.Panacea.ChainID + "/" + channel.Panacea.ConnectionID,
				"ibc-channel:" + channel.Panacea.ChainID + "/" + channel.Panacea.PortID + "/" + channel.Panacea.ChannelID,
				"ibc-channel:" + channel.Osmosis.ChainID + "/" + channel.Osmosis.PortID + "/" + channel.Osmosis.ChannelID,
			}
			row.LowerLevelTests = []string{
				"internal/harness/ibc_upgrade_continuity_test.go",
				"internal/harness/ibc_transfer_test.go",
			}
			row.QueryCoverage = unsupportedUpgradeQueryCoverage("the isolated IBC lane does not exercise this query boundary")
			row.QueryCoverage[0].Reason = "IBC state is queried through topology helpers, but this lane does not emit a structured CLI transport record"
			row.QueryCoverage[0].HistoricalHeightReason = "the isolated IBC lane does not claim a structured height-pinned CLI transport record"
			row.QueryCoverage[1] = harness.IBCUpgradeGRPCQueryCoverage()
			row.Phases = []harness.UpgradeCoveragePhase{
				passedIBCCoveragePhase(harness.UpgradeCoveragePhaseV221Preparation,
					"ibc/provenance.json", "ibc/osmosis-source-contract.json", "ibc/resolved-images.json",
					"ibc/hermes/runtime-identity.json", "ibc/hermes/binary-sha256.txt",
					"ibc/chains/panacea/pre-upgrade/identity.json",
					"ibc/chains/osmosis/identity.json",
					"ibc/state/pre-upgrade-channel.json", "ibc/state/pre-upgrade-bidirectional.json"),
				passedIBCCoveragePhase(harness.UpgradeCoveragePhasePreUpgradeCheckpoint,
					"ibc/upgrade/in-flight-staged.json"),
				passedIBCCoveragePhase(harness.UpgradeCoveragePhasePostUpgradePreservation,
					"ibc-compatibility-matrix.json",
					"ibc/hermes/upgrade-invariance.json",
					"ibc/chains/panacea/post-upgrade/identity.json",
					"ibc/chains/osmosis/genesis-checksums.json",
					"ibc/upgrade/panacea-step.json", "ibc/upgrade/osmosis-height-progress.json",
					"ibc/upgrade/post-upgrade-before-relay.json"),
				passedIBCCoveragePhase(harness.UpgradeCoveragePhasePostUpgradeMutation,
					"ibc/upgrade/in-flight-relay.json", "ibc/upgrade/timeout-refund.json", "ibc/state/post-upgrade-bidirectional.json"),
				passedIBCCoveragePhase(harness.UpgradeCoveragePhasePostRestart,
					"ibc/upgrade/continuity.json", "ibc/upgrade/all-node-restarts.json",
					"ibc/state/post-restart-node-semantics.json",
					"ibc/state/post-restart-balances.json", "ibc/state/post-restart-escrow-balances.json",
					"ibc/state/post-restart-denom-traces.json"),
			}
		}
		rows = append(rows, row)
	}
	return harness.UpgradeCoverageMatrix{
		SchemaVersion: harness.UpgradeCoverageMatrixSchemaVersion,
		RecordedAt:    time.Now().UTC(),
		UpgradeName:   upgradeName,
		SourceVersion: "2.2.1",
		TargetVersion: upgradeBinaryVersion,
		Rows:          rows,
	}
}

func passedIBCCoveragePhase(name harness.UpgradeCoveragePhaseName, artifactPaths ...string) harness.UpgradeCoveragePhase {
	return harness.UpgradeCoveragePhase{
		Name:          name,
		Status:        harness.UpgradeCoverageStatusPassed,
		ArtifactPaths: artifactPaths,
	}
}
