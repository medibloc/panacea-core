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

	for _, dataPassRedeemReceipt := range genState.DataPassRedeemReceipts {
		err := k.SetDataPassRedeemReceipt(ctx, dataPassRedeemReceipt)
		if err != nil {
			panic(err)
		}
	}

	k.SetInstantRevenueDistribute(ctx, &genState.InstantRevenueDistribute)

	for _, history := range genState.SalesHistories {
		k.SetSalesHistory(ctx, history)
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

	dataPassRedeemReceipts, err := k.GetAllDataPassRedeemReceipts(ctx)
	if err != nil {
		panic(err)
	}

	genesis.DataPassRedeemReceipts = append(genesis.DataPassRedeemReceipts, dataPassRedeemReceipts...)

	genesis.InstantRevenueDistribute.PoolIds = append(
		genesis.InstantRevenueDistribute.PoolIds,
		k.GetInstantRevenueDistribute(ctx).PoolIds...,
	)

	genesis.SalesHistories = k.GetAllSalesHistories(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	// this line is used by starport scaffolding # ibc/genesis/export

	return genesis
}
