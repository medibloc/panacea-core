package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgRegisterDataValidator{}

func NewMsgRegisterDataValidator(dataValidator *DataValidator) *MsgRegisterDataValidator {
	return &MsgRegisterDataValidator{
		ValidatorDetail: dataValidator,
	}
}

func (msg *MsgRegisterDataValidator) Route() string {
	return RouterKey
}

func (msg *MsgRegisterDataValidator) Type() string {
	return "RegisterDataValidator"
}

func (msg *MsgRegisterDataValidator) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.ValidatorDetail.Address)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid data validator address (%s)", err)
	}

	if msg.ValidatorDetail.Endpoint == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty data validator endpoint URL")
	}
	return nil
}

func (msg *MsgRegisterDataValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRegisterDataValidator) GetSigners() []sdk.AccAddress {
	dataValidator, err := sdk.AccAddressFromBech32(msg.ValidatorDetail.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{dataValidator}
}

var _ sdk.Msg = &MsgUpdateDataValidator{}

func NewMsgUpdateDataValidator(address, endpoint string) *MsgUpdateDataValidator {
	return &MsgUpdateDataValidator{
		DataValidator: address,
		Endpoint:      endpoint,
	}
}

func (msg *MsgUpdateDataValidator) Route() string {
	return RouterKey
}

func (msg *MsgUpdateDataValidator) Type() string {
	return "UpdateDataValidator"
}

func (msg *MsgUpdateDataValidator) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.DataValidator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid data validator address (%s)", err)
	}

	if msg.Endpoint == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty data validator endpoint URL")
	}
	return nil
}

func (msg *MsgUpdateDataValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateDataValidator) GetSigners() []sdk.AccAddress {
	dataValidator, err := sdk.AccAddressFromBech32(msg.DataValidator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{dataValidator}
}

var _ sdk.Msg = &MsgCreatePool{}

func NewMsgCreatePool(poolParams *PoolParams, deposit sdk.Coin, curator string) *MsgCreatePool {
	return &MsgCreatePool{
		Curator:    curator,
		Deposit:    deposit,
		PoolParams: poolParams,
	}
}

func (msg *MsgCreatePool) Route() string {
	return RouterKey
}

func (msg *MsgCreatePool) Type() string {
	return "CreatePool"
}

func (msg *MsgCreatePool) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Curator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid curator address (%s)", err)
	}

	poolParams := msg.PoolParams

	if len(poolParams.DataSchema) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "no data schema")
	}

	if poolParams.TargetNumData <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "target number of data should be greater than 0")
	}

	if poolParams.MaxNftSupply <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "maximum supply of NFT should be greater than 0")
	}

	if poolParams.NftPrice == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "the price of NFT is empty")
	}
	if !poolParams.NftPrice.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "the price of NFT is invalid")
	}

	for _, validatorAddr := range poolParams.TrustedDataValidators {
		_, err = sdk.AccAddressFromBech32(validatorAddr)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid data validator address of %s", validatorAddr)
		}
	}

	for _, issuerAddr := range poolParams.TrustedDataIssuers {
		_, err = sdk.AccAddressFromBech32(issuerAddr)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid issuer address of %s", issuerAddr)
		}
	}

	return nil
}

func (msg *MsgCreatePool) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreatePool) GetSigners() []sdk.AccAddress {
	curator, err := sdk.AccAddressFromBech32(msg.Curator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{curator}
}

var _ sdk.Msg = &MsgSellData{}

func (msg *MsgSellData) Route() string {
	return RouterKey
}

func (msg *MsgSellData) Type() string {
	return "SellData"
}

func (msg *MsgSellData) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Seller)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid seller address (%s)", err)
	}
	if msg.Cert == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "certificate is nil")
	}

	cert := msg.Cert
	if cert.UnsignedCert == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "unsignedCertificate is nil")
	} else if cert.Signature == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "signature is nil")
	}

	unsignedCert := cert.UnsignedCert
	if unsignedCert.DataHash == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "dataHash is nil")
	} else if unsignedCert.DataValidator == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty data validator address")
	} else if unsignedCert.Requester == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty requester address")
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

var _ sdk.Msg = &MsgBuyDataPass{}

func (msg *MsgBuyDataPass) Route() string {
	return RouterKey
}

func (msg *MsgBuyDataPass) Type() string {
	return "BuyDataPass"
}

func (msg *MsgBuyDataPass) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Buyer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid buyer address (%s)", err)
	}

	if !msg.Payment.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid payment")
	}

	return nil
}

func (msg *MsgBuyDataPass) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgBuyDataPass) GetSigners() []sdk.AccAddress {
	buyer, err := sdk.AccAddressFromBech32(msg.Buyer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{buyer}
}

var _ sdk.Msg = &MsgRedeemDataAccessNFT{}

func (msg *MsgRedeemDataAccessNFT) Route() string {
	return RouterKey
}

func (msg *MsgRedeemDataAccessNFT) Type() string {
	return "RedeemDataAccessNFT"
}

func (msg *MsgRedeemDataAccessNFT) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Redeemer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid redeemer address (%s)", err)
	}
	return nil
}

func (msg *MsgRedeemDataAccessNFT) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRedeemDataAccessNFT) GetSigners() []sdk.AccAddress {
	redeemer, err := sdk.AccAddressFromBech32(msg.Redeemer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{redeemer}
}
