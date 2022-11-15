package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Oracles:                             []Oracle{},
		OracleRegistrations:                 []OracleRegistration{},
		OracleRegistrationVotes:             []OracleRegistrationVote{},
		Params:                              DefaultParams(),
		OracleUpgradeInfo:                   nil,
		OracleRegistrationVoteQueueElements: nil,
	}
}

func GetGenesisStateFromAppState(cdc codec.Codec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
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

	if m.OracleUpgradeInfo != nil {
		if err := m.OracleUpgradeInfo.ValidateBasic(); err != nil {
			return err
		}
	}

	if m.OracleRegistrationVoteQueueElements != nil {
		for _, element := range m.OracleRegistrationVoteQueueElements {
			if err := element.ValidateBasic(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *OracleRegistrationVoteQueueElement) ValidateBasic() error {
	if len(m.UniqueId) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "uniqueID is empty")
	}

	if m.Address == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "address is empty")
	}

	if m.VotingEndTime.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "voting end time can not be 0")
	}
	return nil
}
