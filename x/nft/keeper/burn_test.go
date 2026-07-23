package keeper

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/collections"
	coreaddress "cosmossdk.io/core/address"
	upstreamnft "cosmossdk.io/x/nft"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
)

func TestBurnMovesActiveNFTToPermanentTombstone(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, controller, owner, ownerAddress, data := createNFTForBurnTest(t, &fixture)
	firstToken, found := fixture.keeper.nftKeeper.GetNFT(fixture.ctx, classID, "nft-1")
	require.True(t, found)
	firstLifecycle, err := fixture.keeper.lifecycles.Get(
		fixture.ctx,
		collections.Join(classID, "nft-1"),
	)
	require.NoError(t, err)

	secondRequest := validMintRequest(classID, controller, owner)
	secondRequest.NftId = "nft-2"
	secondRequest.Uri = "https://example.test/nft-2.json"
	secondRequest.UriHash = "sha256:" + strings.Repeat("c", 64)
	_, err = NewMsgServer(fixture.keeper).Mint(
		fixture.ctx,
		secondRequest,
	)
	require.NoError(t, err)
	fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())

	burnedAt := fixture.ctx.BlockTime().Add(time.Hour)
	fixture.ctx = fixture.ctx.WithBlockTime(burnedAt)
	request := &nfttypes.MsgBurnRequest{
		ClassId: classID,
		NftId:   "nft-1",
		Owner:   strings.ToUpper(owner),
	}
	original := *request
	response, err := NewMsgServer(fixture.keeper).Burn(
		fixture.ctx,
		request,
	)
	require.NoError(t, err)
	require.Equal(t, &nfttypes.MsgBurnResponse{}, response)
	require.Equal(t, original, *request)

	_, found = fixture.keeper.nftKeeper.GetNFT(fixture.ctx, classID, "nft-1")
	require.False(t, found)
	require.Empty(t, fixture.keeper.nftKeeper.GetOwner(fixture.ctx, classID, "nft-1"))
	require.Equal(t, uint64(1), fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID))
	require.Equal(t, uint64(1), fixture.keeper.nftKeeper.GetBalance(
		fixture.ctx,
		classID,
		ownerAddress,
	))
	_, err = fixture.keeper.lifecycles.Get(fixture.ctx, collections.Join(classID, "nft-1"))
	require.ErrorIs(t, err, collections.ErrNotFound)

	tombstone, err := fixture.keeper.tombstones.Get(
		fixture.ctx,
		collections.Join(classID, "nft-1"),
	)
	require.NoError(t, err)
	require.Equal(t, classID, tombstone.ClassId)
	require.Equal(t, "nft-1", tombstone.NftId)
	require.Equal(t, firstLifecycle.Mint, tombstone.Mint)
	require.Equal(t, firstToken.Uri, tombstone.Uri)
	require.Equal(t, firstToken.UriHash, tombstone.UriHash)
	require.Equal(t, data.TypeUrl, tombstone.Data.TypeUrl)
	require.Equal(t, data.Value, tombstone.Data.Value)
	require.IsType(t, &nfttypes.BasicNFTData{}, tombstone.Data.GetCachedValue())
	require.Nil(t, tombstone.Revocation)
	require.Equal(t, burnedAt, tombstone.BurnedAt)
	require.Equal(t, owner, tombstone.BurnedBy)

	_, err = fixture.keeper.nftKeeper.NFT(
		fixture.ctx,
		&upstreamnft.QueryNFTRequest{ClassId: classID, Id: "nft-1"},
	)
	require.ErrorIs(t, err, upstreamnft.ErrNFTNotExists)
	ownerResponse, err := fixture.keeper.nftKeeper.Owner(
		fixture.ctx,
		&upstreamnft.QueryOwnerRequest{ClassId: classID, Id: "nft-1"},
	)
	require.NoError(t, err)
	require.Empty(t, ownerResponse.Owner)

	events := fixture.ctx.EventManager().ABCIEvents()
	require.Len(t, events, 1)
	parsedEvent, err := sdk.ParseTypedEvent(events[0])
	require.NoError(t, err)
	require.Equal(t, &upstreamnft.EventBurn{
		ClassId: classID,
		Id:      "nft-1",
		Owner:   owner,
	}, parsedEvent)
	assertClassNFTInvariants(t, &fixture, classID, 2, 1, 1)
}

