package keeper

import (
	"context"
	"errors"
	"fmt"

	upstreamnft "cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ upstreamnft.QueryServer = (*standardQueryServer)(nil)

type standardQueryServer struct {
	upstreamnft.UnimplementedQueryServer
	keeper Keeper
}

// NewStandardQueryServer creates the policy-aware implementation of the
// standard cosmos.nft.v1beta1.Query service.
func NewStandardQueryServer(k Keeper) upstreamnft.QueryServer {
	return &standardQueryServer{keeper: k}
}

// Class returns standard class metadata after verifying its coupled Panacea
// policy state.
func (q standardQueryServer) Class(
	goCtx context.Context,
	request *upstreamnft.QueryClassRequest,
) (*upstreamnft.QueryClassResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := q.keeper.validateCanonicalClassID(request.ClassId); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	record, err := q.keeper.getClassRecord(ctx, request.ClassId)
	if err != nil {
		return nil, mapQueryStateError(err)
	}
	return &upstreamnft.QueryClassResponse{Class: record.Class}, nil
}

// NFT returns standard NFT metadata for an active or revoked NFT.
func (q standardQueryServer) NFT(
	goCtx context.Context,
	request *upstreamnft.QueryNFTRequest,
) (*upstreamnft.QueryNFTResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := q.keeper.validateCanonicalClassID(request.ClassId); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := types.ValidateNFTID(request.Id); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	live, err := q.keeper.getLiveNFTRecord(ctx, request.ClassId, request.Id)
	if err != nil {
		return nil, mapQueryStateError(err)
	}
	return &upstreamnft.QueryNFTResponse{Nft: live.Nft}, nil
}

// Owner returns the current owner of an active or revoked NFT. Burned and
// unissued IDs retain the standard service's empty-owner response.
func (q standardQueryServer) Owner(
	goCtx context.Context,
	request *upstreamnft.QueryOwnerRequest,
) (*upstreamnft.QueryOwnerResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := q.keeper.validateCanonicalClassID(request.ClassId); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := types.ValidateNFTID(request.Id); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	live, err := q.keeper.getLiveNFTRecord(ctx, request.ClassId, request.Id)
	if errors.Is(err, upstreamnft.ErrNFTNotExists) {
		return &upstreamnft.QueryOwnerResponse{}, nil
	}
	if err != nil {
		return nil, mapQueryStateError(err)
	}
	return &upstreamnft.QueryOwnerResponse{Owner: live.Owner}, nil
}

// Supply returns the number of active and revoked NFTs in a class. Missing
// classes retain the standard service's zero response.
func (q standardQueryServer) Supply(
	goCtx context.Context,
	request *upstreamnft.QuerySupplyRequest,
) (*upstreamnft.QuerySupplyResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := q.keeper.validateCanonicalClassID(request.ClassId); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	supply := q.keeper.nftKeeper.GetTotalSupply(ctx, request.ClassId)
	record, err := q.keeper.getClassRecord(ctx, request.ClassId)
	if errors.Is(err, upstreamnft.ErrClassNotExists) {
		if supply == 0 {
			return &upstreamnft.QuerySupplyResponse{}, nil
		}
		return nil, mapQueryStateError(fmt.Errorf(
			"missing class %s has supply %d",
			request.ClassId,
			supply,
		))
	}
	if err != nil {
		return nil, mapQueryStateError(err)
	}
	if supply > record.MintedCount {
		return nil, mapQueryStateError(fmt.Errorf(
			"class %s supply %d exceeds minted count %d",
			request.ClassId,
			supply,
			record.MintedCount,
		))
	}
	return &upstreamnft.QuerySupplyResponse{Amount: supply}, nil
}
