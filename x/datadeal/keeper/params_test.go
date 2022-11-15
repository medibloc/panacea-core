package keeper_test

import (
	"testing"

	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/stretchr/testify/suite"
)

type paramsTestSuite struct {
	testsuite.TestSuite
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(paramsTestSuite))
}

func (suite *paramsTestSuite) TestSetAndGetParams() {
	params := types.DefaultParams()

	suite.DataDealKeeper.SetParams(suite.Ctx, params)

	getParams := suite.DataDealKeeper.GetParams(suite.Ctx)

	suite.Require().Equal(params, getParams)
}
