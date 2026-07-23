package keeper

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/collections"
	upstreamnft "cosmossdk.io/x/nft"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
)

func TestRevokeTransitionsActiveNFTToRevoked(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, creator, owner, ownerAddress := createNFTForRevokeTest(t, &fixture, true)
	controller := updateControllerForMintTest(
		t,
		&fixture,
		classID,
		creator,
		sdk.AccAddress(bytes.Repeat([]byte{72}, 32)),
	)
	revokedAt := fixture.ctx.BlockTime().Add(time.Hour)
	fixture.ctx = fixture.ctx.WithBlockTime(revokedAt)
	tokenBefore, found := fixture.keeper.nftKeeper.GetNFT(fixture.ctx, classID, "nft-1")
	require.True(t, found)
	lifecycleBefore, err := fixture.keeper.lifecycles.Get(
		fixture.ctx,
		collections.Join(classID, "nft-1"),
	)
	require.NoError(t, err)
	classBefore, err := fixture.keeper.getClassRecord(fixture.ctx, classID)
	require.NoError(t, err)

	response, err := NewMsgServer(fixture.keeper).Revoke(
		fixture.ctx,
		&nfttypes.MsgRevokeRequest{
			ClassId:    classID,
			NftId:      "nft-1",
			Controller: strings.ToUpper(controller),
		},
	)
	require.NoError(t, err)
	require.Equal(t, &nfttypes.MsgRevokeResponse{}, response)

	lifecycleAfter, err := fixture.keeper.lifecycles.Get(
		fixture.ctx,
		collections.Join(classID, "nft-1"),
	)
	require.NoError(t, err)
	require.Equal(t, lifecycleBefore.Mint, lifecycleAfter.Mint)
	require.Equal(t, creator, lifecycleAfter.Mint.MintedBy)
	require.NotEqual(t, controller, lifecycleAfter.Mint.MintedBy)
	require.Equal(t, &nfttypes.Revocation{
		RevokedAt: revokedAt,
		RevokedBy: controller,
	}, lifecycleAfter.Revocation)
	tokenAfter, found := fixture.keeper.nftKeeper.GetNFT(fixture.ctx, classID, "nft-1")
	require.True(t, found)
	require.Equal(t, tokenBefore, tokenAfter)
	require.Equal(t, ownerAddress, fixture.keeper.nftKeeper.GetOwner(
		fixture.ctx,
		classID,
		"nft-1",
	))
	require.Equal(t, uint64(1), fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID))
	classAfter, err := fixture.keeper.getClassRecord(fixture.ctx, classID)
	require.NoError(t, err)
	require.Equal(t, classBefore, classAfter)
	require.Equal(t, uint64(1), classAfter.MintedCount)

	queryResponse, err := NewQueryServer(fixture.keeper).NFTRecord(
		fixture.ctx,
		&nfttypes.QueryNFTRecordRequest{ClassId: classID, NftId: "nft-1"},
	)
	require.NoError(t, err)
	live := queryResponse.NftRecord.GetLive()
	require.NotNil(t, live)
	require.Equal(t, owner, live.Owner)
	require.Equal(t, nfttypes.LiveNFTStatus_LIVE_NFT_STATUS_REVOKED, live.Status)
	require.Equal(t, lifecycleAfter.Revocation, live.Revocation)

	events := fixture.ctx.EventManager().ABCIEvents()
	require.Len(t, events, 1)
	parsedEvent, err := sdk.ParseTypedEvent(events[0])
	require.NoError(t, err)
	require.Equal(t, &nfttypes.EventNFTRevoked{
		ClassId:    classID,
		NftId:      "nft-1",
		Controller: controller,
	}, parsedEvent)

	fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())
	receiver := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{73}, 20)))
	_, err = NewStandardMsgServer(fixture.keeper).Send(
		fixture.ctx,
		&upstreamnft.MsgSend{
			ClassId: classID, Id: "nft-1", Sender: owner, Receiver: receiver,
		},
	)
	require.ErrorIs(t, err, nfttypes.ErrNFTRevoked)
	require.Equal(t, ownerAddress, fixture.keeper.nftKeeper.GetOwner(
		fixture.ctx,
		classID,
		"nft-1",
	))
	require.Empty(t, fixture.ctx.EventManager().Events())
}

