package types

type GenesisTokenKey struct {
	Symbol Symbol `json:"symbol"`
}

func (k GenesisTokenKey) Marshal() string {
	return string(k.Symbol)
}

func (k *GenesisTokenKey) Unmarshal(key string) error {
	symbol := Symbol(key)
	if err := symbol.validate(); err != nil {
		return err
	}
	k.Symbol = symbol
	return nil
}
