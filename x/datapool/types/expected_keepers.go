package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

type AccountKeeper interface {
	// Methods imported from account should be defined here
}
