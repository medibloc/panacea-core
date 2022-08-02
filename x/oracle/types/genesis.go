package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Oracles:                 []Oracle{},
		OracleRegistrations:     []OracleRegistration{},
		OracleRegistrationVotes: []OracleRegistrationVote{},
		Params:                  Params{},
		UpgradeOracleInfo:       UpgradeOracleInfo{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (m GenesisState) Validate() error {
	for _, oracle := range m.Oracles {
		if err := oracle.ValidateBasic(); err != nil {
			return err
		}
	}
	for _, oracleRegistration := range m.OracleRegistrations {
		if err := oracleRegistration.ValidateBasic(); err != nil {
			return err
		}
	}

	for _, oracleRegistrationVote := range m.OracleRegistrationVotes {
		if err := oracleRegistrationVote.ValidateBasic(); err != nil {
			return err
		}
	}

	if err := m.Params.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
	}

	return nil
}
