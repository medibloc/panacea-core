package keeper

import (
	"bytes"
	"testing"

	"cosmossdk.io/collections"
	upstreamkeeper "cosmossdk.io/x/nft/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
)

func TestExportGenesisRejectsStaleUnknownOwnerIndex(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	_, _, _, _, _ = createNFTForBurnTest(t, &fixture)
	store := fixture.keeper.nftStoreService.OpenKVStore(fixture.ctx)
	staleKey := append(
		append([]byte(nil), upstreamkeeper.NFTOfClassByOwnerKey...),
		[]byte("stale-owner-index")...,
	)
	require.NoError(t, store.Set(staleKey, upstreamkeeper.Placeholder))

	_, err := fixture.keeper.ExportGenesis(fixture.ctx)
	require.ErrorContains(t, err, "standard nft reverse owner index key count 2 does not match expected count 1")
}

func TestExportGenesisAcceptsCacheContext(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, _, _, _, _ := createNFTForBurnTest(t, &fixture)
	cacheCtx, _ := fixture.ctx.CacheContext()

	exported, err := fixture.keeper.ExportGenesis(cacheCtx)
	require.NoError(t, err)
	require.Len(t, exported.NftState.Classes, 1)
	require.Equal(t, classID, exported.NftState.Classes[0].Id)
	require.Len(t, exported.NftState.Entries, 1)
	require.Len(t, exported.NftState.Entries[0].Nfts, 1)
}

func TestExportGenesisRejectsOrphanStandardKeys(t *testing.T) {
	tests := []struct {
		name        string
		prefix      []byte
		errorString string
	}{
		{
			name:        "nft",
			prefix:      upstreamkeeper.NFTKey,
			errorString: "standard nft key count 2 does not match expected count 1",
		},
		{
			name:        "direct owner",
			prefix:      upstreamkeeper.OwnerKey,
			errorString: "standard nft direct owner key count 2 does not match expected count 1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newKeeperFixture(t, true, true)
			_, _, _, _, _ = createNFTForBurnTest(t, &fixture)
			store := fixture.keeper.nftStoreService.OpenKVStore(fixture.ctx)
			orphanKey := append(append([]byte(nil), test.prefix...), []byte("orphan")...)
			require.NoError(t, store.Set(orphanKey, []byte("orphan-value")))

			_, err := fixture.keeper.ExportGenesis(fixture.ctx)
			require.ErrorContains(t, err, test.errorString)
		})
	}
}

func TestExportGenesisRejectsInvalidDerivedSupply(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(t *testing.T, fixture *keeperFixture, supplyKey []byte)
		errorString string
	}{
		{
			name: "mismatched value",
			mutate: func(t *testing.T, fixture *keeperFixture, supplyKey []byte) {
				store := fixture.keeper.nftStoreService.OpenKVStore(fixture.ctx)
				require.NoError(t, store.Set(supplyKey, sdk.Uint64ToBigEndian(2)))
			},
			errorString: "stored supply 2 does not match live lifecycle count 1",
		},
		{
			name: "missing positive supply",
			mutate: func(t *testing.T, fixture *keeperFixture, supplyKey []byte) {
				store := fixture.keeper.nftStoreService.OpenKVStore(fixture.ctx)
				require.NoError(t, store.Delete(supplyKey))
			},
			errorString: "stored supply 0 does not match live lifecycle count 1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newKeeperFixture(t, true, true)
			classID, _, _, _, _ := createNFTForBurnTest(t, &fixture)
			supplyKey := append(
				append([]byte(nil), upstreamkeeper.ClassTotalSupply...),
				[]byte(classID)...,
			)
			test.mutate(t, &fixture, supplyKey)

			_, err := fixture.keeper.ExportGenesis(fixture.ctx)
			require.ErrorContains(t, err, test.errorString)
		})
	}
}

func TestExportGenesisAcceptsZeroSupplyStorageForms(t *testing.T) {
	t.Run("absent before first mint", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, _ := createClassForMintTest(
			t,
			&fixture,
			sdk.AccAddress(bytes.Repeat([]byte{83}, 20)),
			10,
		)
		store := fixture.keeper.nftStoreService.OpenKVStore(fixture.ctx)
		supplyKey := append(
			append([]byte(nil), upstreamkeeper.ClassTotalSupply...),
			[]byte(classID)...,
		)
		hasSupplyKey, err := store.Has(supplyKey)
		require.NoError(t, err)
		require.False(t, hasSupplyKey)

		_, err = fixture.keeper.ExportGenesis(fixture.ctx)
		require.NoError(t, err)
	})

	t.Run("stored zero after final burn", func(t *testing.T) {
		fixture := newKeeperFixture(t, true, true)
		classID, _, owner, _, _ := createNFTForBurnTest(t, &fixture)
		_, err := NewMsgServer(fixture.keeper).Burn(
			sdk.WrapSDKContext(fixture.ctx),
			&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "nft-1", Owner: owner},
		)
		require.NoError(t, err)
		store := fixture.keeper.nftStoreService.OpenKVStore(fixture.ctx)
		supplyKey := append(
			append([]byte(nil), upstreamkeeper.ClassTotalSupply...),
			[]byte(classID)...,
		)
		hasSupplyKey, err := store.Has(supplyKey)
		require.NoError(t, err)
		require.True(t, hasSupplyKey)

		_, err = fixture.keeper.ExportGenesis(fixture.ctx)
		require.NoError(t, err)
	})
}

func TestExportGenesisRejectsInvalidOwnerClassCounts(t *testing.T) {
	for _, test := range []struct {
		name        string
		mutate      func(t *testing.T, fixture *keeperFixture, classID, owner string)
		errorString string
	}{
		{
			name: "missing",
			mutate: func(t *testing.T, fixture *keeperFixture, classID, owner string) {
				require.NoError(t, fixture.keeper.ownerClassCounts.Remove(
					fixture.ctx,
					collections.Join(classID, owner),
				))
			},
			errorString: "balance count is missing",
		},
		{
			name: "mismatched",
			mutate: func(t *testing.T, fixture *keeperFixture, classID, owner string) {
				require.NoError(t, fixture.keeper.ownerClassCounts.Set(
					fixture.ctx,
					collections.Join(classID, owner),
					2,
				))
			},
			errorString: "balance count 2 does not match expected 1",
		},
		{
			name: "stored zero",
			mutate: func(t *testing.T, fixture *keeperFixture, classID, owner string) {
				require.NoError(t, fixture.keeper.ownerClassCounts.Set(
					fixture.ctx,
					collections.Join(classID, owner),
					0,
				))
			},
			errorString: "stored zero balance count",
		},
		{
			name: "orphan",
			mutate: func(t *testing.T, fixture *keeperFixture, _ string, owner string) {
				require.NoError(t, fixture.keeper.ownerClassCounts.Set(
					fixture.ctx,
					collections.Join("orphan-class", owner),
					1,
				))
			},
			errorString: "balance count has no live nfts",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := newKeeperFixture(t, true, true)
			classID, _, owner, _, _ := createNFTForBurnTest(t, &fixture)
			test.mutate(t, &fixture, classID, owner)

			_, err := fixture.keeper.ExportGenesis(fixture.ctx)
			require.ErrorContains(t, err, test.errorString)
		})
	}
}
