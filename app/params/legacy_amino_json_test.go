package params

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	aoltypes "github.com/medibloc/panacea-core/v2/x/aol/types"
	didtypes "github.com/medibloc/panacea-core/v2/x/did/types"
	"github.com/stretchr/testify/require"
)

func TestLegacyAminoJSONCompatibilityIsLimitedToTopLevelSelectedMessages(t *testing.T) {
	configureTestBech32()

	standardConfig := newLegacyAminoJSONTestConfig(t, false)
	compatConfig := newLegacyAminoJSONTestConfig(t, true)
	addr1 := sdk.AccAddress(repeatedBytes(1))
	addr2 := sdk.AccAddress(repeatedBytes(2))
	signer := newSignerCase(addr1.String(), 0)
	aolMsg := &aoltypes.MsgCreateTopicRequest{
		TopicName:    "topic",
		Description:  "desc",
		OwnerAddress: addr1.String(),
	}
	bankMsg := &banktypes.MsgSend{
		FromAddress: addr1.String(),
		ToAddress:   addr2.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("umed", 10)),
	}

	t.Run("standard top-level message is unchanged", func(t *testing.T) {
		standard := signBytes(t, standardConfig, signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, []sdk.Msg{bankMsg}, signer)
		compatible := signBytes(t, compatConfig, signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, []sdk.Msg{bankMsg}, signer)
		require.Equal(t, standard, compatible)
	})

	t.Run("nested selected message is unchanged", func(t *testing.T) {
		msgExec := authz.NewMsgExec(addr1, []sdk.Msg{aolMsg})
		standard := signBytes(t, standardConfig, signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, []sdk.Msg{&msgExec}, signer)
		compatible := signBytes(t, compatConfig, signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, []sdk.Msg{&msgExec}, signer)
		require.Equal(t, standard, compatible)
	})

	t.Run("direct signing is unchanged", func(t *testing.T) {
		standard := signBytes(t, standardConfig, signing.SignMode_SIGN_MODE_DIRECT, []sdk.Msg{aolMsg}, signer)
		compatible := signBytes(t, compatConfig, signing.SignMode_SIGN_MODE_DIRECT, []sdk.Msg{aolMsg}, signer)
		require.Equal(t, standard, compatible)
	})

	t.Run("only selected element changes in a mixed transaction", func(t *testing.T) {
		standard := signBytes(t, standardConfig, signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, []sdk.Msg{aolMsg, bankMsg}, signer)
		compatible := signBytes(t, compatConfig, signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, []sdk.Msg{aolMsg, bankMsg}, signer)

		standardMessages := signDocMessages(t, standard)
		compatibleMessages := signDocMessages(t, compatible)
		require.Len(t, compatibleMessages, 2)
		require.NotEqual(t, standardMessages[0], compatibleMessages[0])
		require.Equal(t, standardMessages[1], compatibleMessages[1])

		var bareMessage map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(compatibleMessages[0], &bareMessage))
		require.NotContains(t, bareMessage, "type")
		require.Contains(t, bareMessage, "topic_name")
	})
}

func TestBareAminoJSONValueRejectsMalformedEnvelopes(t *testing.T) {
	testCases := []struct {
		name     string
		envelope string
	}{
		{name: "invalid JSON", envelope: "{"},
		{name: "missing type", envelope: `{"value":{}}`},
		{name: "empty type", envelope: `{"type":"","value":{}}`},
		{name: "missing value", envelope: `{"type":"aol/CreateTopic"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := bareAminoJSONValue(json.RawMessage(tc.envelope))
			require.Error(t, err)
		})
	}
}

func TestDIDLegacyAminoJSONCompatibilityPreservesComplexDocumentEncoding(t *testing.T) {
	configureTestBech32()
	encodingConfig := newLegacyAminoJSONTestConfig(t, true)
	address := sdk.AccAddress(repeatedBytes(1)).String()
	did, verificationMethodID, doc := complexDIDDocument(t)
	testCases := []struct {
		name string
		msg  legacySignBytesMsg
	}{
		{name: "create", msg: &didtypes.MsgCreateDIDRequest{
			Did:                  did,
			Document:             doc,
			VerificationMethodId: verificationMethodID,
			Signature:            []byte("signature"),
			FromAddress:          address,
		}},
		{name: "update", msg: &didtypes.MsgUpdateDIDRequest{
			Did:                  did,
			Document:             doc,
			VerificationMethodId: verificationMethodID,
			Signature:            []byte("signature"),
			FromAddress:          address,
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := signBytes(
				t,
				encodingConfig,
				signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
				[]sdk.Msg{tc.msg},
				newSignerCase(address, 0),
			)
			messages := signDocMessages(t, actual)
			require.Len(t, messages, 1)
			require.Equal(t, string(tc.msg.GetSignBytes()), string(messages[0]))
		})
	}
}

type legacySignBytesMsg interface {
	sdk.Msg
	GetSignBytes() []byte
}

func newLegacyAminoJSONTestConfig(t *testing.T, compatible bool) EncodingConfig {
	t.Helper()

	options := []EncodingConfigOption{
		WithCustomGetSigners(aoltypes.CustomGetSigners()...),
		WithAminoJSONEncoderModifiers(didtypes.AminoJSONEncoderModifiers()...),
	}
	if compatible {
		options = append(options, WithV221LegacyAminoJSONCompatibility(legacyAminoJSONBareMessageTypeURLs()...))
	}

	encodingConfig := MakeEncodingConfig(options...)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	authz.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	banktypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	aoltypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	didtypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	require.NoError(t, encodingConfig.InterfaceRegistry.SigningContext().Validate())
	return encodingConfig
}

func signDocMessages(t *testing.T, signBytes []byte) []json.RawMessage {
	t.Helper()

	var signDoc struct {
		Messages []json.RawMessage `json:"msgs"`
	}
	require.NoError(t, json.Unmarshal(signBytes, &signDoc))
	return signDoc.Messages
}
