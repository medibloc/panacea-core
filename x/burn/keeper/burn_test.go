package keeper_test

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/medibloc/panacea-core/types/testsuite"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/medibloc/panacea-core/types/assets"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
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

	initTokens = sdk.TokensFromConsensusPower(200)
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
		err := bankKeeper.AddCoins(ctx, addr, initCoins)
		suite.Require().NoError(err)
	}

	bankKeeper.SetSupply(ctx, banktypes.DefaultSupply())
	supply := bankKeeper.GetSupply(ctx)
	supply.SetTotal(sdk.NewCoins(sdk.NewCoin("umed", sdk.TokensFromConsensusPower(600))))
	bankKeeper.SetSupply(ctx, supply)

	totalCoins := bankKeeper.GetSupply(ctx).GetTotal()

	suite.Require().Equal(1, len(totalCoins))
	suite.Require().Equal(assets.MicroMedDenom, totalCoins[0].Denom)
	suite.Require().Equal(sdk.TokensFromConsensusPower(600), totalCoins[0].Amount)
}

func (suite *BurnTestSuite) TestBurnCoins() {
	ctx := suite.Ctx
	bankKeeper := suite.BankKeeper
	burnKeeper := suite.BurnKeeper

	beforeBurnCoin1 := bankKeeper.GetBalance(ctx, address[0], assets.MicroMedDenom)
	beforeBurnCoin2 := bankKeeper.GetBalance(ctx, address[1], assets.MicroMedDenom)
	beforeBurnCoin3 := bankKeeper.GetBalance(ctx, address[2], assets.MicroMedDenom)
	beforeBurnTotal := bankKeeper.GetSupply(ctx).GetTotal()

	suite.Require().Equal(initCoin, beforeBurnCoin1)
	suite.Require().Equal(initCoin, beforeBurnCoin2)
	suite.Require().Equal(initCoin, beforeBurnCoin3)
	suite.Require().Equal(assets.MicroMedDenom, beforeBurnTotal[0].GetDenom())
	suite.Require().Equal(sdk.TokensFromConsensusPower(600), beforeBurnTotal[0].Amount)

	err := burnKeeper.BurnCoins(ctx, address[1].String())

	suite.Require().NoError(err)

	afterBurnCoin1 := bankKeeper.GetBalance(ctx, address[0], assets.MicroMedDenom)
	afterBurnCoin2 := bankKeeper.GetBalance(ctx, address[1], assets.MicroMedDenom)
	afterBurnCoin3 := bankKeeper.GetBalance(ctx, address[2], assets.MicroMedDenom)
	afterBurnTotal := bankKeeper.GetSupply(ctx).GetTotal()

	suite.Require().True(afterBurnCoin2.IsZero())
	suite.Require().Equal(initCoin, afterBurnCoin1)
	suite.Require().Equal(initCoin, afterBurnCoin3)
	suite.Require().Equal(assets.MicroMedDenom, afterBurnTotal[0].GetDenom())
	suite.Require().Equal(sdk.TokensFromConsensusPower(400), afterBurnTotal[0].Amount)
}

func (suite *BurnTestSuite) TestGetAccount_NotExistAddress() {
	ctx := suite.Ctx
	bankKeeper := suite.BankKeeper
	burnKeeper := suite.BurnKeeper

	beforeBurnCoin1 := bankKeeper.GetBalance(ctx, address[0], assets.MicroMedDenom)
	beforeBurnCoin2 := bankKeeper.GetBalance(ctx, address[1], assets.MicroMedDenom)
	beforeBurnCoin3 := bankKeeper.GetBalance(ctx, address[2], assets.MicroMedDenom)
	beforeBurnTotal := bankKeeper.GetSupply(ctx).GetTotal()

	suite.Require().Equal(initCoin, beforeBurnCoin1)
	suite.Require().Equal(initCoin, beforeBurnCoin2)
	suite.Require().Equal(initCoin, beforeBurnCoin3)
	suite.Require().Equal(assets.MicroMedDenom, beforeBurnTotal[0].GetDenom())
	suite.Require().Equal(sdk.TokensFromConsensusPower(600), beforeBurnTotal[0].Amount)

	err := burnKeeper.BurnCoins(ctx, "invalid address")

	suite.Require().Error(err)

	afterBurnCoin1 := bankKeeper.GetBalance(ctx, address[0], assets.MicroMedDenom)
	afterBurnCoin2 := bankKeeper.GetBalance(ctx, address[1], assets.MicroMedDenom)
	afterBurnCoin3 := bankKeeper.GetBalance(ctx, address[2], assets.MicroMedDenom)
	afterBurnTotal := bankKeeper.GetSupply(ctx).GetTotal()

	suite.Require().Equal(initCoin, afterBurnCoin1)
	suite.Require().Equal(initCoin, afterBurnCoin2)
	suite.Require().Equal(initCoin, afterBurnCoin3)
	suite.Require().Equal(assets.MicroMedDenom, afterBurnTotal[0].GetDenom())
	suite.Require().Equal(sdk.TokensFromConsensusPower(600), afterBurnTotal[0].Amount)

}
