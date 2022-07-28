package types

import (
	"github.com/btcsuite/btcd/btcec"
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

	if _, err := sdk.AccAddressFromBech32(m.OracleRegistrationVote.VoterAddress); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(m.OracleRegistrationVote.VotingTargetAddress); err != nil {
		return err
	}

	if len(m.OracleRegistrationVote.EncryptedOraclePrivKey) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "encryptedOraclePrivKey is empty")
	}
	return nil
}

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

	}

	return nil
}

func (m OracleRegistrationVote) ValidateBasic() error {
	if len(m.UniqueId) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "uniqueID is empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.VoterAddress); err != nil {
		return sdkerrors.Wrapf(err, "voterAddress is invalid. address: %s", m.VoterAddress)
	}

	if _, err := sdk.AccAddressFromBech32(m.VotingTargetAddress); err != nil {
		return sdkerrors.Wrapf(err, "votingTargetAddress is invalid. address: %s", m.VotingTargetAddress)
	}

	if m.EncryptedOraclePrivKey == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "encryptedOraclePrivKey is empty")
	}

	return nil
}
