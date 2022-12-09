package types

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Deals:                         []Deal{},
		NextDealNumber:                uint64(1),
		DataSales:                     []DataSale{},
		DataVerificationVotes:         []DataVerificationVote{},
		DataDeliveryVotes:             []DataDeliveryVote{},
		DataVerificationQueueElements: []DataVerificationQueueElement{},
		DataDeliveryQueueElements:     []DataDeliveryQueueElement{},
		Params:                        DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (m GenesisState) Validate() error {
	for _, dataSale := range m.DataSales {
		if err := dataSale.ValidateBasic(); err != nil {
			return err
		}
	}

	for _, deal := range m.Deals {
		if err := deal.ValidateBasic(); err != nil {
			return err
		}
	}

	for _, dataVerificationVote := range m.DataVerificationVotes {
		if err := dataVerificationVote.ValidateBasic(); err != nil {
			return err
		}
	}

	for _, dataDeliveryVote := range m.DataDeliveryVotes {
		if err := dataDeliveryVote.ValidateBasic(); err != nil {
			return err
		}
	}

	for _, dataVerificationQueueElement := range m.DataVerificationQueueElements {
		if err := dataVerificationQueueElement.ValidateBasic(); err != nil {
			return err
		}
	}

	for _, dataDeliveryQueueElement := range m.DataDeliveryQueueElements {
		if err := dataDeliveryQueueElement.ValidateBasic(); err != nil {
			return err
		}
	}

	if err := m.Params.Validate(); err != nil {
		return err
	}

	return nil
}
