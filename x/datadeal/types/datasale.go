package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func NewDataSale(msg *MsgSellData) *DataSale {
	return &DataSale{
		SellerAddress: msg.SellerAddress,
		DealId:        msg.DealId,
		VerifiableCid: msg.VerifiableCid,
		DataHash:      msg.DataHash,
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

	if m.DataHash == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "dataHash is empty")
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

func (m DataVerificationVote) ValidateBasic() error {
	if m.DealId == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "dealID should be bigger than 0")
	}

	if _, err := sdk.AccAddressFromBech32(m.VoterAddress); err != nil {
		return sdkerrors.Wrapf(err, "voterAddress is invalid. address: %s", m.VoterAddress)
	}

	if len(m.DataHash) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "dataHash is empty")
	}

	if err := m.VoteOption.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

func (m DataVerificationVote) GetConsensusValue() []byte {
	return []byte(m.DataHash)
}

func (m DataDeliveryVote) ValidateBasic() error {
	if m.DealId == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "dealID can not be 0")
	}

	if _, err := sdk.AccAddressFromBech32(m.VoterAddress); err != nil {
		return sdkerrors.Wrapf(err, "voterAddress is invalid. address: %s", m.VoterAddress)
	}

	if len(m.DataHash) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "dataHash is empty")
	}

	if len(m.DeliveredCid) == 0 && m.VoteOption == oracletypes.VOTE_OPTION_YES {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "vote option is yes, but DeliveredCid is empty")
	}

	if err := m.VoteOption.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

func (m DataDeliveryVote) GetConsensusValue() []byte {
	return []byte(m.DeliveredCid)
}
