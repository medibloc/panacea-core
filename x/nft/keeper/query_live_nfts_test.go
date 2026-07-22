package keeper

import (
	"bytes"
	"testing"
	"time"

	"cosmossdk.io/collections"
	upstreamnft "cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
)

func TestListLiveNFTRecordsByClassPaginatesInNFTIDOrder(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, _, _ := createNFTsForClassQueryTest(t, &fixture, "gamma", "alpha", "beta")

	first, firstPage, err := fixture.keeper.listLiveNFTRecordsByClass(
		fixture.ctx,
		classID,
		&query.PageRequest{Limit: 2},
	)
	require.NoError(t, err)
	require.Equal(t, []string{"alpha", "beta"}, liveNFTRecordIDs(first))
	require.NotEmpty(t, firstPage.NextKey)
	require.Zero(t, firstPage.Total)

	second, secondPage, err := fixture.keeper.listLiveNFTRecordsByClass(
		fixture.ctx,
		classID,
		&query.PageRequest{Key: firstPage.NextKey, Limit: 2},
	)
	require.NoError(t, err)
	require.Equal(t, []string{"gamma"}, liveNFTRecordIDs(second))
	require.Empty(t, secondPage.NextKey)
	require.Zero(t, secondPage.Total)
}

func TestListLiveNFTRecordsByClassSupportsReversePagination(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, _, _ := createNFTsForClassQueryTest(t, &fixture, "gamma", "alpha", "beta")

	first, firstPage, err := fixture.keeper.listLiveNFTRecordsByClass(
		fixture.ctx,
		classID,
		&query.PageRequest{Limit: 2, Reverse: true},
	)
	require.NoError(t, err)
	require.Equal(t, []string{"gamma", "beta"}, liveNFTRecordIDs(first))
	require.NotEmpty(t, firstPage.NextKey)

	second, secondPage, err := fixture.keeper.listLiveNFTRecordsByClass(
		fixture.ctx,
		classID,
		&query.PageRequest{Key: firstPage.NextKey, Limit: 2, Reverse: true},
	)
	require.NoError(t, err)
	require.Equal(t, []string{"alpha"}, liveNFTRecordIDs(second))
	require.Empty(t, secondPage.NextKey)
}

func TestListLiveNFTRecordsByClassIncludesRevokedAndExcludesBurned(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, controller, owner := createNFTsForClassQueryTest(t, &fixture, "gamma", "alpha", "beta")

	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	_, err := NewMsgServer(fixture.keeper).Revoke(
		sdk.WrapSDKContext(fixture.ctx),
		&nfttypes.MsgRevokeRequest{ClassId: classID, NftId: "beta", Controller: controller},
	)
	require.NoError(t, err)
	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	_, err = NewMsgServer(fixture.keeper).Burn(
		sdk.WrapSDKContext(fixture.ctx),
		&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "alpha", Owner: owner},
	)
	require.NoError(t, err)

	records, page, err := fixture.keeper.listLiveNFTRecordsByClass(
		fixture.ctx,
		classID,
		&query.PageRequest{Limit: 100},
	)
	require.NoError(t, err)
	require.Equal(t, []string{"beta", "gamma"}, liveNFTRecordIDs(records))
	require.Equal(t, nfttypes.LiveNFTStatus_LIVE_NFT_STATUS_REVOKED, records[0].Status)
	require.Equal(t, nfttypes.LiveNFTStatus_LIVE_NFT_STATUS_ACTIVE, records[1].Status)
	require.Empty(t, page.NextKey)
}

func TestListLiveNFTRecordsByClassReturnsEmptyForUnknownClass(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{95}, 20)))

	records, page, err := fixture.keeper.listLiveNFTRecordsByClass(
		fixture.ctx,
		creator+":unknown",
		&query.PageRequest{Limit: 100},
	)

	require.NoError(t, err)
	require.Empty(t, records)
	require.NotNil(t, page)
	require.Empty(t, page.NextKey)
}

func TestListLiveNFTRecordsByClassRejectsInconsistentState(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, _, _ := createNFTsForClassQueryTest(t, &fixture, "alpha")
	require.NoError(t, fixture.keeper.lifecycles.Remove(
		fixture.ctx,
		collections.Join(classID, "alpha"),
	))

	records, page, err := fixture.keeper.listLiveNFTRecordsByClass(
		fixture.ctx,
		classID,
		&query.PageRequest{Limit: 100},
	)

	require.Nil(t, records)
	require.Nil(t, page)
	require.ErrorContains(t, err, "inconsistent standard, lifecycle, and tombstone state")
}

func TestListLiveNFTRecordsByClassRejectsOrphanClass(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{96}, 20)))
	classID := creator + ":orphan"
	require.NoError(t, fixture.keeper.nftKeeper.SaveClass(
		fixture.ctx,
		upstreamnft.Class{Id: classID},
	))

	records, page, err := fixture.keeper.listLiveNFTRecordsByClass(
		fixture.ctx,
		classID,
		&query.PageRequest{Limit: 100},
	)

	require.Nil(t, records)
	require.Nil(t, page)
	require.ErrorContains(t, err, "inconsistent standard and policy state")
}

func createNFTsForClassQueryTest(
	t *testing.T,
	fixture *keeperFixture,
	nftIDs ...string,
) (classID, controller, owner string) {
	return createNFTsForQueryTest(
		t,
		fixture,
		sdk.AccAddress(bytes.Repeat([]byte{93}, 20)),
		nftIDs...,
	)
}

func createNFTsForQueryTest(
	t *testing.T,
	fixture *keeperFixture,
	creatorAddress sdk.AccAddress,
	nftIDs ...string,
) (classID, controller, owner string) {
	t.Helper()
	classID, controller = createClassForMintTest(
		t,
		fixture,
		creatorAddress,
		uint64(len(nftIDs)),
	)
	ownerAddress := sdk.AccAddress(bytes.Repeat([]byte{94}, 32))
	owner = fixture.accountAddress(t, ownerAddress)
	fixture.accountKeeper.accounts[string(ownerAddress)] =
		authtypes.NewBaseAccountWithAddress(ownerAddress)
	for _, nftID := range nftIDs {
		request := validMintRequest(classID, controller, owner)
		request.NftId = nftID
		request.Uri = "https://example.test/" + nftID + ".json"
		_, err := NewMsgServer(fixture.keeper).Mint(sdk.WrapSDKContext(fixture.ctx), request)
		require.NoError(t, err)
	}
	fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())
	return classID, controller, owner
}

func liveNFTRecordIDs(records []*nfttypes.LiveNFTRecord) []string {
	nftIDs := make([]string, len(records))
	for index, record := range records {
		nftIDs[index] = record.Nft.Id
	}
	return nftIDs
}
