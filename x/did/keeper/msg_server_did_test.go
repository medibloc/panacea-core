package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/did/internal/secp256k1util"
	"github.com/medibloc/panacea-core/x/did/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

type msgServerTestSuite struct {
	suite.Suite
}

type any string

func (a any) GetSignBytes() []byte {
	return sdk.MustSortJSON(types.ModuleCdc.Amino.MustMarshalJSON(a))
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(msgServerTestSuite))
}

func (suite msgServerTestSuite) TestVerifyDIDOwnership() {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	doc, privKey := suite.newDIDDocumentWithSeq(did)

	data := any("random string")
	sig, _ := types.Sign(data, doc.Seq, privKey)

	newSeq, err := verifyDIDOwnership(data, doc.Seq, doc.Document, doc.Document.VerificationMethods[0].ID, sig)
	suite.Require().NoError(err)
	suite.Require().Equal(doc.Seq+1, newSeq)
}

func (suite msgServerTestSuite) TestVerifyDIDOwnership_SigVerificationFailed() {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	doc, privKey := suite.newDIDDocumentWithSeq(did)

	data := any("random string")
	sig, _ := types.Sign(data, doc.Seq+11234, privKey)

	_, err := verifyDIDOwnership(data, doc.Seq, doc.Document, doc.Document.VerificationMethods[0].ID, sig)
	suite.Require().ErrorIs(types.ErrSigVerificationFailed, err)
}

func (suite msgServerTestSuite) newDIDDocumentWithSeq(did string) (types.DIDDocumentWithSeq, crypto.PrivKey) {
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
