package keeper

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/collections"
	corestore "cosmossdk.io/core/store"
	upstreamnft "cosmossdk.io/x/nft"
	"github.com/cosmos/cosmos-sdk/runtime"
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

func TestStandardQueryOwnerIncludesActiveAndRevoked(t *testing.T) {
	t.Run("active", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, _, owner, _, _ := createNFTForBurnTest(t, &fixture)

		response, err := NewStandardQueryServer(fixture.keeper).Owner(
			sdk.WrapSDKContext(fixture.ctx),
			&upstreamnft.QueryOwnerRequest{ClassId: classID, Id: "nft-1"},
		)
		require.NoError(t, err)
		require.Equal(t, owner, response.Owner)
	})

	t.Run("revoked", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, controller, owner, _ := createNFTForRevokeTest(t, &fixture, true)
		fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
		_, err := NewMsgServer(fixture.keeper).Revoke(
			sdk.WrapSDKContext(fixture.ctx),
			&nfttypes.MsgRevokeRequest{
				ClassId: classID, NftId: "nft-1", Controller: controller,
			},
		)
		require.NoError(t, err)

		response, err := NewStandardQueryServer(fixture.keeper).Owner(
			sdk.WrapSDKContext(fixture.ctx),
			&upstreamnft.QueryOwnerRequest{ClassId: classID, Id: "nft-1"},
		)
		require.NoError(t, err)
		require.Equal(t, owner, response.Owner)
	})
}

func TestStandardQueryOwnerReturnsEmptyForBurnedAndUnknownIDs(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, _, owner, _, _ := createNFTForBurnTest(t, &fixture)
	server := NewStandardQueryServer(fixture.keeper)

	response, err := server.Owner(
		sdk.WrapSDKContext(fixture.ctx),
		&upstreamnft.QueryOwnerRequest{ClassId: classID, Id: "unknown"},
	)
	require.NoError(t, err)
	require.Empty(t, response.Owner)

	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	_, err = NewMsgServer(fixture.keeper).Burn(
		sdk.WrapSDKContext(fixture.ctx),
		&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "nft-1", Owner: owner},
	)
	require.NoError(t, err)
	response, err = server.Owner(
		sdk.WrapSDKContext(fixture.ctx),
		&upstreamnft.QueryOwnerRequest{ClassId: classID, Id: "nft-1"},
	)
	require.NoError(t, err)
	require.Empty(t, response.Owner)
}

func TestStandardQueryOwnerErrorMapping(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	server := NewStandardQueryServer(fixture.keeper)
	creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{64}, 20)))
	goCtx := sdk.WrapSDKContext(fixture.ctx)

	_, err := server.Owner(goCtx, nil)
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = server.Owner(goCtx, &upstreamnft.QueryOwnerRequest{
		ClassId: "invalid",
		Id:      "nft-1",
	})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = server.Owner(goCtx, &upstreamnft.QueryOwnerRequest{
		ClassId: creator + ":class",
		Id:      ".",
	})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	classID, _, _, _, _ := createNFTForBurnTest(t, &fixture)
	require.NoError(t, fixture.keeper.lifecycles.Remove(
		fixture.ctx,
		collections.Join(classID, "nft-1"),
	))
	_, err = server.Owner(
		sdk.WrapSDKContext(fixture.ctx),
		&upstreamnft.QueryOwnerRequest{ClassId: classID, Id: "nft-1"},
	)
	require.Equal(t, codes.Internal, status.Code(err))
}

