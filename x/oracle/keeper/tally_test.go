package keeper_test

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/testutil"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
)

type tallyTestSuite struct {
	testutil.OracleBaseTestSuite

	uniqueID string

	oracleAccPubKey  cryptotypes.PubKey
	oracleAccAddr    sdk.AccAddress
	oracleToken      sdk.Int
	oracleAccPubKey2 cryptotypes.PubKey
	oracleAccAddr2   sdk.AccAddress
	oracleToken2     sdk.Int

	newOracleAccPubKey cryptotypes.PubKey
	newOracleAccAddr   sdk.AccAddress
	newToken           sdk.Int

	oraclePrivKey *btcec.PrivateKey
	oraclePubKey  *btcec.PublicKey
}

func TestTallyTestSuite(t *testing.T) {
	suite.Run(t, new(tallyTestSuite))
}

func (suite *tallyTestSuite) BeforeTest(_, _ string) {
	ctx := suite.Ctx

	suite.uniqueID = "correctUniqueID"

	suite.oracleAccPubKey = secp256k1.GenPrivKey().PubKey()
	suite.oracleAccAddr = sdk.AccAddress(suite.oracleAccPubKey.Address())
	suite.oracleToken = sdk.NewInt(70)
	suite.oracleAccPubKey2 = secp256k1.GenPrivKey().PubKey()
	suite.oracleAccAddr2 = sdk.AccAddress(suite.oracleAccPubKey2.Address())
	suite.oracleToken2 = sdk.NewInt(20)

	suite.newOracleAccPubKey = secp256k1.GenPrivKey().PubKey()
	suite.newOracleAccAddr = sdk.AccAddress(suite.newOracleAccPubKey.Address())
	suite.newToken = sdk.NewInt(10)

	suite.oraclePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.oraclePubKey = suite.oraclePrivKey.PubKey()

	suite.OracleKeeper.SetParams(ctx, types.Params{
		OraclePublicKey:          base64.StdEncoding.EncodeToString(suite.oraclePubKey.SerializeCompressed()),
		OraclePubKeyRemoteReport: "",
		UniqueId:                 suite.uniqueID,
		VoteParams: types.VoteParams{
			VotingPeriod: 100,
			JailPeriod:   60,
			Threshold:    sdk.NewDecWithPrec(2, 3),
		},
		SlashParams: types.SlashParams{
			SlashFractionDowntime: sdk.NewDecWithPrec(3, 1),
			SlashFractionForgery:  sdk.NewDecWithPrec(1, 1),
		},
	})
}

func (suite *tallyTestSuite) TestTally() {
	ctx := suite.Ctx

	oracleAccAddr := suite.oracleAccAddr
	oracleAccAddr2 := suite.oracleAccAddr2
	suite.CreateOracleValidator(suite.oracleAccPubKey, suite.oracleToken)
	suite.CreateOracleValidator(suite.oracleAccPubKey2, suite.oracleToken2)

	newOracleAccAddr := suite.newOracleAccAddr
	suite.SetAccount(suite.newOracleAccPubKey)
	suite.SetValidator(suite.newOracleAccPubKey, suite.newToken)

	nodePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	suite.Require().NoError(err)

	oracleRegistration := &types.OracleRegistration{
		UniqueId:               suite.uniqueID,
		Address:                newOracleAccAddr.String(),
		NodePubKey:             nodePrivKey.PubKey().SerializeCompressed(),
		NodePubKeyRemoteReport: []byte("nodePubKey"),
		TrustedBlockHeight:     1,
		TrustedBlockHash:       []byte("Hash"),
		Status:                 types.ORACLE_REGISTRATION_STATUS_VOTING_PERIOD,
		VotingPeriod: &types.VotingPeriod{
			VotingStartTime: time.Now(),
			VotingEndTime:   time.Now(),
		},
	}
	err = suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	consensusValue := []byte("encPriv1")
	vote := &types.OracleRegistrationVote{
		UniqueId:               suite.uniqueID,
		VoterAddress:           oracleAccAddr.String(),
		VotingTargetAddress:    newOracleAccAddr.String(),
		VoteOption:             types.VOTE_OPTION_YES,
		EncryptedOraclePrivKey: consensusValue,
	}
	err = suite.OracleKeeper.SetOracleRegistrationVote(suite.Ctx, vote)
	suite.Require().NoError(err)

	vote2 := &types.OracleRegistrationVote{
		UniqueId:               suite.uniqueID,
		VoterAddress:           oracleAccAddr2.String(),
		VotingTargetAddress:    newOracleAccAddr.String(),
		VoteOption:             types.VOTE_OPTION_YES,
		EncryptedOraclePrivKey: consensusValue,
	}
	err = suite.OracleKeeper.SetOracleRegistrationVote(suite.Ctx, vote2)
	suite.Require().NoError(err)

	iter := suite.OracleKeeper.GetOracleRegistrationVoteIterator(suite.Ctx, suite.uniqueID, newOracleAccAddr.String())
	tallyResult, err := suite.GetTallyKeeper().Tally(
		suite.Ctx,
		iter,
		func() types.Vote {
			return &types.OracleRegistrationVote{}
		},
		func(vote types.Vote) error {
			return suite.OracleKeeper.RemoveOracleRegistrationVote(suite.Ctx, vote.(*types.OracleRegistrationVote))
		},
	)
	suite.Require().NoError(err)

	suite.Require().Equal(sdk.NewInt(90), tallyResult.Yes)
	suite.Require().Equal(sdk.ZeroInt(), tallyResult.No)
	suite.Require().Equal(0, len(tallyResult.InvalidYes))
	suite.Require().Equal(sdk.NewInt(90), tallyResult.Total)
	suite.Require().Equal(consensusValue, tallyResult.ConsensusValue)
}

