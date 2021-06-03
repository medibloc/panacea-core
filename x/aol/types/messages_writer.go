package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateWriter{}

func NewMsgCreateWriter(creator string, moniker string, description string, nanoTimestamp int32) *MsgCreateWriter {
	return &MsgCreateWriter{
		Creator:       creator,
		Moniker:       moniker,
		Description:   description,
		NanoTimestamp: nanoTimestamp,
	}
}

func (msg *MsgCreateWriter) Route() string {
	return RouterKey
}

func (msg *MsgCreateWriter) Type() string {
	return "CreateWriter"
}

func (msg *MsgCreateWriter) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateWriter) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateWriter) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateWriter{}

func NewMsgUpdateWriter(creator string, id uint64, moniker string, description string, nanoTimestamp int32) *MsgUpdateWriter {
	return &MsgUpdateWriter{
		Id:            id,
		Creator:       creator,
		Moniker:       moniker,
		Description:   description,
		NanoTimestamp: nanoTimestamp,
	}
}

func (msg *MsgUpdateWriter) Route() string {
	return RouterKey
}

func (msg *MsgUpdateWriter) Type() string {
	return "UpdateWriter"
}

func (msg *MsgUpdateWriter) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateWriter) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateWriter) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgCreateWriter{}

func NewMsgDeleteWriter(creator string, id uint64) *MsgDeleteWriter {
	return &MsgDeleteWriter{
		Id:      id,
		Creator: creator,
	}
}
func (msg *MsgDeleteWriter) Route() string {
	return RouterKey
}

func (msg *MsgDeleteWriter) Type() string {
	return "DeleteWriter"
}

func (msg *MsgDeleteWriter) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDeleteWriter) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteWriter) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
