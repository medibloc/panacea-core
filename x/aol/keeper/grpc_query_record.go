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

func (k Keeper) RecordAll(c context.Context, req *types.QueryAllRecordRequest) (*types.QueryAllRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var records []*types.Record
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	recordStore := prefix.NewStore(store, types.KeyPrefix(types.RecordKey))

	pageRes, err := query.Paginate(recordStore, req.Pagination, func(key []byte, value []byte) error {
		var record types.Record
		if err := k.cdc.UnmarshalBinaryBare(value, &record); err != nil {
			return err
		}

		records = append(records, &record)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllRecordResponse{Record: records, Pagination: pageRes}, nil
}

func (k Keeper) Record(c context.Context, req *types.QueryGetRecordRequest) (*types.QueryGetRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var record types.Record
	ctx := sdk.UnwrapSDKContext(c)

	if !k.HasRecord(ctx, req.Id) {
		return nil, sdkerrors.ErrKeyNotFound
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RecordKey))
	k.cdc.MustUnmarshalBinaryBare(store.Get(GetRecordIDBytes(req.Id)), &record)

	return &types.QueryGetRecordResponse{Record: &record}, nil
}
