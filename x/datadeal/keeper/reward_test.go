package keeper_test

import (
	"encoding/base64"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

	delegatorAccAddr sdk.AccAddress

	oracleAcc1PrivKey cryptotypes.PrivKey
	oracleAcc1PubKey  cryptotypes.PubKey
	oracleAcc1Addr    sdk.AccAddress

	oracleAcc2PrivKey cryptotypes.PrivKey
	oracleAcc2PubKey  cryptotypes.PubKey
	oracleAcc2Addr    sdk.AccAddress

	oracleAcc3PrivKey cryptotypes.PrivKey
	oracleAcc3PubKey  cryptotypes.PubKey
	oracleAcc3Addr    sdk.AccAddress

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

	suite.delegatorAccAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	suite.oracleAcc1PrivKey = secp256k1.GenPrivKey()
	suite.oracleAcc1PubKey = suite.oracleAcc1PrivKey.PubKey()
	suite.oracleAcc1Addr = sdk.AccAddress(suite.oracleAcc1PubKey.Address())

	suite.oracleAcc2PrivKey = secp256k1.GenPrivKey()
	suite.oracleAcc2PubKey = suite.oracleAcc2PrivKey.PubKey()
	suite.oracleAcc2Addr = sdk.AccAddress(suite.oracleAcc2PubKey.Address())

	suite.oracleAcc3PrivKey = secp256k1.GenPrivKey()
	suite.oracleAcc3PubKey = suite.oracleAcc3PrivKey.PubKey()
	suite.oracleAcc3Addr = sdk.AccAddress(suite.oracleAcc3PubKey.Address())

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

func (suite *rewardTestSuite) TestBasicVerificationRewardDistribution() {
	ctx := suite.Ctx

	err := suite.FundAccount(ctx, suite.buyerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	val1VotingPower := sdk.NewInt(100)
	val1Commission := sdk.NewDecWithPrec(1, 1)
	val1 := suite.SetValidator(suite.oracleAcc1PubKey, val1VotingPower, val1Commission)

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

	oracleBalance := suite.BankKeeper.GetBalance(ctx, suite.oracleAcc1Addr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(0)), oracleBalance)

	dataSale := types.NewDataSale(&types.MsgSellData{
		DealId:        1,
		VerifiableCid: "verifiableCID",
		SellerAddress: suite.sellerAccAddr.String(),
	})

	voters := []*oracletypes.VoterInfo{
		{
			VoterAddress: suite.oracleAcc1Addr.String(),
			VotingPower:  sdk.NewInt(100),
		},
	}

	// distribute data verification rewards
	err = suite.DataDealKeeper.DistributeVerificationRewards(ctx, dataSale, voters)
	suite.Require().NoError(err)

	// balance check after reward distribution

	// 5 MED is remained, which is a reward for data delivery
	dealBalance = suite.BankKeeper.GetBalance(ctx, dealAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(5_000_000)), dealBalance)

	// 90 MED is for seller
	sellerBalance = suite.BankKeeper.GetBalance(ctx, suite.sellerAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(90_000_000)), sellerBalance)

	// 5 MED for data verification reward
	// - 0.5 MED (10% of 5 MED) for commission
	// - 4.5 MED (90% of 5 MED) for reward
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(500_000))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, val1.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(4_500_000))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, val1.GetOperator()).Rewards)
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
func (suite *rewardTestSuite) TestVerificationRewardDistributionMultipleOracles() {
	ctx := suite.Ctx

	err := suite.FundAccount(ctx, suite.buyerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	// set validators
	val1VotingPower := sdk.NewInt(10)
	val2VotingPower := sdk.NewInt(30)

	val1Commission := sdk.NewDecWithPrec(1, 1) // 10% commission
	val2Commission := sdk.NewDecWithPrec(5, 1) // 50% commission
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

	dataSale := types.NewDataSale(&types.MsgSellData{
		DealId:        1,
		VerifiableCid: "verifiableCID",
		SellerAddress: suite.sellerAccAddr.String(),
	})

	// oracle lists to be rewarded (2 oracles)
	voters := []*oracletypes.VoterInfo{
		{
			VoterAddress: suite.oracleAcc1Addr.String(),
			VotingPower:  sdk.NewInt(10),
		},
		{
			VoterAddress: suite.oracleAcc2Addr.String(),
			VotingPower:  sdk.NewInt(30),
		},
	}

	// balance check before distribution
	// the balance of distrModuleAcc is 0 MED
	distrModuleAcc := suite.AccountKeeper.GetModuleAccount(ctx, distrtypes.ModuleName).GetAddress()
	distrBalance := suite.BankKeeper.GetBalance(ctx, distrModuleAcc, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.ZeroInt()), distrBalance)

	// the balance of deal is 400 MED
	dealBalance := suite.BankKeeper.GetBalance(ctx, dealAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(400_000_000)), dealBalance)

	// the balance of seller is 0 MED
	sellerBalance := suite.BankKeeper.GetBalance(ctx, suite.sellerAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.ZeroInt()), sellerBalance)

	// distribute rewards to oracles
	err = suite.DataDealKeeper.DistributeVerificationRewards(ctx, dataSale, voters)
	suite.Require().NoError(err)

	// after distribution, the balance of deal is 20 MED.
	// 90% (360 MED) - seller
	// 5% (20 MED) - distrModuleAcc for data verification reward
	// 5% (20 MED) - remained for data delivery reward
	dealBalance = suite.BankKeeper.GetBalance(ctx, dealAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(20_000_000)), dealBalance)

	// the balance of distrModuleAcc is 40 MED, which is 10% of 400 MED
	distrBalance = suite.BankKeeper.GetBalance(ctx, distrModuleAcc, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(20_000_000)), distrBalance)

	sellerBalance = suite.BankKeeper.GetBalance(ctx, suite.sellerAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(360_000_000)), sellerBalance)

	// Among total oracle commission(20 MED), 25% is for oracle 1 and 75% is for oracle 2 according to the respective voting power
	// 5 MED (25% of 20 MED) for oracle 1
	// - 0.5 MED (validator commission) + 4.5 MED (reward)
	//
	// 15 MED (75% of 20 MED) for oracle 2
	// - 7.5 MED (validator commission) + 7.5 MED (reward)

	val1Updated := suite.StakingKeeper.Validator(ctx, val1.GetOperator())

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(500_000))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, val1Updated.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(4_500_000))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, val1Updated.GetOperator()).Rewards)

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(7_500_000))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, val2.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(7_500_000))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, val2.GetOperator()).Rewards)
}

