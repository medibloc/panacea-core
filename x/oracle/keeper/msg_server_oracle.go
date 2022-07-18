package keeper

import (
	"context"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func (m msgServer) RegisterOracle(goCtx context.Context, msg *types.MsgRegisterOracle) (*types.MsgRegisterOracleResponse, error) {
	panic("implements me")
}

func (m msgServer) VoteOracleRegistration(goCtx context.Context, msg *types.MsgVoteOracleRegistration) (*types.MsgVoteOracleRegistrationResponse, error) {
	panic("implements me")
}
