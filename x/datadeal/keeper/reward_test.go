package keeper_test

import (
	"encoding/base64"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datadeal/testutil"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
)

type rewardTestSuite struct {
	testutil.DataDealBaseTestSuite

	defaultFunds sdk.Coins

	sellerAccPrivKey cryptotypes.PrivKey
	sellerAccPubKey  cryptotypes.PubKey
	sellerAccAddr    sdk.AccAddress

	buyerAccAddr sdk.AccAddress

	oracleAcc1PrivKey cryptotypes.PrivKey
	oracleAcc1PubKey  cryptotypes.PubKey
	oracleAcc1Addr    sdk.AccAddress

	oracleAcc2PrivKey cryptotypes.PrivKey
	oracleAcc2PubKey  cryptotypes.PubKey
	oracleAcc2Addr    sdk.AccAddress

	oraclePrivKey *btcec.PrivateKey
	oraclePubKey  *btcec.PublicKey
}

func TestRewardTestSuite(t *testing.T) {
	suite.Run(t, new(rewardTestSuite))
}

func (suite *rewardTestSuite) BeforeTest(_, _ string) {
	ctx := suite.Ctx

	suite.defaultFunds = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10_000_000_000))) // 10,000 MED

	suite.sellerAccPrivKey = secp256k1.GenPrivKey()
	suite.sellerAccPubKey = suite.sellerAccPrivKey.PubKey()
	suite.sellerAccAddr = sdk.AccAddress(suite.sellerAccPubKey.Address())

	suite.buyerAccAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	suite.oracleAcc1PrivKey = secp256k1.GenPrivKey()
	suite.oracleAcc1PubKey = suite.oracleAcc1PrivKey.PubKey()
	suite.oracleAcc1Addr = sdk.AccAddress(suite.oracleAcc1PubKey.Address())

	suite.oracleAcc2PrivKey = secp256k1.GenPrivKey()
	suite.oracleAcc2PubKey = suite.oracleAcc2PrivKey.PubKey()
	suite.oracleAcc2Addr = sdk.AccAddress(suite.oracleAcc2PubKey.Address())

	suite.oraclePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.oraclePubKey = suite.oraclePrivKey.PubKey()

	err := suite.DataDealKeeper.SetNextDealNumber(ctx, 1)
	suite.Require().NoError(err)

	suite.OracleKeeper.SetParams(ctx, oracletypes.Params{
		OraclePublicKey:          base64.StdEncoding.EncodeToString(suite.oraclePubKey.SerializeCompressed()),
		OraclePubKeyRemoteReport: "",
		UniqueId:                 "uniqueID",
		OracleCommissionRate:     sdk.NewDecWithPrec(1, 1), // 10% oracle commission
		VoteParams: oracletypes.VoteParams{
			VotingPeriod: 100,
			JailPeriod:   60,
			Threshold:    sdk.NewDecWithPrec(2, 3),
		},
		SlashParams: oracletypes.SlashParams{
			SlashFractionDowntime: sdk.NewDecWithPrec(3, 1),
			SlashFractionForgery:  sdk.NewDecWithPrec(1, 1),
		},
	})
}

func (suite rewardTestSuite) TestSellerRewardDistribution() {
	ctx := suite.Ctx

	err := suite.FundAccount(ctx, suite.buyerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	dealAddr := types.NewDealAddress(1)
	msgCreateDeal := &types.MsgCreateDeal{
		DataSchema:   []string{"http://jsonld.com"},
		Budget:       &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(100_000_000)}, // 100 MED,
		MaxNumData:   1,
		BuyerAddress: suite.buyerAccAddr.String(),
	}

	buyer, err := sdk.AccAddressFromBech32(suite.buyerAccAddr.String())
	suite.Require().NoError(err)

	dealID, err := suite.DataDealKeeper.CreateDeal(ctx, buyer, msgCreateDeal)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), dealID)

	// balance check before reward distribution
	dealBalance := suite.BankKeeper.GetBalance(ctx, dealAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(100_000_000)), dealBalance)

	sellerBalance := suite.BankKeeper.GetBalance(ctx, suite.sellerAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(0)), sellerBalance)

	dataSale := types.NewDataSale(&types.MsgSellData{
		DealId:        1,
		VerifiableCid: "verifiableCID",
		SellerAddress: suite.sellerAccAddr.String(),
	})

	// distribute data verification rewards
	err = suite.DataDealKeeper.DistributeVerificationRewards(ctx, dataSale)
	suite.Require().NoError(err)

	// TODO check balance of oracles

	// balance check after reward distribution
	dealBalance = suite.BankKeeper.GetBalance(ctx, dealAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10_000_000)), dealBalance)

	sellerBalance = suite.BankKeeper.GetBalance(ctx, suite.sellerAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(90_000_000)), sellerBalance)

}

