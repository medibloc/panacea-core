package aol

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/types/compkey"
	"github.com/medibloc/panacea-core/x/aol/keeper"
	"github.com/medibloc/panacea-core/x/aol/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	for keyStr, owner := range genState.Owners {
		var key types.OwnerCompositeKey
		compkey.MustDecodeFromString(keyStr, types.GenesisKeySeparator, &key)
		k.SetOwner(ctx, key, *owner)
	}

	for keyStr, topic := range genState.Topics {
		var key types.TopicCompositeKey
		compkey.MustDecodeFromString(keyStr, types.GenesisKeySeparator, &key)
		k.SetTopic(ctx, key, *topic)
	}

	for keyStr, writer := range genState.Writers {
		var key types.WriterCompositeKey
		compkey.MustDecodeFromString(keyStr, types.GenesisKeySeparator, &key)
		k.SetWriter(ctx, key, *writer)
	}

	for keyStr, record := range genState.Records {
		var key types.RecordCompositeKey
		compkey.MustDecodeFromString(keyStr, types.GenesisKeySeparator, &key)
		k.SetRecord(ctx, key, *record)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	ownerKeys, owners := k.GetAllOwners(ctx)
	for i, key := range ownerKeys {
		genesis.Owners[compkey.EncodeToString(&key, types.GenesisKeySeparator)] = &owners[i]
	}

	topicKeys, topics := k.GetAllTopics(ctx)
	for i, key := range topicKeys {
		genesis.Topics[compkey.EncodeToString(&key, types.GenesisKeySeparator)] = &topics[i]
	}

	writerKeys, writers := k.GetAllWriters(ctx)
	for i, key := range writerKeys {
		genesis.Writers[compkey.EncodeToString(&key, types.GenesisKeySeparator)] = &writers[i]
	}

	recordKeys, records := k.GetAllRecords(ctx)
	for i, key := range recordKeys {
		genesis.Records[compkey.EncodeToString(&key, types.GenesisKeySeparator)] = &records[i]
	}

	return genesis
}
