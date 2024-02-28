package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/aol module sentinel errors
var (
	ErrMessageTooLarge     = errors.Register(ModuleName, 2, "message too large")
	ErrInvalidTopic        = errors.Register(ModuleName, 3, "invalid topic")
	ErrInvalidMoniker      = errors.Register(ModuleName, 4, "invalid moniker")
	ErrTopicExists         = errors.Register(ModuleName, 5, "topic already exists")
	ErrWriterExists        = errors.Register(ModuleName, 6, "writer already exists")
	ErrTopicNotFound       = errors.Register(ModuleName, 7, "topic not found")
	ErrWriterNotFound      = errors.Register(ModuleName, 8, "writer not found")
	ErrWriterNotAuthorized = errors.Register(ModuleName, 9, "writer not authorized")
)
