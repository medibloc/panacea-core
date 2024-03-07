package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	"github.com/medibloc/panacea-core/v2/x/pnft/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServiceServer = Keeper{}

func (k Keeper) Denoms(goCtx context.Context, request *types.QueryServiceDenomsRequest) (*types.QueryServiceDenomsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	classRes, err := k.nftKeeper.Classes(goCtx, &nft.QueryClassesRequest{
		Pagination: request.Pagination,
	})
	if err != nil {
		return nil, err
	}

	denoms, err := k.ParseDenoms(classRes.GetClasses())
	if err != nil {
		return nil, err
	}

	return &types.QueryServiceDenomsResponse{
		Denoms:     denoms,
		Pagination: classRes.Pagination,
	}, nil
}

func (k Keeper) DenomsByCreator(goCtx context.Context, request *types.QueryServiceDenomsByCreatorRequest) (*types.QueryServiceDenomsByCreatorResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	classRes, err := k.nftKeeper.Classes(goCtx, &nft.QueryClassesRequest{})
	if err != nil {
		return nil, err
	}

	denoms, err := k.ParseDenoms(classRes.Classes)
	if err != nil {
		return nil, err
	}

	return &types.QueryServiceDenomsByCreatorResponse{
		Denoms: denoms,
	}, nil
}

func (k Keeper) Denom(goCtx context.Context, request *types.QueryServiceDenomRequest) (*types.QueryServiceDenomResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	denom, err := k.GetDenom(ctx, request.GetId())
	if err != nil {
		return nil, err
	}

	return &types.QueryServiceDenomResponse{
		Denom: denom,
	}, nil
}
