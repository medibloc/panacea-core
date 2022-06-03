package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
	"testing"
)

var (
	oraclePrivKey = secp256k1.GenPrivKey()
	oraclePubKey  = oraclePrivKey.PubKey()
	oracle1       = sdk.AccAddress(oraclePubKey.Address())

	fundForOracle = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)))
)

type oracleTestSuite struct {
	testsuite.TestSuite
}

func TestOracleTestSuite(t *testing.T) {
	suite.Run(t, new(oracleTestSuite))
}

func (suite *oracleTestSuite) TestRegisterOracle() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, oracle1, fundForOracle)
	suite.Require().NoError(err)

	suite.setOracleAccount()

	tempOracle := types.Oracle{
		Address:  oracle1.String(),
		Endpoint: "https://my-oracle.org",
	}

	err = suite.OracleKeeper.RegisterOracle(suite.Ctx, tempOracle)
	suite.Require().NoError(err)
}

func (suite *oracleTestSuite) setOracleAccount() {
	oracleAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, oracle1)
	err := oracleAccount.SetPubKey(oraclePubKey)
	suite.Require().NoError(err)
	suite.AccountKeeper.SetAccount(suite.Ctx, oracleAccount)
}

func (suite *oracleTestSuite) TestGetRegisterOracle() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, oracle1, fundForOracle)
	suite.Require().NoError(err)

	suite.setOracleAccount()

	tempOracle := types.Oracle{
		Address:  oracle1.String(),
		Endpoint: "https://my-oracle.org",
	}

	err = suite.OracleKeeper.RegisterOracle(suite.Ctx, tempOracle)
	suite.Require().NoError(err)

	getOracle, err := suite.OracleKeeper.GetOracle(suite.Ctx, oracle1)
	suite.Require().NoError(err)
	suite.Require().Equal(tempOracle.Endpoint, getOracle.Endpoint)
}

func (suite *oracleTestSuite) TestIsOracleDuplicate() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, oracle1, fundForOracle)
	suite.Require().NoError(err)

	suite.setOracleAccount()

	tempOracle := types.Oracle{
		Address:  oracle1.String(),
		Endpoint: "https://my-oralce.org",
	}

	err = suite.OracleKeeper.RegisterOracle(suite.Ctx, tempOracle)
	suite.Require().NoError(err)

	err = suite.OracleKeeper.RegisterOracle(suite.Ctx, tempOracle)
	suite.Require().Error(err, types.ErrOracleAlreadyExist)
}

func (suite *oracleTestSuite) TestUpdateOracle() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, oracle1, fundForOracle)
	suite.Require().NoError(err)

	suite.setOracleAccount()

	tempOracle := types.Oracle{
		Address:  oracle1.String(),
		Endpoint: "https://my-oracle.org",
	}

	err = suite.OracleKeeper.RegisterOracle(suite.Ctx, tempOracle)
	suite.Require().NoError(err)

	updateTempOracle := types.Oracle{
		Address:  oracle1.String(),
		Endpoint: "https://update-my-oracle.org",
	}

	err = suite.OracleKeeper.UpdateOracle(suite.Ctx, oracle1, updateTempOracle.Endpoint)
	suite.Require().NoError(err)

	getOracle, err := suite.OracleKeeper.GetOracle(suite.Ctx, oracle1)
	suite.Require().NoError(err)

	suite.Require().Equal(getOracle.GetAddress(), updateTempOracle.GetAddress())
	suite.Require().Equal(getOracle.GetEndpoint(), updateTempOracle.GetEndpoint())
}
