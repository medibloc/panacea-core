package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (m OracleRegistrationVote) ValidateBasic() error {
	if len(m.UniqueId) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "uniqueID is empty")
	}
	_, err := sdk.AccAddressFromBech32(m.VoterAddress)
	if err != nil {
		return err
	}
	_, err = sdk.AccAddressFromBech32(m.VotingTargetAddress)
	if err != nil {
		return err
	}
	if len(m.EncryptedOraclePrivKey) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "encryptedOraclePrivKey is empty")
	}
	if err := m.VoteOption.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

func (m VoteOption) ValidateBasic() error {
	if m == VOTE_OPTION_VALID ||
		m == VOTE_OPTION_INVALID {
		return nil
	}

	return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "voteOption is invalid")
}
