package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (m *Consent) ValidateBasic() error {
	if m.Certificate == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "certificate is empty")
	}

	if m.Certificate.UnsignedCertificate == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "unsignedCertificate is empty")
	}

	if err := m.Certificate.UnsignedCertificate.ValidateBasic(); err != nil {
		return sdkerrors.Wrapf(err, "failed to validation unsignedCertificate")
	}

	if m.Certificate.Signature == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "signature is empty")
	}

	if m.Certificate.UnsignedCertificate.DealId != m.DealId {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "the certificate is not for the deal %v", m.DealId)
	}

	return nil
}

func (m *UnsignedCertificate) ValidateBasic() error {
	if len(m.Cid) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "cid is empty")
	}

	if len(m.UniqueId) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "uniqueId is empty")
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