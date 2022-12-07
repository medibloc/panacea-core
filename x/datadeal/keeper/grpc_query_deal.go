package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

func (k Keeper) Deals(ctx context.Context, request *types.QueryDealsRequest) (*types.QueryDealsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (k Keeper) Certificates(ctx context.Context, certificates *types.QueryCertificates) (*types.QueryCertificatesResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (k Keeper) Certificate(ctx context.Context, certificate *types.QueryCertificate) (*types.QueryCertificateResponse, error) {
	//TODO implement me
	panic("implement me")
}
