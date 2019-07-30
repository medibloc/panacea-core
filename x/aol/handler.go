package aol

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/aol/types"
)

// NewHandler returns a handler for "aol" type messages
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgCreateTopic:
			return handleMsgCreateTopic(ctx, keeper, msg)
		case MsgAddWriter:
			return handleMsgAddWriter(ctx, keeper, msg)
		case MsgDeleteWriter:
			return handleMsgDeleteWriter(ctx, keeper, msg)
		case MsgAddRecord:
			return handleMsgAddRecord(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized aol Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgCreateTopic(ctx sdk.Context, keeper Keeper, msg MsgCreateTopic) sdk.Result {
	if keeper.HasTopic(ctx, msg.OwnerAddress, msg.TopicName) {
		return types.ErrTopicExists(msg.TopicName).Result()
	}

	owner := keeper.GetOwner(ctx, msg.OwnerAddress).IncreaseTotalTopics()
	keeper.SetOwner(ctx, msg.OwnerAddress, owner)

	keeper.SetTopic(ctx, msg.OwnerAddress, msg.TopicName, types.NewTopic(msg.Description))

	return sdk.Result{}
}

func handleMsgAddWriter(ctx sdk.Context, keeper Keeper, msg MsgAddWriter) sdk.Result {
	if !keeper.HasTopic(ctx, msg.OwnerAddress, msg.TopicName) {
		return types.ErrTopicNotFound(msg.TopicName).Result()
	}
	if keeper.HasWriter(ctx, msg.OwnerAddress, msg.TopicName, msg.WriterAddress) {
		return types.ErrWriterExists(msg.WriterAddress).Result()
	}

	topic := keeper.GetTopic(ctx, msg.OwnerAddress, msg.TopicName).IncreaseTotalWriters()
	keeper.SetTopic(ctx, msg.OwnerAddress, msg.TopicName, topic)

	keeper.SetWriter(ctx, msg.OwnerAddress, msg.TopicName, msg.WriterAddress,
		types.NewWriter(msg.Moniker, msg.Description, ctx.BlockHeader().Time.UnixNano()))

	return sdk.Result{}
}

func handleMsgDeleteWriter(ctx sdk.Context, keeper Keeper, msg MsgDeleteWriter) sdk.Result {
	if !keeper.HasWriter(ctx, msg.OwnerAddress, msg.TopicName, msg.WriterAddress) {
		return types.ErrWriterNotFound(msg.WriterAddress).Result()
	}

	topic := keeper.GetTopic(ctx, msg.OwnerAddress, msg.TopicName).DecreaseTotalWriters()
	keeper.SetTopic(ctx, msg.OwnerAddress, msg.TopicName, topic)

	keeper.DeleteWriter(ctx, msg.OwnerAddress, msg.TopicName, msg.WriterAddress)

	return sdk.Result{}
}

func handleMsgAddRecord(ctx sdk.Context, keeper Keeper, msg MsgAddRecord) sdk.Result {
	if !keeper.HasTopic(ctx, msg.OwnerAddress, msg.TopicName) {
		return types.ErrTopicNotFound(msg.TopicName).Result()
	}
	if !keeper.HasWriter(ctx, msg.OwnerAddress, msg.TopicName, msg.WriterAddress) {
		return types.ErrWriterNotAuthorized(msg.WriterAddress).Result()
	}

	topic := keeper.GetTopic(ctx, msg.OwnerAddress, msg.TopicName)
	offset := topic.NextRecordOffset()
	keeper.SetTopic(ctx, msg.OwnerAddress, msg.TopicName, topic.IncreaseTotalRecords())

	keeper.SetRecord(ctx, msg.OwnerAddress, msg.TopicName, offset,
		types.NewRecord(msg.Key, msg.Value, ctx.BlockHeader().Time.UnixNano(), msg.WriterAddress))

	response := types.NewMsgAddRecordResponse(msg.OwnerAddress, msg.TopicName, offset)
	data := response.MustMarshalJSON()

	return sdk.Result{
		Data: []byte(data),
	}
}
