package aol

import (
	"github.com/medibloc/panacea-core/x/aol/internal/keeper"
	"github.com/medibloc/panacea-core/x/aol/internal/types"
)

const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	RouterKey    = types.RouterKey
)

var (
	NewKeeper           = keeper.NewKeeper
	NewQuerier          = keeper.NewQuerier
	NewMsgCreateTopic   = types.NewMsgCreateTopic
	NewMsgAddWriter     = types.NewMsgAddWriter
	NewMsgDeleteWriter  = types.NewMsgDeleteWriter
	NewMsgAddRecord     = types.NewMsgAddRecord
	ModuleCdc           = types.ModuleCdc
	RegisterCodec       = types.RegisterCodec
)

type (
	Keeper          = keeper.Keeper

	MsgCreateTopic  = types.MsgCreateTopic
	MsgAddWriter    = types.MsgAddWriter
	MsgDeleteWriter = types.MsgDeleteWriter
	MsgAddRecord    = types.MsgAddRecord
)
