package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/medibloc/panacea-core/v2/x/datapool/types"
)

func (m msgServer) RegisterDataValidator(goCtx context.Context, msg *types.MsgRegisterDataValidator) (*types.MsgRegisterDataValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := sdk.AccAddressFromBech32(msg.ValidatorDetail.Address)
	if err != nil {
		return nil, err
	}

	err = m.Keeper.RegisterDataValidator(ctx, *msg.ValidatorDetail)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterDataValidatorResponse{}, nil
}

func (m msgServer) UpdateDataValidator(goCtx context.Context, msg *types.MsgUpdateDataValidator) (*types.MsgUpdateDataValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(msg.DataValidator)
	if err != nil {
		return nil, err
	}

	err = m.Keeper.UpdateDataValidator(ctx, address, msg.Endpoint)
	if err != nil {
		return nil, err
	}
	return &types.MsgUpdateDataValidatorResponse{}, nil
}

func (m msgServer) CreatePool(goCtx context.Context, msg *types.MsgCreatePool) (*types.MsgCreatePoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	curator, err := sdk.AccAddressFromBech32(msg.Curator)
	if err != nil {
		return nil, err
	}

	newPoolId, err := m.Keeper.CreatePool(ctx, curator, *msg.PoolParams)
	if err != nil {
		return nil, err
	}

	// TODO: return curation NFT id
	return &types.MsgCreatePoolResponse{PoolId: newPoolId, Round: 1}, nil
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

func (m msgServer) RegisterNFTContract(goCtx context.Context, msg *types.MsgRegisterNFTContract) (*types.MsgRegisterNFTContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return nil, err
	}

	if err := m.Keeper.DeployAndRegisterNFTContract(ctx, msg.WasmCode); err != nil {
		return nil, err
	}

	return &types.MsgRegisterNFTContractResponse{}, nil
}

func (m msgServer) UpgradeNFTContract(goCtx context.Context, msg *types.MsgUpgradeNFTContract) (*types.MsgUpgradeNFTContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return nil, err
	}

	if err := m.Keeper.MigrateNFTContract(ctx, msg.NewWasmCode); err != nil {
		return nil, err
	}

	return &types.MsgUpgradeNFTContractResponse{}, nil
}
