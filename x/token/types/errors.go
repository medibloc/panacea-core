package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const DefaultCodespace sdk.CodespaceType = ModuleName

const (
	CodeInvalidName        sdk.CodeType = 101
	CodeInvalidSymbol      sdk.CodeType = 102
	CodeSymbolNotAllowed   sdk.CodeType = 103
	CodeInvalidTotalSupply sdk.CodeType = 104
	CodeTokenExists        sdk.CodeType = 105
	CodeTokenNotExists     sdk.CodeType = 106
)

func ErrInvalidName(name string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidName, "Invalid name: %v", name)
}

func ErrInvalidSymbol(symbol string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidSymbol, "Invalid symbol: %v", symbol)
}

func ErrSymbolNotAllowed(symbol string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeSymbolNotAllowed, "Symbol(%v) not allowed", symbol)
}

func ErrInvalidTotalSupply(msg string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidTotalSupply, msg)
}

func ErrTokenExists(symbol Symbol) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeTokenExists, "Token(%v) already exists", symbol)
}

func ErrTokenNotExists(symbol Symbol) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeTokenNotExists, "Token(%v) not exists", symbol)
}
