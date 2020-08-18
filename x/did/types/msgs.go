package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgCreateDID{}
	_ sdk.Msg = &MsgUpdateDID{}
)

// MsgCreateDID defines a CreateDID message.
type MsgCreateDID struct {
	DID          DID            `json:"did"`
	Document     DIDDocument    `json:"document"`
	OwnerAddress sdk.AccAddress `json:"owner_address"`
}

// NewMsgCreateDID is a constructor of MsgCreateDID.
func NewMsgCreateDID(did DID, doc DIDDocument, ownerAddr sdk.AccAddress) MsgCreateDID {
	return MsgCreateDID{did, doc, ownerAddr}
}

// Route returns the name of the module.
func (msg MsgCreateDID) Route() string { return RouterKey }

// Type returns the name of the action.
func (msg MsgCreateDID) Type() string { return "create_did" }

// VaValidateBasic runs stateless checks on the message.
func (msg MsgCreateDID) ValidateBasic() sdk.Error {
	if !msg.DID.Valid() {
		return ErrInvalidDID(msg.DID)
	}
	if !msg.Document.Valid() {
		return ErrInvalidDIDDocument()
	}
	if msg.OwnerAddress.Empty() {
		return sdk.ErrInvalidAddress(msg.OwnerAddress.String())
	}
	return nil
}

// GetSignBytes returns the canonical byte representation of the message. Used to generate a signature.
func (msg MsgCreateDID) GetSignBytes() []byte {
	return sdk.MustSortJSON(didCodec.MustMarshalJSON(msg))
}

// GetSigners return the addresses of signers that must sign.
func (msg MsgCreateDID) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.OwnerAddress}
}

// MsgUpdateDID defines a UpdateDID message.
type MsgUpdateDID struct {
	DID          DID            `json:"did"`
	Document     DIDDocument    `json:"document"`
	SigPubKeyID  PubKeyID       `json:"sig_pubkey_id"`
	Signature    []byte         `json:"signature"`
	OwnerAddress sdk.AccAddress `json:"owner_address"`
}

// NewMsgUpdateDID is a constructor of MsgUpdateDID.
func NewMsgUpdateDID(did DID, doc DIDDocument, sigPubKeyID PubKeyID, sig []byte, ownerAddr sdk.AccAddress) MsgUpdateDID {
	return MsgUpdateDID{did, doc, sigPubKeyID, sig, ownerAddr}
}

// Route returns the name of the module.
func (msg MsgUpdateDID) Route() string { return RouterKey }

// Type returns the name of the action.
func (msg MsgUpdateDID) Type() string { return "update_did" }

// VaValidateBasic runs stateless checks on the message.
func (msg MsgUpdateDID) ValidateBasic() sdk.Error {
	if !msg.DID.Valid() {
		return ErrInvalidDID(msg.DID)
	}
	if !msg.Document.Valid() {
		return ErrInvalidDIDDocument()
	}
	if msg.Signature == nil || len(msg.Signature) == 0 {
		return ErrInvalidSignature(msg.Signature)
	}
	if msg.OwnerAddress.Empty() {
		return sdk.ErrInvalidAddress(msg.OwnerAddress.String())
	}
	return nil
}

// GetSignBytes returns the canonical byte representation of the message. Used to generate a signature.
func (msg MsgUpdateDID) GetSignBytes() []byte {
	return sdk.MustSortJSON(didCodec.MustMarshalJSON(msg))
}

// GetSigners return the addresses of signers that must sign.
func (msg MsgUpdateDID) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.OwnerAddress}
}
