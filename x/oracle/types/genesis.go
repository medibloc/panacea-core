package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Oracles:                    []Oracle{},
		OracleRegistrations:        []OracleRegistration{},
		Params:                     DefaultParams(),
		OracleUpgradeQueueElements: nil,
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
func (m *GenesisState) Validate() error {

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

	if err := m.Params.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if m.OracleUpgradeQueueElements != nil {
		for _, address := range m.OracleUpgradeQueueElements {
			if len(address) == 0 {
				return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "address is empty")
			}
		}
	}

	return nil
}
