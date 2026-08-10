package app

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/medibloc/panacea-core/v2/app/params"
	aoltypes "github.com/medibloc/panacea-core/v2/x/aol/types"
	didtypes "github.com/medibloc/panacea-core/v2/x/did/types"
	pnfttypes "github.com/medibloc/panacea-core/v2/x/pnft/types"
)

// MakeEncodingConfig creates an EncodingConfig for testing
func MakeEncodingConfig() params.EncodingConfig {
	encodingConfig := params.MakeEncodingConfig(
		params.WithCustomGetSigners(aoltypes.CustomGetSigners()...),
		params.WithAminoJSONEncoderModifiers(didtypes.AminoJSONEncoderModifiers()...),
		params.WithV221LegacyAminoJSONCompatibility(
			codectypes.MsgTypeURL(&aoltypes.MsgCreateTopicRequest{}),
			codectypes.MsgTypeURL(&aoltypes.MsgAddWriterRequest{}),
			codectypes.MsgTypeURL(&aoltypes.MsgDeleteWriterRequest{}),
			codectypes.MsgTypeURL(&aoltypes.MsgAddRecordRequest{}),
			codectypes.MsgTypeURL(&didtypes.MsgCreateDIDRequest{}),
			codectypes.MsgTypeURL(&didtypes.MsgUpdateDIDRequest{}),
			codectypes.MsgTypeURL(&didtypes.MsgDeactivateDIDRequest{}),
		),
	)
	// Keep legacy Amino registrations while existing clients migrate to
	// SIGN_MODE_DIRECT. AOL/DID sign bytes retain their v2.2.1 representation.
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	// PNFT is no longer an application module, but its message types remain
	// registered so historical transactions and stored Any values can decode.
	pnfttypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	if err := encodingConfig.InterfaceRegistry.SigningContext().Validate(); err != nil {
		panic(err)
	}
	return encodingConfig
}
