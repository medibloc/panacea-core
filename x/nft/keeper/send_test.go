package keeper

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	upstreamnft "cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
)

func TestStandardMsgSendTransfersActiveOwnerNFT(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, sender, _ := createNFTForSendTest(
		t,
		&fixture,
		nfttypes.TransferPolicy_TRANSFER_POLICY_OWNER_TRANSFERABLE,
	)
	receiverAddress := sdk.AccAddress(bytes.Repeat([]byte{53}, 32))
	receiver := fixture.accountAddress(t, receiverAddress)
	fixture.accountKeeper.accounts[string(receiverAddress)] =
		authtypes.NewBaseAccountWithAddress(receiverAddress)
	tokenBefore, found := fixture.keeper.nftKeeper.GetNFT(fixture.ctx, classID, "nft-1")
	require.True(t, found)
	lifecycleBefore, err := fixture.keeper.lifecycles.Get(
		fixture.ctx,
		collections.Join(classID, "nft-1"),
	)
	require.NoError(t, err)
	classBefore, err := fixture.keeper.getClassRecord(fixture.ctx, classID)
	require.NoError(t, err)
	request := &upstreamnft.MsgSend{
		ClassId:  classID,
		Id:       "nft-1",
		Sender:   strings.ToUpper(sender),
		Receiver: strings.ToUpper(receiver),
	}
	original := *request

	response, err := NewStandardMsgServer(fixture.keeper).Send(
		fixture.ctx,
		request,
	)
	require.NoError(t, err)
	require.Equal(t, &upstreamnft.MsgSendResponse{}, response)
	require.Equal(t, original, *request)
	require.Equal(t, receiverAddress, fixture.keeper.nftKeeper.GetOwner(
		fixture.ctx,
		classID,
		"nft-1",
	))
	require.Equal(t, uint64(1), fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID))

	tokenAfter, found := fixture.keeper.nftKeeper.GetNFT(fixture.ctx, classID, "nft-1")
	require.True(t, found)
	require.Equal(t, tokenBefore, tokenAfter)
	lifecycleAfter, err := fixture.keeper.lifecycles.Get(
		fixture.ctx,
		collections.Join(classID, "nft-1"),
	)
	require.NoError(t, err)
	require.Equal(t, lifecycleBefore, lifecycleAfter)
	classAfter, err := fixture.keeper.getClassRecord(fixture.ctx, classID)
	require.NoError(t, err)
	require.Equal(t, classBefore, classAfter)

	events := fixture.ctx.EventManager().ABCIEvents()
	require.Len(t, events, 1)
	parsedEvent, err := sdk.ParseTypedEvent(events[0])
	require.NoError(t, err)
	require.Equal(t, &upstreamnft.EventSend{
		ClassId:  classID,
		Id:       "nft-1",
		Sender:   sender,
		Receiver: receiver,
	}, parsedEvent)

	queryResponse, err := NewQueryServer(fixture.keeper).NFTRecord(
		fixture.ctx,
		&nfttypes.QueryNFTRecordRequest{ClassId: classID, NftId: "nft-1"},
	)
	require.NoError(t, err)
	require.Equal(t, receiver, queryResponse.NftRecord.GetLive().Owner)
}

func TestStandardMsgSendAllowsSelfTransfer(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, owner, ownerAddress := createNFTForSendTest(
		t,
		&fixture,
		nfttypes.TransferPolicy_TRANSFER_POLICY_OWNER_TRANSFERABLE,
	)

	_, err := NewStandardMsgServer(fixture.keeper).Send(
		fixture.ctx,
		&upstreamnft.MsgSend{
			ClassId:  classID,
			Id:       "nft-1",
			Sender:   owner,
			Receiver: owner,
		},
	)
	require.NoError(t, err)
	require.Equal(t, ownerAddress, fixture.keeper.nftKeeper.GetOwner(
		fixture.ctx,
		classID,
		"nft-1",
	))
	require.Equal(t, uint64(1), fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID))
	require.Len(t, fixture.ctx.EventManager().ABCIEvents(), 1)
}

