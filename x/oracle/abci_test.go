package oracle_test

import (
	"bytes"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle"
	"github.com/medibloc/panacea-core/v2/x/oracle/testutil"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type abciTestSuite struct {
	testutil.OracleBaseTestSuite

	uniqueID string

	oraclePubKey  cryptotypes.PubKey
	oracleAddr    sdk.AccAddress
	oraclePubKey2 cryptotypes.PubKey
	oracleAddr2   sdk.AccAddress
	oraclePubKey3 cryptotypes.PubKey
	oracleAddr3   sdk.AccAddress

	newOraclePubKey cryptotypes.PubKey
	newOracleAddr   sdk.AccAddress
}

func TestAbciTestSuite(t *testing.T) {
	suite.Run(t, new(abciTestSuite))
}

func (suite *abciTestSuite) BeforeTest(_, _ string) {
	ctx := suite.Ctx
	suite.uniqueID = "uniqueID"

	suite.oraclePubKey = secp256k1.GenPrivKey().PubKey()
	suite.oracleAddr = sdk.AccAddress(suite.oraclePubKey.Address())
	suite.oraclePubKey2 = secp256k1.GenPrivKey().PubKey()
	suite.oracleAddr2 = sdk.AccAddress(suite.oraclePubKey2.Address())
	suite.oraclePubKey3 = secp256k1.GenPrivKey().PubKey()
	suite.oracleAddr3 = sdk.AccAddress(suite.oraclePubKey3.Address())
	suite.newOraclePubKey = secp256k1.GenPrivKey().PubKey()
	suite.newOracleAddr = sdk.AccAddress(suite.newOraclePubKey.Address())

	oraclePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	suite.Require().NoError(err)
	suite.OracleKeeper.SetParams(ctx, types.Params{
		OraclePublicKey:          oraclePrivKey.PubKey().SerializeCompressed(),
		OraclePubKeyRemoteReport: []byte("oraclePubKeyRemoteReport"),
		UniqueId:                 uniqueID,
		VoteParams: types.VoteParams{
			VotingPeriod: 100,
			JailPeriod:   60,
			Threshold:    sdk.NewDec(2).Quo(sdk.NewDec(3)),
		},
		SlashParams: types.SlashParams{
			SlashFractionDowntime: sdk.NewDecWithPrec(3, 1),
			SlashFractionForgery:  sdk.NewDecWithPrec(1, 1),
		},
	})
}

func (suite abciTestSuite) TestEndBlockerVotePass() {
	ctx := suite.Ctx

	suite.CreateOracleValidator(suite.oraclePubKey, sdk.NewInt(70))
	suite.CreateOracleValidator(suite.oraclePubKey2, sdk.NewInt(20))
	suite.CreateOracleValidator(suite.oraclePubKey3, sdk.NewInt(10))

	nodePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	suite.Require().NoError(err)
	oracleRegistration := &types.OracleRegistration{
		UniqueId:               suite.uniqueID,
		Address:                suite.newOracleAddr.String(),
		NodePubKey:             nodePrivKey.PubKey().SerializeCompressed(),
		NodePubKeyRemoteReport: []byte("nodePubKeyRemoteReport"),
		TrustedBlockHeight:     10,
		TrustedBlockHash:       []byte("trustedBlockHash"),
		Status:                 types.ORACLE_REGISTRATION_STATUS_VOTING_PERIOD,
		VotingPeriod: &types.VotingPeriod{
			VotingStartTime: time.Now().Add(-2 * time.Second),
			VotingEndTime:   time.Now().Add(-1 * time.Second),
		},
	}
	err = suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	suite.OracleKeeper.AddOracleRegistrationQueue(
		ctx,
		suite.uniqueID,
		suite.newOracleAddr,
		oracleRegistration.VotingPeriod.VotingEndTime,
	)

	vote := types.OracleRegistrationVote{
		UniqueId:               suite.uniqueID,
		VoterAddress:           suite.oracleAddr.String(),
		VotingTargetAddress:    suite.newOracleAddr.String(),
		VoteOption:             types.VOTE_OPTION_YES,
		EncryptedOraclePrivKey: []byte("encryptedOraclePrivKey"),
	}
	vote2 := types.OracleRegistrationVote{
		UniqueId:               suite.uniqueID,
		VoterAddress:           suite.oracleAddr2.String(),
		VotingTargetAddress:    suite.newOracleAddr.String(),
		VoteOption:             types.VOTE_OPTION_YES,
		EncryptedOraclePrivKey: []byte("encryptedOraclePrivKey"),
	}
	vote3 := types.OracleRegistrationVote{
		UniqueId:               suite.uniqueID,
		VoterAddress:           suite.oracleAddr3.String(),
		VotingTargetAddress:    suite.newOracleAddr.String(),
		VoteOption:             types.VOTE_OPTION_YES,
		EncryptedOraclePrivKey: []byte("encryptedOraclePrivKey"),
	}
	err = suite.OracleKeeper.SetOracleRegistrationVote(ctx, &vote)
	suite.Require().NoError(err)
	err = suite.OracleKeeper.SetOracleRegistrationVote(ctx, &vote2)
	suite.Require().NoError(err)
	err = suite.OracleKeeper.SetOracleRegistrationVote(ctx, &vote3)
	suite.Require().NoError(err)

	oracleVotes, err := suite.OracleKeeper.GetAllOracleRegistrationVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(3, len(oracleVotes))

	oracle.EndBlocker(suite.Ctx, suite.OracleKeeper)

	oracleRegistration, err = suite.OracleKeeper.GetOracleRegistration(ctx, suite.uniqueID, suite.newOracleAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(types.ORACLE_REGISTRATION_STATUS_PASSED, oracleRegistration.Status)

	newOracle, err := suite.OracleKeeper.GetOracle(ctx, suite.newOracleAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(suite.newOracleAddr.String(), newOracle.Address)
	suite.Require().Equal(types.ORACLE_STATUS_ACTIVE, newOracle.Status)

	oracleVotes, err = suite.OracleKeeper.GetAllOracleRegistrationVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(oracleVotes))

	events := ctx.EventManager().Events()
	suite.Require().Equal(1, len(events))
	suite.Require().Equal(types.EventTypeRegistrationVote, events[0].Type)
	eventAttributes := events[0].Attributes
	suite.Require().Equal(2, len(eventAttributes))
	suite.Require().Equal(types.AttributeKeyVoteStatus, string(eventAttributes[0].Key))
	suite.Require().Equal(types.AttributeValueVoteStatusEnded, string(eventAttributes[0].Value))
	suite.Require().Equal(types.AttributeKeyOracleAddress, string(eventAttributes[1].Key))
	suite.Require().Equal(oracleRegistration.Address, string(eventAttributes[1].Value))
}

func (suite abciTestSuite) TestEndBlockerVoteReject() {
	ctx := suite.Ctx

	suite.CreateOracleValidator(suite.oraclePubKey, sdk.NewInt(70))
	suite.CreateOracleValidator(suite.oraclePubKey2, sdk.NewInt(20))
	suite.CreateOracleValidator(suite.oraclePubKey3, sdk.NewInt(10))

	nodePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	suite.Require().NoError(err)
	oracleRegistration := &types.OracleRegistration{
		UniqueId:               suite.uniqueID,
		Address:                suite.newOracleAddr.String(),
		NodePubKey:             nodePrivKey.PubKey().SerializeCompressed(),
		NodePubKeyRemoteReport: []byte("nodePubKeyRemoteReport"),
		TrustedBlockHeight:     10,
		TrustedBlockHash:       []byte("trustedBlockHash"),
		Status:                 types.ORACLE_REGISTRATION_STATUS_VOTING_PERIOD,
		VotingPeriod: &types.VotingPeriod{
			VotingStartTime: time.Now().Add(-2 * time.Second),
			VotingEndTime:   time.Now().Add(-1 * time.Second),
		},
	}
	err = suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	suite.OracleKeeper.AddOracleRegistrationQueue(
		ctx,
		suite.uniqueID,
		suite.newOracleAddr,
		oracleRegistration.VotingPeriod.VotingEndTime,
	)

	vote := types.OracleRegistrationVote{
		UniqueId:            suite.uniqueID,
		VoterAddress:        suite.oracleAddr.String(),
		VotingTargetAddress: suite.newOracleAddr.String(),
		VoteOption:          types.VOTE_OPTION_NO,
	}
	vote2 := types.OracleRegistrationVote{
		UniqueId:               suite.uniqueID,
		VoterAddress:           suite.oracleAddr2.String(),
		VotingTargetAddress:    suite.newOracleAddr.String(),
		VoteOption:             types.VOTE_OPTION_YES,
		EncryptedOraclePrivKey: []byte("encryptedOraclePrivKey"),
	}
	vote3 := types.OracleRegistrationVote{
		UniqueId:               suite.uniqueID,
		VoterAddress:           suite.oracleAddr3.String(),
		VotingTargetAddress:    suite.newOracleAddr.String(),
		VoteOption:             types.VOTE_OPTION_YES,
		EncryptedOraclePrivKey: []byte("encryptedOraclePrivKey"),
	}
	err = suite.OracleKeeper.SetOracleRegistrationVote(ctx, &vote)
	suite.Require().NoError(err)
	err = suite.OracleKeeper.SetOracleRegistrationVote(ctx, &vote2)
	suite.Require().NoError(err)
	err = suite.OracleKeeper.SetOracleRegistrationVote(ctx, &vote3)
	suite.Require().NoError(err)

	oracleVotes, err := suite.OracleKeeper.GetAllOracleRegistrationVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(3, len(oracleVotes))

	oracle.EndBlocker(suite.Ctx, suite.OracleKeeper)

	oracleRegistration, err = suite.OracleKeeper.GetOracleRegistration(ctx, suite.uniqueID, suite.newOracleAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(types.ORACLE_REGISTRATION_STATUS_REJECTED, oracleRegistration.Status)

	_, err = suite.OracleKeeper.GetOracle(ctx, suite.newOracleAddr.String())
	suite.Require().Error(err)

	oracleVotes, err = suite.OracleKeeper.GetAllOracleRegistrationVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(oracleVotes))

	events := ctx.EventManager().Events()
	suite.Require().Equal(1, len(events))
	suite.Require().Equal(types.EventTypeRegistrationVote, events[0].Type)
	eventAttributes := events[0].Attributes
	suite.Require().Equal(2, len(eventAttributes))
	suite.Require().Equal(types.AttributeKeyVoteStatus, string(eventAttributes[0].Key))
	suite.Require().Equal(types.AttributeValueVoteStatusEnded, string(eventAttributes[0].Value))
	suite.Require().Equal(types.AttributeKeyOracleAddress, string(eventAttributes[1].Key))
	suite.Require().Equal(oracleRegistration.Address, string(eventAttributes[1].Value))
}

func (suite abciTestSuite) TestEndBlockerVoteRejectSamePower() {
	ctx := suite.Ctx

	suite.CreateOracleValidator(suite.oraclePubKey, sdk.NewInt(10))
	suite.CreateOracleValidator(suite.oraclePubKey2, sdk.NewInt(10))

	nodePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	suite.Require().NoError(err)
	oracleRegistration := &types.OracleRegistration{
		UniqueId:               suite.uniqueID,
		Address:                suite.newOracleAddr.String(),
		NodePubKey:             nodePrivKey.PubKey().SerializeCompressed(),
		NodePubKeyRemoteReport: []byte("nodePubKeyRemoteReport"),
		TrustedBlockHeight:     10,
		TrustedBlockHash:       []byte("trustedBlockHash"),
		Status:                 types.ORACLE_REGISTRATION_STATUS_VOTING_PERIOD,
		VotingPeriod: &types.VotingPeriod{
			VotingStartTime: time.Now().Add(-2 * time.Second),
			VotingEndTime:   time.Now().Add(-1 * time.Second),
		},
	}
	err = suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	suite.OracleKeeper.AddOracleRegistrationQueue(
		ctx,
		suite.uniqueID,
		suite.newOracleAddr,
		oracleRegistration.VotingPeriod.VotingEndTime,
	)

	vote := types.OracleRegistrationVote{
		UniqueId:               suite.uniqueID,
		VoterAddress:           suite.oracleAddr.String(),
		VotingTargetAddress:    suite.newOracleAddr.String(),
		VoteOption:             types.VOTE_OPTION_YES,
		EncryptedOraclePrivKey: []byte("encryptedOraclePrivKey1"),
	}
	vote2 := types.OracleRegistrationVote{
		UniqueId:               suite.uniqueID,
		VoterAddress:           suite.oracleAddr2.String(),
		VotingTargetAddress:    suite.newOracleAddr.String(),
		VoteOption:             types.VOTE_OPTION_YES,
		EncryptedOraclePrivKey: []byte("encryptedOraclePrivKey2"),
	}
	err = suite.OracleKeeper.SetOracleRegistrationVote(ctx, &vote)
	suite.Require().NoError(err)
	err = suite.OracleKeeper.SetOracleRegistrationVote(ctx, &vote2)
	suite.Require().NoError(err)

	oracleVotes, err := suite.OracleKeeper.GetAllOracleRegistrationVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(oracleVotes))

	oracle.EndBlocker(suite.Ctx, suite.OracleKeeper)

	oracleRegistration, err = suite.OracleKeeper.GetOracleRegistration(ctx, suite.uniqueID, suite.newOracleAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(types.ORACLE_REGISTRATION_STATUS_REJECTED, oracleRegistration.Status)
	tallyResult := oracleRegistration.TallyResult
	suite.Require().Equal(sdk.ZeroInt(), tallyResult.Yes)
	suite.Require().Equal(sdk.ZeroInt(), tallyResult.No)
	suite.Require().Equal(2, len(tallyResult.InvalidYes))
	for _, tallyResult := range tallyResult.InvalidYes {
		if bytes.Equal(vote.EncryptedOraclePrivKey, tallyResult.ConsensusValue) {
			suite.Require().Equal(vote.EncryptedOraclePrivKey, tallyResult.ConsensusValue)
			suite.Require().Equal(sdk.NewInt(10), tallyResult.VotingAmount)
		} else if bytes.Equal(vote2.EncryptedOraclePrivKey, tallyResult.ConsensusValue) {
			suite.Require().Equal(vote2.EncryptedOraclePrivKey, tallyResult.ConsensusValue)
			suite.Require().Equal(sdk.NewInt(10), tallyResult.VotingAmount)
		} else {
			panic(fmt.Sprintf("No matching EncryptedOraclePrivKey(%s) found.", tallyResult.ConsensusValue))
		}
	}

	_, err = suite.OracleKeeper.GetOracle(ctx, suite.newOracleAddr.String())
	suite.Require().Error(err)

	oracleVotes, err = suite.OracleKeeper.GetAllOracleRegistrationVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(oracleVotes))

	events := ctx.EventManager().Events()
	suite.Require().Equal(1, len(events))
	suite.Require().Equal(types.EventTypeRegistrationVote, events[0].Type)
	eventAttributes := events[0].Attributes
	suite.Require().Equal(2, len(eventAttributes))
	suite.Require().Equal(types.AttributeKeyVoteStatus, string(eventAttributes[0].Key))
	suite.Require().Equal(types.AttributeValueVoteStatusEnded, string(eventAttributes[0].Value))
	suite.Require().Equal(types.AttributeKeyOracleAddress, string(eventAttributes[1].Key))
	suite.Require().Equal(oracleRegistration.Address, string(eventAttributes[1].Value))
}
