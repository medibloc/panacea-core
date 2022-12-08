package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrOracleRegistration    = sdkerrors.Register(ModuleName, 1, "error while registering a oracle")
	ErrGetOracleRegistration = sdkerrors.Register(ModuleName, 2, "error while getting a oracle registration")
	ErrGetOracle             = sdkerrors.Register(ModuleName, 3, "error while getting a oracle")
	ErrOracleNotFound        = sdkerrors.Register(ModuleName, 4, "oracle not found")
	ErrUpdateOracle          = sdkerrors.Register(ModuleName, 5, "error while updating oracle info")

	ErrCommissionNegative        = sdkerrors.Register(ModuleName, 6, "commission must be positive")
	ErrCommissionGTMaxRate       = sdkerrors.Register(ModuleName, 7, "commission cannot be more than the max rate")
	ErrCommissionUpdateTime      = sdkerrors.Register(ModuleName, 8, "commission cannot be changed more than once in 24h")
	ErrCommissionGTMaxChangeRate = sdkerrors.Register(ModuleName, 9, "commission cannot be changed more than max change rate")
)
