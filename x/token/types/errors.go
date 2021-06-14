package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/token module sentinel errors
var (
	ErrInvalidName        = sdkerrors.Register(ModuleName, 2, "invalid name")
	ErrSymbolNotAllowed   = sdkerrors.Register(ModuleName, 3, "symbol not allowed")
	ErrInvalidSymbol      = sdkerrors.Register(ModuleName, 4, "invalid symbol")
	ErrInvalidTotalSupply = sdkerrors.Register(ModuleName, 5, "invalid total supply")
	ErrTokenExists        = sdkerrors.Register(ModuleName, 6, "token already exists")
	ErrInvalidDenom       = sdkerrors.Register(ModuleName, 7, "invalid denom")
)