// Assumes that 3 oracles(validators)
//
// oracle commission : 10%
//
// ****** oracle1 ******
// voting power : 10
// validator commission rate : 10%
//
// ****** oracle2 ******
// voting power : 20
// validator commission rate : 50%
//
// ****** oracle3 ******
// voting power : 20
// validator commission rate : 20%
// self-delegated voting power : 10
// delegated voting power : 10
func (suite *rewardTestSuite) TestVerificationRewardDistributionWithDelegators() {
	ctx := suite.Ctx

	err := suite.FundAccount(ctx, suite.buyerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	err = suite.FundAccount(ctx, suite.delegatorAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	err = suite.FundAccount(ctx, suite.oracleAcc1Addr, suite.defaultFunds)
	suite.Require().NoError(err)

	err = suite.FundAccount(ctx, suite.oracleAcc2Addr, suite.defaultFunds)
	suite.Require().NoError(err)

	err = suite.FundAccount(ctx, suite.oracleAcc3Addr, suite.defaultFunds)
	suite.Require().NoError(err)

	// set validators
	val1VotingPower := sdk.NewInt(10)
	val2VotingPower := sdk.NewInt(20)
	val3VotingPower := sdk.NewInt(10)

	val1Commission := sdk.NewDecWithPrec(1, 1)
	val2Commission := sdk.NewDecWithPrec(5, 1)
	val3Commission := sdk.NewDecWithPrec(2, 1)
	val1 := suite.SetValidator(suite.oracleAcc1PubKey, val1VotingPower, val1Commission)
	val2 := suite.SetValidator(suite.oracleAcc2PubKey, val2VotingPower, val2Commission)
	val3 := suite.SetValidator(suite.oracleAcc3PubKey, val3VotingPower, val3Commission)

	// self delegation
	_, err = suite.StakingKeeper.Delegate(ctx, suite.oracleAcc3Addr, sdk.NewInt(10), stakingtypes.Unbonded, val3, true)
	suite.Require().NoError(err)

	// delegation
	_, err = suite.StakingKeeper.Delegate(ctx, suite.delegatorAccAddr, sdk.NewInt(10), stakingtypes.Unbonded, val3, true)
	suite.Require().NoError(err)

	// initial setting for validator rewards
	suite.SetupValidatorRewards(val1.GetOperator())
	suite.SetupValidatorRewards(val2.GetOperator())
	suite.SetupValidatorRewards(val3.GetOperator())
	suite.DistrKeeper.SetDelegatorStartingInfo(ctx, val3.GetOperator(), suite.delegatorAccAddr, distrtypes.NewDelegatorStartingInfo(1, sdk.NewInt(10).ToDec(), 1))
	suite.DistrKeeper.SetDelegatorStartingInfo(ctx, val3.GetOperator(), suite.oracleAcc3Addr, distrtypes.NewDelegatorStartingInfo(1, sdk.NewInt(10).ToDec(), 1))

	// set deal
	dealAddr := types.NewDealAddress(1)
	msgCreateDeal := &types.MsgCreateDeal{
		DataSchema:   []string{"http://jsonld.com"},
		Budget:       &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(500_000_000)}, // 500 MED,
		MaxNumData:   1,
		BuyerAddress: suite.buyerAccAddr.String(),
	}

	buyer, err := sdk.AccAddressFromBech32(suite.buyerAccAddr.String())
	suite.Require().NoError(err)

	dealID, err := suite.DataDealKeeper.CreateDeal(ctx, buyer, msgCreateDeal)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), dealID)

	dataSale := types.NewDataSale(&types.MsgSellData{
		DealId:        1,
		VerifiableCid: "verifiableCID",
		SellerAddress: suite.sellerAccAddr.String(),
	})

	// oracle lists to be rewarded (3 oracles)
	voters := []*oracletypes.VoterInfo{
		{
			VoterAddress: suite.oracleAcc1Addr.String(),
			VotingPower:  sdk.NewInt(10),
		},
		{
			VoterAddress: suite.oracleAcc2Addr.String(),
			VotingPower:  sdk.NewInt(20),
		},
		{
			VoterAddress: suite.oracleAcc3Addr.String(),
			VotingPower:  sdk.NewInt(20),
		},
	}

	// before distribution, the balance of distrModuleAcc is 0 MED
	distrModuleAcc := suite.AccountKeeper.GetModuleAccount(ctx, distrtypes.ModuleName).GetAddress()
	distrBalance := suite.BankKeeper.GetBalance(ctx, distrModuleAcc, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.ZeroInt()), distrBalance)

	// the balance of deal is 500 MED
	dealBalance := suite.BankKeeper.GetBalance(ctx, dealAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(500_000_000)), dealBalance)

	// distribute rewards to oracles
	err = suite.DataDealKeeper.DistributeVerificationRewards(ctx, dataSale, voters)
	suite.Require().NoError(err)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// after distribution, the balance of deal is 25 MED.
	// Among oracle commission (50 MED, 10% of 500 MED), half is remained for data delivery reward.
	dealBalance = suite.BankKeeper.GetBalance(ctx, dealAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(25_000_000)), dealBalance)

	// the balance of distrModuleAcc is 25 MED, which is half of 10% of 500 MED
	distrBalance = suite.BankKeeper.GetBalance(ctx, distrModuleAcc, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(25_000_000)), distrBalance)

	// Among total oracle commission(25 MED),
	// 5 MED (20% of 50 MED) for oracle 1
	// - 0.5 MED (validator commission) + 4.5 MED (validator reward)
	//
	// 10 MED (40% of 50 MED) for oracle 2
	// - 5 MED (validator commission) + 5 MED (validator reward)
	//
	// 10 MED (40% of 50 MED) for oracle 3
	// - 2 MED (validator commission) + 4 MED (validator reward) + 4 MED (delegator reward)

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(500_000))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, val1.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(4_500_000))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, val1.GetOperator()).Rewards)

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(5_000_000))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, val2.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(5_000_000))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, val2.GetOperator()).Rewards)

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(2_000_000))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, val3.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(8_000_000))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, val3.GetOperator()).Rewards)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	del := suite.StakingKeeper.Delegation(ctx, suite.delegatorAccAddr, val3.GetOperator())
	selfDel := suite.StakingKeeper.Delegation(ctx, suite.oracleAcc3Addr, val3.GetOperator())

	val3Updated := suite.StakingKeeper.Validator(ctx, val3.GetOperator())
	endingPeriod := suite.DistrKeeper.IncrementValidatorPeriod(ctx, val3Updated)
	delegatorReward := suite.DistrKeeper.CalculateDelegationRewards(ctx, val3Updated, del, endingPeriod)
	validatorReward := suite.DistrKeeper.CalculateDelegationRewards(ctx, val3Updated, selfDel, endingPeriod)

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(4_000_000))), delegatorReward)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(4_000_000))), validatorReward)
}

