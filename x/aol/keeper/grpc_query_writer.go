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

func (k Keeper) WriterAll(c context.Context, req *types.QueryAllWriterRequest) (*types.QueryAllWriterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var writers []*types.Writer
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	writerStore := prefix.NewStore(store, types.KeyPrefix(types.WriterKey))

	pageRes, err := query.Paginate(writerStore, req.Pagination, func(key []byte, value []byte) error {
		var writer types.Writer
		if err := k.cdc.UnmarshalBinaryBare(value, &writer); err != nil {
			return err
		}

		writers = append(writers, &writer)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllWriterResponse{Writer: writers, Pagination: pageRes}, nil
}

func (k Keeper) Writer(c context.Context, req *types.QueryGetWriterRequest) (*types.QueryGetWriterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var writer types.Writer
	ctx := sdk.UnwrapSDKContext(c)

	if !k.HasWriter(ctx, req.Id) {
		return nil, sdkerrors.ErrKeyNotFound
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.WriterKey))
	k.cdc.MustUnmarshalBinaryBare(store.Get(GetWriterIDBytes(req.Id)), &writer)

	return &types.QueryGetWriterResponse{Writer: &writer}, nil
}
