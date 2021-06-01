package types

type GenesisDIDDocumentKey struct {
	DID string `json:"did"`
}

func (k GenesisDIDDocumentKey) Marshal() string {
	return k.DID
}

func (k *GenesisDIDDocumentKey) Unmarshal(key string) error {
	did := key
	if !ValidDID(did) {
		return Error(ErrInvalidDID, key)
	}

	k.DID = did
	return nil
}
