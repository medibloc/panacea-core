package keeper_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

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

const (
	defaultTargetNumData = 100
	defaultMaxNfySupply  = 10
)

var (
	dataValPrivKey = secp256k1.GenPrivKey()
	dataValPubKey  = dataValPrivKey.PubKey()
	dataVal1       = sdk.AccAddress(dataValPubKey.Address())

	curatorPrivKey = secp256k1.GenPrivKey()
	curatorPubKey  = curatorPrivKey.PubKey()
	curatorAddr    = sdk.AccAddress(curatorPubKey.Address())
	buyerPrivKey   = secp256k1.GenPrivKey()
	buyerPubKey    = buyerPrivKey.PubKey()
	buyerAddr      = sdk.AccAddress(buyerPubKey.Address())

	requesterPrivKey = secp256k1.GenPrivKey()
	requesterPubKey  = requesterPrivKey.PubKey()
	requesterAddr    = sdk.AccAddress(requesterPubKey.Address())

	fundForDataVal = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)))
	fundForCurator = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(1000000000000)))
	fundForBuyer   = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)))
	NFTPrice       = sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000)) // 10 MED
	enoughDeposit  = sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(20000000)) // 20 MED

	downloadPeriod = time.Second * 100000000
)

func (suite poolTestSuite) setupNFTContract() {
	// Before test
	wasmCode, err := ioutil.ReadFile("./testdata/cw721_test.wasm")
	suite.Require().NoError(err)

	addr, err := suite.DataPoolKeeper.DeployAndRegisterNFTContract(suite.Ctx, wasmCode)
	suite.Require().NoError(err)

	// set datapool parameters
	params := types.Params{
		DataPoolDepositRate:           types.DefaultDataPoolDepositRate,
		DataPoolNftContractAddress:    addr.String(),
		DataPoolCodeId:                1,
		DataPoolCuratorCommissionRate: types.DefaultDataPoolCuratorCommissionRate,
	}

	suite.DataPoolKeeper.SetParams(suite.Ctx, params)
}

func (suite poolTestSuite) setupCreatePool(targetNumData, maxNftSupply uint64) uint64 {
	suite.setupNFTContract()

	err := suite.BankKeeper.AddCoins(suite.Ctx, curatorAddr, fundForCurator)
	suite.Require().NoError(err)

	// register data validator
	err = suite.BankKeeper.AddCoins(suite.Ctx, dataVal1, fundForDataVal)
	suite.Require().NoError(err)

	suite.setDataValidatorAccount()

	dataValidator := types.DataValidator{
		Address:  dataVal1.String(),
		Endpoint: "https://my-validator.org",
	}

	err = suite.DataPoolKeeper.RegisterDataValidator(suite.Ctx, dataValidator)
	suite.Require().NoError(err)

	newPoolParams := makePoolParamsWithDataValidator(targetNumData, maxNftSupply)

	poolID, err := suite.DataPoolKeeper.CreatePool(suite.Ctx, curatorAddr, enoughDeposit, newPoolParams)
	suite.Require().NoError(err)

	return poolID
}

func (suite *poolTestSuite) TestRegisterDataValidator() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, dataVal1, fundForDataVal)
	suite.Require().NoError(err)

	suite.setDataValidatorAccount()

	tempDataValidator := types.DataValidator{
		Address:  dataVal1.String(),
		Endpoint: "https://my-validator.org",
	}

	err = suite.DataPoolKeeper.RegisterDataValidator(suite.Ctx, tempDataValidator)
	suite.Require().NoError(err)
}

func (suite *poolTestSuite) setDataValidatorAccount() {
	validatorAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, dataVal1)
	err := validatorAccount.SetPubKey(dataValPubKey)
	suite.Require().NoError(err)
	suite.AccountKeeper.SetAccount(suite.Ctx, validatorAccount)
}

