package token

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/token/keeper"
	"github.com/medibloc/panacea-core/x/token/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	// Set all the token
	for _, elem := range genState.TokenList {
		k.SetToken(ctx, *elem)
	}

	// Set token count
	k.SetTokenCount(ctx, uint64(len(genState.TokenList)))

	// this line is used by starport scaffolding # ibc/genesis/init
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	// this line is used by starport scaffolding # genesis/module/export
	// Get all token
	tokenList := k.GetAllToken(ctx)
	for _, elem := range tokenList {
		elem := elem
		genesis.TokenList = append(genesis.TokenList, &elem)
	}

	// this line is used by starport scaffolding # ibc/genesis/export

	return genesis
}
