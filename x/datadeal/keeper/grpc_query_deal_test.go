package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/testutil"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/stretchr/testify/suite"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
)

type queryDealTestSuite struct {
	testutil.DataDealBaseTestSuite
	acc1 sdk.AccAddress

	sellerAccPrivKey cryptotypes.PrivKey
	sellerAccPubKey  cryptotypes.PubKey
	sellerAccAddr    sdk.AccAddress

	verifiableCID string

}

func TestQueryDealTest(t *testing.T) {
	suite.Run(t, new(queryDealTestSuite))
}

func (suite *queryDealTestSuite) BeforeTest(_, _ string) {

	suite.acc1 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	suite.verifiableCID = "verifiableCID"

	suite.sellerAccPrivKey = secp256k1.GenPrivKey()
	suite.sellerAccPubKey = suite.sellerAccPrivKey.PubKey()
	suite.sellerAccAddr = sdk.AccAddress(suite.sellerAccPubKey.Address())
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

func (suite queryDealTestSuite) makeNewDataSale() *types.DataSale {
	return &types.DataSale{
		SellerAddress: suite.sellerAccAddr.String(),
		DealId:        1,
		VerifiableCid: suite.verifiableCID,
		DeliveredCid:  "",
		Status:        types.DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD,
		VotingPeriod: &oracletypes.VotingPeriod{
			VotingStartTime: time.Now(),
			VotingEndTime:   time.Now().Add(5 * time.Second),
		},
		VerificationTallyResult: nil,
		DeliveryTallyResult:     nil,
	}
}

func (suite queryDealTestSuite) TestDataSale() {
	newDataSale := suite.makeNewDataSale()
	err := suite.DataDealKeeper.SetDataSale(suite.Ctx, newDataSale)
	suite.Require().NoError(err)

	req := types.QueryDataSaleRequest{
		DealId:        1,
		VerifiableCid: suite.verifiableCID,
	}

	res, err := suite.DataDealKeeper.DataSale(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Equal(newDataSale.DealId, res.DataSale.DealId)
	suite.Require().Equal(newDataSale.VerifiableCid, res.DataSale.VerifiableCid)
	suite.Require().Equal(newDataSale.VotingPeriod.VotingStartTime.UTC(), res.DataSale.VotingPeriod.VotingStartTime)
	suite.Require().Equal(newDataSale.VotingPeriod.VotingEndTime.UTC(), res.DataSale.VotingPeriod.VotingEndTime)
	suite.Require().Equal(newDataSale.SellerAddress, res.DataSale.SellerAddress)
}
