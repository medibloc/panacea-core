package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/market/types"
	"github.com/stretchr/testify/suite"
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
	acc3                   = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	defaultFunds sdk.Coins = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)))
	zeroFunds    sdk.Coins = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(0)))
	privKey                = secp256k1.GenPrivKey()
	newAddr                = sdk.AccAddress(privKey.PubKey().Address())
)

func (suite *dealTestSuite) BeforeTest(_, _ string) {
	testDeal := makeTestDeal()
	suite.MarketKeeper.SetNextDealNumber(suite.Ctx, 2)
	suite.MarketKeeper.SetDeal(suite.Ctx, testDeal)
}

func (suite *dealTestSuite) TestCreateNewDeal() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, acc1, defaultFunds)
	suite.Require().NoError(err)

	tempDeal := types.Deal{
		DataSchema:            []string{"http://jsonld.com"},
		Budget:                &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(10000000)},
		MaxNumData:            10000,
		TrustedDataValidators: []string{acc2.String()},
		Owner:                 acc1.String(),
	}

	owner, err := sdk.AccAddressFromBech32(tempDeal.GetOwner())
	suite.Require().NoError(err)

	dealId, err := suite.MarketKeeper.CreateNewDeal(suite.Ctx, owner, tempDeal)
	suite.Require().NoError(err)

	expectedId := suite.MarketKeeper.GetNextDealNumberAndIncrement(suite.Ctx) - 1
	suite.Require().Equal(dealId, expectedId)

	deal, err := suite.MarketKeeper.GetDeal(suite.Ctx, dealId)
	suite.Require().NoError(err)
	suite.Require().Equal(deal.GetDataSchema(), tempDeal.GetDataSchema())
	suite.Require().Equal(deal.GetBudget(), tempDeal.GetBudget())
	suite.Require().Equal(deal.GetMaxNumData(), tempDeal.GetMaxNumData())
	suite.Require().Equal(deal.GetCurNumData(), tempDeal.GetCurNumData())
	suite.Require().Equal(deal.GetTrustedDataValidators(), tempDeal.GetTrustedDataValidators())
	suite.Require().Equal(deal.GetOwner(), tempDeal.GetOwner())
	suite.Require().Equal(deal.GetStatus(), ACTIVE)
}

func (suite *dealTestSuite) TestGetDeal() {
	deal, err := suite.MarketKeeper.GetDeal(suite.Ctx, 1)
	suite.Require().NoError(err)
	testDeal := makeTestDeal()

	suite.Require().Equal(deal.GetDealId(), testDeal.GetDealId())
	suite.Require().Equal(deal.GetDealAddress(), testDeal.GetDealAddress())
	suite.Require().Equal(deal.GetDataSchema(), testDeal.GetDataSchema())
	suite.Require().Equal(deal.GetBudget(), testDeal.GetBudget())
	suite.Require().Equal(deal.GetMaxNumData(), testDeal.GetMaxNumData())
	suite.Require().Equal(deal.GetCurNumData(), testDeal.GetCurNumData())
	suite.Require().Equal(deal.GetTrustedDataValidators(), testDeal.GetTrustedDataValidators())
	suite.Require().Equal(deal.GetOwner(), testDeal.GetOwner())
	suite.Require().Equal(deal.GetStatus(), testDeal.GetStatus())
}

func (suite *dealTestSuite) TestGetBalanceOfDeal() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, acc1, defaultFunds)
	suite.Require().NoError(err)

	tempDeal := types.Deal{
		DataSchema:            []string{"http://jsonld.com"},
		Budget:                &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(10000000)},
		MaxNumData:            10000,
		TrustedDataValidators: []string{acc2.String()},
		Owner:                 acc1.String(),
	}

	owner, err := sdk.AccAddressFromBech32(tempDeal.GetOwner())
	suite.Require().NoError(err)

	dealId, err := suite.MarketKeeper.CreateNewDeal(suite.Ctx, owner, tempDeal)
	suite.Require().NoError(err)

	deal, err := suite.MarketKeeper.GetDeal(suite.Ctx, dealId)
	suite.Require().NoError(err)

	addr, err := types.AccDealAddressFromBech32(deal.GetDealAddress())
	suite.Require().NoError(err)

	balance := suite.BankKeeper.GetBalance(suite.Ctx, addr, assets.MicroMedDenom)
	suite.Require().Equal(balance, *tempDeal.GetBudget())

	ownerBalance := suite.BankKeeper.GetBalance(suite.Ctx, acc1, assets.MicroMedDenom)
	suite.Require().Equal(ownerBalance, sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)).Sub(balance))
}

func (suite *dealTestSuite) TestSellOwnData() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, acc1, defaultFunds)
	suite.Require().NoError(err)

	err = suite.BankKeeper.AddCoins(suite.Ctx, acc3, zeroFunds)
	suite.Require().NoError(err)

	tempDeal := types.Deal{
		DataSchema:            []string{"http://jsonld.com"},
		Budget:                &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(10000000)},
		MaxNumData:            10000,
		TrustedDataValidators: []string{newAddr.String()},
		Owner:                 acc1.String(),
	}

	newDealId, err := suite.MarketKeeper.CreateNewDeal(suite.Ctx, acc1, tempDeal)
	suite.Require().NoError(err)

	cert := makeTestCert()
	deal, err := suite.MarketKeeper.GetDeal(suite.Ctx, newDealId)
	suite.Require().NoError(err)

	reward, err := suite.MarketKeeper.SellOwnData(suite.Ctx, acc3, cert)
	suite.Require().NoError(err)
	suite.Require().Equal(cert.UnsignedCert.GetDealId(), deal.GetDealId())

	sellerBalance := suite.BankKeeper.GetBalance(suite.Ctx, acc3, assets.MicroMedDenom)
	suite.Require().Equal(sellerBalance, reward)

	updatedDeal, err := suite.MarketKeeper.GetDeal(suite.Ctx, newDealId)
	suite.Require().NoError(err)
	suite.Require().Equal(updatedDeal.GetCurNumData(), deal.GetCurNumData()+1)
}