func TestStandardMsgSendEnforcesAuthorizationAndPolicy(t *testing.T) {
	t.Run("locked class", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, owner, ownerAddress := createNFTForSendTest(
			t,
			&fixture,
			nfttypes.TransferPolicy_TRANSFER_POLICY_LOCKED,
		)
		receiver := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{54}, 20)))

		_, err := NewStandardMsgServer(fixture.keeper).Send(
			fixture.ctx,
			&upstreamnft.MsgSend{ClassId: classID, Id: "nft-1", Sender: owner, Receiver: receiver},
		)
		require.ErrorIs(t, err, nfttypes.ErrTransferNotAllowed)
		assertFailedSendState(t, &fixture, classID, ownerAddress)
	})

	t.Run("revoked NFT", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, owner, ownerAddress := createNFTForSendTest(
			t,
			&fixture,
			nfttypes.TransferPolicy_TRANSFER_POLICY_OWNER_TRANSFERABLE,
		)
		key := collections.Join(classID, "nft-1")
		lifecycle, err := fixture.keeper.lifecycles.Get(fixture.ctx, key)
		require.NoError(t, err)
		lifecycle.Revocation = &nfttypes.Revocation{
			RevokedAt: fixture.ctx.BlockTime().Add(time.Minute),
			RevokedBy: lifecycle.Mint.MintedBy,
		}
		require.NoError(t, fixture.keeper.lifecycles.Set(fixture.ctx, key, lifecycle))
		receiver := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{55}, 20)))

		_, err = NewStandardMsgServer(fixture.keeper).Send(
			fixture.ctx,
			&upstreamnft.MsgSend{ClassId: classID, Id: "nft-1", Sender: owner, Receiver: receiver},
		)
		require.ErrorIs(t, err, nfttypes.ErrNFTRevoked)
		assertFailedSendState(t, &fixture, classID, ownerAddress)
	})

	t.Run("sender is not owner", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, _, ownerAddress := createNFTForSendTest(
			t,
			&fixture,
			nfttypes.TransferPolicy_TRANSFER_POLICY_OWNER_TRANSFERABLE,
		)
		arbitrarySender := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{56}, 20)))
		receiver := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{57}, 20)))

		_, err := NewStandardMsgServer(fixture.keeper).Send(
			fixture.ctx,
			&upstreamnft.MsgSend{
				ClassId: classID, Id: "nft-1", Sender: arbitrarySender, Receiver: receiver,
			},
		)
		require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
		assertFailedSendState(t, &fixture, classID, ownerAddress)
	})

	t.Run("module account receivers", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, owner, ownerAddress := createNFTForSendTest(
			t,
			&fixture,
			nfttypes.TransferPolicy_TRANSFER_POLICY_OWNER_TRANSFERABLE,
		)
		for _, tc := range []struct {
			name    string
			address sdk.AccAddress
		}{
			{name: "materialized", address: fixture.moduleAccountAddresses[0]},
			{name: "unmaterialized", address: fixture.moduleAccountAddresses[1]},
		} {
			t.Run(tc.name, func(t *testing.T) {
				if tc.name == "unmaterialized" {
					require.Nil(t, fixture.accountKeeper.GetAccount(fixture.ctx, tc.address))
				}
				moduleReceiver := fixture.accountAddress(t, tc.address)
				_, err := NewStandardMsgServer(fixture.keeper).Send(
					fixture.ctx,
					&upstreamnft.MsgSend{
						ClassId: classID, Id: "nft-1", Sender: owner, Receiver: moduleReceiver,
					},
				)
				require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
				assertFailedSendState(t, &fixture, classID, ownerAddress)
			})
		}
	})
}

func TestStandardMsgSendReturnsNFTNotExistsForUnusedAndBurnedIDs(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, controller := createClassForMintTest(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{58}, 20)),
		10,
	)
	receiver := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{59}, 20)))
	server := NewStandardMsgServer(fixture.keeper)

	_, err := server.Send(
		fixture.ctx,
		&upstreamnft.MsgSend{
			ClassId: classID, Id: "unused", Sender: controller, Receiver: receiver,
		},
	)
	require.ErrorIs(t, err, upstreamnft.ErrNFTNotExists)

	require.NoError(t, fixture.keeper.tombstones.Set(
		fixture.ctx,
		collections.Join(classID, "burned"),
		nfttypes.BurnTombstone{ClassId: classID, NftId: "burned"},
	))
	_, err = server.Send(
		fixture.ctx,
		&upstreamnft.MsgSend{
			ClassId: classID, Id: "burned", Sender: controller, Receiver: receiver,
		},
	)
	require.ErrorIs(t, err, upstreamnft.ErrNFTNotExists)
	require.Empty(t, fixture.ctx.EventManager().Events())
}

func TestStandardMsgSendRejectsInconsistentNFTState(t *testing.T) {
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newKeeperFixture(t, true, true)
			classID, owner, _ := createNFTForSendTest(
				t,
				&fixture,
				nfttypes.TransferPolicy_TRANSFER_POLICY_OWNER_TRANSFERABLE,
			)
			tc.mutate(t, &fixture, classID)
			receiver := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{60}, 20)))

			_, err := NewStandardMsgServer(fixture.keeper).Send(
				fixture.ctx,
				&upstreamnft.MsgSend{
					ClassId: classID, Id: "nft-1", Sender: owner, Receiver: receiver,
				},
			)
			require.ErrorContains(t, err, tc.expectedError)
			require.Empty(t, fixture.ctx.EventManager().Events())
		})
	}
}

