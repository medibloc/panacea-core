package keeper

import (
	"context"

	"github.com/medibloc/panacea-core/v2/x/datapool/types"
)

func (k Keeper) Pool(goCtx context.Context, req *types.QueryPoolRequest) (*types.QueryPoolResponse, error) {
	return &types.QueryPoolResponse{}, nil
}
