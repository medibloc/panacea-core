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
		if m.VerificationTallyResult.Yes.IsNegative() {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "yes in TallyResult must not be negative: %s", m.VerificationTallyResult.Yes)
		}

		if m.VerificationTallyResult.No.IsNegative() {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "no in TallyResult must not be negative: %s", m.VerificationTallyResult.Yes)
		}

		if len(m.VerificationTallyResult.InvalidYes) > 0 {
			for _, invalidYes := range m.VerificationTallyResult.InvalidYes {
				if invalidYes.ConsensusValue == nil {
					return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalidConsensusValue in ConsensusValue must not be nil")
				}
				if invalidYes.VotingAmount.IsNegative() {
					return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "votingAmount in ConsensusValue must not be negative: %s", m.VerificationTallyResult.Yes)
				}
			}
		}
		if m.VerificationTallyResult.InvalidYes == nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalidYes in TallyResult must not be negative: %s", m.VerificationTallyResult.Yes)
		}
	}

	if m.DeliveryTallyResult != nil {
		if m.DeliveryTallyResult.Yes.IsNegative() {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "yes in TallyResult must not be negative: %s", m.DeliveryTallyResult.Yes)
		}

		if m.DeliveryTallyResult.No.IsNegative() {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "no in TallyResult must not be negative: %s", m.DeliveryTallyResult.Yes)
		}

		if len(m.DeliveryTallyResult.InvalidYes) > 0 {
			for _, invalidYes := range m.DeliveryTallyResult.InvalidYes {
				if invalidYes.ConsensusValue == nil {
					return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalidConsensusValue in ConsensusValue must not be nil")
				}
				if invalidYes.VotingAmount.IsNegative() {
					return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "votingAmount in ConsensusValue must not be negative: %s", m.DeliveryTallyResult.Yes)
				}
			}
		}
		if m.DeliveryTallyResult.InvalidYes == nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalidYes in TallyResult must not be negative: %s", m.DeliveryTallyResult.Yes)
		}
	}

	return nil
}

func (m DataDeliveryVote) ValidateBasic() error {
	if m.DealId <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "dealID should be bigger than 0")
	}

	if _, err := sdk.AccAddressFromBech32(m.VoterAddress); err != nil {
		return sdkerrors.Wrapf(err, "voterAddress is invalid. address: %s", m.VoterAddress)
	}

	if len(m.VerifiableCid) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "VerifiableCid is empty")
	}

	if len(m.DeliveredCid) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Delivered is empty")
	}

	if err := m.VoteOption.ValidateBasic(); err != nil {
		return err
	}

	return nil
}
