package types

import (
	"testing"

	upstreamnft "cosmossdk.io/x/nft"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/stretchr/testify/require"
)

func TestDefaultGenesisIsEmptyAndNonNil(t *testing.T) {
	genesis := DefaultGenesis()

	require.NotNil(t, genesis)
	require.NotNil(t, genesis.NftState)
	require.Empty(t, genesis.NftState.Classes)
	require.Empty(t, genesis.NftState.Entries)
	require.NotNil(t, genesis.ClassPolicies)
	require.NotNil(t, genesis.Lifecycles)
	require.NotNil(t, genesis.Tombstones)
	require.NoError(t, ValidateGenesis(*genesis, address.NewBech32Codec("panacea")))
}

func TestValidateGenesisRejectsNilStandardNFTState(t *testing.T) {
	genesis := DefaultGenesis()
	genesis.NftState = nil

	err := ValidateGenesis(*genesis, address.NewBech32Codec("panacea"))
	require.ErrorContains(t, err, "nft_state must not be nil")
}

func TestValidateGenesisRejectsNonEmptySkeletonState(t *testing.T) {
	addressCodec := address.NewBech32Codec("panacea")

	t.Run("standard nft state", func(t *testing.T) {
		genesis := DefaultGenesis()
		genesis.NftState = &upstreamnft.GenesisState{
			Classes: []*upstreamnft.Class{{Id: "class-1"}},
		}

		err := ValidateGenesis(*genesis, addressCodec)
		require.ErrorContains(t, err, "non-empty standard nft genesis")
	})

	t.Run("policy state", func(t *testing.T) {
		genesis := DefaultGenesis()
		genesis.ClassPolicies = []*ClassPolicy{{ClassId: "class-1"}}

		err := ValidateGenesis(*genesis, addressCodec)
		require.ErrorContains(t, err, "non-empty nftpolicy genesis")
	})
}