func (suite rewardTestSuite) TestBasicDeliveryRewardDistribution() {
	ctx := suite.Ctx

	err := suite.FundAccount(ctx, suite.buyerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	val1VotingPower := sdk.NewInt(100)
	val1Commission := sdk.NewDecWithPrec(1, 1)
	val1 := suite.SetValidator(suite.oracleAcc1PubKey, val1VotingPower, val1Commission)

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

	oracleBalance := suite.BankKeeper.GetBalance(ctx, suite.oracleAcc1Addr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(0)), oracleBalance)

	dataSale := types.NewDataSale(&types.MsgSellData{
		DealId:        1,
		VerifiableCid: "verifiableCID",
		SellerAddress: suite.sellerAccAddr.String(),
	})

	voters := []*oracletypes.VoterInfo{
		{
			VoterAddress: suite.oracleAcc1Addr.String(),
			VotingPower:  sdk.NewInt(100),
		},
	}

	err = suite.DataDealKeeper.DistributeDeliveryRewards(ctx, dataSale, voters)
	suite.Require().NoError(err)

	// balance check after reward distribution

	// 5 MED is remained, which is a reward for data delivery
	dealBalance = suite.BankKeeper.GetBalance(ctx, dealAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(95_000_000)), dealBalance)

	// 5 MED for data verification reward
	// - 0.5 MED (10% of 5 MED) for commission
	// - 4.5 MED (90% of 5 MED) for reward
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(500_000))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, val1.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(4_500_000))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, val1.GetOperator()).Rewards)
}

