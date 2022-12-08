package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateDeal{}, "datadeal/CreateDeal", nil)
	cdc.RegisterConcrete(&MsgDeactivateDeal{}, "datadeal/MsgDeactivateDeal", nil)
	cdc.RegisterConcrete(&MsgSubmitConsent{}, "datadeal/SubmitConsent", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateDeal{},
		&MsgDeactivateDeal{},
		&MsgSubmitConsent{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)
