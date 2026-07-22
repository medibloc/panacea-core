package keeper

import (
	"bytes"
	"fmt"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
)

var storeAvailabilityProbeKey = []byte{0xff}

type genesisNFTKey struct {
	classID string
	nftID   string
}

type expectedGenesisState struct {
	supply          map[string]uint64
	mintedCount     map[string]uint64
	ownerByNFT      map[genesisNFTKey]sdk.AccAddress
	ownerClassCount map[string]map[string]uint64
}

// InitGenesis atomically restores standard NFT state and its coupled policy
// state after validating the complete input.
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) error {
	if data == nil {
		return fmt.Errorf("nft genesis state must not be nil")
	}
	if err := types.ValidateGenesis(*data, k.addressCodec, k.cdc); err != nil {
		return err
	}
	if err := k.validateGenesisModuleAccounts(*data); err != nil {
		return err
	}
	if err := k.ensureStoresAvailable(ctx); err != nil {
		return err
	}
	expected, err := k.expectedGenesisState(*data)
	if err != nil {
		return err
	}

	cacheCtx, writeCache := ctx.CacheContext()
	standardCtx := cacheCtx.WithEventManager(sdk.NewEventManager())
	k.nftKeeper.InitGenesis(standardCtx, data.NftState)

	for _, policy := range data.ClassPolicies {
		if err := k.classPolicies.Set(cacheCtx, policy.ClassId, *policy); err != nil {
			return fmt.Errorf("initialize class policy %s: %w", policy.ClassId, err)
		}
	}
	for _, lifecycle := range data.Lifecycles {
		key := collections.Join(lifecycle.ClassId, lifecycle.NftId)
		if err := k.lifecycles.Set(cacheCtx, key, *lifecycle); err != nil {
			return fmt.Errorf(
				"initialize lifecycle %s/%s: %w",
				lifecycle.ClassId,
				lifecycle.NftId,
				err,
			)
		}
	}
	for _, tombstone := range data.Tombstones {
		key := collections.Join(tombstone.ClassId, tombstone.NftId)
		if err := k.tombstones.Set(cacheCtx, key, *tombstone); err != nil {
			return fmt.Errorf(
				"initialize tombstone %s/%s: %w",
				tombstone.ClassId,
				tombstone.NftId,
				err,
			)
		}
	}
	for _, class := range data.NftState.Classes {
		if err := k.mintedCounts.Set(cacheCtx, class.Id, expected.mintedCount[class.Id]); err != nil {
			return fmt.Errorf("initialize minted count for class %s: %w", class.Id, err)
		}
	}
	if err := k.verifyInitializedGenesis(cacheCtx, *data, expected); err != nil {
		return err
	}
	writeCache()
	return nil
}

// ExportGenesis exports standard NFT state and Panacea policy state in
// deterministic collection-key order and rejects inconsistent stored state.
func (k Keeper) ExportGenesis(ctx sdk.Context) (*types.GenesisState, error) {
	if err := k.ensureStoresAvailable(ctx); err != nil {
		return nil, err
	}

	classPolicies := make([]*types.ClassPolicy, 0)
	if err := k.classPolicies.Walk(ctx, nil, func(key string, value types.ClassPolicy) (bool, error) {
		if value.ClassId != key {
			return true, fmt.Errorf("class policy key does not match value for %s", key)
		}
		valueCopy := value
		classPolicies = append(classPolicies, &valueCopy)
		return false, nil
	}); err != nil {
		return nil, fmt.Errorf("export class policies: %w", err)
	}

	lifecycles := make([]*types.LifecycleRecord, 0)
	if err := k.lifecycles.Walk(ctx, nil, func(key collections.Pair[string, string], value types.LifecycleRecord) (bool, error) {
		if value.ClassId != key.K1() || value.NftId != key.K2() {
			return true, fmt.Errorf(
				"lifecycle key does not match value for %s/%s",
				key.K1(),
				key.K2(),
			)
		}
		valueCopy := value
		lifecycles = append(lifecycles, &valueCopy)
		return false, nil
	}); err != nil {
		return nil, fmt.Errorf("export lifecycles: %w", err)
	}

	tombstones := make([]*types.BurnTombstone, 0)
	if err := k.tombstones.Walk(ctx, nil, func(key collections.Pair[string, string], value types.BurnTombstone) (bool, error) {
		if value.ClassId != key.K1() || value.NftId != key.K2() {
			return true, fmt.Errorf(
				"tombstone key does not match value for %s/%s",
				key.K1(),
				key.K2(),
			)
		}
		valueCopy := value
		tombstones = append(tombstones, &valueCopy)
		return false, nil
	}); err != nil {
		return nil, fmt.Errorf("export tombstones: %w", err)
	}

	exported := &types.GenesisState{
		NftState:      k.nftKeeper.ExportGenesis(ctx),
		ClassPolicies: classPolicies,
		Lifecycles:    lifecycles,
		Tombstones:    tombstones,
	}
	if err := types.ValidateGenesis(*exported, k.addressCodec, k.cdc); err != nil {
		return nil, fmt.Errorf("validate exported nft genesis: %w", err)
	}
	if err := k.validateGenesisModuleAccounts(*exported); err != nil {
		return nil, fmt.Errorf("validate exported nft genesis: %w", err)
	}
	if err := k.verifyExportedDerivedState(ctx, *exported); err != nil {
		return nil, err
	}
	if err := k.verifyExportedPhysicalState(ctx, *exported); err != nil {
		return nil, err
	}
	return exported, nil
}

