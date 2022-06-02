package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// DONTCOVER

var (
	ErrOracleNotFound     = sdkerrors.Register(ModuleName, 1, "oracle not found")
	ErrOracleAlreadyExist = sdkerrors.Register(ModuleName, 2, "oracle already exists")
)
