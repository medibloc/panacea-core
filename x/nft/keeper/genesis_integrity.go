package keeper

import (
	"bytes"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	upstreamnft "cosmossdk.io/x/nft"
	upstreamkeeper "cosmossdk.io/x/nft/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
)

func (k Keeper) verifyExportedPhysicalState(ctx sdk.Context, data types.GenesisState) error {
	if err := k.verifyExportedStandardKeys(ctx, data); err != nil {
		return err
	}
	return k.verifyExportedOwnerIndexes(ctx, data)
}

func (k Keeper) verifyExportedStandardKeys(ctx sdk.Context, data types.GenesisState) error {
	for _, expected := range data.NftState.Classes {
		actual, found := k.nftKeeper.GetClass(ctx, expected.Id)
		if !found {
			return fmt.Errorf("exported class %s does not match its physical store key", expected.Id)
		}
		actualBytes, err := k.cdc.Marshal(&actual)
		if err != nil {
			return fmt.Errorf("encode stored class %s: %w", expected.Id, err)
		}
		expectedBytes, err := k.cdc.Marshal(expected)
		if err != nil {
			return fmt.Errorf("encode exported class %s: %w", expected.Id, err)
		}
		if !bytes.Equal(actualBytes, expectedBytes) {
			return fmt.Errorf("exported class %s does not match standard store state", expected.Id)
		}
	}
	for _, entry := range data.NftState.Entries {
		for _, expected := range entry.Nfts {
			actual, found := k.nftKeeper.GetNFT(ctx, expected.ClassId, expected.Id)
			if !found {
				return fmt.Errorf(
					"exported nft %s/%s does not match its physical store key",
					expected.ClassId,
					expected.Id,
				)
			}
			actualBytes, err := k.cdc.Marshal(&actual)
			if err != nil {
				return fmt.Errorf("encode stored nft %s/%s: %w", expected.ClassId, expected.Id, err)
			}
			expectedBytes, err := k.cdc.Marshal(expected)
			if err != nil {
				return fmt.Errorf("encode exported nft %s/%s: %w", expected.ClassId, expected.Id, err)
			}
			if !bytes.Equal(actualBytes, expectedBytes) {
				return fmt.Errorf(
					"exported nft %s/%s does not match standard store state",
					expected.ClassId,
					expected.Id,
				)
			}
		}
	}
	var liveNFTCount uint64
	for _, entry := range data.NftState.Entries {
		liveNFTCount += uint64(len(entry.Nfts))
	}
	if err := k.verifyUpstreamPrefixCount(
		ctx,
		upstreamkeeper.ClassKey,
		uint64(len(data.NftState.Classes)),
		"standard class",
	); err != nil {
		return err
	}
	if err := k.verifyUpstreamPrefixCount(
		ctx,
		upstreamkeeper.NFTKey,
		liveNFTCount,
		"standard nft",
	); err != nil {
		return err
	}
	if err := k.verifyUpstreamPrefixCount(
		ctx,
		upstreamkeeper.OwnerKey,
		liveNFTCount,
		"standard nft direct owner",
	); err != nil {
		return err
	}
	return nil
}

func (k Keeper) verifyExportedOwnerIndexes(ctx sdk.Context, data types.GenesisState) error {
	const pageLimit = uint64(100)
	for _, entry := range data.NftState.Entries {
		expected := make(map[genesisNFTKey]*upstreamnft.NFT, len(entry.Nfts))
		for _, token := range entry.Nfts {
			expected[genesisNFTKey{classID: token.ClassId, nftID: token.Id}] = token
		}
		seen := make(map[genesisNFTKey]struct{}, len(entry.Nfts))
		var nextKey []byte
		firstPage := true
		for {
			response, err := k.nftKeeper.NFTs(
				sdk.WrapSDKContext(ctx),
				&upstreamnft.QueryNFTsRequest{
					Owner: entry.Owner,
					Pagination: &query.PageRequest{
						Key:        nextKey,
						Limit:      pageLimit,
						CountTotal: firstPage,
					},
				},
			)
			if err != nil {
				return fmt.Errorf("query owner index for %s: %w", entry.Owner, err)
			}
			if response.Pagination == nil {
				return fmt.Errorf("owner index query for %s returned no pagination", entry.Owner)
			}
			if firstPage && response.Pagination.Total != uint64(len(expected)) {
				return fmt.Errorf(
					"owner %s reverse index count %d does not match direct owner count %d",
					entry.Owner,
					response.Pagination.Total,
					len(expected),
				)
			}
			for _, actual := range response.Nfts {
				key := genesisNFTKey{classID: actual.ClassId, nftID: actual.Id}
				expectedToken, exists := expected[key]
				if !exists {
					return fmt.Errorf(
						"owner %s reverse index contains unexpected nft %s/%s",
						entry.Owner,
						actual.ClassId,
						actual.Id,
					)
				}
				if _, duplicate := seen[key]; duplicate {
					return fmt.Errorf(
						"owner %s reverse index contains duplicate nft %s/%s",
						entry.Owner,
						actual.ClassId,
						actual.Id,
					)
				}
				actualBytes, err := k.cdc.Marshal(actual)
				if err != nil {
					return fmt.Errorf("encode indexed nft %s/%s: %w", actual.ClassId, actual.Id, err)
				}
				expectedBytes, err := k.cdc.Marshal(expectedToken)
				if err != nil {
					return fmt.Errorf("encode exported nft %s/%s: %w", actual.ClassId, actual.Id, err)
				}
				if !bytes.Equal(actualBytes, expectedBytes) {
					return fmt.Errorf(
						"owner %s reverse index nft %s/%s does not match standard state",
						entry.Owner,
						actual.ClassId,
						actual.Id,
					)
				}
				seen[key] = struct{}{}
			}
			if len(response.Pagination.NextKey) == 0 {
				break
			}
			nextKey = response.Pagination.NextKey
			firstPage = false
		}
		if len(seen) != len(expected) {
			return fmt.Errorf(
				"owner %s reverse index contains %d nfts, expected %d",
				entry.Owner,
				len(seen),
				len(expected),
			)
		}
	}
	var expectedCount uint64
	for _, entry := range data.NftState.Entries {
		expectedCount += uint64(len(entry.Nfts))
	}
	return k.verifyUpstreamPrefixCount(
		ctx,
		upstreamkeeper.NFTOfClassByOwnerKey,
		expectedCount,
		"standard nft reverse owner index",
	)
}

func (k Keeper) verifyUpstreamPrefixCount(
	ctx sdk.Context,
	prefix []byte,
	expectedCount uint64,
	name string,
) error {
	// Upstream exports these prefixes, so inspect them directly instead of
	// copying private key encodings. Point and owner-query checks prove expected
	// membership; the global counts detect orphan or stale physical keys.
	store := k.nftStoreService.OpenKVStore(ctx)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return fmt.Errorf("iterate %s keys: %w", name, err)
	}
	defer iterator.Close()
	var actualCount uint64
	for ; iterator.Valid(); iterator.Next() {
		actualCount++
	}
	if err := iterator.Error(); err != nil {
		return fmt.Errorf("iterate %s keys: %w", name, err)
	}
	if actualCount != expectedCount {
		return fmt.Errorf(
			"%s key count %d does not match expected count %d",
			name,
			actualCount,
			expectedCount,
		)
	}
	return nil
}
