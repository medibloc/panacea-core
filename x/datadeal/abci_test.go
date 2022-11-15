package datadeal_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datadeal"
	"github.com/medibloc/panacea-core/v2/x/datadeal/testutil"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
)

type abciTestSuite struct {
	testutil.DataDealBaseTestSuite

	defaultFunds sdk.Coins

	sellerAccPrivKey cryptotypes.PrivKey
	sellerAccPubKey  cryptotypes.PubKey
	sellerAccAddr    sdk.AccAddress

	buyerAccAddr sdk.AccAddress

	verifiableCID  string
	verifiableCID2 string
	dataHash       string
	dataHash2      string

	oraclePubKey  cryptotypes.PubKey
	oracleAddr    sdk.AccAddress
	oraclePubKey2 cryptotypes.PubKey
	oracleAddr2   sdk.AccAddress
	oraclePubKey3 cryptotypes.PubKey
	oracleAddr3   sdk.AccAddress

	uniqueID string
	dealID   uint64

	BlockPeriod time.Duration
}

func TestAbciTestSuite(t *testing.T) {
	suite.Run(t, new(abciTestSuite))
}

func (suite *abciTestSuite) BeforeTest(_, _ string) {

	ctx := suite.Ctx
	suite.uniqueID = "uniqueID"

	suite.defaultFunds = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(100_000_000_000))) // 100,000 MED

	suite.sellerAccPrivKey = secp256k1.GenPrivKey()
	suite.sellerAccPubKey = suite.sellerAccPrivKey.PubKey()
	suite.sellerAccAddr = sdk.AccAddress(suite.sellerAccPubKey.Address())

	suite.buyerAccAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err := suite.FundAccount(ctx, suite.buyerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	suite.dataHash = "dataHash1"
	suite.dataHash2 = "dataHash2"
	suite.verifiableCID = "verifiableCID"
	suite.verifiableCID2 = "verifiableCID2"

	suite.oraclePubKey = secp256k1.GenPrivKey().PubKey()
	suite.oracleAddr = sdk.AccAddress(suite.oraclePubKey.Address())
	suite.oraclePubKey2 = secp256k1.GenPrivKey().PubKey()
	suite.oracleAddr2 = sdk.AccAddress(suite.oraclePubKey2.Address())
	suite.oraclePubKey3 = secp256k1.GenPrivKey().PubKey()
	suite.oracleAddr3 = sdk.AccAddress(suite.oraclePubKey3.Address())

	oraclePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	suite.Require().NoError(err)
	suite.OracleKeeper.SetParams(ctx, oracletypes.Params{
		OraclePublicKey:          base64.StdEncoding.EncodeToString(oraclePrivKey.PubKey().SerializeCompressed()),
		OraclePubKeyRemoteReport: base64.StdEncoding.EncodeToString([]byte("oraclePubKeyRemoteReport")),
		UniqueId:                 suite.uniqueID,
		OracleCommissionRate:     sdk.NewDecWithPrec(1, 1),
		VoteParams: oracletypes.VoteParams{
			VotingPeriod: 30 * time.Second,
			JailPeriod:   10 * time.Minute,
			Threshold:    sdk.NewDec(2).Quo(sdk.NewDec(3)),
		},
		SlashParams: oracletypes.SlashParams{
			SlashFractionDowntime: sdk.NewDecWithPrec(3, 1),
			SlashFractionForgery:  sdk.NewDecWithPrec(1, 1),
		},
	})
	suite.DataDealKeeper.SetParams(suite.Ctx, types.DefaultParams())

	err = suite.DataDealKeeper.SetNextDealNumber(ctx, 1)
	suite.Require().NoError(err)

	budget := &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(10_000_000_000)} // 10,000 MED

	msgCreateDeal := &types.MsgCreateDeal{
		DataSchema:   []string{"http://jsonld.com"},
		Budget:       budget,
		MaxNumData:   10000,
		BuyerAddress: suite.buyerAccAddr.String(),
	}

	buyer, err := sdk.AccAddressFromBech32(suite.buyerAccAddr.String())
	suite.Require().NoError(err)

	dealID, err := suite.DataDealKeeper.CreateDeal(ctx, buyer, msgCreateDeal)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), dealID)
	suite.dealID = dealID

	suite.BlockPeriod = 6 * time.Second
}

