package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/did/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) DIDDocumentWithSeq(c context.Context, req *types.QueryGetDIDRequest) (*types.QueryGetDIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	did := req.DID
	docWithSeq := k.GetDIDDocument(ctx, did)
	if docWithSeq.Empty() {
		return &types.QueryGetDIDResponse{}, types.ErrorWrapf(types.ErrDIDNotFound, "DID: %s", did)
	}
	if docWithSeq.Deactivated() {
		return &types.QueryGetDIDResponse{}, types.ErrorWrapf(types.ErrDIDDeactivated, "DID: %s", did)
	}

	return &types.QueryGetDIDResponse{DIDDocumentWithSeq: &docWithSeq}, nil
}
