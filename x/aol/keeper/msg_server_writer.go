package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/x/aol/types"
)

func (k msgServer) CreateWriter(goCtx context.Context, msg *types.MsgCreateWriter) (*types.MsgCreateWriterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	id := k.AppendWriter(
		ctx,
		msg.Creator,
		msg.Moniker,
		msg.Description,
		msg.NanoTimestamp,
	)

	return &types.MsgCreateWriterResponse{
		Id: id,
	}, nil
}

func (k msgServer) UpdateWriter(goCtx context.Context, msg *types.MsgUpdateWriter) (*types.MsgUpdateWriterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var writer = types.Writer{
		Creator:       msg.Creator,
		Id:            msg.Id,
		Moniker:       msg.Moniker,
		Description:   msg.Description,
		NanoTimestamp: msg.NanoTimestamp,
	}

	// Checks that the element exists
	if !k.HasWriter(ctx, msg.Id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}

	// Checks if the the msg sender is the same as the current owner
	if msg.Creator != k.GetWriterOwner(ctx, msg.Id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	k.SetWriter(ctx, writer)

	return &types.MsgUpdateWriterResponse{}, nil
}

func (k msgServer) DeleteWriter(goCtx context.Context, msg *types.MsgDeleteWriter) (*types.MsgDeleteWriterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasWriter(ctx, msg.Id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}
	if msg.Creator != k.GetWriterOwner(ctx, msg.Id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	k.RemoveWriter(ctx, msg.Id)

	return &types.MsgDeleteWriterResponse{}, nil
}
