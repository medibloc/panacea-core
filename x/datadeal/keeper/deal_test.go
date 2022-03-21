package keeper_test

import (
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/stretchr/testify/suite"
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
	suite.DataDealKeeper.SetNextDealNumber(suite.Ctx, 2)
	suite.DataDealKeeper.SetDeal(suite.Ctx, testDeal)
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

	dealId, err := suite.DataDealKeeper.CreateNewDeal(suite.Ctx, owner, tempDeal)
	suite.Require().NoError(err)

	expectedId := suite.DataDealKeeper.GetNextDealNumberAndIncrement(suite.Ctx) - 1
	suite.Require().Equal(dealId, expectedId)

	deal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, dealId)
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
	deal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, 1)
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

func (suite *dealTestSuite) TestListDeals() {
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

	dealIds := make([]uint64, 0)

	for i := 0; i < 5; i++ {
		dealId, err := suite.DataDealKeeper.CreateNewDeal(suite.Ctx, owner, tempDeal)
		suite.Require().NoError(err)
		dealIds = append(dealIds, dealId)
	}

	deals, err := suite.DataDealKeeper.ListDeals(suite.Ctx)
	suite.Require().NoError(err)

	for i, dealId := range dealIds {
		deal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, dealId)
		suite.Require().NoError(err)

		suite.Require().Equal(deal.GetDealId(), deals[i+1].GetDealId())
		suite.Require().Equal(deal.GetDealAddress(), deals[i+1].GetDealAddress())
		suite.Require().Equal(deal.GetDataSchema(), deals[i+1].GetDataSchema())
		suite.Require().Equal(deal.GetBudget(), deals[i+1].GetBudget())
		suite.Require().Equal(deal.GetMaxNumData(), deals[i+1].GetMaxNumData())
		suite.Require().Equal(deal.GetCurNumData(), deals[i+1].GetCurNumData())
		suite.Require().Equal(deal.GetTrustedDataValidators(), deals[i+1].GetTrustedDataValidators())
		suite.Require().Equal(deal.GetOwner(), deals[i+1].GetOwner())
		suite.Require().Equal(deal.GetStatus(), deals[i+1].GetStatus())
	}
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

	dealId, err := suite.DataDealKeeper.CreateNewDeal(suite.Ctx, owner, tempDeal)
	suite.Require().NoError(err)

	deal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, dealId)
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

	newDealId, err := suite.DataDealKeeper.CreateNewDeal(suite.Ctx, acc1, tempDeal)
	suite.Require().NoError(err)

	deal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, newDealId)
	suite.Require().NoError(err)

	cert := makeTestCert("1a312c1223x2fs3", newAddr, acc3)
	reward, err := suite.DataDealKeeper.SellOwnData(suite.Ctx, acc3, cert)
	suite.Require().NoError(err)
	suite.Require().Equal(cert.UnsignedCert.GetDealId(), deal.GetDealId())

	sellerBalance := suite.BankKeeper.GetBalance(suite.Ctx, acc3, assets.MicroMedDenom)
	suite.Require().Equal(sellerBalance, reward)

	updatedDeal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, newDealId)
	suite.Require().NoError(err)
	suite.Require().Equal(updatedDeal.GetCurNumData(), deal.GetCurNumData()+1)
}

