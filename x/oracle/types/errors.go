package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrInvalidUniqueId       = sdkerrors.Register(ModuleName, 1, "invalid unique id")
	ErrGetOracle             = sdkerrors.Register(ModuleName, 2, "error while get oracle")
	ErrGetOracleRegistration = sdkerrors.Register(ModuleName, 3, "error while get oracleRegistration")

	ErrOracleRegistrationNotFound = sdkerrors.Register(ModuleName, 4, "oracle registration not found")
)
