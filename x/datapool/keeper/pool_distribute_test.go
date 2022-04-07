package keeper_test

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
)

func (suite poolTestSuite) TestDistributeSalesRevenueNotSoldNFTs() {
	// create and instantiate NFT contract
	suite.initCreatePool()

	// create a pool where data sales are not complete.
	poolID, err := suite.DataPoolKeeper.CreatePool(suite.Ctx, curatorAddr, makePoolParamsWithDataValidator(100, 10))
	suite.Require().NoError(err)
	suite.Require().Equal(poolID, uint64(1))

	// create a pool where data sales will be completed
	secondPoolID, err := suite.DataPoolKeeper.CreatePool(suite.Ctx, curatorAddr, makePoolParamsWithDataValidator(1000, 100))
	suite.Require().NoError(err)
	suite.Require().Equal(secondPoolID, uint64(2))

	// create 1000 sellers
	sellers := make([]sdk.AccAddress, 0)
	for i := 0; i < 1000; i++ {
		sellerAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
		sellers = append(sellers, sellerAddr)
	}

	// sell half of seller data to the first pool
	for i, sellerAddr := range sellers[:50] {
		cert, err := makeTestDataCertificate(
			suite.Cdc.Marshaler,
			poolID,
			1,
			[]byte(fmt.Sprintf("dataHash_%v", i)),
			sellerAddr.String(),
		)
		suite.Require().NoError(err)
		shareToken, err := suite.DataPoolKeeper.SellData(suite.Ctx, sellerAddr, *cert)
		suite.Require().NoError(err)
		suite.Require().Equal("DP/1", shareToken.Denom)
		suite.Require().Equal(sdk.NewInt(1), shareToken.Amount)
	}

	// sell all seller data to the second pool
	for i, sellerAddr := range sellers {
		cert, err := makeTestDataCertificate(
			suite.Cdc.Marshaler,
			secondPoolID,
			1,
			[]byte(fmt.Sprintf("dataHash_%v", i)),
			sellerAddr.String(),
		)
		suite.Require().NoError(err)
		shareToken, err := suite.DataPoolKeeper.SellData(suite.Ctx, sellerAddr, *cert)
		suite.Require().NoError(err)
		suite.Require().Equal("DP/2", shareToken.Denom)
		suite.Require().Equal(sdk.NewInt(1), shareToken.Amount)
	}

	// check pool status
	pool, err := suite.DataPoolKeeper.GetPool(suite.Ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(types.PENDING, pool.Status)
	secondPool, err := suite.DataPoolKeeper.GetPool(suite.Ctx, secondPoolID)
	suite.Require().NoError(err)
	suite.Require().Equal(types.ACTIVE, secondPool.Status)

	// check balances of curator and sellers
	depositAmount := suite.DataPoolKeeper.GetParams(suite.Ctx).DataPoolDeposit.Amount
	// minus two times deposit
	expectedCuratorAmount := fundForCurator.AmountOf(assets.MicroMedDenom).Sub(depositAmount.Mul(sdk.NewInt(2)))
	curatorAmount := suite.BankKeeper.GetBalance(suite.Ctx, curatorAddr, assets.MicroMedDenom).Amount
	suite.Require().Equal(expectedCuratorAmount, curatorAmount)
	// amount of all sellers is zero
	for _, sellerAddr := range sellers {
		coin := suite.BankKeeper.GetBalance(suite.Ctx, sellerAddr, assets.MicroMedDenom)
		suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(0)), coin)
	}

	// execute a distribute sales revenue
	err = suite.DataPoolKeeper.DistributeRevenuePools(suite.Ctx)
	suite.Require().NoError(err)

	// check balances of curator and sellers after distribution
	// minus one times deposit
	expectedCuratorAmount = fundForCurator.AmountOf(assets.MicroMedDenom).Sub(depositAmount)
	curatorAmount = suite.BankKeeper.GetBalance(suite.Ctx, curatorAddr, assets.MicroMedDenom).Amount
	suite.Require().Equal(expectedCuratorAmount, curatorAmount)
	secondPool, err = suite.DataPoolKeeper.GetPool(suite.Ctx, secondPoolID)
	suite.Require().NoError(err)
	for _, sellerAddr := range sellers {
		coin := suite.BankKeeper.GetBalance(suite.Ctx, sellerAddr, assets.MicroMedDenom)
		// Since no NFTs have been sold, the seller's balance is zero.
		suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(0)), coin)
	}

	poolAddr, err := sdk.AccAddressFromBech32(pool.PoolAddress)
	suite.Require().NoError(err)
	secondPoolAddr, err := sdk.AccAddressFromBech32(secondPool.PoolAddress)
	suite.Require().NoError(err)
	// The first pool amount only requires a deposit
	deposit := suite.DataPoolKeeper.GetParams(suite.Ctx).DataPoolDeposit
	suite.Require().Equal(deposit, suite.BankKeeper.GetBalance(suite.Ctx, poolAddr, assets.MicroMedDenom))
	// The second pool amount is zero
	suite.Require().Equal(int64(0), suite.BankKeeper.GetBalance(suite.Ctx, secondPoolAddr, assets.MicroMedDenom).Amount.Int64())
}