func (suite *tallyTestSuite) TestTallyOracleJailed() {
	ctx := suite.Ctx

	oracleAccAddr := suite.oracleAccAddr
	oracleAccAddr2 := suite.oracleAccAddr2
	suite.CreateOracleValidator(suite.oracleAccPubKey, suite.oracleToken)
	suite.CreateOracleValidator(suite.oracleAccPubKey2, suite.oracleToken2)

	suite.StakingKeeper.Jail(ctx, oracleAccAddr2.Bytes())

	val, ok := suite.StakingKeeper.GetValidator(ctx, oracleAccAddr2.Bytes())
	suite.Require().True(ok)
	suite.Require().True(val.Jailed)

	newOracleAccAddr := suite.newOracleAccAddr
	suite.SetAccount(suite.newOracleAccPubKey)
	suite.SetValidator(suite.newOracleAccPubKey, suite.newToken)

	nodePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	suite.Require().NoError(err)

	oracleRegistration := &types.OracleRegistration{
		UniqueId:               suite.uniqueID,
		Address:                newOracleAccAddr.String(),
		NodePubKey:             nodePrivKey.PubKey().SerializeCompressed(),
		NodePubKeyRemoteReport: []byte("nodePubKey"),
		TrustedBlockHeight:     1,
		TrustedBlockHash:       []byte("Hash"),
		Status:                 types.ORACLE_REGISTRATION_STATUS_VOTING_PERIOD,
		VotingPeriod: &types.VotingPeriod{
			VotingStartTime: time.Now(),
			VotingEndTime:   time.Now(),
		},
	}
	err = suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	consensusValue := []byte("encPriv1")
	vote := &types.OracleRegistrationVote{
		UniqueId:               suite.uniqueID,
		VoterAddress:           oracleAccAddr.String(),
		VotingTargetAddress:    newOracleAccAddr.String(),
		VoteOption:             types.VOTE_OPTION_YES,
		EncryptedOraclePrivKey: consensusValue,
	}
	err = suite.OracleKeeper.SetOracleRegistrationVote(suite.Ctx, vote)
	suite.Require().NoError(err)

	vote2 := &types.OracleRegistrationVote{
		UniqueId:               suite.uniqueID,
		VoterAddress:           oracleAccAddr2.String(),
		VotingTargetAddress:    newOracleAccAddr.String(),
		VoteOption:             types.VOTE_OPTION_YES,
		EncryptedOraclePrivKey: consensusValue,
	}
	err = suite.OracleKeeper.SetOracleRegistrationVote(suite.Ctx, vote2)
	suite.Require().NoError(err)

	oracleVotes, err := suite.OracleKeeper.GetAllOracleRegistrationVoteList(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(oracleVotes))

	iter := suite.OracleKeeper.GetOracleRegistrationVoteIterator(suite.Ctx, suite.uniqueID, newOracleAccAddr.String())
	tallyResult, err := suite.GetTallyKeeper().Tally(
		suite.Ctx,
		iter,
		func() types.Vote {
			return &types.OracleRegistrationVote{}
		},
		func(vote types.Vote) error {
			return suite.OracleKeeper.RemoveOracleRegistrationVote(suite.Ctx, vote.(*types.OracleRegistrationVote))
		},
	)
	suite.Require().NoError(err)

	suite.Require().Equal(suite.oracleToken, tallyResult.Yes)
	suite.Require().Equal(sdk.ZeroInt(), tallyResult.No)
	suite.Require().Equal(0, len(tallyResult.InvalidYes))
	// not include oracle2. because oracle2 is jailed.
	suite.Require().Equal(sdk.NewInt(70), tallyResult.Total)
	suite.Require().Equal(consensusValue, tallyResult.ConsensusValue)

	oracleVotes, err = suite.OracleKeeper.GetAllOracleRegistrationVoteList(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(oracleVotes))
}
