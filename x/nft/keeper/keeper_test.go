package keeper

import (
	"context"
	"reflect"
	"testing"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	upstreamnft "cosmossdk.io/x/nft"
	upstreamkeeper "cosmossdk.io/x/nft/keeper"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
)

type testAccountKeeper struct {
	addressCodec  address.Codec
	moduleAddress sdk.AccAddress
	requested     []string
}

func (k *testAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	k.requested = append(k.requested, name)
	if name == upstreamnft.ModuleName {
		return k.moduleAddress
	}
	return nil
}

func (*testAccountKeeper) GetAccount(context.Context, sdk.AccAddress) sdk.AccountI { return nil }

func (k *testAccountKeeper) AddressCodec() address.Codec { return k.addressCodec }

type testBankKeeper struct{}

func (testBankKeeper) SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins { return nil }

type keeperFixture struct {
	keeper        Keeper
	ctx           sdk.Context
	cdc           *codec.ProtoCodec
	nftService    *storetypes.KVStoreKey
	policyService *storetypes.KVStoreKey
	accountKeeper *testAccountKeeper
}

func TestNewKeeperOwnsTypedPolicyCollections(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)

	require.Equal(t, []string{upstreamnft.ModuleName}, fixture.accountKeeper.requested)

	collectionsByName := fixture.keeper.schema.ListCollections()
	names := make([]string, 0, len(collectionsByName))
	for _, collection := range collectionsByName {
		names = append(names, collection.GetName())
	}
	require.Equal(t, []string{"class_policies", "lifecycles", "minted_counts", "tombstones"}, names)
}

func TestNFTPairKeyCodecDoesNotConcatenateAmbiguousIDs(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	firstKey := collections.Join("a", "bc")
	secondKey := collections.Join("ab", "c")

	require.NoError(t, fixture.keeper.lifecycles.Set(
		fixture.ctx,
		firstKey,
		types.LifecycleRecord{ClassId: "a", NftId: "bc"},
	))
	require.NoError(t, fixture.keeper.lifecycles.Set(
		fixture.ctx,
		secondKey,
		types.LifecycleRecord{ClassId: "ab", NftId: "c"},
	))

	first, err := fixture.keeper.lifecycles.Get(fixture.ctx, firstKey)
	require.NoError(t, err)
	second, err := fixture.keeper.lifecycles.Get(fixture.ctx, secondKey)
	require.NoError(t, err)
	require.Equal(t, "a", first.ClassId)
	require.Equal(t, "ab", second.ClassId)
}

func TestUpstreamKeeperRemainsPrivate(t *testing.T) {
	keeperType := reflect.TypeOf(Keeper{})
	field, ok := keeperType.FieldByName("nftKeeper")
	require.True(t, ok)
	require.Equal(t, reflect.TypeOf(upstreamkeeper.Keeper{}), field.Type)
	require.NotEmpty(t, field.PkgPath, "nftKeeper must remain unexported")

	upstreamType := reflect.TypeOf(upstreamkeeper.Keeper{})
	upstreamPointerType := reflect.PointerTo(upstreamType)
	for i := 0; i < keeperType.NumMethod(); i++ {
		method := keeperType.Method(i)
		for output := 0; output < method.Type.NumOut(); output++ {
			require.NotEqual(t, upstreamType, method.Type.Out(output))
			require.NotEqual(t, upstreamPointerType, method.Type.Out(output))
		}
	}
}

func TestNewKeeperRequiresBothStoresAndModuleAccount(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	accountKeeper := &testAccountKeeper{addressCodec: addresscodec.NewBech32Codec("panacea")}
	bankKeeper := testBankKeeper{}
	nftService := runtime.NewKVStoreService(fixture.nftService)
	policyService := runtime.NewKVStoreService(fixture.policyService)

	require.PanicsWithValue(t, "nft keeper requires the nft store service", func() {
		NewKeeper(fixture.cdc, nil, policyService, fixture.accountKeeper, bankKeeper)
	})
	require.PanicsWithValue(t, "nft keeper requires the nftpolicy store service", func() {
		NewKeeper(fixture.cdc, nftService, nil, fixture.accountKeeper, bankKeeper)
	})
	require.PanicsWithValue(t, "the nft module account has not been set", func() {
		NewKeeper(fixture.cdc, nftService, policyService, accountKeeper, bankKeeper)
	})
}

func TestKeeperContextRequiresBothMountedStores(t *testing.T) {
	for _, tc := range []struct {
		name        string
		mountNFT    bool
		mountPolicy bool
	}{
		{name: "missing nft store", mountPolicy: true},
		{name: "missing nftpolicy store", mountNFT: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newKeeperFixture(t, tc.mountNFT, tc.mountPolicy)

			require.Panics(t, func() {
				_ = fixture.keeper.InitGenesis(fixture.ctx, types.DefaultGenesis())
			})
		})
	}
}

func TestEmptyGenesisRoundTrip(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)

	require.NoError(t, fixture.keeper.InitGenesis(fixture.ctx, types.DefaultGenesis()))
	exported, err := fixture.keeper.ExportGenesis(fixture.ctx)
	require.NoError(t, err)
	require.NotNil(t, exported.NftState)
	require.Empty(t, exported.NftState.Classes)
	require.Empty(t, exported.NftState.Entries)
	require.NotNil(t, exported.ClassPolicies)
	require.NotNil(t, exported.Lifecycles)
	require.NotNil(t, exported.Tombstones)
}

func newKeeperFixture(t *testing.T, mountNFT, mountPolicy bool) keeperFixture {
	t.Helper()

	registry := cdctypes.NewInterfaceRegistry()
	sdk.RegisterInterfaces(registry)
	types.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	nftKey := storetypes.NewKVStoreKey(types.StoreKey)
	policyKey := storetypes.NewKVStoreKey(types.PolicyStoreKey)
	database := dbm.NewMemDB()
	multiStore := store.NewCommitMultiStore(database, log.NewNopLogger(), metrics.NewNoOpMetrics())
	if mountNFT {
		multiStore.MountStoreWithDB(nftKey, storetypes.StoreTypeIAVL, database)
	}
	if mountPolicy {
		multiStore.MountStoreWithDB(policyKey, storetypes.StoreTypeIAVL, database)
	}
	require.NoError(t, multiStore.LoadLatestVersion())

	accountKeeper := &testAccountKeeper{
		addressCodec:  addresscodec.NewBech32Codec("panacea"),
		moduleAddress: authtypes.NewModuleAddress(upstreamnft.ModuleName),
	}
	nftService := runtime.NewKVStoreService(nftKey)
	policyService := runtime.NewKVStoreService(policyKey)
	k := NewKeeper(cdc, nftService, policyService, accountKeeper, testBankKeeper{})

	return keeperFixture{
		keeper:        k,
		ctx:           sdk.NewContext(multiStore, cmtproto.Header{}, false, log.NewNopLogger()),
		cdc:           cdc,
		nftService:    nftKey,
		policyService: policyKey,
		accountKeeper: accountKeeper,
	}
}
