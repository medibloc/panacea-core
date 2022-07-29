package types

import (
	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (m Oracle) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Address); err != nil {
		return sdkerrors.Wrapf(err, "oracle address is invalid. address: %s", m.Address)
	}
	return nil
}

func (m OracleRegistration) ValidateBasic() error {
	if len(m.UniqueId) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "uniqueID is empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Address); err != nil {
		return sdkerrors.Wrapf(err, "oracle address is invalid. address: %s", m.Address)
	}

	if m.NodePubKey == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "nodePubKey is empty")
	}
	if m.NodePubKey != nil {
		if _, err := btcec.ParsePubKey(m.NodePubKey, btcec.S256()); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "nodePubKey is invalid. %s", err.Error())
		}
	}

	if m.NodePubKeyRemoteReport == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "nodePubKeyRemoteReport is empty")
	}

	if m.TrustedBlockHeight <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "trustedBlockHeight must be greater than zero")
	}

	if m.TrustedBlockHash == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "trustedBlockHash is nil")
	}

	if m.VotingPeriod == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "votingPeriod is nil")
	}

	if m.TallyResult != nil {
		if m.TallyResult.Yes.IsNegative() {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "yes in TallyResult must not be negative: %s", m.TallyResult.Yes)
		}

		if m.TallyResult.No.IsNegative() {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "no in TallyResult must not be negative: %s", m.TallyResult.Yes)
		}

		if m.TallyResult.InvalidYes.IsNegative() {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalidYes in TallyResult must not be negative: %s", m.TallyResult.Yes)
		}

		if m.TallyResult.ConsensusValue == nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "consensusValue in TallyResult must not be nil: %s", m.TallyResult.Yes)
		}
	}

	return nil
}

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
