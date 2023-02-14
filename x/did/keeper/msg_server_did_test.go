package keeper_test

import (
	"testing"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ariesdid "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/did/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
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
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"

	didDocument, privKey, verificationMethodID := makeTestDIDDocument(did)
	signedDidDocument, msg := newMsgCreateDID(suite, did, *didDocument, verificationMethodID, privKey)

	res, err := didMsgServer.CreateDID(sdk.WrapSDKContext(suite.Ctx), &msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(res.Did, did)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(signedDidDocument, *didKeeper.GetDIDDocument(suite.Ctx, did))
}

func (suite *msgServerTestSuite) TestHandleMsgCreateDID_Exists() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"

	didDocument, privKey, verificationMethodID := makeTestDIDDocument(did)
	signedDidDocument, msg := newMsgCreateDID(suite, did, *didDocument, verificationMethodID, privKey)

	// create
	res, err := didMsgServer.CreateDID(sdk.WrapSDKContext(suite.Ctx), &msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(signedDidDocument, *didKeeper.GetDIDDocument(suite.Ctx, did))

	// one more time
	res, err = didMsgServer.CreateDID(sdk.WrapSDKContext(suite.Ctx), &msg)
	suite.Require().ErrorIs(types.ErrDIDExists, err)
	suite.Require().Nil(res)
}

func (suite *msgServerTestSuite) TestHandleMsgCreateDID_Deactivated() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)

	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"

	didDocument, privKey, verificationMethodID := makeTestDIDDocument(did)
	signedDidDocument, msg := newMsgCreateDID(suite, did, *didDocument, verificationMethodID, privKey)

	// create and deactivate
	res, err := didMsgServer.CreateDID(goContext, &msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(signedDidDocument, *didKeeper.GetDIDDocument(suite.Ctx, did))

	deactivateMsg := newMsgDeactivateDID(did)
	deactivateRes, err := didMsgServer.DeactivateDID(goContext, &deactivateMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(deactivateRes)

	// create once again
	res, err = didMsgServer.CreateDID(goContext, &msg)
	suite.Require().ErrorIs(types.ErrDIDDeactivated, err)
	suite.Require().Nil(res)
}

func (suite *msgServerTestSuite) TestHandleMsgUpdateDID() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"

	didDocument, privKey, verificationMethodID := makeTestDIDDocument(did)
	signedDidDocument, createMsg := newMsgCreateDID(suite, did, *didDocument, verificationMethodID, privKey)

	// create
	res, err := didMsgServer.CreateDID(goContext, &createMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(signedDidDocument, *didKeeper.GetDIDDocument(suite.Ctx, did))

	// prepare a new doc

	newDidDocument := didDocument
	newDocument, err := ariesdid.ParseDocument(newDidDocument.Document)
	suite.Require().NoError(err)
	vmID := types.NewVerificationMethodID(did, "key2")
	newPrivKey := secp256k1.GenPrivKey()
	_, btcecNewPubKey := btcec.PrivKeyFromBytes(btcec.S256(), newPrivKey.Bytes())
	vm := types.NewVerificationMethod(vmID, types.ES256K_2019, did, btcecNewPubKey.SerializeUncompressed())
	newDocument.VerificationMethod = append(newDocument.VerificationMethod, vm)

	newDocumentBz, err := newDocument.JSONBytes()
	suite.Require().NoError(err)

	newDidDocument.Document = newDocumentBz

	// call
	signedNewDidDocument, updateMsg := newMsgUpdateDID(suite, did, *newDidDocument, verificationMethodID, 1, privKey)
	updateRes, err := didMsgServer.UpdateDID(goContext, &updateMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(updateRes)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))

	suite.Require().Equal(signedNewDidDocument, *didKeeper.GetDIDDocument(suite.Ctx, did))

	// call again with the same signature (replay-attack! should be failed!)
	updateRes, err = didMsgServer.UpdateDID(goContext, &updateMsg)
	suite.Require().ErrorContains(err, types.ErrInvalidSequence.Error())
}

func (suite *msgServerTestSuite) TestHandleMsgUpdateDID_DIDNotFound() {
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"

	didDocument, privKey, verificationMethodID := makeTestDIDDocument(did)

	// update without creation
	_, updateMsg := newMsgUpdateDID(suite, did, *didDocument, verificationMethodID, 1, privKey)
	res, err := didMsgServer.UpdateDID(goContext, &updateMsg)
	suite.ErrorIs(types.ErrDIDNotFound, err)
	suite.Nil(res)
}

