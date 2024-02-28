package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgServiceCreateTopicRequest{}

func NewMsgCreateTopic(topicName, description, ownerAddress string) *MsgServiceCreateTopicRequest {
	return &MsgServiceCreateTopicRequest{
		TopicName:    topicName,
		Description:  description,
		OwnerAddress: ownerAddress,
	}
}

func (msg *MsgServiceCreateTopicRequest) Route() string {
	return RouterKey
}

func (msg *MsgServiceCreateTopicRequest) Type() string {
	return "CreateTopic"
}

func (msg *MsgServiceCreateTopicRequest) GetSigners() []sdk.AccAddress {
	ownerAddress, err := sdk.AccAddressFromBech32(msg.OwnerAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{ownerAddress}
}

func (msg *MsgServiceCreateTopicRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgServiceCreateTopicRequest) ValidateBasic() error {
	if err := validateTopicName(msg.TopicName); err != nil {
		return err
	}
	if err := validateDescription(msg.Description); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.OwnerAddress); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address (%s)", err)
	}
	return nil
}
