package v2_3_0

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	dbm "github.com/cosmos/cosmos-db"
	pnfttypes "github.com/medibloc/panacea-core/v2/x/pnft/types"
	"github.com/stretchr/testify/require"
)

func TestStoreUpgradeDeletesLegacyPNFTStore(t *testing.T) {
	require.Equal(t, []string{pnfttypes.StoreKey}, Upgrade.StoreUpgrades.Deleted)

	db := dbm.NewMemDB()
	legacyStore := rootmulti.NewStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	pnftKey := storetypes.NewKVStoreKey(pnfttypes.StoreKey)
	stableKey := storetypes.NewKVStoreKey("stable")
	legacyStore.MountStoreWithDB(pnftKey, storetypes.StoreTypeIAVL, nil)
	legacyStore.MountStoreWithDB(stableKey, storetypes.StoreTypeIAVL, nil)
	require.NoError(t, legacyStore.LoadLatestVersion())

	pnftStore := legacyStore.GetStore(pnftKey).(storetypes.KVStore)
	stableStore := legacyStore.GetStore(stableKey).(storetypes.KVStore)
	pnftStore.Set([]byte("legacy"), []byte("discarded"))
	stableStore.Set([]byte("preserved"), []byte("value"))
	require.Equal(t, int64(1), legacyStore.Commit().Version)

	upgradedStore := rootmulti.NewStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	upgradedStore.MountStoreWithDB(stableKey, storetypes.StoreTypeIAVL, nil)
	loader := upgradetypes.UpgradeStoreLoader(2, &Upgrade.StoreUpgrades)
	require.NoError(t, loader(upgradedStore))
	require.Equal(t, []byte("value"), upgradedStore.GetStore(stableKey).(storetypes.KVStore).Get([]byte("preserved")))
	require.Equal(t, int64(2), upgradedStore.Commit().Version)

	commitInfo, err := upgradedStore.GetCommitInfo(2)
	require.NoError(t, err)
	require.Len(t, commitInfo.StoreInfos, 1)
	require.Equal(t, "stable", commitInfo.StoreInfos[0].Name)
}
