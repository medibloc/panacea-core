package did

import (
	"github.com/medibloc/panacea-core/x/did/types"
)

type GenesisDIDDocumentKey struct {
	DID types.DID `json:"did"`
}

func (k GenesisDIDDocumentKey) Marshal() string {
	return string(k.DID)
}

func (k *GenesisDIDDocumentKey) Unmarshal(key string) error {
	did := types.DID(key)
	if !did.Valid() {
		return types.ErrInvalidDID(key)
	}

	k.DID = did
	return nil
}
