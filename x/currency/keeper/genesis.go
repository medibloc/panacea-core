package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/currency/types"
)

func InitGenesis(ctx sdk.Context, k Keeper, data types.GenesisState) {
	for bz, issuance := range data.Issuances {
		var key types.GenesisIssuanceKey
		if err := key.Unmarshal(bz); err != nil {
			panic(err)
		}
		k.SetIssuance(ctx, key.Denom, issuance)
	}
}

func ExportGenesis(ctx sdk.Context, k Keeper) types.GenesisState {
	issuancesMap := make(map[string]types.Issuance)

	for _, denom := range k.ListIssuedDenoms(ctx) {
		key := types.GenesisIssuanceKey{Denom: denom}.Marshal()
		issuancesMap[key] = k.GetIssuance(ctx, denom)
	}

	return types.GenesisState{
		Issuances: issuancesMap,
	}
}
