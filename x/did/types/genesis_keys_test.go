package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/medibloc/panacea-core/types/testsuite"
	"github.com/medibloc/panacea-core/x/did/types"
)

type keyTestSuite struct {
	testsuite.TestSuite
}

func TestKeyTestSuite(t *testing.T) {
	suite.Run(t, new(keyTestSuite))
}

func (suite keyTestSuite) TestGenesisDIDDocumentKey() {
	key := types.GenesisDIDDocumentKey{DID: "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"}

	var newKey types.GenesisDIDDocumentKey
	suite.Require().NoError(newKey.Unmarshal(key.Marshal()))
	suite.Require().Equal(key, newKey)
}

func (suite keyTestSuite) TestGenesisDIDDocumentKey_InvalidDID() {
	invalidKey := types.GenesisDIDDocumentKey{DID: "invalid_did"}.Marshal()

	var key types.GenesisDIDDocumentKey
	suite.Require().Error(key.Unmarshal(invalidKey))
}
