package keeper

import (
	"bytes"
	"sort"
	"testing"
	"time"

	upstreamnft "cosmossdk.io/x/nft"
	upstreamkeeper "cosmossdk.io/x/nft/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
)

func TestListLiveNFTRecordsByOwnerPaginatesInClassAndNFTIDOrder(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	firstClassID, _, owner := createNFTsForClassQueryTest(
		t,
		&fixture,
		"beta",
		"alpha",
	)
	secondClassID, _, secondOwner := createNFTsForQueryTest(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{97}, 20)),
		"delta",
		"gamma",
	)
	require.Equal(t, owner, secondOwner)

	expected := []string{
		firstClassID + "/alpha",
		firstClassID + "/beta",
		secondClassID + "/delta",
		secondClassID + "/gamma",
	}
	sort.Strings(expected)

	first, firstPage, err := fixture.keeper.listLiveNFTRecordsByOwner(
		fixture.ctx,
		"",
		owner,
		&query.PageRequest{Limit: 3},
	)
	require.NoError(t, err)
	require.NotEmpty(t, firstPage.NextKey)
	require.Zero(t, firstPage.Total)

	second, secondPage, err := fixture.keeper.listLiveNFTRecordsByOwner(
		fixture.ctx,
		"",
		owner,
		&query.PageRequest{Key: firstPage.NextKey, Limit: 3},
	)
	require.NoError(t, err)
	require.Empty(t, secondPage.NextKey)
	require.Equal(t, expected, liveNFTRecordKeys(append(first, second...)))

	reverse, reversePage, err := fixture.keeper.listLiveNFTRecordsByOwner(
		fixture.ctx,
		"",
		owner,
		&query.PageRequest{Limit: 100, Reverse: true},
	)
	require.NoError(t, err)
	require.Empty(t, reversePage.NextKey)
	require.Equal(t, reversedStrings(expected), liveNFTRecordKeys(reverse))
}

func TestListLiveNFTRecordsByOwnerSupportsClassIntersection(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, controller, owner := createNFTsForClassQueryTest(
		t,
		&fixture,
		"gamma",
		"alpha",
		"beta",
	)
	_, _, otherClassOwner := createNFTsForQueryTest(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{97}, 20)),
		"delta",
	)
	require.Equal(t, owner, otherClassOwner)

	receiverAddress := sdk.AccAddress(bytes.Repeat([]byte{99}, 32))
	receiver := fixture.accountAddress(t, receiverAddress)
	fixture.accountKeeper.accounts[string(receiverAddress)] =
		authtypes.NewBaseAccountWithAddress(receiverAddress)
	_, err := NewStandardMsgServer(fixture.keeper).Send(
		sdk.WrapSDKContext(fixture.ctx),
		&upstreamnft.MsgSend{
			ClassId:  classID,
			Id:       "beta",
			Sender:   owner,
			Receiver: receiver,
		},
	)
	require.NoError(t, err)

	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	_, err = NewMsgServer(fixture.keeper).Revoke(
		sdk.WrapSDKContext(fixture.ctx),
		&nfttypes.MsgRevokeRequest{
			ClassId:    classID,
			NftId:      "gamma",
			Controller: controller,
		},
	)
	require.NoError(t, err)
	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	_, err = NewMsgServer(fixture.keeper).Burn(
		sdk.WrapSDKContext(fixture.ctx),
		&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "alpha", Owner: owner},
	)
	require.NoError(t, err)

	ownerRecords, page, err := fixture.keeper.listLiveNFTRecordsByOwner(
		fixture.ctx,
		classID,
		owner,
		&query.PageRequest{Limit: 100},
	)
	require.NoError(t, err)
	require.Empty(t, page.NextKey)
	require.Equal(t, []string{classID + "/gamma"}, liveNFTRecordKeys(ownerRecords))
	require.Equal(t, nfttypes.LiveNFTStatus_LIVE_NFT_STATUS_REVOKED, ownerRecords[0].Status)

	receiverRecords, _, err := fixture.keeper.listLiveNFTRecordsByOwner(
		fixture.ctx,
		classID,
		receiver,
		&query.PageRequest{Limit: 100},
	)
	require.NoError(t, err)
	require.Equal(t, []string{classID + "/beta"}, liveNFTRecordKeys(receiverRecords))
}

func TestListLiveNFTRecordsByOwnerRejectsMismatchedOwnerIndex(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, _, owner := createNFTsForClassQueryTest(t, &fixture, "alpha")
	otherOwnerAddress := sdk.AccAddress(bytes.Repeat([]byte{100}, 20))
	otherOwner := fixture.accountAddress(t, otherOwnerAddress)

	ownerKey := make([]byte, 0, len(classID)+len("alpha")+2)
	ownerKey = append(ownerKey, upstreamkeeper.OwnerKey...)
	ownerKey = append(ownerKey, classID...)
	ownerKey = append(ownerKey, upstreamkeeper.Delimiter...)
	ownerKey = append(ownerKey, "alpha"...)
	require.NoError(t, fixture.keeper.nftStoreService.OpenKVStore(fixture.ctx).Set(
		ownerKey,
		otherOwnerAddress,
	))

	records, page, err := fixture.keeper.listLiveNFTRecordsByOwner(
		fixture.ctx,
		classID,
		owner,
		&query.PageRequest{Limit: 100},
	)
	require.Nil(t, records)
	require.Nil(t, page)
	require.ErrorContains(t, err, "has owner "+otherOwner+", expected "+owner)
}

func liveNFTRecordKeys(records []*nfttypes.LiveNFTRecord) []string {
	keys := make([]string, len(records))
	for index, record := range records {
		keys[index] = record.Nft.ClassId + "/" + record.Nft.Id
	}
	return keys
}

func reversedStrings(values []string) []string {
	reversed := append([]string(nil), values...)
	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}
	return reversed
}