func TestRevokeEnforcesControllerAndClassPolicy(t *testing.T) {
	t.Run("non-revocable class", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, controller, _, ownerAddress := createNFTForRevokeTest(t, &fixture, false)
		fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))

		_, err := NewMsgServer(fixture.keeper).Revoke(
			fixture.ctx,
			&nfttypes.MsgRevokeRequest{
				ClassId: classID, NftId: "nft-1", Controller: controller,
			},
		)
		require.ErrorIs(t, err, nfttypes.ErrRevocationNotAllowed)
		assertFailedRevokeState(t, &fixture, classID, ownerAddress)
	})

	t.Run("past controller", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, creator, _, ownerAddress := createNFTForRevokeTest(t, &fixture, true)
		_ = updateControllerForMintTest(
			t,
			&fixture,
			classID,
			creator,
			sdk.AccAddress(bytes.Repeat([]byte{74}, 20)),
		)
		fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))

		_, err := NewMsgServer(fixture.keeper).Revoke(
			fixture.ctx,
			&nfttypes.MsgRevokeRequest{
				ClassId: classID, NftId: "nft-1", Controller: creator,
			},
		)
		require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
		assertFailedRevokeState(t, &fixture, classID, ownerAddress)
	})

	t.Run("arbitrary account", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, _, _, ownerAddress := createNFTForRevokeTest(t, &fixture, true)
		arbitrary := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{75}, 20)))
		fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))

		_, err := NewMsgServer(fixture.keeper).Revoke(
			fixture.ctx,
			&nfttypes.MsgRevokeRequest{
				ClassId: classID, NftId: "nft-1", Controller: arbitrary,
			},
		)
		require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
		assertFailedRevokeState(t, &fixture, classID, ownerAddress)
	})
}

func TestRevokeIsIrreversible(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, controller, _, ownerAddress := createNFTForRevokeTest(t, &fixture, true)
	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	server := NewMsgServer(fixture.keeper)
	request := &nfttypes.MsgRevokeRequest{
		ClassId: classID, NftId: "nft-1", Controller: controller,
	}

	_, err := server.Revoke(fixture.ctx, request)
	require.NoError(t, err)
	first, err := fixture.keeper.lifecycles.Get(
		fixture.ctx,
		collections.Join(classID, "nft-1"),
	)
	require.NoError(t, err)
	fixture.ctx = fixture.ctx.
		WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour)).
		WithEventManager(sdk.NewEventManager())

	_, err = server.Revoke(fixture.ctx, request)
	require.ErrorIs(t, err, nfttypes.ErrNFTRevoked)
	second, err := fixture.keeper.lifecycles.Get(
		fixture.ctx,
		collections.Join(classID, "nft-1"),
	)
	require.NoError(t, err)
	require.Equal(t, first, second)
	require.Equal(t, ownerAddress, fixture.keeper.nftKeeper.GetOwner(
		fixture.ctx,
		classID,
		"nft-1",
	))
	require.Empty(t, fixture.ctx.EventManager().Events())
}

func TestRevokeRejectsInvalidBlockTime(t *testing.T) {
	for _, tc := range []struct {
		name          string
		expectedError string
		blockTime     func(mintedAt time.Time) time.Time
	}{
		{
			name:          "zero block time",
			expectedError: "at zero block time",
			blockTime:     func(time.Time) time.Time { return time.Time{} },
		},
		{
			name:          "before mint time",
			expectedError: "before its mint time",
			blockTime: func(mintedAt time.Time) time.Time {
				return mintedAt.Add(-time.Second)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newKeeperFixture(t, true, true)
			classID, controller, _, ownerAddress := createNFTForRevokeTest(t, &fixture, true)
			lifecycle, err := fixture.keeper.lifecycles.Get(
				fixture.ctx,
				collections.Join(classID, "nft-1"),
			)
			require.NoError(t, err)
			fixture.ctx = fixture.ctx.WithBlockTime(tc.blockTime(lifecycle.Mint.MintedAt))
			before := snapshotRevokeState(t, &fixture, classID, "nft-1")

			_, err = NewMsgServer(fixture.keeper).Revoke(
				fixture.ctx,
				&nfttypes.MsgRevokeRequest{
					ClassId: classID, NftId: "nft-1", Controller: controller,
				},
			)
			require.ErrorContains(t, err, tc.expectedError)
			require.Equal(t, before, snapshotRevokeState(t, &fixture, classID, "nft-1"))
			assertFailedRevokeState(t, &fixture, classID, ownerAddress)
		})
	}
}

