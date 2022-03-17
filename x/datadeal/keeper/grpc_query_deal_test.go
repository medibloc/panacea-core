package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/stretchr/testify/suite"
	"testing"
)

type queryDealTestSuite struct {
	testsuite.TestSuite
}

func TestQueryDealTest(t *testing.T) {
	suite.Run(t, new(queryDealTestSuite))
}

func (suite *queryDealTestSuite) TestQueryDeal() {
	deal := makeTestDeal()
	suite.DatadealKeeper.SetDeal(suite.Ctx, deal)
	suite.DatadealKeeper.SetNextDealNumber(suite.Ctx, 2)

	req := types.QueryDealRequest{
		DealId: deal.GetDealId(),
	}
	res, err := suite.DatadealKeeper.Deal(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(deal, *res.Deal)
}
