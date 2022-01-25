package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateDeal{}

func NewMsgCreateDeal(dataSchema []string, budget *sdk.Coin, maxNumData uint64, trustedDataValidator []string, owner string) *MsgCreateDeal {
	return &MsgCreateDeal{
		DataSchema:            dataSchema,
		Budget:                budget,
		MaxNumData:            maxNumData,
		TrustedDataValidators: trustedDataValidator,
		Owner:                 owner,
	}
}

func (msg *MsgCreateDeal) Route() string {
	return RouterKey
}

func (msg *MsgCreateDeal) Type() string {
	return "CreateDeal"
}

// ValidateBasic is validation for MsgCreateDeal.
func (msg *MsgCreateDeal) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	schema := msg.DataSchema
	if len(schema) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "no data schema")
	}

	budget := msg.Budget
	if budget == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "budget is empty")
	}
	if !budget.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "budget is not a valid Coin object")
	}

	data := msg.MaxNumData
	if data <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "max num of data is negative number")
	}

	for _, validator := range msg.TrustedDataValidators {
		_, err = sdk.AccAddressFromBech32(validator)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid data validator address (%s)", err)
		}
	}
	return nil
}

func (msg *MsgCreateDeal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateDeal) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

var _ sdk.Msg = &MsgSellData{}

func NewMsgSellData(cert DataValidationCertificate, seller string) *MsgSellData {
	return &MsgSellData{
		Cert:   &cert,
		Seller: seller,
	}
}

func (msg *MsgSellData) Route() string {
	return RouterKey
}

func (msg *MsgSellData) Type() string {
	return "SellData"
}

// ValidateBasic is validation for MsgSellData.
func (msg *MsgSellData) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Seller)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.Cert == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty certificate")
	}

	signature := msg.Cert.Signature
	if signature == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty signature")
	}

	unsignedCert := msg.Cert.UnsignedCert
	if unsignedCert == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "no unsigned data certificate")
	}

	if unsignedCert.DealId <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid deal id format")
	}

	if unsignedCert.DataHash == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty data hash")
	}

	if unsignedCert.EncryptedDataUrl == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty encrypted data url")
	}

	if unsignedCert.DataValidatorAddress == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty data validator address")
	}
	_, err = sdk.AccAddressFromBech32(unsignedCert.DataValidatorAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid data validator address (%s)", err)
	}

	if unsignedCert.RequesterAddress == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty requester address")
	}
	_, err = sdk.AccAddressFromBech32(unsignedCert.RequesterAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid requester address (%s)", err)
	}

	return nil
}

func (msg *MsgSellData) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSellData) GetSigners() []sdk.AccAddress {
	seller, err := sdk.AccAddressFromBech32(msg.Seller)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{seller}
}
