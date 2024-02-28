package did

import (
	"testing"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/medibloc/panacea-core/v2/x/did/internal/secp256k1util"

	"github.com/stretchr/testify/suite"

	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/did/types"
)

type genesisTestSuite struct {
	testsuite.TestSuite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}

func (suite *genesisTestSuite) TestGenesis() {
	didKeeper := suite.DIDKeeper

	// prepare a keeper with some data
	did1 := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	doc1, _ := suite.newDIDDocumentWithSeq(did1)
	did2 := "did:panacea:6JamVbJgk8azVgUm7Prd74ry1Uct87nZqL3ny7aR7Cg4"
	doc2, _ := suite.newDIDDocumentWithSeq(did2)

	didKeeper.SetDIDDocument(suite.Ctx, did1, doc1)
	didKeeper.SetDIDDocument(suite.Ctx, did2, doc2)
	doc2Deactivated := doc2.Deactivate(doc2.Sequence + 1)
	didKeeper.SetDIDDocument(suite.Ctx, did2, doc2Deactivated)

	// export a genesis
	state := ExportGenesis(suite.Ctx, didKeeper)
	suite.Require().Equal(2, len(state.Documents))
	suite.Require().Equal(doc1, *state.Documents[suite.newGenesisKey(did1)])
	suite.Require().Equal(doc2Deactivated, *state.Documents[suite.newGenesisKey(did2)])

	// check if the exported genesis is valid
	suite.Require().NoError(state.Validate())

	// import it to a new keeper
	InitGenesis(suite.Ctx, didKeeper, *state)
	suite.Require().Equal(2, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(doc1, didKeeper.GetDIDDocument(suite.Ctx, did1))
	suite.Require().Equal(doc2Deactivated, didKeeper.GetDIDDocument(suite.Ctx, did2))
}

func (suite *genesisTestSuite) newGenesisKey(did string) string {
	return types.GenesisDIDDocumentKey{DID: did}.Marshal()
}

func (suite *genesisTestSuite) newDIDDocumentWithSeq(did string) (types.DIDDocumentWithSeq, crypto.PrivKey) {
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
