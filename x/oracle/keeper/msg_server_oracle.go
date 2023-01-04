package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func (m msgServer) RegisterOracle(goCtx context.Context, msg *types.MsgRegisterOracle) (*types.MsgRegisterOracleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.RegisterOracle(ctx, msg); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(

			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	return &types.MsgRegisterOracleResponse{}, nil
}

func (m msgServer) ApproveOracleRegistration(goCtx context.Context, msg *types.MsgApproveOracleRegistration) (*types.MsgApproveOracleRegistrationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.ApproveOracleRegistration(ctx, msg); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	return &types.MsgApproveOracleRegistrationResponse{}, nil
}

func (m msgServer) UpdateOracleInfo(goCtx context.Context, msg *types.MsgUpdateOracleInfo) (*types.MsgUpdateOracleInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.UpdateOracleInfo(ctx, msg); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	return &types.MsgUpdateOracleInfoResponse{}, nil
}

func (m msgServer) UpgradeOracle(goCtx context.Context, msg *types.MsgUpgradeOracle) (*types.MsgUpgradeOracleResponse, error) {
	if err := m.Keeper.UpgradeOracle(sdk.UnwrapSDKContext(goCtx), msg); err != nil {
		return nil, err
	}
	return &types.MsgUpgradeOracleResponse{}, nil
}

func (m msgServer) ApproveOracleUpgrade(goCtx context.Context, msg *types.MsgApproveOracleUpgrade) (*types.MsgApproveOracleUpgradeResponse, error) {
	// TODO: Implementation
	return &types.MsgApproveOracleUpgradeResponse{}, nil
}
