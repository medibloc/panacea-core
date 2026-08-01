package app_test

import (
	"testing"
	"time"

	"cosmossdk.io/core/header"
	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	panaceaapp "github.com/medibloc/panacea-core/v2/app"
	"github.com/medibloc/panacea-core/v2/app/upgrades/v2_3_0"
)

func TestV230UpgradeMigratesIBCCoreParams(t *testing.T) {
	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)
	require.NoError(t, testApp.LoadLatestVersion())

	const upgradeHeight = int64(10)
	blockTime := time.Now().UTC()
	ctx := testApp.NewUncachedContext(
		false,
		cmtproto.Header{Height: upgradeHeight, Time: blockTime},
	).WithHeaderInfo(header.Info{Height: upgradeHeight, Time: blockTime})

	fromVM := testApp.ModuleManager.GetVersionMap()
	fromVM[ibcexported.ModuleName] = 4
	require.Equal(t, uint64(5), fromVM[ibctransfertypes.ModuleName])
	require.NoError(t, testApp.UpgradeKeeper.SetModuleVersionMap(ctx, fromVM))

	expectedClientParams := ibcclienttypes.NewParams("06-solomachine", "07-tendermint")
	expectedConnectionParams := ibcconnectiontypes.NewParams(uint64(45 * time.Second))
	legacyIBCSubspace := testApp.GetSubspace(ibcexported.ModuleName)
	legacyIBCSubspace.SetParamSet(ctx, &expectedClientParams)
	legacyIBCSubspace.SetParamSet(ctx, &expectedConnectionParams)

	ibcStore := ctx.KVStore(testApp.GetKey(ibcexported.StoreKey))
	require.Nil(t, ibcStore.Get([]byte(ibcclienttypes.ParamsKey)))
	require.Nil(t, ibcStore.Get([]byte(ibcconnectiontypes.ParamsKey)))
	require.Nil(t, ibcStore.Get([]byte(ibcchanneltypes.ParamsKey)))

	require.NoError(t, testApp.UpgradeKeeper.ScheduleUpgrade(ctx, upgradetypes.Plan{
		Name:   v2_3_0.UpgradeName,
		Height: upgradeHeight,
	}))
	_, err := testApp.PreBlocker(ctx, &abci.RequestFinalizeBlock{Height: upgradeHeight})
	require.NoError(t, err)

	toVM, err := testApp.UpgradeKeeper.GetModuleVersionMap(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(6), toVM[ibcexported.ModuleName])
	require.Equal(t, uint64(5), toVM[ibctransfertypes.ModuleName])

	require.Equal(t, expectedClientParams, testApp.IBCKeeper.ClientKeeper.GetParams(ctx))
	require.Equal(t, expectedConnectionParams, testApp.IBCKeeper.ConnectionKeeper.GetParams(ctx))
	require.Equal(t, ibcchanneltypes.DefaultParams(), testApp.IBCKeeper.ChannelKeeper.GetParams(ctx))
}
