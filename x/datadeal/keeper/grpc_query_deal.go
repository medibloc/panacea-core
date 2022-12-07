package keeper

import (
	"context"

	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

func (k Keeper) Deals(ctx context.Context, request *types.QueryDealsRequest) (*types.QueryDealsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (k Keeper) Deal(ctx context.Context, request *types.QueryDealRequest) (*types.QueryDealResponse, error) {
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
