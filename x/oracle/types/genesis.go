package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	for _, oracle := range gs.Oracles {
		oracleAddr, err := sdk.AccAddressFromBech32(oracle.Address)
		if err != nil {
			return sdkerrors.Wrapf(ErrInvalidOracleAccAddress, "oracle address: %s", oracleAddr.String())
		}
	}
	return nil
}
