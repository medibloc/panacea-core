package types

import (
	"fmt"

	"cosmossdk.io/core/address"
	upstreamnft "cosmossdk.io/x/nft"
)

// DefaultGenesis returns the only state accepted by the empty module skeleton.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		NftState:      upstreamnft.DefaultGenesisState(),
		ClassPolicies: []*ClassPolicy{},
		Lifecycles:    []*LifecycleRecord{},
		Tombstones:    []*BurnTombstone{},
	}
}

// ValidateGenesis currently accepts only empty state. Non-empty genesis remains
// disabled until initialization can enforce invariants across both stores.
func ValidateGenesis(data GenesisState, addressCodec address.Codec) error {
	if data.NftState == nil {
		return fmt.Errorf("nft_state must not be nil")
	}
	if err := upstreamnft.ValidateGenesis(*data.NftState, addressCodec); err != nil {
		return fmt.Errorf("invalid standard nft genesis: %w", err)
	}
	if len(data.NftState.Classes) != 0 || len(data.NftState.Entries) != 0 {
		return fmt.Errorf("non-empty standard nft genesis is not supported by the empty module skeleton")
	}
	if len(data.ClassPolicies) != 0 || len(data.Lifecycles) != 0 || len(data.Tombstones) != 0 {
		return fmt.Errorf("non-empty nftpolicy genesis is not supported by the empty module skeleton")
	}
	return nil
}
