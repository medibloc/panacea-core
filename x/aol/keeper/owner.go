package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/compkey"
	"github.com/medibloc/panacea-core/v2/x/aol/types"
)

// SetOwner set a specific owner in the store
func (k Keeper) SetOwner(ctx sdk.Context, key types.OwnerCompositeKey, owner types.Owner) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.OwnerKeyPrefix)
	b := k.cdc.MustMarshal(&owner)
	store.Set(compkey.MustEncode(&key), b)
}

// GetOwner returns a owner from its id
func (k Keeper) GetOwner(ctx sdk.Context, key types.OwnerCompositeKey) types.Owner {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.OwnerKeyPrefix)
	var owner types.Owner
	k.cdc.MustUnmarshal(store.Get(compkey.MustEncode(&key)), &owner)
	return owner
}

// HasOwner checks if the owner exists in the store
func (k Keeper) HasOwner(ctx sdk.Context, key types.OwnerCompositeKey) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.OwnerKeyPrefix)
	return store.Has(compkey.MustEncode(&key))
}

func (k Keeper) GetAllOwners(ctx sdk.Context) ([]types.OwnerCompositeKey, []types.Owner) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.OwnerKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	keys := make([]types.OwnerCompositeKey, 0)
	values := make([]types.Owner, 0)

	for ; iterator.Valid(); iterator.Next() {
		var key types.OwnerCompositeKey
		compkey.MustDecode(iterator.Key(), &key)
		keys = append(keys, key)

		var value types.Owner
		k.cdc.MustUnmarshal(iterator.Value(), &value)
		values = append(values, value)
	}

	return keys, values
}
