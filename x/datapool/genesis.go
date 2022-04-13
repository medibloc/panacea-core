package datapool

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datapool/keeper"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	k.SetPoolNumber(ctx, genState.NextPoolNumber)

	for _, dataValidator := range genState.DataValidators {
		err := k.SetDataValidator(ctx, dataValidator)
		if err != nil {
			panic(err)
		}
	}

	for _, pool := range genState.Pools {
		k.SetPool(ctx, &pool)
	}

	for _, whiteList := range genState.WhiteList {
		addr, err := sdk.AccAddressFromBech32(whiteList.Address)
		if err != nil {
			panic(err)
		}
		k.AddToWhiteList(ctx, whiteList.PoolId, addr)
	}
	// this line is used by starport scaffolding # genesis/module/init

	// this line is used by starport scaffolding # ibc/genesis/init
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	genesis.NextPoolNumber = k.GetNextPoolNumber(ctx)

	dataValidators, err := k.GetAllDataValidators(ctx)
	if err != nil {
		panic(err)
	}

	for _, val := range dataValidators {
		genesis.DataValidators = append(genesis.DataValidators, val)
	}

	pools, err := k.GetAllPools(ctx)
	if err != nil {
		panic(err)
	}

	for _, pool := range pools {
		genesis.Pools = append(genesis.Pools, pool)
	}

	genesis.Params = k.GetParams(ctx)

	whiteLists, err := k.GetAllWhiteLists(ctx)
	if err != nil {
		panic(err)
	}

	for _, list := range whiteLists {
		genesis.WhiteList = append(genesis.WhiteList, list)
	}

	// this line is used by starport scaffolding # genesis/module/export

	// this line is used by starport scaffolding # ibc/genesis/export

	return genesis
}
