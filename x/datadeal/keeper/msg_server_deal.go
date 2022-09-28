package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

// CreateDeal defines a method for creating a deal.
func (k Keeper) CreateDeal(goCtx context.Context, msg *types.MsgCreateDeal) (*types.MsgCreateDealResponse, error) {
	panic("implements me")
}

// SellData defines a method for selling a data.
func (m msgServer) SellData(goCtx context.Context, msg *types.MsgSellData) (*types.MsgSellDataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.SellData(ctx, msg); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	return &types.MsgSellDataResponse{}, nil
}

// VoteDataVerification defines a method for voting data verification.
func (m msgServer) VoteDataVerification(goCtx context.Context, msg *types.MsgVoteDataVerification) (*types.MsgVoteDataVerificationResponse, error) {
	panic("implements me")
}

// VoteDataDelivery defines a method for voting data delivery.
func (m msgServer) VoteDataDelivery(goCtx context.Context, msg *types.MsgVoteDataDelivery) (*types.MsgVoteDataDeliveryResponse, error) {
	panic("implements me")
}

// DeactivateDeal defines a method for deactivating the deal.
func (m msgServer) DeactivateDeal(goCtx context.Context, msg *types.MsgDeactivateDeal) (*types.MsgDeactivateDealResponse, error) {
	panic("implements me")
}
