package did

import (
	"github.com/medibloc/panacea-core/x/did/keeper"
	"github.com/medibloc/panacea-core/x/did/types"
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

	MsgCreateDID     = types.MsgCreateDID
	MsgUpdateDID     = types.MsgUpdateDID
	MsgDeactivateDID = types.MsgDeactivateDID
)
