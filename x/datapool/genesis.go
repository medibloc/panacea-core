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

	for _, delayedNftTransfer := range genState.DelayedNftTransfer {
		addr, err := sdk.AccAddressFromBech32(delayedNftTransfer.Address)
		if err != nil {
			panic(err)
		}
		k.AddToDelayedNftTransfer(ctx, delayedNftTransfer.PoolId, addr)
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

	genesis.DataValidators = append(genesis.DataValidators, dataValidators...)

	pools, err := k.GetAllPools(ctx)
	if err != nil {
		panic(err)
	}

	genesis.Pools = append(genesis.Pools, pools...)

	genesis.Params = k.GetParams(ctx)

	delayedNftTransfers, err := k.GetAllDelayedNftTransfers(ctx)
	if err != nil {
		panic(err)
	}

	genesis.DelayedNftTransfer = append(genesis.DelayedNftTransfer, delayedNftTransfers...)

	// this line is used by starport scaffolding # genesis/module/export

	// this line is used by starport scaffolding # ibc/genesis/export

	return genesis
}
