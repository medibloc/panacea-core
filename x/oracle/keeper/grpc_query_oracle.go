package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Oracle(goCtx context.Context, req *types.QueryOracleRequest) (*types.QueryOracleResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	accOracleAddress, err := sdk.AccAddressFromBech32(req.GetAddress())
	if err != nil {
		return nil, err
	}

	oracle, err := k.GetOracle(ctx, accOracleAddress)
	if err != nil {
		return nil, err
	}

	return &types.QueryOracleResponse{Oracle: &oracle}, nil
}
