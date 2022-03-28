package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// DONTCOVER

// x/datapool module sentinel errors
var (
	ErrDataValidatorNotFound      = sdkerrors.Register(ModuleName, 1, "data validator not found")
	ErrDataValidatorAlreadyExist  = sdkerrors.Register(ModuleName, 2, "data validator already exists")
	ErrPoolAlreadyExist           = sdkerrors.Register(ModuleName, 3, "data pool already exists")
	ErrNotEnoughPoolDeposit       = sdkerrors.Register(ModuleName, 4, "the balance is not enough to make a data pool")
	ErrNotRegisteredDataValidator = sdkerrors.Register(ModuleName, 5, "data validator is not registered")
	ErrNoRegisteredContract       = sdkerrors.Register(ModuleName, 6, "no contract is registered")
)
