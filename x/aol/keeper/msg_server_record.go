package keeper

import (
	"context"
	"cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/aol/types"
)

func (k msgServer) AddRecord(goCtx context.Context, msg *types.MsgAddRecordRequest) (*types.MsgAddRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ownerAddr, err := sdk.AccAddressFromBech32(msg.OwnerAddress)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address: %v", err)
	}
	writerAddr, err := sdk.AccAddressFromBech32(msg.WriterAddress)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid writer address: %v", err)
	}

	topicKey := types.TopicCompositeKey{OwnerAddress: ownerAddr, TopicName: msg.TopicName}
	if !k.HasTopic(ctx, topicKey) {
		return nil, errors.Wrapf(types.ErrTopicNotFound, "topic <%s, %s>", msg.OwnerAddress, msg.TopicName)
	}
	writerKey := types.WriterCompositeKey{OwnerAddress: ownerAddr, TopicName: msg.TopicName, WriterAddress: writerAddr}
	if !k.HasWriter(ctx, writerKey) {
		return nil, errors.Wrapf(types.ErrWriterNotAuthorized, "writer <%s, %s, %s>", msg.OwnerAddress, msg.TopicName, msg.WriterAddress)
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