func (suite *poolTestSuite) TestGetRegisterDataValidator() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, dataVal1, fundForDataVal)
	suite.Require().NoError(err)

	suite.setDataValidatorAccount()

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

	suite.setDataValidatorAccount()

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

	suite.setDataValidatorAccount()

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
	downloadPeriod := time.Hour
	poolParams := types.PoolParams{
		DataSchema:            []string{"https://json.schemastore.org/github-issue-forms.json"},
		TargetNumData:         defaultTargetNumData,
		MaxNftSupply:          defaultMaxNfySupply,
		NftPrice:              &NFTPrice,
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
	poolID := suite.setupCreatePool(defaultTargetNumData, defaultMaxNfySupply)

	suite.Require().Equal(poolID, uint64(1))

	newPoolParams := makePoolParamsWithDataValidator(defaultTargetNumData, defaultMaxNfySupply)

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

	newPoolParams := makePoolParamsWithDataValidator(defaultTargetNumData, defaultMaxNfySupply)

	_, err = suite.DataPoolKeeper.CreatePool(suite.Ctx, curatorAddr, enoughDeposit, newPoolParams)
	suite.Require().Error(err, types.ErrNotRegisteredDataValidator)
}

func (suite poolTestSuite) TestNotEnoughBalanceForDeposit() {
	// create and instantiate NFT contract
	suite.setupNFTContract()

	newPoolParams := makePoolParamsNoDataValidator(defaultMaxNfySupply)

	_, err := suite.DataPoolKeeper.CreatePool(suite.Ctx, curatorAddr, enoughDeposit, newPoolParams)
	suite.Require().Error(err, types.ErrNotEnoughPoolDeposit)
}

func (suite poolTestSuite) TestNotRegisteredNFTContract() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, curatorAddr, fundForCurator)
	suite.Require().NoError(err)

	newPoolParams := makePoolParamsNoDataValidator(defaultMaxNfySupply)

	_, err = suite.DataPoolKeeper.CreatePool(suite.Ctx, curatorAddr, enoughDeposit, newPoolParams)
	suite.Require().Error(err, types.ErrNoRegisteredNFTContract)
}

func (suite poolTestSuite) TestBuyDataAccessNFTPending() {
	// create pool
	poolID := suite.setupCreatePool(defaultTargetNumData, defaultMaxNfySupply)

	err := suite.BankKeeper.AddCoins(suite.Ctx, buyerAddr, fundForBuyer)
	suite.Require().NoError(err)

	err = suite.DataPoolKeeper.BuyDataPass(suite.Ctx, buyerAddr, poolID, 1, NFTPrice)
	suite.Require().NoError(err)

	pool, err := suite.DataPoolKeeper.GetPool(suite.Ctx, poolID)
	suite.Require().NoError(err)

	suite.Require().Equal(pool.GetNumIssuedNfts(), uint64(1))
}

// TODO: TestBuyDataAccessNFTActive - check if data access NFT is mintes successfully

func (suite poolTestSuite) TestBuyDataAccessNFTPoolNotFound() {
	// create pool
	suite.setupCreatePool(defaultTargetNumData, defaultMaxNfySupply)

	err := suite.BankKeeper.AddCoins(suite.Ctx, buyerAddr, fundForBuyer)
	suite.Require().NoError(err)

	// buy NFT other data pool
	err = suite.DataPoolKeeper.BuyDataPass(suite.Ctx, buyerAddr, 2, 1, NFTPrice)
	suite.Require().Error(err, types.ErrPoolNotFound)
}

func (suite poolTestSuite) TestBuyDataAccessNFTSoldOut() {
	// create pool w/ NFT max supply of 1
	poolID := suite.setupCreatePool(defaultTargetNumData, 1)

	err := suite.BankKeeper.AddCoins(suite.Ctx, buyerAddr, fundForBuyer)
	suite.Require().NoError(err)

	// buy 1 NFT
	err = suite.DataPoolKeeper.BuyDataPass(suite.Ctx, buyerAddr, poolID, 1, NFTPrice)
	suite.Require().NoError(err)

	// buy 1 NFT more
	err = suite.DataPoolKeeper.BuyDataPass(suite.Ctx, buyerAddr, poolID, 1, NFTPrice)
	suite.Require().Error(err, types.ErrNFTAllIssued)
}

func (suite poolTestSuite) TestBuyDataAccessNFTRoundNotMatched() {
	// create pool
	poolID := suite.setupCreatePool(defaultTargetNumData, defaultMaxNfySupply)

	err := suite.BankKeeper.AddCoins(suite.Ctx, buyerAddr, fundForBuyer)
	suite.Require().NoError(err)

	// different round
	err = suite.DataPoolKeeper.BuyDataPass(suite.Ctx, buyerAddr, poolID, 2, NFTPrice)
	suite.Require().Error(err, types.ErrRoundNotMatched)
}

