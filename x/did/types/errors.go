package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const DefaultCodespace sdk.CodespaceType = ModuleName

const (
	CodeDIDExists sdk.CodeType = 101
)

func ErrDIDExists(did DID) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeDIDExists, "DID %v already exists", did)
}
