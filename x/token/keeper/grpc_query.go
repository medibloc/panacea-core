package keeper

import (
	"github.com/medibloc/panacea-core/x/token/types"
)

var _ types.QueryServer = Keeper{}
