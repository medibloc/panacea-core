package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/aol module sentinel errors
var (
	ErrMessageTooLarge     = sdkerrors.Register(ModuleName, 1, "message too large")
	ErrInvalidTopic        = sdkerrors.Register(ModuleName, 2, "invalid topic")
	ErrInvalidMoniker      = sdkerrors.Register(ModuleName, 3, "invalid moniker")
	ErrTopicExists         = sdkerrors.Register(ModuleName, 4, "topic already exists")
	ErrWriterExists        = sdkerrors.Register(ModuleName, 5, "writer already exists")
	ErrTopicNotFound       = sdkerrors.Register(ModuleName, 6, "topic not found")
	ErrWriterNotFound      = sdkerrors.Register(ModuleName, 7, "writer not found")
	ErrWriterNotAuthorized = sdkerrors.Register(ModuleName, 8, "writer not authorized")
)
