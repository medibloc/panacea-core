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

	for _, oracleRegistrationVote := range genState.OracleRegistrationVotes {
		if err := k.SetOracleRegistrationVote(ctx, &oracleRegistrationVote); err != nil {
			panic(err)
		}
	}

	k.SetParams(ctx, genState.Params)

	// TODO implements SetUpgradeInfo

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

	oracleRegistrationVotes, err := k.GetAllOracleRegistrationVoteList(ctx)
	if err != nil {
		panic(err)
	}
	genesis.OracleRegistrationVotes = oracleRegistrationVotes

	genesis.Params = k.GetParams(ctx)

	// TODO implements SetUpgradeInfo

	return genesis
}
