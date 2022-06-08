package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

func (m msgServer) CreateDeal(goCtx context.Context, msg *types.MsgCreateDeal) (*types.MsgCreateDealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var deal = types.Deal{
		DataSchema:     msg.DataSchema,
		Budget:         msg.Budget,
		MaxNumData:     msg.MaxNumData,
		TrustedOracles: msg.GetTrustedOracles(),
		Owner:          msg.Owner,
	}

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	newDealID, err := m.Keeper.CreateDeal(ctx, owner, deal)
	if err != nil {
		return nil, err
	}
	return &types.MsgCreateDealResponse{DealId: newDealID}, nil
}

func (m msgServer) SellData(goCtx context.Context, msg *types.MsgSellData) (*types.MsgSellDataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	seller, err := sdk.AccAddressFromBech32(msg.Seller)
	if err != nil {
		return nil, err
	}

	oracleAddr, err := sdk.AccAddressFromBech32(msg.Cert.UnsignedCert.OracleAddress)
	if err != nil {
		return nil, err
	}

	_, err = m.Keeper.VerifyDataCert(ctx, oracleAddr, *msg.Cert)
	if err != nil {
		return nil, err
	}

	reward, err := m.Keeper.SellData(ctx, seller, *msg.Cert)
	if err != nil {
		return nil, err
	}

	return &types.MsgSellDataResponse{Reward: &reward}, nil
}

func (m msgServer) DeactivateDeal(goCtx context.Context, msg *types.MsgDeactivateDeal) (*types.MsgDeactivateDealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	requester, err := sdk.AccAddressFromBech32(msg.DeactivateRequester)
	if err != nil {
		return nil, err
	}

	deactivatedDealID, err := m.Keeper.DeactivateDeal(ctx, msg.DealId, requester)
	if err != nil {
		return nil, err
	}

	return &types.MsgDeactivateDealResponse{DealId: deactivatedDealID}, nil
}
