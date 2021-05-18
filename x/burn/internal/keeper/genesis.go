package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/medibloc/panacea-core/x/burn/internal/types"
)

func InitGenesis(ctx sdk.Context, k Keeper, data types.GenesisState) {

}

func ExportGenesis(ctx sdk.Context, k Keeper) types.GenesisState {
	return types.GenesisState{}
}
