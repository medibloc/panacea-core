package keeper

import (
	"bytes"
	"context"
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/collections"
	corestore "cosmossdk.io/core/store"
	upstreamnft "cosmossdk.io/x/nft"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/gogoproto/proto"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMintCreatesActiveNFTRecord(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	mintTime := time.Date(2026, time.July, 21, 12, 34, 56, 0, time.UTC)
	fixture.ctx = fixture.ctx.WithBlockTime(mintTime)
	classID, creator := createClassForMintTest(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{21}, 20)),
		10,
	)
	controllerAddress := sdk.AccAddress(bytes.Repeat([]byte{22}, 20))
	controller := updateControllerForMintTest(t, &fixture, classID, creator, controllerAddress)
	recipientAddress := sdk.AccAddress(bytes.Repeat([]byte{23}, 32))
	recipient := fixture.accountAddress(t, recipientAddress)
	fixture.accountKeeper.accounts[string(recipientAddress)] =
		authtypes.NewBaseAccountWithAddress(recipientAddress)
	data, err := types.NewAnyWithValue(&nfttypes.BasicNFTData{
		Name:        "Certificate #1",
		Description: "Completion certificate",
		ImageUri:    "https://example.test/certificate.png",
	})
	require.NoError(t, err)
	request := validMintRequest(classID, strings.ToUpper(controller), strings.ToUpper(recipient))
	request.Data = data

	response, err := NewMsgServer(fixture.keeper).Mint(fixture.ctx, request)
	require.NoError(t, err)
	require.Equal(t, &nfttypes.MsgMintResponse{}, response)

	token, hasToken := fixture.keeper.nftKeeper.GetNFT(fixture.ctx, classID, request.NftId)
	require.True(t, hasToken)
	require.Equal(t, classID, token.ClassId)
	require.Equal(t, request.NftId, token.Id)
	require.Equal(t, request.Uri, token.Uri)
	require.Equal(t, request.UriHash, token.UriHash)
	require.Equal(t, data.TypeUrl, token.Data.TypeUrl)
	require.Equal(t, data.Value, token.Data.Value)
	require.Nil(t, token.Data.GetCachedValue())
	var decodedData nfttypes.NFTData
	require.NoError(t, fixture.cdc.UnpackAny(token.Data, &decodedData))
	require.Equal(t, data.GetCachedValue(), decodedData)

	require.Equal(t, recipientAddress, fixture.keeper.nftKeeper.GetOwner(
		fixture.ctx,
		classID,
		request.NftId,
	))
	require.Equal(t, uint64(1), fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID))
	lifecycle, err := fixture.keeper.lifecycles.Get(
		fixture.ctx,
		collections.Join(classID, request.NftId),
	)
	require.NoError(t, err)
	require.Equal(t, &nfttypes.LifecycleRecord{
		ClassId: classID,
		NftId:   request.NftId,
		Mint: &nfttypes.MintRecord{
			MintedAt: mintTime,
			MintedBy: controller,
		},
	}, &lifecycle)
	mintedCount, err := fixture.keeper.mintedCounts.Get(fixture.ctx, classID)
	require.NoError(t, err)
	require.Equal(t, uint64(1), mintedCount)

	classResponse, err := NewQueryServer(fixture.keeper).ClassRecord(
		fixture.ctx,
		&nfttypes.QueryClassRecordRequest{ClassId: classID},
	)
	require.NoError(t, err)
	require.Equal(t, creator, classResponse.ClassRecord.Policy.Creator)
	require.Equal(t, controller, classResponse.ClassRecord.Policy.Controller)
	require.Equal(t, uint64(1), classResponse.ClassRecord.MintedCount)

	nftResponse, err := NewQueryServer(fixture.keeper).NFTRecord(
		fixture.ctx,
		&nfttypes.QueryNFTRecordRequest{ClassId: classID, NftId: request.NftId},
	)
	require.NoError(t, err)
	live := nftResponse.NftRecord.GetLive()
	require.NotNil(t, live)
	tokenBytes, err := fixture.cdc.Marshal(&token)
	require.NoError(t, err)
	liveTokenBytes, err := fixture.cdc.Marshal(live.Nft)
	require.NoError(t, err)
	require.Equal(t, tokenBytes, liveTokenBytes)
	require.Equal(t, recipient, live.Owner)
	require.Equal(t, nfttypes.LiveNFTStatus_LIVE_NFT_STATUS_ACTIVE, live.Status)
	require.Equal(t, lifecycle.Mint, live.Mint)
	require.Nil(t, live.Revocation)
	require.Nil(t, nftResponse.NftRecord.GetBurnTombstone())

	events := fixture.ctx.EventManager().ABCIEvents()
	require.Len(t, events, 1)
	parsedEvent, err := sdk.ParseTypedEvent(events[0])
	require.NoError(t, err)
	require.Equal(t, &upstreamnft.EventMint{
		ClassId: classID,
		Id:      request.NftId,
		Owner:   recipient,
	}, parsedEvent)
}

