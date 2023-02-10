package did

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/did/keeper"
	"github.com/medibloc/panacea-core/v2/x/did/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data types.GenesisState) {
	for did, doc := range data.Documents {
		k.SetDIDDocument(ctx, did, doc)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	documentsMap := make(map[string]*types.DIDDocument)

	for _, did := range k.ListDIDs(ctx) {
		key := types.GenesisDIDDocumentKey{DID: did}.Marshal()
		document := k.GetDIDDocument(ctx, did)
		documentsMap[key] = document
	}

	return &types.GenesisState{
		Documents: documentsMap,
	}
}
