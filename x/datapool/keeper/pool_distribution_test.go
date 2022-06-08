package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
)

const (
	defaultSellerCount = defaultTargetNumData
)

var (
	curatorCommissionPerSalesDataPass = sdk.NewInt(100000)
)

func setupRevenueDistributionTest(suite poolTestSuite, targetNumData, poolMaxNftSupply uint64, sellerCount int) (uint64, []sdk.AccAddress) {
	poolID := suite.setupCreatePool(targetNumData, poolMaxNftSupply)

	// create sellers
	sellers := make([]sdk.AccAddress, 0)
	for i := 0; i < sellerCount; i++ {
		sellerAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
		sellers = append(sellers, sellerAddr)
	}

	return poolID, sellers
}

func buyDataPass(suite poolTestSuite, poolID, count uint64) {
	// Buyer buys DataPass.
	err := suite.BankKeeper.AddCoins(suite.Ctx, buyerAddr, fundForBuyer)
	suite.Require().NoError(err)

	for i := uint64(0); i < count; i++ {
		err = suite.DataPoolKeeper.BuyDataPass(suite.Ctx, buyerAddr, poolID, 1, NFTPrice)
		suite.Require().NoError(err)
	}
}

func (suite poolTestSuite) TestExecuteRevenueDistributionPoolActive() {
	// create and instantiate NFT contract
	poolID, sellers := setupRevenueDistributionTest(suite, defaultTargetNumData, defaultMaxNftSupply, defaultSellerCount)
	// create a pool where data sales are not complete.
	suite.Require().Equal(poolID, uint64(1))

	// sell all seller data to the second pool
	for i, sellerAddr := range sellers {
		cert, err := makeTestDataCert(
			suite.Cdc.Marshaler,
			poolID,
			1,
			[]byte(fmt.Sprintf("dataHash_%v", i)),
			sellerAddr.String(),
		)
		suite.Require().NoError(err)
		err = suite.DataPoolKeeper.SellData(suite.Ctx, sellerAddr, *cert)
		suite.Require().NoError(err)
	}

	// check pool status
	pool, err := suite.DataPoolKeeper.GetPool(suite.Ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(types.ACTIVE, pool.Status)

	// check balances of curator and sellers
	// minus two times deposit
	expectedCuratorAmount := fundForCurator.AmountOf(assets.MicroMedDenom).Sub(enoughDeposit.Amount)
	curatorAmount := suite.BankKeeper.GetBalance(suite.Ctx, curatorAddr, assets.MicroMedDenom).Amount
	suite.Require().Equal(expectedCuratorAmount, curatorAmount)
	// amount of all sellers is zero
	for _, sellerAddr := range sellers {
		coin := suite.BankKeeper.GetBalance(suite.Ctx, sellerAddr, assets.MicroMedDenom)
		suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(0)), coin)
	}

	// Buyer buys DataPass.
	buyDataPass(suite, poolID, 1)

	// execute a distribution sales revenue
	err = suite.DataPoolKeeper.DistributionRevenuePools(suite.Ctx)
	suite.Require().NoError(err)

	// check balances of curator and sellers after distribution
	curatorAmount = suite.BankKeeper.GetBalance(suite.Ctx, curatorAddr, assets.MicroMedDenom).Amount
	suite.Require().Equal(fundForCurator.AmountOf(assets.MicroMedDenom).Add(curatorCommissionPerSalesDataPass), curatorAmount)

	for _, sellerAddr := range sellers {
		coin := suite.BankKeeper.GetBalance(suite.Ctx, sellerAddr, assets.MicroMedDenom)
		suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(99000)), coin)
	}

	// 2nd purchase
	err = suite.DataPoolKeeper.BuyDataPass(suite.Ctx, buyerAddr, poolID, 1, NFTPrice)
	suite.Require().NoError(err)

	// execute a distribution sales revenue
	err = suite.DataPoolKeeper.DistributionRevenuePools(suite.Ctx)
	suite.Require().NoError(err)

	curatorAmount = suite.BankKeeper.GetBalance(suite.Ctx, curatorAddr, assets.MicroMedDenom).Amount
	suite.Require().Equal(fundForCurator.AmountOf(assets.MicroMedDenom).Add(curatorCommissionPerSalesDataPass.Mul(sdk.NewInt(2))), curatorAmount)

	for _, sellerAddr := range sellers {
		coin := suite.BankKeeper.GetBalance(suite.Ctx, sellerAddr, assets.MicroMedDenom)
		suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(198000)), coin)
	}
}

