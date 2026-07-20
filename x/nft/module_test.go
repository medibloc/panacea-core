package nft

import (
	"context"
	"testing"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	upstreamnft "cosmossdk.io/x/nft"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	grpc "github.com/cosmos/gogoproto/grpc"
	"github.com/medibloc/panacea-core/v2/x/nft/keeper"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
	googlegrpc "google.golang.org/grpc"
)

func TestAppModuleBasicRegistersAllNFTInterfaces(t *testing.T) {
	registry := cdctypes.NewInterfaceRegistry()
	sdk.RegisterInterfaces(registry)
	NewAppModuleBasic(addresscodec.NewBech32Codec("panacea")).RegisterInterfaces(registry)

	for _, message := range []sdk.Msg{
		&upstreamnft.MsgSend{},
		&types.MsgCreateClassRequest{},
		&types.MsgUpdateControllerRequest{},
		&types.MsgMintRequest{},
		&types.MsgRevokeRequest{},
		&types.MsgBurnRequest{},
	} {
		packed, err := cdctypes.NewAnyWithValue(message)
		require.NoError(t, err)

		var unpacked sdk.Msg
		require.NoError(t, registry.UnpackAny(packed, &unpacked))
		require.IsType(t, message, unpacked)
	}

	require.Equal(
		t,
		[]string{types.BasicNFTDataTypeURL},
		registry.ListImplementations(types.NFTDataInterfaceName),
	)
}

func TestAppModuleBasicGenesisContract(t *testing.T) {
	addressCodec := addresscodec.NewBech32Codec("panacea")
	basic := NewAppModuleBasic(addressCodec)
	cdc := newModuleTestCodec()

	defaultGenesis := basic.DefaultGenesis(cdc)
	var decoded types.GenesisState
	require.NoError(t, cdc.UnmarshalJSON(defaultGenesis, &decoded))
	require.NotNil(t, decoded.NftState)
	require.NoError(t, basic.ValidateGenesis(cdc, nil, defaultGenesis))

	nilNFTState := cdc.MustMarshalJSON(&types.GenesisState{})
	require.ErrorContains(t, basic.ValidateGenesis(cdc, nil, nilNFTState), "nft_state must not be nil")
}

func TestAppModuleEmptyGenesisRoundTrip(t *testing.T) {
	moduleKeeper, sdkContext, addressCodec, cdc := newModuleTestKeeper(t)
	appModule := NewAppModule(addressCodec, moduleKeeper)
	defaultGenesis := appModule.DefaultGenesis(cdc)

	updates := appModule.InitGenesis(sdkContext, cdc, defaultGenesis)
	require.Empty(t, updates)
	firstExport := appModule.ExportGenesis(sdkContext, cdc)
	secondExport := appModule.ExportGenesis(sdkContext, cdc)
	require.JSONEq(t, string(defaultGenesis), string(firstExport))
	require.Equal(t, firstExport, secondExport)
	require.Equal(t, uint64(1), appModule.ConsensusVersion())
}

func TestAppModuleDoesNotRegisterRuntimeServicesYet(t *testing.T) {
	appModule := NewAppModule(addresscodec.NewBech32Codec("panacea"), keeper.Keeper{})
	configurator := &countingConfigurator{}

	appModule.RegisterServices(configurator)
	appModule.RegisterServices(configurator)

	require.Zero(t, configurator.services)
	require.Zero(t, configurator.migrations)
}

type moduleTestAccountKeeper struct {
	addressCodec address.Codec
}

func (k moduleTestAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	if name == upstreamnft.ModuleName {
		return authtypes.NewModuleAddress(name)
	}
	return nil
}

func (moduleTestAccountKeeper) GetAccount(context.Context, sdk.AccAddress) sdk.AccountI { return nil }

func (k moduleTestAccountKeeper) AddressCodec() address.Codec { return k.addressCodec }

type moduleTestBankKeeper struct{}

func (moduleTestBankKeeper) SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins { return nil }

func newModuleTestCodec() *codec.ProtoCodec {
	registry := cdctypes.NewInterfaceRegistry()
	sdk.RegisterInterfaces(registry)
	types.RegisterInterfaces(registry)
	return codec.NewProtoCodec(registry)
}

func newModuleTestKeeper(t *testing.T) (keeper.Keeper, sdk.Context, address.Codec, *codec.ProtoCodec) {
	t.Helper()

	cdc := newModuleTestCodec()
	addressCodec := addresscodec.NewBech32Codec("panacea")
	nftKey := storetypes.NewKVStoreKey(types.StoreKey)
	policyKey := storetypes.NewKVStoreKey(types.PolicyStoreKey)
	database := dbm.NewMemDB()
	multiStore := store.NewCommitMultiStore(database, log.NewNopLogger(), metrics.NewNoOpMetrics())
	multiStore.MountStoreWithDB(nftKey, storetypes.StoreTypeIAVL, database)
	multiStore.MountStoreWithDB(policyKey, storetypes.StoreTypeIAVL, database)
	require.NoError(t, multiStore.LoadLatestVersion())

	moduleKeeper := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(nftKey),
		runtime.NewKVStoreService(policyKey),
		moduleTestAccountKeeper{addressCodec: addressCodec},
		moduleTestBankKeeper{},
	)
	ctx := sdk.NewContext(multiStore, cmtproto.Header{}, false, log.NewNopLogger())
	return moduleKeeper, ctx, addressCodec, cdc
}

type countingConfigurator struct {
	services   int
	migrations int
}

func (c *countingConfigurator) RegisterService(*googlegrpc.ServiceDesc, interface{}) {
	c.services++
}

func (*countingConfigurator) Error() error { return nil }

func (c *countingConfigurator) MsgServer() grpc.Server { return c }

func (c *countingConfigurator) QueryServer() grpc.Server { return c }

func (c *countingConfigurator) RegisterMigration(string, uint64, module.MigrationHandler) error {
	c.migrations++
	return nil
}
