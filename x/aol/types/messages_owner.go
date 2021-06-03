package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateOwner{}

func NewMsgCreateOwner(creator string, totalTopics int32) *MsgCreateOwner {
	return &MsgCreateOwner{
		Creator:     creator,
		TotalTopics: totalTopics,
	}
}

func (msg *MsgCreateOwner) Route() string {
	return RouterKey
}

func (msg *MsgCreateOwner) Type() string {
	return "CreateOwner"
}

func (msg *MsgCreateOwner) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateOwner) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateOwner) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateOwner{}

func NewMsgUpdateOwner(creator string, id uint64, totalTopics int32) *MsgUpdateOwner {
	return &MsgUpdateOwner{
		Id:          id,
		Creator:     creator,
		TotalTopics: totalTopics,
	}
}

func (msg *MsgUpdateOwner) Route() string {
	return RouterKey
}

func (msg *MsgUpdateOwner) Type() string {
	return "UpdateOwner"
}

func (msg *MsgUpdateOwner) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateOwner) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateOwner) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgCreateOwner{}

func NewMsgDeleteOwner(creator string, id uint64) *MsgDeleteOwner {
	return &MsgDeleteOwner{
		Id:      id,
		Creator: creator,
	}
}
func (msg *MsgDeleteOwner) Route() string {
	return RouterKey
}

func (msg *MsgDeleteOwner) Type() string {
	return "DeleteOwner"
}

func (msg *MsgDeleteOwner) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDeleteOwner) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteOwner) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
