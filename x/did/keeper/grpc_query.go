package keeper

import (
	"github.com/medibloc/panacea-core/v2/x/did/types"
)

var _ types.QueryServer = Keeper{}
