package did

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/medibloc/panacea-core/types/testsuite"
	"github.com/medibloc/panacea-core/x/did/internal/secp256k1util"
	"github.com/medibloc/panacea-core/x/did/types"
)

type handlerTestSuite struct {
	testsuite.TestSuite
}

func (suite handlerTestSuite) BeforeTest(_, _ string) {

}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(handlerTestSuite))
}

func (suite handlerTestSuite) TestHandleMsgCreateDID() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer

	did, docWithSeq, privKey, verificationMethodID := makeTestData()
	msg := newMsgCreateDID(suite, *docWithSeq.Document, verificationMethodID, privKey)

	res, err := didMsgServer.CreateDID(sdk.WrapSDKContext(suite.Ctx), &msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(docWithSeq, didKeeper.GetDIDDocument(suite.Ctx, did))
}

func (suite handlerTestSuite) TestHandleMsgCreateDID_Exists() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer

	did, docWithSeq, privKey, verificationMethodID := makeTestData()
	msg := newMsgCreateDID(suite, *docWithSeq.Document, verificationMethodID, privKey)

	// create
	res, err := didMsgServer.CreateDID(sdk.WrapSDKContext(suite.Ctx), &msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(docWithSeq, didKeeper.GetDIDDocument(suite.Ctx, did))

	// one more time
	res, err = didMsgServer.CreateDID(sdk.WrapSDKContext(suite.Ctx), &msg)
	suite.Require().ErrorIs(types.ErrDIDExists, err)
	suite.Require().Nil(res)
}

func (suite handlerTestSuite) TestHandleMsgCreateDID_Deactivated() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)

	did, docWithSeq, privKey, verificationMethodID := makeTestData()
	msg := newMsgCreateDID(suite, *docWithSeq.Document, verificationMethodID, privKey)

	// create and deactivate
	res, err := didMsgServer.CreateDID(goContext, &msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(docWithSeq, didKeeper.GetDIDDocument(suite.Ctx, did))

	deactivateMsg := newMsgDeactivateDID(suite, did, verificationMethodID, privKey, types.InitialSequence)
	deactivateRes, err := didMsgServer.DeactivateDID(goContext, &deactivateMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(deactivateRes)

	// create once again
	res, err = didMsgServer.CreateDID(goContext, &msg)
	suite.Require().ErrorIs(types.ErrDIDDeactivated, err)
	suite.Require().Nil(res)
}

func (suite handlerTestSuite) TestHandleMsgCreateDID_SigVerificationFailed() {
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)

	did, docWithSeq, privKey, veriMethodID := makeTestData()
	doc := docWithSeq.Document
	sig, _ := types.Sign(doc, types.InitialSequence, privKey)
	sig[0] += 1 // pollute the signature

	msg := types.NewMsgCreateDID(did, *doc, veriMethodID, sig, sdk.AccAddress{}.String())

	res, err := didMsgServer.CreateDID(goContext, &msg)
	suite.Require().ErrorIs(types.ErrSigVerificationFailed, err)
	suite.Require().Nil(res)
}

func (suite handlerTestSuite) TestHandleMsgUpdateDID() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)

	did, origDocWithSeq, privKey, verificationMethodID := makeTestData()
	createMsg := newMsgCreateDID(suite, *origDocWithSeq.Document, verificationMethodID, privKey)

	// create
	res, err := didMsgServer.CreateDID(goContext, &createMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(origDocWithSeq, didKeeper.GetDIDDocument(suite.Ctx, did))

	// prepare a new doc
	newDoc := origDocWithSeq.Document
	verificationMethod := types.NewVerificationMethod(
		types.NewVerificationMethodID(did, "key2"),
		types.ES256K_2019,
		did,
		secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey())),
	)
	newDoc.VerificationMethods = append(newDoc.VerificationMethods, &verificationMethod)

	// call
	updateMsg := newMsgUpdateDID(suite, *newDoc, verificationMethodID, privKey, origDocWithSeq.Seq)
	updateRes, err := didMsgServer.UpdateDID(goContext, &updateMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(updateRes)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))

	updatedDoc := didKeeper.GetDIDDocument(suite.Ctx, did)
	suite.Require().Equal(newDoc, updatedDoc.Document)
	suite.Require().Equal(origDocWithSeq.Seq+1, updatedDoc.Seq)

	// call again with the same signature (replay-attack! should be failed!)
	updateRes, err = didMsgServer.UpdateDID(goContext, &updateMsg)
	suite.Require().ErrorIs(types.ErrSigVerificationFailed, err)
	suite.Require().Nil(updateRes)
}

