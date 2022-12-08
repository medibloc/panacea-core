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
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "dataSchema is empty")
	}
	if m.MaxNumData <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "MaxNumData should be bigger than 0")
	}
	if m.Budget == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "budget is empty")
	}
	if !m.Budget.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "budget is not a valid Coin object")
	}
	if _, err := sdk.AccAddressFromBech32(m.ConsumerAddress); err != nil {
		return sdkerrors.Wrapf(err, "consumer address is invalid. address: %s", m.ConsumerAddress)
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

func (m *Certificate) ValidateBasic() error {
	if m.UnsignedCertificate == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "unsignedCertificate is empty")
	}

	if err := m.UnsignedCertificate.ValidateBasic(); err != nil {
		return sdkerrors.Wrapf(err, "failed to validation unsignedCertificate")
	}

	if m.Signature == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "signature is empty")
	}

	return nil
}

func (m *UnsignedCertificate) ValidateBasic() error {
	if len(m.Cid) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "cid is empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.OracleAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "oracleAddress is invalid. address: %s, error: %s", m.OracleAddress, err.Error())
	}

	if m.DealId <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "dealId is greater than 0")
	}

	if _, err := sdk.AccAddressFromBech32(m.ProviderAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "providerAddress is invalid. address: %s, error: %s", m.OracleAddress, err.Error())
	}

	if len(m.DataHash) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "dataHash is empty")
	}

	return nil
}

var _ sdk.Msg = &MsgDeactivateDeal{}

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
