package e2e_test

import (
	"testing"
	"time"

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
	}, "2.3.1")
	require.Equal(t, "2.3.1", matrix.TargetVersion)

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
			if row.Area == harness.UpgradeCoverageAreaDID && claim.Boundary == harness.UpgradeQueryBoundaryCLI {
				require.True(t, claim.HistoricalHeightSupported)
				require.Len(t, claim.Evidence, 1)
				require.True(t, claim.Evidence[0].HistoricalHeight)
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

	matrix := buildConnectedUpgradeCoverageMatrix(input, "2.3.1")
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

func TestBuildConnectedUpgradeCoverageMatrixLinksStrengthenedP0Evidence(t *testing.T) {
	t.Parallel()

	stakingQueues := validConnectedUpgradeStakingTimeQueueEvidence()
	stakingQueues.SourceValidator = "panaceavaloper1source"
	stakingQueues.DestinationValidator = "panaceavaloper1destination"
	slashing := validConnectedUpgradeSlashingJailEvidence()
	legacy := minimallyValidConnectedUpgradeLegacyAminoEvidence()
	legacy.AOL.StateObjectID = "panacea1aol/legacy-topic"
	legacy.DID.StateObjectID = "did:panacea:legacy"
	halt := minimallyValidConnectedUpgradeHaltMempoolEvidence()
	halt.SignerAddress = "panacea1halt"

	matrix := buildConnectedUpgradeCoverageMatrix(connectedUpgradeCoverageInput{
		Scenario: upgradeRunScenario{
			Name:                "normal-empty-legacy-pnft",
			RunP0BoundaryMatrix: true,
		},
		Preserved: upgradePreservedState{
			Address: "panacea1account",
			DID: upgradeDIDFixtures{
				Updated:     upgradeDIDFixture{DID: "did:panacea:updated"},
				Deactivated: upgradeDIDFixture{DID: "did:panacea:deactivated"},
			},
			AOLFixture: harness.AOLUpgradeFixture{TopicName: "topic"},
		},
		Staking: upgradeStakingFixture{
			DelegatorAddress:       "panacea1delegator",
			ValidatorOperator:      "panaceavaloper1operator",
			ValidatorConsensusAddr: "panaceavalcons1consensus",
		},
		ProposalID:            1,
		PostUpgradeProposalID: 2,
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
			TxHash:        "DIRECT123",
		},
		HaltMempool:       halt,
		StakingTimeQueues: &stakingQueues,
		SlashingJail:      &slashing,
		LegacyAminoCustom: &legacy,
	}, "2.3.1")
	require.NoError(t, matrix.Validate())

	expected := map[harness.UpgradeCoverageArea]struct {
		stateID string
		path    string
	}{
		harness.UpgradeCoverageAreaAuthBank: {
			stateID: "halt-mempool-signer:panacea1halt",
			path:    "upgrade/halt-mempool-carrier/carrier-commit.json",
		},
		harness.UpgradeCoverageAreaStaking: {
			stateID: "redelegation-time-queue:panaceavaloper1source/panaceavaloper1destination",
			path:    "upgrade/staking-time-queue/post-upgrade-pending.json",
		},
		harness.UpgradeCoverageAreaSlashing: {
			stateID: "downtime-validator:panaceavaloper1target",
			path:    "upgrade/system-modules/p0-slashing-supply-accounting.json",
		},
		harness.UpgradeCoverageAreaDID: {
			stateID: "legacy-amino-did:did:panacea:legacy",
			path:    "upgrade/legacy-amino-custom/post-upgrade.json",
		},
		harness.UpgradeCoverageAreaAOL: {
			stateID: "legacy-amino-aol:panacea1aol/legacy-topic",
			path:    "upgrade/legacy-amino-custom/post-upgrade.json",
		},
	}
	for _, row := range matrix.Rows {
		want, ok := expected[row.Area]
		if !ok {
			continue
		}
		require.Equal(t, harness.UpgradeCoverageStatusPassed, row.Status, row.Area)
		require.Contains(t, row.StateObjectIDs, want.stateID, row.Area)
		var paths []string
		for _, phase := range row.Phases {
			paths = append(paths, phase.ArtifactPaths...)
		}
		require.Contains(t, paths, want.path, row.Area)
		delete(expected, row.Area)
	}
	require.Empty(t, expected)
}

