package types

import (
	"fmt"
	// this line is used by starport scaffolding # ibc/genesistype/import
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		// this line is used by starport scaffolding # ibc/genesistype/default
		// this line is used by starport scaffolding # genesis/types/default
		Tokens: map[string]*Token{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// this line is used by starport scaffolding # ibc/genesistype/validate

	// this line is used by starport scaffolding # genesis/types/validate
	// Check for duplicated ID in token
	tokenSymbolMap := make(map[string]bool)

	for _, token := range gs.Tokens {
		if _, ok := tokenSymbolMap[token.Symbol]; ok {
			return fmt.Errorf("duplicated symbol for token: %v", token)
		}

		if err := validateToken(token); err != nil {
			return fmt.Errorf("invalid token: %v", token)
		}

		tokenSymbolMap[token.Symbol] = true
	}

	return nil
}
