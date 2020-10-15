package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/did/types"
)

func InitGenesis(ctx sdk.Context, k Keeper, data types.GenesisState) {
	for bz, doc := range data.Documents {
		var key types.GenesisDIDDocumentKey
		if err := key.Unmarshal(bz); err != nil {
			panic(err)
		}
		k.SetDIDDocument(ctx, key.DID, doc)
	}
}

func ExportGenesis(ctx sdk.Context, k Keeper) types.GenesisState {
	documentsMap := make(map[string]types.DIDDocumentWithSeq)

	for _, did := range k.ListDIDs(ctx) {
		key := types.GenesisDIDDocumentKey{DID: did}.Marshal()
		documentsMap[key] = k.GetDIDDocument(ctx, did)
	}

	return types.GenesisState{
		Documents: documentsMap,
	}
}
