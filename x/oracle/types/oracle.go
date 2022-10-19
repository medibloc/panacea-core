package types

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const NonceSize = 12

func NewOracle(address string, status OracleStatus) *Oracle {
	return &Oracle{
		Address: address,
		Status:  status,
	}
}

func NewOracleRegistration(msg *MsgRegisterOracle) *OracleRegistration {
	return &OracleRegistration{
		UniqueId:               msg.UniqueId,
		Address:                msg.OracleAddress,
		NodePubKey:             msg.NodePubKey,
		NodePubKeyRemoteReport: msg.NodePubKeyRemoteReport,
		TrustedBlockHeight:     msg.TrustedBlockHeight,
		TrustedBlockHash:       msg.TrustedBlockHash,
		Status:                 ORACLE_REGISTRATION_STATUS_VOTING_PERIOD,
		Nonce:                  msg.Nonce,
		RegistrationType:       ORACLE_REGISTRATION_TYPE_NEW,
	}
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

	if err := validateNodeKey(m.NodePubKey); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if err := validateNodePubKeyRemoteReport(m.NodePubKeyRemoteReport); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if err := validateTrustedBlockHeight(m.TrustedBlockHeight); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if err := validateTrustedBlockHash(m.TrustedBlockHash); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, err.Error())
	}

	if m.VotingPeriod == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "votingPeriod is nil")
	}

	if m.TallyResult != nil {
		if err := m.TallyResult.ValidateBasic(); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
		}
	}

	if len(m.Nonce) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "nonce is empty")
	} else if len(m.Nonce) != NonceSize {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "nonce length must be %v", NonceSize)
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

	if _, err := sdk.AccAddressFromBech32(m.VoterAddress); err != nil {
		return sdkerrors.Wrapf(err, "voterAddress is invalid. address: %s", m.VoterAddress)
	}

	if _, err := sdk.AccAddressFromBech32(m.VotingTargetAddress); err != nil {
		return sdkerrors.Wrapf(err, "votingTargetAddress is invalid. address: %s", m.VotingTargetAddress)
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

func validateNodeKey(nodePubKey []byte) error {
	if nodePubKey == nil {
		return ErrEmptyNodePubKey
	}

	if nodePubKey != nil {
		if _, err := btcec.ParsePubKey(nodePubKey, btcec.S256()); err != nil {
			return ErrInvalidNodePubKey
		}
	}

	return nil
}

func validateNodePubKeyRemoteReport(nodePubKeyRemoteReport []byte) error {
	if nodePubKeyRemoteReport == nil {
		return ErrEmptyNodePubKeyRemoteReport
	}

	return nil
}

func validateTrustedBlockHeight(height int64) error {
	if height <= 0 {
		return ErrInvalidTrustedBlockHeight
	}

	return nil
}

func validateTrustedBlockHash(hash []byte) error {
	if hash == nil {
		return ErrTrustedBlockHashNil
	}

	return nil
}

func (m VoteOption) ValidateBasic() error {
	if m == VOTE_OPTION_YES ||
		m == VOTE_OPTION_NO {
		return nil
	}

	return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "voteOption is invalid")
}

func NewTallyResult() *TallyResult {
	return &TallyResult{
		Yes:            sdk.ZeroInt(),
		No:             sdk.ZeroInt(),
		InvalidYes:     make([]*ConsensusTally, 0),
		ConsensusValue: nil,
		Total:          sdk.ZeroInt(),
		ValidVoters:    make([]*VoterInfo, 0),
	}
}

func (t *TallyResult) AddInvalidYes(tally *ConsensusTally) {
	t.InvalidYes = append(t.InvalidYes, tally)
}

func (t TallyResult) IsPassed() bool {
	return t.ConsensusValue != nil
}

func (t TallyResult) ValidateBasic() error {
	if t.Yes.IsNegative() {
		return fmt.Errorf("yes in TallyResult must not be negative: %s", t.Yes)
	}

	if t.No.IsNegative() {
		return fmt.Errorf("no in TallyResult must not be negative: %s", t.Yes)
	}

	if t.Total.IsNegative() {
		return fmt.Errorf("total in TallyResult must not be negative: %s", t.Total)
	}

	if len(t.InvalidYes) > 0 {
		for _, invalidYes := range t.InvalidYes {
			if invalidYes.ConsensusValue == nil {
				return fmt.Errorf("invalidConsensusValue in ConsensusValue must not be nil")
			}
			if invalidYes.VotingAmount.IsNegative() {
				return fmt.Errorf("votingAmount in ConsensusValue must not be negative: %s", t.Yes)
			}
		}
	}

	if t.InvalidYes == nil {
		return fmt.Errorf("invalidYes in TallyResult cannot be nil")
	}

	if len(t.ValidVoters) > 0 {
		for _, voter := range t.ValidVoters {
			if _, err := sdk.AccAddressFromBech32(voter.VoterAddress); err != nil {
				return fmt.Errorf("invalid voter address: %s", voter.VoterAddress)
			}
			if voter.VotingPower.IsNegative() {
				return fmt.Errorf("voting power cannot be negetive: %s", voter.VotingPower)
			}
		}
	}

	if t.ValidVoters == nil {
		return fmt.Errorf("validVoters cannot be nil")
	}

	return nil
}

func (m *MsgUpgradeOracle) ToOracleRegistration() *OracleRegistration {
	return &OracleRegistration{
		UniqueId:               m.UniqueId,
		Address:                m.OracleAddress,
		NodePubKey:             m.NodePubKey,
		NodePubKeyRemoteReport: m.NodePubKeyRemoteReport,
		TrustedBlockHeight:     m.TrustedBlockHeight,
		TrustedBlockHash:       m.TrustedBlockHash,
		Status:                 ORACLE_REGISTRATION_STATUS_VOTING_PERIOD,
		Nonce:                  m.Nonce,
		RegistrationType:       ORACLE_REGISTRATION_TYPE_UPGRADE,
	}
}

func (m *OracleUpgradeInfo) ShouldExecute(ctx sdk.Context) bool {
	if m.Height > 0 {
		return m.Height <= ctx.BlockHeight()
	}
	return false
}
