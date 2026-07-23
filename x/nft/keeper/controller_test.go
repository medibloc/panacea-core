package keeper

import (
	"bytes"
	"strings"
	"testing"

	corestore "cosmossdk.io/core/store"
	upstreamnft "cosmossdk.io/x/nft"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
)

func TestUpdateControllerChangesOnlyController(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	creatorAddress := sdk.AccAddress(bytes.Repeat([]byte{11}, 20))
	newControllerAddress := sdk.AccAddress(bytes.Repeat([]byte{12}, 32))
	classID, creator := createClassForControllerTest(t, &fixture, creatorAddress)
	newController := fixture.accountAddress(t, newControllerAddress)
	fixture.accountKeeper.accounts[string(newControllerAddress)] =
		authtypes.NewBaseAccountWithAddress(newControllerAddress)

	recordBefore, err := fixture.keeper.getClassRecord(fixture.ctx, classID)
	require.NoError(t, err)
	classBefore := *recordBefore.Class
	policyBefore := *recordBefore.Policy
	mintedCountBefore := recordBefore.MintedCount

	response, err := NewMsgServer(fixture.keeper).UpdateController(
		fixture.ctx,
		&types.MsgUpdateControllerRequest{
			ClassId:       classID,
			Controller:    strings.ToUpper(creator),
			NewController: strings.ToUpper(newController),
		},
	)
	require.NoError(t, err)
	require.Equal(t, &types.MsgUpdateControllerResponse{}, response)

	recordAfter, err := fixture.keeper.getClassRecord(fixture.ctx, classID)
	require.NoError(t, err)
	expectedPolicy := policyBefore
	expectedPolicy.Controller = newController
	require.Equal(t, &classBefore, recordAfter.Class)
	require.Equal(t, &expectedPolicy, recordAfter.Policy)
	require.Equal(t, mintedCountBefore, recordAfter.MintedCount)

	queryResponse, err := NewQueryServer(fixture.keeper).ClassRecord(
		fixture.ctx,
		&types.QueryClassRecordRequest{ClassId: classID},
	)
	require.NoError(t, err)
	require.Equal(t, newController, queryResponse.ClassRecord.Policy.Controller)

	events := fixture.ctx.EventManager().ABCIEvents()
	require.Len(t, events, 1)
	parsedEvent, err := sdk.ParseTypedEvent(events[0])
	require.NoError(t, err)
	require.Equal(t, &types.EventControllerUpdated{
		ClassId:       classID,
		OldController: creator,
		NewController: newController,
	}, parsedEvent)
}

func TestUpdateControllerTransfersAuthorityImmediately(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, creator := createClassForControllerTest(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{13}, 20)),
	)
	second := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{14}, 20)))
	third := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{15}, 20)))
	server := NewMsgServer(fixture.keeper)

	_, err := server.UpdateController(
		fixture.ctx,
		&types.MsgUpdateControllerRequest{
			ClassId:       classID,
			Controller:    creator,
			NewController: second,
		},
	)
	require.NoError(t, err)
	fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())

	_, err = server.UpdateController(
		fixture.ctx,
		&types.MsgUpdateControllerRequest{
			ClassId:       classID,
			Controller:    creator,
			NewController: third,
		},
	)
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
	require.Empty(t, fixture.ctx.EventManager().Events())

	_, err = server.UpdateController(
		fixture.ctx,
		&types.MsgUpdateControllerRequest{
			ClassId:       classID,
			Controller:    second,
			NewController: third,
		},
	)
	require.NoError(t, err)

	record, err := fixture.keeper.getClassRecord(fixture.ctx, classID)
	require.NoError(t, err)
	require.Equal(t, third, record.Policy.Controller)
	events := fixture.ctx.EventManager().ABCIEvents()
	require.Len(t, events, 1)
	parsedEvent, err := sdk.ParseTypedEvent(events[0])
	require.NoError(t, err)
	require.Equal(t, &types.EventControllerUpdated{
		ClassId:       classID,
		OldController: second,
		NewController: third,
	}, parsedEvent)
}

