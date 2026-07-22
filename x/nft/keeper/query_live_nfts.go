package keeper

import (
	"errors"
	"fmt"

	upstreamnft "cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
)

// listLiveNFTRecordsByClass expects classID and pagination to have already been
// validated, with pagination normalized by normalizeQueryPageRequest.
func (k Keeper) listLiveNFTRecordsByClass(
	ctx sdk.Context,
	classID string,
	pagination *query.PageRequest,
) ([]*types.LiveNFTRecord, *query.PageResponse, error) {
	return k.listLiveNFTRecords(ctx, classID, "", pagination)
}

// listLiveNFTRecordsByOwner expects owner and any non-empty classID to have
// already been validated, with pagination normalized by
// normalizeQueryPageRequest.
func (k Keeper) listLiveNFTRecordsByOwner(
	ctx sdk.Context,
	classID string,
	owner string,
	pagination *query.PageRequest,
) ([]*types.LiveNFTRecord, *query.PageResponse, error) {
	return k.listLiveNFTRecords(ctx, classID, owner, pagination)
}

func (k Keeper) listLiveNFTRecords(
	ctx sdk.Context,
	classID string,
	owner string,
	pagination *query.PageRequest,
) ([]*types.LiveNFTRecord, *query.PageResponse, error) {
	if classID == "" && owner == "" {
		return nil, nil, fmt.Errorf("list live nfts requires a class or owner")
	}

	response, err := k.nftKeeper.NFTs(
		sdk.WrapSDKContext(ctx),
		&upstreamnft.QueryNFTsRequest{
			ClassId:    classID,
			Owner:      owner,
			Pagination: pagination,
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("list standard nfts: %w", err)
	}
	if response == nil || response.Pagination == nil {
		return nil, nil, fmt.Errorf("standard nft query returned no pagination")
	}

	if classID != "" {
		_, err = k.getClassRecord(ctx, classID)
		if errors.Is(err, upstreamnft.ErrClassNotExists) {
			if len(response.Nfts) == 0 {
				return []*types.LiveNFTRecord{}, response.Pagination, nil
			}
			return nil, nil, fmt.Errorf("listed nfts reference missing class %s", classID)
		}
		if err != nil {
			return nil, nil, fmt.Errorf("load class state for nft list %s: %w", classID, err)
		}
	}

	records := make([]*types.LiveNFTRecord, 0, len(response.Nfts))
	for _, token := range response.Nfts {
		if token == nil {
			return nil, nil, fmt.Errorf("standard nft query returned a nil nft")
		}
		if classID != "" && token.ClassId != classID {
			return nil, nil, fmt.Errorf(
				"listed nft %s has class %s, expected %s",
				token.Id,
				token.ClassId,
				classID,
			)
		}
		if err := k.validateCanonicalClassID(token.ClassId); err != nil {
			return nil, nil, fmt.Errorf(
				"listed nft has invalid stored class ID %q: %w",
				token.ClassId,
				err,
			)
		}
		if err := types.ValidateNFTID(token.Id); err != nil {
			return nil, nil, fmt.Errorf("listed nft has invalid stored ID %q: %w", token.Id, err)
		}
		live, err := k.getLiveNFTRecord(ctx, token.ClassId, token.Id)
		if errors.Is(err, upstreamnft.ErrNFTNotExists) {
			return nil, nil, fmt.Errorf(
				"listed nft %s/%s has no coupled live state",
				token.ClassId,
				token.Id,
			)
		}
		if err != nil {
			return nil, nil, fmt.Errorf(
				"load listed nft %s/%s: %w",
				token.ClassId,
				token.Id,
				err,
			)
		}
		if owner != "" && live.Owner != owner {
			return nil, nil, fmt.Errorf(
				"listed nft %s/%s has owner %s, expected %s",
				token.ClassId,
				token.Id,
				live.Owner,
				owner,
			)
		}
		records = append(records, live)
	}
	return records, response.Pagination, nil
}
