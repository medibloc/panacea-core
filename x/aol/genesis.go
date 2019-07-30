package aol

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/aol/types"
)

type GenesisState struct {
	Owners  map[string]types.Owner  `json:"owners"`
	Topics  map[string]types.Topic  `json:"topics"`
	Writers map[string]types.Writer `json:"writers"`
	Records map[string]types.Record `json:"records"`
}

func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	for bz, value := range data.Owners {
		var key GenesisOwnerKey
		err := key.Unmarshal(bz)
		if err != nil {
			panic(err)
		}
		k.SetOwner(ctx, key.OwnerAddress, value)
	}

	for bz, value := range data.Topics {
		var key GenesisTopicKey
		err := key.Unmarshal(bz)
		if err != nil {
			panic(err)
		}
		k.SetTopic(ctx, key.OwnerAddress, key.TopicName, value)
	}

	for bz, value := range data.Writers {
		var key GenesisWriterKey
		err := key.Unmarshal(bz)
		if err != nil {
			panic(err)
		}
		k.SetWriter(ctx, key.OwnerAddress, key.TopicName, key.WriterAddress, value)
	}

	for bz, value := range data.Records {
		var key GenesisRecordKey
		err := key.Unmarshal(bz)
		if err != nil {
			panic(err)
		}
		k.SetRecord(ctx, key.OwnerAddress, key.TopicName, key.Offset, value)
	}
}

func ValidateGenesis(data GenesisState) error {
	for bz := range data.Owners {
		var key GenesisOwnerKey
		err := key.Unmarshal(bz)
		if err != nil {
			return err
		}
	}
	for bz := range data.Topics {
		var key GenesisTopicKey
		err := key.Unmarshal(bz)
		if err != nil {
			return err
		}
	}
	for bz := range data.Writers {
		var key GenesisWriterKey
		err := key.Unmarshal(bz)
		if err != nil {
			return err
		}
	}
	for bz := range data.Records {
		var key GenesisRecordKey
		err := key.Unmarshal(bz)
		if err != nil {
			return err
		}
	}
	return nil
}

func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	ownersMap := make(map[string]types.Owner)
	topicsMap := make(map[string]types.Topic)
	writersMap := make(map[string]types.Writer)
	recordsMap := make(map[string]types.Record)

	ownerAddrs := k.ListOwner(ctx)
	for _, ownerAddr := range ownerAddrs {
		key := GenesisOwnerKey{
			OwnerAddress: ownerAddr,
		}.Marshal()
		value := k.GetOwner(ctx, ownerAddr)
		ownersMap[key] = value

		topicNames := k.ListTopic(ctx, ownerAddr)
		for _, topicName := range topicNames {
			key := GenesisTopicKey{
				OwnerAddress: ownerAddr,
				TopicName:    topicName,
			}.Marshal()
			value := k.GetTopic(ctx, ownerAddr, topicName)
			totalRecords := value.TotalRecords
			topicsMap[key] = value

			writerAddrs := k.ListWriter(ctx, ownerAddr, topicName)
			for _, writerAddr := range writerAddrs {
				key := GenesisWriterKey{
					OwnerAddress:  ownerAddr,
					TopicName:     topicName,
					WriterAddress: writerAddr,
				}.Marshal()
				value := k.GetWriter(ctx, ownerAddr, topicName, writerAddr)
				writersMap[key] = value
			}

			var offset uint64
			for offset = 0; offset < totalRecords; offset++ {
				key := GenesisRecordKey{
					OwnerAddress: ownerAddr,
					TopicName:    topicName,
					Offset:       offset,
				}.Marshal()
				value := k.GetRecord(ctx, ownerAddr, topicName, offset)
				recordsMap[key] = value
			}
		}
	}

	return GenesisState{
		Owners:  ownersMap,
		Topics:  topicsMap,
		Writers: writersMap,
		Records: recordsMap,
	}
}
