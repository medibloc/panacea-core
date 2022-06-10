package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// DONTCOVER

var (
	ErrOracleNotFound          = sdkerrors.Register(ModuleName, 1, "oracle not found")
	ErrOracleAlreadyExist      = sdkerrors.Register(ModuleName, 2, "oracle already exists")
	ErrInvalidUpdateRequester  = sdkerrors.Register(ModuleName, 3, "invalid update requester")
	ErrNotRegisteredOracle     = sdkerrors.Register(ModuleName, 4, "oracle is not registered")
	ErrInvalidOracleAccAddress = sdkerrors.Register(ModuleName, 5, "invalid oracle account address")
)
