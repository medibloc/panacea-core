package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/burn/expected"
	"github.com/medibloc/panacea-core/x/burn/types"
)

type (
	Keeper struct {
		bankKeeper expected.BankKeeperI
	}
)

func NewKeeper(bankKeeper expected.BankKeeperI) *Keeper {
	return &Keeper{
		bankKeeper: bankKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("burn", fmt.Sprintf("x/%s", types.ModuleName))
}
