package keeper

import (
	"context"
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/pnft/types"
)

type msgServer struct {
	*Keeper
}

func NewMsgServerImpl(keeper *Keeper) types.MsgServiceServer {
	return &msgServer{Keeper: keeper}
}

func (m msgServer) CreateDenom(goCtx context.Context, request *types.MsgServiceCreateDenomRequest) (*types.MsgServiceCreateDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := request.ValidateBasic(); err != nil {
		return nil, err
	}

	err := m.Keeper.SaveDenom(ctx, request.Denom)
	if err != nil {
		return nil, errors.Wrapf(types.ErrCreateDenom, err.Error())
	}
	return &types.MsgServiceCreateDenomResponse{}, nil
}

func (m msgServer) UpdateDenom(goCtx context.Context, request *types.MsgServiceUpdateDenomRequest) (*types.MsgServiceUpdateDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := request.ValidateBasic(); err != nil {
		return nil, err
	}

	if err := m.Keeper.UpdateDenom(
		ctx,
		&types.Denom{
			Id:          request.Id,
			Name:        request.Name,
			Symbol:      request.Symbol,
			Description: request.Description,
			Uri:         request.Uri,
			UriHash:     request.UriHash,
			Creator:     "",
			Data:        request.Data,
		},
		request.Updater,
	); err != nil {
		return nil, errors.Wrapf(types.ErrUpdateDenom, err.Error())
	}

	return &types.MsgServiceUpdateDenomResponse{}, nil
}

func (m msgServer) DeleteDenom(goCtx context.Context, request *types.MsgServiceDeleteDenomRequest) (*types.MsgServiceDeleteDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := request.ValidateBasic(); err != nil {
		return nil, err
	}

	if err := m.Keeper.DeleteDenom(ctx, request.Id, request.Remover); err != nil {
		return nil, errors.Wrapf(types.ErrDeleteDenom, err.Error())
	}

	return &types.MsgServiceDeleteDenomResponse{}, nil
}

func (m msgServer) TransferDenom(goCtx context.Context, request *types.MsgServiceTransferDenomRequest) (*types.MsgServiceTransferDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := request.ValidateBasic(); err != nil {
		return nil, err
	}

	if err := m.Keeper.TransferDenomOwner(ctx, request.Id, request.Sender, request.Receiver); err != nil {
		return nil, err
	}

	return &types.MsgServiceTransferDenomResponse{}, nil
}
