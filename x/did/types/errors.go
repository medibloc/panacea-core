package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const DefaultCodespace sdk.CodespaceType = ModuleName

const (
	CodeDIDExists  sdk.CodeType = 101
	CodeInvalidDID sdk.CodeType = 102
)

func ErrDIDExists(did DID) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeDIDExists, "DID %v already exists", did)
}

func ErrInvalidDID(did DID) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidDID, "Invalid DID %v", did)
}
