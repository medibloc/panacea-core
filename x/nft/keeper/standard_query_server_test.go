package keeper

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/collections"
	upstreamnft "cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStandardQueryClass(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{61}, 20)))
	request := validCreateClassRequest(creator)
	_, err := NewMsgServer(fixture.keeper).CreateClass(
		sdk.WrapSDKContext(fixture.ctx),
		request,
	)
	require.NoError(t, err)

	classID := creator + ":" + request.LocalClassId
	response, err := NewStandardQueryServer(fixture.keeper).Class(
		sdk.WrapSDKContext(fixture.ctx),
		&upstreamnft.QueryClassRequest{ClassId: classID},
	)
	require.NoError(t, err)
	require.NotNil(t, response.Class)
	require.Equal(t, classID, response.Class.Id)
	require.Equal(t, request.Name, response.Class.Name)
	require.Equal(t, request.Symbol, response.Class.Symbol)
	require.Equal(t, request.Description, response.Class.Description)
	require.Equal(t, request.Uri, response.Class.Uri)
	require.Equal(t, request.UriHash, response.Class.UriHash)
	require.Nil(t, response.Class.Data)
}

func TestStandardQueryClassErrorMapping(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	server := NewStandardQueryServer(fixture.keeper)
	creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{62}, 20)))
	goCtx := sdk.WrapSDKContext(fixture.ctx)

	_, err := server.Class(goCtx, nil)
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = server.Class(goCtx, &upstreamnft.QueryClassRequest{ClassId: "invalid"})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = server.Class(goCtx, &upstreamnft.QueryClassRequest{
		ClassId: strings.ToUpper(creator) + ":class",
	})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = server.Class(goCtx, &upstreamnft.QueryClassRequest{
		ClassId: creator + ":missing",
	})
	require.Equal(t, codes.NotFound, status.Code(err))

	orphanClassID := creator + ":orphan"
	require.NoError(t, fixture.keeper.nftKeeper.SaveClass(
		fixture.ctx,
		upstreamnft.Class{Id: orphanClassID},
	))
	_, err = server.Class(goCtx, &upstreamnft.QueryClassRequest{ClassId: orphanClassID})
	require.Equal(t, codes.Internal, status.Code(err))
}

func TestStandardQueryNFTIncludesActiveAndRevoked(t *testing.T) {
	t.Run("active", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, _, _, _, expectedData := createNFTForBurnTest(t, &fixture)

		response, err := NewStandardQueryServer(fixture.keeper).NFT(
			sdk.WrapSDKContext(fixture.ctx),
			&upstreamnft.QueryNFTRequest{ClassId: classID, Id: "nft-1"},
		)
		require.NoError(t, err)
		require.NotNil(t, response.Nft)
		require.Equal(t, classID, response.Nft.ClassId)
		require.Equal(t, "nft-1", response.Nft.Id)
		require.Equal(t, "https://example.test/nft-1.json", response.Nft.Uri)
		require.Equal(t, "sha256:"+strings.Repeat("b", 64), response.Nft.UriHash)
		require.Equal(t, expectedData.TypeUrl, response.Nft.Data.TypeUrl)
		require.Equal(t, expectedData.Value, response.Nft.Data.Value)
	})

	t.Run("revoked", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, controller, _, _ := createNFTForRevokeTest(t, &fixture, true)
		fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
		_, err := NewMsgServer(fixture.keeper).Revoke(
			sdk.WrapSDKContext(fixture.ctx),
			&nfttypes.MsgRevokeRequest{
				ClassId: classID, NftId: "nft-1", Controller: controller,
			},
		)
		require.NoError(t, err)

		response, err := NewStandardQueryServer(fixture.keeper).NFT(
			sdk.WrapSDKContext(fixture.ctx),
			&upstreamnft.QueryNFTRequest{ClassId: classID, Id: "nft-1"},
		)
		require.NoError(t, err)
		require.NotNil(t, response.Nft)
		require.Equal(t, classID, response.Nft.ClassId)
		require.Equal(t, "nft-1", response.Nft.Id)
	})
}

func TestStandardQueryNFTReturnsNotFoundForBurnedAndUnknownIDs(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, _, owner, _, _ := createNFTForBurnTest(t, &fixture)
	server := NewStandardQueryServer(fixture.keeper)
	goCtx := sdk.WrapSDKContext(fixture.ctx)

	_, err := server.NFT(goCtx, &upstreamnft.QueryNFTRequest{
		ClassId: classID,
		Id:      "unknown",
	})
	require.Equal(t, codes.NotFound, status.Code(err))

	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	_, err = NewMsgServer(fixture.keeper).Burn(
		sdk.WrapSDKContext(fixture.ctx),
		&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "nft-1", Owner: owner},
	)
	require.NoError(t, err)
	_, err = server.NFT(
		sdk.WrapSDKContext(fixture.ctx),
		&upstreamnft.QueryNFTRequest{ClassId: classID, Id: "nft-1"},
	)
	require.Equal(t, codes.NotFound, status.Code(err))
}

func TestStandardQueryNFTErrorMapping(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	server := NewStandardQueryServer(fixture.keeper)
	creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{63}, 20)))
	goCtx := sdk.WrapSDKContext(fixture.ctx)

	_, err := server.NFT(goCtx, nil)
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = server.NFT(goCtx, &upstreamnft.QueryNFTRequest{ClassId: "invalid", Id: "nft-1"})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = server.NFT(goCtx, &upstreamnft.QueryNFTRequest{
		ClassId: creator + ":class",
		Id:      ".",
	})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	classID, _, _, _, _ := createNFTForBurnTest(t, &fixture)
	require.NoError(t, fixture.keeper.lifecycles.Remove(
		fixture.ctx,
		collections.Join(classID, "nft-1"),
	))
	_, err = server.NFT(
		sdk.WrapSDKContext(fixture.ctx),
		&upstreamnft.QueryNFTRequest{ClassId: classID, Id: "nft-1"},
	)
	require.Equal(t, codes.Internal, status.Code(err))
}
