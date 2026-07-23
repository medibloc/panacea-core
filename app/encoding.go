package app

import (
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
	)
	// Keep legacy Amino registrations for SIGN_MODE_LEGACY_AMINO_JSON
	// and compatibility with existing clients and hardware wallets.
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
