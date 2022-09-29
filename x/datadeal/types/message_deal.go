package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateDeal{}

func (msg *MsgCreateDeal) Route() string {
	return RouterKey
}

func (msg *MsgCreateDeal) Type() string {
	return "CreateDeal"
}

// ValidateBasic is validation for MsgCreateDeal.
func (msg *MsgCreateDeal) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.BuyerAddress)
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

	return nil
}

func (msg *MsgCreateDeal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateDeal) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.BuyerAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

var _ sdk.Msg = &MsgSellData{}

func NewMsgSellData(dealID uint64, verifiableCID, sellerAddress string) *MsgSellData {
	return &MsgSellData{
		DealId:        dealID,
		VerifiableCid: verifiableCID,
		SellerAddress: sellerAddress,
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
	_, err := sdk.AccAddressFromBech32(msg.SellerAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid seller address (%s)", err)
	}

	if msg.DealId == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty deal ID")
	}

	if len(msg.VerifiableCid) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty verifiableCID")
	}

	return nil
}

func (msg *MsgSellData) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSellData) GetSigners() []sdk.AccAddress {
	seller, err := sdk.AccAddressFromBech32(msg.SellerAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{seller}
}

func NewMsgDeactivateDeal(dealID uint64, requesterAddress string) *MsgDeactivateDeal {
	return &MsgDeactivateDeal{
		DealId:           dealID,
		RequesterAddress: requesterAddress,
	}
}

var _ sdk.Msg = &MsgVoteDataVerification{}

func (msg *MsgVoteDataVerification) Route() string {
	return RouterKey
}

func (msg *MsgVoteDataVerification) Type() string {
	return "VoteDataVerification"
}

func (msg *MsgVoteDataVerification) ValidateBasic() error {
	//TODO implement me
	panic("implement me")
}

func (msg *MsgVoteDataVerification) GetSigners() []sdk.AccAddress {
	//TODO implement me
	panic("implement me")
}

var _ sdk.Msg = &MsgVoteDataDelivery{}

func (msg *MsgVoteDataDelivery) Route() string {
	return RouterKey
}

func (msg *MsgVoteDataDelivery) Type() string {
	return "VoteDataDelivery"
}

func (msg *MsgVoteDataDelivery) ValidateBasic() error {
	//TODO implement me
	panic("implement me")
}

func (msg *MsgVoteDataDelivery) GetSigners() []sdk.AccAddress {
	//TODO implement me
	panic("implement me")
}

func (msg *MsgDeactivateDeal) Route() string {
	return RouterKey
}

func (msg *MsgDeactivateDeal) Type() string {
	return "DeactivateDeal"
}

// ValidateBasic is validation for MsgCreateDeal.
func (msg *MsgDeactivateDeal) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.RequesterAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid requester address (%s)", err)
	}

	if msg.DealId <= 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid deal id format")
	}
	return nil
}

func (msg *MsgDeactivateDeal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeactivateDeal) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.RequesterAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}
