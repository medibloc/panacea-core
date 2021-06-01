package types

// DONTCOVER

import (
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/did module sentinel errors
var (
	ErrDIDExists = sdkerrors.Register(ModuleName, 101, "DID %v already exists")
	ErrInvalidDID = sdkerrors.Register(ModuleName, 102, "Invalid DID %v")
	ErrInvalidDIDDocument = sdkerrors.Register(ModuleName, 103, "Invalid DID Document: %v")
	ErrDIDNotFound = sdkerrors.Register(ModuleName, 104, "DID %v not found")
	ErrInvalidSignature = sdkerrors.Register(ModuleName, 105, "Invalid signature %v")
	ErrInvalidVerificationMethodID = sdkerrors.Register(ModuleName, 106, "Invalid VerificationMethodID: %s")
	ErrVerificationMethodIDNotFound = sdkerrors.Register(ModuleName, 107, "VerificationMethodID %v not found")
	ErrSigVerificationFailed = sdkerrors.Register(ModuleName, 108, "DID signature verification was failed")
	ErrInvalidSecp256k1PublicKey = sdkerrors.Register(ModuleName, 109, "Invalid Secp256k1 public key: %v")
	ErrInvalidNetworkID = sdkerrors.Register(ModuleName, 110, "Invalid network ID: %s")
	ErrInvalidDIDDocumentWithSeq = sdkerrors.Register(ModuleName, 111, "Invalid DIDDocumentWithSeq: %v")
	ErrDIDDeactivated = sdkerrors.Register(ModuleName, 112, "DID was already deactivated: %v")
	ErrVerificationMethodKeyTypeNotImplemented = sdkerrors.Register(ModuleName, 114, "Verification not implemented with key type: %v")

)

func Error(error *sdkerrors.Error, args ...interface{}) error {
	return sdkerrors.New(error.Codespace(), error.ABCICode(), fmt.Sprintf(error.Error(), args))
}
