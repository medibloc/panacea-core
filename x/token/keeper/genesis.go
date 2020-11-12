package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/token/types"
)

func InitGenesis(ctx sdk.Context, k Keeper, data types.GenesisState) {
	for bz, token := range data.Tokens {
		var key types.GenesisTokenKey
		if err := key.Unmarshal(bz); err != nil {
			panic(err)
		}
		k.SetToken(ctx, key.Symbol, token)
	}
}

func ExportGenesis(ctx sdk.Context, k Keeper) types.GenesisState {
	tokens := make(map[string]types.Token)

	for _, symbol := range k.ListTokens(ctx) {
		key := types.GenesisTokenKey{Symbol: symbol}.Marshal()
		tokens[key] = k.GetToken(ctx, symbol)
	}

	return types.GenesisState{
		Tokens: tokens,
	}
}
