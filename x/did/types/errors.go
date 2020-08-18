package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const DefaultCodespace sdk.CodespaceType = ModuleName

const (
	CodeDIDExists                  sdk.CodeType = 101
	CodeInvalidDID                 sdk.CodeType = 102
	CodeInvalidDIDDocument         sdk.CodeType = 103
	CodeDIDNotFound                sdk.CodeType = 104
	CodeInvalidSignature           sdk.CodeType = 105
	CodePubKeyIDNotFound           sdk.CodeType = 106
	CodeSigVerificationFailed      sdk.CodeType = 107
	CodeInvalidSecp256k1PrivateKey sdk.CodeType = 108
	CodeInvalidSecp256k1PublicKey  sdk.CodeType = 109
)

func ErrDIDExists(did DID) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeDIDExists, "DID %v already exists", did)
}

func ErrInvalidDID(did DID) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidDID, "Invalid DID %v", did)
}

func ErrInvalidDIDDocument() sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidDIDDocument, "Invalid DID Document")
}

func ErrDIDNotFound(did DID) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeDIDNotFound, "DID %v not found", did)
}

func ErrInvalidSignature(sig []byte) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidSignature, "Invalid signature %v", sig)
}

func ErrPubKeyIDNotFound(id PubKeyID) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodePubKeyIDNotFound, "PubKeyID %v not found", id)
}

func ErrSigVerificationFailed() sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeSigVerificationFailed, "Signature verification was failed")
}

func ErrInvalidSecp256k1PrivateKey(err error) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidSecp256k1PrivateKey, "Invalid Secp256k1 private key: %v", err)
}

func ErrInvalidSecp256k1PublicKey(err error) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidSecp256k1PublicKey, "Invalid Secp256k1 public key: %v", err)
}
