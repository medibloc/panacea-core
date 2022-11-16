package datadeal

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/keeper"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func EndBlocker(ctx sdk.Context, keeper keeper.Keeper) {
	handleDealDeactivate(ctx, keeper)

	handleDataVerificationVote(ctx, keeper)

	handleDataDeliveryVote(ctx, keeper)
}

func handleDealDeactivate(ctx sdk.Context, keeper keeper.Keeper) {
	keeper.IteratedClosedDealQueue(ctx, ctx.BlockHeader().Height, func(deal *types.Deal) bool {
		keeper.RemoveDealQueue(ctx, deal.Id, ctx.BlockHeader().Height)
		err := keeper.DeactivateDeal(ctx, deal.Id)
		if err != nil {
			panic(err)
		}
		return false
	})
}

func handleDataVerificationVote(ctx sdk.Context, keeper keeper.Keeper) {
	keeper.IterateClosedDataVerificationQueue(ctx, ctx.BlockHeader().Time, func(dataSale *types.DataSale) bool {

		keeper.RemoveDataVerificationQueue(ctx, dataSale.DealId, dataSale.DataHash, dataSale.VerificationVotingPeriod.VotingEndTime)
		iterator := keeper.GetDataVerificationVoteIterator(ctx, dataSale.DealId, dataSale.DataHash)

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
			isDealCompleted, err := keeper.IsDealCompleted(ctx, dataSale.DealId)
			if err != nil {
				panic(err)
			}

			if isDealCompleted {
				dataSale.Status = types.DATA_SALE_STATUS_DEAL_COMPLETED
			} else {
				if err = keeper.IncrementCurNumDataAtDeal(ctx, dataSale.DealId); err != nil {
					panic(err)
				}

				if err := keeper.DistributeVerificationRewards(ctx, dataSale, tallyResult.ValidVoters); err != nil {
					panic(err)
				}

				dataSale.DeliveryVotingPeriod = oracleKeeper.GetVotingPeriod(ctx)
				dataSale.Status = types.DATA_SALE_STATUS_DELIVERY_VOTING_PERIOD
				keeper.AddDataDeliveryQueue(ctx, dataSale.DataHash, dataSale.DealId, oracleKeeper.GetVotingPeriod(ctx).VotingEndTime)

				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						types.EventTypeDataDeliveryVote,
						sdk.NewAttribute(types.AttributeKeyVoteStatus, types.AttributeValueVoteStatusStarted),
						sdk.NewAttribute(types.AttributeKeyDataHash, dataSale.DataHash),
						sdk.NewAttribute(types.AttributeKeyDealID, strconv.FormatUint(dataSale.DealId, 10))),
				)
			}
		} else {
			dataSale.Status = types.DATA_SALE_STATUS_VERIFICATION_FAILED
		}

		dataSale.VerificationTallyResult = tallyResult

		if err := keeper.SetDataSale(ctx, dataSale); err != nil {
			panic(err)
		}

		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeDataVerificationVote,
				sdk.NewAttribute(types.AttributeKeyVoteStatus, types.AttributeValueVoteStatusEnded),
				sdk.NewAttribute(types.AttributeKeyDataHash, dataSale.DataHash),
				sdk.NewAttribute(types.AttributeKeyDealID, strconv.FormatUint(dataSale.DealId, 10)),
			),
		})

		return false
	})
}

func handleDataDeliveryVote(ctx sdk.Context, keeper keeper.Keeper) {
	keeper.IterateClosedDataDeliveryQueue(ctx, ctx.BlockHeader().Time, func(dataSale *types.DataSale) bool {

		keeper.RemoveDataDeliveryQueue(ctx, dataSale.DealId, dataSale.DataHash, dataSale.DeliveryVotingPeriod.VotingEndTime)
		iterator := keeper.GetDataDeliveryVoteIterator(ctx, dataSale.DealId, dataSale.DataHash)
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
			if err := keeper.DistributeDeliveryRewards(ctx, dataSale, tallyResult.ValidVoters); err != nil {
				panic(err)
			}

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
