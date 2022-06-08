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

	for _, pool := range genState.Pools {
		k.SetPool(ctx, &pool)
	}

	for _, dataPassRedeemReceipt := range genState.DataPassRedeemReceipts {
		k.SetDataPassRedeemReceipt(ctx, dataPassRedeemReceipt)
	}

	k.SetInstantRevenueDistribution(ctx, &genState.InstantRevenueDistribution)

	for _, history := range genState.SalesHistories {
		k.SetSalesHistory(ctx, history)
	}

	for _, history := range genState.DataPassRedeemHistories {
		k.SetDataPassRedeemHistory(ctx, history)
	}
	// this line is used by starport scaffolding # genesis/module/init

	// this line is used by starport scaffolding # ibc/genesis/init
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	genesis.NextPoolNumber = k.GetNextPoolNumber(ctx)

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

	genesis.InstantRevenueDistribution.PoolIds = append(
		genesis.InstantRevenueDistribution.PoolIds,
		k.GetInstantRevenueDistribution(ctx).PoolIds...,
	)

	genesis.SalesHistories = k.GetAllSalesHistories(ctx)

	allHistories, err := k.GetAllDataPassRedeemHistory(ctx)
	if err != nil {
		panic(err)
	}

	genesis.DataPassRedeemHistories = allHistories
	// this line is used by starport scaffolding # genesis/module/export

	// this line is used by starport scaffolding # ibc/genesis/export

	return genesis
}
