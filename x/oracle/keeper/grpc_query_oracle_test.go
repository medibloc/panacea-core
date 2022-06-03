package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
	"testing"
)

type queryOracleTestSuite struct {
	testsuite.TestSuite
}

func TestQueryPoolTest(t *testing.T) {
	suite.Run(t, new(queryOracleTestSuite))
}

func (suite *queryOracleTestSuite) TestQueryOracle() {
	oracle := types.Oracle{
		Address:  oracle1.String(),
		Endpoint: "https://my-oracle-url.org",
	}
	err := suite.OracleKeeper.SetOracle(suite.Ctx, oracle)
	suite.Require().NoError(err)

	req := types.QueryOracleRequest{
		Address: oracle1.String(),
	}

	res, err := suite.OracleKeeper.Oracle(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(oracle, *res.Oracle)
}
