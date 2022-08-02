package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/keeper"
)

func EndBlocker(ctx sdk.Context, keeper keeper.Keeper) {

	keeper.GetEndOracleRegistrationVoteQueueIterator()
}
