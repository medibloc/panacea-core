package types_test

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

func (suite genesisTestSuite) TestDefaultGenesisState() {
	require := suite.Require()

	defaultState := types.DefaultGenesis()
	require.Empty(defaultState.Documents)
}
