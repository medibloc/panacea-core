package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
)

var _ sdk.Msg = &MsgCreateDeal{}

func (m *MsgCreateDeal) Route() string {
	return RouterKey
}

func (m *MsgCreateDeal) Type() string {
	return "CreateDeal"
}

// ValidateBasic is validation for MsgCreateDeal.
func (m *MsgCreateDeal) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.BuyerAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	schema := m.DataSchema
	if len(schema) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "no data schema")
	}

	budget := m.Budget
	if budget == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "budget is empty")
	}
	if !budget.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "budget is not a valid Coin object")
	}

	data := m.MaxNumData
	if data <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "max num of data is negative number")
	}

	return nil
}

func (m *MsgCreateDeal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgCreateDeal) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.BuyerAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

var _ sdk.Msg = &MsgSellData{}

func NewMsgSellData(dealID uint64, verifiableCID, dataHash, sellerAddress string) *MsgSellData {
	return &MsgSellData{
		DealId:        dealID,
		VerifiableCid: verifiableCID,
		DataHash:      dataHash,
		SellerAddress: sellerAddress,
	}
}

func (m *MsgSellData) Route() string {
	return RouterKey
}

func (m *MsgSellData) Type() string {
	return "SellData"
}

// ValidateBasic is validation for MsgSellData.
func (m *MsgSellData) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.SellerAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid seller address (%s)", err)
	}

	if m.DealId == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty deal ID")
	}

	if len(m.VerifiableCid) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty verifiableCID")
	}
	if len(m.DataHash) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty dataHash")
	}

	return nil
}

func (m *MsgSellData) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgSellData) GetSigners() []sdk.AccAddress {
	seller, err := sdk.AccAddressFromBech32(m.SellerAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{seller}
}

var _ sdk.Msg = &MsgVoteDataVerification{}

func (m *MsgVoteDataVerification) Route() string {
	return RouterKey
}

func (m *MsgVoteDataVerification) Type() string {
	return "VoteDataVerification"
}

func (m *MsgVoteDataVerification) ValidateBasic() error {
	if m.DataVerificationVote == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "dataVerificationVote cannot be nil")
	}

	if _, err := sdk.AccAddressFromBech32(m.DataVerificationVote.VoterAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid voter address (%s)", err)
	}
	if len(m.DataVerificationVote.DataHash) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "DataHash cannot be empty")
	}
	if m.DataVerificationVote.DealId == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "deal ID must be bigger than zero(0)")
	}

	if m.Signature == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "signature cannot be nil")
	}

	return nil
}

func (m *MsgVoteDataVerification) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgVoteDataVerification) GetSigners() []sdk.AccAddress {
	voterAddress, err := sdk.AccAddressFromBech32(m.DataVerificationVote.VoterAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{voterAddress}
}

var _ sdk.Msg = &MsgVoteDataDelivery{}

func (m *MsgVoteDataDelivery) Route() string {
	return RouterKey
}

func (m *MsgVoteDataDelivery) Type() string {
	return "VoteDataDelivery"
}

func (m *MsgVoteDataDelivery) ValidateBasic() error {
	if m.DataDeliveryVote == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "DataDeliveryVote cannot be nil")
	}

	if m.DataDeliveryVote.DealId == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "deal ID cannot be 0")
	}
	if _, err := sdk.AccAddressFromBech32(m.DataDeliveryVote.VoterAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid voter address (%s)", err)
	}
	if len(m.DataDeliveryVote.DeliveredCid) == 0 && m.DataDeliveryVote.VoteOption == oracletypes.VOTE_OPTION_YES {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Delivered Cid can not be empty when vote option is yes")
	}
	if len(m.DataDeliveryVote.DataHash) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "DataHash can not be empty")
	}
	if m.Signature == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "signature cannot be nil")
	}
	return nil
}

func (m *MsgVoteDataDelivery) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgVoteDataDelivery) GetSigners() []sdk.AccAddress {
	voterAddress, err := sdk.AccAddressFromBech32(m.DataDeliveryVote.VoterAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{voterAddress}
}

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
	_, err := sdk.AccAddressFromBech32(m.RequesterAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid requester address (%s)", err)
	}

	if m.DealId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid deal id format")
	}
	return nil
}

func (m *MsgDeactivateDeal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgDeactivateDeal) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.RequesterAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

var _ sdk.Msg = &MsgReRequestDataDeliveryVote{}

func NewMsgReRequestDataDeliveryVote(dealID uint64, dataHash string, requesterAddress string) *MsgReRequestDataDeliveryVote {
	return &MsgReRequestDataDeliveryVote{
		DealId:           dealID,
		DataHash:         dataHash,
		RequesterAddress: requesterAddress,
	}
}

func (m *MsgReRequestDataDeliveryVote) Route() string {
	return RouterKey
}

func (m *MsgReRequestDataDeliveryVote) Type() string {
	return "RequestDeliveredCid"
}

func (m *MsgReRequestDataDeliveryVote) ValidateBasic() error {

	_, err := sdk.AccAddressFromBech32(m.RequesterAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid requester address (%s)", err)
	}

	if m.DealId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid deal id format")
	}

	if len(m.DataHash) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "DataHash can not be empty")
	}
	return nil
}

func (m *MsgReRequestDataDeliveryVote) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgReRequestDataDeliveryVote) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.RequesterAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}
