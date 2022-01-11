package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/market/types"
)

func (m msgServer) CreateDeal(goCtx context.Context, msg *types.MsgCreateDeal) (*types.MsgCreateDealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var deal = types.Deal{
		DataSchema:           msg.DataSchema,
		Budget:               msg.Budget,
		WantDataCount:        msg.WantDataCount,
		TrustedDataValidator: msg.TrustedDataValidator,
		Owner:                msg.Owner,
	}

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	newDealId, err := m.Keeper.CreateNewDeal(ctx, owner, deal)
	if err != nil {
		return nil, err
	}
	return &types.MsgCreateDealResponse{DealId: newDealId}, nil
}
