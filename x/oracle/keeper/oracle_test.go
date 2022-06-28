package keeper_test

import (
	"sort"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
)

var (
	oraclePrivKey = secp256k1.GenPrivKey()
	oraclePubKey  = oraclePrivKey.PubKey()
	oracle1       = sdk.AccAddress(oraclePubKey.Address())
	oracle2       = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	oracle3       = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	oracle4       = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	oracle5       = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	fundForOracle = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)))
)

type oracleTestSuite struct {
	testsuite.TestSuite
}

func TestOracleTestSuite(t *testing.T) {
	suite.Run(t, new(oracleTestSuite))
}

func (suite *oracleTestSuite) TestRegisterOracle() {
	err := suite.FundAccount(suite.BankKeeper, suite.Ctx, oracle1, fundForOracle)
	suite.Require().NoError(err)

	suite.setOracleAccount(oracle1)

	tempOracle := types.Oracle{
		Address:  oracle1.String(),
		Endpoint: "https://my-oracle.org",
	}

	err = suite.OracleKeeper.RegisterOracle(suite.Ctx, tempOracle)
	suite.Require().NoError(err)
}

func (suite *oracleTestSuite) setOracleAccount(oracleAddr sdk.AccAddress) {
	oracleAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, oracleAddr)
	err := oracleAccount.SetPubKey(oraclePubKey)
	suite.Require().NoError(err)
	suite.AccountKeeper.SetAccount(suite.Ctx, oracleAccount)
}

func (suite *oracleTestSuite) TestGetRegisterOracle() {
	err := suite.FundAccount(suite.BankKeeper, suite.Ctx, oracle1, fundForOracle)
	suite.Require().NoError(err)

	suite.setOracleAccount(oracle1)

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
	err := suite.FundAccount(suite.BankKeeper, suite.Ctx, oracle1, fundForOracle)
	suite.Require().NoError(err)

	suite.setOracleAccount(oracle1)

	tempOracle := types.Oracle{
		Address:  oracle1.String(),
		Endpoint: "https://my-oracle.org",
	}

	err = suite.OracleKeeper.RegisterOracle(suite.Ctx, tempOracle)
	suite.Require().NoError(err)

	err = suite.OracleKeeper.RegisterOracle(suite.Ctx, tempOracle)
	suite.Require().Error(err, types.ErrOracleAlreadyExist)
}

func (suite *oracleTestSuite) TestGetOracle() {
	err := suite.FundAccount(suite.BankKeeper, suite.Ctx, oracle1, fundForOracle)
	suite.Require().NoError(err)

	suite.setOracleAccount(oracle1)

	tempOracle := types.Oracle{
		Address:  oracle1.String(),
		Endpoint: "https://my-oracle.org",
	}

	err = suite.OracleKeeper.RegisterOracle(suite.Ctx, tempOracle)
	suite.Require().NoError(err)

	getOracle, err := suite.OracleKeeper.GetOracle(suite.Ctx, oracle1)
	suite.Require().NoError(err)

	suite.Require().Equal(tempOracle.Address, getOracle.Address)
	suite.Require().Equal(tempOracle.Endpoint, getOracle.Endpoint)
}

func (suite *oracleTestSuite) TestOracleNotFound() {
	err := suite.FundAccount(suite.BankKeeper, suite.Ctx, oracle1, fundForOracle)
	suite.Require().NoError(err)

	suite.setOracleAccount(oracle1)

	_, err = suite.OracleKeeper.GetOracle(suite.Ctx, oracle1)
	suite.Require().Error(err, types.ErrOracleNotFound)
}

func (suite *oracleTestSuite) TestGetAllOracles() {
	oracles := [5]sdk.AccAddress{oracle1, oracle2, oracle3, oracle4, oracle5}

	for _, o := range oracles {
		err := suite.FundAccount(suite.BankKeeper, suite.Ctx, o, fundForOracle)
		suite.Require().NoError(err)

		suite.setOracleAccount(o)
	}

	for _, o := range oracles {
		tempOracle := types.Oracle{
			Address:  o.String(),
			Endpoint: "https://my-oracle.org",
		}

		err := suite.OracleKeeper.RegisterOracle(suite.Ctx, tempOracle)
		suite.Require().NoError(err)
	}

	allOracles, err := suite.OracleKeeper.GetAllOracles(suite.Ctx)
	suite.Require().NoError(err)

	var allOracleStr []string
	var tempOracleStr []string

	for i := 0; i < 5; i++ {
		allOracleStr = append(allOracleStr, allOracles[i].Address)
		tempOracleStr = append(tempOracleStr, oracles[i].String())
	}

	sort.Strings(allOracleStr)
	sort.Strings(tempOracleStr)

	for i := 0; i < 5; i++ {
		suite.Require().Equal(allOracleStr[i], tempOracleStr[i])
		suite.Require().Equal(allOracles[i].Endpoint, "https://my-oracle.org")
	}
}

func (suite *oracleTestSuite) TestUpdateOracle() {
	err := suite.FundAccount(suite.BankKeeper, suite.Ctx, oracle1, fundForOracle)
	suite.Require().NoError(err)

	suite.setOracleAccount(oracle1)

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

func (suite *oracleTestSuite) TestUpdateOracle_invalid_requester() {
	err := suite.FundAccount(suite.BankKeeper, suite.Ctx, oracle1, fundForOracle)
	suite.Require().NoError(err)

	suite.setOracleAccount(oracle1)

	tempOracle := types.Oracle{
		Address:  oracle1.String(),
		Endpoint: "https://my-oracle.org",
	}

	err = suite.OracleKeeper.RegisterOracle(suite.Ctx, tempOracle)
	suite.Require().NoError(err)

	updateTempOracle := types.Oracle{
		Address:  oracle2.String(),
		Endpoint: "https://update-my-oracle.org",
	}

	err = suite.OracleKeeper.UpdateOracle(suite.Ctx, oracle2, updateTempOracle.Endpoint)
	suite.Require().Error(err, types.ErrInvalidUpdateRequester)
}
