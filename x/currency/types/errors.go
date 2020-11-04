package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const DefaultCodespace sdk.CodespaceType = ModuleName

const (
	CodeDenomExists     sdk.CodeType = 101
	CodeDenomNotExists  sdk.CodeType = 102
	CodeDenomNotAllowed sdk.CodeType = 103
	CodeInvalidIssuance sdk.CodeType = 104
)

func ErrDenomExists(denom string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeDenomExists, "Denom(%v) already exists", denom)
}

func ErrDenomNotExists(denom string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeDenomNotExists, "Denom(%v) not exists", denom)
}

func ErrDenomNotAllowed(denom string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeDenomNotAllowed, "Denom(%v) is not allowed", denom)
}

func ErrInvalidIssuance(err sdk.Error) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidIssuance, "Invalid issuance: %v", err)
}
