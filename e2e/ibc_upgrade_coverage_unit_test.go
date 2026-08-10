package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestBuildIBCUpgradeCoverageMatrixIsHonestAboutIsolatedLane(t *testing.T) {
	t.Parallel()
	evidence := harness.IBCUpgradeContinuityEvidence{OriginalChannel: harness.IBCChannelHandshake{
		Panacea: harness.IBCChannelEndpoint{
			ChainID: "panacea-v221", ClientID: "07-tendermint-0", ConnectionID: "connection-0",
			PortID: "transfer", ChannelID: "channel-0",
		},
		Osmosis: harness.IBCChannelEndpoint{
			ChainID: "osmosis-v31", ClientID: "07-tendermint-1", ConnectionID: "connection-0",
			PortID: "transfer", ChannelID: "channel-0",
		},
	}}
	matrix := buildIBCUpgradeCoverageMatrix(evidence)
	require.NoError(t, matrix.Validate())
	require.Len(t, matrix.Rows, 13)
	for _, row := range matrix.Rows {
		if row.Area == harness.UpgradeCoverageAreaIBCTransfer {
			require.Equal(t, harness.UpgradeCoverageStatusPassed, row.Status)
			require.False(t, row.QueryCoverage[0].Supported)
			grpc := row.QueryCoverage[1]
			require.Equal(t, harness.UpgradeQueryBoundaryGRPC, grpc.Boundary)
			require.True(t, grpc.Supported)
			require.True(t, grpc.Exercised)
			require.Equal(t, []string{"queries/results.jsonl"}, grpc.EvidencePaths)
			require.Len(t, grpc.Evidence, 5)
			for _, reference := range grpc.Evidence {
				require.Equal(t, harness.UpgradeQueryBoundaryGRPC, reference.Boundary)
				require.NotEmpty(t, reference.Step)
				require.False(t, reference.HistoricalHeight)
			}
			require.False(t, grpc.HistoricalHeightExercised)
			require.False(t, row.QueryCoverage[2].Supported)
			for _, phase := range row.Phases {
				require.Equal(t, harness.UpgradeCoverageStatusPassed, phase.Status)
				require.NotEmpty(t, phase.ArtifactPaths)
				if phase.Name == harness.UpgradeCoveragePhaseV221Preparation {
					require.Equal(t, []string{
						"ibc/provenance.json",
						"ibc/osmosis-source-contract.json",
						"ibc/resolved-images.json",
						"ibc/hermes/runtime-identity.json",
						"ibc/hermes/binary-sha256.txt",
						"ibc/chains/panacea/pre-upgrade/identity.json",
						"ibc/chains/osmosis/identity.json",
						"ibc/state/pre-upgrade-channel.json",
						"ibc/state/pre-upgrade-bidirectional.json",
					}, phase.ArtifactPaths)
				}
				if phase.Name == harness.UpgradeCoveragePhasePostUpgradePreservation {
					require.Equal(t, []string{
						"ibc-compatibility-matrix.json",
						"ibc/hermes/upgrade-invariance.json",
						"ibc/chains/panacea/post-upgrade/identity.json",
						"ibc/chains/osmosis/genesis-checksums.json",
						"ibc/upgrade/panacea-step.json",
						"ibc/upgrade/osmosis-height-progress.json",
						"ibc/upgrade/post-upgrade-before-relay.json",
					}, phase.ArtifactPaths)
				}
				if phase.Name == harness.UpgradeCoveragePhasePostRestart {
					require.Equal(t, []string{
						"ibc/upgrade/continuity.json",
						"ibc/upgrade/all-node-restarts.json",
						"ibc/state/post-restart-node-semantics.json",
						"ibc/state/post-restart-balances.json",
						"ibc/state/post-restart-escrow-balances.json",
						"ibc/state/post-restart-denom-traces.json",
					}, phase.ArtifactPaths)
				}
			}
			continue
		}
		for _, claim := range row.QueryCoverage {
			require.False(t, claim.Supported, "%s/%s lacks a structured query transport record", row.Area, claim.Boundary)
			require.Empty(t, claim.Evidence)
			require.False(t, claim.HistoricalHeightSupported)
			require.Empty(t, claim.HistoricalHeightEvidence)
		}
		require.Equal(t, harness.UpgradeCoverageStatusNotRun, row.Status)
		for _, phase := range row.Phases {
			require.Equal(t, harness.UpgradeCoverageStatusNotRun, phase.Status)
			require.NotEmpty(t, phase.Reason)
		}
	}
}