func TestBuildConnectedUpgradeCoverageMatrixMarksStrengthenedP0RowsNotRunInAdversarialLane(t *testing.T) {
	t.Parallel()

	matrix := buildConnectedUpgradeCoverageMatrix(connectedUpgradeCoverageInput{
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
		Staking: upgradeStakingFixture{
			DelegatorAddress:       "panacea1delegator",
			ValidatorOperator:      "panaceavaloper1operator",
			ValidatorConsensusAddr: "panaceavalcons1consensus",
		},
		LegacyPNFT: &legacyPNFTUpgradeRunState{Prepared: preparedLegacyPNFTFixture{
			Fixture: legacyPNFTFixture{DenomID: "panacea1legacy:denom", PNFTID: "legacy-nft"},
		}},
	}, "2.3.1")

	wantNotRun := map[harness.UpgradeCoverageArea]struct{}{
		harness.UpgradeCoverageAreaAuthBank: {},
		harness.UpgradeCoverageAreaStaking:  {},
		harness.UpgradeCoverageAreaSlashing: {},
		harness.UpgradeCoverageAreaDID:      {},
		harness.UpgradeCoverageAreaAOL:      {},
	}
	for _, row := range matrix.Rows {
		if _, ok := wantNotRun[row.Area]; !ok {
			continue
		}
		require.Equal(t, harness.UpgradeCoverageStatusNotRun, row.Status, row.Area)
		for _, phase := range row.Phases {
			require.Equal(t, harness.UpgradeCoverageStatusNotRun, phase.Status, row.Area)
			require.Empty(t, phase.ArtifactPaths, row.Area)
			require.Contains(t, phase.Reason, "normal connected P0 boundary lane", row.Area)
		}
		for _, claim := range row.QueryCoverage {
			require.False(t, claim.Exercised, "%s/%s", row.Area, claim.Boundary)
			require.Contains(t, claim.Reason, "normal connected P0 boundary lane", row.Area)
		}
		delete(wantNotRun, row.Area)
	}
	require.Empty(t, wantNotRun)
}

func TestValidateConnectedUpgradeCoverageInputRequiresHaltMempoolEvidenceForNormalLane(t *testing.T) {
	t.Parallel()

	err := validateConnectedUpgradeCoverageInput(connectedUpgradeCoverageInput{
		Scenario: upgradeRunScenario{
			Name:                "normal-empty-legacy-pnft",
			RunP0BoundaryMatrix: true,
		},
	})
	require.ErrorContains(t, err, "halt-mempool")
}

func TestValidateConnectedUpgradeCoverageInputRejectsIncompleteHaltMempoolEvidence(t *testing.T) {
	t.Parallel()

	err := validateConnectedUpgradeCoverageInput(connectedUpgradeCoverageInput{
		Scenario: upgradeRunScenario{
			Name:                "normal-empty-legacy-pnft",
			RunP0BoundaryMatrix: true,
		},
		HaltMempool: &upgradeHaltMempoolFixture{},
	})
	require.ErrorContains(t, err, "transactions")
}

func TestValidateConnectedUpgradeCoverageInputRequiresStakingTimeQueueEvidenceForNormalLane(t *testing.T) {
	t.Parallel()

	err := validateConnectedUpgradeCoverageInput(connectedUpgradeCoverageInput{
		Scenario: upgradeRunScenario{
			Name:                "normal-empty-legacy-pnft",
			RunP0BoundaryMatrix: true,
		},
		HaltMempool: minimallyValidConnectedUpgradeHaltMempoolEvidence(),
	})
	require.ErrorContains(t, err, "staking time-queue")
}

func TestValidateConnectedUpgradeCoverageInputRejectsIncompleteStakingTimeQueueEvidence(t *testing.T) {
	t.Parallel()

	err := validateConnectedUpgradeCoverageInput(connectedUpgradeCoverageInput{
		Scenario: upgradeRunScenario{
			Name:                "normal-empty-legacy-pnft",
			RunP0BoundaryMatrix: true,
		},
		HaltMempool:       minimallyValidConnectedUpgradeHaltMempoolEvidence(),
		StakingTimeQueues: &upgradeP0StakingQueueEvidence{},
	})
	require.ErrorContains(t, err, "unbond")
}

