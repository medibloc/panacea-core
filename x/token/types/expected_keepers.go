package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
)

type BankKeeper interface {
	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error
	GetSupply(ctx sdk.Context) exported.SupplyI
	SetSupply(ctx sdk.Context, supply exported.SupplyI)
}
