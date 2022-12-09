package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrOracleRegistration        = sdkerrors.Register(ModuleName, 1, "error while registering a oracle")
	ErrGetOracleRegistration     = sdkerrors.Register(ModuleName, 2, "error while getting a oracle registration")
	ErrGetOracle                 = sdkerrors.Register(ModuleName, 3, "error while getting a oracle")
	ErrOracleNotFound            = sdkerrors.Register(ModuleName, 4, "oracle not found")
	ErrInvalidUniqueID           = sdkerrors.Register(ModuleName, 5, "invalid unique id")
	ErrApproveOracleRegistration = sdkerrors.Register(ModuleName, 6, "error while approving an oracle registration")
)
