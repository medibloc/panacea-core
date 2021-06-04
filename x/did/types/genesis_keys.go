package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

type GenesisDIDDocumentKey struct {
	DID string `json:"did"`
}

func (k GenesisDIDDocumentKey) Marshal() string {
	return k.DID
}

func (k *GenesisDIDDocumentKey) Unmarshal(key string) error {
	did := key
	if !ValidateDID(did) {
		return sdkerrors.Wrapf(ErrInvalidDID, "DID: %s", key)
	}

	k.DID = did
	return nil
}
