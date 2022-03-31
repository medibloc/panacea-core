package types

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}

type AccountKeeper interface {
	NewAccount(sdk.Context, authtypes.AccountI) authtypes.AccountI

	SetAccount(ctx sdk.Context, acc authtypes.AccountI)
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	GetPubKey(sdk.Context, sdk.AccAddress) (cryptotypes.PubKey, error)
}
