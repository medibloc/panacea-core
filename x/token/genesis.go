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
	for _, token := range genState.Tokens {
		k.SetToken(ctx, *token)
	}

	// this line is used by starport scaffolding # ibc/genesis/init
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	// this line is used by starport scaffolding # genesis/module/export
	// Get all token
	tokenList := k.GetAllToken(ctx)
	for _, token := range tokenList {
		genesis.Tokens[token.Symbol] = &token
	}

	// this line is used by starport scaffolding # ibc/genesis/export

	return genesis
}
