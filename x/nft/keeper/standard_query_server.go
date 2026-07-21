package keeper

import (
	"context"

	upstreamnft "cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
