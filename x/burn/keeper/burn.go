package keeper

import (
	"fmt"

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

	err = k.bankKeeper.SubtractCoins(ctx, burnAcc, burnCoins)
	if err != nil {
		return err
	}

	ctx.Logger().Info("Success burn coin to burnAccount.", fmt.Sprintf("address: %s, coins: %s", acc, burnCoins))

	burnCoinsFromSupply(ctx, k, burnCoins)

	return nil

}

func burnCoinsFromSupply(ctx sdk.Context, k Keeper, amt sdk.Coins) {
	supply := k.bankKeeper.GetSupply(ctx)
	supply.Deflate(amt)
	k.bankKeeper.SetSupply(ctx, supply)

	ctx.Logger().Info("Success burn coin to supply.", fmt.Sprintf("total: %s", supply.GetTotal()))
}
