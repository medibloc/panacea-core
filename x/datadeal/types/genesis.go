package types

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Deals:          map[uint64]Deal{},
		NextDealNumber: uint64(1),
		DataSales:      []DataSale{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for _, dataSale := range gs.DataSales {
		if err := dataSale.ValidateBasic(); err != nil {
			return err
		}
	}
	return nil
}
