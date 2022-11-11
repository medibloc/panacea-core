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

	for _, dataVerificationQueueElement := range genState.DataVerificationQueueElements {
		k.AddDataVerificationQueue(ctx, dataVerificationQueueElement.DataHash, dataVerificationQueueElement.DealId, dataVerificationQueueElement.VotingEndTime)
	}

	for _, dataDeliveryQueueElement := range genState.DataDeliveryQueueElements {
		k.AddDataDeliveryQueue(ctx, dataDeliveryQueueElement.DataHash, dataDeliveryQueueElement.DealId, dataDeliveryQueueElement.VotingEndTime)
	}

	for _, dealQueueElement := range genState.DealQueueElements {
		k.AddDealQueue(ctx, dealQueueElement.DealId, dealQueueElement.DeactivationHeight)
	}

	k.SetParams(ctx, genState.Params)
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

	dataVerificationQueueElements, err := k.GetAllDataVerificationQueueElements(ctx)
	if err != nil {
		panic(err)
	}
	genesis.DataVerificationQueueElements = dataVerificationQueueElements

	dataDeliveryQueuesElements, err := k.GetAllDataDeliveryQueueElements(ctx)
	if err != nil {
		panic(err)
	}
	genesis.DataDeliveryQueueElements = dataDeliveryQueuesElements

	dealQueueElements, err := k.GetAllDealQueueElements(ctx)
	if err != nil {
		panic(err)
	}
	genesis.DealQueueElements = dealQueueElements

	genesis.Params = k.GetParams(ctx)

	return genesis
}