func (k Keeper) validateGenesisModuleAccounts(data types.GenesisState) error {
	validate := func(field string, value string) error {
		if _, _, err := k.canonicalNonModuleAccount(field, value); err != nil {
			return err
		}
		return nil
	}
	for _, policy := range data.ClassPolicies {
		if err := validate("class policy creator", policy.Creator); err != nil {
			return fmt.Errorf("invalid class policy %s: %w", policy.ClassId, err)
		}
		if err := validate("class policy controller", policy.Controller); err != nil {
			return fmt.Errorf("invalid class policy %s: %w", policy.ClassId, err)
		}
	}
	for _, entry := range data.NftState.Entries {
		if err := validate("nft owner", entry.Owner); err != nil {
			return err
		}
	}
	for _, lifecycle := range data.Lifecycles {
		if err := validate("lifecycle minted_by", lifecycle.Mint.MintedBy); err != nil {
			return fmt.Errorf("invalid lifecycle %s/%s: %w", lifecycle.ClassId, lifecycle.NftId, err)
		}
		if lifecycle.Revocation != nil {
			if err := validate("lifecycle revoked_by", lifecycle.Revocation.RevokedBy); err != nil {
				return fmt.Errorf(
					"invalid lifecycle %s/%s: %w",
					lifecycle.ClassId,
					lifecycle.NftId,
					err,
				)
			}
		}
	}
	for _, tombstone := range data.Tombstones {
		if err := validate("tombstone minted_by", tombstone.Mint.MintedBy); err != nil {
			return fmt.Errorf("invalid tombstone %s/%s: %w", tombstone.ClassId, tombstone.NftId, err)
		}
		if tombstone.Revocation != nil {
			if err := validate("tombstone revoked_by", tombstone.Revocation.RevokedBy); err != nil {
				return fmt.Errorf(
					"invalid tombstone %s/%s: %w",
					tombstone.ClassId,
					tombstone.NftId,
					err,
				)
			}
		}
		if err := validate("tombstone burned_by", tombstone.BurnedBy); err != nil {
			return fmt.Errorf("invalid tombstone %s/%s: %w", tombstone.ClassId, tombstone.NftId, err)
		}
	}
	return nil
}

func (k Keeper) expectedGenesisState(data types.GenesisState) (expectedGenesisState, error) {
	expected := expectedGenesisState{
		supply:          make(map[string]uint64, len(data.NftState.Classes)),
		mintedCount:     make(map[string]uint64, len(data.NftState.Classes)),
		ownerByNFT:      make(map[genesisNFTKey]sdk.AccAddress),
		ownerClassCount: make(map[string]map[string]uint64),
	}
	for _, class := range data.NftState.Classes {
		expected.supply[class.Id] = 0
		expected.mintedCount[class.Id] = 0
	}
	for _, entry := range data.NftState.Entries {
		ownerBytes, err := k.addressCodec.StringToBytes(entry.Owner)
		if err != nil {
			return expectedGenesisState{}, fmt.Errorf("decode validated nft owner %s: %w", entry.Owner, err)
		}
		expected.ownerClassCount[entry.Owner] = make(map[string]uint64)
		for _, token := range entry.Nfts {
			key := genesisNFTKey{classID: token.ClassId, nftID: token.Id}
			expected.ownerByNFT[key] = sdk.AccAddress(append([]byte(nil), ownerBytes...))
			expected.supply[token.ClassId]++
			expected.mintedCount[token.ClassId]++
			expected.ownerClassCount[entry.Owner][token.ClassId]++
		}
	}
	for _, tombstone := range data.Tombstones {
		expected.mintedCount[tombstone.ClassId]++
	}
	return expected, nil
}

