package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewDataSale(msg *MsgSellData) *DataSale {
	return &DataSale{
		SellerAddress: msg.SellerAddress,
		DealId:        msg.DealId,
		VerifiableCid: msg.VerifiableCid,
		Status:        DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD,
	}
}

func (m DataSale) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.SellerAddress); err != nil {
		return sdkerrors.Wrapf(err, "seller address is invalid. address: %s", m.SellerAddress)
	}

	if m.DealId == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "dealID should be bigger than 0")
	}

	if m.VerifiableCid == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "verifiable CID is empty")
	}

	if m.VerificationTallyResult != nil {
		if err := m.VerificationTallyResult.ValidateBasic(); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
		}
	}

	if m.DeliveryTallyResult != nil {
		if err := m.DeliveryTallyResult.ValidateBasic(); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
		}
	}

	return nil
}
