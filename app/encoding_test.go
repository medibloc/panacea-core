package app

import (
	"context"
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	aoltypes "github.com/medibloc/panacea-core/v2/x/aol/types"
	didtypes "github.com/medibloc/panacea-core/v2/x/did/types"
	"github.com/stretchr/testify/require"
)

func TestMakeEncodingConfigEnablesV221AOLDIDLegacyAminoCompatibility(t *testing.T) {
	SetConfig()
	encodingConfig := MakeEncodingConfig()
	address := sdk.AccAddress(make([]byte, 20)).String()
	testCases := []struct {
		name          string
		msg           sdk.Msg
		expectedField string
	}{
		{
			name:          "aol/CreateTopic",
			msg:           &aoltypes.MsgCreateTopicRequest{TopicName: "topic", OwnerAddress: address},
			expectedField: "topic_name",
		},
		{
			name:          "aol/AddWriter",
			msg:           &aoltypes.MsgAddWriterRequest{TopicName: "topic", WriterAddress: address, OwnerAddress: address},
			expectedField: "writer_address",
		},
		{
			name:          "aol/DeleteWriter",
			msg:           &aoltypes.MsgDeleteWriterRequest{TopicName: "topic", WriterAddress: address, OwnerAddress: address},
			expectedField: "writer_address",
		},
		{
			name:          "aol/AddRecord",
			msg:           &aoltypes.MsgAddRecordRequest{TopicName: "topic", WriterAddress: address, OwnerAddress: address},
			expectedField: "writer_address",
		},
		{
			name:          "did/CreateDID",
			msg:           &didtypes.MsgCreateDIDRequest{Did: "did:panacea:create", FromAddress: address},
			expectedField: "did",
		},
		{
			name:          "did/UpdateDID",
			msg:           &didtypes.MsgUpdateDIDRequest{Did: "did:panacea:update", FromAddress: address},
			expectedField: "did",
		},
		{
			name:          "did/DeactivateDID",
			msg:           &didtypes.MsgDeactivateDIDRequest{Did: "did:panacea:deactivate", FromAddress: address},
			expectedField: "did",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			txBuilder := encodingConfig.TxConfig.NewTxBuilder()
			txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin("umed", 10)))
			txBuilder.SetGasLimit(200000)
			require.NoError(t, txBuilder.SetMsgs(tc.msg))

			signBytes, err := authsigning.GetSignBytesAdapter(
				context.Background(),
				encodingConfig.TxConfig.SignModeHandler(),
				signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
				authsigning.SignerData{Address: address, ChainID: "test-chain"},
				txBuilder.GetTx(),
			)
			require.NoError(t, err)

			var signDoc struct {
				Messages []map[string]json.RawMessage `json:"msgs"`
			}
			require.NoError(t, json.Unmarshal(signBytes, &signDoc))
			require.Len(t, signDoc.Messages, 1)
			require.NotContains(t, signDoc.Messages[0], "type")
			require.NotContains(t, signDoc.Messages[0], "value")
			require.Contains(t, signDoc.Messages[0], tc.expectedField)
		})
	}
}
