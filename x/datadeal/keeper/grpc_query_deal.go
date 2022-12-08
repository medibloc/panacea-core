package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Deal(goCtx context.Context, req *types.QueryDealRequest) (*types.QueryDealResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	deal, err := k.GetDeal(sdk.UnwrapSDKContext(goCtx), req.DealId)
	if err != nil {
		return nil, err
	}

	return &types.QueryDealResponse{
		Deal: deal,
	}, nil
}

func (k Keeper) Deals(goCtx context.Context, req *types.QueryDealsRequest) (*types.QueryDealsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	store := ctx.KVStore(k.storeKey)
	dealsStore := prefix.NewStore(store, types.DealKey)

	var deals []*types.Deal
	pageRes, err := query.Paginate(dealsStore, req.Pagination, func(_ []byte, value []byte) error {
		var deal types.Deal
		err := k.cdc.UnmarshalLengthPrefixed(value, &deal)
		if err != nil {
			return err
		}
		deals = append(deals, &deal)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDealsResponse{
		Deal:       deals,
		Pagination: pageRes,
	}, nil
}

func (k Keeper) Certificates(ctx context.Context, certificates *types.QueryCertificates) (*types.QueryCertificatesResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (k Keeper) Certificate(ctx context.Context, certificate *types.QueryCertificate) (*types.QueryCertificateResponse, error) {
	//TODO implement me
	panic("implement me")
}
