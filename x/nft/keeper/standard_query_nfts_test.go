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
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStandardQueryNFTsSupportsClassAndOwnerFilters(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, controller, owner := createNFTsForClassQueryTest(
		t,
		&fixture,
		"beta",
		"alpha",
	)
	otherClassID, _, otherClassOwner := createNFTsForQueryTest(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{101}, 20)),
		"delta",
	)
	require.Equal(t, owner, otherClassOwner)

	receiverAddress := sdk.AccAddress(bytes.Repeat([]byte{102}, 32))
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
		&nfttypes.MsgRevokeRequest{
			ClassId:    classID,
			NftId:      "alpha",
			Controller: controller,
		},
	)
	require.NoError(t, err)

	server := NewStandardQueryServer(fixture.keeper)
	first, err := server.NFTs(
		fixture.ctx,
		&upstreamnft.QueryNFTsRequest{
			ClassId:    classID,
			Pagination: &query.PageRequest{Limit: 1},
		},
	)
	require.NoError(t, err)
	require.Equal(t, []string{classID + "/alpha"}, standardNFTKeys(first.Nfts))
	require.Equal(t, "https://example.test/alpha.json", first.Nfts[0].Uri)
	require.NotEmpty(t, first.Pagination.NextKey)

	second, err := server.NFTs(
		fixture.ctx,
		&upstreamnft.QueryNFTsRequest{
			ClassId: classID,
			Pagination: &query.PageRequest{
				Key:   first.Pagination.NextKey,
				Limit: 1,
			},
		},
	)
	require.NoError(t, err)
	require.Equal(t, []string{classID + "/beta"}, standardNFTKeys(second.Nfts))
	require.Empty(t, second.Pagination.NextKey)

	ownerRequest := &upstreamnft.QueryNFTsRequest{
		Owner:      strings.ToUpper(owner),
		Pagination: &query.PageRequest{},
	}
	originalOwner := ownerRequest.Owner
	originalPagination := *ownerRequest.Pagination
	ownerResponse, err := server.NFTs(fixture.ctx, ownerRequest)
	require.NoError(t, err)
	expectedOwnerKeys := []string{classID + "/alpha", otherClassID + "/delta"}
	sort.Strings(expectedOwnerKeys)
	require.Equal(t, expectedOwnerKeys, standardNFTKeys(ownerResponse.Nfts))
	require.Equal(t, originalOwner, ownerRequest.Owner)
	require.Equal(t, originalPagination, *ownerRequest.Pagination)

	intersection, err := server.NFTs(
		fixture.ctx,
		&upstreamnft.QueryNFTsRequest{ClassId: classID, Owner: receiver},
	)
	require.NoError(t, err)
	require.Equal(t, []string{classID + "/beta"}, standardNFTKeys(intersection.Nfts))
	require.NotNil(t, intersection.Pagination)
}

func TestStandardQueryNFTsReturnsEmptyForUnknownFilters(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{103}, 20)))
	owner := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{104}, 20)))

	response, err := NewStandardQueryServer(fixture.keeper).NFTs(
		fixture.ctx,
		&upstreamnft.QueryNFTsRequest{
			ClassId: creator + ":unknown",
			Owner:   owner,
		},
	)
	require.NoError(t, err)
	require.Empty(t, response.Nfts)
	require.NotNil(t, response.Pagination)
	require.Empty(t, response.Pagination.NextKey)
}

func TestStandardQueryNFTsErrorMapping(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	server := NewStandardQueryServer(fixture.keeper)
	goCtx := fixture.ctx
	creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{105}, 20)))

	invalidRequests := []*upstreamnft.QueryNFTsRequest{
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
		_, err := server.NFTs(goCtx, request)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	}

	classID, _, _ := createNFTsForClassQueryTest(t, &fixture, "alpha")
	require.NoError(t, fixture.keeper.lifecycles.Remove(
		fixture.ctx,
		collections.Join(classID, "alpha"),
	))
	_, err := server.NFTs(
		fixture.ctx,
		&upstreamnft.QueryNFTsRequest{ClassId: classID},
	)
	require.Equal(t, codes.Internal, status.Code(err))
}

func standardNFTKeys(nfts []*upstreamnft.NFT) []string {
	keys := make([]string, len(nfts))
	for index, nft := range nfts {
		keys[index] = nft.ClassId + "/" + nft.Id
	}
	return keys
}
