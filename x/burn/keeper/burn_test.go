package keeper_test

import (
	"testing"

	"github.com/medibloc/panacea-core/v2/types/testsuite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/medibloc/panacea-core/v2/types/assets"
)

var (
	pubKeys = []crypto.PubKey{
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
	}

	address = []sdk.AccAddress{
		sdk.AccAddress(pubKeys[0].Address()),
		sdk.AccAddress(pubKeys[1].Address()),
		sdk.AccAddress(pubKeys[2].Address()),
	}

	initTokens = sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	initCoin   = sdk.NewCoin(assets.MicroMedDenom, initTokens)
	initCoins  = sdk.NewCoins(initCoin)
)

type BurnTestSuite struct {
	testsuite.TestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(BurnTestSuite))
}

func (suite *BurnTestSuite) BeforeTest(_, _ string) {
	ctx := suite.Ctx
	bankKeeper := suite.BankKeeper

	for _, addr := range address {
		// mint coins and send to each account
		err := suite.FundAccount(suite.Ctx, addr, initCoins)
		suite.Require().NoError(err)
	}

	totalCoin := bankKeeper.GetSupply(ctx, assets.MicroMedDenom)

	suite.Require().Equal(assets.MicroMedDenom, totalCoin.Denom)
	suite.Require().Equal(sdk.TokensFromConsensusPower(600, sdk.DefaultPowerReduction), totalCoin.Amount)
}

func (suite *BurnTestSuite) TestBurnCoins() {
	ctx := suite.Ctx
	bankKeeper := suite.BankKeeper
	burnKeeper := suite.BurnKeeper

	beforeBurnCoin1 := bankKeeper.GetBalance(ctx, address[0], assets.MicroMedDenom)
	beforeBurnCoin2 := bankKeeper.GetBalance(ctx, address[1], assets.MicroMedDenom)
	beforeBurnCoin3 := bankKeeper.GetBalance(ctx, address[2], assets.MicroMedDenom)
	beforeBurnTotal := bankKeeper.GetSupply(ctx, assets.MicroMedDenom)

	suite.Require().Equal(initCoin, beforeBurnCoin1)
	suite.Require().Equal(initCoin, beforeBurnCoin2)
	suite.Require().Equal(initCoin, beforeBurnCoin3)
	suite.Require().Equal(assets.MicroMedDenom, beforeBurnTotal.GetDenom())
	suite.Require().Equal(sdk.TokensFromConsensusPower(600, sdk.DefaultPowerReduction), beforeBurnTotal.Amount)

	err := burnKeeper.BurnCoins(ctx, address[1].String())

	suite.Require().NoError(err)

	afterBurnCoin1 := bankKeeper.GetBalance(ctx, address[0], assets.MicroMedDenom)
	afterBurnCoin2 := bankKeeper.GetBalance(ctx, address[1], assets.MicroMedDenom)
	afterBurnCoin3 := bankKeeper.GetBalance(ctx, address[2], assets.MicroMedDenom)
	afterBurnTotal := bankKeeper.GetSupply(ctx, assets.MicroMedDenom)

	suite.Require().True(afterBurnCoin2.IsZero())
	suite.Require().Equal(initCoin, afterBurnCoin1)
	suite.Require().Equal(initCoin, afterBurnCoin3)
	suite.Require().Equal(assets.MicroMedDenom, afterBurnTotal.GetDenom())
	suite.Require().Equal(sdk.TokensFromConsensusPower(400, sdk.DefaultPowerReduction), afterBurnTotal.Amount)
}

func (suite *BurnTestSuite) TestGetAccount_NotExistAddress() {
	ctx := suite.Ctx
	bankKeeper := suite.BankKeeper
	burnKeeper := suite.BurnKeeper

	beforeBurnCoin1 := bankKeeper.GetBalance(ctx, address[0], assets.MicroMedDenom)
	beforeBurnCoin2 := bankKeeper.GetBalance(ctx, address[1], assets.MicroMedDenom)
	beforeBurnCoin3 := bankKeeper.GetBalance(ctx, address[2], assets.MicroMedDenom)
	beforeBurnTotal := bankKeeper.GetSupply(ctx, assets.MicroMedDenom)

	suite.Require().Equal(initCoin, beforeBurnCoin1)
	suite.Require().Equal(initCoin, beforeBurnCoin2)
	suite.Require().Equal(initCoin, beforeBurnCoin3)
	suite.Require().Equal(assets.MicroMedDenom, beforeBurnTotal.GetDenom())
	suite.Require().Equal(sdk.TokensFromConsensusPower(600, sdk.DefaultPowerReduction), beforeBurnTotal.Amount)

	err := burnKeeper.BurnCoins(ctx, "invalid address")

	suite.Require().Error(err)

	afterBurnCoin1 := bankKeeper.GetBalance(ctx, address[0], assets.MicroMedDenom)
	afterBurnCoin2 := bankKeeper.GetBalance(ctx, address[1], assets.MicroMedDenom)
	afterBurnCoin3 := bankKeeper.GetBalance(ctx, address[2], assets.MicroMedDenom)
	afterBurnTotal := bankKeeper.GetSupply(ctx, assets.MicroMedDenom)

	suite.Require().Equal(initCoin, afterBurnCoin1)
	suite.Require().Equal(initCoin, afterBurnCoin2)
	suite.Require().Equal(initCoin, afterBurnCoin3)
	suite.Require().Equal(assets.MicroMedDenom, afterBurnTotal.GetDenom())
	suite.Require().Equal(sdk.TokensFromConsensusPower(600, sdk.DefaultPowerReduction), afterBurnTotal.Amount)

}
