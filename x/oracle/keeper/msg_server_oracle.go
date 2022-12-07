package keeper

import (
	"context"

	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func (m msgServer) RegisterOracle(ctx context.Context, oracle *types.MsgRegisterOracle) (*types.MsgRegisterOracleResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m msgServer) ApproveOracleRegistration(ctx context.Context, registration *types.MsgApproveOracleRegistration) (*types.MsgApproveOracleRegistrationResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m msgServer) UpdateOracleInfo(ctx context.Context, info *types.MsgUpdateOracleInfo) (*types.MsgUpdateOracleInfoResponse, error) {
	//TODO implement me
	panic("implement me")
}
