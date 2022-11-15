package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/testutil"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/stretchr/testify/suite"
)

type queryDealTestSuite struct {
	testutil.DataDealBaseTestSuite

	sellerAccPrivKey cryptotypes.PrivKey
	sellerAccPubKey  cryptotypes.PubKey
	sellerAccAddr    sdk.AccAddress

	buyerAccAddr sdk.AccAddress

	oracleAccPrivKey cryptotypes.PrivKey
	oracleAccPubKey  cryptotypes.PubKey
	oracleAccAddr    sdk.AccAddress

	verifiableCID  string
	verifiableCID2 string
	deliveredCID   string

	dataHash1 string
	dataHash2 string
}

func TestQueryDealTest(t *testing.T) {
	suite.Run(t, new(queryDealTestSuite))
}

func (suite *queryDealTestSuite) BeforeTest(_, _ string) {

	suite.buyerAccAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	suite.verifiableCID = "verifiableCID"
	suite.verifiableCID2 = "verifiableCID2"
	suite.deliveredCID = "deliveredCID"
	suite.dataHash1 = "dataHash"
	suite.dataHash2 = "dataHash2"

	suite.sellerAccPrivKey = secp256k1.GenPrivKey()
	suite.sellerAccPubKey = suite.sellerAccPrivKey.PubKey()
	suite.sellerAccAddr = sdk.AccAddress(suite.sellerAccPubKey.Address())

	suite.oracleAccPrivKey = secp256k1.GenPrivKey()
	suite.oracleAccPubKey = suite.oracleAccPrivKey.PubKey()
	suite.oracleAccAddr = sdk.AccAddress(suite.oracleAccPubKey.Address())
}

func (suite *queryDealTestSuite) TestQueryDeal() {
	deal := suite.MakeTestDeal(1, suite.buyerAccAddr, 100)
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
	suite.Require().Equal(deal, res.Deal)
}

func (suite *queryDealTestSuite) TestQueryDeals() {
	// set deals
	deal1 := suite.MakeTestDeal(1, suite.buyerAccAddr, 100)
	err := suite.DataDealKeeper.SetDeal(suite.Ctx, deal1)
	suite.Require().NoError(err)

	deal2 := suite.MakeTestDeal(2, suite.buyerAccAddr, 100)
	err = suite.DataDealKeeper.SetDeal(suite.Ctx, deal2)
	suite.Require().NoError(err)

	deal3 := suite.MakeTestDeal(3, suite.buyerAccAddr, 100)
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

func (suite *queryDealTestSuite) TestDataSale() {
	newDataSale := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID)
	err := suite.DataDealKeeper.SetDataSale(suite.Ctx, newDataSale)
	suite.Require().NoError(err)

	req := types.QueryDataSaleRequest{
		DealId:   1,
		DataHash: suite.dataHash1,
	}

	res, err := suite.DataDealKeeper.DataSale(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Equal(newDataSale.DealId, res.DataSale.DealId)
	suite.Require().Equal(newDataSale.VerifiableCid, res.DataSale.VerifiableCid)
	suite.Require().Equal(newDataSale.VerificationVotingPeriod.VotingStartTime.UTC(), res.DataSale.VerificationVotingPeriod.VotingStartTime)
	suite.Require().Equal(newDataSale.VerificationVotingPeriod.VotingEndTime.UTC(), res.DataSale.VerificationVotingPeriod.VotingEndTime)
	suite.Require().Equal(newDataSale.SellerAddress, res.DataSale.SellerAddress)
}

func (suite *queryDealTestSuite) TestDataSales() {
	dataSale1 := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID)
	err := suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale1)
	suite.Require().NoError(err)

	dataSale2 := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash2, suite.verifiableCID2)
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale2)
	suite.Require().NoError(err)

	req := types.QueryDataSalesRequest{
		DealId: 1,
	}

	res, err := suite.DataDealKeeper.DataSales(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Len(res.DataSale, 2)
	suite.Require().Equal(dataSale1.DataHash, res.DataSale[0].DataHash)
	suite.Require().Equal(dataSale2.DataHash, res.DataSale[1].DataHash)
}

func (suite *queryDealTestSuite) TestDataVerificationVote() {
	dataVerificationVote := suite.MakeNewDataVerificationVote(suite.oracleAccAddr, suite.dataHash1)
	err := suite.DataDealKeeper.SetDataVerificationVote(suite.Ctx, dataVerificationVote)
	suite.Require().NoError(err)

	req := types.QueryDataVerificationVoteRequest{
		DealId:       1,
		DataHash:     suite.dataHash1,
		VoterAddress: suite.oracleAccAddr.String(),
	}

	res, err := suite.DataDealKeeper.DataVerificationVote(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Equal(dataVerificationVote, res.DataVerificationVote)
}

func (suite *queryDealTestSuite) TestDataDeliveryVote() {
	dataVerificationVote := suite.MakeNewDataDeliveryVote(suite.oracleAccAddr, suite.dataHash1, suite.deliveredCID, 1)
	err := suite.DataDealKeeper.SetDataDeliveryVote(suite.Ctx, dataVerificationVote)
	suite.Require().NoError(err)

	req := types.QueryDataDeliveryVoteRequest{
		DealId:       1,
		DataHash:     suite.dataHash1,
		VoterAddress: suite.oracleAccAddr.String(),
	}

	res, err := suite.DataDealKeeper.DataDeliveryVote(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Equal(dataVerificationVote, res.DataDeliveryVote)
}
