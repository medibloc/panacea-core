package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

func (m msgServer) CreateDeal(goCtx context.Context, msg *types.MsgCreateDeal) (*types.MsgCreateDealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	newDealID, err := m.Keeper.CreateDeal(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateDealResponse{DealId: newDealID}, nil
}

func (m msgServer) DeactivateDeal(goCtx context.Context, msg *types.MsgDeactivateDeal) (*types.MsgDeactivateDealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := m.Keeper.DeactivateDeal(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgDeactivateDealResponse{}, nil
}

func (m msgServer) SubmitConsent(ctx context.Context, consent *types.MsgSubmitConsent) (*types.MsgSubmitConsentResponse, error) {
	//TODO implement me
	panic("implement me")
}