func (suite poolTestSuite) TestExecuteRevenueDistributionDataPassSoldOut() {
	// create and instantiate NFT contract
	// create and instantiate NFT contract
	poolID, sellers := setupRevenueDistributionTest(suite, defaultTargetNumData, defaultMaxNftSupply, defaultSellerCount)

	// create a pool where data sales are not complete.
	suite.Require().Equal(poolID, uint64(1))

	// sell all seller data to the second pool
	for i, sellerAddr := range sellers {
		cert, err := makeTestDataCert(
			suite.Cdc.Marshaler,
			poolID,
			1,
			[]byte(fmt.Sprintf("dataHash_%v", i)),
			sellerAddr.String(),
		)
		suite.Require().NoError(err)
		err = suite.DataPoolKeeper.SellData(suite.Ctx, sellerAddr, *cert)
		suite.Require().NoError(err)
	}

	// check pool status
	pool, err := suite.DataPoolKeeper.GetPool(suite.Ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(types.ACTIVE, pool.Status)

	// check balances of curator and sellers
	// minus two times deposit
	expectedCuratorAmount := fundForCurator.AmountOf(assets.MicroMedDenom).Sub(enoughDeposit.Amount)
	curatorAmount := suite.BankKeeper.GetBalance(suite.Ctx, curatorAddr, assets.MicroMedDenom).Amount
	suite.Require().Equal(expectedCuratorAmount, curatorAmount)
	// amount of all sellers is zero
	for _, sellerAddr := range sellers {
		coin := suite.BankKeeper.GetBalance(suite.Ctx, sellerAddr, assets.MicroMedDenom)
		suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(0)), coin)
	}

	// Buyer buys all DataPass.
	buyDataPass(suite, poolID, defaultMaxNftSupply)

	// execute a distribution sales revenue
	err = suite.DataPoolKeeper.DistributionRevenuePools(suite.Ctx)
	suite.Require().NoError(err)

	// check balances of curator and sellers after distribution
	curatorAmount = suite.BankKeeper.GetBalance(suite.Ctx, curatorAddr, assets.MicroMedDenom).Amount
	// Add as much as Curator's commission
	totalCuratorCommissionAmount := curatorCommissionPerSalesDataPass.Mul(sdk.NewInt(defaultMaxNftSupply))
	suite.Require().Equal(fundForCurator.AmountOf(assets.MicroMedDenom).Add(totalCuratorCommissionAmount), curatorAmount)

	for _, sellerAddr := range sellers {
		coin := suite.BankKeeper.GetBalance(suite.Ctx, sellerAddr, assets.MicroMedDenom)
		suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(990000)), coin)
	}
}

