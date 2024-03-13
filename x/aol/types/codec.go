package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	// this line is used by starport scaffolding # 2
	cdc.RegisterConcrete(&MsgCreateTopicRequest{}, "aol/CreateTopic", nil)
	cdc.RegisterConcrete(&MsgAddWriterRequest{}, "aol/AddWriter", nil)
	cdc.RegisterConcrete(&MsgDeleteWriterRequest{}, "aol/DeleteWriter", nil)
	cdc.RegisterConcrete(&MsgAddRecordRequest{}, "aol/AddRecord", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// this line is used by starport scaffolding # 3
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateTopicRequest{},
		&MsgAddWriterRequest{},
		&MsgDeleteWriterRequest{},
		&MsgAddRecordRequest{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)
