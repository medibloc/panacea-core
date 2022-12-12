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

func (m msgServer) ApproveOracleRegistration(ctx context.Context, registration *types.MsgApproveOracleRegistration) (*types.MsgApproveOracleRegistrationResponse, error) {
	//TODO implement me
	panic("implement me")
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
