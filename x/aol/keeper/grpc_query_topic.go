package keeper

import (
	"context"

	"github.com/medibloc/panacea-core/types/compkey"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/medibloc/panacea-core/x/aol/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Topic(c context.Context, req *types.QueryGetTopicRequest) (*types.QueryGetTopicResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	ownerAddr, err := sdk.AccAddressFromBech32(req.OwnerAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner address")
	}

	topicKey := types.TopicCompositeKey{OwnerAddress: ownerAddr, TopicName: req.TopicName}
	if !k.HasTopic(ctx, topicKey) {
		return nil, status.Error(codes.NotFound, "topic not found")
	}

	topic := k.GetTopic(ctx, topicKey)
	return &types.QueryGetTopicResponse{Topic: &topic}, nil
}

func (k Keeper) Topics(c context.Context, req *types.QueryListTopicsRequest) (*types.QueryListTopicsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var topicNames []string
	ctx := sdk.UnwrapSDKContext(c)

	ownerAddr, err := sdk.AccAddressFromBech32(req.OwnerAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner address")
	}
	compKeyPrefix, err := compkey.PartialEncode(&types.TopicCompositeKey{OwnerAddress: ownerAddr, TopicName: ""}, 1)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fail to encode topic key: %v", err.Error())
	}

	store := ctx.KVStore(k.storeKey)
	topicStore := prefix.NewStore(store, append(types.TopicKeyPrefix, compKeyPrefix...))

	pageRes, err := query.Paginate(topicStore, req.Pagination, func(compKeyLast []byte, value []byte) error {
		var compKey types.TopicCompositeKey
		if err := compkey.Decode(append(compKeyPrefix, compKeyLast...), &compKey); err != nil {
			return err
		}
		topicNames = append(topicNames, compKey.TopicName)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryListTopicsResponse{TopicNames: topicNames, Pagination: pageRes}, nil
}
