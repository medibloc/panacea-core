package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeOracleUpgrade string = "OracleUpgrade"
)

func NewOracleUpgradeProposal(title, description string, plan Plan) govtypes.Content {
	return &OracleUpgradeProposal{
		Title:       title,
		Description: description,
		Plan:        plan,
	}
}

var _ govtypes.Content = &OracleUpgradeProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeOracleUpgrade)
	govtypes.RegisterProposalTypeCodec(&OracleUpgradeProposal{}, "panacea/SoftwareUpgradeProposal")
}

func (m *OracleUpgradeProposal) GetTitle() string {
	return m.Title
}

func (m *OracleUpgradeProposal) GetDescription() string {
	return m.Description
}

func (m *OracleUpgradeProposal) ProposalRoute() string {
	return RouterKey
}

func (m *OracleUpgradeProposal) ProposalType() string {
	return ProposalTypeOracleUpgrade
}

func (m *OracleUpgradeProposal) ValidateBasic() error {
	if err := m.Plan.ValidateBasic(); err != nil {
		return err
	}
	return govtypes.ValidateAbstract(m)
}

func (p Plan) ValidateBasic() error {
	if len(p.UniqueId) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "uniqueID cannot be empty")
	}
	if p.Height <= 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "height must be greater than 0")
	}
	return nil
}
