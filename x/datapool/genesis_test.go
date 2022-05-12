package datapool_test

import (
	"fmt"
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
	redeemer       = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
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

	var dataPassRedeemReceipts []types.DataPassRedeemReceipt

	dataPassRedeemReceipt := makeSampleDataPassRedeemReceipt()

	dataPassRedeemReceipts = append(dataPassRedeemReceipts, dataPassRedeemReceipt)

	instantRevenueDistribute := types.InstantRevenueDistribute{
		PoolIds: poolIDs,
	}

	salesHistoryMap := makeSampleSalesHistories()

	genState := &types.GenesisState{
		DataValidators:           dataValidators,
		NextPoolNumber:           2,
		Pools:                    pools,
		Params:                   params,
		DataPassRedeemReceipts:   dataPassRedeemReceipts,
		InstantRevenueDistribute: instantRevenueDistribute,
		SalesHistories:           salesHistoryMap,
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
	instantRevenueDistributeFromKeeper := suite.DataPoolKeeper.GetInstantRevenueDistribute(suite.Ctx)
	suite.Require().Equal(poolIDs, instantRevenueDistributeFromKeeper.PoolIds)

	salesHistoryFromKeeper := suite.DataPoolKeeper.GetSalesHistories(suite.Ctx, poolID, round)
	suite.Require().Equal(2, len(salesHistoryFromKeeper))

	for _, history := range salesHistoryFromKeeper {
		if history.SellerAddress == seller.String() {
			suite.Require().Equal(poolID, history.PoolId)
			suite.Require().Equal(round, history.Round)
			suite.Require().Equal(seller.String(), history.SellerAddress)
			suite.Require().Equal([]byte("data"), history.DataHashes[0])
			suite.Require().Equal("1000000umed", history.PaidCoin.String())
		} else if history.SellerAddress == seller2.String() {
			suite.Require().Equal(poolID, history.PoolId)
			suite.Require().Equal(round, history.Round)
			suite.Require().Equal(seller2.String(), history.SellerAddress)
			suite.Require().Equal([]byte("data2"), history.DataHashes[0])
			suite.Require().Equal("1000000umed", history.PaidCoin.String())
		}
	}

	secondSalesHistoryFromKeeper := suite.DataPoolKeeper.GetSalesHistories(suite.Ctx, secondPoolID, round)
	for _, history := range secondSalesHistoryFromKeeper {
		if history.SellerAddress == seller.String() {
			suite.Require().Equal(secondPoolID, history.PoolId)
			suite.Require().Equal(round, history.Round)
			suite.Require().Equal(seller.String(), history.SellerAddress)
			suite.Require().Equal([]byte("data3"), history.DataHashes[0])
			suite.Require().Equal("1000000umed", history.PaidCoin.String())
		} else if history.SellerAddress == seller2.String() {
			suite.Require().Equal(secondPoolID, history.PoolId)
			suite.Require().Equal(round, history.Round)
			suite.Require().Equal(seller2.String(), history.SellerAddress)
			suite.Require().Equal([]byte("data4"), history.DataHashes[0])
			suite.Require().Equal("1000000umed", history.PaidCoin.String())
		}
	}
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

	suite.DataPoolKeeper.SetInstantRevenueDistribute(
		suite.Ctx,
		&types.InstantRevenueDistribute{
			PoolIds: poolIDs,
		})

	salesHistories := makeSampleSalesHistories()
	for _, salesHistory := range salesHistories {
		suite.DataPoolKeeper.SetSalesHistory(suite.Ctx, salesHistory)
	}

	genesisState := datapool.ExportGenesis(suite.Ctx, suite.DataPoolKeeper)
	suite.Require().Equal(uint64(2), genesisState.NextPoolNumber)
	suite.Require().Len(genesisState.Pools, 1)
	suite.Require().Equal(types.DefaultParams(), genesisState.Params)
	suite.Require().Len(genesisState.DataValidators, 1)
	suite.Require().Len(genesisState.DataPassRedeemReceipts, 1)
	suite.Require().Equal(poolIDs, genesisState.InstantRevenueDistribute.PoolIds)
	fmt.Println(genesisState.SalesHistories)
	suite.Require().True(len(genesisState.SalesHistories) == 4)
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

func makeSampleSalesHistories() []*types.SalesHistory {
	return []*types.SalesHistory{
		{
			PoolId:        poolID,
			Round:         round,
			SellerAddress: seller.String(),
			DataHashes:    [][]byte{[]byte("data")},
			PaidCoin:      &paidCoin,
		},
		{
			PoolId:        poolID,
			Round:         round,
			SellerAddress: seller2.String(),
			DataHashes:    [][]byte{[]byte("data2")},
			PaidCoin:      &paidCoin,
		},
		{
			PoolId:        secondPoolID,
			Round:         round,
			SellerAddress: seller.String(),
			DataHashes:    [][]byte{[]byte("data3")},
			PaidCoin:      &paidCoin,
		},
		{
			PoolId:        secondPoolID,
			Round:         round,
			SellerAddress: seller2.String(),
			DataHashes:    [][]byte{[]byte("data4")},
			PaidCoin:      &paidCoin,
		},
	}
}

func makeSampleDataPassRedeemReceipt() types.DataPassRedeemReceipt {
	return types.DataPassRedeemReceipt{
		PoolId:   poolID,
		Round:    1,
		NftId:    1,
		Redeemer: redeemer.String(),
	}
}
