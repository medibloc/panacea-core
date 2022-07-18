package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func (m msgServer) RegisterOracle(goCtx context.Context, msg *types.MsgRegisterOracle) (*types.MsgRegisterOracleResponse, error) {
	panic("implements me")
}

func (m msgServer) VoteOracleRegistration(goCtx context.Context, msg *types.MsgVoteOracleRegistration) (*types.MsgVoteOracleRegistrationResponse, error) {
	err := m.Keeper.VoteOracleRegistration(sdk.UnwrapSDKContext(goCtx), msg.SignedOracleRegistrationVote)
	if err != nil {
		return nil, err
	}

	return &types.MsgVoteOracleRegistrationResponse{}, nil
}
