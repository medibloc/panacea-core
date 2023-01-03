package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterOracle{}, "oracle/RegisterOracle", nil)
	cdc.RegisterConcrete(&MsgApproveOracleRegistration{}, "oracle/ApproveOracleRegistration", nil)
	cdc.RegisterConcrete(&MsgUpdateOracleInfo{}, "oracle/UpdateOracleInfo", nil)
	cdc.RegisterConcrete(&MsgUpgradeOracle{}, "oracle/UpgradeOracle", nil)
	cdc.RegisterConcrete(&MsgApproveOracleUpgrade{}, "oracle/ApproveOracleUpgrade", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterOracle{},
		&MsgApproveOracleRegistration{},
		&MsgUpdateOracleInfo{},
		&MsgUpgradeOracle{},
		&MsgApproveOracleUpgrade{},
	)
	registry.RegisterImplementations((*govtypes.Content)(nil),
		&OracleUpgradeProposal{},
	)
	registry.RegisterImplementations((*govtypes.Content)(nil),
		&OracleUpgradeProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)