func (suite *abciTestSuite) TestDataVerificationEndBlockerVotePass() {
	ctx := suite.Ctx

	oracle1 := suite.CreateOracleValidator(suite.oraclePubKey, sdk.NewInt(70))
	oracle2 := suite.CreateOracleValidator(suite.oraclePubKey2, sdk.NewInt(20))
	oracle3 := suite.CreateOracleValidator(suite.oraclePubKey3, sdk.NewInt(10))

	dataSale := &types.DataSale{
		SellerAddress: suite.sellerAccAddr.String(),
		DealId:        suite.dealID,
		VerifiableCid: suite.verifiableCID,
		DeliveredCid:  "",
		DataHash:      suite.dataHash,
		Status:        types.DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD,
		VerificationVotingPeriod: &oracletypes.VotingPeriod{
			VotingStartTime: time.Now().Add(-2 * time.Second),
			VotingEndTime:   time.Now().Add(-1 * time.Second),
		},
		DeliveryVotingPeriod:    nil,
		VerificationTallyResult: nil,
		DeliveryTallyResult:     nil,
	}

	deal, err := suite.DataDealKeeper.GetDeal(ctx, suite.dealID)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.SetDataSale(ctx, dataSale)
	suite.Require().NoError(err)

	suite.DataDealKeeper.AddDataVerificationQueue(
		ctx,
		dataSale.DataHash,
		dataSale.DealId,
		dataSale.VerificationVotingPeriod.VotingEndTime,
	)

	vote := types.DataVerificationVote{
		VoterAddress: suite.oracleAddr.String(),
		DealId:       suite.dealID,
		DataHash:     suite.dataHash,
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	vote2 := types.DataVerificationVote{
		VoterAddress: suite.oracleAddr2.String(),
		DealId:       suite.dealID,
		DataHash:     suite.dataHash,
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	vote3 := types.DataVerificationVote{
		VoterAddress: suite.oracleAddr3.String(),
		DealId:       suite.dealID,
		DataHash:     suite.dataHash,
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	err = suite.DataDealKeeper.SetDataVerificationVote(ctx, &vote)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDataVerificationVote(ctx, &vote2)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDataVerificationVote(ctx, &vote3)
	suite.Require().NoError(err)

	votes, err := suite.DataDealKeeper.GetAllDataVerificationVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(3, len(votes))

	// balance check before data verification is passed
	distrModuleAcc := suite.AccountKeeper.GetModuleAccount(ctx, distrtypes.ModuleName).GetAddress()
	distrBalance := suite.BankKeeper.GetBalance(ctx, distrModuleAcc, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.ZeroInt()), distrBalance)

	dealAccAddr, err := sdk.AccAddressFromBech32(deal.GetAddress())
	suite.Require().NoError(err)
	dealBalance := suite.BankKeeper.GetBalance(ctx, dealAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10_000_000_000)), dealBalance)

	sellerBalance := suite.BankKeeper.GetBalance(ctx, suite.sellerAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.ZeroInt()), sellerBalance)

	// End blocker
	datadeal.EndBlocker(ctx, suite.DataDealKeeper)

	// balance check after data verification is passed
	distrBalance = suite.BankKeeper.GetBalance(ctx, distrModuleAcc, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(50_000)), distrBalance) // 0.05 MED

	dealBalance = suite.BankKeeper.GetBalance(ctx, dealAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(9_999_050_000)), dealBalance) // 9,999.05 MED

	sellerBalance = suite.BankKeeper.GetBalance(ctx, suite.sellerAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(900_000)), sellerBalance) // 0.9 MED for 1 data

	// 0.05 MED for oracle rewards for data verification
	// the reward for data verification is distributed to three oracles proportional to their voting power
	// oracle 1 : 0.035 MED
	// oracle 2 : 0.01 MED
	// oracle 3 : 0.05 MED
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(3_500))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, oracle1.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(31_500))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, oracle1.GetOperator()).Rewards)

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(1_000))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, oracle2.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(9_000))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, oracle2.GetOperator()).Rewards)

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(500))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, oracle3.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(4_500))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, oracle3.GetOperator()).Rewards)

	updatedDataSale, err := suite.DataDealKeeper.GetDataSale(ctx, dataSale.DataHash, dataSale.DealId)
	suite.Require().NoError(err)
	suite.Require().Equal(types.DATA_SALE_STATUS_DELIVERY_VOTING_PERIOD, updatedDataSale.Status)
	suite.Require().Equal(suite.dataHash, updatedDataSale.DataHash)

	votes, err = suite.DataDealKeeper.GetAllDataVerificationVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(votes))

	events := ctx.EventManager().Events()

	requiredEvents := map[string]bool{
		types.EventTypeDataDeliveryVote:     false,
		types.EventTypeDataVerificationVote: false,
	}

	for _, e := range events {
		if e.Type == types.EventTypeDataDeliveryVote {
			requiredEvents[e.Type] = true
			suite.Require().Equal(3, len(e.Attributes))
			suite.Require().Equal(types.AttributeKeyVoteStatus, string(e.Attributes[0].Key))
			suite.Require().Equal(types.AttributeValueVoteStatusStarted, string(e.Attributes[0].Value))
			suite.Require().Equal(types.AttributeKeyDataHash, string(e.Attributes[1].Key))
			suite.Require().Equal(types.AttributeKeyDealID, string(e.Attributes[2].Key))
		}

		if e.Type == types.EventTypeDataVerificationVote {
			requiredEvents[e.Type] = true
			suite.Require().Equal(types.AttributeKeyVoteStatus, string(e.Attributes[0].Key))
			suite.Require().Equal(types.AttributeValueVoteStatusEnded, string(e.Attributes[0].Value))
		}
	}

	for _, v := range requiredEvents {
		suite.Require().True(v)
	}
}

