package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgAddWriter{}

func NewMsgAddWriter(topicName, moniker, description, writerAddress, ownerAddress string) *MsgAddWriter {
	return &MsgAddWriter{
		TopicName:     topicName,
		Moniker:       moniker,
		Description:   description,
		WriterAddress: writerAddress,
		OwnerAddress:  ownerAddress,
	}
}

func (msg *MsgAddWriter) Route() string {
	return RouterKey
}

func (msg *MsgAddWriter) Type() string {
	return "AddWriter"
}

func (msg *MsgAddWriter) GetSigners() []sdk.AccAddress {
	ownerAddress, err := sdk.AccAddressFromBech32(msg.OwnerAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{ownerAddress}
}

func (msg *MsgAddWriter) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddWriter) ValidateBasic() error {
	if err := validateTopicName(msg.TopicName); err != nil {
		return err
	}
	if err := validateMoniker(msg.Moniker); err != nil {
		return err
	}
	if err := validateDescription(msg.Description); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.WriterAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid writer address (%s)", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.OwnerAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgDeleteWriter{}

func NewMsgDeleteWriter(topicName, writerAddress, ownerAddress string) *MsgDeleteWriter {
	return &MsgDeleteWriter{
		TopicName:     topicName,
		WriterAddress: writerAddress,
		OwnerAddress:  ownerAddress,
	}
}
func (msg *MsgDeleteWriter) Route() string {
	return RouterKey
}

func (msg *MsgDeleteWriter) Type() string {
	return "DeleteWriter"
}

func (msg *MsgDeleteWriter) GetSigners() []sdk.AccAddress {
	ownerAddress, err := sdk.AccAddressFromBech32(msg.OwnerAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{ownerAddress}
}

func (msg *MsgDeleteWriter) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteWriter) ValidateBasic() error {
	if err := validateTopicName(msg.TopicName); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.WriterAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid writer address (%s)", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.OwnerAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address (%s)", err)
	}
	return nil
}
