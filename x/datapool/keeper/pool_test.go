package keeper_test

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"

	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/stretchr/testify/suite"
)

type poolTestSuite struct {
	testsuite.TestSuite
}

func TestPoolTestSuite(t *testing.T) {
	suite.Run(t, new(poolTestSuite))
}

var (
	privKey        = secp256k1.GenPrivKey()
	pubKey         = privKey.PubKey()
	dataVal1       = sdk.AccAddress(pubKey.Address())
	fundForDataVal = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)))
)

func (suite *poolTestSuite) TestRegisterDataValidator() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, dataVal1, fundForDataVal)
	suite.Require().NoError(err)

	validatorAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, dataVal1)
	err = validatorAccount.SetPubKey(pubKey)
	suite.Require().NoError(err)
	suite.AccountKeeper.SetAccount(suite.Ctx, validatorAccount)

	tempDataValidator := types.DataValidator{
		Address:  dataVal1.String(),
		Endpoint: "https://my-validator.org",
	}

	err = suite.DataPoolKeeper.RegisterDataValidator(suite.Ctx, tempDataValidator)
	suite.Require().NoError(err)
}

func (suite *poolTestSuite) TestGetRegisterDataValidator() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, dataVal1, fundForDataVal)
	suite.Require().NoError(err)

	validatorAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, dataVal1)
	err = validatorAccount.SetPubKey(pubKey)
	suite.Require().NoError(err)
	suite.AccountKeeper.SetAccount(suite.Ctx, validatorAccount)

	tempDataValidatorDetail := types.DataValidator{
		Address:  dataVal1.String(),
		Endpoint: "https://my-validator.org",
	}

	err = suite.DataPoolKeeper.RegisterDataValidator(suite.Ctx, tempDataValidatorDetail)
	suite.Require().NoError(err)

	getDataValidator, err := suite.DataPoolKeeper.GetDataValidator(suite.Ctx, dataVal1)
	suite.Require().NoError(err)
	suite.Require().Equal(tempDataValidatorDetail.Endpoint, getDataValidator.Endpoint)
}

func (suite *poolTestSuite) TestIsDataValidatorDuplicate() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, dataVal1, fundForDataVal)
	suite.Require().NoError(err)

	validatorAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, dataVal1)
	err = validatorAccount.SetPubKey(pubKey)
	suite.Require().NoError(err)
	suite.AccountKeeper.SetAccount(suite.Ctx, validatorAccount)

	tempDataValidatorDetail := types.DataValidator{
		Address:  dataVal1.String(),
		Endpoint: "https://my-validator.org",
	}

	err = suite.DataPoolKeeper.RegisterDataValidator(suite.Ctx, tempDataValidatorDetail)
	suite.Require().NoError(err)

	err = suite.DataPoolKeeper.RegisterDataValidator(suite.Ctx, tempDataValidatorDetail)
	suite.Require().Error(err, types.ErrDataValidatorAlreadyExist)
}

func (suite *poolTestSuite) TestNotGetPubKey() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, dataVal1, fundForDataVal)
	suite.Require().NoError(err)

	tempDataValidatorDetail := types.DataValidator{
		Address:  dataVal1.String(),
		Endpoint: "https://my-validator.org",
	}

	err = suite.DataPoolKeeper.RegisterDataValidator(suite.Ctx, tempDataValidatorDetail)
	suite.Require().Error(err, sdkerrors.ErrKeyNotFound)
}

func (suite *poolTestSuite) TestUpdateDataValidator() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, dataVal1, fundForDataVal)
	suite.Require().NoError(err)

	validatorAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, dataVal1)
	err = validatorAccount.SetPubKey(pubKey)
	suite.Require().NoError(err)
	suite.AccountKeeper.SetAccount(suite.Ctx, validatorAccount)

	tempDataValidator := types.DataValidator{
		Address:  dataVal1.String(),
		Endpoint: "https://my-validator.org",
	}

	err = suite.DataPoolKeeper.RegisterDataValidator(suite.Ctx, tempDataValidator)
	suite.Require().NoError(err)

	updateTempDataValidator := types.DataValidator{
		Address:  dataVal1.String(),
		Endpoint: "https://update-my-validator.org",
	}

	err = suite.DataPoolKeeper.UpdateDataValidator(suite.Ctx, updateTempDataValidator)
	suite.Require().NoError(err)

	getDataValidator, err := suite.DataPoolKeeper.GetDataValidator(suite.Ctx, dataVal1)
	suite.Require().NoError(err)

	suite.Require().Equal(getDataValidator.GetAddress(), updateTempDataValidator.GetAddress())
	suite.Require().Equal(getDataValidator.GetEndpoint(), updateTempDataValidator.GetEndpoint())
}
