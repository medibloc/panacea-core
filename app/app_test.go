package app_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"cosmossdk.io/core/header"
	"cosmossdk.io/log"
	"cosmossdk.io/x/feegrant"
	"cosmossdk.io/x/nft"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	panaceaapp "github.com/medibloc/panacea-core/v2/app"
	pnftlegacy "github.com/medibloc/panacea-core/v2/x/pnft/legacy"
	pnfttypes "github.com/medibloc/panacea-core/v2/x/pnft/types"
)

func TestRuntimeModulesHaveBootstrapBasics(t *testing.T) {
	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)

	runtimeBasics := testApp.BasicManager()
	moduleNames := make([]string, 0, len(testApp.ModuleManager.Modules))
	for name := range testApp.ModuleManager.Modules {
		moduleNames = append(moduleNames, name)
		require.Contains(t, runtimeBasics, name, "runtime module %q has no derived basic", name)
		require.Contains(t, panaceaapp.ModuleBasics, name, "runtime module %q is missing from bootstrap basics", name)
	}

	runtimeBasicNames := make([]string, 0, len(runtimeBasics))
	for name := range runtimeBasics {
		runtimeBasicNames = append(runtimeBasicNames, name)
	}
	require.ElementsMatch(t, moduleNames, runtimeBasicNames)

	var bootstrapOnly []string
	for name := range panaceaapp.ModuleBasics {
		if _, ok := runtimeBasics[name]; !ok {
			bootstrapOnly = append(bootstrapOnly, name)
		}
	}
	require.ElementsMatch(t, []string{ibctm.ModuleName}, bootstrapOnly)
}

func TestCapabilityWiring(t *testing.T) {
	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)

	require.True(t, testApp.CapabilityKeeper.HasModule(ibcexported.ModuleName))
	require.True(t, testApp.CapabilityKeeper.HasModule(ibctransfertypes.ModuleName))
	require.True(t, testApp.CapabilityKeeper.IsSealed())
	require.Contains(t, testApp.ModuleManager.Modules, capabilitytypes.ModuleName)

	require.NotEmpty(t, testApp.ModuleManager.OrderBeginBlockers)
	require.Equal(t, capabilitytypes.ModuleName, testApp.ModuleManager.OrderBeginBlockers[0])
	require.NotEmpty(t, testApp.ModuleManager.OrderInitGenesis)
	require.Equal(t, capabilitytypes.ModuleName, testApp.ModuleManager.OrderInitGenesis[0])
	require.NotEmpty(t, testApp.ModuleManager.OrderExportGenesis)
	require.Equal(t, capabilitytypes.ModuleName, testApp.ModuleManager.OrderExportGenesis[0])
}

func TestPNFTCompatibilityModuleWiring(t *testing.T) {
	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)

	require.NotContains(t, testApp.GetKVStoreKey(), pnfttypes.StoreKey)
	require.NotContains(t, testApp.GetKVStoreKey(), nft.ModuleName)
	require.Contains(t, testApp.ModuleManager.Modules, pnfttypes.ModuleName)
	require.NotContains(t, testApp.ModuleManager.Modules, nft.ModuleName)
	require.NotContains(t, testApp.ModuleManager.OrderInitGenesis, pnfttypes.ModuleName)
	require.NotContains(t, testApp.ModuleManager.OrderExportGenesis, pnfttypes.ModuleName)
}

func TestPNFTMsgRouteUsesLegacyRejectionServer(t *testing.T) {
	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)
	require.NoError(t, testApp.LoadLatestVersion())

	msg := &pnfttypes.MsgMintPNFTRequest{
		DenomId: "legacy-denom",
		Id:      "legacy-pnft",
		Name:    "Legacy PNFT",
		Creator: sdk.AccAddress(bytes.Repeat([]byte{5}, 20)).String(),
	}
	handler := testApp.MsgServiceRouter().Handler(msg)
	require.NotNil(t, handler)

	ctx := testApp.NewUncachedContext(false, cmtproto.Header{Time: time.Now()})
	response, err := handler(ctx, msg)
	require.Nil(t, response)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.ErrorContains(t, err, pnftlegacy.DisabledErrorMessage)
}

