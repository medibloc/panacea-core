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
		iterator := keeper.GetOracleRegistrationVoteIterator(ctx, oracleRegistration.UniqueId, oracleRegistration.Address)

		defer iterator.Close()

		tallyResult, err := keeper.Tally(
			ctx,
			iterator,
			&types.OracleRegistrationVote{},
			func(vote types.Vote) error {
				return keeper.RemoveOracleRegistrationVote(ctx, vote.(*types.OracleRegistrationVote))
			},
		)
		if err != nil {
			panic(err)
		}

		// If ConsensusValue does not exist, consensus has not been passed.
		if tallyResult.IsPassed() {
			oracleRegistration.Status = types.ORACLE_REGISTRATION_STATUS_PASSED
			oracleRegistration.EncryptedOraclePrivKey = tallyResult.ConsensusValue

			oracle := types.NewOracle(oracleRegistration.Address, types.ORACLE_STATUS_ACTIVE)
			if err := keeper.SetOracle(ctx, oracle); err != nil {
				panic(err)
			}
		} else {
			oracleRegistration.Status = types.ORACLE_REGISTRATION_STATUS_REJECTED
		}

		oracleRegistration.TallyResult = tallyResult

		if err := keeper.SetOracleRegistration(ctx, oracleRegistration); err != nil {
			panic(err)
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRegistrationVote,
				sdk.NewAttribute(types.AttributeKeyVoteStatus, types.AttributeValueVoteStatusEnded),
				sdk.NewAttribute(types.AttributeKeyOracleAddress, oracleRegistration.Address),
			),
		)

		return false
	})
}
