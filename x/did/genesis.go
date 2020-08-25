package did

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/did/types"
)

type GenesisState struct {
	Documents map[string]types.DIDDocument
}

func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	for bz, doc := range data.Documents {
		var key GenesisDIDDocumentKey
		if err := key.Unmarshal(bz); err != nil {
			panic(err)
		}
		k.SetDIDDocument(ctx, key.DID, doc)
	}
}

func ValidateGenesis(data GenesisState) error {
	for bz, doc := range data.Documents {
		var key GenesisDIDDocumentKey
		if err := key.Unmarshal(bz); err != nil {
			return err
		}

		if !doc.Valid() {
			return types.ErrInvalidDIDDocument(doc)
		}
	}
	return nil
}

func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	documentsMap := make(map[string]types.DIDDocument)

	for _, did := range k.ListDIDs(ctx) {
		key := GenesisDIDDocumentKey{DID: did}.Marshal()
		documentsMap[key] = k.GetDIDDocument(ctx, did)
	}

	return GenesisState{
		Documents: documentsMap,
	}
}