func TestRevokeReturnsNFTNotExistsForUnusedAndBurnedIDs(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, controller := createClassForMintTest(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{76}, 20)),
		10,
	)
	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	server := NewMsgServer(fixture.keeper)

	_, err := server.Revoke(
		fixture.ctx,
		&nfttypes.MsgRevokeRequest{
			ClassId: classID, NftId: "unused", Controller: controller,
		},
	)
	require.ErrorIs(t, err, upstreamnft.ErrNFTNotExists)

	require.NoError(t, fixture.keeper.tombstones.Set(
		fixture.ctx,
		collections.Join(classID, "burned"),
		nfttypes.BurnTombstone{ClassId: classID, NftId: "burned"},
	))
	_, err = server.Revoke(
		fixture.ctx,
		&nfttypes.MsgRevokeRequest{
			ClassId: classID, NftId: "burned", Controller: controller,
		},
	)
	require.ErrorIs(t, err, upstreamnft.ErrNFTNotExists)
	require.Empty(t, fixture.ctx.EventManager().Events())
}

func TestRevokeRejectsInvalidRequests(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, controller, _, ownerAddress := createNFTForRevokeTest(t, &fixture, true)
	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	creator := strings.Split(classID, ":")[0]
	mixedController := strings.ToUpper(controller[:1]) + controller[1:]

	for _, tc := range []struct {
		name      string
		request   *nfttypes.MsgRevokeRequest
		targetErr error
	}{
		{name: "nil request", targetErr: sdkerrors.ErrInvalidRequest},
		{
			name:      "empty class ID",
			request:   &nfttypes.MsgRevokeRequest{NftId: "nft-1", Controller: controller},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "malformed class creator",
			request: &nfttypes.MsgRevokeRequest{
				ClassId: "badbech32:certificate", NftId: "nft-1", Controller: controller,
			},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "non-canonical class creator",
			request: &nfttypes.MsgRevokeRequest{
				ClassId: strings.ToUpper(creator) + ":certificate",
				NftId:   "nft-1", Controller: controller,
			},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "empty NFT ID",
			request: &nfttypes.MsgRevokeRequest{
				ClassId: classID, Controller: controller,
			},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "invalid NFT ID",
			request: &nfttypes.MsgRevokeRequest{
				ClassId: classID, NftId: ".", Controller: controller,
			},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "invalid controller",
			request: &nfttypes.MsgRevokeRequest{
				ClassId: classID, NftId: "nft-1", Controller: "invalid",
			},
			targetErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "mixed-case controller",
			request: &nfttypes.MsgRevokeRequest{
				ClassId: classID, NftId: "nft-1", Controller: mixedController,
			},
			targetErr: sdkerrors.ErrInvalidAddress,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewMsgServer(fixture.keeper).Revoke(
				fixture.ctx,
				tc.request,
			)
			require.ErrorIs(t, err, tc.targetErr)
			assertFailedRevokeState(t, &fixture, classID, ownerAddress)
		})
	}
}

func TestRevokeRejectsInconsistentNFTState(t *testing.T) {
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
			name:          "revoked NFT without class policy",
			expectedError: "inconsistent standard and policy state",
			mutate: func(t *testing.T, fixture *keeperFixture, classID string) {
				key := collections.Join(classID, "nft-1")
				lifecycle, err := fixture.keeper.lifecycles.Get(fixture.ctx, key)
				require.NoError(t, err)
				lifecycle.Revocation = &nfttypes.Revocation{
					RevokedAt: lifecycle.Mint.MintedAt,
					RevokedBy: lifecycle.Mint.MintedBy,
				}
				require.NoError(t, fixture.keeper.lifecycles.Set(fixture.ctx, key, lifecycle))
				require.NoError(t, fixture.keeper.classPolicies.Remove(fixture.ctx, classID))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newKeeperFixture(t, true, true)
			classID, controller, _, _ := createNFTForRevokeTest(t, &fixture, true)
			tc.mutate(t, &fixture, classID)
			fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
			before := snapshotRevokeState(t, &fixture, classID, "nft-1")

			_, err := NewMsgServer(fixture.keeper).Revoke(
				fixture.ctx,
				&nfttypes.MsgRevokeRequest{
					ClassId: classID, NftId: "nft-1", Controller: controller,
				},
			)
			require.ErrorContains(t, err, tc.expectedError)
			require.Equal(t, before, snapshotRevokeState(t, &fixture, classID, "nft-1"))
			require.Empty(t, fixture.ctx.EventManager().Events())
		})
	}
}

