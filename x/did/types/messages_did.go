package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateDID{}

func NewMsgCreateDID(did string, document DIDDocument, VerificationMethodID string, Signature []byte, FromAddress string) MsgCreateDID {
	return MsgCreateDID{
		Did:                  did,
		Document:             &document,
		VerificationMethodId: VerificationMethodID,
		Signature:            Signature,
		FromAddress:          FromAddress,
	}
}

func (msg *MsgCreateDID) Route() string {
	return RouterKey
}

func (msg *MsgCreateDID) Type() string {
	return "create_did"
}

func (msg *MsgCreateDID) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateDID) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg *MsgCreateDID) ValidateBasic() error {
	if !ValidateDID(msg.Did) {
		return sdkerrors.Wrapf(ErrInvalidDID, "did: %v", msg.Did)
	}
	if !msg.Document.Valid() {
		return sdkerrors.Wrapf(ErrInvalidDIDDocument, "DIDDocument: %v", msg.Document)
	}
	if msg.Signature == nil || len(msg.Signature) == 0 {
		return sdkerrors.Wrapf(ErrInvalidSignature, "Signature: %v", msg.Signature)
	}

	addr, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return err
	}
	if addr.Empty() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Address: %s", addr.String())
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateDID{}

// NewMsgUpdateDID is a constructor of MsgUpdateDID.
func NewMsgUpdateDID(did string, doc DIDDocument, verificationMethodID string, sig []byte, fromAddr string) MsgUpdateDID {
	return MsgUpdateDID{
		Did:                  did,
		Document:             &doc,
		VerificationMethodId: verificationMethodID,
		Signature:            sig,
		FromAddress:          fromAddr,
	}
}

// Route returns the name of the module.
func (msg MsgUpdateDID) Route() string { return RouterKey }

// Type returns the name of the action.
func (msg MsgUpdateDID) Type() string { return "update_did" }

// ValidateBasic runs stateless checks on the message.
func (msg MsgUpdateDID) ValidateBasic() error {
	if !ValidateDID(msg.Did) {
		return sdkerrors.Wrapf(ErrInvalidDID, "DID: %v", msg.Did)
	}
	if !msg.Document.Valid() {
		return sdkerrors.Wrapf(ErrInvalidDIDDocument, "DIDDocument: %v", msg.Document)
	}
	if msg.Signature == nil || len(msg.Signature) == 0 {
		return sdkerrors.Wrapf(ErrInvalidSignature, "Signature: %v", msg.Signature)
	}
	addr, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return err
	}
	if addr.Empty() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Address: %v", addr.String())
	}
	return nil
}

// GetSignBytes returns the canonical byte representation of the message. Used to generate a signature.
func (msg MsgUpdateDID) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners return the addresses of signers that must sign.
func (msg MsgUpdateDID) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

var _ sdk.Msg = &MsgDeactivateDID{}

// NewMsgDeactivateDID is a constructor of MsgDeactivateDID.
func NewMsgDeactivateDID(did string, verificationMethodID string, sig []byte, fromAddr string) MsgDeactivateDID {
	return MsgDeactivateDID{did, verificationMethodID, sig, fromAddr}
}

// Route returns the name of the module.
func (msg MsgDeactivateDID) Route() string { return RouterKey }

// Type returns the name of the action.
func (msg MsgDeactivateDID) Type() string { return "deactivate_did" }

// VaValidateBasic runs stateless checks on the message.
func (msg MsgDeactivateDID) ValidateBasic() error {
	if !ValidateDID(msg.Did) {
		return sdkerrors.Wrapf(ErrInvalidDID, "DID: %v", msg.Did)
	}
	if msg.Signature == nil || len(msg.Signature) == 0 {
		return sdkerrors.Wrapf(ErrInvalidSignature, "Signature: %v", msg.Signature)
	}

	addr, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return err
	}
	if addr.Empty() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Address: %s", addr.String())
	}
	return nil
}

// GetSignBytes returns the canonical byte representation of the message. Used to generate a signature.
func (msg MsgDeactivateDID) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners return the addresses of signers that must sign.
func (msg MsgDeactivateDID) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}
