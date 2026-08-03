package app_test

import (
	"path/filepath"
	"testing"
	"time"

	"cosmossdk.io/core/header"
	"cosmossdk.io/log"
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	aoltypes "github.com/medibloc/panacea-core/v2/x/aol/types"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	pnfttypes "github.com/medibloc/panacea-core/v2/x/pnft/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	panaceaapp "github.com/medibloc/panacea-core/v2/app"
	"github.com/medibloc/panacea-core/v2/app/upgrades/v2_3_0"
)

func TestV230StoreUpgradeAndRestartRehearsal(t *testing.T) {
	panaceaapp.SetConfig()

	const (
		legacyHeight  = int64(2)
		upgradeHeight = int64(3)
	)
	var (
		preservedKey    = []byte("v2.3.0-rehearsal/preserved")
		preservedValue  = []byte("stable-state")
		legacyPNFTKey   = []byte("legacy-pnft")
		legacyPNFTData  = []byte("retained")
		uncommittedKey  = []byte("v2.3.0-rehearsal/uncommitted")
		uncommittedData = []byte("must-not-persist")
	)

	homeDir := t.TempDir()
	dataDir := filepath.Join(homeDir, "data")
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, homeDir)
	plan := upgradetypes.Plan{Name: v2_3_0.UpgradeName, Height: upgradeHeight}
	blockTime := time.Now().UTC()

	// Use the current app only as a source of persistent store keys, codecs, and
	// the upgrade keeper. Its stores are never loaded or committed.
	templateDB := dbm.NewMemDB()
	templateApp := panaceaapp.New(log.NewNopLogger(), templateDB, nil, false, appOpts)
	fromVM := v230LegacyVersionMap(templateApp.ModuleManager.GetVersionMap())

	// Build the previous binary's topology directly, so the two new NFT IAVL
	// stores have never existed in the physical database.
	legacyDB := openRehearsalDB(t, dataDir)
	legacyStore := rootmulti.NewStore(legacyDB, log.NewNopLogger(), metrics.NewNoOpMetrics())
	pnftKey := templateApp.GetKey(pnfttypes.StoreKey)
	require.NotNil(t, pnftKey)
	for name, key := range templateApp.GetKVStoreKey() {
		if name == nfttypes.StoreKey || name == nfttypes.PolicyStoreKey {
			continue
		}
		legacyStore.MountStoreWithDB(key, storetypes.StoreTypeIAVL, nil)
	}
	for _, key := range templateApp.GetTransientStoreKey() {
		legacyStore.MountStoreWithDB(key, storetypes.StoreTypeTransient, nil)
	}
	require.NoError(t, legacyStore.LoadLatestVersion())

	legacyCtx := sdk.NewContext(
		legacyStore,
		cmtproto.Header{Height: 1, Time: blockTime},
		false,
		log.NewNopLogger(),
	).WithHeaderInfo(header.Info{Height: 1, Time: blockTime})
	legacyIBCSubspace := templateApp.GetSubspace(ibcexported.ModuleName)
	legacyIBCClientParams := ibcclienttypes.NewParams(ibcexported.Solomachine, ibcexported.Tendermint)
	legacyIBCConnectionParams := ibcconnectiontypes.DefaultParams()
	legacyIBCSubspace.SetParamSet(legacyCtx, &legacyIBCClientParams)
	legacyIBCSubspace.SetParamSet(legacyCtx, &legacyIBCConnectionParams)
	legacyTransferParams := ibctransfertypes.NewParams(false, true)
	legacyTransferSubspace := templateApp.GetSubspace(ibctransfertypes.ModuleName)
	legacyTransferSubspace.SetParamSet(legacyCtx, &legacyTransferParams)
	require.NoError(t, templateApp.UpgradeKeeper.SetModuleVersionMap(legacyCtx, fromVM))
	require.NoError(t, templateApp.UpgradeKeeper.ScheduleUpgrade(legacyCtx, plan))
	require.NoError(t, templateApp.UpgradeKeeper.DumpUpgradeInfoToDisk(upgradeHeight, plan))
	legacyCtx.KVStore(templateApp.GetKey(aoltypes.StoreKey)).Set(preservedKey, preservedValue)
	legacyStore.GetStore(pnftKey).(storetypes.KVStore).Set(legacyPNFTKey, legacyPNFTData)
	require.Equal(t, int64(1), legacyStore.Commit().Version)
	require.Equal(t, legacyHeight, legacyStore.Commit().Version)
	require.Equal(
		t,
		legacyPNFTData,
		legacyStore.GetStore(pnftKey).(storetypes.KVStore).Get(legacyPNFTKey),
	)
	legacyStoreNames := rehearsalStoreNames(t, legacyStore, legacyHeight)
	require.Contains(t, legacyStoreNames, pnfttypes.StoreKey)
	require.NotContains(t, legacyStoreNames, nfttypes.StoreKey)
	require.NotContains(t, legacyStoreNames, nfttypes.PolicyStoreKey)
	require.NoError(t, templateDB.Close())
	require.NoError(t, legacyDB.Close())

	// Simulate the new process stopping after the store loader and PreBlock
	// succeed but before commit. Its in-memory store changes must not advance or
	// alter the persisted legacy commit.
	interruptedDB := openRehearsalDB(t, dataDir)
	interruptedApp := panaceaapp.New(log.NewNopLogger(), interruptedDB, nil, false, appOpts)
	require.NoError(t, interruptedApp.LoadLatestVersion())
	require.Equal(t, legacyHeight, interruptedApp.LastBlockHeight())

	interruptedCtx := interruptedApp.NewUncachedContext(
		false,
		cmtproto.Header{Height: upgradeHeight, Time: blockTime},
	).WithHeaderInfo(header.Info{Height: upgradeHeight, Time: blockTime})
	interruptedCtx.KVStore(interruptedApp.GetKey(nfttypes.PolicyStoreKey)).Set(uncommittedKey, uncommittedData)
	require.Equal(
		t,
		uncommittedData,
		interruptedCtx.KVStore(interruptedApp.GetKey(nfttypes.PolicyStoreKey)).Get(uncommittedKey),
	)
	uncommittedStoreNames := rehearsalStoreNames(
		t,
		interruptedApp.CommitMultiStore().(*rootmulti.Store),
		legacyHeight,
	)
	require.Contains(t, uncommittedStoreNames, pnfttypes.StoreKey)
	require.NotContains(t, uncommittedStoreNames, nfttypes.StoreKey)
	require.NotContains(t, uncommittedStoreNames, nfttypes.PolicyStoreKey)

	_, interruptedPreBlockErr := interruptedApp.PreBlocker(
		interruptedCtx,
		&abci.RequestFinalizeBlock{Height: upgradeHeight},
	)
	require.NoError(t, interruptedPreBlockErr)
	interruptedVM, interruptedVMErr := interruptedApp.UpgradeKeeper.GetModuleVersionMap(interruptedCtx)
	require.NoError(t, interruptedVMErr)
	requireV230MigratedModuleVersions(t, interruptedVM)
	interruptedDoneHeight, interruptedDoneErr := interruptedApp.UpgradeKeeper.GetDoneHeight(
		interruptedCtx,
		v2_3_0.UpgradeName,
	)
	require.NoError(t, interruptedDoneErr)
	require.Equal(t, upgradeHeight, interruptedDoneHeight)
	_, interruptedPlanErr := interruptedApp.UpgradeKeeper.GetUpgradePlan(interruptedCtx)
	require.ErrorIs(t, interruptedPlanErr, upgradetypes.ErrNoUpgradePlanFound)
	require.NoError(t, interruptedDB.Close())

	// The new binary reads the real upgrade-info file while it is constructed.
	// Restarting from the unchanged legacy commit reapplies the production
	// v2.3.0 loader.
	upgradeDB := openRehearsalDB(t, dataDir)
	upgradedApp := panaceaapp.New(log.NewNopLogger(), upgradeDB, nil, false, appOpts)
	require.NoError(t, upgradedApp.LoadLatestVersion())
	require.Equal(t, legacyHeight, upgradedApp.LastBlockHeight())

	upgradeCtx := upgradedApp.NewUncachedContext(
		false,
		cmtproto.Header{Height: upgradeHeight, Time: blockTime},
	).WithHeaderInfo(header.Info{Height: upgradeHeight, Time: blockTime})
	recoveredVM, recoveredVMErr := upgradedApp.UpgradeKeeper.GetModuleVersionMap(upgradeCtx)
	require.NoError(t, recoveredVMErr)
	require.NotContains(t, recoveredVM, nfttypes.ModuleName)
	require.Equal(t, uint64(1), recoveredVM[pnfttypes.ModuleName])
	recoveredDoneHeight, recoveredDoneErr := upgradedApp.UpgradeKeeper.GetDoneHeight(
		upgradeCtx,
		v2_3_0.UpgradeName,
	)
	require.NoError(t, recoveredDoneErr)
	require.Zero(t, recoveredDoneHeight)
	recoveredPlan, recoveredPlanErr := upgradedApp.UpgradeKeeper.GetUpgradePlan(upgradeCtx)
	require.NoError(t, recoveredPlanErr)
	require.Equal(t, plan, recoveredPlan)
	require.Nil(
		t,
		upgradeCtx.KVStore(upgradedApp.GetKey(nfttypes.PolicyStoreKey)).Get(uncommittedKey),
	)
	_, err := upgradedApp.PreBlocker(upgradeCtx, &abci.RequestFinalizeBlock{Height: upgradeHeight})
	require.NoError(t, err)

	toVM, err := upgradedApp.UpgradeKeeper.GetModuleVersionMap(upgradeCtx)
	require.NoError(t, err)
	requireV230MigratedModuleVersions(t, toVM)
	require.Equal(t, upgradeHeight, upgradedApp.CommitMultiStore().Commit().Version)
	require.NoError(t, upgradeDB.Close())

	// A second process using the same binary must load the committed result
	// without applying the store upgrade again.
	restartDB := openRehearsalDB(t, dataDir)
	restartedApp := panaceaapp.New(log.NewNopLogger(), restartDB, nil, false, appOpts)
	require.NoError(t, restartedApp.LoadLatestVersion())
	require.Equal(t, upgradeHeight, restartedApp.LastBlockHeight())

	restartCtx := restartedApp.NewUncachedContext(
		false,
		cmtproto.Header{Height: upgradeHeight, Time: blockTime},
	).WithHeaderInfo(header.Info{Height: upgradeHeight, Time: blockTime})
	require.Equal(
		t,
		preservedValue,
		restartCtx.KVStore(restartedApp.GetKey(aoltypes.StoreKey)).Get(preservedKey),
	)
	require.Equal(
		t,
		legacyPNFTData,
		restartCtx.KVStore(restartedApp.GetKey(pnfttypes.StoreKey)).Get(legacyPNFTKey),
	)

	restartedVM, err := restartedApp.UpgradeKeeper.GetModuleVersionMap(restartCtx)
	require.NoError(t, err)
	requireV230MigratedModuleVersions(t, restartedVM)
	doneHeight, err := restartedApp.UpgradeKeeper.GetDoneHeight(restartCtx, v2_3_0.UpgradeName)
	require.NoError(t, err)
	require.Equal(t, upgradeHeight, doneHeight)
	_, err = restartedApp.UpgradeKeeper.GetUpgradePlan(restartCtx)
	require.ErrorIs(t, err, upgradetypes.ErrNoUpgradePlanFound)

	genesis, err := restartedApp.NFTKeeper.ExportGenesis(restartCtx)
	require.NoError(t, err)
	require.NotNil(t, genesis.NftState)
	require.Empty(t, genesis.NftState.Classes)
	require.Empty(t, genesis.NftState.Entries)
	require.Empty(t, genesis.ClassPolicies)
	require.Empty(t, genesis.Lifecycles)
	require.Empty(t, genesis.Tombstones)

	committedStoreNames := rehearsalStoreNames(
		t,
		restartedApp.CommitMultiStore().(*rootmulti.Store),
		upgradeHeight,
	)
	require.Contains(t, committedStoreNames, nfttypes.StoreKey)
	require.Contains(t, committedStoreNames, nfttypes.PolicyStoreKey)
	require.Contains(t, committedStoreNames, pnfttypes.StoreKey)
	require.NoError(t, restartDB.Close())
}

