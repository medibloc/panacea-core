package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateDIDRequest{}

func NewMsgCreateDIDResponse(did string, document DIDDocument, VerificationMethodID string, Signature []byte, FromAddress string) MsgCreateDIDRequest {
	return MsgCreateDIDRequest{
		Did:                  did,
		Document:             &document,
		VerificationMethodId: VerificationMethodID,
		Signature:            Signature,
		FromAddress:          FromAddress,
	}
}

func (msg *MsgCreateDIDRequest) Route() string {
	return RouterKey
}

func (msg *MsgCreateDIDRequest) Type() string {
	return "create_did"
}

func (msg *MsgCreateDIDRequest) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateDIDRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg *MsgCreateDIDRequest) ValidateBasic() error {
	if !ValidateDID(msg.Did) {
		return errors.Wrapf(ErrInvalidDID, "did: %v", msg.Did)
	}
	if !msg.Document.Valid() {
		return errors.Wrapf(ErrInvalidDIDDocument, "DIDDocument: %v", msg.Document)
	}
	if msg.Signature == nil || len(msg.Signature) == 0 {
		return errors.Wrapf(ErrInvalidSignature, "Signature: %v", msg.Signature)
	}

	addr, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return err
	}
	if addr.Empty() {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "Address: %s", addr.String())
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateDIDRequest{}

// NewMsgUpdateDID is a constructor of MsgUpdateDID.
func NewMsgUpdateDID(did string, doc DIDDocument, verificationMethodID string, sig []byte, fromAddr string) *MsgUpdateDIDRequest {
	return &MsgUpdateDIDRequest{
		Did:                  did,
		Document:             &doc,
		VerificationMethodId: verificationMethodID,
		Signature:            sig,
		FromAddress:          fromAddr,
	}
}

// Route returns the name of the module.
func (msg *MsgUpdateDIDRequest) Route() string { return RouterKey }

// Type returns the name of the action.
func (msg *MsgUpdateDIDRequest) Type() string { return "update_did" }

// ValidateBasic runs stateless checks on the message.
func (msg *MsgUpdateDIDRequest) ValidateBasic() error {
	if !ValidateDID(msg.Did) {
		return errors.Wrapf(ErrInvalidDID, "DID: %v", msg.Did)
	}
	if !msg.Document.Valid() {
		return errors.Wrapf(ErrInvalidDIDDocument, "DIDDocument: %v", msg.Document)
	}
	if msg.Signature == nil || len(msg.Signature) == 0 {
		return errors.Wrapf(ErrInvalidSignature, "Signature: %v", msg.Signature)
	}
	addr, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return err
	}
	if addr.Empty() {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "Address: %v", addr.String())
	}
	return nil
}

// GetSignBytes returns the canonical byte representation of the message. Used to generate a signature.
func (msg *MsgUpdateDIDRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners return the addresses of signers that must sign.
func (msg *MsgUpdateDIDRequest) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

var _ sdk.Msg = &MsgDeactivateDIDRequest{}

// NewMsgDeactivateDIDRequest is a constructor of MsgDeactivateDID.
func NewMsgDeactivateDIDRequest(did string, verificationMethodID string, sig []byte, fromAddr string) *MsgDeactivateDIDRequest {
	return &MsgDeactivateDIDRequest{did, verificationMethodID, sig, fromAddr}
}

// Route returns the name of the module.
func (msg *MsgDeactivateDIDRequest) Route() string { return RouterKey }

// Type returns the name of the action.
func (msg *MsgDeactivateDIDRequest) Type() string { return "deactivate_did" }

// ValidateBasic runs stateless checks on the message.
func (msg *MsgDeactivateDIDRequest) ValidateBasic() error {
	if !ValidateDID(msg.Did) {
		return errors.Wrapf(ErrInvalidDID, "DID: %v", msg.Did)
	}
	if msg.Signature == nil || len(msg.Signature) == 0 {
		return errors.Wrapf(ErrInvalidSignature, "Signature: %v", msg.Signature)
	}

	addr, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return err
	}
	if addr.Empty() {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "Address: %s", addr.String())
	}
	return nil
}

// GetSignBytes returns the canonical byte representation of the message. Used to generate a signature.
func (msg *MsgDeactivateDIDRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners return the addresses of signers that must sign.
func (msg *MsgDeactivateDIDRequest) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}
