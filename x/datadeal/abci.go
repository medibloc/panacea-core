package datadeal

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/keeper"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func EndBlocker(ctx sdk.Context, keeper keeper.Keeper) {
	keeper.IterateClosedDataDeliveryQueue(ctx, ctx.BlockHeader().Time, func(dataSale *types.DataSale) bool {

		keeper.RemoveDataDeliveryQueue(ctx, dataSale.DealId, dataSale.VerifiableCid, dataSale.VotingPeriod.VotingEndTime)
		iterator := keeper.GetDataDeliveryVoteIterator(ctx, dataSale.DealId, dataSale.VerifiableCid)
		defer iterator.Close()

		oracleKeeper := keeper.GetOracleKeeper()

		tallyResult, err := oracleKeeper.Tally(
			ctx,
			iterator,
			func() oracletypes.Vote {
				return &types.DataDeliveryVote{}
			},
			func(vote oracletypes.Vote) error {
				return keeper.RemoveDataDeliveryVote(ctx, vote.(*types.DataDeliveryVote))
			},
		)
		if err != nil {
			panic(err)
		}

		if tallyResult.IsPassed() {
			dataSale.Status = types.DATA_SALE_STATUS_COMPLETED
			dataSale.DeliveredCid = string(tallyResult.ConsensusValue)

		} else {
			dataSale.Status = types.DATA_SALE_STATUS_FAILED
		}

		dataSale.DeliveryTallyResult = tallyResult
		if err := keeper.SetDataSale(ctx, dataSale); err != nil {
			panic(err)
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeDataDeliveryVote,
				sdk.NewAttribute(types.AttributeKeyVoteStatus, types.AttributeValueVoteStatusEnded),
				sdk.NewAttribute(types.AttributeKeyDeliveredCID, dataSale.DeliveredCid),
			),
		)

		return false
	})
}