func (suite *dealTestSuite) TestIsDataCertDuplicate() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, acc1, sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000))))
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

	_, err = suite.DataDealKeeper.CreateNewDeal(suite.Ctx, acc1, tempDeal)
	suite.Require().NoError(err)

	testCert1 := makeTestCert("1a312c1223x", newAddr, acc3)
	_, err = suite.DataDealKeeper.SellOwnData(suite.Ctx, acc3, testCert1)
	suite.Require().NoError(err)

	testCert2 := makeTestCert("1a312c1223x", newAddr, acc3)
	_, err = suite.DataDealKeeper.SellOwnData(suite.Ctx, acc3, testCert2)
	suite.Require().Error(err, types.ErrDataAlreadyExist)
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
		TrustedDataValidators: []string{errValidator1, errValidator2},
		Owner:                 acc1.String(),
	}

	_, err = suite.DataDealKeeper.CreateNewDeal(suite.Ctx, acc1, tempDeal)
	suite.Require().NoError(err)

	testCert1 := makeTestCert("1a312c1223x2fs3", newAddr, acc3)
	_, err = suite.DataDealKeeper.SellOwnData(suite.Ctx, acc3, testCert1)
	suite.Require().Error(err, sdkerrors.ErrInvalidAddress)
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

	dealId, err := suite.DataDealKeeper.CreateNewDeal(suite.Ctx, acc1, tempDeal)
	suite.Require().NoError(err)
	findDeal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, dealId)
	suite.Require().NoError(err)

	findDeal.Status = INACTIVE
	suite.DataDealKeeper.SetDeal(suite.Ctx, findDeal)

	testCert1 := makeTestCert("1a312c1223x2fs3", newAddr, acc3)
	_, err = suite.DataDealKeeper.SellOwnData(suite.Ctx, acc3, testCert1)
	suite.Require().Error(err, types.ErrInvalidStatus)

	findDeal.Status = COMPLETED
	suite.DataDealKeeper.SetDeal(suite.Ctx, findDeal)
	suite.Require().Error(err, types.ErrInvalidStatus)
}

func (suite *dealTestSuite) TestVerifyDataCertificate() {
	cert := makeTestCert("1a312c1223x2fs3", newAddr, acc3)

	validatorAddr, err := sdk.AccAddressFromBech32(cert.UnsignedCert.GetDataValidatorAddress())
	suite.Require().NoError(err)
	suite.Require().Equal(newAddr, validatorAddr)

	account := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, validatorAddr)
	err = account.SetPubKey(privKey.PubKey())
	suite.Require().NoError(err)
	suite.AccountKeeper.SetAccount(suite.Ctx, account)

	verify, err := suite.DataDealKeeper.VerifyDataCertificate(suite.Ctx, validatorAddr, cert)
	suite.Require().Equal(true, verify)
	suite.Require().NoError(err)
}

func (suite *dealTestSuite) TestIsDealStatusCompleted() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, acc1, defaultFunds)
	suite.Require().NoError(err)

	err = suite.BankKeeper.AddCoins(suite.Ctx, acc2, zeroFunds)
	suite.Require().NoError(err)

	err = suite.BankKeeper.AddCoins(suite.Ctx, acc3, zeroFunds)
	suite.Require().NoError(err)

	tempDeal := types.Deal{
		DataSchema:            []string{"http://jsonld.com"},
		Budget:                &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(10000000)},
		MaxNumData:            2,
		TrustedDataValidators: []string{newAddr.String()},
		Owner:                 acc1.String(),
	}

	dealId, err := suite.DataDealKeeper.CreateNewDeal(suite.Ctx, acc1, tempDeal)
	suite.Require().NoError(err)

	testCert1 := makeTestCert("1a312c1223x2fs3", newAddr, acc3)
	_, err = suite.DataDealKeeper.SellOwnData(suite.Ctx, acc3, testCert1)
	suite.Require().NoError(err)

	testCert2 := makeTestCert("1a312c1223x", newAddr, acc2)
	_, err = suite.DataDealKeeper.SellOwnData(suite.Ctx, acc2, testCert2)
	suite.Require().NoError(err)

	updatedDeal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, dealId)
	suite.Require().NoError(err)

	suite.Require().Equal(updatedDeal.GetStatus(), COMPLETED)
}

