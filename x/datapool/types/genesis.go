package types

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		NextPoolNumber:             uint64(1),
		Pools:                      []Pool{},
		Params:                     DefaultParams(),
		DataPassRedeemReceipts:     []DataPassRedeemReceipt{},
		InstantRevenueDistribution: InstantRevenueDistribution{},
		SalesHistories:             []*SalesHistory{},
		DataPassRedeemHistories:    []DataPassRedeemHistory{},
		// this line is used by starport scaffolding # ibc/genesistype/default
		// this line is used by starport scaffolding # genesis/types/default
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// this line is used by starport scaffolding # ibc/genesistype/validate

	// this line is used by starport scaffolding # genesis/types/validate

	return nil
}
