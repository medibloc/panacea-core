package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrOracleRegistration         = sdkerrors.Register(ModuleName, 1, "error while registering a oracle")
	ErrGetOracleRegistration      = sdkerrors.Register(ModuleName, 2, "error while getting a oracle registration")
	ErrGetOracle                  = sdkerrors.Register(ModuleName, 3, "error while getting a oracle")
	ErrOracleNotFound             = sdkerrors.Register(ModuleName, 4, "oracle not found")
	ErrInvalidUniqueId            = sdkerrors.Register(ModuleName, 5, "invalid unique id")
	ErrOracleRegistrationNotFound = sdkerrors.Register(ModuleName, 6, "oracle registration not found")
)
