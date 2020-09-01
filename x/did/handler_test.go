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
	msg := newMsgCreateDID(t, doc.Document, keyID, privKey)

	res := handleMsgCreateDID(sdk.Context{}, keeper, msg)
	require.True(t, res.IsOK())
	require.Equal(t, 1, len(keeper.ListDIDs(sdk.Context{})))
	require.Equal(t, doc, keeper.GetDIDDocument(sdk.Context{}, did))
}

func TestHandleMsgCreateDID_Exists(t *testing.T) {
	did, doc, privKey, keyID, keeper := ctx()
	msg := newMsgCreateDID(t, doc.Document, keyID, privKey)

	// create
	handleMsgCreateDID(sdk.Context{}, keeper, msg)

	// one more time
	res := handleMsgCreateDID(sdk.Context{}, keeper, msg)
	require.Equal(t, types.ErrDIDExists(did).Result(), res)
}

func TestHandleMsgCreateDID_Deactivated(t *testing.T) {
	did, doc, privKey, keyID, keeper := ctx()
	msg := newMsgCreateDID(t, doc.Document, keyID, privKey)

	// create and deactivate
	handleMsgCreateDID(sdk.Context{}, keeper, msg)
	handleMsgDeactivateDID(sdk.Context{}, keeper, newMsgDeactivateDID(t, did, keyID, privKey, types.InitialSequence))

	// create once again
	res := handleMsgCreateDID(sdk.Context{}, keeper, msg)
	require.Equal(t, types.ErrDIDDeactivated(did).Result(), res)
}

func TestHandleMsgCreateDID_SigVerificationFailed(t *testing.T) {
	did, doc, privKey, keyID, keeper := ctx()
	sig, _ := types.Sign(doc.Document, types.InitialSequence, privKey)
	sig[0] += 1 // pollute the signature

	res := handleMsgCreateDID(sdk.Context{}, keeper, types.NewMsgCreateDID(
		did, doc.Document, keyID, sig, sdk.AccAddress{},
	))
	require.Equal(t, types.ErrSigVerificationFailed().Result(), res)
}

func TestHandleMsgUpdateDID(t *testing.T) {
	did, origDoc, privKey, keyID, keeper := ctx()

	// create
	handleMsgCreateDID(sdk.Context{}, keeper, newMsgCreateDID(t, origDoc.Document, keyID, privKey))

	// prepare a new doc
	newDoc := origDoc.Document
	newDoc.PubKeys = append(newDoc.PubKeys, types.NewPubKey(
		types.NewKeyID(did, "key2"),
		types.ES256K,
		secp256k1.GenPrivKey().PubKey(),
	))

	// call
	msg := newMsgUpdateDID(t, newDoc, keyID, privKey, origDoc.Seq)
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

	// update without creation
	res := handleMsgUpdateDID(sdk.Context{}, keeper, newMsgUpdateDID(t, origDoc.Document, keyID, privKey, origDoc.Seq))
	require.Equal(t, types.ErrDIDNotFound(did).Result(), res)
}

func TestHandleMsgUpdateDID_DIDDeactivated(t *testing.T) {
	did, origDoc, privKey, keyID, keeper := ctx()
	handleMsgCreateDID(sdk.Context{}, keeper, newMsgCreateDID(t, origDoc.Document, keyID, privKey))

	// deactivate
	deactivateMsg := newMsgDeactivateDID(t, did, keyID, privKey, origDoc.Seq)
	require.True(t, handleMsgDeactivateDID(sdk.Context{}, keeper, deactivateMsg).IsOK())

	// update
	res := handleMsgUpdateDID(sdk.Context{}, keeper, newMsgUpdateDID(t, origDoc.Document, keyID, privKey, origDoc.Seq))
	require.Equal(t, types.ErrDIDDeactivated(did).Result(), res)
}