func (suite *abciTestSuite) TestDataVerificationEndBlockerVoteReject() {
	ctx := suite.Ctx

	suite.CreateOracleValidator(suite.oraclePubKey, sdk.NewInt(70))
	suite.CreateOracleValidator(suite.oraclePubKey2, sdk.NewInt(20))
	suite.CreateOracleValidator(suite.oraclePubKey3, sdk.NewInt(10))

	dataSale := &types.DataSale{
		SellerAddress: suite.sellerAccAddr.String(),
		DealId:        suite.dealID,
		VerifiableCid: suite.verifiableCID,
		DeliveredCid:  "",
		Status:        types.DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD,
		VerificationVotingPeriod: &oracletypes.VotingPeriod{
			VotingStartTime: time.Now().Add(-2 * time.Second),
			VotingEndTime:   time.Now().Add(-1 * time.Second),
		},
		DeliveryVotingPeriod:    nil,
		VerificationTallyResult: nil,
		DeliveryTallyResult:     nil,
	}

	err := suite.DataDealKeeper.SetDataSale(ctx, dataSale)
	suite.Require().NoError(err)

	suite.DataDealKeeper.AddDataVerificationQueue(
		ctx,
		dataSale.DataHash,
		dataSale.DealId,
		dataSale.VerificationVotingPeriod.VotingEndTime,
	)

	vote := types.DataVerificationVote{
		VoterAddress: suite.oracleAddr.String(),
		DealId:       suite.dealID,
		DataHash:     suite.dataHash,
		VoteOption:   oracletypes.VOTE_OPTION_NO,
	}

	vote2 := types.DataVerificationVote{
		VoterAddress: suite.oracleAddr2.String(),
		DealId:       suite.dealID,
		DataHash:     suite.dataHash,
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	vote3 := types.DataVerificationVote{
		VoterAddress: suite.oracleAddr3.String(),
		DealId:       suite.dealID,
		DataHash:     suite.dataHash,
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	err = suite.DataDealKeeper.SetDataVerificationVote(ctx, &vote)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDataVerificationVote(ctx, &vote2)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDataVerificationVote(ctx, &vote3)
	suite.Require().NoError(err)

	votes, err := suite.DataDealKeeper.GetAllDataVerificationVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(3, len(votes))

	datadeal.EndBlocker(ctx, suite.DataDealKeeper)

	updatedDataSale, err := suite.DataDealKeeper.GetDataSale(ctx, dataSale.DataHash, dataSale.DealId)
	suite.Require().NoError(err)
	suite.Require().Equal(types.DATA_SALE_STATUS_VERIFICATION_FAILED, updatedDataSale.Status)

	votes, err = suite.DataDealKeeper.GetAllDataVerificationVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(votes))

	events := ctx.EventManager().Events()

	requiredEvents := map[string]bool{
		types.EventTypeDataVerificationVote: false,
	}

	for _, e := range events {
		if e.Type == types.EventTypeDataVerificationVote {
			requiredEvents[e.Type] = true
			suite.Require().Equal(types.AttributeKeyVoteStatus, string(e.Attributes[0].Key))
			suite.Require().Equal(types.AttributeValueVoteStatusEnded, string(e.Attributes[0].Value))
			suite.Require().Equal(types.AttributeKeyDataHash, string(e.Attributes[1].Key))
			suite.Require().Equal(types.AttributeKeyDealID, string(e.Attributes[2].Key))
		}
	}

	for _, v := range requiredEvents {
		suite.Require().True(v)
	}
}

func (suite *abciTestSuite) TestDataVerificationEndBlockerVoteRejectSamePower() {
	ctx := suite.Ctx

	suite.CreateOracleValidator(suite.oraclePubKey, sdk.NewInt(10))
	suite.CreateOracleValidator(suite.oraclePubKey2, sdk.NewInt(10))

	dataSale := &types.DataSale{
		SellerAddress: suite.sellerAccAddr.String(),
		DealId:        suite.dealID,
		VerifiableCid: suite.verifiableCID,
		DeliveredCid:  "",
		DataHash:      suite.dataHash,
		Status:        types.DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD,
		VerificationVotingPeriod: &oracletypes.VotingPeriod{
			VotingStartTime: time.Now().Add(-2 * time.Second),
			VotingEndTime:   time.Now().Add(-1 * time.Second),
		},
		DeliveryVotingPeriod:    nil,
		VerificationTallyResult: nil,
		DeliveryTallyResult:     nil,
	}

	err := suite.DataDealKeeper.SetDataSale(ctx, dataSale)
	suite.Require().NoError(err)

	suite.DataDealKeeper.AddDataVerificationQueue(
		ctx,
		dataSale.DataHash,
		dataSale.DealId,
		dataSale.VerificationVotingPeriod.VotingEndTime,
	)

	vote := types.DataVerificationVote{
		VoterAddress: suite.oracleAddr.String(),
		DealId:       suite.dealID,
		DataHash:     suite.dataHash,
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	vote2 := types.DataVerificationVote{
		VoterAddress: suite.oracleAddr2.String(),
		DealId:       suite.dealID,
		DataHash:     suite.dataHash,
		VoteOption:   oracletypes.VOTE_OPTION_NO,
	}

	err = suite.DataDealKeeper.SetDataVerificationVote(ctx, &vote)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDataVerificationVote(ctx, &vote2)
	suite.Require().NoError(err)

	votes, err := suite.DataDealKeeper.GetAllDataVerificationVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(votes))

	datadeal.EndBlocker(ctx, suite.DataDealKeeper)

	updatedDataSale, err := suite.DataDealKeeper.GetDataSale(ctx, dataSale.DataHash, dataSale.DealId)
	suite.Require().NoError(err)
	suite.Require().Equal(types.DATA_SALE_STATUS_VERIFICATION_FAILED, updatedDataSale.Status)

	tallyResult := updatedDataSale.VerificationTallyResult
	suite.Require().Equal(sdk.ZeroInt(), tallyResult.Yes)
	suite.Require().Equal(sdk.NewInt(10), tallyResult.No)
	suite.Require().Equal(1, len(tallyResult.InvalidYes))

	for _, tallyResult := range tallyResult.InvalidYes {
		if bytes.Equal([]byte(vote.DataHash), tallyResult.ConsensusValue) {
			suite.Require().Equal([]byte(vote.DataHash), tallyResult.ConsensusValue)
			suite.Require().Equal(sdk.NewInt(10), tallyResult.VotingAmount)
		} else if bytes.Equal([]byte(vote2.DataHash), tallyResult.ConsensusValue) {
			suite.Require().Equal([]byte(vote2.DataHash), tallyResult.ConsensusValue)
			suite.Require().Equal(sdk.NewInt(10), tallyResult.VotingAmount)
		} else {
			panic(fmt.Sprintf("No matching VerifiableCID(%s) found.", tallyResult.ConsensusValue))
		}
	}

	votes, err = suite.DataDealKeeper.GetAllDataVerificationVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(votes))

	events := ctx.EventManager().Events()

	requiredEvents := map[string]bool{
		types.EventTypeDataVerificationVote: false,
	}

	for _, e := range events {
		if e.Type == types.EventTypeDataVerificationVote {
			requiredEvents[e.Type] = true
			suite.Require().Equal(types.AttributeKeyVoteStatus, string(e.Attributes[0].Key))
			suite.Require().Equal(types.AttributeValueVoteStatusEnded, string(e.Attributes[0].Value))
			suite.Require().Equal(types.AttributeKeyDataHash, string(e.Attributes[1].Key))
			suite.Require().Equal(types.AttributeKeyDealID, string(e.Attributes[2].Key))
		}
	}

	for _, v := range requiredEvents {
		suite.Require().True(v)
	}
}

