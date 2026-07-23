package keeper

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	upstreamnft "cosmossdk.io/x/nft"
	upstreamkeeper "cosmossdk.io/x/nft/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func BenchmarkMaximumPageListQueryPageCacheCandidates(b *testing.B) {
	for _, filter := range maximumPageNFTFilters() {
		b.Run(filter.name, func(b *testing.B) {
			fixture := newMaximumPageNFTFixture(b, nil)
			classID, owner := filter.values(fixture)
			pagination := &query.PageRequest{Limit: maximumQueryPageLimit}

			b.Run("current", func(b *testing.B) {
				records, _, err := fixture.keeper.listLiveNFTRecords(
					fixture.ctx,
					classID,
					owner,
					pagination,
				)
				require.NoError(b, err)
				require.Len(b, records, int(maximumQueryPageLimit))

				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					records, _, err = fixture.keeper.listLiveNFTRecords(
						fixture.ctx,
						classID,
						owner,
						pagination,
					)
					if err != nil {
						b.Fatal(err)
					}
					if len(records) != int(maximumQueryPageLimit) {
						b.Fatalf(
							"records length %d, expected %d",
							len(records),
							maximumQueryPageLimit,
						)
					}
				}
			})

			b.Run("page_class_supply_cache", func(b *testing.B) {
				records, _, err := listLiveNFTRecordsWithPageCache(
					fixture.keeper,
					fixture.ctx,
					classID,
					owner,
					pagination,
				)
				require.NoError(b, err)
				require.Len(b, records, int(maximumQueryPageLimit))

				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					records, _, err = listLiveNFTRecordsWithPageCache(
						fixture.keeper,
						fixture.ctx,
						classID,
						owner,
						pagination,
					)
					if err != nil {
						b.Fatal(err)
					}
					if len(records) != int(maximumQueryPageLimit) {
						b.Fatalf(
							"records length %d, expected %d",
							len(records),
							maximumQueryPageLimit,
						)
					}
				}
			})
		})
	}
}

func TestMaximumPageListQueryPageCacheStoreReads(t *testing.T) {
	counters := &queryStoreCounters{}
	fixture := newMaximumPageNFTFixture(t, counters)
	for _, filter := range maximumPageNFTFilters() {
		t.Run(filter.name, func(t *testing.T) {
			classID, owner := filter.values(fixture)
			expected, _, err := fixture.keeper.listLiveNFTRecords(
				fixture.ctx,
				classID,
				owner,
				&query.PageRequest{Limit: maximumQueryPageLimit},
			)
			require.NoError(t, err)
			counters.reset()

			records, page, err := listLiveNFTRecordsWithPageCache(
				fixture.keeper,
				fixture.ctx,
				classID,
				owner,
				&query.PageRequest{Limit: maximumQueryPageLimit},
			)
			require.NoError(t, err)
			require.Len(t, records, int(maximumQueryPageLimit))
			require.Equal(t, expected, records)
			require.NotEmpty(t, page.NextKey)
			require.Equal(t, filter.pageCacheNFTReads, counters.nft)
			require.Equal(t, filter.pageCachePolicyReads, counters.policy)
		})
	}
}

