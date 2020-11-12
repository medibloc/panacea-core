package types

type GenesisState struct {
	Tokens map[string]Token `json:"tokens"`
}

func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

func ValidateGenesis(data GenesisState) error {
	for bz, token := range data.Tokens {
		var key GenesisTokenKey
		if err := key.Unmarshal(bz); err != nil {
			return err
		}

		if err := token.validate(); err != nil {
			return err
		}
	}
	return nil
}
