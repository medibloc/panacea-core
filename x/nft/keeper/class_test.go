package keeper

import (
	"bytes"
	"context"
	"errors"
	"math"
	"strings"
	"testing"

	corestore "cosmossdk.io/core/store"
	upstreamnft "cosmossdk.io/x/nft"
	upstreamkeeper "cosmossdk.io/x/nft/keeper"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateClassAndQueryClassRecord(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	creatorAddress := sdk.AccAddress(bytes.Repeat([]byte{1}, 20))
	creator := fixture.accountAddress(t, creatorAddress)
	fixture.accountKeeper.accounts[string(creatorAddress)] = authtypes.NewBaseAccountWithAddress(creatorAddress)
	request := validCreateClassRequest(strings.ToUpper(creator))
	request.MaxSupply = math.MaxUint64

	response, err := NewMsgServer(fixture.keeper).CreateClass(
		sdk.WrapSDKContext(fixture.ctx),
		request,
	)
	require.NoError(t, err)
	expectedClassID := creator + ":certificate"
	require.Equal(t, expectedClassID, response.ClassId)

	class, hasClass := fixture.keeper.nftKeeper.GetClass(fixture.ctx, expectedClassID)
	require.True(t, hasClass)
	require.Equal(t, expectedClassID, class.Id)
	require.Equal(t, request.Name, class.Name)
	require.Equal(t, request.Symbol, class.Symbol)
	require.Equal(t, request.Description, class.Description)
	require.Equal(t, request.Uri, class.Uri)
	require.Equal(t, request.UriHash, class.UriHash)
	require.Nil(t, class.Data)

	policy, err := fixture.keeper.classPolicies.Get(fixture.ctx, expectedClassID)
	require.NoError(t, err)
	require.Equal(t, expectedClassID, policy.ClassId)
	require.Equal(t, creator, policy.Creator)
	require.Equal(t, creator, policy.Controller)
	require.Equal(t, request.TransferPolicy, policy.TransferPolicy)
	require.Equal(t, request.Revocable, policy.Revocable)
	require.Equal(t, uint64(math.MaxUint64), policy.MaxSupply)
	mintedCount, err := fixture.keeper.mintedCounts.Get(fixture.ctx, expectedClassID)
	require.NoError(t, err)
	require.Zero(t, mintedCount)

	queryResponse, err := NewQueryServer(fixture.keeper).ClassRecord(
		sdk.WrapSDKContext(fixture.ctx),
		&types.QueryClassRecordRequest{ClassId: expectedClassID},
	)
	require.NoError(t, err)
	require.Equal(t, &class, queryResponse.ClassRecord.Class)
	require.Equal(t, &policy, queryResponse.ClassRecord.Policy)
	require.Zero(t, queryResponse.ClassRecord.MintedCount)

	events := fixture.ctx.EventManager().ABCIEvents()
	require.Len(t, events, 1)
	parsedEvent, err := sdk.ParseTypedEvent(events[0])
	require.NoError(t, err)
	require.Equal(t, &types.EventClassCreated{
		ClassId: expectedClassID,
		Creator: creator,
	}, parsedEvent)
}

func TestCreateClassAcceptsNonModuleAccountShapes(t *testing.T) {
	testCases := []struct {
		name          string
		addressLength int
		localClassID  string
		maxSupply     uint64
		fullIDLength  int
	}{
		{
			name:          "regular or multisig account with unlimited supply",
			addressLength: 20,
			localClassID:  "class",
			maxSupply:     0,
		},
		{
			name:          "group policy account",
			addressLength: 32,
			localClassID:  strings.Repeat("a", 64),
			fullIDLength:  131,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newKeeperFixture(t, true, true)
			address := sdk.AccAddress(bytes.Repeat([]byte{2}, tc.addressLength))
			creator := fixture.accountAddress(t, address)
			fixture.accountKeeper.accounts[string(address)] = authtypes.NewBaseAccountWithAddress(address)
			request := validCreateClassRequest(creator)
			request.LocalClassId = tc.localClassID
			request.MaxSupply = tc.maxSupply

			response, err := NewMsgServer(fixture.keeper).CreateClass(
				sdk.WrapSDKContext(fixture.ctx),
				request,
			)
			require.NoError(t, err)
			if tc.fullIDLength != 0 {
				require.Len(t, response.ClassId, tc.fullIDLength)
			}
			policy, err := fixture.keeper.classPolicies.Get(fixture.ctx, response.ClassId)
			require.NoError(t, err)
			require.Equal(t, tc.maxSupply, policy.MaxSupply)
		})
	}
}

