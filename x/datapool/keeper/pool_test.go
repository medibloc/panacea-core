package keeper_test

import (
	"io/ioutil"
	"testing"
	"time"

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
	curatorPrivKey = secp256k1.GenPrivKey()
	curatorPubKey  = curatorPrivKey.PubKey()
	curatorAddr    = sdk.AccAddress(curatorPubKey.Address())
	buyerPrivKey   = secp256k1.GenPrivKey()
	buyerPubKey    = buyerPrivKey.PubKey()
	buyerAddr      = sdk.AccAddress(buyerPubKey.Address())

	fundForDataVal = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)))
	fundForCurator = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)))
	fundForBuyer   = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)))
	NFTPrice       = sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000))

	downloadPeriod = time.Duration(time.Second * 100000000)
)

func (suite poolTestSuite) setupNFTContract() {
	// Before test
	wasmCode, err := ioutil.ReadFile("./testdata/cw721_test.wasm")
	suite.Require().NoError(err)

	addr, err := suite.DataPoolKeeper.DeployAndRegisterNFTContract(suite.Ctx, wasmCode)
	suite.Require().NoError(err)

	// set datapool parameters
	params := types.Params{
		DataPoolNftContractAddress: addr.String(),
		DataPoolDeposit:            types.DefaultDataPoolDeposit,
		DataPoolCodeId:             1,
	}

	suite.DataPoolKeeper.SetParams(suite.Ctx, params)
}

func (suite poolTestSuite) setupCreatePool(maxNftSupply uint64) uint64 {
	suite.setupNFTContract()

	err := suite.BankKeeper.AddCoins(suite.Ctx, curatorAddr, fundForCurator)
	suite.Require().NoError(err)

	newPoolParams := makePoolParamsNoDataValidator(maxNftSupply)

	poolID, err := suite.DataPoolKeeper.CreatePool(suite.Ctx, curatorAddr, newPoolParams)
	suite.Require().NoError(err)
	suite.Require().Equal(poolID, uint64(1))

	return poolID
}

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

	err = suite.DataPoolKeeper.UpdateDataValidator(suite.Ctx, dataVal1, updateTempDataValidator.Endpoint)
	suite.Require().NoError(err)

	getDataValidator, err := suite.DataPoolKeeper.GetDataValidator(suite.Ctx, dataVal1)
	suite.Require().NoError(err)

	suite.Require().Equal(getDataValidator.GetAddress(), updateTempDataValidator.GetAddress())
	suite.Require().Equal(getDataValidator.GetEndpoint(), updateTempDataValidator.GetEndpoint())
}

func (suite *poolTestSuite) TestGetPool() {
	poolID := uint64(1)
	nftPrice := sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(1000000))
	downloadPeriod := time.Hour
	poolParams := types.PoolParams{
		DataSchema:            []string{"https://json.schemastore.org/github-issue-forms.json"},
		TargetNumData:         100,
		MaxNftSupply:          10,
		NftPrice:              &nftPrice,
		TrustedDataValidators: []string{dataVal1.String()},
		DownloadPeriod:        &downloadPeriod,
	}

	pool := types.NewPool(poolID, curatorAddr, poolParams)
	suite.DataPoolKeeper.SetPool(suite.Ctx, pool)

	resultPool, err := suite.DataPoolKeeper.GetPool(suite.Ctx, poolID)
	suite.Require().NoError(err)

	suite.Require().Equal(pool.PoolId, resultPool.PoolId)
	suite.Require().Equal(pool.PoolAddress, resultPool.PoolAddress)
	suite.Require().Equal(pool.Round, uint64(1))
	suite.Require().Equal(pool.PoolParams, resultPool.PoolParams)
	suite.Require().Equal(uint64(0), resultPool.CurNumData)
	suite.Require().Equal(pool.NumIssuedNfts, resultPool.NumIssuedNfts)
	suite.Require().Equal(types.PENDING, resultPool.Status)
	suite.Require().Equal(pool.Curator, resultPool.Curator)
}