func TestBurnPreservesRevocationInTombstone(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, controller, owner, _, _ := createNFTForBurnTest(t, &fixture)
	revokedAt := fixture.ctx.BlockTime().Add(time.Hour)
	fixture.ctx = fixture.ctx.WithBlockTime(revokedAt)
	_, err := NewMsgServer(fixture.keeper).Revoke(
		fixture.ctx,
		&nfttypes.MsgRevokeRequest{
			ClassId: classID, NftId: "nft-1", Controller: controller,
		},
	)
	require.NoError(t, err)
	lifecycle, err := fixture.keeper.lifecycles.Get(
		fixture.ctx,
		collections.Join(classID, "nft-1"),
	)
	require.NoError(t, err)
	require.NotNil(t, lifecycle.Revocation)
	fixture.ctx = fixture.ctx.
		WithBlockTime(revokedAt.Add(time.Hour)).
		WithEventManager(sdk.NewEventManager())

	_, err = NewMsgServer(fixture.keeper).Burn(
		fixture.ctx,
		&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "nft-1", Owner: owner},
	)
	require.NoError(t, err)
	tombstone, err := fixture.keeper.tombstones.Get(
		fixture.ctx,
		collections.Join(classID, "nft-1"),
	)
	require.NoError(t, err)
	require.Equal(t, lifecycle.Mint, tombstone.Mint)
	require.Equal(t, lifecycle.Revocation, tombstone.Revocation)
	require.Equal(t, revokedAt, tombstone.Revocation.RevokedAt)
	require.Equal(t, controller, tombstone.Revocation.RevokedBy)
	require.Equal(t, owner, tombstone.BurnedBy)

	assertClassNFTInvariants(t, &fixture, classID, 1, 0, 1)
}

func TestBurnEnforcesCurrentOwner(t *testing.T) {
	for _, tc := range []struct {
		name   string
		signer func(t *testing.T, fixture *keeperFixture, classID, controller, owner string) string
	}{
		{
			name: "controller is not owner",
			signer: func(_ *testing.T, _ *keeperFixture, _, controller, _ string) string {
				return controller
			},
		},
		{
			name: "arbitrary account",
			signer: func(t *testing.T, fixture *keeperFixture, _, _, _ string) string {
				return fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{84}, 20)))
			},
		},
		{
			name: "past owner",
			signer: func(t *testing.T, fixture *keeperFixture, classID, _, owner string) string {
				newOwnerAddress := sdk.AccAddress(bytes.Repeat([]byte{85}, 20))
				newOwner := fixture.accountAddress(t, newOwnerAddress)
				fixture.accountKeeper.accounts[string(newOwnerAddress)] =
					authtypes.NewBaseAccountWithAddress(newOwnerAddress)
				_, err := NewStandardMsgServer(fixture.keeper).Send(
					fixture.ctx,
					&upstreamnft.MsgSend{
						ClassId: classID, Id: "nft-1", Sender: owner, Receiver: newOwner,
					},
				)
				require.NoError(t, err)
				fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())
				return owner
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newKeeperFixture(t, true, true)
			classID, controller, owner, _, _ := createNFTForBurnTest(t, &fixture)
			signer := tc.signer(t, &fixture, classID, controller, owner)
			fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
			before := snapshotRevokeState(t, &fixture, classID, "nft-1")

			_, err := NewMsgServer(fixture.keeper).Burn(
				fixture.ctx,
				&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "nft-1", Owner: signer},
			)
			require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
			require.Equal(t, before, snapshotRevokeState(t, &fixture, classID, "nft-1"))
			require.Empty(t, fixture.ctx.EventManager().Events())
		})
	}
}

