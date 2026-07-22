package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NFTRecords returns active and revoked NFT records selected by class, owner,
// or their intersection. Burn tombstones remain available only by point query.
func (q queryServer) NFTRecords(
	goCtx context.Context,
	request *types.QueryNFTRecordsRequest,
) (*types.QueryNFTRecordsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if request.ClassId == "" && request.Owner == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"must provide at least one of class_id or owner",
		)
	}
	if request.ClassId != "" {
		if err := q.keeper.validateCanonicalClassID(request.ClassId); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	owner := ""
	if request.Owner != "" {
		var err error
		owner, _, err = q.keeper.canonicalAddress("owner", request.Owner)
		if err != nil {
			if errors.Is(err, sdkerrors.ErrInvalidAddress) {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			return nil, mapQueryStateError(err)
		}
	}
	pagination, err := normalizeQueryPageRequest(request.Pagination)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	var records []*types.LiveNFTRecord
	var page *query.PageResponse
	if owner == "" {
		records, page, err = q.keeper.listLiveNFTRecordsByClass(
			ctx,
			request.ClassId,
			pagination,
		)
	} else {
		records, page, err = q.keeper.listLiveNFTRecordsByOwner(
			ctx,
			request.ClassId,
			owner,
			pagination,
		)
	}
	if err != nil {
		return nil, mapQueryStateError(err)
	}

	return &types.QueryNFTRecordsResponse{
		NftRecords: records,
		Pagination: page,
	}, nil
}
