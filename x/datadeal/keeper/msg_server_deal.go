package keeper

import (
	"context"

	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

func (m msgServer) CreateDeal(ctx context.Context, deal *types.MsgCreateDeal) (*types.MsgCreateDealResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m msgServer) DeactivateDeal(ctx context.Context, deal *types.MsgDeactivateDeal) (*types.MsgDeactivateDealResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m msgServer) SubmitConsent(ctx context.Context, consent *types.MsgSubmitConsent) (*types.MsgSubmitConsentResponse, error) {
	//TODO implement me
	panic("implement me")
}
