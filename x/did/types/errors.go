package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/did module sentinel errors
var (
	ErrDIDExists                   = sdkerrors.Register(ModuleName, 1, "DID already exists")
	ErrInvalidDID                  = sdkerrors.Register(ModuleName, 2, "Invalid DID")
	ErrInvalidDIDDocument          = sdkerrors.Register(ModuleName, 3, "Invalid DID Document")
	ErrDIDNotFound                 = sdkerrors.Register(ModuleName, 4, "DID not found")
	ErrInvalidVerificationMethodID = sdkerrors.Register(ModuleName, 5, "Invalid VerificationMethodID")
	ErrDIDDeactivated              = sdkerrors.Register(ModuleName, 6, "DID was already deactivated")
	ErrEmptyDocument               = sdkerrors.Register(ModuleName, 7, "Empty Document")
	ErrDIDNotMatched               = sdkerrors.Register(ModuleName, 8, "Document ID is not matched with did")
	ErrEmptyProof                  = sdkerrors.Register(ModuleName, 9, "Empty Document proof")
	ErrInvalidProof                = sdkerrors.Register(ModuleName, 10, "Verify with document proof was failed")
	ErrParseDocument               = sdkerrors.Register(ModuleName, 11, "Parse document was failed")
	ErrInvalidSequence             = sdkerrors.Register(ModuleName, 12, "Invalid sequence of did document")
	ErrVerifyOwnershipFailed       = sdkerrors.Register(ModuleName, 13, "Verify DID ownership failed")
)
