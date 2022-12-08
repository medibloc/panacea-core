package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateDeal{}

func (m *MsgCreateDeal) Route() string {
	return RouterKey
}

func (m *MsgCreateDeal) Type() string {
	return "CreateDeal"
}

func (m *MsgCreateDeal) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.ConsumerAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid consumer address (%s)", err)
	}

	schema := m.DataSchema
	if len(schema) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "no data schema")
	}

	budget := m.Budget
	if budget == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "budget is empty")
	}
	if !budget.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "budget is not a valid Coin object")
	}

	data := m.MaxNumData
	if data <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "max num of data is negative number")
	}
	return nil
}

func (m *MsgCreateDeal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgCreateDeal) GetSigners() []sdk.AccAddress {
	consumerAddress, err := sdk.AccAddressFromBech32(m.ConsumerAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{consumerAddress}
}
