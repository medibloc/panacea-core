package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestProposalIDFromCommittedTxAcceptsTypedAndLegacyEvents(t *testing.T) {
	for _, eventType := range []string{"cosmos.gov.v1.EventSubmitProposal", "submit_proposal"} {
		result := &harness.TxResult{Events: []harness.TxEvent{{
			Type: eventType,
			Attributes: []harness.TxEventAttribute{{
				Key:   "proposal_id",
				Value: `"17"`,
			}},
		}}}
		proposalID, err := proposalIDFromCommittedTx(result)
		require.NoError(t, err)
		require.Equal(t, uint64(17), proposalID)
	}
}

func TestProposalIDFromCommittedTxRejectsMissingOrAmbiguousID(t *testing.T) {
	_, err := proposalIDFromCommittedTx(nil)
	require.Error(t, err)

	_, err = proposalIDFromCommittedTx(&harness.TxResult{})
	require.Error(t, err)

	_, err = proposalIDFromCommittedTx(&harness.TxResult{Events: []harness.TxEvent{
		{Attributes: []harness.TxEventAttribute{{Key: "proposal_id", Value: "4"}}},
		{Attributes: []harness.TxEventAttribute{{Key: "proposal_id", Value: "5"}}},
	}})
	require.Error(t, err)
}

func TestDIDFromKeyStoreMetadata(t *testing.T) {
	metadata := []byte(`{"address":"did:panacea:mainnet:abc123#key1"}` + "\n")
	did, err := didFromKeyStoreMetadata(metadata)
	require.NoError(t, err)
	require.Equal(t, "did:panacea:mainnet:abc123", did)
}

func TestDIDFromKeyStoreMetadataRejectsSecretBearingDocument(t *testing.T) {
	_, err := didFromKeyStoreMetadata([]byte(`{"address":"did:panacea:mainnet:abc#key1","crypto":{"ciphertext":"secret"}}`))
	require.Error(t, err)
}

func TestDecodeUpgradeInfoAcceptsStringAndNumericHeight(t *testing.T) {
	for _, input := range []string{
		`{"name":"v2.3.0","height":"42"}`,
		`{"name":"v2.3.0","height":42}`,
	} {
		name, height, err := decodeUpgradeInfo([]byte(input))
		require.NoError(t, err)
		require.Equal(t, "v2.3.0", name)
		require.EqualValues(t, 42, height)
	}
	_, _, err := decodeUpgradeInfo([]byte(`{"name":"","height":42}`))
	require.ErrorContains(t, err, "plan name")
	_, _, err = decodeUpgradeInfo([]byte(`{"name":"v2.3.0","height":"bad"}`))
	require.ErrorContains(t, err, "invalid height")
}

func TestDecodeUpgradeBinaryIdentity(t *testing.T) {
	identity, err := decodeUpgradeBinaryIdentity([]byte(`build_deps:
- example.org/dependency@v1.0.0
commit: abc123
cosmos_sdk_version: v0.50.15
name: panacea-core
version: 2.3.0
`))
	require.NoError(t, err)
	require.Equal(t, upgradeBinaryIdentity{
		Name:             "panacea-core",
		Version:          "2.3.0",
		Commit:           "abc123",
		CosmosSDKVersion: "v0.50.15",
	}, identity)

	_, err = decodeUpgradeBinaryIdentity([]byte("name: panacea-core\n"))
	require.ErrorContains(t, err, "incomplete binary identity")
}

func TestValidateCurrentUpgradeBinaryIdentityRequiresV230(t *testing.T) {
	valid := upgradeBinaryIdentity{
		Name:             "panacea-core",
		Version:          "2.3.0",
		Commit:           "current-commit",
		CosmosSDKVersion: "v0.50.15",
	}
	require.NoError(t, validateCurrentUpgradeBinaryIdentity(valid))

	wrongVersion := valid
	wrongVersion.Version = "2.2.1-237-g4cfc858"
	require.ErrorContains(t, validateCurrentUpgradeBinaryIdentity(wrongVersion), "2.3.0")

	oldCommit := valid
	oldCommit.Commit = upgradeV221Commit
	require.ErrorContains(t, validateCurrentUpgradeBinaryIdentity(oldCommit), "old binary commit")
}

func TestValidateExpectedCurrentUpgradeBinaryIdentityRequiresWorktreeCommit(t *testing.T) {
	identity := upgradeBinaryIdentity{
		Name:             "panacea-core",
		Version:          "2.3.0",
		Commit:           "current-commit",
		CosmosSDKVersion: "v0.50.15",
	}

	require.ErrorContains(
		t,
		validateExpectedCurrentUpgradeBinaryIdentity(identity, "2.3.0", ""),
		"PANACEA_E2E_CURRENT_COMMIT",
	)
	require.ErrorContains(
		t,
		validateExpectedCurrentUpgradeBinaryIdentity(identity, "", "current-commit"),
		"PANACEA_E2E_CURRENT_BINARY_VERSION",
	)
	require.NoError(t, validateExpectedCurrentUpgradeBinaryIdentity(identity, "2.3.0", "current-commit"))
	require.ErrorContains(
		t,
		validateExpectedCurrentUpgradeBinaryIdentity(identity, "2.3.0", "different-commit"),
		"current-commit",
	)
}
