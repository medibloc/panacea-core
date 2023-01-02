package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/keeper"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func NewOracleProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.OracleUpgradeProposal:
			return handlerOracleUpgradeProposal(ctx, k, c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized oracle upgrade proposal content type: %T", c)
		}

	}
}

func handlerOracleUpgradeProposal(ctx sdk.Context, k keeper.Keeper, p *types.OracleUpgradeProposal) error {
	if p.Plan.Height < ctx.BlockHeight() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "oracle upgrade cannot be scheduled in the past")
	}

	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: p.Plan.UniqueId,
		Height:   p.Plan.Height,
	}
	if err := k.SetOracleUpgradeInfo(ctx, upgradeInfo); err != nil {
		return err
	}

	// clear OracleUpgradeQueue
	iterator := k.GetOracleUpgradeQueueIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		accAddr := sdk.AccAddress(iterator.Value())
		k.RemoveOracleUpgradeQueue(ctx, accAddr)
	}

	return nil
}
