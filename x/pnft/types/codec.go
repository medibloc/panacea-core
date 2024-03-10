package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgServiceCreateDenomRequest{}, "pnft/CreateDenom", nil)
	cdc.RegisterConcrete(&MsgServiceUpdateDenomRequest{}, "pnft/UpdateDenom", nil)
	cdc.RegisterConcrete(&MsgServiceDeleteDenomRequest{}, "pnft/DeleteDenom", nil)
	cdc.RegisterConcrete(&MsgServiceTransferDenomRequest{}, "pnft/TransferDenom", nil)

	cdc.RegisterConcrete(&MsgServiceMintPNFTRequest{}, "pnft/MintPNFT", nil)
	cdc.RegisterConcrete(&MsgServiceTransferPNFTRequest{}, "pnft/TransferPNFT", nil)
	cdc.RegisterConcrete(&MsgServiceBurnPNFTRequest{}, "pnft/BurnPNFT", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgServiceCreateDenomRequest{},
		&MsgServiceUpdateDenomRequest{},
		&MsgServiceDeleteDenomRequest{},
		&MsgServiceTransferDenomRequest{},
		&MsgServiceMintPNFTRequest{},
		&MsgServiceTransferPNFTRequest{},
		&MsgServiceBurnPNFTRequest{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_MsgService_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)
