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
	seller         = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	seller2        = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	paidCoin       = sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(1000000))
	NFTPrice       = sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000))
	downloadPeriod = time.Second * 100000000
	poolID         = uint64(1)
	secondPoolID   = uint64(2)
	round          = uint64(1)
	poolIDs        = []uint64{uint64(1), uint64(3), uint64(2), uint64(4)}
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

	instantRevenueDistribute := types.InstantRevenueDistribute{
		PoolIds: poolIDs,
	}

	salesHistoryMap := makeSampleSalesHistory()

	genState := &types.GenesisState{
		DataValidators:           dataValidators,
		NextPoolNumber:           2,
		Pools:                    pools,
		Params:                   params,
		InstantRevenueDistribute: instantRevenueDistribute,
		SalesHistory:             salesHistoryMap,
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

	instantRevenueDistributeFromKeeper := suite.DataPoolKeeper.GetInstantRevenueDistribute(suite.Ctx)
	suite.Require().Equal(poolIDs, instantRevenueDistributeFromKeeper.PoolIds)

	salesHistoryFromKeeper := suite.DataPoolKeeper.GetSalesHistory(suite.Ctx, poolID, round)
	suite.Require().Equal(2, len(salesHistoryFromKeeper.SalesInfos))
	suite.Require().Equal(poolID, salesHistoryFromKeeper.SalesInfos[0].PoolId)
	suite.Require().Equal(round, salesHistoryFromKeeper.SalesInfos[0].Round)
	suite.Require().Equal(seller.String(), salesHistoryFromKeeper.SalesInfos[0].Address)
	suite.Require().Equal([]byte("data"), salesHistoryFromKeeper.SalesInfos[0].DataHash)
	suite.Require().Equal("1000000umed", salesHistoryFromKeeper.SalesInfos[0].PaidCoin.String())
	suite.Require().Equal(poolID, salesHistoryFromKeeper.SalesInfos[1].PoolId)
	suite.Require().Equal(round, salesHistoryFromKeeper.SalesInfos[1].Round)
	suite.Require().Equal(seller2.String(), salesHistoryFromKeeper.SalesInfos[1].Address)
	suite.Require().Equal([]byte("data2"), salesHistoryFromKeeper.SalesInfos[1].DataHash)
	suite.Require().Equal("1000000umed", salesHistoryFromKeeper.SalesInfos[1].PaidCoin.String())

	secondSalesHistoryFromKeeper := suite.DataPoolKeeper.GetSalesHistory(suite.Ctx, secondPoolID, round)
	suite.Require().Equal(2, len(secondSalesHistoryFromKeeper.SalesInfos))
	suite.Require().Equal(secondPoolID, secondSalesHistoryFromKeeper.SalesInfos[0].PoolId)
	suite.Require().Equal(round, secondSalesHistoryFromKeeper.SalesInfos[0].Round)
	suite.Require().Equal(seller.String(), secondSalesHistoryFromKeeper.SalesInfos[0].Address)
	suite.Require().Equal([]byte("data3"), secondSalesHistoryFromKeeper.SalesInfos[0].DataHash)
	suite.Require().Equal("1000000umed", secondSalesHistoryFromKeeper.SalesInfos[0].PaidCoin.String())
	suite.Require().Equal(secondPoolID, secondSalesHistoryFromKeeper.SalesInfos[1].PoolId)
	suite.Require().Equal(round, secondSalesHistoryFromKeeper.SalesInfos[1].Round)
	suite.Require().Equal(seller2.String(), secondSalesHistoryFromKeeper.SalesInfos[1].Address)
	suite.Require().Equal([]byte("data4"), secondSalesHistoryFromKeeper.SalesInfos[1].DataHash)
	suite.Require().Equal("1000000umed", secondSalesHistoryFromKeeper.SalesInfos[1].PaidCoin.String())

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

	suite.DataPoolKeeper.SetInstantRevenueDistribute(
		suite.Ctx,
		&types.InstantRevenueDistribute{
			PoolIds: poolIDs,
		})

	salesHistoryMap := makeSampleSalesHistory()
	for key, salesHistory := range salesHistoryMap {
		suite.DataPoolKeeper.SetSalesHistoryByKey(suite.Ctx, []byte(key), &salesHistory)
	}

	genesisState := datapool.ExportGenesis(suite.Ctx, suite.DataPoolKeeper)
	suite.Require().Equal(uint64(2), genesisState.NextPoolNumber)
	suite.Require().Len(genesisState.Pools, 1)
	suite.Require().Equal(types.DefaultParams(), genesisState.Params)
	suite.Require().Len(genesisState.DataValidators, 1)
	suite.Require().Equal(poolIDs, genesisState.InstantRevenueDistribute.PoolIds)
	suite.Require().Equal(salesHistoryMap, genesisState.SalesHistory)
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

func makeSampleSalesHistory() map[string]types.SalesHistory {
	salesHistoryMap := map[string]types.SalesHistory{}

	salesKey1 := string(types.GetKeyPrefixSalesHistory(poolID, round))
	salesHistoryMap[salesKey1] = types.SalesHistory{
		SalesInfos: []*types.SalesInfo{
			{
				PoolId:   poolID,
				Round:    round,
				Address:  seller.String(),
				DataHash: []byte("data"),
				PaidCoin: &paidCoin,
			},
			{
				PoolId:   poolID,
				Round:    round,
				Address:  seller2.String(),
				DataHash: []byte("data2"),
				PaidCoin: &paidCoin,
			},
		},
	}

	salesKey2 := string(types.GetKeyPrefixSalesHistory(secondPoolID, round))
	salesHistoryMap[salesKey2] = types.SalesHistory{
		SalesInfos: []*types.SalesInfo{
			{
				PoolId:   secondPoolID,
				Round:    round,
				Address:  seller.String(),
				DataHash: []byte("data3"),
				PaidCoin: &paidCoin,
			},
			{
				PoolId:   secondPoolID,
				Round:    round,
				Address:  seller2.String(),
				DataHash: []byte("data4"),
				PaidCoin: &paidCoin,
			},
		},
	}
	return salesHistoryMap
}