func (suite poolTestSuite) TestCreatePool() {
	// create and instantiate NFT contract
	suite.setupNFTContract()

	err := suite.BankKeeper.AddCoins(suite.Ctx, curatorAddr, fundForCurator)
	suite.Require().NoError(err)

	// register data validator
	err = suite.BankKeeper.AddCoins(suite.Ctx, dataVal1, fundForDataVal)
	suite.Require().NoError(err)

	validatorAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, dataVal1)
	err = validatorAccount.SetPubKey(pubKey)
	suite.Require().NoError(err)
	suite.AccountKeeper.SetAccount(suite.Ctx, validatorAccount)

	dataValidator := types.DataValidator{
		Address:  dataVal1.String(),
		Endpoint: "https://my-validator.org",
	}

	err = suite.DataPoolKeeper.RegisterDataValidator(suite.Ctx, dataValidator)
	suite.Require().NoError(err)

	newPoolParams := makePoolParamsWithDataValidator()

	poolID, err := suite.DataPoolKeeper.CreatePool(suite.Ctx, curatorAddr, newPoolParams)
	suite.Require().NoError(err)
	suite.Require().Equal(poolID, uint64(1))

	pool, err := suite.DataPoolKeeper.GetPool(suite.Ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(pool.GetPoolId(), uint64(1))
	suite.Require().Equal(pool.GetPoolAddress(), types.NewPoolAddress(poolID).String())
	suite.Require().Equal(pool.GetPoolParams(), &newPoolParams)
	suite.Require().Equal(pool.GetCurator(), curatorAddr.String())
	suite.Require().Equal(pool.GetCurNumData(), uint64(0))
	suite.Require().Equal(pool.GetNumIssuedNfts(), uint64(0))
	suite.Require().Equal(pool.GetRound(), uint64(1))
	suite.Require().Equal(pool.GetStatus(), types.PENDING)
}

func (suite poolTestSuite) TestNotRegisteredDataValidator() {
	// create and instantiate NFT contract
	suite.setupNFTContract()

	err := suite.BankKeeper.AddCoins(suite.Ctx, curatorAddr, fundForCurator)
	suite.Require().NoError(err)

	newPoolParams := makePoolParamsWithDataValidator()

	_, err = suite.DataPoolKeeper.CreatePool(suite.Ctx, curatorAddr, newPoolParams)
	suite.Require().Error(err, types.ErrNotRegisteredDataValidator)
}

func (suite poolTestSuite) TestNotEnoughBalanceForDeposit() {
	// create and instantiate NFT contract
	suite.setupNFTContract()

	newPoolParams := makePoolParamsNoDataValidator(10)

	_, err := suite.DataPoolKeeper.CreatePool(suite.Ctx, curatorAddr, newPoolParams)
	suite.Require().Error(err, types.ErrNotEnoughPoolDeposit)
}

func (suite poolTestSuite) TestNotRegisteredNFTContract() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, curatorAddr, fundForCurator)
	suite.Require().NoError(err)

	newPoolParams := makePoolParamsNoDataValidator(10)

	_, err = suite.DataPoolKeeper.CreatePool(suite.Ctx, curatorAddr, newPoolParams)
	suite.Require().Error(err, types.ErrNoRegisteredNFTContract)
}

