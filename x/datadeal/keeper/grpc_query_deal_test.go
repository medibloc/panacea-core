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
	deal := makeTestDeal()
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

func (suite *queryDealTestSuite) TestQueryDataCert() {
	deal := makeTestDeal()
	suite.DataDealKeeper.SetDeal(suite.Ctx, deal)

	dataCert := makeTestCert("1a312c1223x2fs3", newAddr, acc1)
	suite.DataDealKeeper.SetDataCert(suite.Ctx, 1, dataCert)

	req := types.QueryDataCertRequest{
		DealId:   deal.DealId,
		DataHash: string(dataCert.UnsignedCert.DataHash),
	}
	res, err := suite.DataDealKeeper.DataCert(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(dataCert, *res.DataCert)
}

func (suite *queryDealTestSuite) TestQueryDataCerts() {
	deal := makeTestDeal()
	suite.DataDealKeeper.SetDeal(suite.Ctx, deal)

	for i := 0; i < 5; i++ {
		dataCert := makeTestCert("1a312c1223x2fs3", newAddr, acc1)
		suite.DataDealKeeper.SetDataCert(suite.Ctx, 1, dataCert)
	}

}
