package keeper

import (
	"bytes"
	"math"
	"testing"

	"cosmossdk.io/collections"
	upstreamnft "cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
)

func TestOwnerClassCountsTrackMintSendBurn(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, controller := createClassForMintTest(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{91}, 20)),
		10,
	)
	firstOwnerAddress := sdk.AccAddress(bytes.Repeat([]byte{92}, 20))
	firstOwner := fixture.accountAddress(t, firstOwnerAddress)
	fixture.accountKeeper.accounts[string(firstOwnerAddress)] =
		authtypes.NewBaseAccountWithAddress(firstOwnerAddress)
	secondOwnerAddress := sdk.AccAddress(bytes.Repeat([]byte{93}, 32))
	secondOwner := fixture.accountAddress(t, secondOwnerAddress)
	fixture.accountKeeper.accounts[string(secondOwnerAddress)] =
		authtypes.NewBaseAccountWithAddress(secondOwnerAddress)

	for _, nftID := range []string{"nft-1", "nft-2"} {
		request := validMintRequest(classID, controller, firstOwner)
		request.NftId = nftID
		_, err := NewMsgServer(fixture.keeper).Mint(
			fixture.ctx,
			request,
		)
		require.NoError(t, err)
	}
	requireOwnerClassCount(t, &fixture, classID, firstOwner, 2)
	requireOwnerClassCountMissing(t, &fixture, classID, secondOwner)

	_, err := NewStandardMsgServer(fixture.keeper).Send(
		fixture.ctx,
		&upstreamnft.MsgSend{
			ClassId:  classID,
			Id:       "nft-2",
			Sender:   firstOwner,
			Receiver: firstOwner,
		},
	)
	require.NoError(t, err)
	requireOwnerClassCount(t, &fixture, classID, firstOwner, 2)

	_, err = NewStandardMsgServer(fixture.keeper).Send(
		fixture.ctx,
		&upstreamnft.MsgSend{
			ClassId:  classID,
			Id:       "nft-1",
			Sender:   firstOwner,
			Receiver: secondOwner,
		},
	)
	require.NoError(t, err)
	requireOwnerClassCount(t, &fixture, classID, firstOwner, 1)
	requireOwnerClassCount(t, &fixture, classID, secondOwner, 1)

	_, err = NewMsgServer(fixture.keeper).Burn(
		fixture.ctx,
		&nfttypes.MsgBurnRequest{
			ClassId: classID,
			NftId:   "nft-2",
			Owner:   firstOwner,
		},
	)
	require.NoError(t, err)
	requireOwnerClassCountMissing(t, &fixture, classID, firstOwner)
	requireOwnerClassCount(t, &fixture, classID, secondOwner, 1)

	_, err = NewMsgServer(fixture.keeper).Burn(
		fixture.ctx,
		&nfttypes.MsgBurnRequest{
			ClassId: classID,
			NftId:   "nft-1",
			Owner:   secondOwner,
		},
	)
	require.NoError(t, err)
	requireOwnerClassCountMissing(t, &fixture, classID, firstOwner)
	requireOwnerClassCountMissing(t, &fixture, classID, secondOwner)
}

