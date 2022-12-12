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
		Deals:       deals,
		Pagination: pageRes,
	}, nil
}

func (k Keeper) Certificates(goCtx context.Context, req *types.QueryCertificates) (*types.QueryCertificatesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	store := ctx.KVStore(k.storeKey)

	certsStore := prefix.NewStore(store, append(types.CertificateKey, sdk.Uint64ToBigEndian(req.DealId)...))

	var certs []*types.Certificate
	pageRes, err := query.Paginate(certsStore, req.Pagination, func(_ []byte, value []byte) error {
		var cert types.Certificate
		err := k.cdc.UnmarshalLengthPrefixed(value, &cert)
		if err != nil {
			return err
		}
		certs = append(certs, &cert)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryCertificatesResponse{
		Certificates: certs,
		Pagination:   pageRes,
	}, nil
}

func (k Keeper) Certificate(goCtx context.Context, req *types.QueryCertificate) (*types.QueryCertificateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	cert, err := k.GetCertificate(sdk.UnwrapSDKContext(goCtx), req.DealId, req.DataHash)
	if err != nil {
		return nil, err
	}

	return &types.QueryCertificateResponse{
		Certificate: cert,
	}, nil
}
