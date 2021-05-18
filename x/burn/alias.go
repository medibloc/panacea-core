package burn

import (
	"github.com/medibloc/panacea-core/x/burn/internal/keeper"
	"github.com/medibloc/panacea-core/x/burn/internal/types"
)

const (
	ModuleName = types.ModuleName
)

var (
	NewKeeper = keeper.NewKeeper
)

type (
	Keeper = keeper.Keeper
)
