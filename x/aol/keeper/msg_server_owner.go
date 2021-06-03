package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/x/aol/types"
)

func (k msgServer) CreateOwner(goCtx context.Context, msg *types.MsgCreateOwner) (*types.MsgCreateOwnerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	id := k.AppendOwner(
		ctx,
		msg.Creator,
		msg.TotalTopics,
	)

	return &types.MsgCreateOwnerResponse{
		Id: id,
	}, nil
}

func (k msgServer) UpdateOwner(goCtx context.Context, msg *types.MsgUpdateOwner) (*types.MsgUpdateOwnerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var owner = types.Owner{
		Creator:     msg.Creator,
		Id:          msg.Id,
		TotalTopics: msg.TotalTopics,
	}

	// Checks that the element exists
	if !k.HasOwner(ctx, msg.Id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}

	// Checks if the the msg sender is the same as the current owner
	if msg.Creator != k.GetOwnerOwner(ctx, msg.Id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	k.SetOwner(ctx, owner)

	return &types.MsgUpdateOwnerResponse{}, nil
}

func (k msgServer) DeleteOwner(goCtx context.Context, msg *types.MsgDeleteOwner) (*types.MsgDeleteOwnerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasOwner(ctx, msg.Id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}
	if msg.Creator != k.GetOwnerOwner(ctx, msg.Id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	k.RemoveOwner(ctx, msg.Id)

	return &types.MsgDeleteOwnerResponse{}, nil
}
