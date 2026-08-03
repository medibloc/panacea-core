package app

import (
	"context"
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	aoltypes "github.com/medibloc/panacea-core/v2/x/aol/types"
	"github.com/stretchr/testify/require"
)

func TestMakeEncodingConfigEnablesV221AOLDIDLegacyAminoCompatibility(t *testing.T) {
	SetConfig()
	encodingConfig := MakeEncodingConfig()
	address := sdk.AccAddress(make([]byte, 20)).String()
	msg := &aoltypes.MsgCreateTopicRequest{
		TopicName:    "topic",
		Description:  "desc",
		OwnerAddress: address,
	}

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin("umed", 10)))
	txBuilder.SetGasLimit(200000)
	require.NoError(t, txBuilder.SetMsgs(msg))

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
	require.Contains(t, signDoc.Messages[0], "topic_name")
}
