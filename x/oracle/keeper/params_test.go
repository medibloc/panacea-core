package keeper_test

import (
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
	"testing"
)

type paramsTestSuite struct {
	testsuite.TestSuite
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(paramsTestSuite))
}

func (suite *paramsTestSuite) TestSetAndGetParams() {
	params := types.DefaultParams()

	suite.OracleKeeper.SetParams(suite.Ctx, params)

	getParams := suite.OracleKeeper.GetParams(suite.Ctx)
	suite.Require().Equal(params, getParams)
}
