package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datadeal/testutil"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/stretchr/testify/suite"
)

type dealTestSuite struct {
	testutil.DataDealBaseTestSuite

	defaultFunds sdk.Coins
	buyerAccAddr sdk.AccAddress
}

func TestDealTestSuite(t *testing.T) {
	suite.Run(t, new(dealTestSuite))
}

func (suite *dealTestSuite) BeforeTest(_, _ string) {
	suite.buyerAccAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	suite.defaultFunds = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)))

	testDeal := suite.MakeTestDeal(1, suite.buyerAccAddr, 100)
	err := suite.DataDealKeeper.SetNextDealNumber(suite.Ctx, 2)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDeal(suite.Ctx, testDeal)
	suite.Require().NoError(err)
}

func (suite *dealTestSuite) TestCreateNewDeal() {

	err := suite.FundAccount(suite.Ctx, suite.buyerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	budget := &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(10000000)}

	msgCreateDeal := &types.MsgCreateDeal{
		DataSchema:   []string{"http://jsonld.com"},
		Budget:       budget,
		MaxNumData:   10000,
		BuyerAddress: suite.buyerAccAddr.String(),
	}

	buyer, err := sdk.AccAddressFromBech32(msgCreateDeal.BuyerAddress)
	suite.Require().NoError(err)

	dealID, err := suite.DataDealKeeper.CreateDeal(suite.Ctx, buyer, msgCreateDeal)
	suite.Require().NoError(err)

	expectedId, err := suite.DataDealKeeper.GetNextDealNumberAndIncrement(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(dealID, expectedId-uint64(1))

	deal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, dealID)
	suite.Require().NoError(err)
	suite.Require().Equal(deal.GetDataSchema(), msgCreateDeal.GetDataSchema())
	suite.Require().Equal(deal.GetBudget(), msgCreateDeal.GetBudget())
	suite.Require().Equal(deal.GetMaxNumData(), msgCreateDeal.GetMaxNumData())
	suite.Require().Equal(deal.GetConsumerAddress(), msgCreateDeal.GetBuyerAddress())
	suite.Require().Equal(deal.GetStatus(), types.DEAL_STATUS_ACTIVE)
}

func (suite *dealTestSuite) TestCheckDealCurNumDataAndIncrement() {
	err := suite.FundAccount(suite.Ctx, suite.buyerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	budget := &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(10000000)}

	msgCreateDeal := &types.MsgCreateDeal{
		DataSchema:   []string{"http://jsonld.com"},
		Budget:       budget,
		MaxNumData:   1,
		BuyerAddress: suite.buyerAccAddr.String(),
	}

	buyer, err := sdk.AccAddressFromBech32(msgCreateDeal.BuyerAddress)
	suite.Require().NoError(err)

	dealID, err := suite.DataDealKeeper.CreateDeal(suite.Ctx, buyer, msgCreateDeal)
	suite.Require().NoError(err)

	check, err := suite.DataDealKeeper.IsDealCompleted(suite.Ctx, dealID)
	suite.Require().NoError(err)
	suite.Equal(false, check)

	err = suite.DataDealKeeper.IncrementCurNumDataAtDeal(suite.Ctx, dealID)
	suite.Require().NoError(err)
	updatedDeal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, dealID)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), updatedDeal.CurNumData)

	check, err = suite.DataDealKeeper.IsDealCompleted(suite.Ctx, dealID)
	suite.Require().NoError(err)
	suite.Require().Equal(true, check)

}