func TestValidateConnectedUpgradeCoverageInputRequiresSlashingJailEvidenceForNormalLane(t *testing.T) {
	t.Parallel()

	staking := validConnectedUpgradeStakingTimeQueueEvidence()
	err := validateConnectedUpgradeCoverageInput(connectedUpgradeCoverageInput{
		Scenario: upgradeRunScenario{
			Name:                "normal-empty-legacy-pnft",
			RunP0BoundaryMatrix: true,
		},
		HaltMempool:       minimallyValidConnectedUpgradeHaltMempoolEvidence(),
		StakingTimeQueues: &staking,
	})
	require.ErrorContains(t, err, "slashing jail")
}

func TestValidateConnectedUpgradeCoverageInputRejectsIncompleteSlashingJailEvidence(t *testing.T) {
	t.Parallel()

	staking := validConnectedUpgradeStakingTimeQueueEvidence()
	err := validateConnectedUpgradeCoverageInput(connectedUpgradeCoverageInput{
		Scenario: upgradeRunScenario{
			Name:                "normal-empty-legacy-pnft",
			RunP0BoundaryMatrix: true,
		},
		HaltMempool:       minimallyValidConnectedUpgradeHaltMempoolEvidence(),
		StakingTimeQueues: &staking,
		SlashingJail:      &upgradeP0SlashingEvidence{},
	})
	require.ErrorContains(t, err, "outage")
}

func TestValidateConnectedUpgradeCoverageInputRequiresLegacyAminoEvidenceForNormalLane(t *testing.T) {
	t.Parallel()

	staking := validConnectedUpgradeStakingTimeQueueEvidence()
	slashing := validConnectedUpgradeSlashingJailEvidence()
	err := validateConnectedUpgradeCoverageInput(connectedUpgradeCoverageInput{
		Scenario: upgradeRunScenario{
			Name:                "normal-empty-legacy-pnft",
			RunP0BoundaryMatrix: true,
		},
		HaltMempool:       minimallyValidConnectedUpgradeHaltMempoolEvidence(),
		StakingTimeQueues: &staking,
		SlashingJail:      &slashing,
	})
	require.ErrorContains(t, err, "legacy-amino")
}

func TestValidateConnectedUpgradeCoverageInputRejectsIncompleteLegacyAminoEvidence(t *testing.T) {
	t.Parallel()

	staking := validConnectedUpgradeStakingTimeQueueEvidence()
	slashing := validConnectedUpgradeSlashingJailEvidence()
	err := validateConnectedUpgradeCoverageInput(connectedUpgradeCoverageInput{
		Scenario: upgradeRunScenario{
			Name:                "normal-empty-legacy-pnft",
			RunP0BoundaryMatrix: true,
		},
		HaltMempool:       minimallyValidConnectedUpgradeHaltMempoolEvidence(),
		StakingTimeQueues: &staking,
		SlashingJail:      &slashing,
		LegacyAminoCustom: &upgradeV221LegacyAminoCustomTxsFixture{},
	})
	require.ErrorContains(t, err, "AOL")
}

func TestValidateConnectedUpgradeCoverageInputRequiresOldAminoSignMode(t *testing.T) {
	t.Parallel()

	staking := validConnectedUpgradeStakingTimeQueueEvidence()
	slashing := validConnectedUpgradeSlashingJailEvidence()
	err := validateConnectedUpgradeCoverageInput(connectedUpgradeCoverageInput{
		Scenario: upgradeRunScenario{
			Name:                "normal-empty-legacy-pnft",
			RunP0BoundaryMatrix: true,
		},
		HaltMempool:       minimallyValidConnectedUpgradeHaltMempoolEvidence(),
		StakingTimeQueues: &staking,
		SlashingJail:      &slashing,
		LegacyAminoCustom: &upgradeV221LegacyAminoCustomTxsFixture{
			AOL: upgradeV221LegacyAminoCustomTxFixture{Kind: upgradeV221LegacyAminoAOLCreateTopic},
			DID: upgradeV221LegacyAminoCustomTxFixture{Kind: upgradeV221LegacyAminoDIDUpdate},
		},
	})
	require.ErrorContains(t, err, upgradeV221LegacyAminoSignMode)
}

