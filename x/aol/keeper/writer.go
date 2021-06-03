package keeper

import (
	"encoding/binary"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/aol/types"
	"strconv"
)

// GetWriterCount get the total number of writer
func (k Keeper) GetWriterCount(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.WriterCountKey))
	byteKey := types.KeyPrefix(types.WriterCountKey)
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

// SetWriterCount set the total number of writer
func (k Keeper) SetWriterCount(ctx sdk.Context, count uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.WriterCountKey))
	byteKey := types.KeyPrefix(types.WriterCountKey)
	bz := []byte(strconv.FormatUint(count, 10))
	store.Set(byteKey, bz)
}

// AppendWriter appends a writer in the store with a new id and update the count
func (k Keeper) AppendWriter(
	ctx sdk.Context,
	creator string,
	moniker string,
	description string,
	nanoTimestamp int32,
) uint64 {
	// Create the writer
	count := k.GetWriterCount(ctx)
	var writer = types.Writer{
		Creator:       creator,
		Id:            count,
		Moniker:       moniker,
		Description:   description,
		NanoTimestamp: nanoTimestamp,
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.WriterKey))
	value := k.cdc.MustMarshalBinaryBare(&writer)
	store.Set(GetWriterIDBytes(writer.Id), value)

	// Update writer count
	k.SetWriterCount(ctx, count+1)

	return count
}

// SetWriter set a specific writer in the store
func (k Keeper) SetWriter(ctx sdk.Context, writer types.Writer) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.WriterKey))
	b := k.cdc.MustMarshalBinaryBare(&writer)
	store.Set(GetWriterIDBytes(writer.Id), b)
}

// GetWriter returns a writer from its id
func (k Keeper) GetWriter(ctx sdk.Context, id uint64) types.Writer {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.WriterKey))
	var writer types.Writer
	k.cdc.MustUnmarshalBinaryBare(store.Get(GetWriterIDBytes(id)), &writer)
	return writer
}

// HasWriter checks if the writer exists in the store
func (k Keeper) HasWriter(ctx sdk.Context, id uint64) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.WriterKey))
	return store.Has(GetWriterIDBytes(id))
}

// GetWriterOwner returns the creator of the writer
func (k Keeper) GetWriterOwner(ctx sdk.Context, id uint64) string {
	return k.GetWriter(ctx, id).Creator
}

// RemoveWriter removes a writer from the store
func (k Keeper) RemoveWriter(ctx sdk.Context, id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.WriterKey))
	store.Delete(GetWriterIDBytes(id))
}

// GetAllWriter returns all writer
func (k Keeper) GetAllWriter(ctx sdk.Context) (list []types.Writer) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.WriterKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Writer
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetWriterIDBytes returns the byte representation of the ID
func GetWriterIDBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return bz
}

// GetWriterIDFromBytes returns ID in uint64 format from a byte array
func GetWriterIDFromBytes(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}