func TestOwnerClassCountFailuresRollBackNFTMutations(t *testing.T) {
	t.Run("mint count overflow", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, controller, owner, ownerAddress, _ := createNFTForBurnTest(t, &fixture)
		key := collections.Join(classID, owner)
		require.NoError(t, fixture.keeper.ownerClassCounts.Set(fixture.ctx, key, math.MaxUint64))
		request := validMintRequest(classID, controller, owner)
		request.NftId = "nft-2"

		_, err := NewMsgServer(fixture.keeper).Mint(
			fixture.ctx,
			request,
		)
		require.ErrorContains(t, err, "balance count overflows")
		require.False(t, fixture.keeper.nftKeeper.HasNFT(fixture.ctx, classID, "nft-2"))
		require.Empty(t, fixture.keeper.nftKeeper.GetOwner(fixture.ctx, classID, "nft-2"))
		require.Equal(t, uint64(1), fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID))
		require.Equal(t, uint64(1), fixture.keeper.nftKeeper.GetBalance(
			fixture.ctx,
			classID,
			ownerAddress,
		))
		requireOwnerClassCount(t, &fixture, classID, owner, math.MaxUint64)
		require.Empty(t, fixture.ctx.EventManager().Events())
	})

	t.Run("send missing sender count", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, owner, ownerAddress := createNFTForSendTest(
			t,
			&fixture,
			nfttypes.TransferPolicy_TRANSFER_POLICY_OWNER_TRANSFERABLE,
		)
		receiverAddress := sdk.AccAddress(bytes.Repeat([]byte{94}, 20))
		receiver := fixture.accountAddress(t, receiverAddress)
		fixture.accountKeeper.accounts[string(receiverAddress)] =
			authtypes.NewBaseAccountWithAddress(receiverAddress)
		require.NoError(t, fixture.keeper.ownerClassCounts.Remove(
			fixture.ctx,
			collections.Join(classID, owner),
		))

		_, err := NewStandardMsgServer(fixture.keeper).Send(
			fixture.ctx,
			&upstreamnft.MsgSend{
				ClassId: classID, Id: "nft-1", Sender: owner, Receiver: receiver,
			},
		)
		require.ErrorContains(t, err, "balance count is missing")
		require.Equal(t, ownerAddress, fixture.keeper.nftKeeper.GetOwner(
			fixture.ctx,
			classID,
			"nft-1",
		))
		requireOwnerClassCountMissing(t, &fixture, classID, owner)
		requireOwnerClassCountMissing(t, &fixture, classID, receiver)
		require.Empty(t, fixture.ctx.EventManager().Events())
	})

	t.Run("send receiver count overflow", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, owner, ownerAddress := createNFTForSendTest(
			t,
			&fixture,
			nfttypes.TransferPolicy_TRANSFER_POLICY_OWNER_TRANSFERABLE,
		)
		receiverAddress := sdk.AccAddress(bytes.Repeat([]byte{95}, 20))
		receiver := fixture.accountAddress(t, receiverAddress)
		fixture.accountKeeper.accounts[string(receiverAddress)] =
			authtypes.NewBaseAccountWithAddress(receiverAddress)
		require.NoError(t, fixture.keeper.ownerClassCounts.Set(
			fixture.ctx,
			collections.Join(classID, receiver),
			math.MaxUint64,
		))

		_, err := NewStandardMsgServer(fixture.keeper).Send(
			fixture.ctx,
			&upstreamnft.MsgSend{
				ClassId: classID, Id: "nft-1", Sender: owner, Receiver: receiver,
			},
		)
		require.ErrorContains(t, err, "balance count overflows")
		require.Equal(t, ownerAddress, fixture.keeper.nftKeeper.GetOwner(
			fixture.ctx,
			classID,
			"nft-1",
		))
		requireOwnerClassCount(t, &fixture, classID, owner, 1)
		requireOwnerClassCount(t, &fixture, classID, receiver, math.MaxUint64)
		require.Empty(t, fixture.ctx.EventManager().Events())
	})

	for _, test := range []struct {
		name        string
		corrupt     func(t *testing.T, fixture *keeperFixture, key collections.Pair[string, string])
		errorString string
	}{
		{
			name: "burn missing count",
			corrupt: func(t *testing.T, fixture *keeperFixture, key collections.Pair[string, string]) {
				require.NoError(t, fixture.keeper.ownerClassCounts.Remove(fixture.ctx, key))
			},
			errorString: "balance count is missing",
		},
		{
			name: "burn stored zero count",
			corrupt: func(t *testing.T, fixture *keeperFixture, key collections.Pair[string, string]) {
				require.NoError(t, fixture.keeper.ownerClassCounts.Set(fixture.ctx, key, 0))
			},
			errorString: "stored zero balance count",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := newKeeperFixture(t, true, true)
			classID, _, owner, ownerAddress, _ := createNFTForBurnTest(t, &fixture)
			key := collections.Join(classID, owner)
			test.corrupt(t, &fixture, key)
			before := snapshotRevokeState(t, &fixture, classID, "nft-1")

			_, err := NewMsgServer(fixture.keeper).Burn(
				fixture.ctx,
				&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "nft-1", Owner: owner},
			)
			require.ErrorContains(t, err, test.errorString)
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

func requireOwnerClassCount(
	t *testing.T,
	fixture *keeperFixture,
	classID string,
	owner string,
	expected uint64,
) {
	t.Helper()
	actual, err := fixture.keeper.ownerClassCounts.Get(
		fixture.ctx,
		collections.Join(classID, owner),
	)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func requireOwnerClassCountMissing(
	t *testing.T,
	fixture *keeperFixture,
	classID string,
	owner string,
) {
	t.Helper()
	_, err := fixture.keeper.ownerClassCounts.Get(
		fixture.ctx,
		collections.Join(classID, owner),
	)
	require.ErrorIs(t, err, collections.ErrNotFound)
}