func TestStandardMsgSendRejectsInvalidRequests(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, owner, ownerAddress := createNFTForSendTest(
		t,
		&fixture,
		nfttypes.TransferPolicy_TRANSFER_POLICY_OWNER_TRANSFERABLE,
	)
	receiver := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{61}, 20)))
	creator := strings.Split(classID, ":")[0]
	mixedOwner := strings.ToUpper(owner[:1]) + owner[1:]
	mixedReceiver := strings.ToUpper(receiver[:1]) + receiver[1:]

	for _, tc := range []struct {
		name      string
		request   *upstreamnft.MsgSend
		targetErr error
	}{
		{name: "nil request", targetErr: sdkerrors.ErrInvalidRequest},
		{
			name: "empty class ID",
			request: &upstreamnft.MsgSend{
				Id: "nft-1", Sender: owner, Receiver: receiver,
			},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "invalid class ID",
			request: &upstreamnft.MsgSend{
				ClassId: "invalid", Id: "nft-1", Sender: owner, Receiver: receiver,
			},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "non-canonical class creator",
			request: &upstreamnft.MsgSend{
				ClassId: strings.ToUpper(creator) + ":certificate",
				Id:      "nft-1", Sender: owner, Receiver: receiver,
			},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "malformed class creator",
			request: &upstreamnft.MsgSend{
				ClassId: "badbech32:certificate",
				Id:      "nft-1", Sender: owner, Receiver: receiver,
			},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "empty NFT ID",
			request: &upstreamnft.MsgSend{
				ClassId: classID, Sender: owner, Receiver: receiver,
			},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "invalid NFT ID",
			request: &upstreamnft.MsgSend{
				ClassId: classID, Id: ".", Sender: owner, Receiver: receiver,
			},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "invalid sender",
			request: &upstreamnft.MsgSend{
				ClassId: classID, Id: "nft-1", Sender: "invalid", Receiver: receiver,
			},
			targetErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "mixed-case sender",
			request: &upstreamnft.MsgSend{
				ClassId: classID, Id: "nft-1", Sender: mixedOwner, Receiver: receiver,
			},
			targetErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid receiver",
			request: &upstreamnft.MsgSend{
				ClassId: classID, Id: "nft-1", Sender: owner, Receiver: "invalid",
			},
			targetErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "mixed-case receiver",
			request: &upstreamnft.MsgSend{
				ClassId: classID, Id: "nft-1", Sender: owner, Receiver: mixedReceiver,
			},
			targetErr: sdkerrors.ErrInvalidAddress,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewStandardMsgServer(fixture.keeper).Send(
				fixture.ctx,
				tc.request,
			)
			require.ErrorIs(t, err, tc.targetErr)
			if tc.targetErr == sdkerrors.ErrInvalidRequest {
				codespace, code, _ := errorsmod.ABCIInfo(err, false)
				expectedCodespace, expectedCode, _ := errorsmod.ABCIInfo(
					sdkerrors.ErrInvalidRequest,
					false,
				)
				require.Equal(t, expectedCodespace, codespace)
				require.Equal(t, expectedCode, code)
			}
			assertFailedSendState(t, &fixture, classID, ownerAddress)
		})
	}
}

func createNFTForSendTest(
	t *testing.T,
	fixture *keeperFixture,
	policy nfttypes.TransferPolicy,
) (string, string, sdk.AccAddress) {
	t.Helper()
	fixture.ctx = fixture.ctx.WithBlockTime(time.Date(2026, time.July, 22, 0, 0, 0, 0, time.UTC))
	creatorAddress := sdk.AccAddress(bytes.Repeat([]byte{51}, 20))
	creator := fixture.accountAddress(t, creatorAddress)
	fixture.accountKeeper.accounts[string(creatorAddress)] =
		authtypes.NewBaseAccountWithAddress(creatorAddress)
	classRequest := validCreateClassRequest(creator)
	classRequest.TransferPolicy = policy
	classResponse, err := NewMsgServer(fixture.keeper).CreateClass(
		fixture.ctx,
		classRequest,
	)
	require.NoError(t, err)
	fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())

	ownerAddress := sdk.AccAddress(bytes.Repeat([]byte{52}, 20))
	owner := fixture.accountAddress(t, ownerAddress)
	fixture.accountKeeper.accounts[string(ownerAddress)] =
		authtypes.NewBaseAccountWithAddress(ownerAddress)
	_, err = NewMsgServer(fixture.keeper).Mint(
		fixture.ctx,
		validMintRequest(classResponse.ClassId, creator, owner),
	)
	require.NoError(t, err)
	fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())
	return classResponse.ClassId, owner, ownerAddress
}

func assertFailedSendState(
	t *testing.T,
	fixture *keeperFixture,
	classID string,
	expectedOwner sdk.AccAddress,
) {
	t.Helper()
	require.Equal(t, expectedOwner, fixture.keeper.nftKeeper.GetOwner(
		fixture.ctx,
		classID,
		"nft-1",
	))
	require.Equal(t, uint64(1), fixture.keeper.nftKeeper.GetTotalSupply(fixture.ctx, classID))
	require.Empty(t, fixture.ctx.EventManager().Events())
}
