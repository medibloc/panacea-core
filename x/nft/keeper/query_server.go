package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type queryServer struct {
	types.UnimplementedQueryServer
	keeper Keeper
}

// NewQueryServer creates Panacea's combined NFT query server.
func NewQueryServer(k Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

// ClassRecord returns standard class metadata and Panacea policy state.
func (q queryServer) ClassRecord(
	goCtx context.Context,
	request *types.QueryClassRecordRequest,
) (*types.QueryClassRecordResponse, error) {
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
	return &types.QueryClassRecordResponse{ClassRecord: record}, nil
}

// NFTRecord returns one live NFT or its permanent burn tombstone.
func (q queryServer) NFTRecord(
	goCtx context.Context,
	request *types.QueryNFTRecordRequest,
) (*types.QueryNFTRecordResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := q.keeper.validateCanonicalClassID(request.ClassId); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := types.ValidateNFTID(request.NftId); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	record, err := q.keeper.getNFTRecord(ctx, request.ClassId, request.NftId)
	if err != nil {
		return nil, mapQueryStateError(err)
	}
	return &types.QueryNFTRecordResponse{NftRecord: record}, nil
}
