package keeper

import (
	"bytes"
	"sort"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/collections"
	upstreamnft "cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestQueryNFTRecordsSupportsClassAndOwnerFilters(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, controller, owner := createNFTsForClassQueryTest(
		t,
		&fixture,
		"gamma",
		"beta",
		"alpha",
	)
	otherClassID, _, otherClassOwner := createNFTsForQueryTest(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{107}, 20)),
		"delta",
	)
	require.Equal(t, owner, otherClassOwner)

	receiverAddress := sdk.AccAddress(bytes.Repeat([]byte{108}, 32))
	receiver := fixture.accountAddress(t, receiverAddress)
	fixture.accountKeeper.accounts[string(receiverAddress)] =
		authtypes.NewBaseAccountWithAddress(receiverAddress)
	_, err := NewStandardMsgServer(fixture.keeper).Send(
		fixture.ctx,
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
		fixture.ctx,
		&types.MsgRevokeRequest{
			ClassId:    classID,
			NftId:      "alpha",
			Controller: controller,
		},
	)
	require.NoError(t, err)
	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	_, err = NewMsgServer(fixture.keeper).Burn(
		fixture.ctx,
		&types.MsgBurnRequest{ClassId: classID, NftId: "gamma", Owner: owner},
	)
	require.NoError(t, err)

	server := NewQueryServer(fixture.keeper)
	first, err := server.NFTRecords(
		fixture.ctx,
		&types.QueryNFTRecordsRequest{
			ClassId:    classID,
			Pagination: &query.PageRequest{Limit: 1},
		},
	)
	require.NoError(t, err)
	require.Equal(t, []string{classID + "/alpha"}, liveNFTRecordKeys(first.NftRecords))
	require.Equal(t, types.LiveNFTStatus_LIVE_NFT_STATUS_REVOKED, first.NftRecords[0].Status)
	require.Equal(t, owner, first.NftRecords[0].Owner)
	require.NotNil(t, first.NftRecords[0].Revocation)
	require.NotEmpty(t, first.Pagination.NextKey)

	second, err := server.NFTRecords(
		fixture.ctx,
		&types.QueryNFTRecordsRequest{
			ClassId: classID,
			Pagination: &query.PageRequest{
				Key:   first.Pagination.NextKey,
				Limit: 1,
			},
		},
	)
	require.NoError(t, err)
	require.Equal(t, []string{classID + "/beta"}, liveNFTRecordKeys(second.NftRecords))
	require.Equal(t, types.LiveNFTStatus_LIVE_NFT_STATUS_ACTIVE, second.NftRecords[0].Status)
	require.Equal(t, receiver, second.NftRecords[0].Owner)
	require.Empty(t, second.Pagination.NextKey)

	ownerRequest := &types.QueryNFTRecordsRequest{
		Owner:      strings.ToUpper(owner),
		Pagination: &query.PageRequest{},
	}
	originalOwner := ownerRequest.Owner
	originalPagination := *ownerRequest.Pagination
	ownerResponse, err := server.NFTRecords(fixture.ctx, ownerRequest)
	require.NoError(t, err)
	expectedOwnerKeys := []string{classID + "/alpha", otherClassID + "/delta"}
	sort.Strings(expectedOwnerKeys)
	require.Equal(t, expectedOwnerKeys, liveNFTRecordKeys(ownerResponse.NftRecords))
	require.Equal(t, originalOwner, ownerRequest.Owner)
	require.Equal(t, originalPagination, *ownerRequest.Pagination)

	intersection, err := server.NFTRecords(
		fixture.ctx,
		&types.QueryNFTRecordsRequest{ClassId: classID, Owner: receiver},
	)
	require.NoError(t, err)
	require.Equal(t, []string{classID + "/beta"}, liveNFTRecordKeys(intersection.NftRecords))
	require.NotNil(t, intersection.Pagination)

	burned, err := server.NFTRecord(
		fixture.ctx,
		&types.QueryNFTRecordRequest{ClassId: classID, NftId: "gamma"},
	)
	require.NoError(t, err)
	require.NotNil(t, burned.NftRecord.GetBurnTombstone())
}

func TestQueryNFTRecordsReturnsEmptyForUnknownFilters(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{109}, 20)))
	owner := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{110}, 20)))

	response, err := NewQueryServer(fixture.keeper).NFTRecords(
		fixture.ctx,
		&types.QueryNFTRecordsRequest{
			ClassId: creator + ":unknown",
			Owner:   owner,
		},
	)
	require.NoError(t, err)
	require.Empty(t, response.NftRecords)
	require.NotNil(t, response.Pagination)
	require.Empty(t, response.Pagination.NextKey)
}

func TestQueryNFTRecordsErrorMapping(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	server := NewQueryServer(fixture.keeper)
	goCtx := fixture.ctx
	creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{111}, 20)))

	invalidRequests := []*types.QueryNFTRecordsRequest{
		nil,
		{},
		{ClassId: "invalid"},
		{ClassId: strings.ToUpper(creator) + ":class"},
		{Owner: "invalid"},
		{
			ClassId:    creator + ":class",
			Pagination: &query.PageRequest{Offset: 1},
		},
		{
			ClassId:    creator + ":class",
			Pagination: &query.PageRequest{CountTotal: true},
		},
		{
			ClassId:    creator + ":class",
			Pagination: &query.PageRequest{Limit: maximumQueryPageLimit + 1},
		},
	}
	for _, request := range invalidRequests {
		_, err := server.NFTRecords(goCtx, request)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	}

	classID, _, _ := createNFTsForClassQueryTest(t, &fixture, "alpha")
	require.NoError(t, fixture.keeper.lifecycles.Remove(
		fixture.ctx,
		collections.Join(classID, "alpha"),
	))
	_, err := server.NFTRecords(
		fixture.ctx,
		&types.QueryNFTRecordsRequest{ClassId: classID},
	)
	require.Equal(t, codes.Internal, status.Code(err))
}
