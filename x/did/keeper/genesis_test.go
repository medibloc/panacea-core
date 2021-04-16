package keeper

import (
	"testing"

	"github.com/medibloc/panacea-core/x/did/internal/secp256k1util"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/did/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	ctx := sdk.Context{}

	// prepare a keeper with some data
	did1 := types.DID("did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm")
	doc1, _ := newDIDDocumentWithSeq(did1)
	did2 := types.DID("did:panacea:6JamVbJgk8azVgUm7Prd74ry1Uct87nZqL3ny7aR7Cg4")
	doc2, _ := newDIDDocumentWithSeq(did2)
	keeper := newMockKeeper()
	keeper.SetDIDDocument(ctx, did1, doc1)
	keeper.SetDIDDocument(ctx, did2, doc2)
	doc2Deactivated := doc2.Deactivate(doc2.Seq + 1)
	keeper.SetDIDDocument(ctx, did2, doc2Deactivated)

	// export a genesis
	state := ExportGenesis(ctx, keeper)
	require.Equal(t, 2, len(state.Documents))
	require.Equal(t, doc1, state.Documents[newGenesisKey(did1)])
	require.Equal(t, doc2Deactivated, state.Documents[newGenesisKey(did2)])

	// check if the exported genesis is valid
	require.NoError(t, types.ValidateGenesis(state))

	// import it to a new keeper
	newK := newMockKeeper()
	InitGenesis(ctx, newK, state)
	require.Equal(t, 2, len(newK.ListDIDs(ctx)))
	require.Equal(t, doc1, newK.GetDIDDocument(ctx, did1))
	require.Equal(t, doc2Deactivated, newK.GetDIDDocument(ctx, did2))
}

func newGenesisKey(did types.DID) string {
	return types.GenesisDIDDocumentKey{DID: did}.Marshal()
}

// mockKeeper implements the did.Keeper interface
type mockKeeper struct {
	docs map[types.DID]types.DIDDocumentWithSeq
}

func newMockKeeper() *mockKeeper {
	return &mockKeeper{docs: make(map[types.DID]types.DIDDocumentWithSeq)}
}

func (k mockKeeper) Codec() *codec.Codec {
	return codec.New()
}

func (k mockKeeper) SetDIDDocument(ctx sdk.Context, did types.DID, doc types.DIDDocumentWithSeq) {
	k.docs[did] = doc
}

func (k mockKeeper) GetDIDDocument(ctx sdk.Context, did types.DID) types.DIDDocumentWithSeq {
	doc := k.docs[did]
	return doc
}

func (k mockKeeper) ListDIDs(ctx sdk.Context) []types.DID {
	dids := make([]types.DID, 0)
	for did := range k.docs {
		dids = append(dids, did)
	}
	return dids
}

func newDIDDocumentWithSeq(did types.DID) (types.DIDDocumentWithSeq, crypto.PrivKey) {
	privKey := secp256k1.GenPrivKey()
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(privKey))
	veriMethodID := types.NewVeriMethodID(did, "key1")
	veriMethods := []types.VeriMethod{types.NewVeriMethod(veriMethodID, types.ES256K_2019, did, pubKey)}
	authentications := []types.Authentication{types.NewAuthentication(veriMethods[0].ID)}
	doc := types.NewDIDDocumentWithSeq(types.NewDIDDocument(did, veriMethods, authentications), types.InitialSequence)
	return doc, privKey
}
