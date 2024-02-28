package keeper_test

import (
	"testing"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/did/internal/secp256k1util"
	didkeeper "github.com/medibloc/panacea-core/v2/x/did/keeper"
	"github.com/medibloc/panacea-core/v2/x/did/types"
	"github.com/stretchr/testify/suite"
)

type msgServerTestSuite struct {
	testsuite.TestSuite
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(msgServerTestSuite))
}

func (suite *msgServerTestSuite) TestHandleMsgCreateDID() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer

	did, docWithSeq, privKey, verificationMethodID := suite.makeTestData()
	msg := newMsgCreateDID(suite, *docWithSeq.Document, verificationMethodID, privKey)

	res, err := didMsgServer.CreateDID(sdk.WrapSDKContext(suite.Ctx), &msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(docWithSeq, didKeeper.GetDIDDocument(suite.Ctx, did))
}

func (suite *msgServerTestSuite) TestHandleMsgCreateDID_Exists() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer

	did, docWithSeq, privKey, verificationMethodID := suite.makeTestData()
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

func (suite *msgServerTestSuite) TestHandleMsgCreateDID_Deactivated() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)

	did, docWithSeq, privKey, verificationMethodID := suite.makeTestData()
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

func (suite *msgServerTestSuite) TestHandleMsgCreateDID_SigVerificationFailed() {
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)

	did, docWithSeq, privKey, veriMethodID := suite.makeTestData()
	doc := docWithSeq.Document
	sig, _ := types.Sign(doc, types.InitialSequence, privKey)
	sig[0] += 1 // pollute the signature

	msg := types.NewMsgServiceCreateDIDResponse(did, *doc, veriMethodID, sig, sdk.AccAddress{}.String())

	res, err := didMsgServer.CreateDID(goContext, &msg)
	suite.Require().ErrorIs(types.ErrSigVerificationFailed, err)
	suite.Require().Nil(res)
}

func (suite *msgServerTestSuite) TestHandleMsgUpdateDID() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)

	did, origDocWithSeq, privKey, verificationMethodID := suite.makeTestData()
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
	updateMsg := newMsgUpdateDID(suite, *newDoc, verificationMethodID, privKey, origDocWithSeq.Sequence)
	updateRes, err := didMsgServer.UpdateDID(goContext, &updateMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(updateRes)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))

	updatedDoc := didKeeper.GetDIDDocument(suite.Ctx, did)
	suite.Require().Equal(newDoc, updatedDoc.Document)
	suite.Require().Equal(origDocWithSeq.Sequence+1, updatedDoc.Sequence)

	// call again with the same signature (replay-attack! should be failed!)
	updateRes, err = didMsgServer.UpdateDID(goContext, &updateMsg)
	suite.Require().ErrorIs(types.ErrSigVerificationFailed, err)
	suite.Require().Nil(updateRes)
}

func (suite *msgServerTestSuite) TestHandleMsgUpdateDID_DIDNotFound() {
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)

	_, origDocWithSeq, privKey, verificationMethodID := suite.makeTestData()

	// update without creation
	msg := newMsgUpdateDID(suite, *origDocWithSeq.Document, verificationMethodID, privKey, origDocWithSeq.Sequence)
	res, err := didMsgServer.UpdateDID(goContext, &msg)
	suite.ErrorIs(types.ErrDIDNotFound, err)
	suite.Nil(res)
}

func (suite *msgServerTestSuite) TestHandleMsgUpdateDID_DIDDeactivated() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)

	did, origDocWithSeq, privKey, verificationMethodID := suite.makeTestData()

	msg := newMsgCreateDID(suite, *origDocWithSeq.Document, verificationMethodID, privKey)
	res, err := didMsgServer.CreateDID(goContext, &msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(origDocWithSeq, didKeeper.GetDIDDocument(suite.Ctx, did))

	// deactivate
	deactivateMsg := newMsgDeactivateDID(suite, did, verificationMethodID, privKey, origDocWithSeq.Sequence)
	deactivateRes, err := didMsgServer.DeactivateDID(goContext, &deactivateMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(deactivateRes)

	// update
	updateMsg := newMsgUpdateDID(suite, *origDocWithSeq.Document, verificationMethodID, privKey, origDocWithSeq.Sequence)
	updateRes, err := didMsgServer.UpdateDID(goContext, &updateMsg)
	suite.Require().ErrorIs(types.ErrDIDDeactivated, err)
	suite.Require().Nil(updateRes)
}

func (suite *msgServerTestSuite) TestHandleMsgDeactivateDID() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)

	did, docWithSeq, privKey, verificationMethodID := suite.makeTestData()

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
	suite.Require().Equal(types.InitialSequence+1, tombstone.Sequence)
}

func (suite *msgServerTestSuite) TestHandleMsgDeactivateDID_DIDNotFound() {
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)

	did, _, privKey, verificationMethodID := suite.makeTestData()

	// deactivate without creation
	msg := newMsgDeactivateDID(suite, did, verificationMethodID, privKey, types.InitialSequence)
	res, err := didMsgServer.DeactivateDID(goContext, &msg)
	suite.Require().ErrorIs(types.ErrDIDNotFound, err)
	suite.Require().Nil(res)
}

