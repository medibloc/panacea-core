package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// this line is used by starport scaffolding # 1
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterDataValidator{}, "datapool/RegisterDataValidator", nil)
	cdc.RegisterConcrete(&MsgUpdateDataValidator{}, "datapool/UpdateDataValidator", nil)
	cdc.RegisterConcrete(&MsgCreatePool{}, "datapool/CreatePool", nil)
	cdc.RegisterConcrete(&MsgSellData{}, "datapool/SellData", nil)
	cdc.RegisterConcrete(&MsgBuyDataAccessNFT{}, "datapool/BuyDataAccessNFT", nil)
	cdc.RegisterConcrete(&MsgRedeemDataAccessNFT{}, "datapool/RedeemDataAccessNFT", nil)
	cdc.RegisterConcrete(&MsgRegisterNFTContract{}, "datapool/RegisterNFTContract", nil)
	cdc.RegisterConcrete(&MsgUpgradeNFTContract{}, "datapool/UpgradeNFTContract", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterDataValidator{},
		&MsgUpdateDataValidator{},
		&MsgCreatePool{},
		&MsgSellData{},
		&MsgBuyDataAccessNFT{},
		&MsgRedeemDataAccessNFT{},
		&MsgRegisterNFTContract{},
		&MsgUpgradeNFTContract{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)
