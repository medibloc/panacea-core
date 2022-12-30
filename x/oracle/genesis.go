package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/keeper"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	for _, oracle := range genState.Oracles {
		if err := k.SetOracle(ctx, &oracle); err != nil {
			panic(err)
		}
	}

	for _, oracleRegistration := range genState.OracleRegistrations {
		if err := k.SetOracleRegistration(ctx, &oracleRegistration); err != nil {
			panic(err)
		}
	}

	k.SetParams(ctx, genState.Params)

	if genState.OracleUpgradeQueueElements != nil {
		for _, address := range genState.OracleUpgradeQueueElements {
			accAddr, err := sdk.AccAddressFromBech32(address)
			if err != nil {
				panic(err)
			}
			k.AddOracleUpgradeQueue(ctx, accAddr)
		}
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	oracles, err := k.GetAllOracleList(ctx)
	if err != nil {
		panic(err)
	}
	genesis.Oracles = oracles

	oracleRegistrations, err := k.GetAllOracleRegistrationList(ctx)
	if err != nil {
		panic(err)
	}
	genesis.OracleRegistrations = oracleRegistrations

	genesis.Params = k.GetParams(ctx)

	queueElements, err := k.GetAllOracleUpgradeQueueElements(ctx)
	if err != nil {
		panic(err)
	}
	genesis.OracleUpgradeQueueElements = queueElements

	return genesis
}
