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
func (k Keeper) SellData(goCtx context.Context, msg *types.MsgSellData) (*types.MsgSellDataResponse, error) {
	panic("implements me")
}

// VoteDataVerification defines a method for voting data verification.
func (k Keeper) VoteDataVerification(goCtx context.Context, msg *types.MsgVoteDataVerification) (*types.MsgVoteDataVerificationResponse, error) {
	panic("implements me")
}

// VoteDataDelivery defines a method for voting data delivery.
func (k Keeper) VoteDataDelivery(goCtx context.Context, msg *types.MsgVoteDataDelivery) (*types.MsgVoteDataDeliveryResponse, error) {
	panic("implements me")
}

// DeactivateDeal defines a method for deactivating the deal.
func (k Keeper) DeactivateDeal(goCtx context.Context, msg *types.MsgDeactivateDeal) (*types.MsgDeactivateDealResponse, error) {
	panic("implements me")
}
