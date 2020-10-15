package types

type GenesisState struct {
	Documents map[string]DIDDocumentWithSeq
}

func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

func ValidateGenesis(data GenesisState) error {
	for bz, doc := range data.Documents {
		var key GenesisDIDDocumentKey
		if err := key.Unmarshal(bz); err != nil {
			return err
		}

		if !doc.Valid() {
			return ErrInvalidDIDDocumentWithSeq(doc)
		}
	}
	return nil
}
