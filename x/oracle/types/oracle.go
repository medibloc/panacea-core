package types

import (
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type OracleKeeper interface {
	VoteOracleRegistration(sdk.Context, *OracleRegistrationVote, []byte) error

	GetAllOracleList(sdk.Context) ([]Oracle, error)

	GetOracle(sdk.Context, string) (*Oracle, error)

	SetOracle(sdk.Context, *Oracle) error

	GetAllOracleRegistrationList(sdk.Context) ([]OracleRegistration, error)

	GetOracleRegistration(sdk.Context, string, string) (*OracleRegistration, error)

	SetOracleRegistration(sdk.Context, *OracleRegistration) error

	GetAllOracleRegistrationVoteList(sdk.Context) ([]OracleRegistrationVote, error)

	GetOracleRegistrationVoteIterator(sdk.Context, string, string) sdk.Iterator

	GetOracleRegistrationVote(sdk.Context, string, string, string) (*OracleRegistrationVote, error)

	SetOracleRegistrationVote(sdk.Context, *OracleRegistrationVote) error

	GetParams(sdk.Context) Params

	SetParams(sdk.Context, Params)
}

func (m Oracle) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Address); err != nil {
		return sdkerrors.Wrapf(err, "oracle address is invalid. address: %s", m.Address)
	}
	return nil
}

func (m Oracle) IsActivated() bool {
	return m.Status == ORACLE_STATUS_ACTIVE
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

func (m OracleRegistration) MustGetOracleAccAddress() sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		panic(fmt.Sprintf("failed convert address to AccAddress. address: %s, error: %v", m.Address, err))
	}
	return accAddr
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

func (m OracleRegistrationVote) GetConsensusValue() []byte {
	return m.EncryptedOraclePrivKey
}

func (m VoteOption) ValidateBasic() error {
	if m == VOTE_OPTION_VALID ||
		m == VOTE_OPTION_INVALID {
		return nil
	}

	return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "voteOption is invalid")
}

func NewTallyResult() *TallyResult {
	return &TallyResult{
		Yes:            sdk.ZeroInt(),
		No:             sdk.ZeroInt(),
		InvalidYes:     sdk.ZeroInt(),
		ConsensusValue: nil,
		Total:          sdk.ZeroInt(),
	}
}