func (suite handlerTestSuite) TestHandleMsgUpdateDID_DIDNotFound() {
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)

	_, origDocWithSeq, privKey, verificationMethodID := makeTestData()

	// update without creation
	msg := newMsgUpdateDID(suite, *origDocWithSeq.Document, verificationMethodID, privKey, origDocWithSeq.Seq)
	res, err := didMsgServer.UpdateDID(goContext, &msg)
	suite.ErrorIs(types.ErrDIDNotFound, err)
	suite.Nil(res)
}

func (suite handlerTestSuite) TestHandleMsgUpdateDID_DIDDeactivated() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)

	did, origDocWithSeq, privKey, verificationMethodID := makeTestData()

	msg := newMsgCreateDID(suite, *origDocWithSeq.Document, verificationMethodID, privKey)
	res, err := didMsgServer.CreateDID(goContext, &msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(origDocWithSeq, didKeeper.GetDIDDocument(suite.Ctx, did))

	// deactivate
	deactivateMsg := newMsgDeactivateDID(suite, did, verificationMethodID, privKey, origDocWithSeq.Seq)
	deactivateRes, err := didMsgServer.DeactivateDID(goContext, &deactivateMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(deactivateRes)

	// update
	updateMsg := newMsgUpdateDID(suite, *origDocWithSeq.Document, verificationMethodID, privKey, origDocWithSeq.Seq)
	updateRes, err := didMsgServer.UpdateDID(goContext, &updateMsg)
	suite.Require().ErrorIs(types.ErrDIDDeactivated, err)
	suite.Require().Nil(updateRes)
}

func (suite handlerTestSuite) TestHandleMsgDeactivateDID() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)

	did, docWithSeq, privKey, verificationMethodID := makeTestData()

	createMsg := newMsgCreateDID(suite, *docWithSeq.Document, verificationMethodID, privKey)
	createRes, err := didMsgServer.CreateDID(goContext, &createMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(createRes)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(docWithSeq, didKeeper.GetDIDDocument(suite.Ctx, did))
	// deactivate
	deactivateMsg := newMsgDeactivateDID(suite, did, verificationMethodID, privKey, types.InitialSequence)
	deactivateRes, err := didMsgServer.DeactivateDID(goContext, &deactivateMsg)

	suite.Require().NoError(err)
	suite.Require().NotNil(deactivateRes)

	// check if it's really deactivated
	tombstone := didKeeper.GetDIDDocument(suite.Ctx, did)
	suite.Require().False(tombstone.Empty())
	suite.Require().True(tombstone.Deactivated())
	suite.Require().Equal(types.InitialSequence+1, tombstone.Seq)
}

func (suite handlerTestSuite) TestHandleMsgDeactivateDID_DIDNotFound() {
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)

	did, _, privKey, verificationMethodID := makeTestData()

	// deactivate without creation
	msg := newMsgDeactivateDID(suite, did, verificationMethodID, privKey, types.InitialSequence)
	res, err := didMsgServer.DeactivateDID(goContext, &msg)
	suite.Require().ErrorIs(types.ErrDIDNotFound, err)
	suite.Require().Nil(res)
}

