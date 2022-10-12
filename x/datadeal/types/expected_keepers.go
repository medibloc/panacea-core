package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type AccountKeeper interface {
	NewAccount(sdk.Context, authtypes.AccountI) authtypes.AccountI

	SetAccount(sdk.Context, authtypes.AccountI)
	GetAccount(sdk.Context, sdk.AccAddress) authtypes.AccountI
}

type BankKeeper interface {
	SendCoins(sdk.Context, sdk.AccAddress, sdk.AccAddress, sdk.Coins) error
	GetBalance(sdk.Context, sdk.AccAddress, string) sdk.Coin
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}
