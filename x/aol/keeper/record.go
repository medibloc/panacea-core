package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/types/compkey"
	"github.com/medibloc/panacea-core/x/aol/types"
)

// SetRecord set a specific record in the store
func (k Keeper) SetRecord(ctx sdk.Context, key types.RecordCompositeKey, record types.Record) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RecordKey))
	b := k.cdc.MustMarshalBinaryBare(&record)
	store.Set(compkey.MustEncode(&key), b)
}

// GetRecord returns a record from its id
func (k Keeper) GetRecord(ctx sdk.Context, key types.RecordCompositeKey) types.Record {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RecordKey))
	var record types.Record
	k.cdc.MustUnmarshalBinaryBare(store.Get(compkey.MustEncode(&key)), &record)
	return record
}

// HasRecord checks if the record exists in the store
func (k Keeper) HasRecord(ctx sdk.Context, key types.RecordCompositeKey) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RecordKey))
	return store.Has(compkey.MustEncode(&key))
}

// GetAllRecords returns all records
func (k Keeper) GetAllRecords(ctx sdk.Context) ([]types.RecordCompositeKey, []types.Record) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RecordKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	keys := make([]types.RecordCompositeKey, 0)
	values := make([]types.Record, 0)

	for ; iterator.Valid(); iterator.Next() {
		var key types.RecordCompositeKey
		compkey.MustDecode(iterator.Key(), &key)
		keys = append(keys, key)

		var value types.Record
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &value)
		values = append(values, value)
	}

	return keys, values
}
