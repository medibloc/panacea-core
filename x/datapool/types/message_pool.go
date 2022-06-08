package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

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

	if len(poolParams.TrustedOracles) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "at least 1 trusted oracle is required")
	}

	for _, oracleAddr := range poolParams.TrustedOracles {
		_, err = sdk.AccAddressFromBech32(oracleAddr)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid oracle address of %s", oracleAddr)
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
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "unsignedCert is nil")
	} else if cert.Signature == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "signature is nil")
	}

	unsignedCert := cert.UnsignedCert
	if unsignedCert.DataHash == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "dataHash is nil")
	} else if unsignedCert.Oracle == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty oracle address")
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

var _ sdk.Msg = &MsgRedeemDataPass{}

func NewMsgRedeemDataPass(poolID, round, dataPassID uint64, redeemer string) *MsgRedeemDataPass {
	return &MsgRedeemDataPass{
		PoolId:     poolID,
		Round:      round,
		DataPassId: dataPassID,
		Redeemer:   redeemer,
	}
}

func (msg *MsgRedeemDataPass) Route() string {
	return RouterKey
}

func (msg *MsgRedeemDataPass) Type() string {
	return "RedeemDataPass"
}

func (msg *MsgRedeemDataPass) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Redeemer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid redeemer address (%s)", err)
	}

	if msg.PoolId <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid pool ID (%s)", err)
	}

	if msg.Round <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid round (%s)", err)
	}

	if msg.DataPassId <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid NFT ID (%s)", err)
	}

	return nil
}

func (msg *MsgRedeemDataPass) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRedeemDataPass) GetSigners() []sdk.AccAddress {
	redeemer, err := sdk.AccAddressFromBech32(msg.Redeemer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{redeemer}
}
