package types

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterInterfaces preserves decoding of legacy PNFT messages embedded in
// historical transactions and gov, group, or authz Any values. PNFT execution
// is disabled; keep this registration while historical decoding is required.
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
