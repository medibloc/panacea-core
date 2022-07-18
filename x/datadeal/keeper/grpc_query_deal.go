package keeper

import (
	"context"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

func (k Keeper) Deal(goCtx context.Context, req *types.QueryDealRequest) (*types.QueryDealResponse, error) {
	panic("implements me")
}

func (k Keeper) Deals(goCtx context.Context, req *types.QueryDealsRequest) (*types.QueryDealsResponse, error) {
	panic("implements me")
}

func (k Keeper) DataSale(goCtx context.Context, req *types.QueryDataSaleRequest) (*types.QueryDataSaleResponse, error) {
	panic("implements me")
}