func (suite *dealTestSuite) TestGetDataCertificate() {
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

	newDealId, err := suite.DataDealKeeper.CreateNewDeal(suite.Ctx, acc1, tempDeal)
	suite.Require().NoError(err)

	cert := makeTestCert("1a312c1223x2fs3", newAddr, acc3)
	_, err = suite.DataDealKeeper.GetDeal(suite.Ctx, newDealId)
	suite.Require().NoError(err)

	_, err = suite.DataDealKeeper.SellOwnData(suite.Ctx, acc3, cert)
	suite.Require().NoError(err)

	getCertificate, err := suite.DataDealKeeper.GetDataCertificate(suite.Ctx, cert)
	suite.Require().NoError(err)

	suite.Require().Equal(getCertificate.GetSignature(), cert.GetSignature())
	suite.Require().Equal(getCertificate.UnsignedCert.GetDealId(), cert.UnsignedCert.GetDealId())
	suite.Require().Equal(getCertificate.UnsignedCert.GetDataHash(), cert.UnsignedCert.GetDataHash())
	suite.Require().Equal(getCertificate.UnsignedCert.GetEncryptedDataUrl(), cert.UnsignedCert.GetEncryptedDataUrl())
	suite.Require().Equal(getCertificate.UnsignedCert.GetDataValidatorAddress(), cert.UnsignedCert.GetDataValidatorAddress())
	suite.Require().Equal(getCertificate.UnsignedCert.GetRequesterAddress(), cert.UnsignedCert.GetRequesterAddress())
}

func (suite *dealTestSuite) TestListDataCertificates() {
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

	newDealId, err := suite.DataDealKeeper.CreateNewDeal(suite.Ctx, acc1, tempDeal)
	suite.Require().NoError(err)

	cert := makeTestCert("1a312c1223x2fs3", newAddr, acc3)
	_, err = suite.DataDealKeeper.GetDeal(suite.Ctx, newDealId)
	suite.Require().NoError(err)

	dataCertificates := make([]types.DataValidationCertificate, 0)

	for i := 0; i < 5; i++ {
		dataCertificates = append(dataCertificates, cert)
	}

	listDataCertificates, err := suite.DataDealKeeper.ListDataCertificates(suite.Ctx)
	suite.Require().NoError(err)

	for i, dataCertificate := range listDataCertificates {
		suite.Require().Equal(dataCertificate.GetSignature(), dataCertificates[i+1].GetSignature())
		suite.Require().Equal(dataCertificate.UnsignedCert.GetDealId(), dataCertificates[i+1].UnsignedCert.GetDealId())
		suite.Require().Equal(dataCertificate.UnsignedCert.GetDataHash(), dataCertificates[i+1].UnsignedCert.GetDataHash())
		suite.Require().Equal(dataCertificate.UnsignedCert.GetEncryptedDataUrl(), dataCertificates[i+1].UnsignedCert.GetEncryptedDataUrl())
		suite.Require().Equal(dataCertificate.UnsignedCert.GetDataValidatorAddress(), dataCertificates[i+1].UnsignedCert.GetDataValidatorAddress())
		suite.Require().Equal(dataCertificate.UnsignedCert.GetRequesterAddress(), dataCertificates[i+1].UnsignedCert.GetRequesterAddress())
	}
}

func (suite *dealTestSuite) TestDeactivateDeal() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, acc1, sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000))))
	suite.Require().NoError(err)

	err = suite.BankKeeper.AddCoins(suite.Ctx, acc3, zeroFunds)
	suite.Require().NoError(err)

	tempDeal := types.Deal{
		DataSchema:            []string{"http://jsonld.com"},
		Budget:                &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(1000000000)},
		MaxNumData:            10000,
		TrustedDataValidators: []string{newAddr.String()},
		Owner:                 acc1.String(),
	}

	dealId, err := suite.DataDealKeeper.CreateNewDeal(suite.Ctx, acc1, tempDeal)
	suite.Require().NoError(err)

	testCert := makeTestCert("1a312c1223x2fs3", newAddr, acc3)
	_, err = suite.DataDealKeeper.SellOwnData(suite.Ctx, acc3, testCert)
	suite.Require().NoError(err)

	findDeal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, dealId)
	suite.Require().NoError(err)

	dealAddress, err := types.AccDealAddressFromBech32(findDeal.GetDealAddress())
	suite.Require().NoError(err)

	dealBalance := suite.BankKeeper.GetBalance(suite.Ctx, dealAddress, assets.MicroMedDenom)

	beforeDeactivateBalance := suite.BankKeeper.GetBalance(suite.Ctx, acc1, assets.MicroMedDenom)

	deactivatedDealId, err := suite.DataDealKeeper.DeactivateDeal(suite.Ctx, dealId, acc1)
	suite.Require().NoError(err)

	newAcc1Balance := suite.BankKeeper.GetBalance(suite.Ctx, acc1, assets.MicroMedDenom)

	suite.Require().Equal(deactivatedDealId, dealId)

	suite.Require().Equal(newAcc1Balance, beforeDeactivateBalance.Add(dealBalance))
}