func (suite *abciTestSuite) TestDataVerificationEndBlockerVoteRejectDealCompleted() {
	ctx := suite.Ctx

	budget := &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(10000000)}

	msgCreateDeal := &types.MsgCreateDeal{
		DataSchema:   []string{"http://jsonld.com"},
		Budget:       budget,
		MaxNumData:   1,
		BuyerAddress: suite.buyerAccAddr.String(),
	}

	buyer, err := sdk.AccAddressFromBech32(msgCreateDeal.BuyerAddress)
	suite.Require().NoError(err)

	dealID, err := suite.DataDealKeeper.CreateDeal(suite.Ctx, buyer, msgCreateDeal)
	suite.Require().NoError(err)

	suite.CreateOracleValidator(suite.oraclePubKey, sdk.NewInt(70))
	suite.CreateOracleValidator(suite.oraclePubKey2, sdk.NewInt(20))
	suite.CreateOracleValidator(suite.oraclePubKey3, sdk.NewInt(10))

	dataSale := &types.DataSale{
		SellerAddress: suite.sellerAccAddr.String(),
		DealId:        dealID,
		VerifiableCid: suite.verifiableCID,
		DeliveredCid:  "",
		DataHash:      suite.dataHash,
		Status:        types.DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD,
		VerificationVotingPeriod: &oracletypes.VotingPeriod{
			VotingStartTime: time.Now().Add(-2 * time.Second),
			VotingEndTime:   time.Now().Add(-1 * time.Second),
		},
		DeliveryVotingPeriod:    nil,
		VerificationTallyResult: nil,
		DeliveryTallyResult:     nil,
	}

	dataSale2 := &types.DataSale{
		SellerAddress: suite.sellerAccAddr.String(),
		DealId:        dealID,
		VerifiableCid: suite.verifiableCID2,
		DeliveredCid:  "",
		DataHash:      suite.dataHash2,
		Status:        types.DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD,
		VerificationVotingPeriod: &oracletypes.VotingPeriod{
			VotingStartTime: time.Now().Add(-2 * time.Second),
			VotingEndTime:   time.Now().Add(-1 * time.Second),
		},
		DeliveryVotingPeriod:    nil,
		VerificationTallyResult: nil,
		DeliveryTallyResult:     nil,
	}

	err = suite.DataDealKeeper.SetDataSale(ctx, dataSale)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.SetDataSale(ctx, dataSale2)
	suite.Require().NoError(err)

	suite.DataDealKeeper.AddDataVerificationQueue(
		ctx,
		dataSale.DataHash,
		dataSale.DealId,
		dataSale.VerificationVotingPeriod.VotingEndTime,
	)

	suite.DataDealKeeper.AddDataVerificationQueue(
		ctx,
		dataSale2.DataHash,
		dataSale2.DealId,
		dataSale2.VerificationVotingPeriod.VotingEndTime,
	)

	vote := types.DataVerificationVote{
		VoterAddress: suite.oracleAddr.String(),
		DealId:       dataSale.DealId,
		DataHash:     suite.dataHash,
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	vote2 := types.DataVerificationVote{
		VoterAddress: suite.oracleAddr2.String(),
		DealId:       dataSale.DealId,
		DataHash:     suite.dataHash,
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	vote3 := types.DataVerificationVote{
		VoterAddress: suite.oracleAddr3.String(),
		DealId:       dataSale.DealId,
		DataHash:     suite.dataHash,
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	vote4 := types.DataVerificationVote{
		VoterAddress: suite.oracleAddr.String(),
		DealId:       dataSale2.DealId,
		DataHash:     suite.dataHash2,
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	vote5 := types.DataVerificationVote{
		VoterAddress: suite.oracleAddr2.String(),
		DealId:       dataSale2.DealId,
		DataHash:     suite.dataHash2,
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	vote6 := types.DataVerificationVote{
		VoterAddress: suite.oracleAddr3.String(),
		DealId:       dataSale2.DealId,
		DataHash:     suite.dataHash2,
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	err = suite.DataDealKeeper.SetDataVerificationVote(ctx, &vote)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDataVerificationVote(ctx, &vote2)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDataVerificationVote(ctx, &vote3)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDataVerificationVote(ctx, &vote4)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDataVerificationVote(ctx, &vote5)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDataVerificationVote(ctx, &vote6)
	suite.Require().NoError(err)

	votes, err := suite.DataDealKeeper.GetAllDataVerificationVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(6, len(votes))

	datadeal.EndBlocker(ctx, suite.DataDealKeeper)

	updatedDataSale, err := suite.DataDealKeeper.GetDataSale(suite.Ctx, suite.dataHash, dealID)
	suite.Require().NoError(err)
	suite.Require().Equal(types.DATA_SALE_STATUS_DELIVERY_VOTING_PERIOD, updatedDataSale.Status)

	updatedDataSale2, err := suite.DataDealKeeper.GetDataSale(suite.Ctx, suite.dataHash2, dealID)
	suite.Require().NoError(err)
	suite.Require().Equal(types.DATA_SALE_STATUS_DEAL_COMPLETED, updatedDataSale2.Status)

	updatedDeal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, dealID)
	suite.Require().NoError(err)
	suite.Require().Equal(types.DEAL_STATUS_COMPLETED, updatedDeal.Status)
	suite.Require().Equal(uint64(1), updatedDeal.CurNumData)
}

func (suite *abciTestSuite) TestDataDeliveryEndBlockerVotePass() {
	ctx := suite.Ctx

	oracle1 := suite.CreateOracleValidator(suite.oraclePubKey, sdk.NewInt(70))
	oracle2 := suite.CreateOracleValidator(suite.oraclePubKey2, sdk.NewInt(20))
	oracle3 := suite.CreateOracleValidator(suite.oraclePubKey3, sdk.NewInt(10))

	dataSale := &types.DataSale{
		SellerAddress:            suite.sellerAccAddr.String(),
		DealId:                   suite.dealID,
		VerifiableCid:            suite.verifiableCID,
		DeliveredCid:             "",
		DataHash:                 suite.dataHash,
		Status:                   types.DATA_SALE_STATUS_DELIVERY_VOTING_PERIOD,
		VerificationVotingPeriod: nil,
		DeliveryVotingPeriod: &oracletypes.VotingPeriod{
			VotingStartTime: time.Now().Add(-2 * time.Second),
			VotingEndTime:   time.Now().Add(-1 * time.Second),
		},
		VerificationTallyResult: nil,
		DeliveryTallyResult:     nil,
	}

	deal, err := suite.DataDealKeeper.GetDeal(ctx, suite.dealID)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.SetDataSale(ctx, dataSale)
	suite.Require().NoError(err)

	suite.DataDealKeeper.AddDataDeliveryQueue(
		ctx,
		dataSale.DataHash,
		dataSale.DealId,
		dataSale.DeliveryVotingPeriod.VotingEndTime,
	)

	vote := types.DataDeliveryVote{
		VoterAddress: suite.oracleAddr.String(),
		DealId:       suite.dealID,
		DataHash:     suite.dataHash,
		DeliveredCid: "deliveredCID",
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	vote2 := types.DataDeliveryVote{
		VoterAddress: suite.oracleAddr2.String(),
		DealId:       suite.dealID,
		DataHash:     suite.dataHash,
		DeliveredCid: "deliveredCID",
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	vote3 := types.DataDeliveryVote{
		VoterAddress: suite.oracleAddr3.String(),
		DealId:       suite.dealID,
		DataHash:     suite.dataHash,
		DeliveredCid: "deliveredCID",
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	err = suite.DataDealKeeper.SetDataDeliveryVote(ctx, &vote)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDataDeliveryVote(ctx, &vote2)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDataDeliveryVote(ctx, &vote3)
	suite.Require().NoError(err)

	votes, err := suite.DataDealKeeper.GetAllDataDeliveryVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(3, len(votes))

	// balance check before data verification is passed
	distrModuleAcc := suite.AccountKeeper.GetModuleAccount(ctx, distrtypes.ModuleName).GetAddress()
	distrBalance := suite.BankKeeper.GetBalance(ctx, distrModuleAcc, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.ZeroInt()), distrBalance)

	dealAccAddr, err := sdk.AccAddressFromBech32(deal.GetAddress())
	suite.Require().NoError(err)
	dealBalance := suite.BankKeeper.GetBalance(ctx, dealAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10_000_000_000)), dealBalance)

	// End blocker
	datadeal.EndBlocker(ctx, suite.DataDealKeeper)

	// balance check after data verification is passed
	distrBalance = suite.BankKeeper.GetBalance(ctx, distrModuleAcc, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(50_000)), distrBalance) // 0.05 MED

	dealBalance = suite.BankKeeper.GetBalance(ctx, dealAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(9_999_950_000)), dealBalance) // 9,999.05 MED

	// 0.05 MED for oracle rewards for data verification
	// the reward for data verification is distributed to three oracles proportional to their voting power
	// oracle 1 : 0.035 MED
	// oracle 2 : 0.01 MED
	// oracle 3 : 0.05 MED
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(3_500))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, oracle1.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(31_500))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, oracle1.GetOperator()).Rewards)

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(1_000))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, oracle2.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(9_000))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, oracle2.GetOperator()).Rewards)

	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(500))), suite.DistrKeeper.GetValidatorAccumulatedCommission(ctx, oracle3.GetOperator()).Commission)
	suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin(assets.MicroMedDenom, sdk.NewInt(4_500))), suite.DistrKeeper.GetValidatorCurrentRewards(ctx, oracle3.GetOperator()).Rewards)

	updatedDataSale, err := suite.DataDealKeeper.GetDataSale(ctx, dataSale.DataHash, dataSale.DealId)
	suite.Require().NoError(err)
	suite.Require().Equal(types.DATA_SALE_STATUS_COMPLETED, updatedDataSale.Status)
	suite.Require().Equal("deliveredCID", updatedDataSale.DeliveredCid)

	votes, err = suite.DataDealKeeper.GetAllDataDeliveryVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(votes))

	events := ctx.EventManager().Events()

	requiredEvents := map[string]bool{
		types.EventTypeDataDeliveryVote: false,
	}

	for _, e := range events {
		if e.Type == types.EventTypeDataDeliveryVote {
			requiredEvents[e.Type] = true
			suite.Require().Len(e.Attributes, 3)
			suite.Require().Equal(types.AttributeKeyVoteStatus, string(e.Attributes[0].Key))
			suite.Require().Equal(types.AttributeValueVoteStatusEnded, string(e.Attributes[0].Value))
			suite.Require().Equal(types.AttributeKeyDeliveredCID, string(e.Attributes[1].Key))
			suite.Require().Equal(types.AttributeKeyDealID, string(e.Attributes[2].Key))
		}
	}

	for _, v := range requiredEvents {
		suite.Require().True(v)
	}
}