func TestBurnReturnsNFTNotExistsForUnusedAndBurnedIDs(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, _, owner, _, _ := createNFTForBurnTest(t, &fixture)
	server := NewMsgServer(fixture.keeper)
	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))

	_, err := server.Burn(
		fixture.ctx,
		&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "unused", Owner: owner},
	)
	require.ErrorIs(t, err, upstreamnft.ErrNFTNotExists)
	require.Empty(t, fixture.ctx.EventManager().Events())

	_, err = server.Burn(
		fixture.ctx,
		&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "nft-1", Owner: owner},
	)
	require.NoError(t, err)
	fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())
	_, err = server.Burn(
		fixture.ctx,
		&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "nft-1", Owner: owner},
	)
	require.ErrorIs(t, err, upstreamnft.ErrNFTNotExists)
	require.Empty(t, fixture.ctx.EventManager().Events())
}

func TestBurnRejectsInvalidRequests(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, _, owner, ownerAddress, _ := createNFTForBurnTest(t, &fixture)
	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	creator := strings.Split(classID, ":")[0]
	mixedOwner := strings.ToUpper(owner[:1]) + owner[1:]
	moduleOwner := fixture.accountAddress(t, fixture.accountKeeper.moduleAddress)

	for _, tc := range []struct {
		name      string
		request   *nfttypes.MsgBurnRequest
		targetErr error
	}{
		{name: "nil request", targetErr: sdkerrors.ErrInvalidRequest},
		{
			name:      "empty class ID",
			request:   &nfttypes.MsgBurnRequest{NftId: "nft-1", Owner: owner},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "non-canonical class creator",
			request: &nfttypes.MsgBurnRequest{
				ClassId: strings.ToUpper(creator) + ":certificate",
				NftId:   "nft-1",
				Owner:   owner,
			},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name:      "empty NFT ID",
			request:   &nfttypes.MsgBurnRequest{ClassId: classID, Owner: owner},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "invalid NFT ID",
			request: &nfttypes.MsgBurnRequest{
				ClassId: classID, NftId: ".", Owner: owner,
			},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "invalid owner",
			request: &nfttypes.MsgBurnRequest{
				ClassId: classID, NftId: "nft-1", Owner: "invalid",
			},
			targetErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "mixed-case owner",
			request: &nfttypes.MsgBurnRequest{
				ClassId: classID, NftId: "nft-1", Owner: mixedOwner,
			},
			targetErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "module account owner",
			request: &nfttypes.MsgBurnRequest{
				ClassId: classID, NftId: "nft-1", Owner: moduleOwner,
			},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := snapshotRevokeState(t, &fixture, classID, "nft-1")
			_, err := NewMsgServer(fixture.keeper).Burn(
				fixture.ctx,
				tc.request,
			)
			require.ErrorIs(t, err, tc.targetErr)
			require.Equal(t, before, snapshotRevokeState(t, &fixture, classID, "nft-1"))
			require.Equal(t, ownerAddress, fixture.keeper.nftKeeper.GetOwner(
				fixture.ctx,
				classID,
				"nft-1",
			))
			require.Empty(t, fixture.ctx.EventManager().Events())
		})
	}
}