func (suite poolTestSuite) TestBuyDataAccessNFTPaymentNotMatched() {
	// create pool
	poolID := suite.setupCreatePool(defaultTargetNumData, defaultMaxNfySupply)

	err := suite.BankKeeper.AddCoins(suite.Ctx, buyerAddr, fundForBuyer)
	suite.Require().NoError(err)

	// buy NFT with different payment
	err = suite.DataPoolKeeper.BuyDataPass(suite.Ctx, buyerAddr, poolID, 1, sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(5000000)))
	suite.Require().Error(err, types.ErrPaymentNotMatched)
}

func (suite poolTestSuite) TestBuyDataAccessNFTInsufficientBalance() {
	// create pool
	poolID := suite.setupCreatePool(defaultTargetNumData, defaultMaxNfySupply)

	// buyer with small balance
	err := suite.BankKeeper.AddCoins(suite.Ctx, buyerAddr, sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(1000))))
	suite.Require().NoError(err)

	err = suite.DataPoolKeeper.BuyDataPass(suite.Ctx, buyerAddr, poolID, 1, NFTPrice)
	suite.Require().Error(err, sdkerrors.ErrInsufficientFunds)
}

func (suite poolTestSuite) TestNotEnoughDeposit() {
	err := suite.BankKeeper.AddCoins(suite.Ctx, curatorAddr, fundForCurator)
	suite.Require().NoError(err)

	newPoolParams := makePoolParamsNoDataValidator(defaultMaxNfySupply)

	// 10 MED is required to create pool. but 5 MED for deposit
	deposit := sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(500000))

	_, err = suite.DataPoolKeeper.CreatePool(suite.Ctx, curatorAddr, deposit, newPoolParams)
	suite.Require().Error(err, types.ErrNotEnoughPoolDeposit)
}

