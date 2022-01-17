package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/market module sentinel errors
var (
	ErrDealAlreadyExist = sdkerrors.Register(ModuleName, 1, "deal already exist")
	ErrNotEnoughBalance = sdkerrors.Register(ModuleName, 2, "The owner's balance is not enough to make deal")
)
