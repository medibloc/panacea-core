package aol

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/aol/keeper"
	"github.com/medibloc/panacea-core/x/aol/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	// Set all the owner
	for _, elem := range genState.OwnerList {
		k.SetOwner(ctx, *elem)
	}

	// Set owner count
	k.SetOwnerCount(ctx, uint64(len(genState.OwnerList)))

	// Set all the record
	for _, elem := range genState.RecordList {
		k.SetRecord(ctx, *elem)
	}

	// Set record count
	k.SetRecordCount(ctx, uint64(len(genState.RecordList)))

	// Set all the writer
	for _, elem := range genState.WriterList {
		k.SetWriter(ctx, *elem)
	}

	// Set writer count
	k.SetWriterCount(ctx, uint64(len(genState.WriterList)))

	// Set all the topic
	for _, elem := range genState.TopicList {
		k.SetTopic(ctx, *elem)
	}

	// Set topic count
	k.SetTopicCount(ctx, uint64(len(genState.TopicList)))

	// this line is used by starport scaffolding # ibc/genesis/init
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	// this line is used by starport scaffolding # genesis/module/export
	// Get all owner
	ownerList := k.GetAllOwner(ctx)
	for _, elem := range ownerList {
		elem := elem
		genesis.OwnerList = append(genesis.OwnerList, &elem)
	}

	// Get all record
	recordList := k.GetAllRecord(ctx)
	for _, elem := range recordList {
		elem := elem
		genesis.RecordList = append(genesis.RecordList, &elem)
	}

	// Get all writer
	writerList := k.GetAllWriter(ctx)
	for _, elem := range writerList {
		elem := elem
		genesis.WriterList = append(genesis.WriterList, &elem)
	}

	// Get all topic
	topicList := k.GetAllTopic(ctx)
	for _, elem := range topicList {
		elem := elem
		genesis.TopicList = append(genesis.TopicList, &elem)
	}

	// this line is used by starport scaffolding # ibc/genesis/export

	return genesis
}
