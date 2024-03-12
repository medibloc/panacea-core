package types

func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Denoms: []*Denom{},
		Pnfts:  []*Pnft{},
	}
}

func (data GenesisState) ValidateBasic() error {
	for _, denom := range data.Denoms {
		if err := denom.ValidateBasic(); err != nil {
			return err
		}
	}

	for _, pnft := range data.Pnfts {
		if err := pnft.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}
