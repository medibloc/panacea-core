package aol

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/aol/types"
)

// Keeper maintains the link to data storage
// and exposes getter/setter methods for
// the various parts of the state machine
type Keeper struct {
	// Unexposed key to access store from sdk.Context
	storeKey sdk.StoreKey

	cdc *codec.Codec
}

// NewKeeper creates new instances of the aol Keeper
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return Keeper{
		storeKey: storeKey,
		cdc:      cdc,
	}
}

func (k Keeper) SetOwner(ctx sdk.Context, ownerAddr sdk.AccAddress, owner types.Owner) {
	store := ctx.KVStore(k.storeKey)
	ownerKey := OwnerKey(ownerAddr)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(owner)
	store.Set(ownerKey, bz)
}

func (k Keeper) GetOwner(ctx sdk.Context, ownerAddr sdk.AccAddress) types.Owner {
	store := ctx.KVStore(k.storeKey)
	ownerKey := OwnerKey(ownerAddr)
	bz := store.Get(ownerKey)
	if bz == nil {
		return types.Owner{}
	}
	var owner types.Owner
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &owner)
	return owner
}

func (k Keeper) ListOwner(ctx sdk.Context) []sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)
	addrs := make([]sdk.AccAddress, 0)

	ownerKey := OwnerKey(sdk.AccAddress{})
	iter := sdk.KVStorePrefixIterator(store, ownerKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		addr := bytes.Split(iter.Key(), ownerKey)[1]
		addrs = append(addrs, addr)
	}
	return addrs
}

func (k Keeper) SetTopic(ctx sdk.Context, ownerAddr sdk.AccAddress, topicName string, topic types.Topic) {
	store := ctx.KVStore(k.storeKey)
	topicKey := TopicKey(ownerAddr, topicName)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(topic)
	store.Set(topicKey, bz)
}

func (k Keeper) GetTopic(ctx sdk.Context, ownerAddr sdk.AccAddress, topicName string) types.Topic {
	store := ctx.KVStore(k.storeKey)
	topicKey := TopicKey(ownerAddr, topicName)
	bz := store.Get(topicKey)
	if bz == nil {
		return types.Topic{}
	}
	var topic types.Topic
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &topic)
	return topic
}

func (k Keeper) HasTopic(ctx sdk.Context, ownerAddr sdk.AccAddress, topicName string) bool {
	store := ctx.KVStore(k.storeKey)
	topicKey := TopicKey(ownerAddr, topicName)
	return store.Has(topicKey)
}

func (k Keeper) ListTopic(ctx sdk.Context, ownerAddr sdk.AccAddress) []string {
	store := ctx.KVStore(k.storeKey)
	topics := make([]string, 0)

	topicKey := TopicKey(ownerAddr, "")
	iter := sdk.KVStorePrefixIterator(store, topicKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		topic := string(bytes.Split(iter.Key(), []byte(ownerAddr))[1])
		topics = append(topics, topic[len(KeyDelimiter):])
	}
	return topics
}

func (k Keeper) SetWriter(ctx sdk.Context, ownerAddr sdk.AccAddress, topic string, writerAddr sdk.AccAddress, writer types.Writer) {
	store := ctx.KVStore(k.storeKey)
	writerKey := ACLWriterKey(ownerAddr, topic, writerAddr)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(writer)
	store.Set(writerKey, bz)
}

func (k Keeper) GetWriter(ctx sdk.Context, ownerAddr sdk.AccAddress, topic string, writerAddr sdk.AccAddress) types.Writer {
	store := ctx.KVStore(k.storeKey)
	writerKey := ACLWriterKey(ownerAddr, topic, writerAddr)
	bz := store.Get(writerKey)
	if bz == nil {
		return types.Writer{}
	}
	var writer types.Writer
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &writer)
	return writer
}

func (k Keeper) HasWriter(ctx sdk.Context, ownerAddr sdk.AccAddress, topic string, writerAddr sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	writerKey := ACLWriterKey(ownerAddr, topic, writerAddr)
	return store.Has(writerKey)
}

func (k Keeper) DeleteWriter(ctx sdk.Context, ownerAddr sdk.AccAddress, topic string, writerAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	writerKey := ACLWriterKey(ownerAddr, topic, writerAddr)
	store.Delete(writerKey)
}

//TODO Max Retrieve
func (k Keeper) ListWriter(ctx sdk.Context, ownerAddr sdk.AccAddress, topic string) []sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)
	writers := make([]sdk.AccAddress, 0)

	writerKey := ACLWriterKey(ownerAddr, topic, sdk.AccAddress{})
	iter := sdk.KVStorePrefixIterator(store, writerKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		writer := sdk.AccAddress{}
		data := bytes.Split(iter.Key(), []byte(topic))[1]
		data = data[len(KeyDelimiter):]
		if err := writer.Unmarshal(data); err != nil {
			return []sdk.AccAddress{}
		}
		writers = append(writers, writer)
	}
	return writers
}

func (k Keeper) SetRecord(ctx sdk.Context, ownerAddr sdk.AccAddress, topic string, offset uint64, record types.Record) {
	store := ctx.KVStore(k.storeKey)
	recordKey := RecordKey(ownerAddr, topic, offset)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(record)
	store.Set(recordKey, bz)
}

func (k Keeper) GetRecord(ctx sdk.Context, ownerAddr sdk.AccAddress, topic string, offset uint64) types.Record {
	store := ctx.KVStore(k.storeKey)
	recordKey := RecordKey(ownerAddr, topic, offset)
	bz := store.Get(recordKey)
	if bz == nil {
		return types.Record{}
	}
	var record types.Record
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &record)
	return record
}

func (k Keeper) HasRecord(ctx sdk.Context, ownerAddr sdk.AccAddress, topic string, offset uint64) bool {
	store := ctx.KVStore(k.storeKey)
	recordKey := RecordKey(ownerAddr, topic, offset)
	return store.Has(recordKey)
}
