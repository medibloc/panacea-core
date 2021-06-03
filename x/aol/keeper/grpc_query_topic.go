package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/medibloc/panacea-core/x/aol/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) TopicAll(c context.Context, req *types.QueryAllTopicRequest) (*types.QueryAllTopicResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var topics []*types.Topic
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	topicStore := prefix.NewStore(store, types.KeyPrefix(types.TopicKey))

	pageRes, err := query.Paginate(topicStore, req.Pagination, func(key []byte, value []byte) error {
		var topic types.Topic
		if err := k.cdc.UnmarshalBinaryBare(value, &topic); err != nil {
			return err
		}

		topics = append(topics, &topic)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllTopicResponse{Topic: topics, Pagination: pageRes}, nil
}

func (k Keeper) Topic(c context.Context, req *types.QueryGetTopicRequest) (*types.QueryGetTopicResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var topic types.Topic
	ctx := sdk.UnwrapSDKContext(c)

	if !k.HasTopic(ctx, req.Id) {
		return nil, sdkerrors.ErrKeyNotFound
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TopicKey))
	k.cdc.MustUnmarshalBinaryBare(store.Get(GetTopicIDBytes(req.Id)), &topic)

	return &types.QueryGetTopicResponse{Topic: &topic}, nil
}
