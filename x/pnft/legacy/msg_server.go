package legacy

import (
	"context"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/x/pnft/types"
)

const DisabledErrorMessage = "legacy PNFT messages are disabled"

var _ types.MsgServer = MsgServer{}

// MsgServer retains the legacy PNFT message routes only so messages stored in
// proposals fail deterministically after the stateful PNFT module is removed.
// It must not depend on a keeper or access a store.
type MsgServer struct{}

func NewMsgServer() types.MsgServer {
	return MsgServer{}
}

func (MsgServer) CreateDenom(context.Context, *types.MsgCreateDenomRequest) (*types.MsgCreateDenomResponse, error) {
	return nil, disabledError()
}

func (MsgServer) UpdateDenom(context.Context, *types.MsgUpdateDenomRequest) (*types.MsgUpdateDenomResponse, error) {
	return nil, disabledError()
}

func (MsgServer) DeleteDenom(context.Context, *types.MsgDeleteDenomRequest) (*types.MsgDeleteDenomResponse, error) {
	return nil, disabledError()
}

func (MsgServer) TransferDenom(context.Context, *types.MsgTransferDenomRequest) (*types.MsgTransferDenomResponse, error) {
	return nil, disabledError()
}

func (MsgServer) MintPNFT(context.Context, *types.MsgMintPNFTRequest) (*types.MsgMintPNFTResponse, error) {
	return nil, disabledError()
}

func (MsgServer) TransferPNFT(context.Context, *types.MsgTransferPNFTRequest) (*types.MsgTransferPNFTResponse, error) {
	return nil, disabledError()
}

func (MsgServer) BurnPNFT(context.Context, *types.MsgBurnPNFTRequest) (*types.MsgBurnPNFTResponse, error) {
	return nil, disabledError()
}

func disabledError() error {
	return sdkerrors.ErrInvalidRequest.Wrap(DisabledErrorMessage)
}