func TestStandardQuerySupplyTracksLiveNFTs(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, controller, owner, _, _ := createNFTForBurnTest(t, &fixture)
	server := NewStandardQueryServer(fixture.keeper)

	response, err := server.Supply(
		sdk.WrapSDKContext(fixture.ctx),
		&upstreamnft.QuerySupplyRequest{ClassId: classID},
	)
	require.NoError(t, err)
	require.Equal(t, uint64(1), response.Amount)

	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	_, err = NewMsgServer(fixture.keeper).Revoke(
		sdk.WrapSDKContext(fixture.ctx),
		&nfttypes.MsgRevokeRequest{
			ClassId: classID, NftId: "nft-1", Controller: controller,
		},
	)
	require.NoError(t, err)
	response, err = server.Supply(
		sdk.WrapSDKContext(fixture.ctx),
		&upstreamnft.QuerySupplyRequest{ClassId: classID},
	)
	require.NoError(t, err)
	require.Equal(t, uint64(1), response.Amount)

	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	_, err = NewMsgServer(fixture.keeper).Burn(
		sdk.WrapSDKContext(fixture.ctx),
		&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "nft-1", Owner: owner},
	)
	require.NoError(t, err)
	response, err = server.Supply(
		sdk.WrapSDKContext(fixture.ctx),
		&upstreamnft.QuerySupplyRequest{ClassId: classID},
	)
	require.NoError(t, err)
	require.Zero(t, response.Amount)
}

func TestStandardQuerySupplyReturnsZeroForEmptyAndUnknownClasses(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, creator := createClassForMintTest(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{65}, 20)),
		10,
	)
	server := NewStandardQueryServer(fixture.keeper)

	for _, targetClassID := range []string{classID, creator + ":unknown"} {
		response, err := server.Supply(
			sdk.WrapSDKContext(fixture.ctx),
			&upstreamnft.QuerySupplyRequest{ClassId: targetClassID},
		)
		require.NoError(t, err)
		require.Zero(t, response.Amount)
	}
}

func TestStandardQuerySupplyErrorMapping(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	server := NewStandardQueryServer(fixture.keeper)
	creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{66}, 20)))
	goCtx := sdk.WrapSDKContext(fixture.ctx)

	_, err := server.Supply(goCtx, nil)
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = server.Supply(goCtx, &upstreamnft.QuerySupplyRequest{ClassId: "invalid"})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	orphanClassID := creator + ":orphan"
	require.NoError(t, fixture.keeper.nftKeeper.SaveClass(
		fixture.ctx,
		upstreamnft.Class{Id: orphanClassID},
	))
	_, err = server.Supply(goCtx, &upstreamnft.QuerySupplyRequest{ClassId: orphanClassID})
	require.Equal(t, codes.Internal, status.Code(err))

	classID, _, _, _, _ := createNFTForBurnTest(t, &fixture)
	require.NoError(t, fixture.keeper.mintedCounts.Set(fixture.ctx, classID, 0))
	_, err = server.Supply(
		sdk.WrapSDKContext(fixture.ctx),
		&upstreamnft.QuerySupplyRequest{ClassId: classID},
	)
	require.Equal(t, codes.Internal, status.Code(err))
}

func TestStandardQueryBalanceTracksLiveOwnership(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, controller, owner, _, _ := createNFTForBurnTest(t, &fixture)
	server := NewStandardQueryServer(fixture.keeper)
	request := &upstreamnft.QueryBalanceRequest{
		ClassId: classID,
		Owner:   strings.ToUpper(owner),
	}

	response, err := server.Balance(sdk.WrapSDKContext(fixture.ctx), request)
	require.NoError(t, err)
	require.Equal(t, uint64(1), response.Amount)

	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	_, err = NewMsgServer(fixture.keeper).Revoke(
		sdk.WrapSDKContext(fixture.ctx),
		&nfttypes.MsgRevokeRequest{
			ClassId: classID, NftId: "nft-1", Controller: controller,
		},
	)
	require.NoError(t, err)
	response, err = server.Balance(sdk.WrapSDKContext(fixture.ctx), request)
	require.NoError(t, err)
	require.Equal(t, uint64(1), response.Amount)

	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	_, err = NewMsgServer(fixture.keeper).Burn(
		sdk.WrapSDKContext(fixture.ctx),
		&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "nft-1", Owner: owner},
	)
	require.NoError(t, err)
	response, err = server.Balance(sdk.WrapSDKContext(fixture.ctx), request)
	require.NoError(t, err)
	require.Zero(t, response.Amount)
}