func TestHandleMsgDeactivateDID(t *testing.T) {
	did, doc, privKey, keyID, keeper := ctx()
	handleMsgCreateDID(sdk.Context{}, keeper, newMsgCreateDID(t, doc.Document, keyID, privKey))

	// deactivate
	msg := newMsgDeactivateDID(t, did, keyID, privKey, types.InitialSequence)
	res := handleMsgDeactivateDID(sdk.Context{}, keeper, msg)
	require.True(t, res.IsOK())

	// check if it's really deactivated
	tombstone := keeper.GetDIDDocument(sdk.Context{}, did)
	require.False(t, tombstone.Empty())
	require.True(t, tombstone.Deactivated())
	require.Equal(t, types.InitialSequence+1, tombstone.Seq)
}

func TestHandleMsgDeactivateDID_DIDNotFound(t *testing.T) {
	did, _, privKey, keyID, keeper := ctx()

	// deactivate without creation
	msg := newMsgDeactivateDID(t, did, keyID, privKey, types.InitialSequence)
	res := handleMsgDeactivateDID(sdk.Context{}, keeper, msg)
	require.Equal(t, types.ErrDIDNotFound(did).Result(), res)
}

func TestHandleMsgDeactivateDID_DIDDeactivated(t *testing.T) {
	did, doc, privKey, keyID, keeper := ctx()
	handleMsgCreateDID(sdk.Context{}, keeper, newMsgCreateDID(t, doc.Document, keyID, privKey))

	// deactivate
	msg := newMsgDeactivateDID(t, did, keyID, privKey, types.InitialSequence)
	handleMsgDeactivateDID(sdk.Context{}, keeper, msg)

	// one more time
	res := handleMsgDeactivateDID(sdk.Context{}, keeper, msg)
	require.Equal(t, types.ErrDIDDeactivated(did).Result(), res)
}

func TestHandleMsgDeactivateDID_SigVerificationFailed(t *testing.T) {
	did, doc, privKey, keyID, keeper := ctx()
	handleMsgCreateDID(sdk.Context{}, keeper, newMsgCreateDID(t, doc.Document, keyID, privKey))

	sig, _ := types.Sign(did, doc.Seq, privKey)
	sig[0] += 1 // pollute the signature

	msg := types.NewMsgDeactivateDID(did, keyID, sig, sdk.AccAddress{})
	res := handleMsgDeactivateDID(sdk.Context{}, keeper, msg)
	require.Equal(t, types.ErrSigVerificationFailed().Result(), res)
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

func ctx() (types.DID, types.DIDDocumentWithSeq, crypto.PrivKey, types.KeyID, Keeper) {
	did := types.DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	doc, privKey := newDIDDocumentWithSeq(did)
	return did, doc, privKey, doc.Document.PubKeys[0].ID, newMockKeeper()
}

func newDIDDocumentWithSeq(did types.DID) (types.DIDDocumentWithSeq, crypto.PrivKey) {
	privKey := secp256k1.GenPrivKey()
	keyID := types.NewKeyID(did, "key1")
	pubKey := types.NewPubKey(keyID, types.ES256K, privKey.PubKey())
	doc := types.NewDIDDocumentWithSeq(types.NewDIDDocument(did, pubKey), types.InitialSequence)
	return doc, privKey
}

func newMsgCreateDID(t *testing.T, doc types.DIDDocument, keyID types.KeyID, privKey crypto.PrivKey) MsgCreateDID {
	sig, err := types.Sign(doc, types.InitialSequence, privKey)
	require.NoError(t, err)
	return types.NewMsgCreateDID(doc.ID, doc, keyID, sig, sdk.AccAddress{})
}

func newMsgUpdateDID(t *testing.T, newDoc types.DIDDocument, keyID types.KeyID, privKey crypto.PrivKey, seq types.Sequence) MsgUpdateDID {
	sig, err := types.Sign(newDoc, seq, privKey)
	require.NoError(t, err)
	return types.NewMsgUpdateDID(newDoc.ID, newDoc, keyID, sig, sdk.AccAddress{})
}

func newMsgDeactivateDID(t *testing.T, did types.DID, keyID types.KeyID, privKey crypto.PrivKey, seq types.Sequence) MsgDeactivateDID {
	sig, err := types.Sign(did, seq, privKey)
	require.NoError(t, err)
	return types.NewMsgDeactivateDID(did, keyID, sig, sdk.AccAddress{})
}
