package datapool_test

import (
	"testing"
	"time"

	"github.com/medibloc/panacea-core/v2/x/datapool"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"

	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

var (
	dataVal        = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	curator        = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	redeemer       = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	NFTPrice       = sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000))
	downloadPeriod = time.Duration(time.Second * 100000000)
	poolID         = uint64(1)
)

type genesisTestSuite struct {
	testsuite.TestSuite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}

func (suite genesisTestSuite) TestDataPoolInitGenesis() {
	var dataValidators []types.DataValidator

	dataValidator := makeSampleDataValidator()

	dataValidators = append(dataValidators, dataValidator)

	pool := makeSamplePool()

	pools := []types.Pool{pool}

	params := types.DefaultParams()

	var dataPassRedeemReceipts []types.DataPassRedeemReceipt

	dataPassRedeemReceipt := makeSampleDataPassRedeemReceipt()

	dataPassRedeemReceipts = append(dataPassRedeemReceipts, dataPassRedeemReceipt)

	genState := &types.GenesisState{
		DataValidators:         dataValidators,
		NextPoolNumber:         2,
		Pools:                  pools,
		Params:                 params,
		DataPassRedeemReceipts: dataPassRedeemReceipts,
	}

	datapool.InitGenesis(suite.Ctx, suite.DataPoolKeeper, *genState)

	// check data validator
	dataValidatorFromKeeper, err := suite.DataPoolKeeper.GetDataValidator(suite.Ctx, dataVal)
	suite.Require().NoError(err)
	suite.Require().Equal(dataValidator, dataValidatorFromKeeper)

	// check the next pool number
	suite.Require().Equal(uint64(2), suite.DataPoolKeeper.GetNextPoolNumber(suite.Ctx))

	// check pool
	poolFromKeeper, err := suite.DataPoolKeeper.GetPool(suite.Ctx, uint64(1))
	suite.Require().NoError(err)
	suite.Require().Equal(pool, *poolFromKeeper)

	// check params
	paramsFromKeeper := suite.DataPoolKeeper.GetParams(suite.Ctx)
	suite.Require().Equal(params, paramsFromKeeper)

	// check all data pass redeem receipts
	dataPassRedeemReceiptsFromKeeper, err := suite.DataPoolKeeper.GetAllDataPassRedeemReceipts(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(dataPassRedeemReceipts, dataPassRedeemReceiptsFromKeeper)
}

func (suite genesisTestSuite) TestDataPoolExportGenesis() {
	// register data validator
	dataValidator := makeSampleDataValidator()
	err := suite.DataPoolKeeper.SetDataValidator(suite.Ctx, dataValidator)
	suite.Require().NoError(err)

	// create pool
	pool := makeSamplePool()
	suite.DataPoolKeeper.SetPool(suite.Ctx, &pool)
	suite.DataPoolKeeper.SetPoolNumber(suite.Ctx, uint64(2))

	// set params
	suite.DataPoolKeeper.SetParams(suite.Ctx, types.DefaultParams())

	// set data pass redeem receipt
	dataPassRedeemReceipt := makeSampleDataPassRedeemReceipt()
	err = suite.DataPoolKeeper.SetDataPassRedeemReceipt(suite.Ctx, dataPassRedeemReceipt)
	suite.Require().NoError(err)

	genesisState := datapool.ExportGenesis(suite.Ctx, suite.DataPoolKeeper)
	suite.Require().Equal(uint64(2), genesisState.NextPoolNumber)
	suite.Require().Len(genesisState.Pools, 1)
	suite.Require().Equal(types.DefaultParams(), genesisState.Params)
	suite.Require().Len(genesisState.DataValidators, 1)
	suite.Require().Len(genesisState.DataPassRedeemReceipts, 1)
}

func makeSampleDataValidator() types.DataValidator {
	return types.DataValidator{
		Address:  dataVal.String(),
		Endpoint: "https://my-validator.org",
	}
}

func makeSamplePool() types.Pool {
	return types.Pool{
		PoolId:        poolID,
		PoolAddress:   types.NewPoolAddress(uint64(1)).String(),
		Round:         1,
		PoolParams:    makeSamplePoolParams(),
		CurNumData:    0,
		NumIssuedNfts: 1,
		Status:        types.PENDING,
		Curator:       curator.String(),
		Deposit:       types.ZeroFund,
	}
}

func makeSamplePoolParams() *types.PoolParams {
	return &types.PoolParams{
		DataSchema:            []string{"https://www.json.ld"},
		TargetNumData:         100,
		MaxNftSupply:          10,
		NftPrice:              &NFTPrice,
		TrustedDataValidators: []string{dataVal.String()},
		TrustedDataIssuers:    []string(nil),
		DownloadPeriod:        &downloadPeriod,
	}
}

func makeSampleDataPassRedeemReceipt() types.DataPassRedeemReceipt {
	return types.DataPassRedeemReceipt{
		PoolId:     poolID,
		Round:      1,
		DataPassId: 1,
		Redeemer:   redeemer.String(),
	}
}