func TestLoadNFTStateRejectsInvalidRevocation(t *testing.T) {
	for _, tc := range []struct {
		name          string
		expectedError string
		revocation    func(lifecycle nfttypes.LifecycleRecord) *nfttypes.Revocation
	}{
		{
			name:          "zero revocation time",
			expectedError: "no revocation time",
			revocation: func(lifecycle nfttypes.LifecycleRecord) *nfttypes.Revocation {
				return &nfttypes.Revocation{RevokedBy: lifecycle.Mint.MintedBy}
			},
		},
		{
			name:          "revocation predates mint",
			expectedError: "revocation predates mint",
			revocation: func(lifecycle nfttypes.LifecycleRecord) *nfttypes.Revocation {
				return &nfttypes.Revocation{
					RevokedAt: lifecycle.Mint.MintedAt.Add(-time.Second),
					RevokedBy: lifecycle.Mint.MintedBy,
				}
			},
		},
		{
			name:          "invalid revoker",
			expectedError: "invalid revoker",
			revocation: func(lifecycle nfttypes.LifecycleRecord) *nfttypes.Revocation {
				return &nfttypes.Revocation{
					RevokedAt: lifecycle.Mint.MintedAt.Add(time.Second),
					RevokedBy: "invalid",
				}
			},
		},
		{
			name:          "non-canonical revoker",
			expectedError: "revoker is not canonical",
			revocation: func(lifecycle nfttypes.LifecycleRecord) *nfttypes.Revocation {
				return &nfttypes.Revocation{
					RevokedAt: lifecycle.Mint.MintedAt.Add(time.Second),
					RevokedBy: strings.ToUpper(lifecycle.Mint.MintedBy),
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newKeeperFixture(t, true, true)
			classID, controller, _, _ := createNFTForRevokeTest(t, &fixture, true)
			key := collections.Join(classID, "nft-1")
			lifecycle, err := fixture.keeper.lifecycles.Get(fixture.ctx, key)
			require.NoError(t, err)
			lifecycle.Revocation = tc.revocation(lifecycle)
			require.NoError(t, fixture.keeper.lifecycles.Set(fixture.ctx, key, lifecycle))
			before := snapshotRevokeState(t, &fixture, classID, "nft-1")

			_, err = NewMsgServer(fixture.keeper).Revoke(
				fixture.ctx,
				&nfttypes.MsgRevokeRequest{
					ClassId: classID, NftId: "nft-1", Controller: controller,
				},
			)
			require.ErrorContains(t, err, tc.expectedError)
			require.Equal(t, before, snapshotRevokeState(t, &fixture, classID, "nft-1"))
			require.Empty(t, fixture.ctx.EventManager().Events())
		})
	}
}

func TestRevokeRollsBackWhenLifecycleWriteFails(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, controller, _, ownerAddress := createNFTForRevokeTest(t, &fixture, true)
	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	original, err := fixture.keeper.lifecycles.Get(
		fixture.ctx,
		collections.Join(classID, "nft-1"),
	)
	require.NoError(t, err)
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

	_, err = NewMsgServer(failingKeeper).Revoke(
		fixture.ctx,
		&nfttypes.MsgRevokeRequest{
			ClassId: classID, NftId: "nft-1", Controller: controller,
		},
	)
	require.ErrorContains(t, err, "forced set failure")
	stored, err := fixture.keeper.lifecycles.Get(
		fixture.ctx,
		collections.Join(classID, "nft-1"),
	)
	require.NoError(t, err)
	require.Equal(t, original, stored)
	require.Equal(t, ownerAddress, fixture.keeper.nftKeeper.GetOwner(
		fixture.ctx,
		classID,
		"nft-1",
	))
	require.Equal(t, uint64(1), fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID))
	require.Empty(t, fixture.ctx.EventManager().Events())
}

