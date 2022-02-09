package types

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Deals:            map[uint64]*Deal{},
		DataCertificates: map[string]*DataValidationCertificate{},
		NextDealNumber:   1,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
// TODO: Validate genesis state after implement GetDealList Deal.
func (gs GenesisState) Validate() error {
	return nil
}
