package e2e_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeV221LegacyAminoCustomTxFixedContracts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		raw          string
		wantKind     upgradeV221LegacyAminoCustomTxKind
		wantSigner   string
		wantObject   string
		wantSequence uint64
	}{
		{
			name: "AOL create-topic",
			raw: `{
				"body":{"messages":[{"@type":"/panacea.aol.v2.MsgCreateTopicRequest","topic_name":"legacy-topic","description":"prepared by v2.2.1","owner_address":"panacea1aol"}]},
				"auth_info":{"signer_infos":[{"mode_info":{"single":{"mode":"SIGN_MODE_LEGACY_AMINO_JSON"}},"sequence":"7"}],"fee":{"amount":[{"denom":"umed","amount":"5000000"}],"gas_limit":"500000"}},
				"signatures":["AQID"]
			}`,
			wantKind:     upgradeV221LegacyAminoAOLCreateTopic,
			wantSigner:   "panacea1aol",
			wantObject:   "panacea1aol/legacy-topic",
			wantSequence: 7,
		},
		{
			name: "DID update",
			raw: `{
				"body":{"messages":[{"@type":"/panacea.did.v2.MsgUpdateDIDRequest","did":"did:panacea:fixture","document":{"contexts":["https://www.w3.org/ns/did/v1"],"id":"did:panacea:fixture","verification_methods":[],"authentications":[],"services":[{"id":"did:panacea:fixture#service","type":"LinkedDomains","service_endpoint":"https://example.invalid"}]},"verification_method_id":"did:panacea:fixture#key1","signature":"BAUG","from_address":"panacea1did"}]},
				"auth_info":{"signer_infos":[{"mode_info":{"single":{"mode":"SIGN_MODE_LEGACY_AMINO_JSON"}},"sequence":"11"}],"fee":{"amount":[{"denom":"umed","amount":"5000000"}],"gas_limit":"1000000"}},
				"signatures":["BwgJ"]
			}`,
			wantKind:     upgradeV221LegacyAminoDIDUpdate,
			wantSigner:   "panacea1did",
			wantObject:   "did:panacea:fixture",
			wantSequence: 11,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := decodeV221LegacyAminoCustomTx([]byte(test.raw))
			require.NoError(t, err)
			require.Equal(t, test.wantKind, decoded.Kind)
			require.Equal(t, test.wantSigner, decoded.SignerAddress)
			require.Equal(t, test.wantObject, decoded.StateObjectID)
			require.Equal(t, test.wantSequence, decoded.Sequence)
			require.Equal(t, upgradeV221LegacyAminoFee, decoded.Fee)
			require.NotEmpty(t, decoded.Signature)
		})
	}
}

func TestDecodeV221LegacyAminoCustomTxRejectsWrongModeAndMalformedPayload(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"body":{"messages":[{"@type":"/panacea.aol.v2.MsgCreateTopicRequest","topic_name":"legacy-topic","description":"prepared by v2.2.1","owner_address":"panacea1aol"}]},
		"auth_info":{"signer_infos":[{"mode_info":{"single":{"mode":"SIGN_MODE_DIRECT"}},"sequence":"7"}],"fee":{"amount":[{"denom":"umed","amount":"5000000"}],"gas_limit":"500000"}},
		"signatures":["AQID"]
	}`)
	_, err := decodeV221LegacyAminoCustomTx(raw)
	require.ErrorContains(t, err, "SIGN_MODE_LEGACY_AMINO_JSON")

	var document map[string]any
	require.NoError(t, json.Unmarshal(raw, &document))
	signer := document["auth_info"].(map[string]any)["signer_infos"].([]any)[0].(map[string]any)
	signer["mode_info"].(map[string]any)["single"].(map[string]any)["mode"] = upgradeV221LegacyAminoSignMode
	body := document["body"].(map[string]any)
	message := body["messages"].([]any)[0].(map[string]any)
	message["owner_address"] = ""
	changed, err := json.Marshal(document)
	require.NoError(t, err)
	_, err = decodeV221LegacyAminoCustomTx(changed)
	require.ErrorContains(t, err, "AOL")
}

func TestTamperV221LegacyAminoSignedTxChangesOnlyMessage(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"body":{"messages":[{"@type":"/panacea.aol.v2.MsgCreateTopicRequest","topic_name":"legacy-topic","description":"prepared by v2.2.1","owner_address":"panacea1aol"}],"memo":"fixed"},
		"auth_info":{"signer_infos":[{"mode_info":{"single":{"mode":"SIGN_MODE_LEGACY_AMINO_JSON"}},"sequence":"7"}],"fee":{"amount":[{"denom":"umed","amount":"5000000"}],"gas_limit":"500000"}},
		"signatures":["AQID"]
	}`)

	tampered, err := tamperV221LegacyAminoSignedTx(raw)
	require.NoError(t, err)
	before := decodeJSONDocument(t, raw)
	after := decodeJSONDocument(t, tampered)
	require.Equal(t, before["auth_info"], after["auth_info"])
	require.Equal(t, before["signatures"], after["signatures"])
	require.NotEqual(t, before["body"], after["body"])

	decoded, err := decodeV221LegacyAminoCustomTx(tampered)
	require.NoError(t, err)
	require.Equal(t, "legacy-topic-tampered", decoded.TopicName)
}

func decodeJSONDocument(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var document map[string]any
	require.NoError(t, json.Unmarshal(raw, &document))
	return document
}