func TestValidateConnectedUpgradeCoverageInputRequiresExactLegacyAminoTamperRejection(t *testing.T) {
	t.Parallel()

	staking := validConnectedUpgradeStakingTimeQueueEvidence()
	slashing := validConnectedUpgradeSlashingJailEvidence()
	legacy := upgradeV221LegacyAminoCustomTxsFixture{
		AOL: upgradeV221LegacyAminoCustomTxFixture{
			Kind:            upgradeV221LegacyAminoAOLCreateTopic,
			SignMode:        upgradeV221LegacyAminoSignMode,
			TamperedCheckTx: &harness.TxResult{Height: "0", TxHash: "AOL-TAMPER", Codespace: "sdk", Code: 3},
		},
		DID: upgradeV221LegacyAminoCustomTxFixture{
			Kind:            upgradeV221LegacyAminoDIDUpdate,
			SignMode:        upgradeV221LegacyAminoSignMode,
			TamperedCheckTx: &harness.TxResult{Height: "0", TxHash: "DID-TAMPER", Codespace: "sdk", Code: 4},
		},
	}
	err := validateConnectedUpgradeCoverageInput(connectedUpgradeCoverageInput{
		Scenario: upgradeRunScenario{
			Name:                "normal-empty-legacy-pnft",
			RunP0BoundaryMatrix: true,
		},
		HaltMempool:       minimallyValidConnectedUpgradeHaltMempoolEvidence(),
		StakingTimeQueues: &staking,
		SlashingJail:      &slashing,
		LegacyAminoCustom: &legacy,
	})
	require.ErrorContains(t, err, "sdk/4")
}

func TestValidateConnectedUpgradeCoverageInputRequiresCommittedLegacyAminoMutations(t *testing.T) {
	t.Parallel()

	staking := validConnectedUpgradeStakingTimeQueueEvidence()
	slashing := validConnectedUpgradeSlashingJailEvidence()
	legacy := upgradeV221LegacyAminoCustomTxsFixture{
		AOL: upgradeV221LegacyAminoCustomTxFixture{
			Kind:            upgradeV221LegacyAminoAOLCreateTopic,
			SignMode:        upgradeV221LegacyAminoSignMode,
			TamperedCheckTx: &harness.TxResult{Height: "0", TxHash: "AOL-TAMPER", Codespace: "sdk", Code: 4},
		},
		DID: upgradeV221LegacyAminoCustomTxFixture{
			Kind:            upgradeV221LegacyAminoDIDUpdate,
			SignMode:        upgradeV221LegacyAminoSignMode,
			TamperedCheckTx: &harness.TxResult{Height: "0", TxHash: "DID-TAMPER", Codespace: "sdk", Code: 4},
		},
	}
	err := validateConnectedUpgradeCoverageInput(connectedUpgradeCoverageInput{
		Scenario: upgradeRunScenario{
			Name:                "normal-empty-legacy-pnft",
			RunP0BoundaryMatrix: true,
		},
		HaltMempool:       minimallyValidConnectedUpgradeHaltMempoolEvidence(),
		StakingTimeQueues: &staking,
		SlashingJail:      &slashing,
		LegacyAminoCustom: &legacy,
	})
	require.ErrorContains(t, err, "committed tx hash")
}

func TestValidateConnectedUpgradeCoverageInputRequiresExactHaltMempoolHashLineage(t *testing.T) {
	t.Parallel()

	staking := validConnectedUpgradeStakingTimeQueueEvidence()
	slashing := validConnectedUpgradeSlashingJailEvidence()
	legacy := minimallyValidConnectedUpgradeLegacyAminoEvidence()
	halt := upgradeHaltMempoolFixture{
		Transactions: []upgradeHaltMempoolSignedTx{
			{
				CheckTx:   &harness.TxResult{Height: "0", TxHash: "HALT-0", Code: 0},
				Committed: &harness.TxResult{Height: "101", TxHash: "HALT-0", Code: 0},
			},
			{
				CheckTx:   &harness.TxResult{Height: "0", TxHash: "HALT-1", Code: 0},
				Committed: &harness.TxResult{Height: "102", TxHash: "OTHER", Code: 0},
			},
		},
	}
	err := validateConnectedUpgradeCoverageInput(connectedUpgradeCoverageInput{
		Scenario: upgradeRunScenario{
			Name:                "normal-empty-legacy-pnft",
			RunP0BoundaryMatrix: true,
		},
		HaltMempool:       &halt,
		StakingTimeQueues: &staking,
		SlashingJail:      &slashing,
		LegacyAminoCustom: &legacy,
	})
	require.ErrorContains(t, err, "hash lineage")
}