func (k Keeper) verifyInitializedGenesis(
	ctx sdk.Context,
	data types.GenesisState,
	expected expectedGenesisState,
) error {
	for _, class := range data.NftState.Classes {
		classID := class.Id
		expectedSupply := expected.supply[classID]
		if actual := k.nftKeeper.GetTotalSupply(ctx, classID); actual != expectedSupply {
			return fmt.Errorf(
				"initialized class %s supply %d does not match expected %d",
				classID,
				actual,
				expectedSupply,
			)
		}
		expectedCount := expected.mintedCount[classID]
		actual, err := k.mintedCounts.Get(ctx, classID)
		if err != nil {
			return fmt.Errorf("load initialized minted count for class %s: %w", classID, err)
		}
		if actual != expectedCount {
			return fmt.Errorf(
				"initialized class %s minted count %d does not match expected %d",
				classID,
				actual,
				expectedCount,
			)
		}
	}
	for _, entry := range data.NftState.Entries {
		ownerBytes, err := k.addressCodec.StringToBytes(entry.Owner)
		if err != nil {
			return fmt.Errorf("decode validated nft owner %s: %w", entry.Owner, err)
		}
		checkedClasses := make(map[string]struct{})
		for _, token := range entry.Nfts {
			key := genesisNFTKey{classID: token.ClassId, nftID: token.Id}
			expectedOwner := expected.ownerByNFT[key]
			actualOwner := k.nftKeeper.GetOwner(ctx, key.classID, key.nftID)
			if !bytes.Equal(actualOwner, expectedOwner) {
				return fmt.Errorf(
					"initialized nft %s/%s owner does not match input",
					key.classID,
					key.nftID,
				)
			}
			if _, checked := checkedClasses[token.ClassId]; checked {
				continue
			}
			checkedClasses[token.ClassId] = struct{}{}
			expectedCount := expected.ownerClassCount[entry.Owner][token.ClassId]
			actual := k.nftKeeper.GetBalance(ctx, token.ClassId, sdk.AccAddress(ownerBytes))
			if actual != expectedCount {
				return fmt.Errorf(
					"initialized owner %s class %s balance %d does not match expected %d",
					entry.Owner,
					token.ClassId,
					actual,
					expectedCount,
				)
			}
		}
	}
	return nil
}

func (k Keeper) verifyExportedDerivedState(ctx sdk.Context, data types.GenesisState) error {
	liveCounts := make(map[string]uint64, len(data.NftState.Classes))
	mintedCounts := make(map[string]uint64, len(data.NftState.Classes))
	policyClasses := make(map[string]struct{}, len(data.ClassPolicies))
	for _, policy := range data.ClassPolicies {
		policyClasses[policy.ClassId] = struct{}{}
	}
	for _, lifecycle := range data.Lifecycles {
		liveCounts[lifecycle.ClassId]++
		mintedCounts[lifecycle.ClassId]++
	}
	for _, tombstone := range data.Tombstones {
		mintedCounts[tombstone.ClassId]++
	}
	for _, policy := range data.ClassPolicies {
		classID := policy.ClassId
		actualMintedCount, err := k.mintedCounts.Get(ctx, classID)
		if err != nil {
			return fmt.Errorf("load minted count for exported class %s: %w", classID, err)
		}
		if actualMintedCount != mintedCounts[classID] {
			return fmt.Errorf(
				"class %s stored minted count %d does not match derived count %d",
				classID,
				actualMintedCount,
				mintedCounts[classID],
			)
		}
		if actualSupply := k.nftKeeper.GetTotalSupply(ctx, classID); actualSupply != liveCounts[classID] {
			return fmt.Errorf(
				"class %s stored supply %d does not match live lifecycle count %d",
				classID,
				actualSupply,
				liveCounts[classID],
			)
		}
	}
	if err := k.mintedCounts.Walk(ctx, nil, func(classID string, _ uint64) (bool, error) {
		if _, exists := policyClasses[classID]; !exists {
			return true, fmt.Errorf("minted count for class %s has no class policy", classID)
		}
		return false, nil
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) ensureStoresAvailable(ctx sdk.Context) error {
	if _, err := k.nftStoreService.OpenKVStore(ctx).Has(storeAvailabilityProbeKey); err != nil {
		return fmt.Errorf("open nft store: %w", err)
	}
	if _, err := k.policyStoreService.OpenKVStore(ctx).Has(storeAvailabilityProbeKey); err != nil {
		return fmt.Errorf("open panacea_nft store: %w", err)
	}
	return nil
}
