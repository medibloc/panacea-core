package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func (m msgServer) RegisterOracle(ctx context.Context, oracle *types.MsgRegisterOracle) (*types.MsgRegisterOracleResponse, error) {
	//TODO implement me
	panic("implement me")
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

func (m msgServer) UpdateOracleInfo(ctx context.Context, info *types.MsgUpdateOracleInfo) (*types.MsgUpdateOracleInfoResponse, error) {
	//TODO implement me
	panic("implement me")
}
