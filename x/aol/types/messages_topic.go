package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateTopic{}

func NewMsgCreateTopic(creator string, description string, totalRecords int32, totalWriter int32) *MsgCreateTopic {
	return &MsgCreateTopic{
		Creator:      creator,
		Description:  description,
		TotalRecords: totalRecords,
		TotalWriter:  totalWriter,
	}
}

func (msg *MsgCreateTopic) Route() string {
	return RouterKey
}

func (msg *MsgCreateTopic) Type() string {
	return "CreateTopic"
}

func (msg *MsgCreateTopic) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateTopic) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateTopic) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateTopic{}

func NewMsgUpdateTopic(creator string, id uint64, description string, totalRecords int32, totalWriter int32) *MsgUpdateTopic {
	return &MsgUpdateTopic{
		Id:           id,
		Creator:      creator,
		Description:  description,
		TotalRecords: totalRecords,
		TotalWriter:  totalWriter,
	}
}

func (msg *MsgUpdateTopic) Route() string {
	return RouterKey
}

func (msg *MsgUpdateTopic) Type() string {
	return "UpdateTopic"
}

func (msg *MsgUpdateTopic) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateTopic) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateTopic) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgCreateTopic{}

func NewMsgDeleteTopic(creator string, id uint64) *MsgDeleteTopic {
	return &MsgDeleteTopic{
		Id:      id,
		Creator: creator,
	}
}
func (msg *MsgDeleteTopic) Route() string {
	return RouterKey
}

func (msg *MsgDeleteTopic) Type() string {
	return "DeleteTopic"
}

func (msg *MsgDeleteTopic) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDeleteTopic) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteTopic) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