func TestCreateClassRejectsInvalidRequests(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	creatorAddress := sdk.AccAddress(bytes.Repeat([]byte{3}, 20))
	creator := fixture.accountAddress(t, creatorAddress)
	fixture.accountKeeper.accounts[string(creatorAddress)] = authtypes.NewBaseAccountWithAddress(creatorAddress)
	moduleCreator := fixture.accountAddress(t, fixture.accountKeeper.moduleAddress)
	mixedCaseCreator := strings.ToUpper(creator[:1]) + creator[1:]

	testCases := []struct {
		name      string
		request   *types.MsgCreateClassRequest
		targetErr error
	}{
		{name: "nil request", targetErr: sdkerrors.ErrInvalidRequest},
		{
			name: "invalid local class ID",
			request: func() *types.MsgCreateClassRequest {
				request := validCreateClassRequest(creator)
				request.LocalClassId = "UPPERCASE"
				return request
			}(),
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name:      "invalid creator",
			request:   validCreateClassRequest("not-an-address"),
			targetErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name:      "mixed case creator",
			request:   validCreateClassRequest(mixedCaseCreator),
			targetErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name:      "module account creator",
			request:   validCreateClassRequest(moduleCreator),
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "unspecified transfer policy",
			request: func() *types.MsgCreateClassRequest {
				request := validCreateClassRequest(creator)
				request.TransferPolicy = types.TransferPolicy_TRANSFER_POLICY_UNSPECIFIED
				return request
			}(),
			targetErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "invalid URI hash",
			request: func() *types.MsgCreateClassRequest {
				request := validCreateClassRequest(creator)
				request.UriHash = "sha256:bad"
				return request
			}(),
			targetErr: sdkerrors.ErrInvalidRequest,
		},
	}

	server := NewMsgServer(fixture.keeper)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := server.CreateClass(sdk.WrapSDKContext(fixture.ctx), tc.request)
			require.ErrorIs(t, err, tc.targetErr)
		})
	}
	require.Empty(t, fixture.keeper.nftKeeper.GetClasses(fixture.ctx))
	require.Empty(t, fixture.ctx.EventManager().Events())
}

func TestCreateClassRejectsDuplicate(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{4}, 20)))
	request := validCreateClassRequest(creator)
	server := NewMsgServer(fixture.keeper)

	_, err := server.CreateClass(sdk.WrapSDKContext(fixture.ctx), request)
	require.NoError(t, err)
	_, err = server.CreateClass(sdk.WrapSDKContext(fixture.ctx), request)
	require.ErrorIs(t, err, upstreamnft.ErrClassExists)
	require.Len(t, fixture.keeper.nftKeeper.GetClasses(fixture.ctx), 1)
	require.Len(t, fixture.ctx.EventManager().Events(), 1)
}

func TestCreateClassWritesBothStoresAtomically(t *testing.T) {
	for _, failStore := range []string{"nft", "policy_nft"} {
		t.Run(failStore, func(t *testing.T) {
			fixture := newKeeperFixture(t, true, true)
			creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{5}, 20)))
			nftService := corestore.KVStoreService(runtime.NewKVStoreService(fixture.nftService))
			policyService := corestore.KVStoreService(runtime.NewKVStoreService(fixture.policyService))
			if failStore == "nft" {
				nftService = failingSetStoreService{delegate: nftService}
			} else {
				policyService = failingSetStoreService{delegate: policyService}
			}
			failingKeeper := NewKeeper(
				fixture.cdc,
				nftService,
				policyService,
				fixture.accountKeeper,
				testBankKeeper{},
				fixture.moduleAccountAddresses,
			)

			_, err := NewMsgServer(failingKeeper).CreateClass(
				sdk.WrapSDKContext(fixture.ctx),
				validCreateClassRequest(creator),
			)
			require.ErrorContains(t, err, "forced set failure")

			classID := creator + ":certificate"
			require.False(t, fixture.keeper.nftKeeper.HasClass(fixture.ctx, classID))
			policyExists, err := fixture.keeper.classPolicies.Has(fixture.ctx, classID)
			require.NoError(t, err)
			require.False(t, policyExists)
			mintedCountExists, err := fixture.keeper.mintedCounts.Has(fixture.ctx, classID)
			require.NoError(t, err)
			require.False(t, mintedCountExists)
			require.Empty(t, fixture.ctx.EventManager().Events())
		})
	}
}

