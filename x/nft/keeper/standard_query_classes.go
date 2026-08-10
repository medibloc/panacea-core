package keeper

import (
	"context"
	"errors"
	"fmt"

	upstreamnft "cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Classes returns standard class metadata in deterministic class ID order.
func (q standardQueryServer) Classes(
	goCtx context.Context,
	request *upstreamnft.QueryClassesRequest,
) (*upstreamnft.QueryClassesResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	pagination, err := normalizeQueryPageRequest(request.Pagination)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	response, err := q.keeper.nftKeeper.Classes(
		goCtx,
		&upstreamnft.QueryClassesRequest{Pagination: pagination},
	)
	if err != nil {
		return nil, mapQueryStateError(err)
	}
	if response == nil || response.Pagination == nil {
		return nil, mapQueryStateError(fmt.Errorf("standard class query returned no pagination"))
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	classes := make([]*upstreamnft.Class, 0, len(response.Classes))
	for _, class := range response.Classes {
		if class == nil {
			return nil, mapQueryStateError(fmt.Errorf("standard class query returned a nil class"))
		}
		if err := q.keeper.validateCanonicalClassID(class.Id); err != nil {
			return nil, mapQueryStateError(fmt.Errorf(
				"listed class %q has an invalid stored ID: %w",
				class.Id,
				err,
			))
		}
		record, err := q.keeper.getClassRecord(ctx, class.Id)
		if errors.Is(err, upstreamnft.ErrClassNotExists) {
			return nil, mapQueryStateError(fmt.Errorf(
				"listed class %s has no coupled class state",
				class.Id,
			))
		}
		if err != nil {
			return nil, mapQueryStateError(err)
		}
		classes = append(classes, record.Class)
	}

	return &upstreamnft.QueryClassesResponse{
		Classes:    classes,
		Pagination: response.Pagination,
	}, nil
}