func TestPNFTQueryRoutesAreNotRegistered(t *testing.T) {
	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)

	for _, path := range []string{
		"/panacea.pnft.v2.Query/Denoms",
		"/panacea.pnft.v2.Query/DenomsByOwner",
		"/panacea.pnft.v2.Query/Denom",
		"/panacea.pnft.v2.Query/PNFTs",
		"/panacea.pnft.v2.Query/PNFTsByDenomOwner",
		"/panacea.pnft.v2.Query/PNFT",
	} {
		t.Run(path, func(t *testing.T) {
			require.Nil(t, testApp.GRPCQueryRouter().Route(path))
		})
	}
}

func TestFeeGrantWiringWithStandaloneModule(t *testing.T) {
	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)
	require.NoError(t, testApp.LoadLatestVersion())

	require.Contains(t, testApp.GetKVStoreKey(), feegrant.StoreKey)
	require.Contains(t, testApp.ModuleManager.Modules, feegrant.ModuleName)

	ctx := testApp.NewUncachedContext(false, cmtproto.Header{Time: time.Now()})
	granter := sdk.AccAddress(bytes.Repeat([]byte{3}, 20))
	grantee := sdk.AccAddress(bytes.Repeat([]byte{4}, 20))
	spendLimit := sdk.NewCoins(sdk.NewInt64Coin("umed", 100))

	msg, err := feegrant.NewMsgGrantAllowance(
		&feegrant.BasicAllowance{SpendLimit: spendLimit},
		granter,
		grantee,
	)
	require.NoError(t, err)

	handler := testApp.MsgServiceRouter().Handler(msg)
	require.NotNil(t, handler)
	_, err = handler(ctx, msg)
	require.NoError(t, err)

	allowance, err := testApp.FeeGrantKeeper.GetAllowance(ctx, granter, grantee)
	require.NoError(t, err)
	basicAllowance, ok := allowance.(*feegrant.BasicAllowance)
	require.True(t, ok)
	require.Equal(t, spendLimit, basicAllowance.SpendLimit)
}

func TestUpgradeWiringWithStandaloneModule(t *testing.T) {
	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)
	require.NoError(t, testApp.LoadLatestVersion())

	require.Contains(t, testApp.GetKVStoreKey(), upgradetypes.StoreKey)
	require.Contains(t, testApp.ModuleManager.Modules, upgradetypes.ModuleName)
	require.Equal(t, []string{upgradetypes.ModuleName}, testApp.ModuleManager.OrderPreBlockers)
	require.Same(t, testApp.BaseApp, testApp.UpgradeKeeper.GetVersionSetter())

	const (
		upgradeName   = "test-standalone-upgrade"
		upgradeHeight = int64(10)
	)
	blockTime := time.Now()
	ctx := testApp.NewUncachedContext(false, cmtproto.Header{Height: upgradeHeight, Time: blockTime}).
		WithHeaderInfo(header.Info{Height: upgradeHeight, Time: blockTime})
	require.NoError(t, testApp.UpgradeKeeper.SetModuleVersionMap(ctx, testApp.ModuleManager.GetVersionMap()))

	handlerCalled := false
	testApp.UpgradeKeeper.SetUpgradeHandler(
		upgradeName,
		func(_ context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			handlerCalled = true
			return fromVM, nil
		},
	)

	plan := upgradetypes.Plan{Name: upgradeName, Height: upgradeHeight}
	require.NoError(t, testApp.UpgradeKeeper.ScheduleUpgrade(ctx, plan))
	storedPlan, err := testApp.UpgradeKeeper.GetUpgradePlan(ctx)
	require.NoError(t, err)
	require.Equal(t, plan, storedPlan)

	response, err := testApp.PreBlocker(ctx, &abci.RequestFinalizeBlock{})
	require.NoError(t, err)
	require.True(t, response.ConsensusParamsChanged)
	require.True(t, handlerCalled)

	doneHeight, err := testApp.UpgradeKeeper.GetDoneHeight(ctx, upgradeName)
	require.NoError(t, err)
	require.Equal(t, upgradeHeight, doneHeight)
	_, err = testApp.UpgradeKeeper.GetUpgradePlan(ctx)
	require.ErrorIs(t, err, upgradetypes.ErrNoUpgradePlanFound)
}
