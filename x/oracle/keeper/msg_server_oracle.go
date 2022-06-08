package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func (m msgServer) RegisterOracle(goCtx context.Context, msg *types.MsgRegisterOracle) (*types.MsgRegisterOracleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := sdk.AccAddressFromBech32(msg.OracleDetail.Address)
	if err != nil {
		return nil, err
	}

	err = m.Keeper.RegisterOracle(ctx, *msg.OracleDetail)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterOracleResponse{}, nil
}

func (m msgServer) UpdateOracle(goCtx context.Context, msg *types.MsgUpdateOracle) (*types.MsgUpdateOracleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(msg.Oracle)
	if err != nil {
		return nil, err
	}

	err = m.Keeper.UpdateOracle(ctx, address, msg.Endpoint)
	if err != nil {
		return nil, err
	}
	return &types.MsgUpdateOracleResponse{}, nil
}
