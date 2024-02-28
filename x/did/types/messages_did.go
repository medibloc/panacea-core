package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgServiceCreateDIDRequest{}

func NewMsgServiceCreateDIDResponse(did string, document DIDDocument, VerificationMethodID string, Signature []byte, FromAddress string) MsgServiceCreateDIDRequest {
	return MsgServiceCreateDIDRequest{
		Did:                  did,
		Document:             &document,
		VerificationMethodId: VerificationMethodID,
		Signature:            Signature,
		FromAddress:          FromAddress,
	}
}

func (msg *MsgServiceCreateDIDRequest) Route() string {
	return RouterKey
}

func (msg *MsgServiceCreateDIDRequest) Type() string {
	return "create_did"
}

func (msg *MsgServiceCreateDIDRequest) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgServiceCreateDIDRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg *MsgServiceCreateDIDRequest) ValidateBasic() error {
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

var _ sdk.Msg = &MsgServiceUpdateDIDRequest{}

// NewMsgUpdateDID is a constructor of MsgUpdateDID.
func NewMsgUpdateDID(did string, doc DIDDocument, verificationMethodID string, sig []byte, fromAddr string) *MsgServiceUpdateDIDRequest {
	return &MsgServiceUpdateDIDRequest{
		Did:                  did,
		Document:             &doc,
		VerificationMethodId: verificationMethodID,
		Signature:            sig,
		FromAddress:          fromAddr,
	}
}

// Route returns the name of the module.
func (msg *MsgServiceUpdateDIDRequest) Route() string { return RouterKey }

// Type returns the name of the action.
func (msg *MsgServiceUpdateDIDRequest) Type() string { return "update_did" }

// ValidateBasic runs stateless checks on the message.
func (msg *MsgServiceUpdateDIDRequest) ValidateBasic() error {
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
func (msg *MsgServiceUpdateDIDRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners return the addresses of signers that must sign.
func (msg *MsgServiceUpdateDIDRequest) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

var _ sdk.Msg = &MsgServiceDeactivateDIDRequest{}

// NewMsgServiceDeactivateDIDRequest is a constructor of MsgDeactivateDID.
func NewMsgServiceDeactivateDIDRequest(did string, verificationMethodID string, sig []byte, fromAddr string) *MsgServiceDeactivateDIDRequest {
	return &MsgServiceDeactivateDIDRequest{did, verificationMethodID, sig, fromAddr}
}

// Route returns the name of the module.
func (msg *MsgServiceDeactivateDIDRequest) Route() string { return RouterKey }

// Type returns the name of the action.
func (msg *MsgServiceDeactivateDIDRequest) Type() string { return "deactivate_did" }

// ValidateBasic runs stateless checks on the message.
func (msg *MsgServiceDeactivateDIDRequest) ValidateBasic() error {
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
func (msg *MsgServiceDeactivateDIDRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners return the addresses of signers that must sign.
func (msg *MsgServiceDeactivateDIDRequest) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}
