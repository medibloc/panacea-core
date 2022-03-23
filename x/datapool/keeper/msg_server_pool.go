package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/medibloc/panacea-core/v2/x/datapool/types"
)

func (m msgServer) RegisterDataValidator(goCtx context.Context, msg *types.MsgRegisterDataValidator) (*types.MsgRegisterDataValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	dataValidator, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, err
	}

	err = m.Keeper.RegisterDataValidator(ctx, dataValidator, *msg.ValidatorDetail)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterDataValidatorResponse{}, nil
}

func (m msgServer) CreatePool(goCtx context.Context, msg *types.MsgCreatePool) (*types.MsgCreatePoolResponse, error) {
	return &types.MsgCreatePoolResponse{}, nil
}

func (m msgServer) SellData(goCtx context.Context, msg *types.MsgSellData) (*types.MsgSellDataResponse, error) {
	return &types.MsgSellDataResponse{}, nil
}

func (m msgServer) BuyDataAccessNFT(goCtx context.Context, msg *types.MsgBuyDataAccessNFT) (*types.MsgBuyDataAccessNFTResponse, error) {
	return &types.MsgBuyDataAccessNFTResponse{}, nil
}

func (m msgServer) RedeemDataAccessNFT(goCtx context.Context, msg *types.MsgRedeemDataAccessNFT) (*types.MsgRedeemDataAccessNFTResponse, error) {
	return &types.MsgRedeemDataAccessNFTResponse{}, nil
}
