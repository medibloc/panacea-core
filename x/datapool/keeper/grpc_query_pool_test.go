package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
	"github.com/stretchr/testify/suite"
)

type queryPoolTestSuite struct {
	testsuite.TestSuite
}

func TestQueryPoolTest(t *testing.T) {
	suite.Run(t, new(queryPoolTestSuite))
}

func (suite *queryPoolTestSuite) TestQueryDataValidator() {
	dataValidator := types.DataValidator{
		Address:  dataVal1.String(),
		Endpoint: "https://my-validator-url.org",
	}
	suite.DataPoolKeeper.SetDataValidator(suite.Ctx, dataValidator)

	req := types.QueryDataValidatorRequest{
		Address: dataVal1.String(),
	}

	res, err := suite.DataPoolKeeper.DataValidator(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(dataValidator, *res.DataValidator)
}
