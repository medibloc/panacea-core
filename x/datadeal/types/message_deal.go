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

var _ sdk.Msg = &MsgDeactivateDeal{}

func NewMsgDeactivateDeal(dealID uint64, requesterAddress string) *MsgDeactivateDeal {
	return &MsgDeactivateDeal{
		DealId:           dealID,
		RequesterAddress: requesterAddress,
	}
}

func (m *MsgDeactivateDeal) Route() string {
	return RouterKey
}

func (m *MsgDeactivateDeal) Type() string {
	return "DeactivateDeal"
}

func (m *MsgDeactivateDeal) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.RequesterAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid requester address (%s)", err)
	}

	if m.DealId <= 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid deal id format")
	}
	return nil
}

func (m *MsgDeactivateDeal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgDeactivateDeal) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.RequesterAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}
