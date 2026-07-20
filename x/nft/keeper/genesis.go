package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
)

var storeAvailabilityProbeKey = []byte{0xff}

// InitGenesis initializes the empty dual-store module skeleton.
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) error {
	if data == nil {
		return fmt.Errorf("nft genesis state must not be nil")
	}
	if err := types.ValidateGenesis(*data, k.addressCodec); err != nil {
		return err
	}
	if err := k.ensureStoresAvailable(ctx); err != nil {
		return err
	}

	k.nftKeeper.InitGenesis(ctx, data.NftState)
	return nil
}

// ExportGenesis exports standard NFT state and Panacea policy state in
// deterministic collection-key order.
func (k Keeper) ExportGenesis(ctx sdk.Context) (*types.GenesisState, error) {
	if err := k.ensureStoresAvailable(ctx); err != nil {
		return nil, err
	}

	classPolicies := make([]*types.ClassPolicy, 0)
	if err := k.classPolicies.Walk(ctx, nil, func(_ string, value types.ClassPolicy) (bool, error) {
		valueCopy := value
		classPolicies = append(classPolicies, &valueCopy)
		return false, nil
	}); err != nil {
		return nil, fmt.Errorf("export class policies: %w", err)
	}

	lifecycles := make([]*types.LifecycleRecord, 0)
	if err := k.lifecycles.Walk(ctx, nil, func(_ collections.Pair[string, string], value types.LifecycleRecord) (bool, error) {
		valueCopy := value
		lifecycles = append(lifecycles, &valueCopy)
		return false, nil
	}); err != nil {
		return nil, fmt.Errorf("export lifecycles: %w", err)
	}

	tombstones := make([]*types.BurnTombstone, 0)
	if err := k.tombstones.Walk(ctx, nil, func(_ collections.Pair[string, string], value types.BurnTombstone) (bool, error) {
		valueCopy := value
		tombstones = append(tombstones, &valueCopy)
		return false, nil
	}); err != nil {
		return nil, fmt.Errorf("export tombstones: %w", err)
	}

	return &types.GenesisState{
		NftState:      k.nftKeeper.ExportGenesis(ctx),
		ClassPolicies: classPolicies,
		Lifecycles:    lifecycles,
		Tombstones:    tombstones,
	}, nil
}

func (k Keeper) ensureStoresAvailable(ctx sdk.Context) error {
	if _, err := k.nftStoreService.OpenKVStore(ctx).Has(storeAvailabilityProbeKey); err != nil {
		return fmt.Errorf("open nft store: %w", err)
	}
	if _, err := k.policyStoreService.OpenKVStore(ctx).Has(storeAvailabilityProbeKey); err != nil {
		return fmt.Errorf("open nftpolicy store: %w", err)
	}
	return nil
}
