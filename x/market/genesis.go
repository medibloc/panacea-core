package market

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/market/keeper"
	"github.com/medibloc/panacea-core/v2/x/market/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetNextDealNumber(ctx, genState.NextDealNumber)

	for _, deal := range genState.Deals {
		k.SetDeal(ctx, *deal)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	//genesis := types.DefaultGenesis()

	dealsMap := make(map[uint64]*types.Deal)

	//TODO: Implement GetDeals and Mapping to dealsMap
	//for _, deal := k.GetDeals(ctx) {
	//}

	return &types.GenesisState{
		Deals:          dealsMap,
		NextDealNumber: k.GetNextDealNumberAndIncrement(ctx),
	}
}
