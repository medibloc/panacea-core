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
	for _, dataSale := range genState.DataSales {
		if err := k.SetDataSale(ctx, &dataSale); err != nil {
			panic(err)
		}
	}

	if err := k.SetNextDealNumber(ctx, genState.NextDealNumber); err != nil {
		panic(err)
	}

	for _, dataDeliveryVote := range genState.DataDeliveryVotes {
		if err := k.SetDataDeliveryVote(ctx, &dataDeliveryVote); err != nil {
			panic(err)
		}
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

	dataDeliveryVotes, err := k.GetAllDataDeliveryVoteList(ctx)
	if err != nil {
		panic(err)
	}
	genesis.DataDeliveryVotes = dataDeliveryVotes

	return genesis
}
