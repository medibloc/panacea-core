package keeper

import (
	"context"

	"github.com/medibloc/panacea-core/v2/types/compkey"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/medibloc/panacea-core/v2/x/aol/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Writer(c context.Context, req *types.QueryServiceWriterRequest) (*types.QueryServiceWriterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	ownerAddr, err := sdk.AccAddressFromBech32(req.OwnerAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner address")
	}
	writerAddr, err := sdk.AccAddressFromBech32(req.WriterAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid writer address")
	}

	writerKey := types.WriterCompositeKey{OwnerAddress: ownerAddr, TopicName: req.TopicName, WriterAddress: writerAddr}
	if !k.HasWriter(ctx, writerKey) {
		return nil, status.Error(codes.NotFound, "writer not found")
	}

	writer := k.GetWriter(ctx, writerKey)
	return &types.QueryServiceWriterResponse{Writer: &writer}, nil
}

func (k Keeper) Writers(c context.Context, req *types.QueryServiceWritersRequest) (*types.QueryServiceWritersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var writerAddresses []string
	ctx := sdk.UnwrapSDKContext(c)

	ownerAddr, err := sdk.AccAddressFromBech32(req.OwnerAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner address")
	}
	compKeyPrefix, err := compkey.PartialEncode(&types.WriterCompositeKey{OwnerAddress: ownerAddr, TopicName: req.TopicName, WriterAddress: nil}, 2)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to writer key")
	}

	store := ctx.KVStore(k.storeKey)
	writerStore := prefix.NewStore(store, append(types.WriterKeyPrefix, compKeyPrefix...))

	pageRes, err := query.Paginate(writerStore, req.Pagination, func(compKeyLast []byte, value []byte) error {
		var compKey types.WriterCompositeKey
		if err := compkey.Decode(append(compKeyPrefix, compKeyLast...), &compKey); err != nil {
			return err
		}
		writerAddresses = append(writerAddresses, compKey.WriterAddress.String())
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryServiceWritersResponse{WriterAddresses: writerAddresses, Pagination: pageRes}, nil
}
