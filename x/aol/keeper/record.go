package keeper

import (
	"encoding/binary"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/aol/types"
	"strconv"
)

// GetRecordCount get the total number of record
func (k Keeper) GetRecordCount(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RecordCountKey))
	byteKey := types.KeyPrefix(types.RecordCountKey)
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

// SetRecordCount set the total number of record
func (k Keeper) SetRecordCount(ctx sdk.Context, count uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RecordCountKey))
	byteKey := types.KeyPrefix(types.RecordCountKey)
	bz := []byte(strconv.FormatUint(count, 10))
	store.Set(byteKey, bz)
}

// AppendRecord appends a record in the store with a new id and update the count
func (k Keeper) AppendRecord(
	ctx sdk.Context,
	creator string,
	key string,
	value string,
	nanoTimestamp int32,
	writerAddress string,
) uint64 {
	// Create the record
	count := k.GetRecordCount(ctx)
	var record = types.Record{
		Creator:       creator,
		Id:            count,
		Key:           key,
		Value:         value,
		NanoTimestamp: nanoTimestamp,
		WriterAddress: writerAddress,
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RecordKey))
	value := k.cdc.MustMarshalBinaryBare(&record)
	store.Set(GetRecordIDBytes(record.Id), value)

	// Update record count
	k.SetRecordCount(ctx, count+1)

	return count
}

// SetRecord set a specific record in the store
func (k Keeper) SetRecord(ctx sdk.Context, record types.Record) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RecordKey))
	b := k.cdc.MustMarshalBinaryBare(&record)
	store.Set(GetRecordIDBytes(record.Id), b)
}

// GetRecord returns a record from its id
func (k Keeper) GetRecord(ctx sdk.Context, id uint64) types.Record {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RecordKey))
	var record types.Record
	k.cdc.MustUnmarshalBinaryBare(store.Get(GetRecordIDBytes(id)), &record)
	return record
}

// HasRecord checks if the record exists in the store
func (k Keeper) HasRecord(ctx sdk.Context, id uint64) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RecordKey))
	return store.Has(GetRecordIDBytes(id))
}

// GetRecordOwner returns the creator of the record
func (k Keeper) GetRecordOwner(ctx sdk.Context, id uint64) string {
	return k.GetRecord(ctx, id).Creator
}

// RemoveRecord removes a record from the store
func (k Keeper) RemoveRecord(ctx sdk.Context, id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RecordKey))
	store.Delete(GetRecordIDBytes(id))
}

// GetAllRecord returns all record
func (k Keeper) GetAllRecord(ctx sdk.Context) (list []types.Record) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RecordKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Record
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetRecordIDBytes returns the byte representation of the ID
func GetRecordIDBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return bz
}

// GetRecordIDFromBytes returns ID in uint64 format from a byte array
func GetRecordIDFromBytes(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}
