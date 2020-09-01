package did

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/did/types"

	"github.com/stretchr/testify/require"
)

func TestDefaultGenesisState(t *testing.T) {
	defaultState := DefaultGenesisState()
	require.Empty(t, defaultState.Documents)
}

func TestGenesis(t *testing.T) {
	ctx := sdk.Context{}

	// prepare a keeper with some data
	did1 := types.DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	doc1, _ := newDIDDocumentWithSeq(did1)
	did2 := types.DID("did:panacea:testnet:6Me8MCctZBYrPKS5zGZt6")
	doc2, _ := newDIDDocumentWithSeq(did2)
	keeper := newMockKeeper()
	keeper.SetDIDDocument(ctx, did1, doc1)
	keeper.SetDIDDocument(ctx, did2, doc2)

	// export a genesis
	state := ExportGenesis(ctx, keeper)
	require.Equal(t, 2, len(state.Documents))
	require.Equal(t, doc1, state.Documents[newGenesisKey(did1)])
	require.Equal(t, doc2, state.Documents[newGenesisKey(did2)])

	// check if the exported genesis is valid
	require.NoError(t, ValidateGenesis(state))

	// import it to a new keeper
	newK := newMockKeeper()
	InitGenesis(ctx, newK, state)
	require.Equal(t, 2, len(newK.ListDIDs(ctx)))
	require.Equal(t, doc1, newK.GetDIDDocument(ctx, did1))
	require.Equal(t, doc2, newK.GetDIDDocument(ctx, did2))
}

func newGenesisKey(did types.DID) string {
	return GenesisDIDDocumentKey{DID: did}.Marshal()
}