func TestBurnRejectsInvalidBlockTime(t *testing.T) {
	for _, tc := range []struct {
		name          string
		expectedError string
		prepare       func(t *testing.T, fixture *keeperFixture, classID, controller string) time.Time
	}{
		{
			name:          "zero block time",
			expectedError: "at zero block time",
			prepare:       func(_ *testing.T, _ *keeperFixture, _, _ string) time.Time { return time.Time{} },
		},
		{
			name:          "before mint time",
			expectedError: "before its mint time",
			prepare: func(t *testing.T, fixture *keeperFixture, classID, _ string) time.Time {
				lifecycle, err := fixture.keeper.lifecycles.Get(
					fixture.ctx,
					collections.Join(classID, "nft-1"),
				)
				require.NoError(t, err)
				return lifecycle.Mint.MintedAt.Add(-time.Second)
			},
		},
		{
			name:          "before revocation time",
			expectedError: "before its revocation time",
			prepare: func(t *testing.T, fixture *keeperFixture, classID, controller string) time.Time {
				revokedAt := fixture.ctx.BlockTime().Add(time.Hour)
				fixture.ctx = fixture.ctx.WithBlockTime(revokedAt)
				_, err := NewMsgServer(fixture.keeper).Revoke(
					fixture.ctx,
					&nfttypes.MsgRevokeRequest{
						ClassId: classID, NftId: "nft-1", Controller: controller,
					},
				)
				require.NoError(t, err)
				fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())
				return revokedAt.Add(-time.Second)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newKeeperFixture(t, true, true)
			classID, controller, owner, _, _ := createNFTForBurnTest(t, &fixture)
			fixture.ctx = fixture.ctx.WithBlockTime(tc.prepare(t, &fixture, classID, controller))
			before := snapshotRevokeState(t, &fixture, classID, "nft-1")

			_, err := NewMsgServer(fixture.keeper).Burn(
				fixture.ctx,
				&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "nft-1", Owner: owner},
			)
			require.ErrorContains(t, err, tc.expectedError)
			require.Equal(t, before, snapshotRevokeState(t, &fixture, classID, "nft-1"))
			require.Empty(t, fixture.ctx.EventManager().Events())
		})
	}
}

func TestBurnRejectsInconsistentLiveState(t *testing.T) {
	for _, tc := range []struct {
		name          string
		expectedError string
		mutate        func(t *testing.T, fixture *keeperFixture, classID string)
	}{
		{
			name:          "standard NFT without lifecycle",
			expectedError: "inconsistent standard, lifecycle, and tombstone state",
			mutate: func(t *testing.T, fixture *keeperFixture, classID string) {
				require.NoError(t, fixture.keeper.lifecycles.Remove(
					fixture.ctx,
					collections.Join(classID, "nft-1"),
				))
			},
		},
		{
			name:          "lifecycle without standard NFT",
			expectedError: "inconsistent standard, lifecycle, and tombstone state",
			mutate: func(t *testing.T, fixture *keeperFixture, classID string) {
				require.NoError(t, fixture.keeper.nftKeeper.Burn(fixture.ctx, classID, "nft-1"))
				fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())
			},
		},
		{
			name:          "live NFT with tombstone",
			expectedError: "inconsistent standard, lifecycle, and tombstone state",
			mutate: func(t *testing.T, fixture *keeperFixture, classID string) {
				require.NoError(t, fixture.keeper.tombstones.Set(
					fixture.ctx,
					collections.Join(classID, "nft-1"),
					nfttypes.BurnTombstone{ClassId: classID, NftId: "nft-1"},
				))
			},
		},
		{
			name:          "live NFT without class policy",
			expectedError: "inconsistent standard and policy state",
			mutate: func(t *testing.T, fixture *keeperFixture, classID string) {
				require.NoError(t, fixture.keeper.classPolicies.Remove(fixture.ctx, classID))
			},
		},
		{
			name:          "invalid stored URI metadata",
			expectedError: "invalid stored URI metadata",
			mutate: func(t *testing.T, fixture *keeperFixture, classID string) {
				token, found := fixture.keeper.nftKeeper.GetNFT(fixture.ctx, classID, "nft-1")
				require.True(t, found)
				token.Uri = ""
				require.NoError(t, fixture.keeper.nftKeeper.Update(fixture.ctx, token))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newKeeperFixture(t, true, true)
			classID, _, owner, _, _ := createNFTForBurnTest(t, &fixture)
			tc.mutate(t, &fixture, classID)
			fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
			before := snapshotRevokeState(t, &fixture, classID, "nft-1")

			_, err := NewMsgServer(fixture.keeper).Burn(
				fixture.ctx,
				&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "nft-1", Owner: owner},
			)
			require.ErrorContains(t, err, tc.expectedError)
			require.Equal(t, before, snapshotRevokeState(t, &fixture, classID, "nft-1"))
			require.Empty(t, fixture.ctx.EventManager().Events())
		})
	}
}

