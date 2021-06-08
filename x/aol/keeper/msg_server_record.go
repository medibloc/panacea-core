package keeper

import (
	"context"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/aol/types"
)

func (k msgServer) AddRecord(goCtx context.Context, msg *types.MsgAddRecord) (*types.MsgAddRecordResponse, error) {
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
	if !k.HasWriter(ctx, writerKey) {
		return nil, sdkerrors.Wrapf(types.ErrWriterNotAuthorized, "writer <%s, %s, %s>", msg.OwnerAddress, msg.TopicName, msg.WriterAddress)
	}

	topic := k.GetTopic(ctx, topicKey)
	offset := topic.NextRecordOffset()
	k.SetTopic(ctx, topicKey, topic.IncreaseTotalRecords())

	recordKey := types.RecordCompositeKey{OwnerAddress: ownerAddr, TopicName: msg.TopicName, Offset: offset}
	record := types.Record{
		Key:           msg.Key,
		Value:         msg.Value,
		NanoTimestamp: ctx.BlockTime().UnixNano(),
		WriterAddress: msg.WriterAddress,
	}
	k.SetRecord(ctx, recordKey, record)

	return &types.MsgAddRecordResponse{
		OwnerAddress: msg.OwnerAddress,
		TopicName:    msg.TopicName,
		Offset:       offset,
	}, nil
}
