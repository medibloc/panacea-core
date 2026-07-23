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
		ctx,
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

	cache := newLiveNFTListCache()
	if classID != "" {
		_, err = cache.classRecord(k, ctx, classID)
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
		live, err := k.getLiveNFTRecordForList(ctx, token.ClassId, token.Id, cache)
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

type liveNFTListClassState struct {
	record      *types.ClassRecord
	supply      uint64
	supplyKnown bool
}

type liveNFTListCache struct {
	classes map[string]*liveNFTListClassState
}

func newLiveNFTListCache() *liveNFTListCache {
	return &liveNFTListCache{classes: make(map[string]*liveNFTListClassState)}
}

func (c *liveNFTListCache) classRecord(
	k Keeper,
	ctx sdk.Context,
	classID string,
) (*types.ClassRecord, error) {
	if cached, exists := c.classes[classID]; exists {
		return cached.record, nil
	}
	record, err := k.getClassRecord(ctx, classID)
	if err != nil {
		return nil, err
	}
	c.classes[classID] = &liveNFTListClassState{record: record}
	return record, nil
}

func (c *liveNFTListCache) classSupply(
	k Keeper,
	ctx sdk.Context,
	classID string,
) (uint64, error) {
	cached, exists := c.classes[classID]
	if !exists {
		if _, err := c.classRecord(k, ctx, classID); err != nil {
			return 0, err
		}
		cached = c.classes[classID]
	}
	if !cached.supplyKnown {
		cached.supply = k.nftKeeper.GetTotalSupply(ctx, classID)
		cached.supplyKnown = true
	}
	return cached.supply, nil
}

func (k Keeper) getLiveNFTRecordForList(
	ctx sdk.Context,
	classID string,
	nftID string,
	cache *liveNFTListCache,
) (*types.LiveNFTRecord, error) {
	state, err := k.loadLiveNFTStateOnly(ctx, classID, nftID)
	if err != nil {
		return nil, err
	}
	classRecord, err := cache.classRecord(k, ctx, classID)
	if errors.Is(err, upstreamnft.ErrClassNotExists) {
		return nil, fmt.Errorf(
			"live nft %s in class %s references missing class state",
			nftID,
			classID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("load class state for nft %s: %w", nftID, err)
	}
	liveSupply, err := cache.classSupply(k, ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("load supply for nft %s: %w", nftID, err)
	}
	return k.buildLiveNFTRecord(ctx, classID, nftID, state, classRecord, liveSupply)
}