func TestBurnRollsBackWhenTombstoneWriteFails(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, _, owner, _, _ := createNFTForBurnTest(t, &fixture)
	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	before := snapshotRevokeState(t, &fixture, classID, "nft-1")
	ownerAddress := before.owner
	balanceBefore := fixture.keeper.nftKeeper.GetBalance(fixture.ctx, classID, ownerAddress)
	ownerNFTsBefore := fixture.keeper.nftKeeper.GetNFTsOfClassByOwner(
		fixture.ctx,
		classID,
		ownerAddress,
	)
	setCalls := 0
	failingKeeper := NewKeeper(
		fixture.cdc,
		runtime.NewKVStoreService(fixture.nftService),
		failingNthSetStoreService{
			delegate: runtime.NewKVStoreService(fixture.policyService),
			calls:    &setCalls,
			failAt:   1,
		},
		fixture.accountKeeper,
		testBankKeeper{},
		fixture.moduleAccountAddresses,
	)

	_, err := NewMsgServer(failingKeeper).Burn(
		fixture.ctx,
		&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "nft-1", Owner: owner},
	)
	require.ErrorContains(t, err, "forced set failure")
	require.Equal(t, before, snapshotRevokeState(t, &fixture, classID, "nft-1"))
	require.Equal(t, balanceBefore, fixture.keeper.nftKeeper.GetBalance(
		fixture.ctx,
		classID,
		ownerAddress,
	))
	require.Equal(t, ownerNFTsBefore, fixture.keeper.nftKeeper.GetNFTsOfClassByOwner(
		fixture.ctx,
		classID,
		ownerAddress,
	))
	require.Empty(t, fixture.ctx.EventManager().Events())
}

func TestBurnRollsBackWhenUpstreamBurnEventFails(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, _, owner, ownerAddress, _ := createNFTForBurnTest(t, &fixture)
	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	before := snapshotRevokeState(t, &fixture, classID, "nft-1")
	balanceBefore := fixture.keeper.nftKeeper.GetBalance(fixture.ctx, classID, ownerAddress)
	ownerNFTsBefore := fixture.keeper.nftKeeper.GetNFTsOfClassByOwner(
		fixture.ctx,
		classID,
		ownerAddress,
	)
	bytesToStringCalls := 0
	failingAccountKeeper := *fixture.accountKeeper
	failingAccountKeeper.addressCodec = failingNthBytesToStringCodec{
		delegate: fixture.accountKeeper.addressCodec,
		calls:    &bytesToStringCalls,
		failAt:   6,
	}
	failingKeeper := NewKeeper(
		fixture.cdc,
		runtime.NewKVStoreService(fixture.nftService),
		runtime.NewKVStoreService(fixture.policyService),
		&failingAccountKeeper,
		testBankKeeper{},
		fixture.moduleAccountAddresses,
	)

	_, err := NewMsgServer(failingKeeper).Burn(
		fixture.ctx,
		&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "nft-1", Owner: owner},
	)
	require.ErrorContains(t, err, "emitted 0 events instead of one")
	require.Equal(t, 6, bytesToStringCalls)
	require.Equal(t, before, snapshotRevokeState(t, &fixture, classID, "nft-1"))
	require.Equal(t, balanceBefore, fixture.keeper.nftKeeper.GetBalance(
		fixture.ctx,
		classID,
		ownerAddress,
	))
	require.Equal(t, ownerNFTsBefore, fixture.keeper.nftKeeper.GetNFTsOfClassByOwner(
		fixture.ctx,
		classID,
		ownerAddress,
	))
	require.Empty(t, fixture.ctx.EventManager().Events())
}

