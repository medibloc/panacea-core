package keeper

import (
	"context"
	"cosmossdk.io/errors"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/aol/types"
)

func (k msgServer) CreateTopic(goCtx context.Context, msg *types.MsgServiceCreateTopicRequest) (*types.MsgServiceCreateTopicResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ownerAddr, err := sdk.AccAddressFromBech32(msg.OwnerAddress)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address: %v", err)
	}

	topicKey := types.TopicCompositeKey{OwnerAddress: ownerAddr, TopicName: msg.TopicName}
	if k.HasTopic(ctx, topicKey) {
		return nil, errors.Wrapf(types.ErrTopicExists, "topic <%s, %s>", msg.OwnerAddress, msg.TopicName)
	}

	ownerKey := types.OwnerCompositeKey{OwnerAddress: ownerAddr}
	owner := k.GetOwner(ctx, ownerKey).IncreaseTotalTopics()
	k.SetOwner(ctx, ownerKey, owner)

	topic := types.Topic{Description: msg.Description}
	k.SetTopic(ctx, topicKey, topic)

	return &types.MsgServiceCreateTopicResponse{}, nil
}
