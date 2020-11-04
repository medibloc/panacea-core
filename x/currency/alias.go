package currency

import (
	"github.com/medibloc/panacea-core/x/currency/keeper"
	"github.com/medibloc/panacea-core/x/currency/types"
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

	MsgIssueCurrency = types.MsgIssueCurrency
)
