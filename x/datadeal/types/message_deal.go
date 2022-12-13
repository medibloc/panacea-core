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
	if len(m.DataSchema) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "no data schema")
	}
	if m.MaxNumData <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "max num of data is negative number")
	}
	if m.Budget == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "budget is empty")
	}
	if !m.Budget.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "budget is not a valid Coin object")
	}
	if _, err := sdk.AccAddressFromBech32(m.ConsumerAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid consumer address (%s)", err)
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

var _ sdk.Msg = &MsgSubmitConsent{}

func NewMsgSubmitConsent(dataCert *Certificate) *MsgSubmitConsent {
	return &MsgSubmitConsent{
		Certificate: dataCert,
	}
}
func (m *MsgSubmitConsent) Route() string {
	return RouterKey
}

func (m *MsgSubmitConsent) Type() string {
	return "SubmitConsent"
}

func (m *MsgSubmitConsent) ValidateBasic() error {
	if m.Certificate == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "certificate is empty")
	}

	if err := m.Certificate.ValidateBasic(); err != nil {
		return sdkerrors.Wrapf(err, "failed to validation certificate")
	}

	return nil
}

func (m *MsgSubmitConsent) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgSubmitConsent) GetSigners() []sdk.AccAddress {
	oracleAddress, err := sdk.AccAddressFromBech32(m.Certificate.UnsignedCertificate.OracleAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{oracleAddress}
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

	if _, err := sdk.AccAddressFromBech32(m.RequesterAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "requesterAddress is invalid. address: %s, error: %s", m.RequesterAddress, err.Error())
	}
	if m.DealId <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "dealId is greater than 0")
	}

	return nil
}

func (m *MsgDeactivateDeal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgDeactivateDeal) GetSigners() []sdk.AccAddress {
	requesterAddress, err := sdk.AccAddressFromBech32(m.RequesterAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{requesterAddress}
}
