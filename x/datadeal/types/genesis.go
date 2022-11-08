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

	for _, deal := range gs.Deals {
		if err := deal.ValidateBasic(); err != nil {
			return err
		}
	}

	for _, dataVerificationVote := range gs.DataVerificationVotes {
		if err := dataVerificationVote.ValidateBasic(); err != nil {
			return err
		}
	}

	for _, dataDeliveryVote := range gs.DataDeliveryVotes {
		if err := dataDeliveryVote.ValidateBasic(); err != nil {
			return err
		}
	}

	for _, dataVerificationQueueElement := range gs.DataVerificationQueueElements {
		if err := dataVerificationQueueElement.ValidateBasic(); err != nil {
			return err
		}
	}

	for _, dataDeliveryQueueElement := range gs.DataDeliveryQueueElements {
		if err := dataDeliveryQueueElement.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}