func (suite poolTestSuite) TestExecuteRevenueDistributionPoolPending() {
	// create and instantiate NFT contract
	poolID, sellers := setupRevenueDistributionTest(suite, defaultTargetNumData, defaultMaxNftSupply, 50)

	// create a pool where data sales are not complete.
	suite.Require().Equal(poolID, uint64(1))

	// sell all seller data to the second pool
	for i, sellerAddr := range sellers {
		cert, err := makeTestDataCert(
			suite.Cdc.Marshaler,
			poolID,
			1,
			[]byte(fmt.Sprintf("dataHash_%v", i)),
			sellerAddr.String(),
		)
		suite.Require().NoError(err)
		err = suite.DataPoolKeeper.SellData(suite.Ctx, sellerAddr, *cert)
		suite.Require().NoError(err)
	}

	// check pool status
	pool, err := suite.DataPoolKeeper.GetPool(suite.Ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(types.PENDING, pool.Status)

	// check balances of curator and sellers
	// minus two times deposit
	expectedCuratorAmount := fundForCurator.AmountOf(assets.MicroMedDenom).Sub(enoughDeposit.Amount)
	curatorAmount := suite.BankKeeper.GetBalance(suite.Ctx, curatorAddr, assets.MicroMedDenom).Amount
	suite.Require().Equal(expectedCuratorAmount, curatorAmount)
	// amount of all sellers is zero
	for _, sellerAddr := range sellers {
		coin := suite.BankKeeper.GetBalance(suite.Ctx, sellerAddr, assets.MicroMedDenom)
		suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(0)), coin)
	}

	// Buyer buys DataPass.
	buyDataPass(suite, poolID, 1)

	// execute a distribution sales revenue
	err = suite.DataPoolKeeper.DistributionRevenuePools(suite.Ctx)
	suite.Require().NoError(err)

	// check balances of curator and sellers after distribution
	curatorAmount = suite.BankKeeper.GetBalance(suite.Ctx, curatorAddr, assets.MicroMedDenom).Amount
	// (fundForCurator - deposit) + commissionPerSalesDataPass
	expectedCuratorAmount = fundForCurator.AmountOf(assets.MicroMedDenom).Sub(enoughDeposit.Amount).Add(curatorCommissionPerSalesDataPass)
	suite.Require().Equal(expectedCuratorAmount, curatorAmount)

	for _, sellerAddr := range sellers {
		coin := suite.BankKeeper.GetBalance(suite.Ctx, sellerAddr, assets.MicroMedDenom)
		suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(99000)), coin)
	}

	// 2nd purchase
	err = suite.DataPoolKeeper.BuyDataPass(suite.Ctx, buyerAddr, poolID, 1, NFTPrice)
	suite.Require().NoError(err)

	// execute a distribution sales revenue
	err = suite.DataPoolKeeper.DistributionRevenuePools(suite.Ctx)
	suite.Require().NoError(err)

	// check balances of curator and sellers after distribution
	curatorAmount = suite.BankKeeper.GetBalance(suite.Ctx, curatorAddr, assets.MicroMedDenom).Amount
	// (fundForCurator - deposit) + (commissionPerSalesDataPass * 2)
	expectedCuratorAmount = expectedCuratorAmount.Add(curatorCommissionPerSalesDataPass)
	suite.Require().Equal(expectedCuratorAmount, curatorAmount)

	for _, sellerAddr := range sellers {
		coin := suite.BankKeeper.GetBalance(suite.Ctx, sellerAddr, assets.MicroMedDenom)
		suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(198000)), coin)
	}
}

func (suite poolTestSuite) TestExecuteRevenueDistributionPoolPendingDataPassSoldOut() {
	// create and instantiate NFT contract
	poolID, sellers := setupRevenueDistributionTest(suite, defaultTargetNumData, defaultMaxNftSupply, 5)

	// create a pool where data sales are not complete.
	suite.Require().Equal(poolID, uint64(1))

	// sell all seller data to the second pool
	for i, sellerAddr := range sellers {
		cert, err := makeTestDataCert(
			suite.Cdc.Marshaler,
			poolID,
			1,
			[]byte(fmt.Sprintf("dataHash_%v", i)),
			sellerAddr.String(),
		)
		suite.Require().NoError(err)
		err = suite.DataPoolKeeper.SellData(suite.Ctx, sellerAddr, *cert)
		suite.Require().NoError(err)
	}

	// check pool status
	pool, err := suite.DataPoolKeeper.GetPool(suite.Ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(types.PENDING, pool.Status)

	// check balances of curator and sellers
	expectedCuratorAmount := fundForCurator.AmountOf(assets.MicroMedDenom).Sub(enoughDeposit.Amount)
	curatorAmount := suite.BankKeeper.GetBalance(suite.Ctx, curatorAddr, assets.MicroMedDenom).Amount
	suite.Require().Equal(expectedCuratorAmount, curatorAmount)
	// amount of all sellers is zero
	for _, sellerAddr := range sellers {
		coin := suite.BankKeeper.GetBalance(suite.Ctx, sellerAddr, assets.MicroMedDenom)
		suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(0)), coin)
	}

	// Buyer buys all DataPass.
	buyDataPass(suite, poolID, defaultMaxNftSupply)

	// execute a distribution sales revenue
	err = suite.DataPoolKeeper.DistributionRevenuePools(suite.Ctx)
	suite.Require().NoError(err)

	// check balances of curator and sellers after distribution
	curatorAmount = suite.BankKeeper.GetBalance(suite.Ctx, curatorAddr, assets.MicroMedDenom).Amount
	// (fundForCurator - deposit) + (commissionPerSalesDataPass * defaultMaxNftSupply)
	expectedCuratorAmount = expectedCuratorAmount.Add(curatorCommissionPerSalesDataPass.Mul(sdk.NewInt(defaultMaxNftSupply)))
	suite.Require().Equal(expectedCuratorAmount, curatorAmount)
	pool, err = suite.DataPoolKeeper.GetPool(suite.Ctx, poolID)
	suite.Require().NoError(err)

	for _, sellerAddr := range sellers {
		coin := suite.BankKeeper.GetBalance(suite.Ctx, sellerAddr, assets.MicroMedDenom)
		suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(990000)), coin)
	}

	poolAddr, err := sdk.AccAddressFromBech32(pool.PoolAddress)
	suite.Require().NoError(err)
	poolBalance := suite.BankKeeper.GetBalance(suite.Ctx, poolAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(114050000)), poolBalance)
}