func TestMintRejectsZeroBlockTime(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, controller := createClassForMintTest(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{24}, 20)),
		10,
	)
	recipientAddress := sdk.AccAddress(bytes.Repeat([]byte{25}, 20))
	recipient := fixture.accountAddress(t, recipientAddress)
	fixture.ctx = fixture.ctx.
		WithBlockTime(time.Time{}).
		WithEventManager(sdk.NewEventManager())
	request := validMintRequest(classID, controller, recipient)

	_, err := NewMsgServer(fixture.keeper).Mint(fixture.ctx, request)
	require.ErrorContains(t, err, "cannot mint nft")
	require.ErrorContains(t, err, "at zero block time")

	require.False(t, fixture.keeper.nftKeeper.HasNFT(fixture.ctx, classID, request.NftId))
	require.Empty(t, fixture.keeper.nftKeeper.GetOwner(fixture.ctx, classID, request.NftId))
	require.Zero(t, fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID))
	require.Zero(t, fixture.keeper.nftKeeper.GetBalance(fixture.ctx, classID, recipientAddress))
	lifecycleExists, err := fixture.keeper.lifecycles.Has(
		fixture.ctx,
		collections.Join(classID, request.NftId),
	)
	require.NoError(t, err)
	require.False(t, lifecycleExists)
	mintedCount, err := fixture.keeper.mintedCounts.Get(fixture.ctx, classID)
	require.NoError(t, err)
	require.Zero(t, mintedCount)
	requireOwnerClassCountMissing(t, &fixture, classID, recipient)
	require.Empty(t, fixture.ctx.EventManager().Events())
}

func TestMintValidatesAnyWireDataWithOrWithoutCache(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, controller := createClassForMintTest(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{24}, 20)),
		0,
	)
	recipient := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{25}, 20)))
	data, err := types.NewAnyWithValue(&nfttypes.BasicNFTData{Name: "metadata"})
	require.NoError(t, err)
	original := validMintRequest(classID, controller, recipient)
	original.Data = data

	binary, err := fixture.cdc.Marshal(original)
	require.NoError(t, err)
	var withCache nfttypes.MsgMintRequest
	require.NoError(t, fixture.cdc.Unmarshal(binary, &withCache))
	require.NotNil(t, withCache.Data.GetCachedValue())
	var withoutCache nfttypes.MsgMintRequest
	require.NoError(t, proto.Unmarshal(binary, &withoutCache))
	require.Nil(t, withoutCache.Data.GetCachedValue())
	withoutCache.NftId = "nft-2"

	server := NewMsgServer(fixture.keeper)
	_, err = server.Mint(fixture.ctx, &withCache)
	require.NoError(t, err)
	_, err = server.Mint(fixture.ctx, &withoutCache)
	require.NoError(t, err)

	first, found := fixture.keeper.nftKeeper.GetNFT(fixture.ctx, classID, withCache.NftId)
	require.True(t, found)
	second, found := fixture.keeper.nftKeeper.GetNFT(fixture.ctx, classID, withoutCache.NftId)
	require.True(t, found)
	require.Equal(t, first.Data.TypeUrl, second.Data.TypeUrl)
	require.Equal(t, first.Data.Value, second.Data.Value)
	require.Equal(t, uint64(2), fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID))
}

