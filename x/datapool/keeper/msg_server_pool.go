package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/medibloc/panacea-core/v2/x/datapool/types"
)

func (m msgServer) CreatePool(goCtx context.Context, msg *types.MsgCreatePool) (*types.MsgCreatePoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	curator, err := sdk.AccAddressFromBech32(msg.Curator)
	if err != nil {
		return nil, err
	}

	newPoolId, err := m.Keeper.CreatePool(ctx, curator, msg.Deposit, *msg.PoolParams)
	if err != nil {
		return nil, err
	}

	// TODO: return curation NFT id
	return &types.MsgCreatePoolResponse{PoolId: newPoolId, Round: 1}, nil
}

func (m msgServer) SellData(goCtx context.Context, msg *types.MsgSellData) (*types.MsgSellDataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	seller, err := sdk.AccAddressFromBech32(msg.Seller)
	if err != nil {
		return nil, err
	}

	err = m.Keeper.SellData(ctx, seller, *msg.Cert)
	if err != nil {
		return nil, err
	}

	return &types.MsgSellDataResponse{}, nil
}

func (m msgServer) BuyDataPass(goCtx context.Context, msg *types.MsgBuyDataPass) (*types.MsgBuyDataPassResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	buyer, err := sdk.AccAddressFromBech32(msg.Buyer)
	if err != nil {
		return nil, err
	}

	err = m.Keeper.BuyDataPass(ctx, buyer, msg.PoolId, msg.Round, *msg.Payment)
	if err != nil {
		return nil, err
	}

	return &types.MsgBuyDataPassResponse{PoolId: msg.PoolId, Round: msg.Round}, nil
}

func (m msgServer) RedeemDataPass(goCtx context.Context, msg *types.MsgRedeemDataPass) (*types.MsgRedeemDataPassResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	nftRedeemReceipt, err := m.Keeper.RedeemDataPass(ctx, *msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgRedeemDataPassResponse{Receipt: nftRedeemReceipt}, nil
}
