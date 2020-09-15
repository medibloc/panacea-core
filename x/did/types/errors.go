package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const DefaultCodespace sdk.CodespaceType = ModuleName

const (
	CodeDIDExists                 sdk.CodeType = 101
	CodeInvalidDID                sdk.CodeType = 102
	CodeInvalidDIDDocument        sdk.CodeType = 103
	CodeDIDNotFound               sdk.CodeType = 104
	CodeInvalidSignature          sdk.CodeType = 105
	CodeInvalidVeriMethodID       sdk.CodeType = 106
	CodeVeriMethodIDNotFound      sdk.CodeType = 107
	CodeSigVerificationFailed     sdk.CodeType = 108
	CodeInvalidSecp256k1PublicKey sdk.CodeType = 109
	CodeInvalidNetworkID          sdk.CodeType = 110
	CodeInvalidDIDDocumentWithSeq sdk.CodeType = 111
	CodeDIDDeactivated            sdk.CodeType = 112
	CodeInvalidKeyController      sdk.CodeType = 113
)

func ErrDIDExists(did DID) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeDIDExists, "DID %v already exists", did)
}

func ErrInvalidDID(did string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidDID, "Invalid DID %v", did)
}

func ErrInvalidDIDDocument(doc DIDDocument) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidDIDDocument, "Invalid DID Document: %v", doc)
}

func ErrDIDNotFound(did DID) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeDIDNotFound, "DID %v not found", did)
}

func ErrInvalidSignature(sig []byte) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidSignature, "Invalid signature %v", sig)
}

func ErrInvalidVeriMethodID(id string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidVeriMethodID, "Invalid VeriMethodID: %s", id)
}

func ErrVeriMethodIDNotFound(id VeriMethodID) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeVeriMethodIDNotFound, "VeriMethodID %v not found", id)
}

func ErrSigVerificationFailed() sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeSigVerificationFailed, "DID signature verification was failed")
}

func ErrInvalidSecp256k1PublicKey(err error) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidSecp256k1PublicKey, "Invalid Secp256k1 public key: %v", err)
}

func ErrInvalidNetworkID(id string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidNetworkID, "Invalid network ID: %s", id)
}

func ErrInvalidDIDDocumentWithSeq(doc DIDDocumentWithSeq) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidDIDDocumentWithSeq, "Invalid DIDDocumentWithSeq: %v", doc)
}

func ErrDIDDeactivated(did DID) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeDIDDeactivated, "DID was already deactivated: %v", did)
}

func ErrInvalidKeyController(did DID) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidKeyController, "Invalid key controller: %v", did)
}