func TestQueryClassRecordErrorMapping(t *testing.T) {
	t.Run("invalid argument", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		server := NewQueryServer(fixture.keeper)
		creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{8}, 20)))

		_, err := server.ClassRecord(sdk.WrapSDKContext(fixture.ctx), nil)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		_, err = server.ClassRecord(
			sdk.WrapSDKContext(fixture.ctx),
			&types.QueryClassRecordRequest{ClassId: "invalid"},
		)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		_, err = server.ClassRecord(
			sdk.WrapSDKContext(fixture.ctx),
			&types.QueryClassRecordRequest{ClassId: strings.ToUpper(creator) + ":class"},
		)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("not found", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{6}, 20)))
		_, err := NewQueryServer(fixture.keeper).ClassRecord(
			sdk.WrapSDKContext(fixture.ctx),
			&types.QueryClassRecordRequest{ClassId: creator + ":missing"},
		)
		require.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("inconsistent state", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{7}, 20)))
		classID := creator + ":orphan"
		require.NoError(t, fixture.keeper.nftKeeper.SaveClass(fixture.ctx, upstreamnft.Class{Id: classID}))

		_, err := NewQueryServer(fixture.keeper).ClassRecord(
			sdk.WrapSDKContext(fixture.ctx),
			&types.QueryClassRecordRequest{ClassId: classID},
		)
		require.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestGetClassRecordRejectsMismatchedStandardClassID(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{9}, 20)))
	response, err := NewMsgServer(fixture.keeper).CreateClass(
		sdk.WrapSDKContext(fixture.ctx),
		validCreateClassRequest(creator),
	)
	require.NoError(t, err)

	class, found := fixture.keeper.nftKeeper.GetClass(fixture.ctx, response.ClassId)
	require.True(t, found)
	class.Id = creator + ":different"
	classBytes, err := fixture.cdc.Marshal(&class)
	require.NoError(t, err)
	classKey := append(append([]byte(nil), upstreamkeeper.ClassKey...), response.ClassId...)
	require.NoError(t, fixture.keeper.nftStoreService.OpenKVStore(fixture.ctx).Set(
		classKey,
		classBytes,
	))

	record, err := fixture.keeper.getClassRecord(fixture.ctx, response.ClassId)
	require.Nil(t, record)
	require.ErrorContains(t, err, "standard class key does not match value")
}

func validCreateClassRequest(creator string) *types.MsgCreateClassRequest {
	return &types.MsgCreateClassRequest{
		Creator:        creator,
		LocalClassId:   "certificate",
		Name:           "Certificate",
		Symbol:         "CERT",
		Description:    "Completion certificate",
		Uri:            "https://example.com/class.json",
		UriHash:        "sha256:" + strings.Repeat("a", 64),
		TransferPolicy: types.TransferPolicy_TRANSFER_POLICY_OWNER_TRANSFERABLE,
		Revocable:      true,
		MaxSupply:      100,
	}
}

func (f keeperFixture) accountAddress(t *testing.T, address sdk.AccAddress) string {
	t.Helper()
	encoded, err := f.accountKeeper.addressCodec.BytesToString(address)
	require.NoError(t, err)
	return encoded
}

type failingSetStoreService struct {
	delegate corestore.KVStoreService
}

func (s failingSetStoreService) OpenKVStore(ctx context.Context) corestore.KVStore {
	return failingSetStore{KVStore: s.delegate.OpenKVStore(ctx)}
}

type failingSetStore struct {
	corestore.KVStore
}

func (failingSetStore) Set([]byte, []byte) error {
	return errors.New("forced set failure")
}
