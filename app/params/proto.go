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
	customGetSigners                   []txsigning.CustomGetSigner
	aminoJSONEncoderModifiers          []func(aminojson.Encoder) aminojson.Encoder
	legacyAminoJSONBareMessageTypeURLs []string
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

// WithV221LegacyAminoJSONCompatibility preserves the pre-v0.50 bare JSON
// representation for the listed top-level message type URLs. Remove it after
// all clients migrate to SIGN_MODE_DIRECT.
func WithV221LegacyAminoJSONCompatibility(typeURLs ...string) EncodingConfigOption {
	return func(options *encodingConfigOptions) {
		options.legacyAminoJSONBareMessageTypeURLs = append(options.legacyAminoJSONBareMessageTypeURLs, typeURLs...)
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

	legacyAminoJSONHandler := txsigning.SignModeHandler(aminojson.NewSignModeHandler(aminojson.SignModeHandlerOptions{
		FileResolver: interfaceRegistry,
		Encoder:      &aminoJSONEncoder,
	}))
	if len(options.legacyAminoJSONBareMessageTypeURLs) > 0 {
		legacyAminoJSONHandler = newLegacyAminoJSONCompatHandler(
			legacyAminoJSONHandler,
			options.legacyAminoJSONBareMessageTypeURLs,
		)
	}

	// SIGN_MODE_TEXTUAL is intentionally not enabled. It requires an online coin
	// metadata query and does not support offline signing. Panacea retains direct,
	// direct-aux, and legacy Amino JSON signing during the client migration.
	txConfig, err := tx.NewTxConfigWithOptions(protoCodec, tx.ConfigOptions{
		EnabledSignModes: []signingtypes.SignMode{
			signingtypes.SignMode_SIGN_MODE_DIRECT,
			signingtypes.SignMode_SIGN_MODE_DIRECT_AUX,
		},
		CustomSignModes: []txsigning.SignModeHandler{
			legacyAminoJSONHandler,
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
