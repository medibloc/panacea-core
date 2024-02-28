package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgServiceAddWriterRequest{}

func NewMsgAddWriter(topicName, moniker, description, writerAddress, ownerAddress string) *MsgServiceAddWriterRequest {
	return &MsgServiceAddWriterRequest{
		TopicName:     topicName,
		Moniker:       moniker,
		Description:   description,
		WriterAddress: writerAddress,
		OwnerAddress:  ownerAddress,
	}
}

func (msg *MsgServiceAddWriterRequest) Route() string {
	return RouterKey
}

func (msg *MsgServiceAddWriterRequest) Type() string {
	return "AddWriter"
}

func (msg *MsgServiceAddWriterRequest) GetSigners() []sdk.AccAddress {
	ownerAddress, err := sdk.AccAddressFromBech32(msg.OwnerAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{ownerAddress}
}

func (msg *MsgServiceAddWriterRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgServiceAddWriterRequest) ValidateBasic() error {
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
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid writer address (%s)", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.OwnerAddress); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgServiceDeleteWriterRequest{}

func NewMsgDeleteWriter(topicName, writerAddress, ownerAddress string) *MsgServiceDeleteWriterRequest {
	return &MsgServiceDeleteWriterRequest{
		TopicName:     topicName,
		WriterAddress: writerAddress,
		OwnerAddress:  ownerAddress,
	}
}
func (msg *MsgServiceDeleteWriterRequest) Route() string {
	return RouterKey
}

func (msg *MsgServiceDeleteWriterRequest) Type() string {
	return "DeleteWriter"
}

func (msg *MsgServiceDeleteWriterRequest) GetSigners() []sdk.AccAddress {
	ownerAddress, err := sdk.AccAddressFromBech32(msg.OwnerAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{ownerAddress}
}

func (msg *MsgServiceDeleteWriterRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgServiceDeleteWriterRequest) ValidateBasic() error {
	if err := validateTopicName(msg.TopicName); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.WriterAddress); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid writer address (%s)", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.OwnerAddress); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address (%s)", err)
	}
	return nil
}
