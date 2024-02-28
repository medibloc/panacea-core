package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/aol/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Record(c context.Context, req *types.QueryServiceRecordRequest) (*types.QueryServiceRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	ownerAddr, err := sdk.AccAddressFromBech32(req.OwnerAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner address")
	}

	recordKey := types.RecordCompositeKey{OwnerAddress: ownerAddr, TopicName: req.TopicName, Offset: req.Offset}
	if !k.HasRecord(ctx, recordKey) {
		return nil, status.Error(codes.NotFound, "record not found")
	}

	record := k.GetRecord(ctx, recordKey)
	return &types.QueryServiceRecordResponse{Record: &record}, nil
}
