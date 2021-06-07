package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgAddRecord{}

func NewMsgAddRecord(topicName string, key, value []byte, writerAddress, ownerAddress, feePayerAddress string) *MsgAddRecord {
	return &MsgAddRecord{
		TopicName:       topicName,
		Key:             key,
		Value:           value,
		WriterAddress:   writerAddress,
		OwnerAddress:    ownerAddress,
		FeePayerAddress: feePayerAddress,
	}
}

func (msg *MsgAddRecord) Route() string {
	return RouterKey
}

func (msg *MsgAddRecord) Type() string {
	return "AddRecord"
}

func (msg *MsgAddRecord) GetSigners() []sdk.AccAddress {
	writerAddress, err := sdk.AccAddressFromBech32(msg.WriterAddress)
	if err != nil {
		panic(err)
	}

	if msg.FeePayerAddress != "" {
		feePayerAddress, err := sdk.AccAddressFromBech32(msg.FeePayerAddress)
		if err != nil {
			panic(err)
		}
		return []sdk.AccAddress{feePayerAddress, writerAddress}
	} else {
		return []sdk.AccAddress{writerAddress}
	}
}

func (msg *MsgAddRecord) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddRecord) ValidateBasic() error {
	if err := validateTopicName(msg.TopicName); err != nil {
		return err
	}
	if err := validateRecordKey(msg.Key); err != nil {
		return err
	}
	if err := validateRecordValue(msg.Value); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.WriterAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid writer address (%s)", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.OwnerAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address (%s)", err)
	}
	if msg.FeePayerAddress != "" {
		if _, err := sdk.AccAddressFromBech32(msg.FeePayerAddress); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid feePayer address (%s)", err)
		}
	}
	return nil
}
