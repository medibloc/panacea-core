package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/market"
	"github.com/medibloc/panacea-core/v2/x/market/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"testing"
)

type dealTestSuite struct {
	testsuite.TestSuite
}

func TestDealTestSuite(t *testing.T) {
	suite.Run(t, new(dealTestSuite))
}

var (
	acc1                   = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc2                   = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	defaultFunds sdk.Coins = sdk.NewCoins(sdk.NewCoin("umed", sdk.NewInt(10000000000)))
)

func (suite *dealTestSuite) BeforeTest(_, _ string) {
	testDeal := makeTestDeal()
	market.InitGenesis(suite.Ctx, suite.MarketKeeper, types.GenesisState{Deals: map[uint64]*types.Deal{
		testDeal.GetDealId(): &testDeal,
	}, NextDealNumber: 2})

	suite.MarketKeeper.SetDeal(suite.Ctx, testDeal)

	err := suite.BankKeeper.AddCoins(suite.Ctx, acc1, defaultFunds)
	if err != nil {
		panic(err)
	}
}

func (suite *dealTestSuite) TestCreateNewDeal() {
	tempDeal := types.Deal{
		DataSchema:           []string{acc1.String()},
		Budget:               &sdk.Coin{Denom: "umed", Amount: sdk.NewInt(10000000)},
		WantDataCount:        10000,
		TrustedDataValidator: []string{acc2.String()},
		Owner:                acc1.String(),
	}

	owner, _ := sdk.AccAddressFromBech32(tempDeal.GetOwner())

	dealId, _ := suite.MarketKeeper.CreateNewDeal(suite.Ctx, owner, tempDeal)
	suite.Require().Equal(dealId, uint64(2))
	deal, _ := suite.MarketKeeper.GetDeal(suite.Ctx, 2)
	suite.Require().Equal(deal.GetDataSchema(), tempDeal.GetDataSchema())
	suite.Require().Equal(deal.GetBudget(), tempDeal.GetBudget())
	suite.Require().Equal(deal.GetWantDataCount(), tempDeal.GetWantDataCount())
	suite.Require().Equal(deal.GetTrustedDataValidator(), tempDeal.GetTrustedDataValidator())
	suite.Require().Equal(deal.GetOwner(), tempDeal.GetOwner())
	suite.Require().Equal(deal.GetStatus(), "ACTIVE")
}

func (suite *dealTestSuite) TestGetDeal() {
	deal, _ := suite.MarketKeeper.GetDeal(suite.Ctx, 1)
	testDeal := makeTestDeal()

	suite.Require().Equal(deal.GetDealId(), testDeal.GetDealId())
	suite.Require().Equal(deal.GetDealAddress(), testDeal.GetDealAddress())
	suite.Require().Equal(deal.GetDataSchema(), testDeal.GetDataSchema())
	suite.Require().Equal(deal.GetBudget(), testDeal.GetBudget())
	suite.Require().Equal(deal.GetWantDataCount(), testDeal.GetWantDataCount())
	suite.Require().Equal(deal.GetTrustedDataValidator(), testDeal.GetTrustedDataValidator())
	suite.Require().Equal(deal.GetOwner(), testDeal.GetOwner())
	suite.Require().Equal(deal.GetStatus(), testDeal.GetStatus())
}

func makeTestDeal() types.Deal {
	return types.Deal{
		DealId:               1,
		DealAddress:          types.NewDealAddress(1).String(),
		DataSchema:           []string{acc1.String()},
		Budget:               &sdk.Coin{Denom: "umed", Amount: sdk.NewInt(1000000000)},
		WantDataCount:        10000,
		CompleteDataCount:    0,
		TrustedDataValidator: []string{acc2.String()},
		Owner:                acc1.String(),
		Status:               "ACTIVE",
	}
}
