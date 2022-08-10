package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/keeper"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func EndBlocker(ctx sdk.Context, keeper keeper.Keeper) {

	// Iterate through the closed OracleRegistration.
	keeper.IterateClosedOracleRegistrationQueue(ctx, ctx.BlockHeader().Time, func(oracleRegistration *types.OracleRegistration) bool {
		// TODO When a particular OracleRegistration fails, we need to consider whether to skip this OracleRegistration or fail all of them.

		// Remove the closed oracleRegistration from the queue.
		keeper.RemoveOracleRegistrationQueue(ctx, oracleRegistration.UniqueId, oracleRegistration.MustGetOracleAccAddress(), oracleRegistration.VotingPeriod.VotingEndTime)

		// Iterate the OracleRegistrationVote corresponding to the OracleRegistration and calculate the TallyResult.
		keeper.IterateOracleRegistrationVote(ctx, oracleRegistration.UniqueId, oracleRegistration.Address, func(vote *types.OracleRegistrationVote) bool {
			iterator := keeper.GetOracleRegistrationVoteIterator(ctx, oracleRegistration.UniqueId, oracleRegistration.Address)

			defer iterator.Close()

			tallyResult, err := keeper.Tally(ctx, iterator, &types.OracleRegistrationVote{})
			if err != nil {
				panic("")
			}

			// If ConsensusValue does not exist, consensus has not been passed.
			if tallyResult.ConsensusValue != nil {
				oracleRegistration.Status = types.ORACLE_REGISTRATION_STATUS_PASSED
				oracleRegistration.EncryptedOraclePrivKey = tallyResult.ConsensusValue
			} else {
				oracleRegistration.Status = types.ORACLE_REGISTRATION_STATUS_REJECTED
			}

			oracleRegistration.TallyResult = tallyResult

			err = keeper.SetOracleRegistration(ctx, oracleRegistration)
			if err != nil {
				panic("")
			}

			return false
		})
		return false
	})
}
