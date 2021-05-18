package keeper

import (
	"github.com/medibloc/panacea-core/x/burn/internal/types"
)

type Keeper struct {
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	supplyKeeper  types.SupplyKeeper
}

func NewKeeper(accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper, supplyKeeper types.SupplyKeeper) Keeper {
	return Keeper{
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		supplyKeeper:  supplyKeeper,
	}
}