func TestValidateConnectedUpgradeCoverageInputRequiresCompleteHaltMempoolReconciliation(t *testing.T) {
	t.Parallel()

	staking := validConnectedUpgradeStakingTimeQueueEvidence()
	slashing := validConnectedUpgradeSlashingJailEvidence()
	legacy := minimallyValidConnectedUpgradeLegacyAminoEvidence()
	halt := minimallyValidConnectedUpgradeHaltMempoolEvidence()
	halt.Reconciliation.CommittedPrefix = 1
	halt.Reconciliation.MissingSuffix = []int{1}
	err := validateConnectedUpgradeCoverageInput(connectedUpgradeCoverageInput{
		Scenario: upgradeRunScenario{
			Name:                "normal-empty-legacy-pnft",
			RunP0BoundaryMatrix: true,
		},
		HaltMempool:       halt,
		StakingTimeQueues: &staking,
		SlashingJail:      &slashing,
		LegacyAminoCustom: &legacy,
	})
	require.ErrorContains(t, err, "reconciliation")
}

func minimallyValidConnectedUpgradeLegacyAminoEvidence() upgradeV221LegacyAminoCustomTxsFixture {
	return upgradeV221LegacyAminoCustomTxsFixture{
		AOL: upgradeV221LegacyAminoCustomTxFixture{
			Kind:            upgradeV221LegacyAminoAOLCreateTopic,
			SignMode:        upgradeV221LegacyAminoSignMode,
			TamperedCheckTx: &harness.TxResult{Height: "0", TxHash: "AOL-TAMPER", Codespace: "sdk", Code: 4},
			TxHash:          "AOL-COMMITTED",
		},
		DID: upgradeV221LegacyAminoCustomTxFixture{
			Kind:            upgradeV221LegacyAminoDIDUpdate,
			SignMode:        upgradeV221LegacyAminoSignMode,
			TamperedCheckTx: &harness.TxResult{Height: "0", TxHash: "DID-TAMPER", Codespace: "sdk", Code: 4},
			TxHash:          "DID-COMMITTED",
		},
	}
}

func minimallyValidConnectedUpgradeHaltMempoolEvidence() *upgradeHaltMempoolFixture {
	return &upgradeHaltMempoolFixture{
		InitialSequence: 7,
		Transactions: []upgradeHaltMempoolSignedTx{
			{
				CheckTx:   &harness.TxResult{Height: "0", TxHash: "HALT-0", Code: 0},
				Committed: &harness.TxResult{Height: "101", TxHash: "HALT-0", Code: 0},
			},
			{
				CheckTx:   &harness.TxResult{Height: "0", TxHash: "HALT-1", Code: 0},
				Committed: &harness.TxResult{Height: "102", TxHash: "HALT-1", Code: 0},
			},
		},
		Reconciliation: upgradeHaltMempoolReconcilePlan{
			InitialSequence:  7,
			ObservedSequence: 9,
			TransactionCount: 2,
			CommittedPrefix:  2,
		},
	}
}

func validConnectedUpgradeSlashingJailEvidence() upgradeP0SlashingEvidence {
	jailedUntil := time.Date(2026, 8, 5, 12, 0, 45, 0, time.UTC)
	return upgradeP0SlashingEvidence{
		UpgradeHeight:        100,
		StoppedAt:            100,
		OutageBlocksObserved: upgradeP0SlashingMinimumMisses,
		MissedBlocksObserved: 7,
		Before: upgradeP0SlashingCheckpoint{
			Height:         99,
			RecordedAt:     jailedUntil.Add(-time.Minute),
			ValidatorPower: 100,
			Validator: upgradeValidatorState{
				OperatorAddress: "panaceavaloper1target",
				Tokens:          "10000",
			},
			SigningInfo: upgradeSigningInfoState{
				Address:             "panaceavalcons1target",
				IndexOffset:         90,
				MissedBlocksCounter: 1,
			},
		},
		Jailed: upgradeP0SlashingCheckpoint{
			Height:         100 + upgradeP0SlashingMinimumMisses,
			RecordedAt:     jailedUntil.Add(-30 * time.Second),
			ValidatorPower: 0,
			Validator: upgradeValidatorState{
				OperatorAddress: "panaceavaloper1target",
				Tokens:          "9900",
				Jailed:          true,
			},
			SigningInfo: upgradeSigningInfoState{
				Address:     "panaceavalcons1target",
				IndexOffset: 99,
				JailedUntil: jailedUntil.Format(time.RFC3339Nano),
			},
		},
		EarlyUnjail: harness.TxResult{
			Height: "109", TxHash: "EARLY", Codespace: "slashing", Code: 4,
		},
		UnjailTxHash: "UNJAIL",
		Unjailed: upgradeP0SlashingCheckpoint{
			Height:         140,
			RecordedAt:     jailedUntil.Add(time.Second),
			ValidatorPower: 0,
			Validator: upgradeValidatorState{
				OperatorAddress: "panaceavaloper1target",
				Tokens:          "9900",
			},
			SigningInfo: upgradeSigningInfoState{
				Address:     "panaceavalcons1target",
				IndexOffset: 99,
				JailedUntil: jailedUntil.Format(time.RFC3339Nano),
			},
		},
		Rejoined: upgradeP0SlashingCheckpoint{
			Height:         143,
			RecordedAt:     jailedUntil.Add(4 * time.Second),
			ValidatorPower: 99,
			Validator: upgradeValidatorState{
				OperatorAddress: "panaceavaloper1target",
				Tokens:          "9900",
			},
			SigningInfo: upgradeSigningInfoState{
				Address:     "panaceavalcons1target",
				IndexOffset: 102,
				JailedUntil: jailedUntil.Format(time.RFC3339Nano),
			},
		},
		SignedCommitHeight: 142,
		PostRestart: upgradeP0SlashingCheckpoint{
			Height:         150,
			RecordedAt:     jailedUntil.Add(time.Minute),
			ValidatorPower: 99,
			Validator: upgradeValidatorState{
				OperatorAddress: "panaceavaloper1target",
				Tokens:          "9900",
			},
			SigningInfo: upgradeSigningInfoState{
				Address:     "panaceavalcons1target",
				IndexOffset: 109,
				JailedUntil: jailedUntil.Format(time.RFC3339Nano),
			},
		},
	}
}

