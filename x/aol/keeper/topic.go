package keeper

import (
	"encoding/binary"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/aol/types"
	"strconv"
)

// GetTopicCount get the total number of topic
func (k Keeper) GetTopicCount(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TopicCountKey))
	byteKey := types.KeyPrefix(types.TopicCountKey)
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

// SetTopicCount set the total number of topic
func (k Keeper) SetTopicCount(ctx sdk.Context, count uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TopicCountKey))
	byteKey := types.KeyPrefix(types.TopicCountKey)
	bz := []byte(strconv.FormatUint(count, 10))
	store.Set(byteKey, bz)
}

// AppendTopic appends a topic in the store with a new id and update the count
func (k Keeper) AppendTopic(
	ctx sdk.Context,
	creator string,
	description string,
	totalRecords int32,
	totalWriter int32,
) uint64 {
	// Create the topic
	count := k.GetTopicCount(ctx)
	var topic = types.Topic{
		Creator:      creator,
		Id:           count,
		Description:  description,
		TotalRecords: totalRecords,
		TotalWriter:  totalWriter,
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TopicKey))
	value := k.cdc.MustMarshalBinaryBare(&topic)
	store.Set(GetTopicIDBytes(topic.Id), value)

	// Update topic count
	k.SetTopicCount(ctx, count+1)

	return count
}

// SetTopic set a specific topic in the store
func (k Keeper) SetTopic(ctx sdk.Context, topic types.Topic) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TopicKey))
	b := k.cdc.MustMarshalBinaryBare(&topic)
	store.Set(GetTopicIDBytes(topic.Id), b)
}

// GetTopic returns a topic from its id
func (k Keeper) GetTopic(ctx sdk.Context, id uint64) types.Topic {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TopicKey))
	var topic types.Topic
	k.cdc.MustUnmarshalBinaryBare(store.Get(GetTopicIDBytes(id)), &topic)
	return topic
}

// HasTopic checks if the topic exists in the store
func (k Keeper) HasTopic(ctx sdk.Context, id uint64) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TopicKey))
	return store.Has(GetTopicIDBytes(id))
}

// GetTopicOwner returns the creator of the topic
func (k Keeper) GetTopicOwner(ctx sdk.Context, id uint64) string {
	return k.GetTopic(ctx, id).Creator
}

// RemoveTopic removes a topic from the store
func (k Keeper) RemoveTopic(ctx sdk.Context, id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TopicKey))
	store.Delete(GetTopicIDBytes(id))
}

// GetAllTopic returns all topic
func (k Keeper) GetAllTopic(ctx sdk.Context) (list []types.Topic) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TopicKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Topic
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetTopicIDBytes returns the byte representation of the ID
func GetTopicIDBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return bz
}

// GetTopicIDFromBytes returns ID in uint64 format from a byte array
func GetTopicIDFromBytes(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}
