package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/x/aol/types"
)

func (k msgServer) CreateTopic(goCtx context.Context, msg *types.MsgCreateTopic) (*types.MsgCreateTopicResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	id := k.AppendTopic(
		ctx,
		msg.Creator,
		msg.Description,
		msg.TotalRecords,
		msg.TotalWriter,
	)

	return &types.MsgCreateTopicResponse{
		Id: id,
	}, nil
}

func (k msgServer) UpdateTopic(goCtx context.Context, msg *types.MsgUpdateTopic) (*types.MsgUpdateTopicResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var topic = types.Topic{
		Creator:      msg.Creator,
		Id:           msg.Id,
		Description:  msg.Description,
		TotalRecords: msg.TotalRecords,
		TotalWriter:  msg.TotalWriter,
	}

	// Checks that the element exists
	if !k.HasTopic(ctx, msg.Id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}

	// Checks if the the msg sender is the same as the current owner
	if msg.Creator != k.GetTopicOwner(ctx, msg.Id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	k.SetTopic(ctx, topic)

	return &types.MsgUpdateTopicResponse{}, nil
}

func (k msgServer) DeleteTopic(goCtx context.Context, msg *types.MsgDeleteTopic) (*types.MsgDeleteTopicResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasTopic(ctx, msg.Id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}
	if msg.Creator != k.GetTopicOwner(ctx, msg.Id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	k.RemoveTopic(ctx, msg.Id)

	return &types.MsgDeleteTopicResponse{}, nil
}
