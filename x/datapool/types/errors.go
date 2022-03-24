package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// DONTCOVER

// x/datapool module sentinel errors
var (
	ErrDataValidatorNotFound     = sdkerrors.Register(ModuleName, 1, "data validator not found")
	ErrDataValidatorAlreadyExist = sdkerrors.Register(ModuleName, 2, "data validator already exists")
)
