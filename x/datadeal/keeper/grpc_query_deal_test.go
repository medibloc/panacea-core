package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/testutil"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

type queryDealTestSuite struct {
	testutil.DataDealBaseTestSuite
	acc1 sdk.AccAddress
}

func TestQueryDealTest(t *testing.T) {
	suite.Run(t, new(queryDealTestSuite))
}

func (suite *queryDealTestSuite) BeforeTest(_, _ string) {

	suite.acc1 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
}

func (suite *queryDealTestSuite) TestQueryDeal() {
	deal := suite.MakeTestDeal(1, suite.acc1)
	err := suite.DataDealKeeper.SetDeal(suite.Ctx, deal)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetNextDealNumber(suite.Ctx, 2)
	suite.Require().NoError(err)

	req := types.QueryDealRequest{
		DealId: deal.GetId(),
	}
	res, err := suite.DataDealKeeper.Deal(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(deal, *res.Deal)
}

func (suite queryDealTestSuite) TestQueryDeals() {
	// set deals
	deal1 := suite.MakeTestDeal(1, suite.acc1)
	err := suite.DataDealKeeper.SetDeal(suite.Ctx, deal1)
	suite.Require().NoError(err)

	deal2 := suite.MakeTestDeal(2, suite.acc1)
	err = suite.DataDealKeeper.SetDeal(suite.Ctx, deal2)
	suite.Require().NoError(err)

	deal3 := suite.MakeTestDeal(3, suite.acc1)
	err = suite.DataDealKeeper.SetDeal(suite.Ctx, deal3)
	suite.Require().NoError(err)

	req := types.QueryDealsRequest{}

	res, err := suite.DataDealKeeper.Deals(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Len(res.Deal, 3)
	suite.Require().Equal(res.Deal[0].Address, deal1.Address)
	suite.Require().Equal(res.Deal[1].Address, deal2.Address)
	suite.Require().Equal(res.Deal[2].Address, deal3.Address)
}
