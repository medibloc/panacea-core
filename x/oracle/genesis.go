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
		err := k.SetOracle(ctx, oracle)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	oracles, err := k.GetAllOracles(ctx)
	if err != nil {
		panic(err)
	}

	oracleMap := make(map[string]types.Oracle)
	for _, oracle := range oracles {
		oracleAddr, err := sdk.AccAddressFromBech32(oracle.GetAddress())
		if err != nil {
			panic(err)
		}

		oracleKey := types.GetKeyPrefixOracle(oracleAddr)
		oracleMap[string(oracleKey)] = oracle
	}

	return &types.GenesisState{
		Oracles: oracleMap,
	}
}
