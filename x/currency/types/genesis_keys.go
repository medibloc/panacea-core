package types

type GenesisIssuanceKey struct {
	Denom string `json:"denom"`
}

func (k GenesisIssuanceKey) Marshal() string {
	return k.Denom
}

func (k *GenesisIssuanceKey) Unmarshal(key string) error {
	k.Denom = key
	return nil
}
