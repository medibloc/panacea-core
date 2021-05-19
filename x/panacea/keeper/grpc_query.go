package keeper

import (
	"github.com/medibloc/panacea-core/x/panacea/types"
)

var _ types.QueryServer = Keeper{}
