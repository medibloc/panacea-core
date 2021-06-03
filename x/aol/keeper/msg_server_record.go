package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/x/aol/types"
)

func (k msgServer) CreateRecord(goCtx context.Context, msg *types.MsgCreateRecord) (*types.MsgCreateRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	id := k.AppendRecord(
		ctx,
		msg.Creator,
		msg.Key,
		msg.Value,
		msg.NanoTimestamp,
		msg.WriterAddress,
	)

	return &types.MsgCreateRecordResponse{
		Id: id,
	}, nil
}

func (k msgServer) UpdateRecord(goCtx context.Context, msg *types.MsgUpdateRecord) (*types.MsgUpdateRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var record = types.Record{
		Creator:       msg.Creator,
		Id:            msg.Id,
		Key:           msg.Key,
		Value:         msg.Value,
		NanoTimestamp: msg.NanoTimestamp,
		WriterAddress: msg.WriterAddress,
	}

	// Checks that the element exists
	if !k.HasRecord(ctx, msg.Id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}

	// Checks if the the msg sender is the same as the current owner
	if msg.Creator != k.GetRecordOwner(ctx, msg.Id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	k.SetRecord(ctx, record)

	return &types.MsgUpdateRecordResponse{}, nil
}

func (k msgServer) DeleteRecord(goCtx context.Context, msg *types.MsgDeleteRecord) (*types.MsgDeleteRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasRecord(ctx, msg.Id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}
	if msg.Creator != k.GetRecordOwner(ctx, msg.Id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	k.RemoveRecord(ctx, msg.Id)

	return &types.MsgDeleteRecordResponse{}, nil
}
