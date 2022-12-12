package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

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
