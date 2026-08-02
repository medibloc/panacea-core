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
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	panaceaapp "github.com/medibloc/panacea-core/v2/app"
	"github.com/medibloc/panacea-core/v2/app/upgrades/v2_3_0"
)

func TestV230UpgradeMigratesTransferState(t *testing.T) {
	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)
	require.NoError(t, testApp.LoadLatestVersion())

	const upgradeHeight = int64(10)
	blockTime := time.Unix(1_700_000_000, 0).UTC()
	ctx := testApp.NewUncachedContext(
		false,
		cmtproto.Header{Height: upgradeHeight, Time: blockTime},
	).WithHeaderInfo(header.Info{Height: upgradeHeight, Time: blockTime})

	fromVM := testApp.ModuleManager.GetVersionMap()
	fromVM[ibctransfertypes.ModuleName] = 3
	require.Equal(t, uint64(6), fromVM[ibcexported.ModuleName])
	require.NoError(t, testApp.UpgradeKeeper.SetModuleVersionMap(ctx, fromVM))

	expectedParams := ibctransfertypes.NewParams(false, true)
	legacyTransferSubspace := testApp.GetSubspace(ibctransfertypes.ModuleName)
	legacyTransferSubspace.SetParamSet(ctx, &expectedParams)

	denomTrace := ibctransfertypes.DenomTrace{
		Path:      "transfer/channel-1",
		BaseDenom: "uosmo",
	}
	require.NoError(t, denomTrace.Validate())
	testApp.TransferKeeper.SetDenomTrace(ctx, denomTrace)
	storedTrace, found := testApp.TransferKeeper.GetDenomTrace(ctx, denomTrace.Hash())
	require.True(t, found)
	require.Equal(t, denomTrace, storedTrace)
	require.False(t, testApp.BankKeeper.HasDenomMetaData(ctx, denomTrace.IBCDenom()))

	transferModuleAccount := authtypes.NewEmptyModuleAccount(
		ibctransfertypes.ModuleName,
		authtypes.Minter,
		authtypes.Burner,
	)
	testApp.AccountKeeper.SetModuleAccount(ctx, transferModuleAccount)
	require.NoError(t, testApp.BankKeeper.MintCoins(
		ctx,
		ibctransfertypes.ModuleName,
		sdk.NewCoins(sdk.NewInt64Coin(denomTrace.IBCDenom(), 123)),
	))
	balanceBeforeUpgrade := testApp.BankKeeper.GetBalance(
		ctx,
		transferModuleAccount.GetAddress(),
		denomTrace.IBCDenom(),
	)
	require.Equal(t, sdk.NewInt64Coin(denomTrace.IBCDenom(), 123), balanceBeforeUpgrade)

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

	require.Equal(t, expectedParams, testApp.TransferKeeper.GetParams(ctx))

	storedTrace, found = testApp.TransferKeeper.GetDenomTrace(ctx, denomTrace.Hash())
	require.True(t, found)
	require.Equal(t, denomTrace, storedTrace)

	metadata, found := testApp.BankKeeper.GetDenomMetaData(ctx, denomTrace.IBCDenom())
	require.True(t, found)
	require.Equal(t, denomTrace.IBCDenom(), metadata.Base)
	require.Equal(t, "transfer/channel-1/uosmo", metadata.Display)
	require.Equal(t, "IBC token from transfer/channel-1/uosmo", metadata.Description)
	require.Equal(t, "transfer/channel-1/uosmo IBC token", metadata.Name)
	require.Equal(t, "UOSMO", metadata.Symbol)
	require.Len(t, metadata.DenomUnits, 1)
	require.Equal(t, "uosmo", metadata.DenomUnits[0].Denom)
	require.Equal(t, uint32(0), metadata.DenomUnits[0].Exponent)

	require.Equal(
		t,
		balanceBeforeUpgrade,
		testApp.BankKeeper.GetBalance(
			ctx,
			transferModuleAccount.GetAddress(),
			denomTrace.IBCDenom(),
		),
	)
}
