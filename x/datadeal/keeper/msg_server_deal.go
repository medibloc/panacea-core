package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

func (m msgServer) CreateDeal(goCtx context.Context, msg *types.MsgCreateDeal) (*types.MsgCreateDealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	consumer, err := sdk.AccAddressFromBech32(msg.ConsumerAddress)
	if err != nil {
		return nil, err
	}
	newDealID, err := m.Keeper.CreateDeal(ctx, consumer, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateDealResponse{DealId: newDealID}, nil
}

func (m msgServer) DeactivateDeal(ctx context.Context, deal *types.MsgDeactivateDeal) (*types.MsgDeactivateDealResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m msgServer) SubmitConsent(ctx context.Context, consent *types.MsgSubmitConsent) (*types.MsgSubmitConsentResponse, error) {
	//TODO implement me
	panic("implement me")
}
