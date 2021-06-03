package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	// this line is used by starport scaffolding # 2
	cdc.RegisterConcrete(&MsgCreateOwner{}, "aol/CreateOwner", nil)
	cdc.RegisterConcrete(&MsgUpdateOwner{}, "aol/UpdateOwner", nil)
	cdc.RegisterConcrete(&MsgDeleteOwner{}, "aol/DeleteOwner", nil)

	cdc.RegisterConcrete(&MsgCreateRecord{}, "aol/CreateRecord", nil)
	cdc.RegisterConcrete(&MsgUpdateRecord{}, "aol/UpdateRecord", nil)
	cdc.RegisterConcrete(&MsgDeleteRecord{}, "aol/DeleteRecord", nil)

	cdc.RegisterConcrete(&MsgCreateWriter{}, "aol/CreateWriter", nil)
	cdc.RegisterConcrete(&MsgUpdateWriter{}, "aol/UpdateWriter", nil)
	cdc.RegisterConcrete(&MsgDeleteWriter{}, "aol/DeleteWriter", nil)

	cdc.RegisterConcrete(&MsgCreateTopic{}, "aol/CreateTopic", nil)
	cdc.RegisterConcrete(&MsgUpdateTopic{}, "aol/UpdateTopic", nil)
	cdc.RegisterConcrete(&MsgDeleteTopic{}, "aol/DeleteTopic", nil)

}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// this line is used by starport scaffolding # 3
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateOwner{},
		&MsgUpdateOwner{},
		&MsgDeleteOwner{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateRecord{},
		&MsgUpdateRecord{},
		&MsgDeleteRecord{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateWriter{},
		&MsgUpdateWriter{},
		&MsgDeleteWriter{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateTopic{},
		&MsgUpdateTopic{},
		&MsgDeleteTopic{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)