func makePoolParamsWithDataValidator(TargetNumData, MaxNftSupply uint64) types.PoolParams {
	return types.PoolParams{
		DataSchema:            []string{"https://www.json.ld"},
		TargetNumData:         TargetNumData,
		MaxNftSupply:          MaxNftSupply,
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

func (suite *poolTestSuite) TestSetDataCertificate() {
	poolID := uint64(1)
	round := uint64(1)
	dataHash := []byte("dataHash")

	unsignedCert := types.UnsignedDataValidationCertificate{
		PoolId:        poolID,
		Round:         round,
		DataHash:      dataHash,
		DataValidator: dataVal1.String(),
		Requester:     requesterAddr.String(),
	}

	bz, err := suite.Cdc.Marshaler.MarshalBinaryBare(&unsignedCert)
	suite.Require().NoError(err)

	sign, err := dataValPrivKey.Sign(bz)
	suite.Require().NoError(err)

	cert := types.DataValidationCertificate{
		UnsignedCert: &unsignedCert,
		Signature:    sign,
	}

	suite.DataPoolKeeper.SetDataValidationCertificate(suite.Ctx, cert)

	getCert, err := suite.DataPoolKeeper.GetDataValidationCertificate(suite.Ctx, poolID, round, dataHash)
	suite.Require().NoError(err)

	suite.Require().Equal(cert.UnsignedCert.PoolId, getCert.UnsignedCert.PoolId)
	suite.Require().Equal(cert.UnsignedCert.Round, getCert.UnsignedCert.Round)
	suite.Require().Equal(cert.UnsignedCert.DataHash, getCert.UnsignedCert.DataHash)
	suite.Require().Equal(cert.UnsignedCert.DataValidator, getCert.UnsignedCert.DataValidator)
	suite.Require().Equal(cert.UnsignedCert.PoolId, getCert.UnsignedCert.PoolId)
	suite.Require().Equal(cert.UnsignedCert.Requester, getCert.UnsignedCert.Requester)
	suite.Require().Equal(cert.Signature, getCert.Signature)
}

func (suite *poolTestSuite) TestSellData() {
	poolID := suite.setupCreatePool(defaultTargetNumData, defaultMaxNfySupply)

	pool, err := suite.DataPoolKeeper.GetPool(suite.Ctx, poolID)
	suite.Require().NoError(err)
	round := pool.Round
	dataHash := []byte("dataHash")

	cert, err := makeTestDataCertificate(suite.Cdc.Marshaler, poolID, round, dataHash, requesterAddr.String())
	suite.Require().NoError(err)

	err = suite.DataPoolKeeper.SellData(suite.Ctx, requesterAddr, *cert)
	suite.Require().NoError(err)

	// Check the current pool status
	getPool, err := suite.DataPoolKeeper.GetPool(suite.Ctx, poolID)
	suite.Require().NoError(err)
	suite.Equal(uint64(1), getPool.CurNumData)
	suite.Equal(types.PENDING, getPool.Status)
}

func (suite *poolTestSuite) TestSellData_change_status_activity() {
	poolID := suite.setupCreatePool(defaultTargetNumData, defaultMaxNfySupply)

	pool, err := suite.DataPoolKeeper.GetPool(suite.Ctx, poolID)
	suite.Require().NoError(err)
	round := pool.Round
	dataHash := []byte("dataHash")

	// Modify the target data of the pool
	pool.PoolParams.TargetNumData = 1
	suite.DataPoolKeeper.SetPool(suite.Ctx, pool)

	cert, err := makeTestDataCertificate(suite.Cdc.Marshaler, poolID, round, dataHash, requesterAddr.String())
	suite.Require().NoError(err)

	err = suite.DataPoolKeeper.SellData(suite.Ctx, requesterAddr, *cert)
	suite.Require().NoError(err)

	// Check the current pool status
	getPool, err := suite.DataPoolKeeper.GetPool(suite.Ctx, poolID)
	suite.Require().NoError(err)
	suite.Equal(uint64(1), getPool.CurNumData)
	suite.Equal(types.ACTIVE, getPool.Status)
}

func (suite *poolTestSuite) TestSellData_not_same_seller() {
	poolID := uint64(1)
	round := uint64(1)
	dataHash := []byte("dataHash")

	cert, err := makeTestDataCertificate(suite.Cdc.Marshaler, poolID, round, dataHash, requesterAddr.String())
	suite.Require().NoError(err)

	// A curator requests to sell data
	err = suite.DataPoolKeeper.SellData(suite.Ctx, curatorAddr, *cert)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrNotEqualsSeller.Error(), err.Error())
}

func (suite *poolTestSuite) TestSellData_failed_get_publicKey_validator_in_signature() {
	poolID := uint64(1)
	round := uint64(1)
	dataHash := []byte("dataHash")

	// Unregistered data validator publicKey

	cert, err := makeTestDataCertificate(suite.Cdc.Marshaler, poolID, round, dataHash, requesterAddr.String())
	suite.Require().NoError(err)

	unsignedCertBz, err := suite.Cdc.Marshaler.MarshalBinaryBare(cert.UnsignedCert)
	suite.Require().NoError(err)

	curatorSign, err := curatorPrivKey.Sign(unsignedCertBz)
	suite.Require().NoError(err)
	cert.Signature = curatorSign

	err = suite.DataPoolKeeper.SellData(suite.Ctx, requesterAddr, *cert)
	suite.Require().Error(err)
	suite.Require().True(strings.HasSuffix(err.Error(), types.ErrInvalidSignature.Error()))
}

func (suite *poolTestSuite) TestSellData_invalid_signature() {
	poolID := uint64(1)
	round := uint64(1)
	dataHash := []byte("dataHash")

	suite.setDataValidatorAccount()

	cert, err := makeTestDataCertificate(suite.Cdc.Marshaler, poolID, round, dataHash, requesterAddr.String())
	suite.Require().NoError(err)

	unsignedCertBz, err := suite.Cdc.Marshaler.MarshalBinaryBare(cert.UnsignedCert)
	suite.Require().NoError(err)

	// Curator's signature
	curatorSign, err := curatorPrivKey.Sign(unsignedCertBz)
	suite.Require().NoError(err)
	cert.Signature = curatorSign

	err = suite.DataPoolKeeper.SellData(suite.Ctx, requesterAddr, *cert)
	suite.Require().Error(err)
	fmt.Println(err)
	suite.Require().True(strings.Contains(err.Error(), "invalid signature"))
	suite.Require().True(strings.HasSuffix(err.Error(), types.ErrInvalidSignature.Error()))
}

func (suite *poolTestSuite) TestSellData_duplicate_data() {
	poolID := suite.setupCreatePool(defaultTargetNumData, defaultMaxNfySupply)

	pool, err := suite.DataPoolKeeper.GetPool(suite.Ctx, poolID)
	suite.Require().NoError(err)
	round := pool.Round
	dataHash := []byte("dataHash")

	cert, err := makeTestDataCertificate(suite.Cdc.Marshaler, poolID, round, dataHash, requesterAddr.String())
	suite.Require().NoError(err)

	err = suite.DataPoolKeeper.SellData(suite.Ctx, requesterAddr, *cert)
	suite.Require().NoError(err)

	// Request same sell data
	err = suite.DataPoolKeeper.SellData(suite.Ctx, requesterAddr, *cert)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrExistSameDataHash.Error(), err.Error())
}

