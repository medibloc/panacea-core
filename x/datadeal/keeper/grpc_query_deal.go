package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

func (k Keeper) Deal(goCtx context.Context, req *types.QueryDealRequest) (*types.QueryDealResponse, error) {
	panic("implements me")
}

func (k Keeper) Deals(goCtx context.Context, req *types.QueryDealsRequest) (*types.QueryDealsResponse, error) {
	panic("implements me")
}

func (k Keeper) DataSale(goCtx context.Context, req *types.QueryDataSaleRequest) (*types.QueryDataSaleResponse, error) {
	dataSale, err := k.GetDataSale(sdk.UnwrapSDKContext(goCtx), req.VerifiableCid, req.DealId)
	if err != nil {
		return nil, err
	}

	return &types.QueryDataSaleResponse{
		DataSale: dataSale,
	}, nil
}
