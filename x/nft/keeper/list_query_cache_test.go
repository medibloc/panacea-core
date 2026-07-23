package keeper

import (
	"bytes"
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

func TestListQueryPageCacheHandlesMultipleClasses(t *testing.T) {
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
	counters.reset()

	records, page, err := fixture.keeper.listLiveNFTRecordsByOwner(
		fixture.ctx,
		"",
		owner,
		&query.PageRequest{Limit: maximumQueryPageLimit},
	)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{
		firstClassID + "/alpha",
		firstClassID + "/beta",
		secondClassID + "/delta",
		secondClassID + "/gamma",
	}, liveNFTRecordKeys(records))
	require.Empty(t, page.NextKey)
	var revoked int
	for _, record := range records {
		if record.Status == nfttypes.LiveNFTStatus_LIVE_NFT_STATUS_REVOKED {
			revoked++
		}
	}
	require.Equal(t, 1, revoked)
	require.Equal(t, storeReadCounters{
		gets:          16,
		iterators:     1,
		iteratorNexts: 4,
	}, counters.nft)
	require.Equal(t, storeReadCounters{gets: 12}, counters.policy)
}
