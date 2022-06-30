package keeper

import (
	"fmt"

	"github.com/medibloc/panacea-core/v2/types/assets"

	"github.com/medibloc/panacea-core/v2/x/burn/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) BurnCoins(ctx sdk.Context, acc string) error {
	burnAcc, err := sdk.AccAddressFromBech32(acc)

	if err != nil {
		return err
	}

	burnCoins := k.bankKeeper.GetAllBalances(ctx, burnAcc)
	if burnCoins.Empty() {
		return nil
	}

	ctx.Logger().Info("find burn coin.", fmt.Sprintf("address: %s, coins: %s", acc, burnCoins))

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, burnAcc, types.ModuleName, burnCoins)
	if err != nil {
		return err
	}

	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, burnCoins)
	if err != nil {
		return err
	}

	ctx.Logger().Info("Success burn coin to burnAccount.", fmt.Sprintf("address: %s, coins: %s", acc, burnCoins))

	ctx.Logger().Info("Success burn coin to supply.", fmt.Sprintf("total: %s", k.bankKeeper.GetSupply(ctx, assets.MicroMedDenom)))

	return nil

}
