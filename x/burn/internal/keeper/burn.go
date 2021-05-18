package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	accountExported "github.com/cosmos/cosmos-sdk/x/auth/exported"
)

func (k Keeper) BurnCoins(ctx sdk.Context, acc string) sdk.Error {
	bunAccount, err := getAccount(ctx, k, acc)
	if err != nil {
		return sdk.ErrInvalidAddress(err.Error())
	}

	if bunAccount == nil {
		return nil
	}

	amt := bunAccount.GetCoins()

	if amt.Empty() {
		return nil
	}

	ctx.Logger().Info("find burn coin.", fmt.Sprintf("address: %s, coins: %s", bunAccount.GetAddress(), bunAccount.GetCoins()))

	_, err = k.bankKeeper.SubtractCoins(ctx, bunAccount.GetAddress(), amt)

	if err != nil {
		return sdk.ErrInvalidCoins(err.Error())
	}

	ctx.Logger().Info("Success burn coin to bunAccount.", fmt.Sprintf("address: %s, coins: %s", acc, amt))

	burnCoinsToSupply(ctx, k, amt)

	return nil
}

func getAccount(ctx sdk.Context, k Keeper, acc string) (accountExported.Account, error) {
	burnAcc, err := sdk.AccAddressFromBech32(acc)

	if err != nil {
		return nil, err
	}

	return k.accountKeeper.GetAccount(ctx, burnAcc), nil
}

func burnCoinsToSupply(ctx sdk.Context, k Keeper, amt sdk.Coins) {
	supply := k.supplyKeeper.GetSupply(ctx)
	supply = supply.Deflate(amt)
	k.supplyKeeper.SetSupply(ctx, supply)

	ctx.Logger().Info("Success burn coin to supply.", fmt.Sprintf("total: %s", supply.GetTotal()))
}
