package keeper

import (
	"context"
	"encoding/base64"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/did/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) DID(c context.Context, req *types.QueryDIDRequest) (*types.QueryDIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	didBz, err := base64.StdEncoding.DecodeString(req.DidBase64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid did_base64")
	}

	did := string(didBz)
	docWithSeq := k.GetDIDDocument(ctx, did)
	if docWithSeq.Empty() {
		return nil, status.Error(codes.NotFound, "DID not found")
	}
	if docWithSeq.Deactivated() {
		return nil, status.Error(codes.NotFound, "DID deactivated")
	}

	return &types.QueryDIDResponse{DidDocumentWithSeq: &docWithSeq}, nil
}
