package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// ModuleCdc generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc := codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}

// RegisterCodec registers concrete types on Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreateDID{}, "did/MsgCreateDID", nil)
	cdc.RegisterConcrete(MsgUpdateDID{}, "did/MsgUpdateDID", nil)
	cdc.RegisterConcrete(MsgDeactivateDID{}, "did/MsgDeactivateDID", nil)
}
