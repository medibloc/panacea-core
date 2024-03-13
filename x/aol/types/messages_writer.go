package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgAddWriterRequest{}

func NewMsgAddWriter(topicName, moniker, description, writerAddress, ownerAddress string) *MsgAddWriterRequest {
	return &MsgAddWriterRequest{
		TopicName:     topicName,
		Moniker:       moniker,
		Description:   description,
		WriterAddress: writerAddress,
		OwnerAddress:  ownerAddress,
	}
}

func (msg *MsgAddWriterRequest) Route() string {
	return RouterKey
}

func (msg *MsgAddWriterRequest) Type() string {
	return "AddWriter"
}

func (msg *MsgAddWriterRequest) GetSigners() []sdk.AccAddress {
	ownerAddress, err := sdk.AccAddressFromBech32(msg.OwnerAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{ownerAddress}
}

func (msg *MsgAddWriterRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddWriterRequest) ValidateBasic() error {
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

var _ sdk.Msg = &MsgDeleteWriterRequest{}

func NewMsgDeleteWriter(topicName, writerAddress, ownerAddress string) *MsgDeleteWriterRequest {
	return &MsgDeleteWriterRequest{
		TopicName:     topicName,
		WriterAddress: writerAddress,
		OwnerAddress:  ownerAddress,
	}
}
func (msg *MsgDeleteWriterRequest) Route() string {
	return RouterKey
}

func (msg *MsgDeleteWriterRequest) Type() string {
	return "DeleteWriter"
}

func (msg *MsgDeleteWriterRequest) GetSigners() []sdk.AccAddress {
	ownerAddress, err := sdk.AccAddressFromBech32(msg.OwnerAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{ownerAddress}
}

func (msg *MsgDeleteWriterRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteWriterRequest) ValidateBasic() error {
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
