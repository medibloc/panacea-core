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

func (k Keeper) OwnerAll(c context.Context, req *types.QueryAllOwnerRequest) (*types.QueryAllOwnerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var owners []*types.Owner
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	ownerStore := prefix.NewStore(store, types.KeyPrefix(types.OwnerKey))

	pageRes, err := query.Paginate(ownerStore, req.Pagination, func(key []byte, value []byte) error {
		var owner types.Owner
		if err := k.cdc.UnmarshalBinaryBare(value, &owner); err != nil {
			return err
		}

		owners = append(owners, &owner)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllOwnerResponse{Owner: owners, Pagination: pageRes}, nil
}

func (k Keeper) Owner(c context.Context, req *types.QueryGetOwnerRequest) (*types.QueryGetOwnerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var owner types.Owner
	ctx := sdk.UnwrapSDKContext(c)

	if !k.HasOwner(ctx, req.Id) {
		return nil, sdkerrors.ErrKeyNotFound
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OwnerKey))
	k.cdc.MustUnmarshalBinaryBare(store.Get(GetOwnerIDBytes(req.Id)), &owner)

	return &types.QueryGetOwnerResponse{Owner: &owner}, nil
}