func createNFTForRevokeTest(
	t *testing.T,
	fixture *keeperFixture,
	revocable bool,
) (string, string, string, sdk.AccAddress) {
	t.Helper()
	fixture.ctx = fixture.ctx.WithBlockTime(time.Date(2026, time.July, 24, 0, 0, 0, 0, time.UTC))
	controllerAddress := sdk.AccAddress(bytes.Repeat([]byte{70}, 20))
	controller := fixture.accountAddress(t, controllerAddress)
	fixture.accountKeeper.accounts[string(controllerAddress)] =
		authtypes.NewBaseAccountWithAddress(controllerAddress)
	classRequest := validCreateClassRequest(controller)
	classRequest.Revocable = revocable
	classResponse, err := NewMsgServer(fixture.keeper).CreateClass(
		fixture.ctx,
		classRequest,
	)
	require.NoError(t, err)
	fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())

	ownerAddress := sdk.AccAddress(bytes.Repeat([]byte{71}, 20))
	owner := fixture.accountAddress(t, ownerAddress)
	fixture.accountKeeper.accounts[string(ownerAddress)] =
		authtypes.NewBaseAccountWithAddress(ownerAddress)
	_, err = NewMsgServer(fixture.keeper).Mint(
		fixture.ctx,
		validMintRequest(classResponse.ClassId, controller, owner),
	)
	require.NoError(t, err)
	fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())
	return classResponse.ClassId, controller, owner, ownerAddress
}

func assertFailedRevokeState(
	t *testing.T,
	fixture *keeperFixture,
	classID string,
	expectedOwner sdk.AccAddress,
) {
	t.Helper()
	lifecycle, err := fixture.keeper.lifecycles.Get(
		fixture.ctx,
		collections.Join(classID, "nft-1"),
	)
	require.NoError(t, err)
	require.Nil(t, lifecycle.Revocation)
	require.Equal(t, expectedOwner, fixture.keeper.nftKeeper.GetOwner(
		fixture.ctx,
		classID,
		"nft-1",
	))
	require.Equal(t, uint64(1), fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID))
	require.Empty(t, fixture.ctx.EventManager().Events())
}

type revokeStateSnapshot struct {
	class          upstreamnft.Class
	hasClass       bool
	classPolicy    nfttypes.ClassPolicy
	hasClassPolicy bool
	mintedCount    uint64
	hasMintedCount bool
	token          upstreamnft.NFT
	hasToken       bool
	owner          sdk.AccAddress
	supply         uint64
	lifecycle      nfttypes.LifecycleRecord
	hasLifecycle   bool
	tombstone      nfttypes.BurnTombstone
	hasTombstone   bool
}

func snapshotRevokeState(
	t *testing.T,
	fixture *keeperFixture,
	classID string,
	nftID string,
) revokeStateSnapshot {
	t.Helper()
	snapshot := revokeStateSnapshot{}
	snapshot.class, snapshot.hasClass = fixture.keeper.nftKeeper.GetClass(fixture.ctx, classID)
	snapshot.token, snapshot.hasToken = fixture.keeper.nftKeeper.GetNFT(fixture.ctx, classID, nftID)
	snapshot.owner = append(sdk.AccAddress(nil), fixture.keeper.nftKeeper.GetOwner(
		fixture.ctx,
		classID,
		nftID,
	)...)
	snapshot.supply = fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID)

	var err error
	snapshot.hasClassPolicy, err = fixture.keeper.classPolicies.Has(fixture.ctx, classID)
	require.NoError(t, err)
	if snapshot.hasClassPolicy {
		snapshot.classPolicy, err = fixture.keeper.classPolicies.Get(fixture.ctx, classID)
		require.NoError(t, err)
	}
	snapshot.hasMintedCount, err = fixture.keeper.mintedCounts.Has(fixture.ctx, classID)
	require.NoError(t, err)
	if snapshot.hasMintedCount {
		snapshot.mintedCount, err = fixture.keeper.mintedCounts.Get(fixture.ctx, classID)
		require.NoError(t, err)
	}
	key := collections.Join(classID, nftID)
	snapshot.hasLifecycle, err = fixture.keeper.lifecycles.Has(fixture.ctx, key)
	require.NoError(t, err)
	if snapshot.hasLifecycle {
		snapshot.lifecycle, err = fixture.keeper.lifecycles.Get(fixture.ctx, key)
		require.NoError(t, err)
	}
	snapshot.hasTombstone, err = fixture.keeper.tombstones.Has(fixture.ctx, key)
	require.NoError(t, err)
	if snapshot.hasTombstone {
		snapshot.tombstone, err = fixture.keeper.tombstones.Get(fixture.ctx, key)
		require.NoError(t, err)
	}
	return snapshot
}