// Assumes that 3 oracles(validators)
//
// oracle commission : 10%
//
// ****** oracle1 ******
// voting power : 10
// validator commission rate : 10%
//
// ****** oracle2 ******
// voting power : 20
// validator commission rate : 50%
//
// ****** oracle3 ******
// voting power : 20
// validator commission rate : 20%
// self-delegated voting power : 10
// delegated voting power : 10
func (suite *rewardTestSuite) TestDeliveryRewardDistributionWithDelegators() {
	ctx := suite.Ctx

	err := suite.FundAccount(ctx, suite.buyerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	err = suite.FundAccount(ctx, suite.delegatorAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	err = suite.FundAccount(ctx, suite.oracleAcc1Addr, suite.defaultFunds)
	suite.Require().NoError(err)

	err = suite.FundAccount(ctx, suite.oracleAcc2Addr, suite.defaultFunds)
	suite.Require().NoError(err)

	err = suite.FundAccount(ctx, suite.oracleAcc3Addr, suite.defaultFunds)
	suite.Require().NoError(err)

	// set validators
	val1VotingPower := sdk.NewInt(10)
	val2VotingPower := sdk.NewInt(20)
	val3VotingPower := sdk.NewInt(10)

	val1Commission := sdk.NewDecWithPrec(1, 1)
	val2Commission := sdk.NewDecWithPrec(5, 1)
	val3Commission := sdk.NewDecWithPrec(2, 1)
	val1 := suite.SetValidator(suite.oracleAcc1PubKey, val1VotingPower, val1Commission)
	val2 := suite.SetValidator(suite.oracleAcc2PubKey, val2VotingPower, val2Commission)
	val3 := suite.SetValidator(suite.oracleAcc3PubKey, val3VotingPower, val3Commission)

	// self delegation
	_, err = suite.StakingKeeper.Delegate(ctx, suite.oracleAcc3Addr, sdk.NewInt(10), stakingtypes.Unbonded, val3, true)
	suite.Require().NoError(err)

	// delegation
	_, err = suite.StakingKeeper.Delegate(ctx, suite.delegatorAccAddr, sdk.NewInt(10), stakingtypes.Unbonded, val3, true)
	suite.Require().NoError(err)

	// initial setting for validator rewards
	suite.SetupValidatorRewards(val1.GetOperator())
	suite.SetupValidatorRewards(val2.GetOperator())
	suite.SetupValidatorRewards(val3.GetOperator())
	suite.DistrKeeper.SetDelegatorStartingInfo(ctx, val3.GetOperator(), suite.delegatorAccAddr, distrtypes.NewDelegatorStartingInfo(1, sdk.NewInt(10).ToDec(), 1))
	suite.DistrKeeper.SetDelegatorStartingInfo(ctx, val3.GetOperator(), suite.oracleAcc3Addr, distrtypes.NewDelegatorStartingInfo(1, sdk.NewInt(10).ToDec(), 1))

	// set deal
	dealAddr := types.NewDealAddress(1)
	msgCreateDeal := &types.MsgCreateDeal{
		DataSchema:   []string{"http://jsonld.com"},
		Budget:       &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(500_000_000)}, // 500 MED,
		MaxNumData:   1,
		BuyerAddress: suite.buyerAccAddr.String(),
	}

	buyer, err := sdk.AccAddressFromBech32(suite.buyerAccAddr.String())
	suite.Require().NoError(err)

	dealID, err := suite.DataDealKeeper.CreateDeal(ctx, buyer, msgCreateDeal)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), dealID)

	dataSale := types.NewDataSale(&types.MsgSellData{
		DealId:        1,
		VerifiableCid: "verifiableCID",
		SellerAddress: suite.sellerAccAddr.String(),
	})

	// oracle lists to be rewarded (3 oracles)
	voters := []*oracletypes.VoterInfo{
		{
			VoterAddress: suite.oracleAcc1Addr.String(),
			VotingPower:  sdk.NewInt(10),
		},
		{
			VoterAddress: suite.oracleAcc2Addr.String(),
			VotingPower:  sdk.NewInt(20),
		},
		{
			VoterAddress: suite.oracleAcc3Addr.String(),
			VotingPower:  sdk.NewInt(20),
		},
	}

	// before distribution, the balance of distrModuleAcc is 0 MED
	distrModuleAcc := suite.AccountKeeper.GetModuleAccount(ctx, distrtypes.ModuleName).GetAddress()
	distrBalance := suite.BankKeeper.GetBalance(ctx, distrModuleAcc, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.ZeroInt()), distrBalance)

	// the balance of deal is 500 MED before any reward distribution
	dealBalance := suite.BankKeeper.GetBalance(ctx, dealAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(500_000_000)), dealBalance)

	// distribute rewards to oracles
	err = suite.DataDealKeeper.DistributeDeliveryRewards(ctx, dataSale, voters)
	suite.Require().NoError(err)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// after distribution, the balance of deal is 475 MED.
	// Among oracle commission (50 MED, 10% of 500 MED), half is distributed for data delivery reward.
	dealBalance = suite.BankKeeper.GetBalance(ctx, dealAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(475_000_000)), dealBalance)

	// the balance of distrModuleAcc is 25 MED, which is half of 10% of 500 MED
	distrBalance = suite.BankKeeper.GetBalance(ctx, distrModuleAcc, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(25_000_000)), distrBalance)

	// Among total oracle commission(25 MED),
	// 5 MED (20% of 50 MED) for oracle 1
	// - 0.5 MED (validator commission) + 4.5 MED (validator reward)
	//
	// 10 MED (40% of 50 MED) for oracle 2
	// - 5 MED (validator commission) + 5 MED (validator reward)
	//
	// 10 MED (40% of 50 MED) for oracle 3
	// - 2 MED (validator commission) + 4 MED (validator reward) + 4 MED (delegator reward)

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(500_000))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, val1.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(4_500_000))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, val1.GetOperator()).Rewards)

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(5_000_000))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, val2.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(5_000_000))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, val2.GetOperator()).Rewards)

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(2_000_000))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, val3.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(8_000_000))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, val3.GetOperator()).Rewards)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	del := suite.StakingKeeper.Delegation(ctx, suite.delegatorAccAddr, val3.GetOperator())
	selfDel := suite.StakingKeeper.Delegation(ctx, suite.oracleAcc3Addr, val3.GetOperator())

	val3Updated := suite.StakingKeeper.Validator(ctx, val3.GetOperator())
	endingPeriod := suite.DistrKeeper.IncrementValidatorPeriod(ctx, val3Updated)
	delegatorReward := suite.DistrKeeper.CalculateDelegationRewards(ctx, val3Updated, del, endingPeriod)
	validatorReward := suite.DistrKeeper.CalculateDelegationRewards(ctx, val3Updated, selfDel, endingPeriod)

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(4_000_000))), delegatorReward)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(4_000_000))), validatorReward)
}
