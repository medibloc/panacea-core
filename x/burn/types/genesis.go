package types

// DefaultIndex is the default burn global index.
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default burn genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	return nil
}
