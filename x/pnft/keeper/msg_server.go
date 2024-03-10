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
		return nil, errors.Wrap(types.ErrCreateDenom, err.Error())
	}

	err := m.Keeper.SaveDenom(
		ctx,
		&types.Denom{
			Id:          request.Id,
			Name:        request.Name,
			Symbol:      request.Symbol,
			Description: request.Description,
			Uri:         request.Uri,
			UriHash:     request.UriHash,
			Owner:       request.Creator,
			Data:        request.Data,
		},
	)
	if err != nil {
		return nil, errors.Wrapf(types.ErrCreateDenom, err.Error())
	}
	return &types.MsgServiceCreateDenomResponse{}, nil
}

func (m msgServer) UpdateDenom(goCtx context.Context, request *types.MsgServiceUpdateDenomRequest) (*types.MsgServiceUpdateDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := request.ValidateBasic(); err != nil {
		return nil, errors.Wrap(types.ErrUpdateDenom, err.Error())
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
			Owner:       request.Updater,
			Data:        request.Data,
		},
	); err != nil {
		return nil, errors.Wrapf(types.ErrUpdateDenom, err.Error())
	}

	return &types.MsgServiceUpdateDenomResponse{}, nil
}

func (m msgServer) DeleteDenom(goCtx context.Context, request *types.MsgServiceDeleteDenomRequest) (*types.MsgServiceDeleteDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := request.ValidateBasic(); err != nil {
		return nil, errors.Wrap(types.ErrDeleteDenom, err.Error())
	}

	if err := m.Keeper.DeleteDenom(ctx, request.Id, request.Remover); err != nil {
		return nil, errors.Wrapf(types.ErrDeleteDenom, err.Error())
	}

	return &types.MsgServiceDeleteDenomResponse{}, nil
}

func (m msgServer) TransferDenom(goCtx context.Context, request *types.MsgServiceTransferDenomRequest) (*types.MsgServiceTransferDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := request.ValidateBasic(); err != nil {
		return nil, errors.Wrap(types.ErrTransferDenom, err.Error())
	}

	if err := m.Keeper.TransferDenomOwner(ctx, request.Id, request.Sender, request.Receiver); err != nil {
		return nil, errors.Wrap(types.ErrTransferDenom, err.Error())
	}

	return &types.MsgServiceTransferDenomResponse{}, nil
}

func (m msgServer) MintPNFT(goCtx context.Context, request *types.MsgServiceMintPNFTRequest) (*types.MsgServiceMintPNFTResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := request.ValidateBasic(); err != nil {
		return nil, errors.Wrap(types.ErrMintPNFT, err.Error())
	}

	msg := &types.PNFT{
		DenomId:     request.DenomId,
		Id:          request.Id,
		Name:        request.Name,
		Description: request.Description,
		Uri:         request.Uri,
		UriHash:     request.UriHash,
		Data:        request.Data,
		Creator:     request.Creator,
		CreatedAt:   ctx.BlockTime(),
	}

	if err := m.Keeper.MintPNFT(ctx, msg); err != nil {
		return nil, errors.Wrap(types.ErrMintPNFT, err.Error())
	}

	return &types.MsgServiceMintPNFTResponse{}, nil
}

func (m msgServer) TransferPNFT(goCtx context.Context, request *types.MsgServiceTransferPNFTRequest) (*types.MsgServiceTransferPNFTResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := request.ValidateBasic(); err != nil {
		return nil, errors.Wrap(types.ErrTransferPNFT, err.Error())
	}

	if err := m.Keeper.TransferPNFT(
		ctx,
		request.DenomId,
		request.Id,
		request.Sender,
		request.Receiver,
	); err != nil {
		return nil, errors.Wrap(types.ErrTransferPNFT, err.Error())
	}

	return &types.MsgServiceTransferPNFTResponse{}, nil
}

func (m msgServer) BurnPNFT(goCtx context.Context, request *types.MsgServiceBurnPNFTRequest) (*types.MsgServiceBurnPNFTResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := request.ValidateBasic(); err != nil {
		return nil, errors.Wrap(types.ErrBurnPNFT, err.Error())
	}

	if err := m.Keeper.BurnPNFT(
		ctx,
		request.DenomId,
		request.Id,
		request.Burner,
	); err != nil {
		return nil, errors.Wrap(types.ErrBurnPNFT, err.Error())
	}

	return &types.MsgServiceBurnPNFTResponse{}, nil
}
