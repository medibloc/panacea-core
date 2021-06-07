package did

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/medibloc/panacea-core/types/testsuite"
	"github.com/medibloc/panacea-core/x/did/types"
)

type genesisTestSuite struct {
	testsuite.TestSuite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}

func (suite genesisTestSuite) TestGenesis() {
	didKeeper := suite.DIDKeeper

	// prepare a keeper with some data
	did1 := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	doc1, _ := newDIDDocumentWithSeq(did1)
	did2 := "did:panacea:6JamVbJgk8azVgUm7Prd74ry1Uct87nZqL3ny7aR7Cg4"
	doc2, _ := newDIDDocumentWithSeq(did2)

	didKeeper.SetDIDDocument(suite.Ctx, did1, doc1)
	didKeeper.SetDIDDocument(suite.Ctx, did2, doc2)
	doc2Deactivated := doc2.Deactivate(doc2.Seq + 1)
	didKeeper.SetDIDDocument(suite.Ctx, did2, doc2Deactivated)

	// export a genesis
	state := ExportGenesis(suite.Ctx, didKeeper)
	suite.Require().Equal(2, len(state.Documents))
	suite.Require().Equal(doc1, *state.Documents[newGenesisKey(did1)])
	suite.Require().Equal(doc2Deactivated, *state.Documents[newGenesisKey(did2)])

	// check if the exported genesis is valid
	suite.Require().NoError(state.Validate())

	// import it to a new keeper
	InitGenesis(suite.Ctx, didKeeper, *state)
	suite.Require().Equal(2, len(didKeeper.ListDIDs(suite.Ctx)))
	suite.Require().Equal(doc1, didKeeper.GetDIDDocument(suite.Ctx, did1))
	suite.Require().Equal(doc2Deactivated, didKeeper.GetDIDDocument(suite.Ctx, did2))
}

func newGenesisKey(did string) string {
	return types.GenesisDIDDocumentKey{DID: did}.Marshal()
}