func (suite *msgServerTestSuite) TestHandleMsgDeactivateDID_DIDDeactivated() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)

	did, docWithSeq, privKey, verificationMethodID := suite.makeTestData()

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

func (suite *msgServerTestSuite) TestHandleMsgDeactivateDID_SigVerificationFailed() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)

	did, docWithSeq, privKey, verificationMethodID := suite.makeTestData()
	doc := *docWithSeq.Document

	createMsg := newMsgCreateDID(suite, doc, verificationMethodID, privKey)
	createRes, err := didMsgServer.CreateDID(goContext, &createMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(createRes)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(docWithSeq, didKeeper.GetDIDDocument(suite.Ctx, did))

	sig, _ := types.Sign(&doc, docWithSeq.Sequence, privKey)
	sig[0] += 1 // pollute the signature

	deactivateMsg := types.NewMsgServiceDeactivateDIDRequest(did, verificationMethodID, sig, sdk.AccAddress{}.String())
	deactivateRes, err := didMsgServer.DeactivateDID(goContext, &deactivateMsg)
	suite.Require().Nil(deactivateRes)
	suite.Require().ErrorIs(types.ErrSigVerificationFailed, err)
}

func (suite *msgServerTestSuite) TestVerifyDIDOwnership() {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	docWithSeq, privKey := suite.newDIDDocumentWithSeq(did)
	doc := docWithSeq.Document

	sig, _ := types.Sign(doc, docWithSeq.Sequence, privKey)

	newSeq, err := didkeeper.VerifyDIDOwnership(doc, docWithSeq.Sequence, docWithSeq.Document, docWithSeq.Document.VerificationMethods[0].Id, sig)
	suite.Require().NoError(err)
	suite.Require().Equal(docWithSeq.Sequence+1, newSeq)
}

func (suite *msgServerTestSuite) TestVerifyDIDOwnership_SigVerificationFailed() {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	docWithSeq, privKey := suite.newDIDDocumentWithSeq(did)
	doc := docWithSeq.Document

	sig, _ := types.Sign(doc, docWithSeq.Sequence+11234, privKey)

	_, err := didkeeper.VerifyDIDOwnership(doc, docWithSeq.Sequence, docWithSeq.Document, docWithSeq.Document.VerificationMethods[0].Id, sig)
	suite.Require().ErrorIs(types.ErrSigVerificationFailed, err)
}

func (suite *msgServerTestSuite) makeTestData() (string, types.DIDDocumentWithSeq, crypto.PrivKey, string) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	doc, privKey := suite.newDIDDocumentWithSeq(did)
	return did, doc, privKey, doc.Document.VerificationMethods[0].Id
}

func (suite *msgServerTestSuite) newDIDDocumentWithSeq(did string) (types.DIDDocumentWithSeq, crypto.PrivKey) {
	privKey := secp256k1.GenPrivKey()
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(privKey))
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	es256VerificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey)
	blsVerificationMethod := types.NewVerificationMethod(verificationMethodID, types.BLS1281G2_2020, did, []byte("dummy BBS+ pub key"))
	verificationMethods := []*types.VerificationMethod{
		&es256VerificationMethod,
		&blsVerificationMethod,
	}
	verificationRelationship := types.NewVerificationRelationship(verificationMethods[0].Id)
	authentications := []types.VerificationRelationship{
		verificationRelationship,
	}
	doc := types.NewDIDDocument(did, types.WithVerificationMethods(verificationMethods), types.WithAuthentications(authentications))
	docWithSeq := types.NewDIDDocumentWithSeq(
		&doc,
		types.InitialSequence,
	)
	return docWithSeq, privKey
}

func newMsgCreateDID(suite *msgServerTestSuite, doc types.DIDDocument, verificationMethodID string, privKey crypto.PrivKey) types.MsgCreateDID {
	sig, err := types.Sign(&doc, types.InitialSequence, privKey)
	suite.Require().NoError(err)
	return types.NewMsgServiceCreateDIDResponse(doc.Id, doc, verificationMethodID, sig, sdk.AccAddress{}.String())
}

func newMsgUpdateDID(suite *msgServerTestSuite, newDoc types.DIDDocument, verificationMethodID string, privKey crypto.PrivKey, seq uint64) types.MsgUpdateDID {
	sig, err := types.Sign(&newDoc, seq, privKey)
	suite.Require().NoError(err)
	return types.NewMsgUpdateDID(newDoc.Id, newDoc, verificationMethodID, sig, sdk.AccAddress{}.String())
}

func newMsgDeactivateDID(suite *msgServerTestSuite, did string, verificationMethodID string, privKey crypto.PrivKey, seq uint64) types.MsgDeactivateDID {
	doc := types.DIDDocument{
		Id: did,
	}

	sig, err := types.Sign(&doc, seq, privKey)
	suite.Require().NoError(err)
	return types.NewMsgServiceDeactivateDIDRequest(did, verificationMethodID, sig, sdk.AccAddress{}.String())
}