func v230LegacyVersionMap(current module.VersionMap) module.VersionMap {
	current[authtypes.ModuleName] = 4
	current[stakingtypes.ModuleName] = 4
	current[slashingtypes.ModuleName] = 3
	current[govtypes.ModuleName] = 4
	current[ibcexported.ModuleName] = 4
	current[ibctransfertypes.ModuleName] = 3
	delete(current, nfttypes.ModuleName)
	current[pnfttypes.ModuleName] = 1
	return current
}

func requireV230MigratedModuleVersions(t *testing.T, versions module.VersionMap) {
	t.Helper()

	require.Equal(t, uint64(5), versions[authtypes.ModuleName])
	require.Equal(t, uint64(5), versions[stakingtypes.ModuleName])
	require.Equal(t, uint64(4), versions[slashingtypes.ModuleName])
	require.Equal(t, uint64(5), versions[govtypes.ModuleName])
	require.Equal(t, uint64(6), versions[ibcexported.ModuleName])
	require.Equal(t, uint64(5), versions[ibctransfertypes.ModuleName])
	require.Equal(t, uint64(1), versions[nfttypes.ModuleName])
	require.Equal(t, uint64(1), versions[pnfttypes.ModuleName])
}

func openRehearsalDB(t *testing.T, dataDir string) dbm.DB {
	t.Helper()

	db, err := dbm.NewDB("application", dbm.GoLevelDBBackend, dataDir)
	require.NoError(t, err)
	return db
}

func rehearsalStoreNames(t *testing.T, store *rootmulti.Store, height int64) []string {
	t.Helper()

	commitInfo, err := store.GetCommitInfo(height)
	require.NoError(t, err)
	names := make([]string, 0, len(commitInfo.StoreInfos))
	for _, storeInfo := range commitInfo.StoreInfos {
		names = append(names, storeInfo.Name)
	}
	return names
}
