package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const DefaultCodespace sdk.CodespaceType = ModuleName

const (
	CodeMessageTooLarge     sdk.CodeType = 101
	CodeTopicExists         sdk.CodeType = 102
	CodeTopicNotFound       sdk.CodeType = 103
	CodeWriterExists        sdk.CodeType = 104
	CodeWriterNotFound      sdk.CodeType = 104
	CodeWriterNotAuthorized sdk.CodeType = 105
	CodeInvalidTopic        sdk.CodeType = 106
	CodeInvalidMoniker      sdk.CodeType = 107
)

func ErrMessageTooLarge(descriptor string, got, max int) sdk.Error {
	msg := fmt.Sprintf("bad message length for %v, got length %v, max is %v", descriptor, got, max)
	return sdk.NewError(DefaultCodespace, CodeMessageTooLarge, msg)
}

func ErrTopicExists(topic string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeTopicExists, "topic %v already exists for this owner", topic)
}

func ErrTopicNotFound(topic string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeTopicNotFound, "topic %v not found", topic)
}

func ErrWriterExists(writer sdk.AccAddress) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeWriterExists, "writer %v already exists for this topic", writer)
}

func ErrWriterNotFound(writer sdk.AccAddress) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeWriterNotFound, "writer %v not found", writer)
}

func ErrWriterNotAuthorized(writer sdk.AccAddress) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeWriterNotAuthorized, "writer %v not authorized", writer)
}

func ErrInvalidTopic(topic string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidTopic, "invalid topic %v", topic)
}

func ErrInvalidMoniker(moniker string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidMoniker, "invalid moniker %v", moniker)
}
