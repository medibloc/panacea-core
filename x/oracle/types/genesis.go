package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Oracles: []Oracle{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for i, oracle := range gs.Oracles {
		if gs.Oracles[i].Address != oracle.Address {
			return sdkerrors.Wrapf(ErrInvalidGenesisOracle, "oracle address: %s", oracle.GetAddress())
		}
	}
	return nil
}
