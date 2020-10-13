package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/aol/types"
)

func InitGenesis(ctx sdk.Context, k Keeper, data types.GenesisState) {
	for bz, value := range data.Owners {
		var key types.GenesisOwnerKey
		err := key.Unmarshal(bz)
		if err != nil {
			panic(err)
		}
		k.SetOwner(ctx, key.OwnerAddress, value)
	}

	for bz, value := range data.Topics {
		var key types.GenesisTopicKey
		err := key.Unmarshal(bz)
		if err != nil {
			panic(err)
		}
		k.SetTopic(ctx, key.OwnerAddress, key.TopicName, value)
	}

	for bz, value := range data.Writers {
		var key types.GenesisWriterKey
		err := key.Unmarshal(bz)
		if err != nil {
			panic(err)
		}
		k.SetWriter(ctx, key.OwnerAddress, key.TopicName, key.WriterAddress, value)
	}

	for bz, value := range data.Records {
		var key types.GenesisRecordKey
		err := key.Unmarshal(bz)
		if err != nil {
			panic(err)
		}
		k.SetRecord(ctx, key.OwnerAddress, key.TopicName, key.Offset, value)
	}
}

func ExportGenesis(ctx sdk.Context, k Keeper) types.GenesisState {
	ownersMap := make(map[string]types.Owner)
	topicsMap := make(map[string]types.Topic)
	writersMap := make(map[string]types.Writer)
	recordsMap := make(map[string]types.Record)

	ownerAddrs := k.ListOwner(ctx)
	for _, ownerAddr := range ownerAddrs {
		key := types.GenesisOwnerKey{
			OwnerAddress: ownerAddr,
		}.Marshal()
		value := k.GetOwner(ctx, ownerAddr)
		ownersMap[key] = value

		topicNames := k.ListTopic(ctx, ownerAddr)
		for _, topicName := range topicNames {
			key := types.GenesisTopicKey{
				OwnerAddress: ownerAddr,
				TopicName:    topicName,
			}.Marshal()
			value := k.GetTopic(ctx, ownerAddr, topicName)
			totalRecords := value.TotalRecords
			topicsMap[key] = value

			writerAddrs := k.ListWriter(ctx, ownerAddr, topicName)
			for _, writerAddr := range writerAddrs {
				key := types.GenesisWriterKey{
					OwnerAddress:  ownerAddr,
					TopicName:     topicName,
					WriterAddress: writerAddr,
				}.Marshal()
				value := k.GetWriter(ctx, ownerAddr, topicName, writerAddr)
				writersMap[key] = value
			}

			var offset uint64
			for offset = 0; offset < totalRecords; offset++ {
				key := types.GenesisRecordKey{
					OwnerAddress: ownerAddr,
					TopicName:    topicName,
					Offset:       offset,
				}.Marshal()
				value := k.GetRecord(ctx, ownerAddr, topicName, offset)
				recordsMap[key] = value
			}
		}
	}

	return types.GenesisState{
		Owners:  ownersMap,
		Topics:  topicsMap,
		Writers: writersMap,
		Records: recordsMap,
	}
}
