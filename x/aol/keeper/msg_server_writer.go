package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/x/aol/types"
)

func (k msgServer) AddWriter(goCtx context.Context, msg *types.MsgAddWriter) (*types.MsgAddWriterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ownerAddr, err := sdk.AccAddressFromBech32(msg.OwnerAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address: %v", err)
	}
	writerAddr, err := sdk.AccAddressFromBech32(msg.WriterAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid writer address: %v", err)
	}

	topicKey := types.TopicCompositeKey{OwnerAddress: ownerAddr, TopicName: msg.TopicName}
	if !k.HasTopic(ctx, topicKey) {
		return nil, sdkerrors.Wrapf(types.ErrTopicNotFound, "topic <%s, %s>", msg.OwnerAddress, msg.TopicName)
	}
	writerKey := types.WriterCompositeKey{OwnerAddress: ownerAddr, TopicName: msg.TopicName, WriterAddress: writerAddr}
	if k.HasWriter(ctx, writerKey) {
		return nil, sdkerrors.Wrapf(types.ErrWriterExists, "writer <%s, %s, %s>", msg.OwnerAddress, msg.TopicName, msg.WriterAddress)
	}

	topic := k.GetTopic(ctx, topicKey).IncreaseTotalWriters()
	k.SetTopic(ctx, topicKey, topic)

	writer := types.Writer{
		Moniker:       msg.Moniker,
		Description:   msg.Description,
		NanoTimestamp: ctx.BlockTime().UnixNano(),
	}
	k.SetWriter(ctx, writerKey, writer)

	return &types.MsgAddWriterResponse{}, nil
}

func (k msgServer) DeleteWriter(goCtx context.Context, msg *types.MsgDeleteWriter) (*types.MsgDeleteWriterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ownerAddr, err := sdk.AccAddressFromBech32(msg.OwnerAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address: %v", err)
	}
	writerAddr, err := sdk.AccAddressFromBech32(msg.WriterAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid writer address: %v", err)
	}

	topicKey := types.TopicCompositeKey{OwnerAddress: ownerAddr, TopicName: msg.TopicName}
	writerKey := types.WriterCompositeKey{OwnerAddress: ownerAddr, TopicName: msg.TopicName, WriterAddress: writerAddr}
	if !k.HasWriter(ctx, writerKey) {
		return nil, sdkerrors.Wrapf(types.ErrWriterNotFound, "writer <%s, %s, %s>", msg.OwnerAddress, msg.TopicName, msg.WriterAddress)
	}

	topic := k.GetTopic(ctx, topicKey).DecreaseTotalWriters()
	k.SetTopic(ctx, topicKey, topic)

	k.RemoveWriter(ctx, writerKey)

	return &types.MsgDeleteWriterResponse{}, nil
}