func (suite *msgServerTestSuite) TestHandleMsgUpdateDID_DIDDeactivated() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"

	didDocument, privKey, verificationMethodID := makeTestDIDDocument(did)

	signedDidDocument, createMsg := newMsgCreateDID(suite, did, *didDocument, verificationMethodID, privKey)
	res, err := didMsgServer.CreateDID(goContext, &createMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(signedDidDocument, *didKeeper.GetDIDDocument(suite.Ctx, did))

	// deactivate
	deactivateMsg := newMsgDeactivateDID(did)
	deactivateRes, err := didMsgServer.DeactivateDID(goContext, &deactivateMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(deactivateRes)

	// update
	_, updateMsg := newMsgUpdateDID(suite, did, *didDocument, verificationMethodID, 1, privKey)
	_, err = didMsgServer.UpdateDID(goContext, &updateMsg)
	suite.Require().ErrorIs(types.ErrDIDDeactivated, err)
}

func (suite *msgServerTestSuite) TestHandleMsgDeactivateDID() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"

	didDocument, privKey, verificationMethodID := makeTestDIDDocument(did)

	signedDidDocument, createMsg := newMsgCreateDID(suite, did, *didDocument, verificationMethodID, privKey)
	createRes, err := didMsgServer.CreateDID(goContext, &createMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(createRes)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(signedDidDocument, *didKeeper.GetDIDDocument(suite.Ctx, did))
	// deactivate
	deactivateMsg := newMsgDeactivateDID(did)
	deactivateRes, err := didMsgServer.DeactivateDID(goContext, &deactivateMsg)

	suite.Require().NoError(err)
	suite.Require().NotNil(deactivateRes)

	// check if it's really deactivated
	tombstone := didKeeper.GetDIDDocument(suite.Ctx, did)
	suite.Require().False(tombstone.Empty())
	suite.Require().True(tombstone.Deactivated)
}

func (suite *msgServerTestSuite) TestHandleMsgDeactivateDID_DIDNotFound() {
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"

	// deactivate without creation
	deactivateMsg := newMsgDeactivateDID(did)
	deactivateRes, err := didMsgServer.DeactivateDID(goContext, &deactivateMsg)
	suite.Require().ErrorIs(types.ErrDIDNotFound, err)
	suite.Require().Nil(deactivateRes)
}

func (suite *msgServerTestSuite) TestHandleMsgDeactivateDID_DIDDeactivated() {
	didKeeper := suite.DIDKeeper
	didMsgServer := suite.DIDMsgServer
	goContext := sdk.WrapSDKContext(suite.Ctx)
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"

	didDocument, privKey, verificationMethodID := makeTestDIDDocument(did)

	signedDidDocument, createMsg := newMsgCreateDID(suite, did, *didDocument, verificationMethodID, privKey)
	createRes, err := didMsgServer.CreateDID(goContext, &createMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(createRes)
	suite.Require().Equal(1, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(signedDidDocument, *didKeeper.GetDIDDocument(suite.Ctx, did))

	// deactivate
	deactivateMsg := newMsgDeactivateDID(did)
	deactivateRes, err := didMsgServer.DeactivateDID(goContext, &deactivateMsg)
	suite.Require().NotNil(deactivateRes)
	suite.Require().NoError(err)

	// one more time
	deactivateRes, err = didMsgServer.DeactivateDID(goContext, &deactivateMsg)
	suite.Require().Nil(deactivateRes)
	suite.Require().ErrorIs(types.ErrDIDDeactivated, err)
}

func newMsgCreateDID(suite *msgServerTestSuite, did string, didDocument types.DIDDocument, verificationMethodID string, privKey *btcec.PrivateKey) (types.DIDDocument, types.MsgCreateDID) {

	signedDoc, err := types.SignDocument(didDocument.Document, verificationMethodID, types.InitialSequence, privKey)
	suite.Require().NoError(err)

	didDocument.Document = signedDoc

	return didDocument, types.NewMsgCreateDID(did, didDocument, sdk.AccAddress{}.String())
}

func newMsgUpdateDID(suite *msgServerTestSuite, did string, didDocument types.DIDDocument, verificationMethodID string, sequence uint64, privKey *btcec.PrivateKey) (types.DIDDocument, types.MsgUpdateDID) {

	signedDoc, err := types.SignDocument(didDocument.Document, verificationMethodID, sequence, privKey)
	suite.Require().NoError(err)

	didDocument.Document = signedDoc

	return didDocument, types.NewMsgUpdateDID(did, didDocument, sdk.AccAddress{}.String())
}

func newMsgDeactivateDID(did string) types.MsgDeactivateDID {

	return types.NewMsgDeactivateDID(did, sdk.AccAddress{}.String())
}
