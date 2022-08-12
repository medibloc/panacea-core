package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func (m msgServer) RegisterOracle(goCtx context.Context, msg *types.MsgRegisterOracle) (*types.MsgRegisterOracleResponse, error) {
	err := m.Keeper.RegisterOracle(sdk.UnwrapSDKContext(goCtx), msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterOracleResponse{}, nil
}

func (m msgServer) VoteOracleRegistration(goCtx context.Context, msg *types.MsgVoteOracleRegistration) (*types.MsgVoteOracleRegistrationResponse, error) {
	err := m.Keeper.VoteOracleRegistration(sdk.UnwrapSDKContext(goCtx), msg.OracleRegistrationVote, msg.Signature)
	if err != nil {
		return nil, err
	}

	return &types.MsgVoteOracleRegistrationResponse{}, nil
}
