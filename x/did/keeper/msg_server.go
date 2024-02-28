package keeper

import "github.com/medibloc/panacea-core/v2/x/did/types"

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(keeper Keeper) types.MsgServiceServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServiceServer = msgServer{}
