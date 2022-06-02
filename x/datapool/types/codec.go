package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// this line is used by starport scaffolding # 1
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterOracle{}, "datapool/RegisterOracle", nil)
	cdc.RegisterConcrete(&MsgUpdateOracle{}, "datapool/UpdateOracle", nil)
	cdc.RegisterConcrete(&MsgCreatePool{}, "datapool/CreatePool", nil)
	cdc.RegisterConcrete(&MsgSellData{}, "datapool/SellData", nil)
	cdc.RegisterConcrete(&MsgBuyDataPass{}, "datapool/BuyDataPass", nil)
	cdc.RegisterConcrete(&MsgRedeemDataPass{}, "datapool/RedeemDataPass", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterOracle{},
		&MsgUpdateOracle{},
		&MsgCreatePool{},
		&MsgSellData{},
		&MsgBuyDataPass{},
		&MsgRedeemDataPass{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)
