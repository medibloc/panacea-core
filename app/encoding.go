package app

import (
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/medibloc/panacea-core/v2/app/params"
	aoltypes "github.com/medibloc/panacea-core/v2/x/aol/types"
	didtypes "github.com/medibloc/panacea-core/v2/x/did/types"
)

// MakeEncodingConfig creates an EncodingConfig for testing
func MakeEncodingConfig() params.EncodingConfig {
	encodingConfig := params.MakeEncodingConfig(
		params.WithCustomGetSigners(aoltypes.CustomGetSigners()...),
		params.WithAminoJSONEncoderModifiers(didtypes.AminoJSONEncoderModifiers()...),
	)
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	if err := encodingConfig.InterfaceRegistry.SigningContext().Validate(); err != nil {
		panic(err)
	}
	return encodingConfig
}
