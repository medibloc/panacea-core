package v2_3_0

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	dbm "github.com/cosmos/cosmos-db"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	pnfttypes "github.com/medibloc/panacea-core/v2/x/pnft/types"
	"github.com/stretchr/testify/require"
)

func TestNormalizeMigratedExpeditedVotingPeriod(t *testing.T) {
	t.Parallel()

	t.Run("short legacy period is clamped", func(t *testing.T) {
		regular := 20 * time.Second
		expedited := govv1.DefaultExpeditedPeriod
		params := govv1.DefaultParams()
		params.VotingPeriod = &regular
		params.ExpeditedVotingPeriod = &expedited

		require.NoError(t, normalizeMigratedExpeditedVotingPeriod(&params))
		require.Equal(t, 10*time.Second, *params.ExpeditedVotingPeriod)
	})

	t.Run("valid mainnet periods remain unchanged", func(t *testing.T) {
		regular := 3 * 24 * time.Hour
		expedited := 24 * time.Hour
		params := govv1.DefaultParams()
		params.VotingPeriod = &regular
		params.ExpeditedVotingPeriod = &expedited
		originalRegular := params.VotingPeriod
		originalExpedited := params.ExpeditedVotingPeriod

		require.NoError(t, normalizeMigratedExpeditedVotingPeriod(&params))
		require.Same(t, originalRegular, params.VotingPeriod)
		require.Equal(t, 3*24*time.Hour, *params.VotingPeriod)
		require.Same(t, originalExpedited, params.ExpeditedVotingPeriod)
		require.Equal(t, 24*time.Hour, *params.ExpeditedVotingPeriod)
	})

	t.Run("nil inputs fail closed", func(t *testing.T) {
		require.ErrorContains(t, normalizeMigratedExpeditedVotingPeriod(nil), "nil governance params")

		params := govv1.DefaultParams()
		params.VotingPeriod = nil
		require.ErrorContains(t, normalizeMigratedExpeditedVotingPeriod(&params), "without regular voting period")

		regular := 20 * time.Second
		params.VotingPeriod = &regular
		params.ExpeditedVotingPeriod = nil
		require.ErrorContains(t, normalizeMigratedExpeditedVotingPeriod(&params), "nil expedited voting period")
	})

	for _, regular := range []time.Duration{-time.Nanosecond, 0, time.Nanosecond} {
		regular := regular
		t.Run("regular period cannot be safely halved/"+regular.String(), func(t *testing.T) {
			expedited := regular
			params := govv1.DefaultParams()
			params.VotingPeriod = &regular
			params.ExpeditedVotingPeriod = &expedited

			require.ErrorContains(t, normalizeMigratedExpeditedVotingPeriod(&params), "cannot derive")
		})
	}
}

func TestStoreUpgradeAddsNFTStoresAndRetainsLegacyPNFTStore(t *testing.T) {
	require.Equal(t, []string{nfttypes.StoreKey, nfttypes.PolicyStoreKey}, Upgrade.StoreUpgrades.Added)
	require.Empty(t, Upgrade.StoreUpgrades.Deleted)

	db := dbm.NewMemDB()
	legacyStore := rootmulti.NewStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	pnftKey := storetypes.NewKVStoreKey(pnfttypes.StoreKey)
	stableKey := storetypes.NewKVStoreKey("stable")
	legacyStore.MountStoreWithDB(pnftKey, storetypes.StoreTypeIAVL, nil)
	legacyStore.MountStoreWithDB(stableKey, storetypes.StoreTypeIAVL, nil)
	require.NoError(t, legacyStore.LoadLatestVersion())

	pnftStore := legacyStore.GetStore(pnftKey).(storetypes.KVStore)
	stableStore := legacyStore.GetStore(stableKey).(storetypes.KVStore)
	pnftStore.Set([]byte("legacy"), []byte("retained"))
	stableStore.Set([]byte("preserved"), []byte("value"))
	require.Equal(t, int64(1), legacyStore.Commit().Version)

	upgradedStore := rootmulti.NewStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	nftKey := storetypes.NewKVStoreKey(nfttypes.StoreKey)
	policyKey := storetypes.NewKVStoreKey(nfttypes.PolicyStoreKey)
	upgradedStore.MountStoreWithDB(stableKey, storetypes.StoreTypeIAVL, nil)
	upgradedStore.MountStoreWithDB(pnftKey, storetypes.StoreTypeIAVL, nil)
	upgradedStore.MountStoreWithDB(nftKey, storetypes.StoreTypeIAVL, nil)
	upgradedStore.MountStoreWithDB(policyKey, storetypes.StoreTypeIAVL, nil)
	loader := upgradetypes.UpgradeStoreLoader(2, &Upgrade.StoreUpgrades)
	require.NoError(t, loader(upgradedStore))
	require.Equal(t, []byte("value"), upgradedStore.GetStore(stableKey).(storetypes.KVStore).Get([]byte("preserved")))
	require.Equal(t, []byte("retained"), upgradedStore.GetStore(pnftKey).(storetypes.KVStore).Get([]byte("legacy")))

	require.NotNil(t, upgradedStore.GetStore(nftKey))
	require.NotNil(t, upgradedStore.GetStore(policyKey))
	require.Equal(t, int64(2), upgradedStore.Commit().Version)

	commitInfo, err := upgradedStore.GetCommitInfo(2)
	require.NoError(t, err)
	storeNames := make([]string, 0, len(commitInfo.StoreInfos))
	for _, storeInfo := range commitInfo.StoreInfos {
		storeNames = append(storeNames, storeInfo.Name)
	}
	require.ElementsMatch(
		t,
		[]string{"stable", pnfttypes.StoreKey, nfttypes.StoreKey, nfttypes.PolicyStoreKey},
		storeNames,
	)
}