func TestMintRejectsInvalidRequests(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, controller := createClassForMintTest(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{26}, 20)),
		100,
	)
	recipient := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{27}, 20)))
	arbitraryController := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{28}, 20)))
	moduleRecipient := fixture.accountAddress(t, fixture.accountKeeper.moduleAddress)
	mixedController := strings.ToUpper(controller[:1]) + controller[1:]
	mixedRecipient := strings.ToUpper(recipient[:1]) + recipient[1:]
	validData, err := types.NewAnyWithValue(&nfttypes.BasicNFTData{Name: "cached"})
	require.NoError(t, err)
	validData.Value = nil

	testCases := []struct {
		name      string
		request   *nfttypes.MsgMintRequest
		targetErr error
	}{
		{name: "nil request", targetErr: sdkerrors.ErrInvalidRequest},
		{
			name:      "invalid class ID",
			request:   validMintRequest("invalid", controller, recipient),
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "non-canonical class ID",
			request: validMintRequest(
				strings.ToUpper(strings.Split(classID, ":")[0])+":certificate",
				controller,
				recipient,
			),
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "invalid NFT ID",
			request: func() *nfttypes.MsgMintRequest {
				request := validMintRequest(classID, controller, recipient)
				request.NftId = "."
				return request
			}(),
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "hash without URI",
			request: func() *nfttypes.MsgMintRequest {
				request := validMintRequest(classID, controller, recipient)
				request.Uri = ""
				return request
			}(),
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name:      "invalid current controller",
			request:   validMintRequest(classID, "not-an-address", recipient),
			targetErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name:      "mixed-case current controller",
			request:   validMintRequest(classID, mixedController, recipient),
			targetErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name:      "unauthorized controller",
			request:   validMintRequest(classID, arbitraryController, recipient),
			targetErr: sdkerrors.ErrUnauthorized,
		},
		{
			name:      "invalid recipient",
			request:   validMintRequest(classID, controller, "not-an-address"),
			targetErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name:      "mixed-case recipient",
			request:   validMintRequest(classID, controller, mixedRecipient),
			targetErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name:      "module account recipient",
			request:   validMintRequest(classID, controller, moduleRecipient),
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "unknown data type",
			request: func() *nfttypes.MsgMintRequest {
				request := validMintRequest(classID, controller, recipient)
				request.Data = &types.Any{
					TypeUrl: "/panacea.nft.v1.UnknownNFTData",
					Value:   []byte{0x0a, 0x01, 'x'},
				}
				return request
			}(),
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "cached data with empty wire value",
			request: func() *nfttypes.MsgMintRequest {
				request := validMintRequest(classID, controller, recipient)
				request.Data = validData
				return request
			}(),
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name:      "missing class",
			request:   validMintRequest(controller+":missing", controller, recipient),
			targetErr: upstreamnft.ErrClassNotExists,
		},
	}

	server := NewMsgServer(fixture.keeper)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := server.Mint(fixture.ctx, tc.request)
			require.ErrorIs(t, err, tc.targetErr)
		})
	}

	require.Empty(t, fixture.keeper.nftKeeper.GetNFTsOfClass(fixture.ctx, classID))
	require.Zero(t, fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID))
	mintedCount, err := fixture.keeper.mintedCounts.Get(fixture.ctx, classID)
	require.NoError(t, err)
	require.Zero(t, mintedCount)
	lifecycleExists, err := fixture.keeper.lifecycles.Has(
		fixture.ctx,
		collections.Join(classID, "nft-1"),
	)
	require.NoError(t, err)
	require.False(t, lifecycleExists)
	require.Empty(t, fixture.ctx.EventManager().Events())
}

