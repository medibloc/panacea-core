package keeper

import (
	"github.com/medibloc/panacea-core/v2/x/token/types"
)

var _ types.QueryServer = Keeper{}
