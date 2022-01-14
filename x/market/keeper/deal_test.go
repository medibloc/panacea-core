package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
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

const (
	ACTIVE    = "ACTIVE"    // When deal is activated.
	INACTIVE  = "INACTIVE"  // When deal is deactivated.
	COMPLETED = "COMPLETED" // When deal is completed.
)

var (
	acc1                   = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc2                   = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	defaultFunds sdk.Coins = sdk.NewCoins(sdk.NewCoin("umed", sdk.NewInt(10000000000)))
)

func (suite *dealTestSuite) BeforeTest(_, _ string) {
	testDeal := makeTestDeal()
	suite.MarketKeeper.SetNextDealNumber(suite.Ctx, 2)
	suite.MarketKeeper.SetDeal(suite.Ctx, testDeal)
}

func (suite *dealTestSuite) TestCreateNewDeal() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, acc1, defaultFunds)
	if err != nil {
		panic(err)
	}

	tempDeal := types.Deal{
		DataSchema:            []string{acc1.String()},
		Budget:                &sdk.Coin{Denom: "umed", Amount: sdk.NewInt(10000000)},
		TargetNumData:         10000,
		TrustedDataValidators: []string{acc2.String()},
		Owner:                 acc1.String(),
	}

	owner, _ := sdk.AccAddressFromBech32(tempDeal.GetOwner())

	dealId, _ := suite.MarketKeeper.CreateNewDeal(suite.Ctx, owner, tempDeal)
	expectedId := suite.MarketKeeper.GetNextDealNumberAndIncrement(suite.Ctx) - 1
	suite.Require().Equal(dealId, expectedId)

	deal, _ := suite.MarketKeeper.GetDeal(suite.Ctx, dealId)
	suite.Require().Equal(deal.GetDataSchema(), tempDeal.GetDataSchema())
	suite.Require().Equal(deal.GetBudget(), tempDeal.GetBudget())
	suite.Require().Equal(deal.GetTargetNumData(), tempDeal.GetTargetNumData())
	suite.Require().Equal(deal.GetTrustedDataValidators(), tempDeal.GetTrustedDataValidators())
	suite.Require().Equal(deal.GetOwner(), tempDeal.GetOwner())
	suite.Require().Equal(deal.GetStatus(), ACTIVE)
}

func (suite *dealTestSuite) TestGetDeal() {
	deal, _ := suite.MarketKeeper.GetDeal(suite.Ctx, 1)
	testDeal := makeTestDeal()

	suite.Require().Equal(deal.GetDealId(), testDeal.GetDealId())
	suite.Require().Equal(deal.GetDealAddress(), testDeal.GetDealAddress())
	suite.Require().Equal(deal.GetDataSchema(), testDeal.GetDataSchema())
	suite.Require().Equal(deal.GetBudget(), testDeal.GetBudget())
	suite.Require().Equal(deal.GetTargetNumData(), testDeal.GetTargetNumData())
	suite.Require().Equal(deal.GetTrustedDataValidators(), testDeal.GetTrustedDataValidators())
	suite.Require().Equal(deal.GetOwner(), testDeal.GetOwner())
	suite.Require().Equal(deal.GetStatus(), testDeal.GetStatus())
}

func (suite *dealTestSuite) TestGetBalanceOfDeal() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, acc1, defaultFunds)
	if err != nil {
		panic(err)
	}

	tempDeal := types.Deal{
		DataSchema:            []string{acc1.String()},
		Budget:                &sdk.Coin{Denom: "umed", Amount: sdk.NewInt(10000000)},
		TargetNumData:         10000,
		TrustedDataValidators: []string{acc2.String()},
		Owner:                 acc1.String(),
	}

	owner, _ := sdk.AccAddressFromBech32(tempDeal.GetOwner())

	dealId, _ := suite.MarketKeeper.CreateNewDeal(suite.Ctx, owner, tempDeal)
	deal, _ := suite.MarketKeeper.GetDeal(suite.Ctx, dealId)
	addr, _ := types.AccDealAddressFromBech32(deal.GetDealAddress())

	balance := suite.BankKeeper.GetBalance(suite.Ctx, addr, "umed")
	suite.Require().Equal(balance, *tempDeal.GetBudget())
	ownerBalance := suite.BankKeeper.GetBalance(suite.Ctx, acc1, "umed")
	suite.Require().Equal(ownerBalance, sdk.NewCoin("umed", sdk.NewInt(10000000000)).Sub(balance))
}

func makeTestDeal() types.Deal {
	return types.Deal{
		DealId:                1,
		DealAddress:           types.NewDealAddress(1).String(),
		DataSchema:            []string{acc1.String()},
		Budget:                &sdk.Coin{Denom: "umed", Amount: sdk.NewInt(1000000000)},
		TargetNumData:         10000,
		FilledNumData:         0,
		TrustedDataValidators: []string{acc2.String()},
		Owner:                 acc1.String(),
		Status:                ACTIVE,
	}
}