func TestMintEnforcesLifetimeSupply(t *testing.T) {
	t.Run("finite supply", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, controller := createClassForMintTest(
			t,
			&fixture,
			sdk.AccAddress(bytes.Repeat([]byte{29}, 20)),
			1,
		)
		recipient := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{30}, 20)))
		server := NewMsgServer(fixture.keeper)
		_, err := server.Mint(
			fixture.ctx,
			validMintRequest(classID, controller, recipient),
		)
		require.NoError(t, err)
		fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())
		second := validMintRequest(classID, controller, recipient)
		second.NftId = "nft-2"
		_, err = server.Mint(fixture.ctx, second)
		require.ErrorIs(t, err, nfttypes.ErrMaxSupplyReached)
		require.False(t, fixture.keeper.nftKeeper.HasNFT(fixture.ctx, classID, second.NftId))
		require.Equal(t, uint64(1), fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID))
		mintedCount, err := fixture.keeper.mintedCounts.Get(fixture.ctx, classID)
		require.NoError(t, err)
		require.Equal(t, uint64(1), mintedCount)
		require.Empty(t, fixture.ctx.EventManager().Events())
	})

	t.Run("unlimited supply", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, controller := createClassForMintTest(
			t,
			&fixture,
			sdk.AccAddress(bytes.Repeat([]byte{31}, 20)),
			0,
		)
		recipient := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{32}, 20)))
		server := NewMsgServer(fixture.keeper)
		for _, nftID := range []string{"nft-1", "nft-2"} {
			request := validMintRequest(classID, controller, recipient)
			request.NftId = nftID
			_, err := server.Mint(fixture.ctx, request)
			require.NoError(t, err)
		}
		require.Equal(t, uint64(2), fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID))
		mintedCount, err := fixture.keeper.mintedCounts.Get(fixture.ctx, classID)
		require.NoError(t, err)
		require.Equal(t, uint64(2), mintedCount)
	})
}

func TestNextMintedCountBoundaries(t *testing.T) {
	testCases := []struct {
		name        string
		current     uint64
		maxSupply   uint64
		expected    uint64
		expectError bool
	}{
		{name: "first unlimited mint", expected: 1},
		{name: "last uint64 unlimited mint", current: math.MaxUint64 - 1, expected: math.MaxUint64},
		{name: "unlimited arithmetic exhausted", current: math.MaxUint64, expectError: true},
		{name: "finite limit reached", current: 1, maxSupply: 1, expectError: true},
		{name: "last maximum finite mint", current: math.MaxUint64 - 1, maxSupply: math.MaxUint64, expected: math.MaxUint64},
		{name: "maximum finite arithmetic exhausted", current: math.MaxUint64, maxSupply: math.MaxUint64, expectError: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			next, err := nextMintedCount(tc.current, tc.maxSupply)
			if tc.expectError {
				require.ErrorIs(t, err, nfttypes.ErrMaxSupplyReached)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expected, next)
		})
	}
}

func TestMintRejectsLiveAndBurnedIDReuse(t *testing.T) {
	t.Run("live NFT", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, controller := createClassForMintTest(
			t,
			&fixture,
			sdk.AccAddress(bytes.Repeat([]byte{33}, 20)),
			10,
		)
		recipient := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{34}, 20)))
		request := validMintRequest(classID, controller, recipient)
		server := NewMsgServer(fixture.keeper)
		_, err := server.Mint(fixture.ctx, request)
		require.NoError(t, err)
		fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())

		_, err = server.Mint(fixture.ctx, request)
		require.ErrorIs(t, err, upstreamnft.ErrNFTExists)
		require.Equal(t, uint64(1), fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID))
		mintedCount, err := fixture.keeper.mintedCounts.Get(fixture.ctx, classID)
		require.NoError(t, err)
		require.Equal(t, uint64(1), mintedCount)
		require.Empty(t, fixture.ctx.EventManager().Events())
	})

	t.Run("burn tombstone", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, controller := createClassForMintTest(
			t,
			&fixture,
			sdk.AccAddress(bytes.Repeat([]byte{35}, 20)),
			10,
		)
		recipient := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{36}, 20)))
		request := validMintRequest(classID, controller, recipient)
		require.NoError(t, fixture.keeper.tombstones.Set(
			fixture.ctx,
			collections.Join(classID, request.NftId),
			nfttypes.BurnTombstone{ClassId: classID, NftId: request.NftId},
		))

		_, err := NewMsgServer(fixture.keeper).Mint(fixture.ctx, request)
		require.ErrorIs(t, err, nfttypes.ErrNFTIDPermanentlyUsed)
		require.False(t, fixture.keeper.nftKeeper.HasNFT(fixture.ctx, classID, request.NftId))
		require.Zero(t, fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID))
		mintedCount, err := fixture.keeper.mintedCounts.Get(fixture.ctx, classID)
		require.NoError(t, err)
		require.Zero(t, mintedCount)
		require.Empty(t, fixture.ctx.EventManager().Events())
	})
}

