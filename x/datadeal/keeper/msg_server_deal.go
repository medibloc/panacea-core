package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

// CreateDeal defines a method for creating a deal.
func (m msgServer) CreateDeal(goCtx context.Context, msg *types.MsgCreateDeal) (*types.MsgCreateDealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	buyer, err := sdk.AccAddressFromBech32(msg.BuyerAddress)
	if err != nil {
		return nil, err
	}
	newDealID, err := m.Keeper.CreateDeal(ctx, buyer, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateDealResponse{DealId: newDealID}, nil

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
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := m.Keeper.VoteDataVerification(ctx, msg.DataVerificationVote, msg.Signature)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	return &types.MsgVoteDataVerificationResponse{}, nil
}

// VoteDataDelivery defines a method for voting data delivery.
func (m msgServer) VoteDataDelivery(goCtx context.Context, msg *types.MsgVoteDataDelivery) (*types.MsgVoteDataDeliveryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := m.Keeper.VoteDataDelivery(ctx, msg.DataDeliveryVote, msg.Signature)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	return &types.MsgVoteDataDeliveryResponse{}, nil
}

// DeactivateDeal defines a method for deactivating the deal.
func (m msgServer) DeactivateDeal(goCtx context.Context, msg *types.MsgDeactivateDeal) (*types.MsgDeactivateDealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := m.Keeper.DeactivateDeal(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgDeactivateDealResponse{}, nil
}
