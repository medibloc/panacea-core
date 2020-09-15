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
	did, doc, privKey, veriMethodID, keeper := ctx()
	msg := newMsgCreateDID(t, doc.Document, veriMethodID, privKey)

	res := handleMsgCreateDID(sdk.Context{}, keeper, msg)
	require.True(t, res.IsOK())
	require.Equal(t, 1, len(keeper.ListDIDs(sdk.Context{})))
	require.Equal(t, doc, keeper.GetDIDDocument(sdk.Context{}, did))
}

func TestHandleMsgCreateDID_Exists(t *testing.T) {
	did, doc, privKey, veriMethodID, keeper := ctx()
	msg := newMsgCreateDID(t, doc.Document, veriMethodID, privKey)

	// create
	handleMsgCreateDID(sdk.Context{}, keeper, msg)

	// one more time
	res := handleMsgCreateDID(sdk.Context{}, keeper, msg)
	require.Equal(t, types.ErrDIDExists(did).Result(), res)
}

func TestHandleMsgCreateDID_Deactivated(t *testing.T) {
	did, doc, privKey, veriMethodID, keeper := ctx()
	msg := newMsgCreateDID(t, doc.Document, veriMethodID, privKey)

	// create and deactivate
	handleMsgCreateDID(sdk.Context{}, keeper, msg)
	handleMsgDeactivateDID(sdk.Context{}, keeper, newMsgDeactivateDID(t, did, veriMethodID, privKey, types.InitialSequence))

	// create once again
	res := handleMsgCreateDID(sdk.Context{}, keeper, msg)
	require.Equal(t, types.ErrDIDDeactivated(did).Result(), res)
}

func TestHandleMsgCreateDID_SigVerificationFailed(t *testing.T) {
	did, doc, privKey, veriMethodID, keeper := ctx()
	sig, _ := types.Sign(doc.Document, types.InitialSequence, privKey)
	sig[0] += 1 // pollute the signature

	res := handleMsgCreateDID(sdk.Context{}, keeper, types.NewMsgCreateDID(
		did, doc.Document, veriMethodID, sig, sdk.AccAddress{},
	))
	require.Equal(t, types.ErrSigVerificationFailed().Result(), res)
}

func TestHandleMsgUpdateDID(t *testing.T) {
	did, origDoc, privKey, veriMethodID, keeper := ctx()

	// create
	handleMsgCreateDID(sdk.Context{}, keeper, newMsgCreateDID(t, origDoc.Document, veriMethodID, privKey))

	// prepare a new doc
	newDoc := origDoc.Document
	newDoc.VeriMethods = append(newDoc.VeriMethods, types.NewVeriMethod(
		types.NewVeriMethodID(did, "key2"),
		types.ES256K,
		did,
		secp256k1.GenPrivKey().PubKey(),
	))

	// call
	msg := newMsgUpdateDID(t, newDoc, veriMethodID, privKey, origDoc.Seq)
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
	did, origDoc, privKey, veriMethodID, keeper := ctx()

	// update without creation
	res := handleMsgUpdateDID(sdk.Context{}, keeper, newMsgUpdateDID(t, origDoc.Document, veriMethodID, privKey, origDoc.Seq))
	require.Equal(t, types.ErrDIDNotFound(did).Result(), res)
}

func TestHandleMsgUpdateDID_DIDDeactivated(t *testing.T) {
	did, origDoc, privKey, veriMethodID, keeper := ctx()
	handleMsgCreateDID(sdk.Context{}, keeper, newMsgCreateDID(t, origDoc.Document, veriMethodID, privKey))

	// deactivate
	deactivateMsg := newMsgDeactivateDID(t, did, veriMethodID, privKey, origDoc.Seq)
	require.True(t, handleMsgDeactivateDID(sdk.Context{}, keeper, deactivateMsg).IsOK())

	// update
	res := handleMsgUpdateDID(sdk.Context{}, keeper, newMsgUpdateDID(t, origDoc.Document, veriMethodID, privKey, origDoc.Seq))
	require.Equal(t, types.ErrDIDDeactivated(did).Result(), res)
}

func TestHandleMsgDeactivateDID(t *testing.T) {
	did, doc, privKey, veriMethodID, keeper := ctx()
	handleMsgCreateDID(sdk.Context{}, keeper, newMsgCreateDID(t, doc.Document, veriMethodID, privKey))

	// deactivate
	msg := newMsgDeactivateDID(t, did, veriMethodID, privKey, types.InitialSequence)
	res := handleMsgDeactivateDID(sdk.Context{}, keeper, msg)
	require.True(t, res.IsOK())

	// check if it's really deactivated
	tombstone := keeper.GetDIDDocument(sdk.Context{}, did)
	require.False(t, tombstone.Empty())
	require.True(t, tombstone.Deactivated())
	require.Equal(t, types.InitialSequence+1, tombstone.Seq)
}

