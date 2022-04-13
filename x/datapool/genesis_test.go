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
	buyer          = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	buyer2         = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
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

	whiteList := makeSampleWhiteList()

	genState := &types.GenesisState{
		DataValidators: dataValidators,
		NextPoolNumber: 2,
		Pools:          pools,
		Params:         params,
		WhiteList:      whiteList,
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

	// check white list
	whiteListFromKeeper, err := suite.DataPoolKeeper.GetAllWhiteLists(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Len(whiteListFromKeeper, 2)
	suite.Require().Equal(poolID, whiteListFromKeeper[0].GetPoolId())
	suite.Require().Equal(buyer.String(), whiteListFromKeeper[0].GetAddress())
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

	// set white list
	whiteList := makeSampleWhiteList()
	for _, list := range whiteList {
		addr, err := sdk.AccAddressFromBech32(list.Address)
		suite.Require().NoError(err)
		suite.DataPoolKeeper.AddToWhiteList(suite.Ctx, list.PoolId, addr)
	}

	genesisState := datapool.ExportGenesis(suite.Ctx, suite.DataPoolKeeper)
	suite.Require().Equal(uint64(2), genesisState.NextPoolNumber)
	suite.Require().Len(genesisState.Pools, 1)
	suite.Require().Equal(types.DefaultParams(), genesisState.Params)
	suite.Require().Len(genesisState.DataValidators, 1)
	suite.Require().Len(genesisState.WhiteList, 2)
	suite.Require().Contains(genesisState.WhiteList, whiteList[0])
	suite.Require().Contains(genesisState.WhiteList, whiteList[1])
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

func makeSampleWhiteList() []types.WhiteList {
	whiteList1 := types.WhiteList{
		PoolId:  poolID,
		Address: buyer.String(),
	}

	whiteList2 := types.WhiteList{
		PoolId:  poolID,
		Address: buyer2.String(),
	}

	return []types.WhiteList{whiteList1, whiteList2}
}
