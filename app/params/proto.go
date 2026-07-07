package params

import (
	txsigning "cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	gogoproto "github.com/cosmos/gogoproto/proto"
)

type EncodingConfigOption func(*encodingConfigOptions)

type encodingConfigOptions struct {
	customGetSigners          []txsigning.CustomGetSigner
	aminoJSONEncoderModifiers []func(aminojson.Encoder) aminojson.Encoder
}

func WithCustomGetSigners(customGetSigners ...txsigning.CustomGetSigner) EncodingConfigOption {
	return func(options *encodingConfigOptions) {
		options.customGetSigners = append(options.customGetSigners, customGetSigners...)
	}
}

func WithAminoJSONEncoderModifiers(modifiers ...func(aminojson.Encoder) aminojson.Encoder) EncodingConfigOption {
	return func(options *encodingConfigOptions) {
		options.aminoJSONEncoderModifiers = append(options.aminoJSONEncoderModifiers, modifiers...)
	}
}

// MakeEncodingConfig creates an EncodingConfig for an amino based test configuration.
func MakeEncodingConfig(configOptions ...EncodingConfigOption) EncodingConfig {
	var options encodingConfigOptions
	for _, applyOption := range configOptions {
		applyOption(&options)
	}

	cdc := codec.NewLegacyAmino()

	signingOptions := txsigning.Options{
		AddressCodec:          authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
	}
	for _, signer := range options.customGetSigners {
		signingOptions.DefineCustomGetSigners(signer.MsgType, signer.Fn)
	}

	interfaceRegistry, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles:     gogoproto.HybridResolver,
		SigningOptions: signingOptions,
	})
	if err != nil {
		panic(err)
	}

	protoCodec := codec.NewProtoCodec(interfaceRegistry)
	aminoJSONEncoder := aminojson.NewEncoder(aminojson.EncoderOptions{
		FileResolver: interfaceRegistry,
	})
	for _, modifier := range options.aminoJSONEncoderModifiers {
		aminoJSONEncoder = modifier(aminoJSONEncoder)
	}

	txConfig, err := tx.NewTxConfigWithOptions(protoCodec, tx.ConfigOptions{
		EnabledSignModes: []signingtypes.SignMode{
			signingtypes.SignMode_SIGN_MODE_DIRECT,
			signingtypes.SignMode_SIGN_MODE_DIRECT_AUX,
		},
		CustomSignModes: []txsigning.SignModeHandler{
			aminojson.NewSignModeHandler(aminojson.SignModeHandlerOptions{
				FileResolver: interfaceRegistry,
				Encoder:      &aminoJSONEncoder,
			}),
		},
		SigningOptions: &signingOptions,
	})
	if err != nil {
		panic(err)
	}

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             protoCodec,
		TxConfig:          txConfig,
		Amino:             cdc,
	}
}