func createNFTForBurnTest(
	t *testing.T,
	fixture *keeperFixture,
) (string, string, string, sdk.AccAddress, *types.Any) {
	t.Helper()
	fixture.ctx = fixture.ctx.WithBlockTime(time.Date(2026, time.July, 25, 0, 0, 0, 0, time.UTC))
	controllerAddress := sdk.AccAddress(bytes.Repeat([]byte{81}, 20))
	classID, controller := createClassForMintTest(t, fixture, controllerAddress, 10)
	ownerAddress := sdk.AccAddress(bytes.Repeat([]byte{82}, 32))
	owner := fixture.accountAddress(t, ownerAddress)
	fixture.accountKeeper.accounts[string(ownerAddress)] =
		authtypes.NewBaseAccountWithAddress(ownerAddress)
	data, err := types.NewAnyWithValue(&nfttypes.BasicNFTData{
		Name:        "Certificate #1",
		Description: "Completion certificate",
		ImageUri:    "https://example.test/certificate.png",
	})
	require.NoError(t, err)
	request := validMintRequest(classID, controller, owner)
	request.Data = data
	_, err = NewMsgServer(fixture.keeper).Mint(fixture.ctx, request)
	require.NoError(t, err)
	fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())
	return classID, controller, owner, ownerAddress, data
}

func assertClassNFTInvariants(
	t *testing.T,
	fixture *keeperFixture,
	classID string,
	wantMintedCount uint64,
	wantLive uint64,
	wantTombstones uint64,
) {
	t.Helper()
	mintedCount, err := fixture.keeper.mintedCounts.Get(fixture.ctx, classID)
	require.NoError(t, err)
	require.Equal(t, wantMintedCount, mintedCount)
	require.Equal(t, wantLive, fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID))

	var liveCount uint64
	err = fixture.keeper.lifecycles.Walk(
		fixture.ctx,
		collections.NewPrefixedPairRange[string, string](classID),
		func(_ collections.Pair[string, string], _ nfttypes.LifecycleRecord) (bool, error) {
			liveCount++
			return false, nil
		},
	)
	require.NoError(t, err)
	var tombstoneCount uint64
	err = fixture.keeper.tombstones.Walk(
		fixture.ctx,
		collections.NewPrefixedPairRange[string, string](classID),
		func(_ collections.Pair[string, string], _ nfttypes.BurnTombstone) (bool, error) {
			tombstoneCount++
			return false, nil
		},
	)
	require.NoError(t, err)
	var ownerClassCount uint64
	err = fixture.keeper.ownerClassCounts.Walk(
		fixture.ctx,
		collections.NewPrefixedPairRange[string, string](classID),
		func(_ collections.Pair[string, string], count uint64) (bool, error) {
			require.NotZero(t, count)
			ownerClassCount += count
			return false, nil
		},
	)
	require.NoError(t, err)
	require.Equal(t, wantLive, liveCount)
	require.Equal(t, wantTombstones, tombstoneCount)
	require.Equal(t, wantLive, ownerClassCount)
	require.Equal(t, mintedCount, liveCount+tombstoneCount)
	require.Equal(t, fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID), liveCount)
}

type failingNthBytesToStringCodec struct {
	delegate coreaddress.Codec
	calls    *int
	failAt   int
}

func (c failingNthBytesToStringCodec) StringToBytes(value string) ([]byte, error) {
	return c.delegate.StringToBytes(value)
}

func (c failingNthBytesToStringCodec) BytesToString(value []byte) (string, error) {
	*c.calls++
	if *c.calls == c.failAt {
		return "", fmt.Errorf("forced address encoding failure")
	}
	return c.delegate.BytesToString(value)
}