func TestStandardQueryBalanceReturnsZeroForUnownedAndUnknownClasses(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, creator, _, _, _ := createNFTForBurnTest(t, &fixture)
	otherOwner := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{67}, 20)))
	server := NewStandardQueryServer(fixture.keeper)

	for _, targetClassID := range []string{
		classID,
		creator + ":unknown",
	} {
		response, err := server.Balance(
			sdk.WrapSDKContext(fixture.ctx),
			&upstreamnft.QueryBalanceRequest{ClassId: targetClassID, Owner: otherOwner},
		)
		require.NoError(t, err)
		require.Zero(t, response.Amount)
	}
}

func TestStandardQueryBalanceErrorMapping(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	server := NewStandardQueryServer(fixture.keeper)
	creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{68}, 20)))
	owner := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{69}, 20)))
	goCtx := sdk.WrapSDKContext(fixture.ctx)

	_, err := server.Balance(goCtx, nil)
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = server.Balance(goCtx, &upstreamnft.QueryBalanceRequest{
		ClassId: "invalid",
		Owner:   owner,
	})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = server.Balance(goCtx, &upstreamnft.QueryBalanceRequest{
		ClassId: creator + ":class",
		Owner:   "invalid",
	})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	orphanClassID := creator + ":orphan"
	require.NoError(t, fixture.keeper.nftKeeper.SaveClass(
		fixture.ctx,
		upstreamnft.Class{Id: orphanClassID},
	))
	_, err = server.Balance(goCtx, &upstreamnft.QueryBalanceRequest{
		ClassId: orphanClassID,
		Owner:   owner,
	})
	require.Equal(t, codes.Internal, status.Code(err))
}

func TestStandardQueryBalanceRejectsOwnerClassCountReadFailures(t *testing.T) {
	t.Run("stored zero", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, _, owner, _, _ := createNFTForBurnTest(t, &fixture)
		require.NoError(t, fixture.keeper.ownerClassCounts.Set(
			fixture.ctx,
			collections.Join(classID, owner),
			0,
		))

		_, err := NewStandardQueryServer(fixture.keeper).Balance(
			sdk.WrapSDKContext(fixture.ctx),
			&upstreamnft.QueryBalanceRequest{ClassId: classID, Owner: owner},
		)
		require.Equal(t, codes.Internal, status.Code(err))
		require.ErrorContains(t, err, "stored zero balance count")
	})

	t.Run("store error", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, _, owner, _, _ := createNFTForBurnTest(t, &fixture)
		getCalls := 0
		failingKeeper := NewKeeper(
			fixture.cdc,
			runtime.NewKVStoreService(fixture.nftService),
			failingNthGetStoreService{
				delegate: runtime.NewKVStoreService(fixture.policyService),
				calls:    &getCalls,
				failAt:   3,
			},
			fixture.accountKeeper,
			testBankKeeper{},
			fixture.moduleAccountAddresses,
		)

		_, err := NewStandardQueryServer(failingKeeper).Balance(
			sdk.WrapSDKContext(fixture.ctx),
			&upstreamnft.QueryBalanceRequest{ClassId: classID, Owner: owner},
		)
		require.Equal(t, codes.Internal, status.Code(err))
		require.ErrorContains(t, err, "forced get failure")
		require.Equal(t, 3, getCalls)
	})
}

type failingNthGetStoreService struct {
	delegate corestore.KVStoreService
	calls    *int
	failAt   int
}

func (s failingNthGetStoreService) OpenKVStore(ctx context.Context) corestore.KVStore {
	return failingNthGetStore{
		KVStore: s.delegate.OpenKVStore(ctx),
		calls:   s.calls,
		failAt:  s.failAt,
	}
}

type failingNthGetStore struct {
	corestore.KVStore
	calls  *int
	failAt int
}

func (s failingNthGetStore) Get(key []byte) ([]byte, error) {
	*s.calls++
	if *s.calls == s.failAt {
		return nil, errors.New("forced get failure")
	}
	return s.KVStore.Get(key)
}