func validConnectedUpgradeStakingTimeQueueEvidence() upgradeP0StakingQueueEvidence {
	completion := time.Date(2026, 8, 5, 12, 3, 0, 0, time.UTC)
	queued := upgradeP0StakingQueueCheckpoint{
		Phase:                        "queued",
		Height:                       95,
		RecordedAt:                   completion.Add(-2 * time.Minute),
		BankBalance:                  "995",
		SourceDelegationBalance:      "700",
		DestinationDelegationBalance: "200",
		Pool:                         upgradeStakingPoolState{BondedTokens: "3900", NotBondedTokens: "100"},
		Unbonding: []upgradeP0QueueEntry{{
			CompletionTime: completion,
			InitialBalance: "100",
			Balance:        "100",
		}},
		Redelegation: []upgradeP0QueueEntry{{
			CompletionTime:    completion,
			InitialBalance:    "200",
			Balance:           "200",
			SharesDestination: "200.000000000000000000",
		}},
	}
	postUpgradePending := queued
	postUpgradePending.Phase = "post-upgrade-pending"
	postUpgradePending.Height = 104
	postUpgradePending.RecordedAt = completion.Add(-time.Minute)
	return upgradeP0StakingQueueEvidence{
		UnbondAmount:              "100",
		RedelegateAmount:          "200",
		FeePerTx:                  "5",
		UnbondWithdrawnReward:     "3",
		RedelegateWithdrawnReward: "2",
		Before: upgradeP0StakingQueueCheckpoint{
			Phase:                        "before",
			Height:                       90,
			RecordedAt:                   completion.Add(-3 * time.Minute),
			BankBalance:                  "1000",
			SourceDelegationBalance:      "1000",
			DestinationDelegationBalance: "0",
			Pool:                         upgradeStakingPoolState{BondedTokens: "4000", NotBondedTokens: "0"},
		},
		Queued:             queued,
		PostUpgradePending: postUpgradePending,
		Completed: upgradeP0StakingQueueCheckpoint{
			Phase:                        "completed",
			Height:                       190,
			RecordedAt:                   completion.Add(time.Second),
			BankBalance:                  "1095",
			SourceDelegationBalance:      "700",
			DestinationDelegationBalance: "200",
			Pool:                         upgradeStakingPoolState{BondedTokens: "3900", NotBondedTokens: "0"},
		},
		PostRestart: upgradeP0StakingQueueCheckpoint{
			Phase:                        "post-restart",
			Height:                       200,
			RecordedAt:                   completion.Add(time.Minute),
			BankBalance:                  "1095",
			SourceDelegationBalance:      "700",
			DestinationDelegationBalance: "200",
			Pool:                         upgradeStakingPoolState{BondedTokens: "3900", NotBondedTokens: "0"},
		},
	}
}
