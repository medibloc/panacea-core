package keeper

import (
	"encoding/binary"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/aol/types"
	"strconv"
)

// GetOwnerCount get the total number of owner
func (k Keeper) GetOwnerCount(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OwnerCountKey))
	byteKey := types.KeyPrefix(types.OwnerCountKey)
	bz := store.Get(byteKey)

	// Count doesn't exist: no element
	if bz == nil {
		return 0
	}

	// Parse bytes
	count, err := strconv.ParseUint(string(bz), 10, 64)
	if err != nil {
		// Panic because the count should be always formattable to iint64
		panic("cannot decode count")
	}

	return count
}

// SetOwnerCount set the total number of owner
func (k Keeper) SetOwnerCount(ctx sdk.Context, count uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OwnerCountKey))
	byteKey := types.KeyPrefix(types.OwnerCountKey)
	bz := []byte(strconv.FormatUint(count, 10))
	store.Set(byteKey, bz)
}

// AppendOwner appends a owner in the store with a new id and update the count
func (k Keeper) AppendOwner(
	ctx sdk.Context,
	creator string,
	totalTopics int32,
) uint64 {
	// Create the owner
	count := k.GetOwnerCount(ctx)
	var owner = types.Owner{
		Creator:     creator,
		Id:          count,
		TotalTopics: totalTopics,
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OwnerKey))
	value := k.cdc.MustMarshalBinaryBare(&owner)
	store.Set(GetOwnerIDBytes(owner.Id), value)

	// Update owner count
	k.SetOwnerCount(ctx, count+1)

	return count
}

// SetOwner set a specific owner in the store
func (k Keeper) SetOwner(ctx sdk.Context, owner types.Owner) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OwnerKey))
	b := k.cdc.MustMarshalBinaryBare(&owner)
	store.Set(GetOwnerIDBytes(owner.Id), b)
}

// GetOwner returns a owner from its id
func (k Keeper) GetOwner(ctx sdk.Context, id uint64) types.Owner {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OwnerKey))
	var owner types.Owner
	k.cdc.MustUnmarshalBinaryBare(store.Get(GetOwnerIDBytes(id)), &owner)
	return owner
}

// HasOwner checks if the owner exists in the store
func (k Keeper) HasOwner(ctx sdk.Context, id uint64) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OwnerKey))
	return store.Has(GetOwnerIDBytes(id))
}

// GetOwnerOwner returns the creator of the owner
func (k Keeper) GetOwnerOwner(ctx sdk.Context, id uint64) string {
	return k.GetOwner(ctx, id).Creator
}

// RemoveOwner removes a owner from the store
func (k Keeper) RemoveOwner(ctx sdk.Context, id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OwnerKey))
	store.Delete(GetOwnerIDBytes(id))
}

// GetAllOwner returns all owner
func (k Keeper) GetAllOwner(ctx sdk.Context) (list []types.Owner) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OwnerKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Owner
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetOwnerIDBytes returns the byte representation of the ID
func GetOwnerIDBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return bz
}

// GetOwnerIDFromBytes returns ID in uint64 format from a byte array
func GetOwnerIDFromBytes(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}
