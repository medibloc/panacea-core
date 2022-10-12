package datadeal

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/keeper"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func EndBlocker(ctx sdk.Context, keeper keeper.Keeper) {
	keeper.IterateClosedDataVerificationQueue(ctx, ctx.BlockHeader().Time, func(dataSale *types.DataSale) bool {

		keeper.RemoveDataVerificationQueue(ctx, dataSale.DealId, dataSale.VerifiableCid, dataSale.VotingPeriod.VotingEndTime)
		iterator := keeper.GetDataVerificationVoteIterator(ctx, dataSale.DealId, dataSale.VerifiableCid)

		defer iterator.Close()

		oracleKeeper := keeper.GetOracleKeeper()

		tallyResult, err := oracleKeeper.Tally(
			ctx,
			iterator,
			func() oracletypes.Vote {
				return &types.DataVerificationVote{}
			},
			func(vote oracletypes.Vote) error {
				return keeper.RemoveDataVerificationVote(ctx, vote.(*types.DataVerificationVote))
			},
		)

		if err != nil {
			panic(err)
		}

		if tallyResult.IsPassed() {
			dataSale.Status = types.DATA_SALE_STATUS_DELIVERY_VOTING_PERIOD
			dataSale.VerifiableCid = string(tallyResult.ConsensusValue)

			keeper.AddDataDeliveryQueue(ctx, dataSale.VerifiableCid, dataSale.DealId, oracleKeeper.GetVotingPeriod(ctx).VotingEndTime)
		} else {
			dataSale.Status = types.DATA_SALE_STATUS_VERIFICATION_FAILED
		}

		dataSale.VerificationTallyResult = tallyResult

		if err := keeper.SetDataSale(ctx, dataSale); err != nil {
			panic(err)
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeDataVerificationVote,
				sdk.NewAttribute(types.AttributeKeyVoteStatus, types.AttributeValueVoteStatusEnded),
				sdk.NewAttribute(types.AttributeKeyVerifiableCID, dataSale.VerifiableCid),
				sdk.NewAttribute(types.AttributeKeyDealID, strconv.FormatUint(dataSale.DealId, 10)),
			),
		)

		return false
	})

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
			dataSale.Status = types.DATA_SALE_STATUS_DELIVERY_FAILED
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
				sdk.NewAttribute(types.AttributeKeyDealID, strconv.FormatUint(dataSale.DealId, 10)),
			),
		)

		return false
	})
}