func (suite poolTestSuite) TestExecuteRevenueDistributionDuplicateSeller() {
	// create and instantiate NFT contract
	poolID, sellers := setupRevenueDistributionTest(suite, defaultTargetNumData, defaultMaxNftSupply, 50)

	// create a pool where data sales are not complete.
	suite.Require().Equal(poolID, uint64(1))

	// sell all seller data to the second pool
	for i, sellerAddr := range sellers[:5] {
		cert, err := makeTestDataCert(
			suite.Cdc.Marshaler,
			poolID,
			1,
			[]byte(fmt.Sprintf("dataHash_%v", i)),
			sellerAddr.String(),
		)
		suite.Require().NoError(err)
		err = suite.DataPoolKeeper.SellData(suite.Ctx, sellerAddr, *cert)
		suite.Require().NoError(err)
	}

	// sell all seller data to the second pool
	for i, sellerAddr := range sellers {
		cert, err := makeTestDataCert(
			suite.Cdc.Marshaler,
			poolID,
			1,
			[]byte(fmt.Sprintf("dataHash2_%v", i)),
			sellerAddr.String(),
		)
		suite.Require().NoError(err)
		err = suite.DataPoolKeeper.SellData(suite.Ctx, sellerAddr, *cert)
		suite.Require().NoError(err)
	}

	// check pool status
	pool, err := suite.DataPoolKeeper.GetPool(suite.Ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(types.PENDING, pool.Status)

	// check balances of curator and sellers
	// minus two times deposit
	expectedCuratorAmount := fundForCurator.AmountOf(assets.MicroMedDenom).Sub(enoughDeposit.Amount)
	curatorAmount := suite.BankKeeper.GetBalance(suite.Ctx, curatorAddr, assets.MicroMedDenom).Amount
	suite.Require().Equal(expectedCuratorAmount, curatorAmount)
	// amount of all sellers is zero
	for _, sellerAddr := range sellers {
		coin := suite.BankKeeper.GetBalance(suite.Ctx, sellerAddr, assets.MicroMedDenom)
		suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(0)), coin)
	}

	// Buyer buys DataPass.
	buyDataPass(suite, poolID, 1)

	// execute a distribution sales revenue
	err = suite.DataPoolKeeper.DistributionRevenuePools(suite.Ctx)
	suite.Require().NoError(err)

	// check balances of curator and sellers after distribution
	curatorAmount = suite.BankKeeper.GetBalance(suite.Ctx, curatorAddr, assets.MicroMedDenom).Amount
	// (fundForCurator - deposit) + commissionPerSalesDataPass
	expectedCuratorAmount = expectedCuratorAmount.Add(curatorCommissionPerSalesDataPass)
	suite.Require().Equal(expectedCuratorAmount, curatorAmount)

	for i, sellerAddr := range sellers {
		coin := suite.BankKeeper.GetBalance(suite.Ctx, sellerAddr, assets.MicroMedDenom)
		if i < 5 {
			suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(198000)), coin)
		} else {
			suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(99000)), coin)
		}
	}

	// 2nd purchase
	buyDataPass(suite, poolID, 1)

	// execute a distribution sales revenue
	err = suite.DataPoolKeeper.DistributionRevenuePools(suite.Ctx)
	suite.Require().NoError(err)

	// check balances of curator and sellers after distribution
	curatorAmount = suite.BankKeeper.GetBalance(suite.Ctx, curatorAddr, assets.MicroMedDenom).Amount
	// (fundForCurator - deposit) + (commissionPerSalesDataPass * 2)
	expectedCuratorAmount = expectedCuratorAmount.Add(curatorCommissionPerSalesDataPass)
	suite.Require().Equal(expectedCuratorAmount, curatorAmount)

	for i, sellerAddr := range sellers {
		coin := suite.BankKeeper.GetBalance(suite.Ctx, sellerAddr, assets.MicroMedDenom)
		if i < 5 {
			suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(396000)), coin)
		} else {
			suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(198000)), coin)
		}
	}
}

