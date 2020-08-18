package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var didCodec = codec.New()

func init() {
	RegisterCodec(didCodec)
}

// RegisterCodec registers concrete types on Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreateDID{}, "did/MsgCreateDID", nil)
	cdc.RegisterConcrete(MsgUpdateDID{}, "did/MsgUpdateDID", nil)
}