func TestMintRejectsInconsistentNFTState(t *testing.T) {
	t.Run("standard NFT without lifecycle", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, controller := createClassForMintTest(
			t,
			&fixture,
			sdk.AccAddress(bytes.Repeat([]byte{37}, 20)),
			10,
		)
		recipientAddress := sdk.AccAddress(bytes.Repeat([]byte{38}, 20))
		recipient := fixture.accountAddress(t, recipientAddress)
		request := validMintRequest(classID, controller, recipient)
		require.NoError(t, fixture.keeper.nftKeeper.Mint(
			fixture.ctx,
			upstreamnft.NFT{ClassId: classID, Id: request.NftId},
			recipientAddress,
		))
		fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())

		_, err := NewMsgServer(fixture.keeper).Mint(fixture.ctx, request)
		require.ErrorContains(t, err, "inconsistent standard, lifecycle, and tombstone state")
		require.Empty(t, fixture.ctx.EventManager().Events())
	})

	t.Run("lifecycle without standard NFT", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, controller := createClassForMintTest(
			t,
			&fixture,
			sdk.AccAddress(bytes.Repeat([]byte{39}, 20)),
			10,
		)
		recipient := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{40}, 20)))
		request := validMintRequest(classID, controller, recipient)
		require.NoError(t, fixture.keeper.lifecycles.Set(
			fixture.ctx,
			collections.Join(classID, request.NftId),
			nfttypes.LifecycleRecord{
				ClassId: classID,
				NftId:   request.NftId,
				Mint: &nfttypes.MintRecord{
					MintedAt: fixture.ctx.BlockTime(),
					MintedBy: controller,
				},
			},
		))

		_, err := NewMsgServer(fixture.keeper).Mint(fixture.ctx, request)
		require.ErrorContains(t, err, "inconsistent standard, lifecycle, and tombstone state")
		require.Empty(t, fixture.ctx.EventManager().Events())
	})
}

func TestMintRollsBackWhenPolicyWriteFails(t *testing.T) {
	for _, tc := range []struct {
		name   string
		failAt int
	}{
		{name: "lifecycle write", failAt: 1},
		{name: "minted count write", failAt: 2},
		{name: "owner class count write", failAt: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newKeeperFixture(t, true, true)
			classID, controller := createClassForMintTest(
				t,
				&fixture,
				sdk.AccAddress(bytes.Repeat([]byte{41}, 20)),
				10,
			)
			recipientAddress := sdk.AccAddress(bytes.Repeat([]byte{42}, 20))
			recipient := fixture.accountAddress(t, recipientAddress)
			setCalls := 0
			failingKeeper := NewKeeper(
				fixture.cdc,
				runtime.NewKVStoreService(fixture.nftService),
				failingNthSetStoreService{
					delegate: runtime.NewKVStoreService(fixture.policyService),
					calls:    &setCalls,
					failAt:   tc.failAt,
				},
				fixture.accountKeeper,
				testBankKeeper{},
				fixture.moduleAccountAddresses,
			)
			request := validMintRequest(classID, controller, recipient)

			_, err := NewMsgServer(failingKeeper).Mint(fixture.ctx, request)
			require.ErrorContains(t, err, "forced set failure")
			require.False(t, fixture.keeper.nftKeeper.HasNFT(fixture.ctx, classID, request.NftId))
			require.Empty(t, fixture.keeper.nftKeeper.GetOwner(fixture.ctx, classID, request.NftId))
			require.Zero(t, fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID))
			lifecycleExists, err := fixture.keeper.lifecycles.Has(
				fixture.ctx,
				collections.Join(classID, request.NftId),
			)
			require.NoError(t, err)
			require.False(t, lifecycleExists)
			mintedCount, err := fixture.keeper.mintedCounts.Get(fixture.ctx, classID)
			require.NoError(t, err)
			require.Zero(t, mintedCount)
			require.Empty(t, fixture.ctx.EventManager().Events())
		})
	}
}

