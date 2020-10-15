package types

type GenesisDIDDocumentKey struct {
	DID DID `json:"did"`
}

func (k GenesisDIDDocumentKey) Marshal() string {
	return string(k.DID)
}

func (k *GenesisDIDDocumentKey) Unmarshal(key string) error {
	did := DID(key)
	if !did.Valid() {
		return ErrInvalidDID(key)
	}

	k.DID = did
	return nil
}