func (suite handlerTestSuite) TestHandleMsgDeactivateDID_DIDDeactivated() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)

	did, docWithSeq, privKey, verificationMethodID := makeTestData()

	msg := newMsgCreateDID(suite, *docWithSeq.Document, verificationMethodID, privKey)
	createRes, err := didMsgServer.CreateDID(goContext, &msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(createRes)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(docWithSeq, didKeeper.GetDIDDocument(suite.Ctx, did))

	// deactivate
	deactivateMsg := newMsgDeactivateDID(suite, did, verificationMethodID, privKey, types.InitialSequence)
	deactivateRes, err := didMsgServer.DeactivateDID(goContext, &deactivateMsg)
	suite.Require().NotNil(deactivateRes)
	suite.Require().NoError(err)

	// one more time
	deactivateRes, err = didMsgServer.DeactivateDID(goContext, &deactivateMsg)
	suite.Require().Nil(deactivateRes)
	suite.Require().ErrorIs(types.ErrDIDDeactivated, err)
}

func (suite handlerTestSuite) TestHandleMsgDeactivateDID_SigVerificationFailed() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)

	did, docWithSeq, privKey, verificationMethodID := makeTestData()

	createMsg := newMsgCreateDID(suite, *docWithSeq.Document, verificationMethodID, privKey)
	createRes, err := didMsgServer.CreateDID(goContext, &createMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(createRes)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(docWithSeq, didKeeper.GetDIDDocument(suite.Ctx, did))

	signableDID := types.SignableDID(did)
	sig, _ := types.Sign(signableDID, docWithSeq.Seq, privKey)
	sig[0] += 1 // pollute the signature

	deactivateMsg := types.NewMsgDeactivateDID(did, verificationMethodID, sig, sdk.AccAddress{}.String())
	deactivateRes, err := didMsgServer.DeactivateDID(goContext, &deactivateMsg)
	suite.Require().Nil(deactivateRes)
	suite.Require().ErrorIs(types.ErrSigVerificationFailed, err)
}

func makeTestData() (string, types.DIDDocumentWithSeq, crypto.PrivKey, string) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	doc, privKey := newDIDDocumentWithSeq(did)
	return did, doc, privKey, doc.Document.VerificationMethods[0].ID
}

func newDIDDocumentWithSeq(did string) (types.DIDDocumentWithSeq, crypto.PrivKey) {
	privKey := secp256k1.GenPrivKey()
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(privKey))
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	es256VerificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey)
	blsVerificationMethod := types.NewVerificationMethod(verificationMethodID, types.BLS1281G2_2020, did, []byte("dummy BBS+ pub key"))
	verificationMethods := []*types.VerificationMethod{
		&es256VerificationMethod,
		&blsVerificationMethod,
	}
	verificationRelationship := types.NewVerificationRelationship(verificationMethods[0].ID)
	authentications := []*types.VerificationRelationship{
		&verificationRelationship,
	}
	doc := types.NewDIDDocument(did, types.WithVerificationMethods(verificationMethods), types.WithAuthentications(authentications))
	docWithSeq := types.NewDIDDocumentWithSeq(
		&doc,
		types.InitialSequence,
	)
	return docWithSeq, privKey
}

func newMsgCreateDID(suite handlerTestSuite, doc types.DIDDocument, verificationMethodID string, privKey crypto.PrivKey) types.MsgCreateDID {
	sig, err := types.Sign(doc, types.InitialSequence, privKey)
	suite.Require().NoError(err)
	return types.NewMsgCreateDID(doc.ID, doc, verificationMethodID, sig, sdk.AccAddress{}.String())
}

func newMsgUpdateDID(suite handlerTestSuite, newDoc types.DIDDocument, verificationMethodID string, privKey crypto.PrivKey, seq uint64) types.MsgUpdateDID {
	sig, err := types.Sign(newDoc, seq, privKey)
	suite.Require().NoError(err)
	return types.NewMsgUpdateDID(newDoc.ID, newDoc, verificationMethodID, sig, sdk.AccAddress{}.String())
}

func newMsgDeactivateDID(suite handlerTestSuite, did string, verificationMethodID string, privKey crypto.PrivKey, seq uint64) types.MsgDeactivateDID {
	sig, err := types.Sign(types.SignableDID(did), seq, privKey)
	suite.Require().NoError(err)
	return types.NewMsgDeactivateDID(did, verificationMethodID, sig, sdk.AccAddress{}.String())
}