func (suite *poolTestSuite) TestSellData_not_exist_pool() {
	poolID := uint64(1)
	round := uint64(1)
	dataHash := []byte("dataHash")

	suite.setDataValidatorAccount()

	// Unregistered pool

	cert, err := makeTestDataCertificate(suite.Cdc.Marshaler, poolID, round, dataHash, requesterAddr.String())
	suite.Require().NoError(err)

	err = suite.DataPoolKeeper.SellData(suite.Ctx, requesterAddr, *cert)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrPoolNotFound, err)
}

func (suite *poolTestSuite) TestSellData_impossible_status_pool() {
	poolID := uint64(1)
	round := uint64(1)
	dataHash := []byte("dataHash")

	suite.setDataValidatorAccount()

	// Already activate status
	pool := makeTestDataPool(poolID)
	pool.Status = types.ACTIVE
	suite.DataPoolKeeper.SetPool(suite.Ctx, pool)

	cert, err := makeTestDataCertificate(suite.Cdc.Marshaler, poolID, round, dataHash, requesterAddr.String())
	suite.Require().NoError(err)

	err = suite.DataPoolKeeper.SellData(suite.Ctx, requesterAddr, *cert)
	suite.Require().Error(err)
	suite.Require().True(strings.Contains(err.Error(), "the status of the pool is not 'PENDING'"))
	suite.Require().True(strings.HasSuffix(err.Error(), types.ErrInvalidDataValidationCert.Error()))
}

func (suite *poolTestSuite) TestSellData_mismatch_certificate_and_pool_round() {
	poolID := uint64(1)
	// Wrong rounds recorded in the certificate
	round := uint64(2)
	dataHash := []byte("dataHash")

	suite.setDataValidatorAccount()

	pool := makeTestDataPool(poolID)
	suite.DataPoolKeeper.SetPool(suite.Ctx, pool)

	cert, err := makeTestDataCertificate(suite.Cdc.Marshaler, poolID, round, dataHash, requesterAddr.String())
	suite.Require().NoError(err)

	err = suite.DataPoolKeeper.SellData(suite.Ctx, requesterAddr, *cert)
	suite.Require().Error(err)
	suite.Require().True(strings.Contains(err.Error(), "pool round do not matched. pool round: 1"))
	suite.Require().True(strings.HasSuffix(err.Error(), types.ErrInvalidDataValidationCert.Error()))
}

func makeTestDataCertificate(marshaler codec.Marshaler, poolID, round uint64, dataHash []byte, requester string) (*types.DataValidationCertificate, error) {
	unsignedCert := types.UnsignedDataValidationCertificate{
		PoolId:        poolID,
		Round:         round,
		DataHash:      dataHash,
		DataValidator: dataVal1.String(),
		Requester:     requester,
	}

	bz, err := marshaler.MarshalBinaryBare(&unsignedCert)
	if err != nil {
		return nil, err
	}

	sign, err := dataValPrivKey.Sign(bz)
	if err != nil {
		return nil, err
	}

	return &types.DataValidationCertificate{
		UnsignedCert: &unsignedCert,
		Signature:    sign,
	}, nil
}

func makeTestDataPool(poolID uint64) *types.Pool {
	downloadPeriod := time.Hour
	poolParams := types.PoolParams{
		DataSchema:            []string{"https://json.schemastore.org/github-issue-forms.json"},
		TargetNumData:         defaultTargetNumData,
		MaxNftSupply:          defaultMaxNfySupply,
		NftPrice:              &NFTPrice,
		TrustedDataValidators: []string{dataVal1.String()},
		DownloadPeriod:        &downloadPeriod,
	}

	return types.NewPool(poolID, curatorAddr, poolParams)
}
