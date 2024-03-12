package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	"github.com/medibloc/panacea-core/v2/x/pnft/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Denoms(goCtx context.Context, request *types.QueryDenomsRequest) (*types.QueryDenomsResponse, error) {
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

	return &types.QueryDenomsResponse{
		Denoms:     denoms,
		Pagination: classRes.Pagination,
	}, nil
}

func (k Keeper) DenomsByOwner(goCtx context.Context, request *types.QueryDenomsByOwnerRequest) (*types.QueryDenomsByOwnerResponse, error) {
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

	return &types.QueryDenomsByOwnerResponse{
		Denoms: denoms,
	}, nil
}

func (k Keeper) Denom(goCtx context.Context, request *types.QueryDenomRequest) (*types.QueryDenomResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	denom, err := k.GetDenom(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	return &types.QueryDenomResponse{
		Denom: denom,
	}, nil
}

func (k Keeper) PNFTs(goCtx context.Context, request *types.QueryPNFTsRequest) (*types.QueryPNFTsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pnfts, err := k.GetPNFTsByDenomId(ctx, request.DenomId)
	if err != nil {
		return nil, err
	}
	return &types.QueryPNFTsResponse{Pnfts: pnfts}, nil
}

func (k Keeper) PNFTsByDenomOwner(goCtx context.Context, request *types.QueryPNFTsByDenomOwnerRequest) (*types.QueryPNFTsByDenomOwnerResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pnfts, err := k.GetPNFTsByDenomIdAndOwner(ctx, request.DenomId, request.Owner)
	if err != nil {
		return nil, err
	}
	return &types.QueryPNFTsByDenomOwnerResponse{Pnfts: pnfts}, nil
}

func (k Keeper) PNFT(goCtx context.Context, request *types.QueryPNFTRequest) (*types.QueryPNFTResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pnft, err := k.GetPNFT(ctx, request.DenomId, request.Id)
	if err != nil {
		return nil, err
	}

	return &types.QueryPNFTResponse{Pnft: pnft}, nil
}