func (suite *dealTestSuite) TestIsDataCertDuplicate() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, acc1, defaultFunds)
	suite.Require().NoError(err)

	err = suite.BankKeeper.AddCoins(suite.Ctx, acc3, zeroFunds)
	suite.Require().NoError(err)

	tempDeal := types.Deal{
		DataSchema:            []string{"http://jsonld.com"},
		Budget:                &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(10000000)},
		MaxNumData:            10000,
		TrustedDataValidators: []string{newAddr.String()},
		Owner:                 acc1.String(),
	}

	_, err = suite.MarketKeeper.CreateNewDeal(suite.Ctx, acc1, tempDeal)
	suite.Require().NoError(err)

	testCert1 := makeTestCert()
	_, err = suite.MarketKeeper.SellOwnData(suite.Ctx, acc3, testCert1)
	suite.Require().NoError(err)

	testCert2 := makeTestCert()
	_, err = suite.MarketKeeper.SellOwnData(suite.Ctx, acc3, testCert2)
	suite.Require().Error(err, "duplicated data")
}

func (suite *dealTestSuite) TestIsTrustedDataValidator_Invalid() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, acc1, defaultFunds)
	suite.Require().NoError(err)

	err = suite.BankKeeper.AddCoins(suite.Ctx, acc3, zeroFunds)
	suite.Require().NoError(err)

	errValidator1 := secp256k1.GenPrivKey().PubKey().Address().String()
	errValidator2 := secp256k1.GenPrivKey().PubKey().Address().String()

	tempDeal := types.Deal{
		DataSchema:            []string{"http://jsonld.com"},
		Budget:                &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(10000000)},
		MaxNumData:            10000,
		TrustedDataValidators: []string{errValidator1,  errValidator2},
		Owner:                 acc1.String(),
	}

	_, err = suite.MarketKeeper.CreateNewDeal(suite.Ctx, acc1, tempDeal)
	suite.Require().NoError(err)

	testCert1 := makeTestCert()
	_, err = suite.MarketKeeper.SellOwnData(suite.Ctx, acc3, testCert1)
	suite.Require().Error(err, "data validator is invalid address")
}

func (suite *dealTestSuite) TestDealStatusInactiveOrCompleted() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, acc1, defaultFunds)
	suite.Require().NoError(err)

	err = suite.BankKeeper.AddCoins(suite.Ctx, acc3, zeroFunds)
	suite.Require().NoError(err)

	tempDeal := types.Deal{
		DataSchema:            []string{"http://jsonld.com"},
		Budget:                &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(10000000)},
		MaxNumData:            10000,
		TrustedDataValidators: []string{newAddr.String()},
		Owner:                 acc1.String(),
	}

	dealId, err := suite.MarketKeeper.CreateNewDeal(suite.Ctx, acc1, tempDeal)
	suite.Require().NoError(err)
	findDeal, err := suite.MarketKeeper.GetDeal(suite.Ctx, dealId)
	suite.Require().NoError(err)

	findDeal.Status = INACTIVE
	suite.MarketKeeper.SetDeal(suite.Ctx, findDeal)

	testCert1 := makeTestCert()
	_, err = suite.MarketKeeper.SellOwnData(suite.Ctx, acc3, testCert1)
	suite.Require().Error(err, "the deal's state is INACTIVE")

	findDeal.Status = COMPLETED
	suite.MarketKeeper.SetDeal(suite.Ctx, findDeal)
	suite.Require().Error(err, "the deal's state is COMPLETED")
}

func (suite *dealTestSuite) TestVerify() {
	cert := makeTestCert()

	validatorAddr, err := sdk.AccAddressFromBech32(cert.UnsignedCert.GetDataValidatorAddress())
	suite.Require().NoError(err)
	suite.Require().Equal(newAddr, validatorAddr)

	account := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, validatorAddr)
	err = account.SetPubKey(privKey.PubKey())
	suite.Require().NoError(err)
	suite.AccountKeeper.SetAccount(suite.Ctx, account)

	verify, err := suite.MarketKeeper.Verify(suite.Ctx, validatorAddr, cert)
	suite.Require().Equal(true, verify)
	suite.Require().NoError(err)
}

func makeTestDeal() types.Deal {
	return types.Deal{
		DealId:                1,
		DealAddress:           types.NewDealAddress(1).String(),
		DataSchema:            []string{acc1.String()},
		Budget:                &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(1000000000)},
		MaxNumData:            10000,
		CurNumData:            0,
		TrustedDataValidators: []string{acc2.String()},
		Owner:                 acc1.String(),
		Status:                ACTIVE,
	}
}

func makeTestCert() types.DataValidationCertificate {
	uCert := types.UnsignedDataValidationCertificate{
		DealId:               2,
		DataHash:             "1a312c123x23",
		EncryptedDataUrl:     "https://panacea.org/a/123.json",
		DataValidatorAddress: newAddr.String(),
		RequesterAddress:     acc3.String(),
	}

	marshal, err := uCert.Marshal()
	if err != nil {
		return types.DataValidationCertificate{}
	}

	sign, err := privKey.Sign(marshal)
	if err != nil {
		return types.DataValidationCertificate{}
	}

	return types.DataValidationCertificate{
		UnsignedCert: &uCert,
		Signature:    sign,
	}

}
