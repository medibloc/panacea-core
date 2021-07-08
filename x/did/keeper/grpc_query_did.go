package keeper

import (
	"context"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/did/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) DIDDocumentWithSeq(c context.Context, req *types.QueryDIDRequest) (*types.QueryDIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	did := req.Did
	docWithSeq := k.GetDIDDocument(ctx, did)
	if docWithSeq.Empty() {
		return &types.QueryDIDResponse{}, sdkerrors.Wrapf(types.ErrDIDNotFound, "DID: %s", did)
	}
	if docWithSeq.Deactivated() {
		return &types.QueryDIDResponse{}, sdkerrors.Wrapf(types.ErrDIDDeactivated, "DID: %s", did)
	}

	return &types.QueryDIDResponse{DidDocumentWithSeq: &docWithSeq}, nil
}
