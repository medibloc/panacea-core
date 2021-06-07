package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgIssueToken{}

func NewMsgIssueToken(name string, shortSymbol string, totalSupplyMicro sdk.Int, mintable bool, ownerAddress string) *MsgIssueToken {
	return &MsgIssueToken{
		Name:             name,
		ShortSymbol:      shortSymbol,
		TotalSupplyMicro: sdk.IntProto{Int: totalSupplyMicro},
		Mintable:         mintable,
		OwnerAddress:     ownerAddress,
	}
}

func (msg *MsgIssueToken) Route() string {
	return RouterKey
}

func (msg *MsgIssueToken) Type() string {
	return "IssueToken"
}

func (msg *MsgIssueToken) GetSigners() []sdk.AccAddress {
	ownerAddress, err := sdk.AccAddressFromBech32(msg.OwnerAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{ownerAddress}
}

func (msg *MsgIssueToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgIssueToken) ValidateBasic() error {
	if err := validateName(msg.Name); err != nil {
		return err
	}
	if err := validateShortSymbol(msg.ShortSymbol); err != nil {
		return err
	}
	if err := validateTotalSupplyAmount(msg.TotalSupplyMicro.Int); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.OwnerAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address (%s)", err)
	}
	return nil
}
