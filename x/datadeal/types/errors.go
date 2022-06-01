package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/datadeal module sentinel errors
var (
	ErrDealAlreadyExist = sdkerrors.Register(ModuleName, 1, "deal already exist")
	ErrNotEnoughBalance = sdkerrors.Register(ModuleName, 2, "The balance is not enough to make deal")
	ErrInvalidSignature = sdkerrors.Register(ModuleName, 3, "invalid oracle signature")
	ErrInvalidStatus    = sdkerrors.Register(ModuleName, 4, "The deal's status is not invalid.")
	ErrDataAlreadyExist = sdkerrors.Register(ModuleName, 5, "data already exist")
	ErrDealNotFound     = sdkerrors.Register(ModuleName, 6, "deal is not found")
	ErrDataNotFound     = sdkerrors.Register(ModuleName, 7, "data is not found")
)
