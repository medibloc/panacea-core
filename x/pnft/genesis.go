package pnft

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/pnft/keeper"
	"github.com/medibloc/panacea-core/v2/x/pnft/types"
)

func InitGenesis(ctx sdk.Context, k *keeper.Keeper, genState types.GenesisState) {
	for _, denom := range genState.Denoms {
		if err := k.SaveDenom(ctx, denom); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the pnft module's exported genesis.
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	denoms, err := k.GetAllDenoms(ctx)
	if err != nil {
		panic(err)
	}

	genesis.Denoms = denoms

	return genesis
}
