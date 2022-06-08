package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Oracles: map[string]Oracle{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for oracleMapKey, oracle := range gs.Oracles {
		oracleAddr, err := sdk.AccAddressFromBech32(oracle.GetAddress())
		if err != nil {
			return err
		}

		key := string(GetKeyPrefixOracle(oracleAddr))
		if oracleMapKey != key {
			return sdkerrors.Wrapf(ErrInvalidGenesisOracle, "oracle address: %s", oracle.GetAddress())
		}
	}
	return nil
}
