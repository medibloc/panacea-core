package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/stretchr/testify/suite"
)

type queryDealTestSuite struct {
	testsuite.TestSuite
}

func TestQueryDealTest(t *testing.T) {
	suite.Run(t, new(queryDealTestSuite))
}

func (suite *queryDealTestSuite) TestQueryDeal() {
	deal := makeTestDeal(1)
	suite.DataDealKeeper.SetDeal(suite.Ctx, deal)
	suite.DataDealKeeper.SetNextDealNumber(suite.Ctx, 2)

	req := types.QueryDealRequest{
		DealId: deal.GetDealId(),
	}
	res, err := suite.DataDealKeeper.Deal(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(deal, *res.Deal)
}

func (suite queryDealTestSuite) TestQueryDeals() {
	// set deals
	deal1 := makeTestDeal(1)
	suite.DataDealKeeper.SetDeal(suite.Ctx, deal1)

	deal2 := makeTestDeal(2)
	suite.DataDealKeeper.SetDeal(suite.Ctx, deal2)

	deal3 := makeTestDeal(3)
	suite.DataDealKeeper.SetDeal(suite.Ctx, deal3)

	req := types.QueryDealsRequest{}

	res, err := suite.DataDealKeeper.Deals(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Len(res.Deals, 3)
	suite.Require().Contains(res.Deals, deal1)
	suite.Require().Contains(res.Deals, deal2)
	suite.Require().Contains(res.Deals, deal3)
}
