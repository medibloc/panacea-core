package keeper

import (
	"context"
	"errors"

	upstreamnft "cosmossdk.io/x/nft"
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
	if errors.Is(err, upstreamnft.ErrClassNotExists) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryClassRecordResponse{ClassRecord: record}, nil
}
