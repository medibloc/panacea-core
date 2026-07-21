package keeper

import (
	"testing"

	upstreamkeeper "cosmossdk.io/x/nft/keeper"
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