// Assumes that 2 oracles(validators)
//
// oracle commission : 10%
//
// ****** oracle1 ******
// voting power : 10
// validator commission rate : 10%
//
// ****** oracle2 ******
// voting power : 30
// validator commission rate : 50%
func (suite *rewardTestSuite) TestOracleRewardDistribution() {
	ctx := suite.Ctx

	err := suite.FundAccount(ctx, suite.buyerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	// set validators
	val1VotingPower := sdk.NewInt(10)
	val2VotingPower := sdk.NewInt(30)

	val1Commission := sdk.NewDecWithPrec(1, 1)
	val2Commission := sdk.NewDecWithPrec(5, 1)
	val1 := suite.SetValidator(suite.oracleAcc1PubKey, val1VotingPower, val1Commission)
	val2 := suite.SetValidator(suite.oracleAcc2PubKey, val2VotingPower, val2Commission)

	// set deal
	dealAddr := types.NewDealAddress(1)
	msgCreateDeal := &types.MsgCreateDeal{
		DataSchema:   []string{"http://jsonld.com"},
		Budget:       &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(400_000_000)}, // 400 MED,
		MaxNumData:   1,
		BuyerAddress: suite.buyerAccAddr.String(),
	}

	buyer, err := sdk.AccAddressFromBech32(suite.buyerAccAddr.String())
	suite.Require().NoError(err)

	dealID, err := suite.DataDealKeeper.CreateDeal(ctx, buyer, msgCreateDeal)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), dealID)

	// the balance of deal is 400 MED
	dealBalance := suite.BankKeeper.GetBalance(ctx, dealAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(400_000_000)), dealBalance)

	// oracle lists to be rewarded (2 oracles)
	oracles := make(map[string]*oracletypes.OracleValidatorInfo)
	oracles[suite.oracleAcc1Addr.String()] = &oracletypes.OracleValidatorInfo{
		Address:         suite.oracleAcc1Addr.String(),
		OracleActivated: true,
		BondedTokens:    sdk.NewInt(10),
		ValidatorJailed: false,
	}
	oracles[suite.oracleAcc2Addr.String()] = &oracletypes.OracleValidatorInfo{
		Address:         suite.oracleAcc2Addr.String(),
		OracleActivated: true,
		BondedTokens:    sdk.NewInt(30),
		ValidatorJailed: false,
	}

	// before distribution, the balance of distrModuleAcc is 0 MED
	distrModuleAcc := suite.AccountKeeper.GetModuleAccount(ctx, distrtypes.ModuleName).GetAddress()
	distrBalance := suite.BankKeeper.GetBalance(ctx, distrModuleAcc, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.ZeroInt()), distrBalance)

	// distribute rewards to oracles
	suite.DataDealKeeper.DistributeOracleRewards(ctx, dealID, oracles)

	// after distribution, the balance of deal is 360 MED. 10%(40 MED) was sent to distrModuleAcc for reward
	dealBalance = suite.BankKeeper.GetBalance(ctx, dealAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(360_000_000)), dealBalance)

	// the balance of distrModuleAcc is 40 MED, which is 10% of 400 MED
	distrBalance = suite.BankKeeper.GetBalance(ctx, distrModuleAcc, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(40_000_000)), distrBalance)

	// Among total oracle commission(40 MED), 25% is for oracle 1 and 75% is for oracle 2 according to the respective voting power
	// 10 MED (25% of 40 MED) for oracle 1
	// - 1 MED (validator commission) + 9 MED (reward)
	//
	// 30 MED (75% of 40 MED) for oracle 2
	// - 15 MED (validator commission) + 15 MED (reward)

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(1_000_000))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, val1.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(9_000_000))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, val1.GetOperator()).Rewards)

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(15_000_000))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, val2.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(15_000_000))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, val2.GetOperator()).Rewards)
}
