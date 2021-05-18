package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	accountExported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	supplyExported "github.com/cosmos/cosmos-sdk/x/supply/exported"
)

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) accountExported.Account
}

type BankKeeper interface {
	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Error)
}

// SupplyKeeper defines the expected bank Keeper
type SupplyKeeper interface {
	GetSupply(ctx sdk.Context) (supply supplyExported.SupplyI)
	SetSupply(ctx sdk.Context, supply supplyExported.SupplyI)
}
