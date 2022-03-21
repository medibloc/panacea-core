package keeper

import (
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
)

var _ types.QueryServer = Keeper{}
