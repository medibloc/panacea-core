package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (m SignedOracleRegistrationVote) ValidateBasic() error {
	if m.OracleRegistrationVote == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "oracleRegistrationVote is empty")
	}

	if len(m.OracleRegistrationVote.UniqueId) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "uniqueID is empty")
	}
	if len(m.OracleRegistrationVote.UniqueId) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "uniqueID is empty")
	}
	_, err := sdk.AccAddressFromBech32(m.OracleRegistrationVote.VoterAddress)
	if err != nil {
		return err
	}
	_, err = sdk.AccAddressFromBech32(m.OracleRegistrationVote.VotingTargetAddress)
	if err != nil {
		return err
	}
	if len(m.OracleRegistrationVote.EncryptedOraclePrivKey) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "encryptedOraclePrivKey is empty")
	}
	return nil
}
