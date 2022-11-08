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
	dealsStore := prefix.NewStore(store, types.KeyPrefixDeals)

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

func (k Keeper) DataSale(goCtx context.Context, req *types.QueryDataSaleRequest) (*types.QueryDataSaleResponse, error) {
	dataSale, err := k.GetDataSale(sdk.UnwrapSDKContext(goCtx), req.DataHash, req.DealId)
	if err != nil {
		return nil, err
	}

	return &types.QueryDataSaleResponse{
		DataSale: dataSale,
	}, nil
}

func (k Keeper) DataSales(goCtx context.Context, req *types.QueryDataSalesRequest) (*types.QueryDataSalesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	store := ctx.KVStore(k.storeKey)
	dataSalesStore := prefix.NewStore(store, types.GetDataSalesKey(req.DealId))

	var dataSales []*types.DataSale
	pageRes, err := query.Paginate(dataSalesStore, req.Pagination, func(_, value []byte) error {
		var dataSale types.DataSale
		err := k.cdc.UnmarshalLengthPrefixed(value, &dataSale)
		if err != nil {
			return err
		}
		dataSales = append(dataSales, &dataSale)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDataSalesResponse{
		DataSale:   dataSales,
		Pagination: pageRes,
	}, nil
}

func (k Keeper) DataVerificationVote(goCtx context.Context, req *types.QueryDataVerificationVoteRequest) (*types.QueryDataVerificationVoteResponse, error) {
	dataVerificationVote, err := k.GetDataVerificationVote(sdk.UnwrapSDKContext(goCtx), req.DataHash, req.VoterAddress, req.DealId)
	if err != nil {
		return nil, err
	}

	return &types.QueryDataVerificationVoteResponse{
		DataVerificationVote: dataVerificationVote,
	}, nil
}

func (k Keeper) DataDeliveryVote(goCtx context.Context, req *types.QueryDataDeliveryVoteRequest) (*types.QueryDataDeliveryVoteResponse, error) {
	dataDeliveryVote, err := k.GetDataDeliveryVote(sdk.UnwrapSDKContext(goCtx), req.DataHash, req.VoterAddress, req.DealId)
	if err != nil {
		return nil, err
	}

	return &types.QueryDataDeliveryVoteResponse{
		DataDeliveryVote: dataDeliveryVote,
	}, nil
}
