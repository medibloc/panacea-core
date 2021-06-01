package keeper

import (
	"github.com/medibloc/panacea-core/x/did/types"
)

var _ types.QueryServer = Keeper{}
