package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateDenomRequest{}, "pnft/CreateDenom", nil)
	cdc.RegisterConcrete(&MsgUpdateDenomRequest{}, "pnft/UpdateDenom", nil)
	cdc.RegisterConcrete(&MsgDeleteDenomRequest{}, "pnft/DeleteDenom", nil)
	cdc.RegisterConcrete(&MsgTransferDenomRequest{}, "pnft/TransferDenom", nil)

	cdc.RegisterConcrete(&MsgMintPNFTRequest{}, "pnft/MintPNFT", nil)
	cdc.RegisterConcrete(&MsgTransferPNFTRequest{}, "pnft/TransferPNFT", nil)
	cdc.RegisterConcrete(&MsgBurnPNFTRequest{}, "pnft/BurnPNFT", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateDenomRequest{},
		&MsgUpdateDenomRequest{},
		&MsgDeleteDenomRequest{},
		&MsgTransferDenomRequest{},
		&MsgMintPNFTRequest{},
		&MsgTransferPNFTRequest{},
		&MsgBurnPNFTRequest{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)
