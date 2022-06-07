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

	dataCert, err := makeTestCert("1a312c1223x2fs3", newAddr, acc1)
	suite.Require().NoError(err)

	suite.DataDealKeeper.SetDataCert(suite.Ctx, 1, dataCert)

	req := types.QueryDataCertRequest{
		DealId:   deal.GetDealId(),
		DataHash: string(dataCert.UnsignedCert.GetDataHash()),
	}
	res, err := suite.DataDealKeeper.DataCert(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(dataCert, *res.DataCert)
}

func (suite *queryDealTestSuite) TestQueryDataCerts() {
	deal := makeTestDeal()
	suite.DataDealKeeper.SetDeal(suite.Ctx, deal)

	req := types.QueryDataCertsRequest{
		DealId: deal.GetDealId(),
	}

	res, err := suite.DataDealKeeper.DataCerts(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Len(res.DataCerts, 0)

	dataCert1, err := makeTestCert("1a312c1223x2fs3", newAddr, acc1)
	suite.Require().NoError(err)

	suite.DataDealKeeper.SetDataCert(suite.Ctx, deal.DealId, dataCert1)

	dataCert2, err := makeTestCert("1a312c1223x2fs2", newAddr, acc1)
	suite.Require().NoError(err)

	suite.DataDealKeeper.SetDataCert(suite.Ctx, deal.DealId, dataCert2)

	res, err = suite.DataDealKeeper.DataCerts(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Len(res.DataCerts, 2)
	suite.Require().Contains(res.DataCerts, dataCert1)
	suite.Require().Contains(res.DataCerts, dataCert2)
}