func TestListedValuesCannotReplacePhysicalKeyValidation(t *testing.T) {
	t.Run("class", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, _ := createClassForMintTest(
			t,
			&fixture,
			sdk.AccAddress(bytes.Repeat([]byte{115}, 20)),
			10,
		)
		class, found := fixture.keeper.nftKeeper.GetClass(fixture.ctx, classID)
		require.True(t, found)
		class.Id = classID + "different"
		value, err := fixture.cdc.Marshal(&class)
		require.NoError(t, err)
		key := append(append([]byte(nil), upstreamkeeper.ClassKey...), classID...)
		require.NoError(t, fixture.keeper.nftStoreService.OpenKVStore(fixture.ctx).Set(
			key,
			value,
		))

		upstream, err := fixture.keeper.nftKeeper.Classes(
			sdk.WrapSDKContext(fixture.ctx),
			&upstreamnft.QueryClassesRequest{
				Pagination: &query.PageRequest{Limit: 1},
			},
		)
		require.NoError(t, err)
		require.Len(t, upstream.Classes, 1)
		require.Equal(t, class.Id, upstream.Classes[0].Id)

		_, err = NewStandardQueryServer(fixture.keeper).Classes(
			sdk.WrapSDKContext(fixture.ctx),
			&upstreamnft.QueryClassesRequest{
				Pagination: &query.PageRequest{Limit: 1},
			},
		)
		require.Equal(t, codes.Internal, status.Code(err))
		require.ErrorContains(t, err, "has no coupled class state")
	})

	t.Run("nft", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, _, _, _, _ := createNFTForBurnTest(t, &fixture)
		token, found := fixture.keeper.nftKeeper.GetNFT(fixture.ctx, classID, "nft-1")
		require.True(t, found)
		token.Id = "different"
		value, err := fixture.cdc.Marshal(&token)
		require.NoError(t, err)
		key := make(
			[]byte,
			0,
			len(upstreamkeeper.NFTKey)+
				len(classID)+
				len(upstreamkeeper.Delimiter)+
				len("nft-1"),
		)
		key = append(key, upstreamkeeper.NFTKey...)
		key = append(key, classID...)
		key = append(key, upstreamkeeper.Delimiter...)
		key = append(key, "nft-1"...)
		require.NoError(t, fixture.keeper.nftStoreService.OpenKVStore(fixture.ctx).Set(
			key,
			value,
		))

		upstream, err := fixture.keeper.nftKeeper.NFTs(
			sdk.WrapSDKContext(fixture.ctx),
			&upstreamnft.QueryNFTsRequest{
				ClassId:    classID,
				Pagination: &query.PageRequest{Limit: 1},
			},
		)
		require.NoError(t, err)
		require.Len(t, upstream.Nfts, 1)
		require.Equal(t, token.Id, upstream.Nfts[0].Id)

		_, err = NewStandardQueryServer(fixture.keeper).NFTs(
			sdk.WrapSDKContext(fixture.ctx),
			&upstreamnft.QueryNFTsRequest{
				ClassId:    classID,
				Pagination: &query.PageRequest{Limit: 1},
			},
		)
		require.Equal(t, codes.Internal, status.Code(err))
		require.ErrorContains(t, err, "has no coupled live state")
	})
}

func TestListQueryPageCacheMatchesCurrentAcrossClasses(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	counters := &queryStoreCounters{}
	fixture.keeper = newQueryCountingKeeper(fixture, counters)
	firstClassID, firstController, owner := createNFTsForQueryTest(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{116}, 20)),
		"alpha",
		"beta",
	)
	secondClassID, _, secondOwner := createNFTsForQueryTest(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{117}, 20)),
		"delta",
		"gamma",
	)
	require.Equal(t, owner, secondOwner)
	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	_, err := NewMsgServer(fixture.keeper).Revoke(
		sdk.WrapSDKContext(fixture.ctx),
		&nfttypes.MsgRevokeRequest{
			ClassId:    firstClassID,
			NftId:      "alpha",
			Controller: firstController,
		},
	)
	require.NoError(t, err)

	pagination := &query.PageRequest{Limit: maximumQueryPageLimit}
	expected, expectedPage, err := fixture.keeper.listLiveNFTRecordsByOwner(
		fixture.ctx,
		"",
		owner,
		pagination,
	)
	require.NoError(t, err)
	counters.reset()
	actual, actualPage, err := listLiveNFTRecordsWithPageCache(
		fixture.keeper,
		fixture.ctx,
		"",
		owner,
		pagination,
	)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
	require.Equal(t, expectedPage, actualPage)
	require.ElementsMatch(t, []string{
		firstClassID + "/alpha",
		firstClassID + "/beta",
		secondClassID + "/delta",
		secondClassID + "/gamma",
	}, liveNFTRecordKeys(actual))
	require.Equal(t, storeReadCounters{
		gets:          16,
		iterators:     1,
		iteratorNexts: 4,
	}, counters.nft)
	require.Equal(t, storeReadCounters{gets: 12}, counters.policy)
}

type pageClassState struct {
	record      *nfttypes.ClassRecord
	supply      uint64
	supplyKnown bool
}

type liveNFTPageCache struct {
	classes map[string]*pageClassState
}

func newLiveNFTPageCache() *liveNFTPageCache {
	return &liveNFTPageCache{classes: make(map[string]*pageClassState)}
}

func (c *liveNFTPageCache) classRecord(
	keeper Keeper,
	ctx sdk.Context,
	classID string,
) (*nfttypes.ClassRecord, error) {
	if cached, exists := c.classes[classID]; exists {
		return cached.record, nil
	}
	record, err := keeper.getClassRecord(ctx, classID)
	if err != nil {
		return nil, err
	}
	c.classes[classID] = &pageClassState{record: record}
	return record, nil
}

func (c *liveNFTPageCache) classSupply(
	keeper Keeper,
	ctx sdk.Context,
	classID string,
) (uint64, error) {
	cached, exists := c.classes[classID]
	if !exists {
		_, err := c.classRecord(keeper, ctx, classID)
		if err != nil {
			return 0, err
		}
		cached = c.classes[classID]
	}
	if !cached.supplyKnown {
		cached.supply = keeper.nftKeeper.GetTotalSupply(ctx, classID)
		cached.supplyKnown = true
	}
	return cached.supply, nil
}

