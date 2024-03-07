package types

func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Denoms: []*Denom{},
	}
}

func (data GenesisState) ValidateBasic() error {
	for _, denom := range data.Denoms {
		if err := denom.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}
