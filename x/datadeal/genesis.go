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
	genesis.DataVerificationQueue = dataVerificationQueues

	dataDeliveryQueues, err := k.GetAllDataDeliveryQueue(ctx)
	if err != nil {
		panic(err)
	}

	genesis.DataDeliveryQueue = dataDeliveryQueues

	return genesis
}