func TestUpdateControllerRejectsInvalidRequests(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	creatorAddress := sdk.AccAddress(bytes.Repeat([]byte{16}, 20))
	classID, creator := createClassForControllerTest(t, &fixture, creatorAddress)
	newController := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{17}, 20)))
	arbitraryController := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{18}, 20)))
	moduleController := fixture.accountAddress(t, fixture.accountKeeper.moduleAddress)
	mixedCaseController := strings.ToUpper(creator[:1]) + creator[1:]

	testCases := []struct {
		name      string
		request   *types.MsgUpdateControllerRequest
		targetErr error
	}{
		{name: "nil request", targetErr: sdkerrors.ErrInvalidRequest},
		{
			name: "invalid class ID",
			request: &types.MsgUpdateControllerRequest{
				ClassId:       "invalid",
				Controller:    creator,
				NewController: newController,
			},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "non-canonical class ID",
			request: &types.MsgUpdateControllerRequest{
				ClassId:       strings.ToUpper(creator) + ":certificate",
				Controller:    creator,
				NewController: newController,
			},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "invalid current controller",
			request: &types.MsgUpdateControllerRequest{
				ClassId:       classID,
				Controller:    "not-an-address",
				NewController: newController,
			},
			targetErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "mixed-case current controller",
			request: &types.MsgUpdateControllerRequest{
				ClassId:       classID,
				Controller:    mixedCaseController,
				NewController: newController,
			},
			targetErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid new controller",
			request: &types.MsgUpdateControllerRequest{
				ClassId:       classID,
				Controller:    creator,
				NewController: "not-an-address",
			},
			targetErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "module account as current controller",
			request: &types.MsgUpdateControllerRequest{
				ClassId:       classID,
				Controller:    moduleController,
				NewController: newController,
			},
			targetErr: sdkerrors.ErrUnauthorized,
		},
		{
			name: "module account as new controller",
			request: &types.MsgUpdateControllerRequest{
				ClassId:       classID,
				Controller:    creator,
				NewController: moduleController,
			},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "arbitrary current controller",
			request: &types.MsgUpdateControllerRequest{
				ClassId:       classID,
				Controller:    arbitraryController,
				NewController: newController,
			},
			targetErr: sdkerrors.ErrUnauthorized,
		},
		{
			name: "same canonical controller",
			request: &types.MsgUpdateControllerRequest{
				ClassId:       classID,
				Controller:    creator,
				NewController: strings.ToUpper(creator),
			},
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "missing class",
			request: &types.MsgUpdateControllerRequest{
				ClassId:       creator + ":missing",
				Controller:    creator,
				NewController: newController,
			},
			targetErr: upstreamnft.ErrClassNotExists,
		},
	}

	server := NewMsgServer(fixture.keeper)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := server.UpdateController(fixture.ctx, tc.request)
			require.ErrorIs(t, err, tc.targetErr)
		})
	}

	policy, err := fixture.keeper.classPolicies.Get(fixture.ctx, classID)
	require.NoError(t, err)
	require.Equal(t, creator, policy.Controller)
	require.Empty(t, fixture.ctx.EventManager().Events())
}

func TestUpdateControllerRollsBackWhenPolicyWriteFails(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, creator := createClassForControllerTest(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{19}, 20)),
	)
	newController := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{20}, 20)))
	failingKeeper := NewKeeper(
		fixture.cdc,
		runtime.NewKVStoreService(fixture.nftService),
		failingSetStoreService{
			delegate: corestore.KVStoreService(runtime.NewKVStoreService(fixture.policyService)),
		},
		fixture.accountKeeper,
		testBankKeeper{},
		fixture.moduleAccountAddresses,
	)

	_, err := NewMsgServer(failingKeeper).UpdateController(
		fixture.ctx,
		&types.MsgUpdateControllerRequest{
			ClassId:       classID,
			Controller:    creator,
			NewController: newController,
		},
	)
	require.ErrorContains(t, err, "forced set failure")

	record, err := fixture.keeper.getClassRecord(fixture.ctx, classID)
	require.NoError(t, err)
	require.Equal(t, creator, record.Policy.Controller)
	require.Empty(t, fixture.ctx.EventManager().Events())
}

func createClassForControllerTest(
	t *testing.T,
	fixture *keeperFixture,
	creatorAddress sdk.AccAddress,
) (string, string) {
	t.Helper()
	creator := fixture.accountAddress(t, creatorAddress)
	fixture.accountKeeper.accounts[string(creatorAddress)] =
		authtypes.NewBaseAccountWithAddress(creatorAddress)
	response, err := NewMsgServer(fixture.keeper).CreateClass(
		fixture.ctx,
		validCreateClassRequest(creator),
	)
	require.NoError(t, err)
	fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())
	return response.ClassId, creator
}
