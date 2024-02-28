package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/did module sentinel errors
var (
	ErrDIDExists                               = errors.Register(ModuleName, 2, "DID already exists")
	ErrInvalidDID                              = errors.Register(ModuleName, 3, "Invalid DID")
	ErrInvalidDIDDocument                      = errors.Register(ModuleName, 4, "Invalid DID Document")
	ErrDIDNotFound                             = errors.Register(ModuleName, 5, "DID not found")
	ErrInvalidSignature                        = errors.Register(ModuleName, 6, "Invalid signature")
	ErrInvalidVerificationMethodID             = errors.Register(ModuleName, 7, "Invalid VerificationMethodID")
	ErrVerificationMethodIDNotFound            = errors.Register(ModuleName, 8, "VerificationMethodID not found")
	ErrSigVerificationFailed                   = errors.Register(ModuleName, 9, "DID signature verification was failed")
	ErrInvalidSecp256k1PublicKey               = errors.Register(ModuleName, 10, "Invalid Secp256k1 public key")
	ErrInvalidNetworkID                        = errors.Register(ModuleName, 11, "Invalid network ID")
	ErrInvalidDIDDocumentWithSeq               = errors.Register(ModuleName, 12, "Invalid DIDDocumentWithSeq")
	ErrDIDDeactivated                          = errors.Register(ModuleName, 13, "DID was already deactivated")
	CodeInvalidKeyController                   = errors.Register(ModuleName, 14, "Invalid key controller")
	ErrVerificationMethodKeyTypeNotImplemented = errors.Register(ModuleName, 15, "Verification not implemented with key type")
)
