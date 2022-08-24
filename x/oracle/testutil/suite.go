package testutil

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

type OracleBaseTestSuite struct {
	testsuite.TestSuite
}

// CreateOracleValidator defines to register Oracle and Validator by default.
func (suite *OracleBaseTestSuite) CreateOracleValidator(pubKey cryptotypes.PubKey, amount sdk.Int) {
	suite.SetAccount(pubKey)

	suite.SetValidator(pubKey, amount)

	oracleAccAddr := sdk.AccAddress(pubKey.Address().Bytes())
	oracle := &types.Oracle{
		Address:  oracleAccAddr.String(),
		Status:   types.ORACLE_STATUS_ACTIVE,
		Uptime:   0,
		JailedAt: nil,
	}

	suite.Require().NoError(suite.OracleKeeper.SetOracle(suite.Ctx, oracle))
}

func (suite *OracleBaseTestSuite) SetAccount(pubKey cryptotypes.PubKey) {
	oracleAccAddr := sdk.AccAddress(pubKey.Address().Bytes())
	oracleAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, oracleAccAddr)
	suite.Require().NoError(oracleAccount.SetPubKey(pubKey))
	suite.AccountKeeper.SetAccount(suite.Ctx, oracleAccount)
}

func (suite *OracleBaseTestSuite) SetValidator(pubKey cryptotypes.PubKey, amount sdk.Int) {
	varAddr := sdk.ValAddress(pubKey.Address().Bytes())
	validator, err := stakingtypes.NewValidator(varAddr, pubKey, stakingtypes.Description{})
	suite.Require().NoError(err)
	validator = validator.UpdateStatus(stakingtypes.Bonded)
	validator, _ = validator.AddTokensFromDel(amount)

	suite.StakingKeeper.SetValidator(suite.Ctx, validator)
	err = suite.StakingKeeper.SetValidatorByConsAddr(suite.Ctx, validator)
	suite.Require().NoError(err)
}