func (suite poolTestSuite) TestBuyDataAccessNFTPending() {
	// create pool
	poolID := suite.setupCreatePool(10)

	err := suite.BankKeeper.AddCoins(suite.Ctx, buyerAddr, fundForBuyer)
	suite.Require().NoError(err)

	err = suite.DataPoolKeeper.BuyDataAccessNFT(suite.Ctx, buyerAddr, poolID, 1, NFTPrice)
	suite.Require().NoError(err)

	whiteList, err := suite.DataPoolKeeper.GetWhiteList(suite.Ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Len(whiteList, 1)

	pool, err := suite.DataPoolKeeper.GetPool(suite.Ctx, poolID)
	suite.Require().NoError(err)

	suite.Require().Equal(pool.GetNumIssuedNfts(), uint64(1))
}

// TODO: TestBuyDataAccessNFTActive - check if data access NFT is mintes successfully

func (suite poolTestSuite) TestBuyDataAccessNFTPoolNotFound() {
	// create pool
	suite.setupCreatePool(10)

	err := suite.BankKeeper.AddCoins(suite.Ctx, buyerAddr, fundForBuyer)
	suite.Require().NoError(err)

	// buy NFT other data pool
	err = suite.DataPoolKeeper.BuyDataAccessNFT(suite.Ctx, buyerAddr, 2, 1, NFTPrice)
	suite.Require().Error(err, types.ErrPoolNotFound)
}

func (suite poolTestSuite) TestBuyDataAccessNFTSoldOut() {
	// create pool w/ NFT max supply of 1
	poolID := suite.setupCreatePool(1)

	err := suite.BankKeeper.AddCoins(suite.Ctx, buyerAddr, fundForBuyer)
	suite.Require().NoError(err)

	// buy 1 NFT
	err = suite.DataPoolKeeper.BuyDataAccessNFT(suite.Ctx, buyerAddr, poolID, 1, NFTPrice)
	suite.Require().NoError(err)

	// buy 1 NFT more
	err = suite.DataPoolKeeper.BuyDataAccessNFT(suite.Ctx, buyerAddr, poolID, 1, NFTPrice)
	suite.Require().Error(err, types.ErrNFTAllIssued)
}

func (suite poolTestSuite) TestBuyDataAccessNFTRoundNotMatched() {
	// create pool
	poolID := suite.setupCreatePool(10)

	err := suite.BankKeeper.AddCoins(suite.Ctx, buyerAddr, fundForBuyer)
	suite.Require().NoError(err)

	// different round
	err = suite.DataPoolKeeper.BuyDataAccessNFT(suite.Ctx, buyerAddr, poolID, 2, NFTPrice)
	suite.Require().Error(err, types.ErrRoundNotMatched)
}

func (suite poolTestSuite) TestBuyDataAccessNFTPaymentNotMatched() {
	// create pool
	poolID := suite.setupCreatePool(10)

	err := suite.BankKeeper.AddCoins(suite.Ctx, buyerAddr, fundForBuyer)
	suite.Require().NoError(err)

	// buy NFT with different payment
	err = suite.DataPoolKeeper.BuyDataAccessNFT(suite.Ctx, buyerAddr, poolID, 1, sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(5000000)))
	suite.Require().Error(err, types.ErrPaymentNotMatched)
}

func (suite poolTestSuite) TestBuyDataAccessNFTInsufficientBalance() {
	// create pool
	poolID := suite.setupCreatePool(10)

	// buyer with small balance
	err := suite.BankKeeper.AddCoins(suite.Ctx, buyerAddr, sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(1000))))
	suite.Require().NoError(err)

	err = suite.DataPoolKeeper.BuyDataAccessNFT(suite.Ctx, buyerAddr, poolID, 1, NFTPrice)
	suite.Require().Error(err, sdkerrors.ErrInsufficientFunds)
}

func makePoolParamsWithDataValidator() types.PoolParams {
	return types.PoolParams{
		DataSchema:            []string{"https://www.json.ld"},
		TargetNumData:         100,
		MaxNftSupply:          10,
		NftPrice:              &NFTPrice,
		TrustedDataValidators: []string{dataVal1.String()},
		TrustedDataIssuers:    []string(nil),
		DownloadPeriod:        &downloadPeriod,
	}
}

func makePoolParamsNoDataValidator(maxNftSupply uint64) types.PoolParams {
	return types.PoolParams{
		DataSchema:            []string{"https://www.json.ld"},
		TargetNumData:         100,
		MaxNftSupply:          maxNftSupply,
		NftPrice:              &NFTPrice,
		TrustedDataValidators: []string(nil),
		TrustedDataIssuers:    []string(nil),
		DownloadPeriod:        &downloadPeriod,
	}
}
