package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterOracle{}, "oracle/RegisterOracle", nil)
	cdc.RegisterConcrete(&MsgApproveOracleRegistration{}, "oracle/ApproveOracleRegistration", nil)
	cdc.RegisterConcrete(&MsgUpdateOracleInfo{}, "oracle/UpdateOracleInfo", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterOracle{},
		&MsgApproveOracleRegistration{},
		&MsgUpdateOracleInfo{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)
