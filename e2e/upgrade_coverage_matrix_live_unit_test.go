package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestBuildConnectedUpgradeCoverageMatrixIsHonestAboutSeparateIBCLane(t *testing.T) {
	t.Parallel()

	matrix := buildConnectedUpgradeCoverageMatrix(connectedUpgradeCoverageInput{
		Scenario: upgradeRunScenario{Name: "normal-empty-legacy-pnft"},
		Preserved: upgradePreservedState{
			Address: "panacea1account",
			DID: upgradeDIDFixtures{
				Updated:     upgradeDIDFixture{DID: "did:panacea:updated"},
				Deactivated: upgradeDIDFixture{DID: "did:panacea:deactivated"},
			},
			AOLFixture: harness.AOLUpgradeFixture{TopicName: "topic"},
		},
		ProposalID:            1,
		PostUpgradeProposalID: 2,
		Staking: upgradeStakingFixture{
			DelegatorAddress:       "panacea1delegator",
			ValidatorOperator:      "panaceavaloper1operator",
			ValidatorConsensusAddr: "panaceavalcons1consensus",
		},
		AuthzFeegrant: upgradeAuthzFeegrantFixture{
			GranterAddress:      "panacea1granter",
			GranteeAddress:      "panacea1grantee",
			AuthzMessageTypeURL: "/cosmos.bank.v1beta1.MsgSend",
		},
		GroupVesting: upgradeGroupVestingFixture{
			GroupID:               7,
			VestingAccountAddress: "panacea1vesting",
		},
		NFT: nftLifecycleEvidence{ClassID: "panacea1creator:class", NFTID: "nft-1"},
		CompatibleSignedBank: upgradeCompatibleSignedTxFixture{
			SignerAddress: "panacea1oldsigned",
			TxHash:        "ABC123",
		},
	})

	require.NoError(t, matrix.Validate())
	require.Len(t, matrix.Rows, 13)
	for _, row := range matrix.Rows {
		for _, claim := range row.QueryCoverage {
			if claim.Exercised {
				require.NotEmpty(t, claim.Evidence, "%s/%s", row.Area, claim.Boundary)
				for _, reference := range claim.Evidence {
					require.Equal(t, "queries/results.jsonl", reference.ArtifactPath)
					require.Equal(t, claim.Boundary, reference.Boundary)
					require.NotEmpty(t, reference.Step)
				}
			} else {
				require.Empty(t, claim.Evidence, "%s/%s", row.Area, claim.Boundary)
			}
			if claim.HistoricalHeightExercised {
				require.Equal(t, harness.UpgradeQueryBoundaryREST, claim.Boundary)
				require.NotEmpty(t, claim.HistoricalHeightEvidence)
			}
		}
		if row.Area == harness.UpgradeCoverageAreaIBCTransfer || row.Area == harness.UpgradeCoverageAreaLegacyPNFT {
			require.Equal(t, harness.UpgradeCoverageStatusNotRun, row.Status)
			for _, phase := range row.Phases {
				require.Equal(t, harness.UpgradeCoverageStatusNotRun, phase.Status)
				if row.Area == harness.UpgradeCoverageAreaIBCTransfer {
					require.Contains(t, phase.Reason, "ibc-upgrade runner lane")
				} else {
					require.Contains(t, phase.Reason, "adversarial")
				}
			}
			continue
		}
		require.Equal(t, harness.UpgradeCoverageStatusPassed, row.Status, row.Area)
	}
}

func TestBuildConnectedUpgradeCoverageMatrixUsesAdversarialLegacyEvidence(t *testing.T) {
	t.Parallel()

	input := connectedUpgradeCoverageInput{
		Scenario: upgradeRunScenario{
			Name:                         "adversarial-non-empty-legacy-pnft",
			LegacyPNFTAdversarialFixture: true,
		},
		Preserved: upgradePreservedState{
			Address: "panacea1account",
			DID: upgradeDIDFixtures{
				Updated:     upgradeDIDFixture{DID: "did:panacea:updated"},
				Deactivated: upgradeDIDFixture{DID: "did:panacea:deactivated"},
			},
			AOLFixture: harness.AOLUpgradeFixture{TopicName: "topic"},
		},
		ProposalID:            1,
		PostUpgradeProposalID: 2,
		Staking: upgradeStakingFixture{
			DelegatorAddress:       "panacea1delegator",
			ValidatorOperator:      "panaceavaloper1operator",
			ValidatorConsensusAddr: "panaceavalcons1consensus",
		},
		AuthzFeegrant: upgradeAuthzFeegrantFixture{
			GranterAddress:      "panacea1granter",
			GranteeAddress:      "panacea1grantee",
			AuthzMessageTypeURL: "/cosmos.bank.v1beta1.MsgSend",
		},
		GroupVesting: upgradeGroupVestingFixture{
			GroupID:               7,
			VestingAccountAddress: "panacea1vesting",
		},
		NFT: nftLifecycleEvidence{ClassID: "panacea1creator:class", NFTID: "nft-1"},
		CompatibleSignedBank: upgradeCompatibleSignedTxFixture{
			SignerAddress: "panacea1oldsigned",
			TxHash:        "ABC123",
		},
		LegacyPNFT: &legacyPNFTUpgradeRunState{Prepared: preparedLegacyPNFTFixture{
			Fixture: legacyPNFTFixture{DenomID: "panacea1legacy:denom", PNFTID: "legacy-nft"},
		}},
	}

	matrix := buildConnectedUpgradeCoverageMatrix(input)
	require.NoError(t, matrix.Validate())
	for _, row := range matrix.Rows {
		if row.Area != harness.UpgradeCoverageAreaLegacyPNFT {
			continue
		}
		require.Equal(t, []string{
			"legacy-denom:panacea1legacy:denom",
			"legacy-nft:legacy-nft",
		}, row.StateObjectIDs)
		require.Contains(t, row.Phases[0].ArtifactPaths, "upgrade/legacy-pnft-v221-fixture.json")
		require.Contains(t, row.Phases[4].ArtifactPaths, "upgrade/legacy-pnft-standard-isolation-post-restart.json")
		for _, claim := range row.QueryCoverage {
			if claim.Boundary != harness.UpgradeQueryBoundaryGRPC {
				continue
			}
			require.True(t, claim.Exercised)
			require.Len(t, claim.Evidence, 3)
			steps := make([]string, 0, len(claim.Evidence))
			for _, evidence := range claim.Evidence {
				steps = append(steps, evidence.Step)
			}
			require.ElementsMatch(t, []string{
				"upgrade-post-upgrade-preservation-empty-panacea-nft-records",
				"upgrade-post-upgrade-preservation-panacea-class-record-absent",
				"upgrade-post-upgrade-preservation-panacea-nft-record-absent",
			}, steps)
			return
		}
		t.Fatal("legacy PNFT gRPC coverage claim not found")
		return
	}
	t.Fatal("legacy PNFT coverage row not found")
}
