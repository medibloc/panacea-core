package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateRecord{}

func NewMsgCreateRecord(creator string, key string, value string, nanoTimestamp int32, writerAddress string) *MsgCreateRecord {
	return &MsgCreateRecord{
		Creator:       creator,
		Key:           key,
		Value:         value,
		NanoTimestamp: nanoTimestamp,
		WriterAddress: writerAddress,
	}
}

func (msg *MsgCreateRecord) Route() string {
	return RouterKey
}

func (msg *MsgCreateRecord) Type() string {
	return "CreateRecord"
}

func (msg *MsgCreateRecord) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateRecord) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateRecord) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateRecord{}

func NewMsgUpdateRecord(creator string, id uint64, key string, value string, nanoTimestamp int32, writerAddress string) *MsgUpdateRecord {
	return &MsgUpdateRecord{
		Id:            id,
		Creator:       creator,
		Key:           key,
		Value:         value,
		NanoTimestamp: nanoTimestamp,
		WriterAddress: writerAddress,
	}
}

func (msg *MsgUpdateRecord) Route() string {
	return RouterKey
}

func (msg *MsgUpdateRecord) Type() string {
	return "UpdateRecord"
}

func (msg *MsgUpdateRecord) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateRecord) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateRecord) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgCreateRecord{}

func NewMsgDeleteRecord(creator string, id uint64) *MsgDeleteRecord {
	return &MsgDeleteRecord{
		Id:      id,
		Creator: creator,
	}
}
func (msg *MsgDeleteRecord) Route() string {
	return RouterKey
}

func (msg *MsgDeleteRecord) Type() string {
	return "DeleteRecord"
}

func (msg *MsgDeleteRecord) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDeleteRecord) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteRecord) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
