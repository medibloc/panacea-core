package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	// this line is used by starport scaffolding # 2
	cdc.RegisterConcrete(&MsgServiceCreateTopicRequest{}, "aol/CreateTopic", nil)
	cdc.RegisterConcrete(&MsgServiceAddWriterRequest{}, "aol/AddWriter", nil)
	cdc.RegisterConcrete(&MsgServiceDeleteWriterRequest{}, "aol/DeleteWriter", nil)
	cdc.RegisterConcrete(&MsgServiceAddRecordRequest{}, "aol/AddRecord", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// this line is used by starport scaffolding # 3
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgServiceCreateTopicRequest{},
		&MsgServiceAddWriterRequest{},
		&MsgServiceDeleteWriterRequest{},
		&MsgServiceAddRecordRequest{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_MsgService_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)