func TestHandleMsgDeactivateDID_DIDNotFound(t *testing.T) {
	did, _, privKey, veriMethodID, keeper := ctx()

	// deactivate without creation
	msg := newMsgDeactivateDID(t, did, veriMethodID, privKey, types.InitialSequence)
	res := handleMsgDeactivateDID(sdk.Context{}, keeper, msg)
	require.Equal(t, types.ErrDIDNotFound(did).Result(), res)
}

func TestHandleMsgDeactivateDID_DIDDeactivated(t *testing.T) {
	did, doc, privKey, veriMethodID, keeper := ctx()
	handleMsgCreateDID(sdk.Context{}, keeper, newMsgCreateDID(t, doc.Document, veriMethodID, privKey))

	// deactivate
	msg := newMsgDeactivateDID(t, did, veriMethodID, privKey, types.InitialSequence)
	handleMsgDeactivateDID(sdk.Context{}, keeper, msg)

	// one more time
	res := handleMsgDeactivateDID(sdk.Context{}, keeper, msg)
	require.Equal(t, types.ErrDIDDeactivated(did).Result(), res)
}

func TestHandleMsgDeactivateDID_SigVerificationFailed(t *testing.T) {
	did, doc, privKey, veriMethodID, keeper := ctx()
	handleMsgCreateDID(sdk.Context{}, keeper, newMsgCreateDID(t, doc.Document, veriMethodID, privKey))

	sig, _ := types.Sign(did, doc.Seq, privKey)
	sig[0] += 1 // pollute the signature

	msg := types.NewMsgDeactivateDID(did, veriMethodID, sig, sdk.AccAddress{})
	res := handleMsgDeactivateDID(sdk.Context{}, keeper, msg)
	require.Equal(t, types.ErrSigVerificationFailed().Result(), res)
}

func TestVerifyDIDOwnership(t *testing.T) {
	_, doc, privKey, veriMethodID, _ := ctx()
	data := any("random string")
	sig, _ := types.Sign(data, doc.Seq, privKey)

	newSeq, err := verifyDIDOwnership(data, doc.Seq, doc.Document, veriMethodID, sig)
	require.NoError(t, err)
	require.Equal(t, doc.Seq+1, newSeq)
}

func TestVerifyDIDOwnership_VeriMethodIDNotFound(t *testing.T) {
	_, doc, privKey, veriMethodID, _ := ctx()
	data := any("random string")
	sig, _ := types.Sign(data, doc.Seq, privKey)

	dummyVeriMethodID := veriMethodID + "dummy"
	_, err := verifyDIDOwnership(data, doc.Seq, doc.Document, dummyVeriMethodID, sig)
	require.EqualError(t, err, types.ErrVeriMethodIDNotFound(dummyVeriMethodID).Error())
}

func TestVerifyDIDOwnership_SigVerificationFailed(t *testing.T) {
	_, doc, privKey, veriMethodID, _ := ctx()
	data := any("random string")
	sig, _ := types.Sign(data, doc.Seq+11234, privKey)

	_, err := verifyDIDOwnership(data, doc.Seq, doc.Document, veriMethodID, sig)
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

func ctx() (types.DID, types.DIDDocumentWithSeq, crypto.PrivKey, types.VeriMethodID, Keeper) {
	did := types.DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	doc, privKey := newDIDDocumentWithSeq(did)
	return did, doc, privKey, doc.Document.VeriMethods[0].ID, newMockKeeper()
}

func newDIDDocumentWithSeq(did types.DID) (types.DIDDocumentWithSeq, crypto.PrivKey) {
	privKey := secp256k1.GenPrivKey()
	veriMethodID := types.NewVeriMethodID(did, "key1")
	pubKey := types.NewVeriMethod(veriMethodID, types.ES256K, did, privKey.PubKey())
	doc := types.NewDIDDocumentWithSeq(types.NewDIDDocument(did, pubKey), types.InitialSequence)
	return doc, privKey
}

func newMsgCreateDID(t *testing.T, doc types.DIDDocument, veriMethodID types.VeriMethodID, privKey crypto.PrivKey) MsgCreateDID {
	sig, err := types.Sign(doc, types.InitialSequence, privKey)
	require.NoError(t, err)
	return types.NewMsgCreateDID(doc.ID, doc, veriMethodID, sig, sdk.AccAddress{})
}

func newMsgUpdateDID(t *testing.T, newDoc types.DIDDocument, veriMethodID types.VeriMethodID, privKey crypto.PrivKey, seq types.Sequence) MsgUpdateDID {
	sig, err := types.Sign(newDoc, seq, privKey)
	require.NoError(t, err)
	return types.NewMsgUpdateDID(newDoc.ID, newDoc, veriMethodID, sig, sdk.AccAddress{})
}

func newMsgDeactivateDID(t *testing.T, did types.DID, veriMethodID types.VeriMethodID, privKey crypto.PrivKey, seq types.Sequence) MsgDeactivateDID {
	sig, err := types.Sign(did, seq, privKey)
	require.NoError(t, err)
	return types.NewMsgDeactivateDID(did, veriMethodID, sig, sdk.AccAddress{})
}
