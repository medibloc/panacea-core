package did

import (
	"testing"

	ariesdid "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/medibloc/panacea-core/v2/x/did/internal/secp256k1util"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

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
	doc1, _ := suite.newDIDDocument(did1)
	did2 := "did:panacea:6JamVbJgk8azVgUm7Prd74ry1Uct87nZqL3ny7aR7Cg4"
	doc2, _ := suite.newDIDDocument(did2)

	didKeeper.SetDIDDocument(suite.Ctx, did1, doc1)
	didKeeper.SetDIDDocument(suite.Ctx, did2, doc2)

	// export a genesis
	state := ExportGenesis(suite.Ctx, didKeeper)
	suite.Require().Equal(2, len(state.Documents))
	suite.Require().Equal(doc1, state.Documents[suite.newGenesisKey(did1)])
	suite.Require().Equal(doc2, state.Documents[suite.newGenesisKey(did2)])

	// check if the exported genesis is valid
	suite.Require().NoError(state.Validate())

	// import it to a new keeper
	InitGenesis(suite.Ctx, didKeeper, *state)
	suite.Require().Equal(2, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(doc1, didKeeper.GetDIDDocument(suite.Ctx, did1))
	suite.Require().Equal(doc2, didKeeper.GetDIDDocument(suite.Ctx, did2))
}

func (suite *genesisTestSuite) newGenesisKey(did string) string {
	return types.GenesisDIDDocumentKey{DID: did}.Marshal()
}

func (suite *genesisTestSuite) newDIDDocument(did string) (*types.DIDDocument, crypto.PrivKey) {
	privKey := secp256k1.GenPrivKey()
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(privKey))
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	es256VerificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey)
	blsVerificationMethod := types.NewVerificationMethod(verificationMethodID, types.BLS1281G2_2020, did, []byte("dummy BBS+ pub key"))

	authentication := types.NewVerification(es256VerificationMethod, ariesdid.Authentication)
	document := types.NewDocument(did,
		ariesdid.WithVerificationMethod([]ariesdid.VerificationMethod{es256VerificationMethod, blsVerificationMethod}),
		ariesdid.WithAuthentication([]ariesdid.Verification{authentication}))

	documentBz, _ := document.JSONBytes()

	didDocument, _ := types.NewDIDDocument(documentBz, types.DidDocumentDataType)

	return &didDocument, privKey
}
