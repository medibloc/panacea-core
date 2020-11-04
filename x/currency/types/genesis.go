package types

type GenesisState struct {
	Issuances map[string]Issuance `json:"issuances"`
}

func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

func ValidateGenesis(data GenesisState) error {
	for bz, issuance := range data.Issuances {
		var key GenesisIssuanceKey
		if err := key.Unmarshal(bz); err != nil {
			return err
		}

		if err := issuance.Valid(); err != nil {
			return ErrInvalidIssuance(err)
		}
	}
	return nil
}
