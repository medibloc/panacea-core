package aol

import (
	"github.com/medibloc/panacea-core/x/aol/keeper"
	"github.com/medibloc/panacea-core/x/aol/types"
)

const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	RouterKey    = types.RouterKey
	QuerierRoute = types.QuerierRoute
)

var (
	RegisterCodec = types.RegisterCodec
	ModuleCdc     = types.ModuleCdc
	NewKeeper     = keeper.NewKeeper
)

type (
	Keeper = keeper.Keeper

	MsgCreateTopic  = types.MsgCreateTopic
	MsgAddWriter    = types.MsgAddWriter
	MsgDeleteWriter = types.MsgDeleteWriter
	MsgAddRecord    = types.MsgAddRecord
)
