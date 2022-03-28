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

var _ sdk.Msg = &MsgCreatePool{}

func NewMsgCreatePool(poolParams *PoolParams, curator string) *MsgCreatePool {
	params := &PoolParams{
		DataSchema:            poolParams.DataSchema,
		TargetNumData:         poolParams.TargetNumData,
		MaxNftSupply:          poolParams.MaxNftSupply,
		NftPrice:              poolParams.NftPrice,
		TrustedDataValidators: poolParams.TrustedDataValidators,
		TrustedDataIssuers:    poolParams.TrustedDataIssuers,
		Deposit:               poolParams.Deposit,
		DownloadPeriod:        poolParams.DownloadPeriod,
	}

	return &MsgCreatePool{
		Curator:    curator,
		PoolParams: params,
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

	if poolParams.Deposit == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "the deposit is empty")
	}
	if !poolParams.Deposit.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "the deposit is invalid")
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

var _ sdk.Msg = &MsgBuyDataAccessNFT{}

func (msg *MsgBuyDataAccessNFT) Route() string {
	return RouterKey
}

func (msg *MsgBuyDataAccessNFT) Type() string {
	return "BuyDataAccessNFT"
}

func (msg *MsgBuyDataAccessNFT) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Buyer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid buyer address (%s)", err)
	}
	return nil
}

func (msg *MsgBuyDataAccessNFT) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgBuyDataAccessNFT) GetSigners() []sdk.AccAddress {
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

var _ sdk.Msg = &MsgRegisterNFTContract{}

func NewMsgRegisterNFTContract(wasmCode []byte, sender string) *MsgRegisterNFTContract {
	return &MsgRegisterNFTContract{
		WasmCode: wasmCode,
		Sender:   sender,
	}
}

func (msg *MsgRegisterNFTContract) Route() string {
	return RouterKey
}

func (msg *MsgRegisterNFTContract) Type() string {
	return "RegisterNFTContract"
}

func (msg *MsgRegisterNFTContract) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid redeemer address (%s)", err)
	}

	return nil
}

func (msg *MsgRegisterNFTContract) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRegisterNFTContract) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgUpgradeNFTContract{}

func NewMsgUpgradeNFTContract(newWasmCode []byte, sender string) *MsgUpgradeNFTContract {
	return &MsgUpgradeNFTContract{
		NewWasmCode: newWasmCode,
		Sender:      sender,
	}
}

func (msg *MsgUpgradeNFTContract) Route() string {
	return RouterKey
}

func (msg *MsgUpgradeNFTContract) Type() string {
	return "UpgradeNFTContract"
}

func (msg *MsgUpgradeNFTContract) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid redeemer address (%s)", err)
	}

	return nil
}

func (msg *MsgUpgradeNFTContract) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpgradeNFTContract) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}