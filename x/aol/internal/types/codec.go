package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
}

// RegisterCodec registers concrete types on Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreateTopic{},  "aol/MsgCreateTopic", nil)
	cdc.RegisterConcrete(MsgAddWriter{},    "aol/MsgAddWriter", nil)
	cdc.RegisterConcrete(MsgDeleteWriter{}, "aol/MsgDeleteWriter", nil)
	cdc.RegisterConcrete(MsgAddRecord{},    "aol/MsgAddRecord", nil)
	cdc.RegisterConcrete(ResAddRecord{},    "aol/ResAddRecord", nil)
}
