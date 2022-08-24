package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func (k Keeper) Oracle(goCtx context.Context, req *types.QueryOracleRequest) (*types.QueryOracleResponse, error) {
	panic("implements me")
}

// Oracles returns a list of oracles.
func (k Keeper) Oracles(goCtx context.Context, req *types.QueryOraclesRequest) (*types.QueryOraclesResponse, error) {
	panic("implements me")
}

// OracleRegistration returns a OracleRegistration details.
func (k Keeper) OracleRegistration(goCtx context.Context, req *types.QueryOracleRegistrationRequest) (*types.QueryOracleRegistrationResponse, error) {
	oracleRegistration, err := k.GetOracleRegistration(sdk.UnwrapSDKContext(goCtx), req.UniqueId, req.Address)
	if err != nil {
		return nil, err
	}

	return &types.QueryOracleRegistrationResponse{OracleRegistration: oracleRegistration}, nil
}

// Params returns params of oracle module.
func (k Keeper) Params(goCtx context.Context, req *types.QueryOracleParamsRequest) (*types.QueryParamsResponse, error) {
	panic("implements me")
}
