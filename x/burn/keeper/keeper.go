package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/burn/types"
)

type (
	Keeper struct {
		bankKeeper types.BankKeeperI
	}
)

func NewKeeper(bankKeeper types.BankKeeperI) *Keeper {
	return &Keeper{
		bankKeeper: bankKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("burn", fmt.Sprintf("x/%s", types.ModuleName))
}
