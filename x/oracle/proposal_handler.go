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
	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: p.Plan.UniqueId,
		Height:   p.Plan.Height,
	}
	if err := k.SetOracleUpgradeInfo(ctx, upgradeInfo); err != nil {
		return err
	}

	return nil
}
