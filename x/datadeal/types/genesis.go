package types

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		NextDealNumber: uint64(1),
		Deals:          []Deal{},
		Certificates:   []Certificate{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (m GenesisState) Validate() error {

	return nil
}
