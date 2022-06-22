package keeper

import (
	"context"
	"encoding/base64"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/types/query"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Deal(goCtx context.Context, req *types.QueryDealRequest) (*types.QueryDealResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	deal, err := k.GetDeal(ctx, req.DealId)
	if err != nil {
		return nil, err
	}

	return &types.QueryDealResponse{Deal: &deal}, nil
}

func (k Keeper) Deals(goCtx context.Context, req *types.QueryDealsRequest) (*types.QueryDealsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	store := ctx.KVStore(k.storeKey)
	dealsStore := prefix.NewStore(store, types.KeyPrefixDeals)

	var deals []types.Deal
	pageRes, err := query.Paginate(dealsStore, req.Pagination, func(_ []byte, value []byte) error {
		var deal types.Deal
		err := k.cdc.UnmarshalLengthPrefixed(value, &deal)
		if err != nil {
			return err
		}
		deals = append(deals, deal)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDealsResponse{
		Deals:      deals,
		Pagination: pageRes,
	}, nil
}

func (k Keeper) DataCert(goCtx context.Context, req *types.QueryDataCertRequest) (*types.QueryDataCertResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	bz, err := base64.StdEncoding.DecodeString(req.DataHash)
	if err != nil {
		return nil, err
	}

	dataCert, err := k.GetDataCert(ctx, req.DealId, bz)
	if err != nil {
		return nil, err
	}

	return &types.QueryDataCertResponse{DataCert: &dataCert}, nil
}

func (k Keeper) DataCerts(goCtx context.Context, req *types.QueryDataCertsRequest) (*types.QueryDataCertsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	store := ctx.KVStore(k.storeKey)
	certsStore := prefix.NewStore(store, types.GetKeyPrefixDataCertsByDealID(req.DealId))

	var certs []types.DataCert

	pageRes, err := query.Paginate(certsStore, req.Pagination, func(_ []byte, value []byte) error {
		var cert types.DataCert
		err := k.cdc.UnmarshalLengthPrefixed(value, &cert)
		if err != nil {
			return err
		}
		certs = append(certs, cert)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDataCertsResponse{
		DataCerts:  certs,
		Pagination: pageRes,
	}, nil
}
