package datadeal

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/keeper"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	if err := k.SetNextDealNumber(ctx, genState.NextDealNumber); err != nil {
		panic(err)
	}

	for _, deal := range genState.Deals {
		if err := k.SetDeal(ctx, deal); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {

	deals, err := k.GetAllDeals(ctx)
	if err != nil {
		panic(err)
	}

	nextDealNum, err := k.GetNextDealNumber(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		Deals:          deals,
		NextDealNumber: nextDealNum,
	}
}