func (suite poolTestSuite) TestExecuteRevenueionTarget101() {
	// create and instantiate NFT contract
	poolID, sellers := setupRevenueDistributionTest(suite, defaultTargetNumData+1, defaultMaxNftSupply, defaultSellerCount+1)

	// create a pool where data sales are not complete.
	suite.Require().Equal(poolID, uint64(1))

	// sell all seller data to the second pool
	for i, sellerAddr := range sellers {
		cert, err := makeTestDataCert(
			suite.Cdc.Marshaler,
			poolID,
			1,
			[]byte(fmt.Sprintf("dataHash_%v", i)),
			sellerAddr.String(),
		)
		suite.Require().NoError(err)
		err = suite.DataPoolKeeper.SellData(suite.Ctx, sellerAddr, *cert)
		suite.Require().NoError(err)
	}

	// check pool status
	pool, err := suite.DataPoolKeeper.GetPool(suite.Ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(types.ACTIVE, pool.Status)

	// check balances of curator and sellers
	// minus two times deposit
	expectedCuratorAmount := fundForCurator.AmountOf(assets.MicroMedDenom).Sub(enoughDeposit.Amount)
	curatorAmount := suite.BankKeeper.GetBalance(suite.Ctx, curatorAddr, assets.MicroMedDenom).Amount
	suite.Require().Equal(expectedCuratorAmount, curatorAmount)
	// amount of all sellers is zero
	for _, sellerAddr := range sellers {
		coin := suite.BankKeeper.GetBalance(suite.Ctx, sellerAddr, assets.MicroMedDenom)
		suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(0)), coin)
	}

	// Buyer buys all DataPass.
	buyDataPass(suite, poolID, defaultMaxNftSupply)

	// execute a distribution sales revenue
	err = suite.DataPoolKeeper.DistributionRevenuePools(suite.Ctx)
	suite.Require().NoError(err)

	// check balances of curator and sellers after distribution
	curatorAmount = suite.BankKeeper.GetBalance(suite.Ctx, curatorAddr, assets.MicroMedDenom).Amount
	expectedCuratorAmount = fundForCurator.AmountOf(assets.MicroMedDenom).Add(curatorCommissionPerSalesDataPass.Mul(sdk.NewInt(defaultMaxNftSupply)))
	suite.Require().Equal(expectedCuratorAmount, curatorAmount)

	for _, sellerAddr := range sellers {
		coin := suite.BankKeeper.GetBalance(suite.Ctx, sellerAddr, assets.MicroMedDenom)
		suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(980198)), coin)
	}

	//
	poolAddr, err := sdk.AccAddressFromBech32(pool.PoolAddress)
	suite.Require().NoError(err)
	poolBalance := suite.BankKeeper.GetBalance(suite.Ctx, poolAddr, assets.MicroMedDenom)
	// After dividing revenue, there is a remainder.
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(2)), poolBalance)
}