func listLiveNFTRecordsWithPageCache(
	keeper Keeper,
	ctx sdk.Context,
	classID string,
	owner string,
	pagination *query.PageRequest,
) ([]*nfttypes.LiveNFTRecord, *query.PageResponse, error) {
	if classID == "" && owner == "" {
		return nil, nil, fmt.Errorf("list live nfts requires a class or owner")
	}
	response, err := keeper.nftKeeper.NFTs(
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

	cache := newLiveNFTPageCache()
	if classID != "" {
		_, err = cache.classRecord(keeper, ctx, classID)
		if errors.Is(err, upstreamnft.ErrClassNotExists) {
			if len(response.Nfts) == 0 {
				return []*nfttypes.LiveNFTRecord{}, response.Pagination, nil
			}
			return nil, nil, fmt.Errorf("listed nfts reference missing class %s", classID)
		}
		if err != nil {
			return nil, nil, fmt.Errorf("load class state for nft list %s: %w", classID, err)
		}
	}

	records := make([]*nfttypes.LiveNFTRecord, 0, len(response.Nfts))
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
		if err := keeper.validateCanonicalClassID(token.ClassId); err != nil {
			return nil, nil, fmt.Errorf(
				"listed nft has invalid stored class ID %q: %w",
				token.ClassId,
				err,
			)
		}
		if err := nfttypes.ValidateNFTID(token.Id); err != nil {
			return nil, nil, fmt.Errorf("listed nft has invalid stored ID %q: %w", token.Id, err)
		}
		live, err := getLiveNFTRecordWithPageCache(
			keeper,
			ctx,
			token.ClassId,
			token.Id,
			cache,
		)
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

func getLiveNFTRecordWithPageCache(
	keeper Keeper,
	ctx sdk.Context,
	classID string,
	nftID string,
	cache *liveNFTPageCache,
) (*nfttypes.LiveNFTRecord, error) {
	state, err := keeper.loadNFTState(ctx, classID, nftID)
	if err != nil {
		return nil, err
	}
	if err := state.validateLiveCombination(classID, nftID); err != nil {
		return nil, err
	}
	if !state.hasNFT {
		return nil, upstreamnft.ErrNFTNotExists.Wrapf(
			"nft %s in class %s not found",
			nftID,
			classID,
		)
	}
	classRecord, err := cache.classRecord(keeper, ctx, classID)
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
	liveSupply, err := cache.classSupply(keeper, ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("load supply for nft %s: %w", nftID, err)
	}
	if liveSupply == 0 ||
		classRecord.MintedCount == 0 ||
		liveSupply > classRecord.MintedCount {
		return nil, fmt.Errorf(
			"nft %s in class %s has inconsistent supply %d and minted count %d",
			nftID,
			classID,
			liveSupply,
			classRecord.MintedCount,
		)
	}

	if err := nfttypes.ValidateURI(state.nft.Uri, state.nft.UriHash); err != nil {
		return nil, fmt.Errorf("nft %s has invalid stored URI metadata: %w", nftID, err)
	}
	canonicalData, err := nfttypes.CanonicalizeNFTData(keeper.cdc, state.nft.Data)
	if err != nil {
		return nil, fmt.Errorf("nft %s has invalid stored data: %w", nftID, err)
	}
	state.nft.Data = canonicalData

	ownerAddress := keeper.nftKeeper.GetOwner(ctx, classID, nftID)
	if len(ownerAddress) == 0 {
		return nil, fmt.Errorf("nft %s in class %s has no owner", nftID, classID)
	}
	owner, err := keeper.addressCodec.BytesToString(ownerAddress)
	if err != nil {
		return nil, fmt.Errorf("encode owner for nft %s: %w", nftID, err)
	}
	canonicalOwner, _, err := keeper.canonicalNonModuleAccount("stored owner", owner)
	if err != nil {
		return nil, fmt.Errorf("nft %s has invalid stored owner: %w", nftID, err)
	}
	status := nfttypes.LiveNFTStatus_LIVE_NFT_STATUS_ACTIVE
	if state.lifecycle.Revocation != nil {
		status = nfttypes.LiveNFTStatus_LIVE_NFT_STATUS_REVOKED
	}

	return &nfttypes.LiveNFTRecord{
		Nft:        &state.nft,
		Owner:      canonicalOwner,
		Status:     status,
		Mint:       state.lifecycle.Mint,
		Revocation: state.lifecycle.Revocation,
	}, nil
}

func (f maximumPageNFTFilter) values(
	fixture maximumPageQueryFixture,
) (classID string, owner string) {
	if f.includeClass {
		classID = fixture.classID
	}
	if f.includeOwner {
		owner = fixture.owner
	}
	return classID, owner
}
