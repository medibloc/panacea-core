package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	tokentypes "github.com/medibloc/panacea-core/x/token/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"testing"

	"github.com/medibloc/panacea-core/types/testsuite"
	"github.com/stretchr/testify/suite"
)

var (
	pubKey = secp256k1.GenPrivKey().PubKey()
	ownerAddress = sdk.AccAddress(pubKey.Address())
)

type tokenTestSuite struct {
	testsuite.TestSuite
}

func TestTokenTestSuite(t *testing.T) {
	suite.Run(t, new(tokenTestSuite))


}

func (suite *tokenTestSuite) BeforeTest(_, _ string) {
	ctx := suite.Ctx
	bankKeeper := suite.BankKeeper

	bankKeeper.SetSupply(ctx, banktypes.DefaultSupply())
}

func (suite tokenTestSuite) TestTokenOneSymbol() {
	ctx := suite.Ctx
	tokenKeeper := suite.TokenKeeper
	bankKeeper := suite.BankKeeper

	totalCoin := sdk.NewCoin("umedc", sdk.TokensFromConsensusPower(1000000000000000))
	token := tokentypes.Token{
		Name:         "MediBloc_Child",
		Symbol:       "medc",
		TotalSupply:  totalCoin,
		Mintable:     true,
		OwnerAddress: ownerAddress.String(),

	}

	tokenKeeper.SetToken(ctx, token)

	// verify
	supply := bankKeeper.GetSupply(ctx)
	suite.Require().Equal(sdk.NewCoins(token.GetTotalSupply()), supply.GetTotal())

	ownerCoin := bankKeeper.GetBalance(ctx, ownerAddress, totalCoin.Denom)
	suite.Require().Equal(totalCoin.Denom, ownerCoin.Denom)
	suite.Require().Equal(totalCoin.Amount, ownerCoin.Amount)

	// verify HasToken
	suite.Require().True(tokenKeeper.HasToken(ctx, token.Symbol))

	// verify GetToken
	resultToken := tokenKeeper.GetToken(ctx, token.Symbol)
	suite.Require().Equal(token, resultToken)

	// verify AllToken
	resultTokens := tokenKeeper.GetAllToken(ctx)
	suite.Require().Equal([]tokentypes.Token{token}, resultTokens)
}

func (suite tokenTestSuite) TestTokenMultiSymbol() {
	ctx := suite.Ctx
	tokenKeeper := suite.TokenKeeper
	bankKeeper := suite.BankKeeper

	medc1TotalCoin := sdk.NewCoin("umedc1", sdk.TokensFromConsensusPower(1000000000000000))
	medc1Token := tokentypes.Token{
		Name:         "MediBloc_Child",
		Symbol:       "medc1",
		TotalSupply:  medc1TotalCoin,
		Mintable:     true,
		OwnerAddress: ownerAddress.String(),
	}

	medc2TotalCoin := sdk.NewCoin("umedc2", sdk.TokensFromConsensusPower(1000000000000000))
	medc2Token := tokentypes.Token{
		Name:         "MediBloc_Child",
		Symbol:       "medc2",
		TotalSupply:  medc2TotalCoin,
		Mintable:     true,
		OwnerAddress: ownerAddress.String(),
	}

	tokenKeeper.SetToken(ctx, medc1Token)
	tokenKeeper.SetToken(ctx, medc2Token)

	// verify
	supply := bankKeeper.GetSupply(ctx)
	suite.Require().Equal(sdk.NewCoins(medc1Token.GetTotalSupply(), medc2Token.GetTotalSupply()), supply.GetTotal())

	// verify HasToken
	suite.Require().True(tokenKeeper.HasToken(ctx, medc1Token.Symbol))
	suite.Require().True(tokenKeeper.HasToken(ctx, medc2Token.Symbol))

	// verify GetToken
	resultMedi1Token := tokenKeeper.GetToken(ctx, medc1Token.Symbol)
	resultMedi2Token := tokenKeeper.GetToken(ctx, medc2Token.Symbol)
	suite.Require().Equal(medc1Token, resultMedi1Token)
	suite.Require().Equal(medc2Token, resultMedi2Token)

	// verify AllToken
	resultTokens := tokenKeeper.GetAllToken(ctx)
	suite.Require().Equal([]tokentypes.Token{medc1Token, medc2Token}, resultTokens)
}

func (suite tokenTestSuite) TestInvalidFromAddress() {
	defer func() {
		recover()
	}()

	ctx := suite.Ctx
	tokenKeeper := suite.TokenKeeper

	totalCoin := sdk.NewCoin("umedc", sdk.TokensFromConsensusPower(1000000000000000))
	token := tokentypes.Token{
		Name:         "MediBloc_Child",
		Symbol:       "medc",
		TotalSupply:  totalCoin,
		Mintable:     true,
		OwnerAddress: "invalid address",

	}

	tokenKeeper.SetToken(ctx, token)

	// if not panic, then test error!
	suite.T().Error("did not panic")
}

func (suite tokenTestSuite) TestInvalidNewCoin() {
	defer func() {
		recover()
	}()

	ctx := suite.Ctx
	tokenKeeper := suite.TokenKeeper

	totalCoin := sdk.NewCoin("", sdk.TokensFromConsensusPower(1000000000000000))
	token := tokentypes.Token{
		Name:         "MediBloc_Child",
		Symbol:       "medc",
		TotalSupply:  totalCoin,
		Mintable:     true,
		OwnerAddress: ownerAddress.String(),

	}

	tokenKeeper.SetToken(ctx, token)

	// if not panic, then test error!
	suite.T().Error("did not panic")
}
