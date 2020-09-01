package did

import (
	"testing"

	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/did/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
)

func TestHandleMsgCreateDID(t *testing.T) {
	did, doc, privKey, keyID, keeper := ctx()
	sig, _ := types.Sign(doc.Document, types.NewSequence(), privKey)

	res := handleMsgCreateDID(sdk.Context{}, keeper, types.NewMsgCreateDID(
		did, doc.Document, keyID, sig, sdk.AccAddress{},
	))
	require.True(t, res.IsOK())
	require.Equal(t, 1, len(keeper.ListDIDs(sdk.Context{})))
	require.Equal(t, doc, keeper.GetDIDDocument(sdk.Context{}, did))
}

func TestHandleMsgCreateDID_Exists(t *testing.T) {
	did, doc, privKey, keyID, keeper := ctx()
	sig, _ := types.Sign(doc.Document, types.NewSequence(), privKey)

	keeper.SetDIDDocument(sdk.Context{}, did, doc)

	res := handleMsgCreateDID(sdk.Context{}, keeper, types.NewMsgCreateDID(
		did, doc.Document, keyID, sig, sdk.AccAddress{},
	))
	require.Equal(t, types.ErrDIDExists(did).Result(), res)
}

func TestHandleMsgCreateDID_SigVerificationFailed(t *testing.T) {
	did, doc, privKey, keyID, keeper := ctx()
	sig, _ := types.Sign(doc.Document, types.NewSequence(), privKey)

	sig[0] += 1 // pollute the signature

	res := handleMsgCreateDID(sdk.Context{}, keeper, types.NewMsgCreateDID(
		did, doc.Document, keyID, sig, sdk.AccAddress{},
	))
	require.Equal(t, types.ErrSigVerificationFailed().Result(), res)
}

func TestHandleMsgUpdateDID(t *testing.T) {
	did, origDoc, privKey, keyID, keeper := ctx()
	keeper.SetDIDDocument(sdk.Context{}, did, origDoc)

	// prepare a new doc
	newDoc := origDoc.Document
	newDoc.PubKeys = append(newDoc.PubKeys, types.NewPubKey(
		types.NewKeyID(did, "key2"),
		types.ES256K,
		secp256k1.GenPrivKey().PubKey(),
	))
	sig, _ := types.Sign(newDoc, origDoc.Seq, privKey)

	// call
	msg := types.NewMsgUpdateDID(did, newDoc, keyID, sig, sdk.AccAddress{})
	res := handleMsgUpdateDID(sdk.Context{}, keeper, msg)
	require.True(t, res.IsOK())
	require.Equal(t, 1, len(keeper.ListDIDs(sdk.Context{})))
	updated := keeper.GetDIDDocument(sdk.Context{}, did)
	require.Equal(t, newDoc, updated.Document)
	require.Equal(t, origDoc.Seq+1, updated.Seq)

	// call again with the same signature (replay-attack! should be failed!)
	res = handleMsgUpdateDID(sdk.Context{}, keeper, msg)
	require.False(t, res.IsOK())
	require.Equal(t, types.ErrSigVerificationFailed().Code(), res.Code)
}

func TestHandleMsgUpdateDID_DIDNotFound(t *testing.T) {
	did, origDoc, privKey, keyID, keeper := ctx()
	newDoc := origDoc.Document
	sig, _ := types.Sign(newDoc, origDoc.Seq, privKey)

	msg := types.NewMsgUpdateDID(did, newDoc, keyID, sig, sdk.AccAddress{})
	res := handleMsgUpdateDID(sdk.Context{}, keeper, msg)
	require.Equal(t, types.ErrDIDNotFound(did).Result(), res)
}

func TestHandleMsgDeleteDID(t *testing.T) {
	did, doc, privKey, keyID, keeper := ctx()
	keeper.SetDIDDocument(sdk.Context{}, did, doc)
	sig, _ := types.Sign(did, doc.Seq, privKey)

	res := handleMsgDeleteDID(sdk.Context{}, keeper, types.NewMsgDeleteDID(did, keyID, sig, sdk.AccAddress{}))
	require.True(t, res.IsOK())
	require.Empty(t, keeper.ListDIDs(sdk.Context{}))
}

func TestHandleMsgDeleteDID_DIDNotFound(t *testing.T) {
	did, doc, privKey, keyID, keeper := ctx()
	sig, _ := types.Sign(did, doc.Seq, privKey)

	msg := types.NewMsgDeleteDID(did, keyID, sig, sdk.AccAddress{})
	res := handleMsgDeleteDID(sdk.Context{}, keeper, msg)
	require.Equal(t, types.ErrDIDNotFound(did).Result(), res)
}

func TestVerifyDIDOwnership(t *testing.T) {
	_, doc, privKey, keyID, _ := ctx()
	data := any("random string")
	sig, _ := types.Sign(data, doc.Seq, privKey)

	newSeq, err := verifyDIDOwnership(data, doc.Seq, doc.Document, keyID, sig)
	require.NoError(t, err)
	require.Equal(t, doc.Seq+1, newSeq)
}

func TestVerifyDIDOwnership_KeyIDNotFound(t *testing.T) {
	_, doc, privKey, keyID, _ := ctx()
	data := any("random string")
	sig, _ := types.Sign(data, doc.Seq, privKey)

	dummyKeyID := keyID + "dummy"
	_, err := verifyDIDOwnership(data, doc.Seq, doc.Document, dummyKeyID, sig)
	require.EqualError(t, err, types.ErrKeyIDNotFound(dummyKeyID).Error())
}

func TestVerifyDIDOwnership_SigVerificationFailed(t *testing.T) {
	_, doc, privKey, keyID, _ := ctx()
	data := any("random string")
	sig, _ := types.Sign(data, doc.Seq+11234, privKey)

	_, err := verifyDIDOwnership(data, doc.Seq, doc.Document, keyID, sig)
	require.EqualError(t, err, types.ErrSigVerificationFailed().Error())
}

type any string

func (a any) GetSignBytes() []byte {
	return sdk.MustSortJSON(codec.New().MustMarshalJSON(a))
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

func (k mockKeeper) DeleteDID(ctx sdk.Context, did types.DID) {
	delete(k.docs, did)
}

func ctx() (types.DID, types.DIDDocumentWithSeq, crypto.PrivKey, types.KeyID, Keeper) {
	did := types.DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	doc, privKey := newDIDDocumentWithSeq(did)
	return did, doc, privKey, doc.Document.PubKeys[0].ID, newMockKeeper()
}

func newDIDDocumentWithSeq(did types.DID) (types.DIDDocumentWithSeq, crypto.PrivKey) {
	privKey := secp256k1.GenPrivKey()
	keyID := types.NewKeyID(did, "key1")
	pubKey := types.NewPubKey(keyID, types.ES256K, privKey.PubKey())
	seq := types.NewSequence()
	doc := types.NewDIDDocumentWithSeq(types.NewDIDDocument(did, pubKey), seq)
	return doc, privKey
}
