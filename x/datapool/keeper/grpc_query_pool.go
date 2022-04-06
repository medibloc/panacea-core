package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/medibloc/panacea-core/v2/x/datapool/types"
)

func (k Keeper) Pool(goCtx context.Context, req *types.QueryPoolRequest) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pool, err := k.GetPool(ctx, req.PoolId)
	if err != nil {
		return nil, err
	}

	return &types.QueryPoolResponse{
		Pool: pool,
	}, nil
}

func (k Keeper) NFTContract(goCtx context.Context, req *types.QueryNFTContractRequest) (*types.QueryNFTContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	contract, err := k.GetNFTContractAddress(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryNFTContractResponse{NftContractAddress: contract.String()}, nil
}

func (k Keeper) DataValidator(goCtx context.Context, req *types.QueryDataValidatorRequest) (*types.QueryDataValidatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	accValidatorAddress, err := sdk.AccAddressFromBech32(req.GetAddress())
	if err != nil {
		return nil, err
	}

	dataValidator, err := k.GetDataValidator(ctx, accValidatorAddress)
	if err != nil {
		return nil, err
	}

	return &types.QueryDataValidatorResponse{DataValidator: &dataValidator}, nil
}
