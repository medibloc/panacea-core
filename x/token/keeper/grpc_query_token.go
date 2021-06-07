package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/medibloc/panacea-core/x/token/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) TokenAll(c context.Context, req *types.QueryAllTokenRequest) (*types.QueryAllTokenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var tokens []*types.Token
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	tokenStore := prefix.NewStore(store, types.KeyPrefix(types.TokenKey))

	pageRes, err := query.Paginate(tokenStore, req.Pagination, func(key []byte, value []byte) error {
		var token types.Token
		if err := k.cdc.UnmarshalBinaryBare(value, &token); err != nil {
			return err
		}

		tokens = append(tokens, &token)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllTokenResponse{Token: tokens, Pagination: pageRes}, nil
}

func (k Keeper) Token(c context.Context, req *types.QueryGetTokenRequest) (*types.QueryGetTokenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var token types.Token
	ctx := sdk.UnwrapSDKContext(c)

	if !k.HasToken(ctx, req.Symbol) {
		return nil, sdkerrors.ErrKeyNotFound
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenKey))
	k.cdc.MustUnmarshalBinaryBare(store.Get(GetSymbolBytes(req.Symbol)), &token)

	return &types.QueryGetTokenResponse{Token: &token}, nil
}
