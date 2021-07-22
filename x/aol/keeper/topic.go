package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/compkey"
	"github.com/medibloc/panacea-core/v2/x/aol/types"
)

// SetTopic set a specific topic in the store
func (k Keeper) SetTopic(ctx sdk.Context, key types.TopicCompositeKey, topic types.Topic) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TopicKeyPrefix)
	b := k.cdc.MustMarshalBinaryBare(&topic)
	store.Set(compkey.MustEncode(&key), b)
}

// GetTopic returns a topic from its id
func (k Keeper) GetTopic(ctx sdk.Context, key types.TopicCompositeKey) types.Topic {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TopicKeyPrefix)
	var topic types.Topic
	k.cdc.MustUnmarshalBinaryBare(store.Get(compkey.MustEncode(&key)), &topic)
	return topic
}

// HasTopic checks if the topic exists in the store
func (k Keeper) HasTopic(ctx sdk.Context, key types.TopicCompositeKey) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TopicKeyPrefix)
	return store.Has(compkey.MustEncode(&key))
}

// GetAllTopics returns all topics
func (k Keeper) GetAllTopics(ctx sdk.Context) ([]types.TopicCompositeKey, []types.Topic) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TopicKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	keys := make([]types.TopicCompositeKey, 0)
	values := make([]types.Topic, 0)

	for ; iterator.Valid(); iterator.Next() {
		var key types.TopicCompositeKey
		compkey.MustDecode(iterator.Key(), &key)
		keys = append(keys, key)

		var value types.Topic
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &value)
		values = append(values, value)
	}

	return keys, values
}
