package datadeal

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/keeper"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	for _, deal := range genState.Deals {
		if err := k.SetDeal(ctx, &deal); err != nil {
			panic(err)
		}
	}

	if err := k.SetNextDealNumber(ctx, genState.NextDealNumber); err != nil {
		panic(err)
	}

	for _, dataSale := range genState.DataSales {
		if err := k.SetDataSale(ctx, &dataSale); err != nil {
			panic(err)
		}
	}

	for _, dataVerificationVote := range genState.DataVerificationVotes {
		if err := k.SetDataVerificationVote(ctx, &dataVerificationVote); err != nil {
			panic(err)
		}
	}

	for _, dataDeliveryVote := range genState.DataDeliveryVotes {
		if err := k.SetDataDeliveryVote(ctx, &dataDeliveryVote); err != nil {
			panic(err)
		}
	}

	for _, dataVerificationQueue := range genState.DataVerificationQueue {
		k.AddDataVerificationQueue(ctx, dataVerificationQueue.DataHash, dataVerificationQueue.DealId, dataVerificationQueue.VotingEndTime)
	}

	for _, dataDeliveryQueue := range genState.DataDeliveryQueue {
		k.AddDataDeliveryQueue(ctx, dataDeliveryQueue.DataHash, dataDeliveryQueue.DealId, dataDeliveryQueue.VotingEndTime)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	deals, err := k.GetAllDeals(ctx)
	if err != nil {
		panic(err)
	}
	genesis.Deals = deals

	dataSales, err := k.GetAllDataSaleList(ctx)
	if err != nil {
		panic(err)
	}
	genesis.DataSales = dataSales

	nextDealNum, err := k.GetNextDealNumber(ctx)
	if err != nil {
		panic(err)
	}
	genesis.NextDealNumber = nextDealNum

	dataVerificationVotes, err := k.GetAllDataVerificationVoteList(ctx)
	if err != nil {
		panic(err)
	}
	genesis.DataVerificationVotes = dataVerificationVotes

	dataDeliveryVotes, err := k.GetAllDataDeliveryVoteList(ctx)
	if err != nil {
		panic(err)
	}
	genesis.DataDeliveryVotes = dataDeliveryVotes

	dataVerificationQueues, err := k.GetAllDataVerificationQueue(ctx)
	if err != nil {
		panic(err)
	}
	genesisDataVerificationQueues := make([]types.DataVerificationQueue, 0)

	for _, dataVerificationQueue := range dataVerificationQueues {
		votingEndTime, dealID, dataHash, err := types.SplitDataDeliveryQueueKey(types.GetDataVerificationQueueKey(dataVerificationQueue.DataHash, dataVerificationQueue.DealId, dataVerificationQueue.VotingEndTime))
		if err != nil {
			panic(err)
		}

		dataVerificationQueue := types.DataVerificationQueue{
			DataHash:      dataHash,
			DealId:        dealID,
			VotingEndTime: *votingEndTime,
		}

		genesisDataVerificationQueues = append(genesisDataVerificationQueues, dataVerificationQueue)
	}
	genesis.DataVerificationQueue = genesisDataVerificationQueues

	dataDeliveryQueues, err := k.GetAllDataDeliveryQueue(ctx)
	if err != nil {
		panic(err)
	}

	genesisDataDeliveryQueues := make([]types.DataDeliveryQueue, 0)

	for _, dataDeliveryQueue := range dataDeliveryQueues {
		votingEndTime, dealID, dataHash, err := types.SplitDataVerificationQueueKey(types.GetDataDeliveryQueueKey(dataDeliveryQueue.DealId, dataDeliveryQueue.DataHash, dataDeliveryQueue.VotingEndTime))
		if err != nil {
			panic(err)
		}

		dataDeliveryQueue := types.DataDeliveryQueue{
			DataHash:      dataHash,
			DealId:        dealID,
			VotingEndTime: *votingEndTime,
		}

		genesisDataDeliveryQueues = append(genesisDataDeliveryQueues, dataDeliveryQueue)
	}
	genesis.DataDeliveryQueue = genesisDataDeliveryQueues

	return genesis
}
