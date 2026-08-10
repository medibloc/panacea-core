package harness

import (
	"testing"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	txsigning "cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
)

const rawTxTestMnemonic = "alcohol woman abuse must during monitor noble actual mixed trade anger aisle"

func TestBuildSignedRawTxRejectsWalletWithoutMnemonic(t *testing.T) {
	t.Parallel()

	_, err := buildSignedRawTx(rawTxBuildRequest{
		Signer:        stubRawWallet{address: []byte{1, 2, 3}},
		Message:       RawProtoMessage{TypeURL: "/example.Msg", Value: []byte{1}},
		ChainID:       "panacea-test",
		CoinType:      371,
		Denom:         "umed",
		GasLimit:      500_000,
		FeeAmount:     sdkmath.NewInt(2_500_000),
		AccountNumber: 1,
		Sequence:      2,
	})
	require.ErrorContains(t, err, "mnemonic")
}

func TestBuildSignedRawTxPreservesVerbatimMessage(t *testing.T) {
	t.Parallel()

	path := hd.CreateHDPath(371, 0, 0).String()
	derived, err := hd.Secp256k1.Derive()(rawTxTestMnemonic, "", path)
	require.NoError(t, err)
	privateKey := hd.Secp256k1.Generate()(derived)

	message := RawProtoMessage{
		TypeURL: "/panacea.nft.v1.MsgMintRequest",
		Value:   []byte{0x0a, 0x01, 'x', 0x3a, 0x02, 0xff, 0x00},
	}
	encoded, err := buildSignedRawTx(rawTxBuildRequest{
		Signer: stubRawWallet{
			mnemonic: rawTxTestMnemonic,
			address:  privateKey.PubKey().Address(),
		},
		Message:       message,
		ChainID:       "panacea-test",
		CoinType:      371,
		Denom:         "umed",
		GasLimit:      500_000,
		FeeAmount:     sdkmath.NewInt(2_500_000),
		AccountNumber: 7,
		Sequence:      11,
	})
	require.NoError(t, err)

	var raw txtypes.TxRaw
	require.NoError(t, proto.Unmarshal(encoded, &raw))
	require.Len(t, raw.Signatures, 1)

	var body txtypes.TxBody
	require.NoError(t, proto.Unmarshal(raw.BodyBytes, &body))
	require.Len(t, body.Messages, 1)
	require.Equal(t, message.TypeURL, body.Messages[0].TypeUrl)
	require.Equal(t, message.Value, body.Messages[0].Value)

	var authInfo txtypes.AuthInfo
	require.NoError(t, proto.Unmarshal(raw.AuthInfoBytes, &authInfo))
	require.Len(t, authInfo.SignerInfos, 1)
	require.EqualValues(t, 11, authInfo.SignerInfos[0].Sequence)
	require.Equal(t, signing.SignMode_SIGN_MODE_DIRECT, authInfo.SignerInfos[0].ModeInfo.GetSingle().Mode)
	require.EqualValues(t, 500_000, authInfo.Fee.GasLimit)
	require.Equal(t, "2500000umed", authInfo.Fee.Amount.String())

	signDoc := &txtypes.SignDoc{
		BodyBytes:     raw.BodyBytes,
		AuthInfoBytes: raw.AuthInfoBytes,
		ChainId:       "panacea-test",
		AccountNumber: 7,
	}
	signBytes, err := proto.Marshal(signDoc)
	require.NoError(t, err)
	require.True(t, privateKey.PubKey().VerifySignature(signBytes, raw.Signatures[0]))
}

func TestBuildSignedRawTxSupportsDistinctFeePayer(t *testing.T) {
	t.Parallel()

	path := hd.CreateHDPath(371, 0, 0).String()
	derived, err := hd.Secp256k1.Derive()(rawTxTestMnemonic, "", path)
	require.NoError(t, err)
	privateKey := hd.Secp256k1.Generate()(derived)
	wallet := stubRawWallet{
		mnemonic: rawTxTestMnemonic,
		address:  privateKey.PubKey().Address(),
	}

	encoded, err := buildSignedRawTx(rawTxBuildRequest{
		Signer:                wallet,
		Message:               RawProtoMessage{TypeURL: "/example.Msg", Value: []byte{1}},
		ChainID:               "panacea-test",
		CoinType:              371,
		Denom:                 "umed",
		GasLimit:              500_000,
		FeeAmount:             sdkmath.NewInt(2_500_000),
		AccountNumber:         7,
		Sequence:              11,
		FeePayer:              wallet,
		FeePayerAccountNumber: 13,
		FeePayerSequence:      17,
	})
	require.NoError(t, err)

	var raw txtypes.TxRaw
	require.NoError(t, proto.Unmarshal(encoded, &raw))
	require.Len(t, raw.Signatures, 2)

	var authInfo txtypes.AuthInfo
	require.NoError(t, proto.Unmarshal(raw.AuthInfoBytes, &authInfo))
	require.Len(t, authInfo.SignerInfos, 2)
	require.EqualValues(t, 11, authInfo.SignerInfos[0].Sequence)
	require.EqualValues(t, 17, authInfo.SignerInfos[1].Sequence)
	require.Equal(t, wallet.FormattedAddress(), authInfo.Fee.Payer)

	for i, accountNumber := range []uint64{7, 13} {
		signBytes, signErr := proto.Marshal(&txtypes.SignDoc{
			BodyBytes:     raw.BodyBytes,
			AuthInfoBytes: raw.AuthInfoBytes,
			ChainId:       "panacea-test",
			AccountNumber: accountNumber,
		})
		require.NoError(t, signErr)
		require.True(t, privateKey.PubKey().VerifySignature(signBytes, raw.Signatures[i]))
	}
}

