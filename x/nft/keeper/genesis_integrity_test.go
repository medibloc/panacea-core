package keeper

import (
	"bytes"
	"testing"

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
