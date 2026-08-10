package cli

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/app/params"
	aoltypes "github.com/medibloc/panacea-core/v2/x/aol/types"
)

func TestSignWithMultiSignersUsesTxConfigHandlerMap(t *testing.T) {
	encodingConfig := params.MakeEncodingConfig(
		params.WithCustomGetSigners(aoltypes.CustomGetSigners()...),
	)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	aoltypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	require.NoError(t, encodingConfig.InterfaceRegistry.SigningContext().Validate())

	kr := keyring.NewInMemory(encodingConfig.Codec)
	feePayerRecord, _, err := kr.NewMnemonic(
		"fee-payer",
		keyring.English,
		sdk.FullFundraiserPath,
		keyring.DefaultBIP39Passphrase,
		hd.Secp256k1,
	)
	require.NoError(t, err)
	writerRecord, _, err := kr.NewMnemonic(
		"writer",
		keyring.English,
		sdk.FullFundraiserPath,
		keyring.DefaultBIP39Passphrase,
		hd.Secp256k1,
	)
	require.NoError(t, err)

	feePayerAddress, err := feePayerRecord.GetAddress()
	require.NoError(t, err)
	writerAddress, err := writerRecord.GetAddress()
	require.NoError(t, err)

	msg := aoltypes.NewMsgAddRecordRequest(
		"topic",
		[]byte("key"),
		[]byte("value"),
		writerAddress.String(),
		writerAddress.String(),
		feePayerAddress.String(),
	)
	require.Equal(t, []sdk.AccAddress{feePayerAddress, writerAddress}, msg.GetSigners())

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	require.NoError(t, txBuilder.SetMsgs(msg))

	accounts := []accountInKeyring{
		{
			keyInfo:       feePayerRecord,
			address:       feePayerAddress,
			accountNumber: 11,
			sequence:      12,
		},
		{
			keyInfo:       writerRecord,
			address:       writerAddress,
			accountNumber: 21,
			sequence:      22,
		},
	}
	clientCtx := client.Context{}.
		WithTxConfig(encodingConfig.TxConfig).
		WithKeyring(kr).
		WithChainID("panacea-test")
	signMode := signing.SignMode_SIGN_MODE_DIRECT

	require.NoError(t, gatherAllSignerInfos(txBuilder, signMode, accounts))
	require.NoError(t, signWithMultiSigners(clientCtx, txBuilder, signMode, accounts))

	sigTx, ok := txBuilder.GetTx().(authsigning.SigVerifiableTx)
	require.True(t, ok)
	signatures, err := sigTx.GetSignaturesV2()
	require.NoError(t, err)
	require.Len(t, signatures, len(accounts))

	for i, account := range accounts {
		publicKey, err := account.keyInfo.GetPubKey()
		require.NoError(t, err)
		require.True(t, publicKey.Equals(signatures[i].PubKey))
		require.Equal(t, account.sequence, signatures[i].Sequence)

		signatureData, ok := signatures[i].Data.(*signing.SingleSignatureData)
		require.True(t, ok)
		require.Equal(t, signMode, signatureData.SignMode)
		require.NotEmpty(t, signatureData.Signature)

		signBytes, err := authsigning.GetSignBytesAdapter(
			context.Background(),
			encodingConfig.TxConfig.SignModeHandler(),
			signMode,
			authsigning.SignerData{
				Address:       account.address.String(),
				ChainID:       clientCtx.ChainID,
				AccountNumber: account.accountNumber,
				Sequence:      account.sequence,
				PubKey:        publicKey,
			},
			txBuilder.GetTx(),
		)
		require.NoError(t, err)
		require.True(t, publicKey.VerifySignature(signBytes, signatureData.Signature))
	}
}