func TestQueryNFTRecordErrorMapping(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{43}, 20)))
	server := NewQueryServer(fixture.keeper)

	_, err := server.NFTRecord(fixture.ctx, nil)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	_, err = server.NFTRecord(
		fixture.ctx,
		&nfttypes.QueryNFTRecordRequest{ClassId: "invalid", NftId: "nft-1"},
	)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	_, err = server.NFTRecord(
		fixture.ctx,
		&nfttypes.QueryNFTRecordRequest{ClassId: creator + ":missing", NftId: "."},
	)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	_, err = server.NFTRecord(
		fixture.ctx,
		&nfttypes.QueryNFTRecordRequest{ClassId: creator + ":missing", NftId: "nft-1"},
	)
	require.Equal(t, codes.NotFound, status.Code(err))

	classID, controller := createClassForMintTest(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{44}, 20)),
		10,
	)
	require.NoError(t, fixture.keeper.lifecycles.Set(
		fixture.ctx,
		collections.Join(classID, "nft-1"),
		nfttypes.LifecycleRecord{
			ClassId: classID,
			NftId:   "nft-1",
			Mint: &nfttypes.MintRecord{
				MintedAt: fixture.ctx.BlockTime(),
				MintedBy: controller,
			},
		},
	))
	_, err = server.NFTRecord(
		fixture.ctx,
		&nfttypes.QueryNFTRecordRequest{ClassId: classID, NftId: "nft-1"},
	)
	require.Equal(t, codes.Internal, status.Code(err))
}

func validMintRequest(classID, controller, recipient string) *nfttypes.MsgMintRequest {
	return &nfttypes.MsgMintRequest{
		ClassId:    classID,
		NftId:      "nft-1",
		Controller: controller,
		Recipient:  recipient,
		Uri:        "https://example.test/nft-1.json",
		UriHash:    "sha256:" + strings.Repeat("b", 64),
	}
}

func createClassForMintTest(
	t *testing.T,
	fixture *keeperFixture,
	creatorAddress sdk.AccAddress,
	maxSupply uint64,
) (string, string) {
	t.Helper()
	if fixture.ctx.BlockTime().IsZero() {
		fixture.ctx = fixture.ctx.WithBlockTime(time.Date(2026, time.July, 21, 0, 0, 0, 0, time.UTC))
	}
	creator := fixture.accountAddress(t, creatorAddress)
	fixture.accountKeeper.accounts[string(creatorAddress)] =
		authtypes.NewBaseAccountWithAddress(creatorAddress)
	request := validCreateClassRequest(creator)
	request.MaxSupply = maxSupply
	response, err := NewMsgServer(fixture.keeper).CreateClass(
		fixture.ctx,
		request,
	)
	require.NoError(t, err)
	fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())
	return response.ClassId, creator
}

func updateControllerForMintTest(
	t *testing.T,
	fixture *keeperFixture,
	classID string,
	controller string,
	newControllerAddress sdk.AccAddress,
) string {
	t.Helper()
	newController := fixture.accountAddress(t, newControllerAddress)
	fixture.accountKeeper.accounts[string(newControllerAddress)] =
		authtypes.NewBaseAccountWithAddress(newControllerAddress)
	_, err := NewMsgServer(fixture.keeper).UpdateController(
		fixture.ctx,
		&nfttypes.MsgUpdateControllerRequest{
			ClassId:       classID,
			Controller:    controller,
			NewController: newController,
		},
	)
	require.NoError(t, err)
	fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())
	return newController
}

type failingNthSetStoreService struct {
	delegate corestore.KVStoreService
	calls    *int
	failAt   int
}

func (s failingNthSetStoreService) OpenKVStore(ctx context.Context) corestore.KVStore {
	return failingNthSetStore{
		KVStore: s.delegate.OpenKVStore(ctx),
		calls:   s.calls,
		failAt:  s.failAt,
	}
}

type failingNthSetStore struct {
	corestore.KVStore
	calls  *int
	failAt int
}

func (s failingNthSetStore) Set(key, value []byte) error {
	*s.calls++
	if *s.calls == s.failAt {
		return errors.New("forced set failure")
	}
	return s.KVStore.Set(key, value)
}