func (suite *abciTestSuite) TestDataDeliveryEndBlockerVoteReject() {
	ctx := suite.Ctx

	suite.CreateOracleValidator(suite.oraclePubKey, sdk.NewInt(70))
	suite.CreateOracleValidator(suite.oraclePubKey2, sdk.NewInt(20))
	suite.CreateOracleValidator(suite.oraclePubKey3, sdk.NewInt(10))

	dataSale := &types.DataSale{
		SellerAddress:            suite.sellerAccAddr.String(),
		DealId:                   suite.dealID,
		VerifiableCid:            suite.verifiableCID,
		DeliveredCid:             "",
		DataHash:                 suite.dataHash,
		Status:                   types.DATA_SALE_STATUS_DELIVERY_VOTING_PERIOD,
		VerificationVotingPeriod: nil,
		DeliveryVotingPeriod: &oracletypes.VotingPeriod{
			VotingStartTime: time.Now().Add(-2 * time.Second),
			VotingEndTime:   time.Now().Add(-1 * time.Second),
		},
		VerificationTallyResult: nil,
		DeliveryTallyResult:     nil,
	}

	err := suite.DataDealKeeper.SetDataSale(ctx, dataSale)
	suite.Require().NoError(err)

	suite.DataDealKeeper.AddDataDeliveryQueue(
		ctx,
		dataSale.DataHash,
		dataSale.DealId,
		dataSale.DeliveryVotingPeriod.VotingEndTime,
	)

	vote := types.DataDeliveryVote{
		VoterAddress: suite.oracleAddr.String(),
		DealId:       suite.dealID,
		DataHash:     suite.dataHash,
		DeliveredCid: "deliveredCID",
		VoteOption:   oracletypes.VOTE_OPTION_NO,
	}

	vote2 := types.DataDeliveryVote{
		VoterAddress: suite.oracleAddr2.String(),
		DealId:       suite.dealID,
		DataHash:     suite.dataHash,
		DeliveredCid: "deliveredCID",
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	vote3 := types.DataDeliveryVote{
		VoterAddress: suite.oracleAddr3.String(),
		DealId:       suite.dealID,
		DataHash:     suite.dataHash,
		DeliveredCid: "deliveredCID",
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	err = suite.DataDealKeeper.SetDataDeliveryVote(ctx, &vote)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDataDeliveryVote(ctx, &vote2)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDataDeliveryVote(ctx, &vote3)
	suite.Require().NoError(err)

	votes, err := suite.DataDealKeeper.GetAllDataDeliveryVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(3, len(votes))

	datadeal.EndBlocker(ctx, suite.DataDealKeeper)

	updatedDataSale, err := suite.DataDealKeeper.GetDataSale(ctx, dataSale.DataHash, dataSale.DealId)
	suite.Require().NoError(err)
	suite.Require().Equal(types.DATA_SALE_STATUS_DELIVERY_FAILED, updatedDataSale.Status)

	votes, err = suite.DataDealKeeper.GetAllDataDeliveryVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(votes))

	events := ctx.EventManager().Events()

	requiredEvents := map[string]bool{
		types.EventTypeDataDeliveryVote: false,
	}

	for _, e := range events {
		if e.Type == types.EventTypeDataDeliveryVote {
			requiredEvents[e.Type] = true
			suite.Require().Equal(types.AttributeKeyVoteStatus, string(e.Attributes[0].Key))
			suite.Require().Equal(types.AttributeValueVoteStatusEnded, string(e.Attributes[0].Value))
			suite.Require().Equal(types.AttributeKeyDeliveredCID, string(e.Attributes[1].Key))
			suite.Require().Equal(types.AttributeKeyDealID, string(e.Attributes[2].Key))
		}
	}

	for _, v := range requiredEvents {
		suite.Require().True(v)
	}
}

func (suite *abciTestSuite) TestDataDeliveryEndBlockerVoteRejectSamePower() {
	ctx := suite.Ctx

	suite.CreateOracleValidator(suite.oraclePubKey, sdk.NewInt(10))
	suite.CreateOracleValidator(suite.oraclePubKey2, sdk.NewInt(10))

	dataSale := &types.DataSale{
		SellerAddress:            suite.sellerAccAddr.String(),
		DealId:                   suite.dealID,
		VerifiableCid:            suite.verifiableCID,
		DeliveredCid:             "",
		DataHash:                 suite.dataHash,
		Status:                   types.DATA_SALE_STATUS_DELIVERY_VOTING_PERIOD,
		VerificationVotingPeriod: nil,
		DeliveryVotingPeriod: &oracletypes.VotingPeriod{
			VotingStartTime: time.Now().Add(-2 * time.Second),
			VotingEndTime:   time.Now().Add(-1 * time.Second),
		},
		VerificationTallyResult: nil,
		DeliveryTallyResult:     nil,
	}

	err := suite.DataDealKeeper.SetDataSale(ctx, dataSale)
	suite.Require().NoError(err)

	suite.DataDealKeeper.AddDataDeliveryQueue(
		ctx,
		dataSale.DataHash,
		dataSale.DealId,
		dataSale.DeliveryVotingPeriod.VotingEndTime,
	)

	vote := types.DataDeliveryVote{
		VoterAddress: suite.oracleAddr.String(),
		DealId:       1,
		DataHash:     suite.dataHash,
		DeliveredCid: "deliveredCID1",
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	vote2 := types.DataDeliveryVote{
		VoterAddress: suite.oracleAddr2.String(),
		DealId:       1,
		DataHash:     suite.dataHash,
		DeliveredCid: "deliveredCID2",
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	err = suite.DataDealKeeper.SetDataDeliveryVote(ctx, &vote)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDataDeliveryVote(ctx, &vote2)
	suite.Require().NoError(err)

	votes, err := suite.DataDealKeeper.GetAllDataDeliveryVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(votes))

	datadeal.EndBlocker(ctx, suite.DataDealKeeper)

	updatedDataSale, err := suite.DataDealKeeper.GetDataSale(ctx, dataSale.DataHash, dataSale.DealId)
	suite.Require().NoError(err)
	suite.Require().Equal(types.DATA_SALE_STATUS_DELIVERY_FAILED, updatedDataSale.Status)
	tallyResult := updatedDataSale.DeliveryTallyResult
	suite.Require().Equal(sdk.ZeroInt(), tallyResult.Yes)
	suite.Require().Equal(sdk.ZeroInt(), tallyResult.No)
	suite.Require().Equal(2, len(tallyResult.InvalidYes))
	for _, tallyResult := range tallyResult.InvalidYes {
		if bytes.Equal([]byte(vote.DeliveredCid), tallyResult.ConsensusValue) {
			suite.Require().Equal([]byte(vote.DeliveredCid), tallyResult.ConsensusValue)
			suite.Require().Equal(sdk.NewInt(10), tallyResult.VotingAmount)
		} else if bytes.Equal([]byte(vote2.DeliveredCid), tallyResult.ConsensusValue) {
			suite.Require().Equal([]byte(vote2.DeliveredCid), tallyResult.ConsensusValue)
			suite.Require().Equal(sdk.NewInt(10), tallyResult.VotingAmount)
		} else {
			panic(fmt.Sprintf("No matching DeliveredCid(%s) found.", tallyResult.ConsensusValue))
		}
	}

	votes, err = suite.DataDealKeeper.GetAllDataDeliveryVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(votes))

	events := ctx.EventManager().Events()

	requiredEvents := map[string]bool{
		types.EventTypeDataDeliveryVote: false,
	}

	for _, e := range events {
		if e.Type == types.EventTypeDataDeliveryVote {
			requiredEvents[types.EventTypeDataDeliveryVote] = true
			suite.Require().Equal(types.AttributeKeyVoteStatus, string(e.Attributes[0].Key))
			suite.Require().Equal(types.AttributeValueVoteStatusEnded, string(e.Attributes[0].Value))
			suite.Require().Equal(types.AttributeKeyDeliveredCID, string(e.Attributes[1].Key))
			suite.Require().Equal(types.AttributeKeyDealID, string(e.Attributes[2].Key))
		}
	}

	for _, v := range requiredEvents {
		suite.Require().True(v)
	}
}

func (suite *abciTestSuite) TestDealDeactivateEndBlockerPass() {
	ctx := suite.Ctx

	msgDeactivateDeal := &types.MsgDeactivateDeal{
		DealId:           suite.dealID,
		RequesterAddress: suite.buyerAccAddr.String(),
	}

	err := suite.DataDealKeeper.RequestDeactivateDeal(ctx, msgDeactivateDeal)
	suite.Require().NoError(err)

	// start block height of deactivation
	datadeal.EndBlocker(ctx, suite.DataDealKeeper)
	getDeal, err := suite.DataDealKeeper.GetDeal(ctx, suite.dealID)
	suite.Require().NoError(err)
	suite.Require().Equal(types.DEAL_STATUS_DEACTIVATING, getDeal.Status)

	msgSellData := &types.MsgSellData{
		DealId:        suite.dealID,
		VerifiableCid: suite.verifiableCID,
		DataHash:      suite.dataHash,
		SellerAddress: suite.sellerAccAddr.String(),
	}
	err = suite.DataDealKeeper.SellData(ctx, msgSellData)
	suite.Require().ErrorContains(err, "deal status is not ACTIVE")

	// block height reached deactivationHeight
	oracleParams := suite.OracleKeeper.GetParams(ctx)
	VotingPeriod := oracleParams.VoteParams.VotingPeriod
	datadealParams := suite.DataDealKeeper.GetParams(ctx)
	dealDeactivationParam := datadealParams.DealDeactivationParam

	deactivationHeight := ctx.BlockHeader().Height + dealDeactivationParam*int64(VotingPeriod/suite.BlockPeriod) + 1

	ctx = ctx.WithBlockHeight(deactivationHeight)

	datadeal.EndBlocker(ctx, suite.DataDealKeeper)

	getDeal, err = suite.DataDealKeeper.GetDeal(ctx, suite.dealID)
	suite.Require().NoError(err)
	suite.Require().Equal(types.DEAL_STATUS_DEACTIVATED, getDeal.Status)

}
