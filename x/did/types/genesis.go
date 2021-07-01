package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (data GenesisState) Validate() error {
	for bz, doc := range data.Documents {
		var key GenesisDIDDocumentKey
		if err := key.Unmarshal(bz); err != nil {
			return err
		}

		if !doc.Valid() {
			return sdkerrors.Wrapf(ErrInvalidDIDDocumentWithSeq, "DIDDocumentWithSeq: %v", doc)
		}
	}
	return nil
}