func TestMalformedSignerFieldsHaveStableUndefinedABCIClassification(t *testing.T) {
	t.Parallel()

	path := hd.CreateHDPath(371, 0, 0).String()
	derived, err := hd.Secp256k1.Derive()(rawTxTestMnemonic, "", path)
	require.NoError(t, err)
	privateKey := hd.Secp256k1.Generate()(derived)
	wallet := stubRawWallet{
		mnemonic: rawTxTestMnemonic,
		address:  privateKey.PubKey().Address(),
	}

	message := &banktypes.MsgSend{FromAddress: "not-an-address"}
	messageValue, err := proto.Marshal(message)
	require.NoError(t, err)
	encoded, err := buildSignedRawTx(rawTxBuildRequest{
		Signer: wallet,
		Message: RawProtoMessage{
			TypeURL: sdk.MsgTypeURL(message),
			Value:   messageValue,
		},
		ChainID:       "panacea-test",
		CoinType:      371,
		Denom:         "umed",
		GasLimit:      500_000,
		FeeAmount:     sdkmath.NewInt(2_500_000),
		AccountNumber: 1,
		Sequence:      2,
	})
	require.NoError(t, err)

	registry, err := cdctypes.NewInterfaceRegistryWithOptions(cdctypes.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: txsigning.Options{
			AddressCodec:          authcodec.NewBech32Codec("panacea"),
			ValidatorAddressCodec: authcodec.NewBech32Codec("panaceavaloper"),
		},
	})
	require.NoError(t, err)
	std.RegisterInterfaces(registry)
	banktypes.RegisterInterfaces(registry)
	txConfig := authtx.NewTxConfig(codec.NewProtoCodec(registry), authtx.DefaultSignModes)
	decoded, err := txConfig.TxDecoder()(encoded)
	require.NoError(t, err, "malformed signer fields are not transaction decode errors")
	basic, ok := decoded.(sdk.HasValidateBasic)
	require.True(t, ok)
	err = basic.ValidateBasic()
	require.ErrorContains(t, err, "decoding bech32 failed")

	// The v0.50 auth TxDecoder defers signer extraction to ValidateBasic. The
	// address codec's bare error therefore has the stable undefined/1 ABCI
	// classification; sdk/2 is reserved for failures inside TxDecoder itself.
	codespace, code, log := errorsmod.ABCIInfo(err, false)
	require.Equal(t, "undefined", codespace)
	require.EqualValues(t, 1, code)
	require.Contains(t, log, "decoding bech32 failed")
}

func TestParseRawBroadcastResponseClassifiesCheckTx(t *testing.T) {
	t.Parallel()

	result, err := parseRawBroadcastResponse([]byte(`{
		"jsonrpc":"2.0",
		"id":1,
		"result":{
			"code":18,
			"data":"",
			"log":"invalid request",
			"codespace":"sdk",
			"hash":"AABBCC"
		}
	}`))
	require.NoError(t, err)
	require.Equal(t, "0", result.Height)
	require.Equal(t, "AABBCC", result.TxHash)
	require.Equal(t, "sdk", result.Codespace)
	require.EqualValues(t, 18, result.Code)
	require.Equal(t, "invalid request", result.RawLog)
}

type stubRawWallet struct {
	mnemonic string
	address  []byte
}

func (w stubRawWallet) KeyName() string          { return "raw-test" }
func (w stubRawWallet) FormattedAddress() string { return "panacea1rawtest" }
func (w stubRawWallet) Mnemonic() string         { return w.mnemonic }
func (w stubRawWallet) Address() []byte          { return w.address }
