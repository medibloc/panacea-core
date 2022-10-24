package keeper_test

import (
	"testing"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/stretchr/testify/suite"

	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/did/internal/secp256k1util"
	"github.com/medibloc/panacea-core/v2/x/did/types"
)

type didTestSuite struct {
	testsuite.TestSuite
}

func TestDIDTestSuite(t *testing.T) {
	suite.Run(t, new(didTestSuite))
}

func (suite *didTestSuite) TestSetGetDIDDocument() {
	didKeeper := suite.DIDKeeper

	// Input two DIDDocument
	did1 := "did1:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	docWithSeq1, _ := makeTestDIDDocumentWithSeq(did1)
	didKeeper.SetDIDDocument(suite.Ctx, did1, docWithSeq1)
	did2 := "did1:panacea:1Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgap"
	docWithSeq2, _ := makeTestDIDDocumentWithSeq(did1)
	didKeeper.SetDIDDocument(suite.Ctx, did2, docWithSeq2)

	// Test one DIDDocument
	resDocWithSeq := didKeeper.GetDIDDocument(suite.Ctx, did1)
	suite.Require().NotNil(resDocWithSeq)
	suite.Require().Equal(docWithSeq1, resDocWithSeq)

	// Test all DIDs
	resDIDs := didKeeper.ListDIDs(suite.Ctx)
	suite.Require().Equal(2, len(resDIDs))
	suite.Require().Equal(did2, resDIDs[0])
	suite.Require().Equal(did1, resDIDs[1])
}

func makeTestDIDDocumentWithSeq(did string) (types.DIDDocumentWithSeq, crypto.PrivKey) {
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
