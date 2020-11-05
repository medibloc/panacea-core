package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// BankKeeper defines the expected bank Keeper
type BankKeeper interface {
	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Error)
}
