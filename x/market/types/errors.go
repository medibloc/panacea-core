package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/market module sentinel errors
var (
	ErrDealAlreadyExist = sdkerrors.Register(ModuleName, 1, "deal already exist")
	ErrNotEnoughBalance = sdkerrors.Register(ModuleName, 2, "The balance is not enough to make deal")
	ErrInvalidSignature  = sdkerrors.Register(ModuleName, 3, "invalid data validator signature")
)
