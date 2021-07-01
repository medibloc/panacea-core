package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateTopic{}

func NewMsgCreateTopic(topicName, description, ownerAddress string) *MsgCreateTopic {
	return &MsgCreateTopic{
		TopicName:    topicName,
		Description:  description,
		OwnerAddress: ownerAddress,
	}
}

func (msg *MsgCreateTopic) Route() string {
	return RouterKey
}

func (msg *MsgCreateTopic) Type() string {
	return "CreateTopic"
}

func (msg *MsgCreateTopic) GetSigners() []sdk.AccAddress {
	ownerAddress, err := sdk.AccAddressFromBech32(msg.OwnerAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{ownerAddress}
}

func (msg *MsgCreateTopic) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateTopic) ValidateBasic() error {
	if err := validateTopicName(msg.TopicName); err != nil {
		return err
	}
	if err := validateDescription(msg.Description); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.OwnerAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address (%s)", err)
	}
	return nil
}