func (suite *dealTestSuite) TestIsNotEqualOwner() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, acc1, sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000))))
	suite.Require().NoError(err)

	err = suite.BankKeeper.AddCoins(suite.Ctx, acc3, zeroFunds)
	suite.Require().NoError(err)

	tempDeal := types.Deal{
		DataSchema:            []string{"http://jsonld.com"},
		Budget:                &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(1000000000)},
		MaxNumData:            10000,
		TrustedDataValidators: []string{newAddr.String()},
		Owner:                 acc1.String(),
	}

	dealId, err := suite.DataDealKeeper.CreateNewDeal(suite.Ctx, acc1, tempDeal)
	suite.Require().NoError(err)

	testCert := makeTestCert("1a312c1223x2fs3", newAddr, acc3)
	_, err = suite.DataDealKeeper.SellOwnData(suite.Ctx, acc3, testCert)
	suite.Require().NoError(err)

	_, err = suite.DataDealKeeper.DeactivateDeal(suite.Ctx, dealId, acc2)
	suite.Require().Error(err, "the owner of deal and requester is not equal")
}

func (suite *dealTestSuite) TestDealIsNotActive() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, acc1, sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000))))
	suite.Require().NoError(err)

	err = suite.BankKeeper.AddCoins(suite.Ctx, acc3, zeroFunds)
	suite.Require().NoError(err)

	tempDeal := types.Deal{
		DataSchema:            []string{"http://jsonld.com"},
		Budget:                &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(1000000000)},
		MaxNumData:            10,
		TrustedDataValidators: []string{newAddr.String()},
		Owner:                 acc1.String(),
	}

	dealId, err := suite.DataDealKeeper.CreateNewDeal(suite.Ctx, acc1, tempDeal)
	suite.Require().NoError(err)

	findDeal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, dealId)
	suite.Require().NoError(err)

	findDeal.Status = INACTIVE
	suite.DataDealKeeper.SetDeal(suite.Ctx, findDeal)

	_, err = suite.DataDealKeeper.DeactivateDeal(suite.Ctx, dealId, acc1)
	suite.Require().Error(err, types.ErrInvalidStatus)
	suite.Require().Error(err, "the deal's status is not activated")

	findDeal.Status = ACTIVE
	suite.DataDealKeeper.SetDeal(suite.Ctx, findDeal)

	dataHash := "123456"
	for i := 0; i < 10; i++ {
		cert := makeTestCert(dataHash+strconv.Itoa(i), newAddr, acc1)
		_, err := suite.DataDealKeeper.SellOwnData(suite.Ctx, acc1, cert)
		suite.Require().NoError(err)
	}

	completedDeal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, dealId)
	suite.Require().NoError(err)

	_, err = suite.DataDealKeeper.DeactivateDeal(suite.Ctx, dealId, acc1)
	suite.Require().Error(err, types.ErrInvalidStatus)
	suite.Require().Equal(completedDeal.GetStatus(), COMPLETED)
	suite.Require().Error(err, "the deal's status is not activated")
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

func makeTestCert(dataHash string, validatorAddress sdk.AccAddress, requesterAddress sdk.AccAddress) types.DataValidationCertificate {
	uCert := types.UnsignedDataValidationCertificate{
		DealId:               2,
		DataHash:             []byte(dataHash),
		EncryptedDataUrl:     []byte("https://panacea.org/a/123.json"),
		DataValidatorAddress: validatorAddress.String(),
		RequesterAddress:     requesterAddress.String(),
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
