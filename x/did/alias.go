package did

import "github.com/medibloc/panacea-core/x/did/types"

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
	MsgCreateDID = types.MsgCreateDID
	MsgUpdateDID = types.MsgUpdateDID
)
