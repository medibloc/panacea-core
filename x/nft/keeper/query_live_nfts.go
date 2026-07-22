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
	response, err := k.nftKeeper.NFTs(
		sdk.WrapSDKContext(ctx),
		&upstreamnft.QueryNFTsRequest{
			ClassId:    classID,
			Pagination: pagination,
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("list standard nfts for class %s: %w", classID, err)
	}
	if response == nil || response.Pagination == nil {
		return nil, nil, fmt.Errorf("standard nft query for class %s returned no pagination", classID)
	}

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

	records := make([]*types.LiveNFTRecord, 0, len(response.Nfts))
	for _, token := range response.Nfts {
		if token == nil {
			return nil, nil, fmt.Errorf("standard nft query for class %s returned a nil nft", classID)
		}
		if token.ClassId != classID {
			return nil, nil, fmt.Errorf(
				"listed nft %s has class %s, expected %s",
				token.Id,
				token.ClassId,
				classID,
			)
		}
		if err := types.ValidateNFTID(token.Id); err != nil {
			return nil, nil, fmt.Errorf("listed nft has invalid stored ID %q: %w", token.Id, err)
		}
		live, err := k.getLiveNFTRecord(ctx, classID, token.Id)
		if errors.Is(err, upstreamnft.ErrNFTNotExists) {
			return nil, nil, fmt.Errorf("listed nft %s/%s has no coupled live state", classID, token.Id)
		}
		if err != nil {
			return nil, nil, fmt.Errorf("load listed nft %s/%s: %w", classID, token.Id, err)
		}
		records = append(records, live)
	}
	return records, response.Pagination, nil
}
