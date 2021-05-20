package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	accountExported "github.com/cosmos/cosmos-sdk/x/auth/exported"
)

func (k Keeper) BurnCoins(ctx sdk.Context, acc string) error {
	burnAccount, err := getAccount(ctx, k, acc)
	if err != nil {
		return err
	}

	if burnAccount == nil {
		return nil
	}

	amt := burnAccount.GetCoins()
	if amt.Empty() {
		return nil
	}

	ctx.Logger().Info("find burn coin.", fmt.Sprintf("address: %s, coins: %s", burnAccount.GetAddress(), burnAccount.GetCoins()))

	_, err = k.bankKeeper.SubtractCoins(ctx, burnAccount.GetAddress(), amt)
	if err != nil {
		return err
	}

	ctx.Logger().Info("Success burn coin to burnAccount.", fmt.Sprintf("address: %s, coins: %s", acc, amt))

	burnCoinsFromSupply(ctx, k, amt)

	return nil
}

func getAccount(ctx sdk.Context, k Keeper, acc string) (accountExported.Account, error) {
	burnAcc, err := sdk.AccAddressFromBech32(acc)
	if err != nil {
		return nil, err
	}

	return k.accountKeeper.GetAccount(ctx, burnAcc), nil
}

func burnCoinsFromSupply(ctx sdk.Context, k Keeper, amt sdk.Coins) {
	supply := k.supplyKeeper.GetSupply(ctx)
	supply = supply.Deflate(amt)
	k.supplyKeeper.SetSupply(ctx, supply)

	ctx.Logger().Info("Success burn coin to supply.", fmt.Sprintf("total: %s", supply.GetTotal()))
}
