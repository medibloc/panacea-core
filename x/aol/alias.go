package aol

import "github.com/medibloc/panacea-core/x/aol/types"

const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	RouterKey    = types.RouterKey
	QuerierRoute = types.QuerierRoute
)

var (
	RegisterCodec = types.RegisterCodec
)

type (
	MsgCreateTopic  = types.MsgCreateTopic
	MsgAddWriter    = types.MsgAddWriter
	MsgDeleteWriter = types.MsgDeleteWriter
	MsgAddRecord    = types.MsgAddRecord
)
