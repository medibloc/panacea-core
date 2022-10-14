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
func (suite *rewardTestSuite) TestRewardDistribution() {
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
	suite.DataDealKeeper.DistributeRewards(ctx, dealID, oracles)

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
func (suite *rewardTestSuite) TestRewardDistributionWithDelegators() {
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

	// the balance of deal is 500 MED
	dealBalance := suite.BankKeeper.GetBalance(ctx, dealAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(500_000_000)), dealBalance)

	// oracle lists to be rewarded (3 oracles)
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
		BondedTokens:    sdk.NewInt(20),
		ValidatorJailed: false,
	}
	oracles[suite.oracleAcc3Addr.String()] = &oracletypes.OracleValidatorInfo{
		Address:         suite.oracleAcc3Addr.String(),
		OracleActivated: true,
		BondedTokens:    sdk.NewInt(20),
		ValidatorJailed: false,
	}

	// before distribution, the balance of distrModuleAcc is 0 MED
	distrModuleAcc := suite.AccountKeeper.GetModuleAccount(ctx, distrtypes.ModuleName).GetAddress()
	distrBalance := suite.BankKeeper.GetBalance(ctx, distrModuleAcc, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.ZeroInt()), distrBalance)

	// distribute rewards to oracles
	suite.DataDealKeeper.DistributeRewards(ctx, dealID, oracles)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// after distribution, the balance of deal is 450 MED. 10%(50 MED) was sent to distrModuleAcc for reward
	dealBalance = suite.BankKeeper.GetBalance(ctx, dealAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(450_000_000)), dealBalance)

	// the balance of distrModuleAcc is 50 MED, which is 10% of 500 MED
	distrBalance = suite.BankKeeper.GetBalance(ctx, distrModuleAcc, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(50_000_000)), distrBalance)

	// Among total oracle commission(50 MED),
	// 10 MED (20% of 50 MED) for oracle 1
	// - 1 MED (validator commission) + 9 MED (validator reward)
	//
	// 20 MED (40% of 50 MED) for oracle 2
	// - 10 MED (validator commission) + 10 MED (validator reward)
	//
	// 20 MED (40% of 50 MED) for oracle 3
	// - 4 MED (validator commission) + 8 MED (validator reward) + 8 MED (delegator reward)

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(1_000_000))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, val1.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(9_000_000))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, val1.GetOperator()).Rewards)

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(10_000_000))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, val2.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(10_000_000))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, val2.GetOperator()).Rewards)

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(4_000_000))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, val3.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(16_000_000))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, val3.GetOperator()).Rewards)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	del := suite.StakingKeeper.Delegation(ctx, suite.delegatorAccAddr, val3.GetOperator())
	selfDel := suite.StakingKeeper.Delegation(ctx, suite.oracleAcc3Addr, val3.GetOperator())

	val3Updated := suite.StakingKeeper.Validator(ctx, val3.GetOperator())
	endingPeriod := suite.DistrKeeper.IncrementValidatorPeriod(ctx, val3Updated)
	delegatorReward := suite.DistrKeeper.CalculateDelegationRewards(ctx, val3Updated, del, endingPeriod)
	validatorReward := suite.DistrKeeper.CalculateDelegationRewards(ctx, val3Updated, selfDel, endingPeriod)

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(8_000_000))), delegatorReward)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(8_000_000))), validatorReward)
}
