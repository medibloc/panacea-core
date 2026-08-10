package harness

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

var validImportedGenesis = []byte(`{
  "genesis_time":"2026-08-04T00:00:00Z",
  "chain_id":"panacea-source",
  "initial_height":"42",
  "app_state":{"bank":{"balances":[{"address":"panacea1owner","coins":[{"denom":"umed","amount":"7"}]}]},"nft":{"class_records":[{"id":"class-1"}]}},
  "consensus":{"validators":[{"address":"AABB","pub_key":{"type":"tendermint/PubKeyEd25519","value":"cHVibGljLWtleQ=="},"power":"1","name":"validator"}]}
}`)

var validImportedValidatorKey = []byte(`{
  "address":"AABB",
  "pub_key":{"type":"tendermint/PubKeyEd25519","value":"cHVibGljLWtleQ=="},
  "priv_key":{"type":"tendermint/PrivKeyEd25519","value":"dGVzdC1wcml2YXRlLWtleQ=="}
}`)

func TestValidateImportedValidatorKeyRequiresExactConsensusIdentity(t *testing.T) {
	require.NoError(t, validateImportedValidatorKey(validImportedGenesis, validImportedValidatorKey))

	mismatchedKey := []byte(`{
  "address":"AABB",
  "pub_key":{"type":"tendermint/PubKeyEd25519","value":"bWlzbWF0Y2hlZA=="},
  "priv_key":{"type":"tendermint/PrivKeyEd25519","value":"dGVzdC1wcml2YXRlLWtleQ=="}
}`)
	require.ErrorContains(t, validateImportedValidatorKey(validImportedGenesis, mismatchedKey), "does not match")

	missingPublicIdentityGenesis := []byte(`{
  "consensus":{"validators":[{"address":"","pub_key":{"type":"","value":""},"power":"1"}]}
}`)
	missingPublicIdentityKey := []byte(`{
  "address":"",
  "pub_key":{"type":"","value":""},
  "priv_key":{"type":"tendermint/PrivKeyEd25519","value":"dGVzdC1wcml2YXRlLWtleQ=="}
}`)
	require.ErrorContains(
		t,
		validateImportedValidatorKey(missingPublicIdentityGenesis, missingPublicIdentityKey),
		"public identity",
	)
}

func TestConfigureExportBootstrapPreservesChainIdentityAndSkipsGentx(t *testing.T) {
	spec, err := NewPanaceaChainSpec("import-hook", CurrentImage(), Topology{Validators: 1, FullNodes: 1})
	require.NoError(t, err)
	genesis := append([]byte(nil), validImportedGenesis...)
	validatorKey := append([]byte(nil), validImportedValidatorKey...)

	err = configureExportBootstrap(context.Background(), spec, Config{
		NumValidators:   1,
		NumFullNodes:    1,
		exportedGenesis: genesis,
		validatorKey:    validatorKey,
	})
	require.NoError(t, err)
	require.Equal(t, "panacea-source", spec.ChainConfig.ChainID)
	require.True(t, spec.ChainConfig.SkipGenTx)
	require.NotNil(t, spec.ChainConfig.ModifyGenesis)
	require.NotNil(t, spec.ChainConfig.PreGenesis)
}

func TestConfigureExportBootstrapRestoresExactExportAfterUpstreamGenesisMutation(t *testing.T) {
	spec, err := NewPanaceaChainSpec("import-hook", CurrentImage(), Topology{Validators: 1, FullNodes: 1})
	require.NoError(t, err)
	exportedGenesis := bytes.Replace(
		validImportedGenesis,
		[]byte(`"app_state":{`),
		[]byte(`"app_state":{"distribution":{"delegator_starting_infos":[{"starting_info":{"stake":"7.000000000000000000"}}]},`),
		1,
	)
	require.Contains(t, string(exportedGenesis), `"stake"`)

	err = configureExportBootstrap(context.Background(), spec, Config{
		NumValidators:   1,
		NumFullNodes:    1,
		exportedGenesis: exportedGenesis,
		validatorKey:    validImportedValidatorKey,
	})
	require.NoError(t, err)
	require.NotNil(t, spec.ChainConfig.ModifyGenesis)

	mutatedByUpstream := bytes.ReplaceAll(exportedGenesis, []byte(`"stake"`), []byte(`"umed"`))
	mutatedByUpstream = append(mutatedByUpstream, []byte(" injected-faucet")...)
	restored, err := spec.ChainConfig.ModifyGenesis(spec.ChainConfig, mutatedByUpstream)
	require.NoError(t, err)
	require.Equal(t, exportedGenesis, restored)
}

func TestConfigureExportBootstrapRejectsUnsafeTopology(t *testing.T) {
	spec, err := NewPanaceaChainSpec("import-hook", CurrentImage(), Topology{Validators: 1, FullNodes: 0})
	require.NoError(t, err)

	err = configureExportBootstrap(context.Background(), spec, Config{
		NumValidators:   1,
		NumFullNodes:    0,
		exportedGenesis: validImportedGenesis,
		validatorKey:    validImportedValidatorKey,
	})
	require.ErrorContains(t, err, "full node")
}
