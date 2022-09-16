package keeper_test

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/medibloc/panacea-core/v2/x/oracle/testutil"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
)

type oracleTestSuite struct {
	testutil.OracleBaseTestSuite

	uniqueID string

	oracleAccPrivKey cryptotypes.PrivKey
	oracleAccPubKey  cryptotypes.PubKey
	oracleAccAddr    sdk.AccAddress

	newOracleAccPrivKey cryptotypes.PrivKey
	newOracleAccPubKey  cryptotypes.PubKey
	newOracleAccAddr    sdk.AccAddress

	oraclePrivKey *btcec.PrivateKey
	oraclePubKey  *btcec.PublicKey

	nodePrivKey *btcec.PrivateKey
	nodePubKey  *btcec.PublicKey

	nodePubKeyRemoteReport []byte

	trustedBlockHeight int64
	trustedBlockHash   []byte
}

func TestOracleTestSuite(t *testing.T) {
	suite.Run(t, new(oracleTestSuite))
}

func (suite *oracleTestSuite) BeforeTest(_, _ string) {
	ctx := suite.Ctx

	suite.uniqueID = "correctUniqueID"
	suite.oracleAccPrivKey = secp256k1.GenPrivKey()
	suite.oracleAccPubKey = suite.oracleAccPrivKey.PubKey()
	suite.oracleAccAddr = sdk.AccAddress(suite.oracleAccPubKey.Address())

	suite.newOracleAccPrivKey = secp256k1.GenPrivKey()
	suite.newOracleAccPubKey = suite.newOracleAccPrivKey.PubKey()
	suite.newOracleAccAddr = sdk.AccAddress(suite.newOracleAccPubKey.Address())

	suite.oraclePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.oraclePubKey = suite.oraclePrivKey.PubKey()

	suite.nodePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.nodePubKey = suite.nodePrivKey.PubKey()

	suite.nodePubKeyRemoteReport = []byte("nodePubKeyRemoteReport")

	suite.trustedBlockHeight = int64(1)
	suite.trustedBlockHash = []byte("trustedBlockHash")

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

func (suite oracleTestSuite) makeNewOracleRegistration() *types.OracleRegistration {
	return &types.OracleRegistration{
		UniqueId:               suite.uniqueID,
		Address:                suite.newOracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: nil,
		TrustedBlockHeight:     0,
		TrustedBlockHash:       nil,
		EncryptedOraclePrivKey: nil,
		Status:                 types.ORACLE_REGISTRATION_STATUS_VOTING_PERIOD,
		VotingPeriod: &types.VotingPeriod{
			VotingStartTime: time.Now(),
			VotingEndTime:   time.Now().Add(5 * time.Second),
		},
	}
}

func (suite oracleTestSuite) TestRegisterOracleSuccess() {
	ctx := suite.Ctx

	// set validator
	suite.SetValidator(suite.oracleAccPubKey, sdk.NewInt(70))

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:               suite.uniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     suite.trustedBlockHeight,
		TrustedBlockHash:       suite.trustedBlockHash,
	}

	err := suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().NoError(err)

	votingPeriod := suite.OracleKeeper.GetVotingPeriod(ctx)

	oracleFromKeeper, err := suite.OracleKeeper.GetOracleRegistration(ctx, suite.uniqueID, suite.oracleAccAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(suite.uniqueID, oracleFromKeeper.UniqueId)
	suite.Require().Equal(suite.oracleAccAddr.String(), oracleFromKeeper.Address)
	suite.Require().Equal(suite.nodePubKey.SerializeCompressed(), oracleFromKeeper.NodePubKey)
	suite.Require().Equal(suite.nodePubKeyRemoteReport, oracleFromKeeper.NodePubKeyRemoteReport)
	suite.Require().Equal(suite.trustedBlockHeight, oracleFromKeeper.TrustedBlockHeight)
	suite.Require().Equal(suite.trustedBlockHash, oracleFromKeeper.TrustedBlockHash)
	suite.Require().Nil(oracleFromKeeper.EncryptedOraclePrivKey)
	suite.Require().Equal(types.ORACLE_REGISTRATION_STATUS_VOTING_PERIOD, oracleFromKeeper.Status)
	suite.Require().Equal(votingPeriod, oracleFromKeeper.VotingPeriod)
	suite.Require().Nil(oracleFromKeeper.TallyResult)

	events := suite.Ctx.EventManager().Events()
	suite.Require().Equal(1, len(events))
	suite.Require().Equal(types.EventTypeRegistrationVote, events[0].Type)

	eventVoteAttributes := events[0].Attributes
	suite.Require().Equal(types.AttributeKeyVoteStatus, string(eventVoteAttributes[0].Key))
	suite.Require().Equal(types.AttributeValueVoteStatusStarted, string(eventVoteAttributes[0].Value))
	suite.Require().Equal(types.AttributeKeyOracleAddress, string(eventVoteAttributes[1].Key))
	suite.Require().Equal(suite.oracleAccAddr.String(), string(eventVoteAttributes[1].Value))
}

func (suite oracleTestSuite) TestRegisterOracleFailedValidatorNotFound() {
	ctx := suite.Ctx

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:               suite.uniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     suite.trustedBlockHeight,
		TrustedBlockHash:       suite.trustedBlockHash,
	}

	err := suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, types.ErrValidatorNotFound)
}

func (suite oracleTestSuite) TestRegisterOracleFailedValidatorJailed() {
	ctx := suite.Ctx

	// set jailed validator
	suite.SetValidator(suite.oracleAccPubKey, sdk.NewInt(70))
	suite.StakingKeeper.Jail(ctx, suite.oracleAccAddr.Bytes())

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:               suite.uniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     suite.trustedBlockHeight,
		TrustedBlockHash:       suite.trustedBlockHash,
	}

	err := suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, types.ErrJailedValidator)
}

func (suite oracleTestSuite) TestRegisterOracleFailedInvalidUniqueID() {
	ctx := suite.Ctx

	// set validator
	suite.SetValidator(suite.oracleAccPubKey, sdk.NewInt(70))

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:               "invalidUniqueID",
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     suite.trustedBlockHeight,
		TrustedBlockHash:       suite.trustedBlockHash,
	}

	err := suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, types.ErrRegisterOracle)
}

func (suite oracleTestSuite) TestRegisterOracleFailedStatusVotingPeriod() {
	ctx := suite.Ctx

	suite.SetValidator(suite.oracleAccPubKey, sdk.NewInt(70))

	votingOracleRegistration := &types.OracleRegistration{
		UniqueId:               suite.uniqueID,
		Address:                suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     suite.trustedBlockHeight,
		TrustedBlockHash:       suite.trustedBlockHash,
		Status:                 types.ORACLE_REGISTRATION_STATUS_VOTING_PERIOD,
		VotingPeriod:           suite.OracleKeeper.GetVotingPeriod(ctx),
	}

	err := suite.OracleKeeper.SetOracleRegistration(ctx, votingOracleRegistration)
	suite.Require().NoError(err)

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:               suite.uniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     suite.trustedBlockHeight,
		TrustedBlockHash:       suite.trustedBlockHash,
	}

	err = suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, types.ErrRegisterOracle)
}

func (suite oracleTestSuite) TestRegisterOracleFailedStatusPassedNotJailed() {
	ctx := suite.Ctx

	suite.SetValidator(suite.oracleAccPubKey, sdk.NewInt(70))

	existingOracleRegistration := &types.OracleRegistration{
		UniqueId:               suite.uniqueID,
		Address:                suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     suite.trustedBlockHeight,
		TrustedBlockHash:       suite.trustedBlockHash,
		Status:                 types.ORACLE_REGISTRATION_STATUS_PASSED,
		VotingPeriod:           suite.OracleKeeper.GetVotingPeriod(ctx),
	}

	err := suite.OracleKeeper.SetOracleRegistration(ctx, existingOracleRegistration)
	suite.Require().NoError(err)

	existingOracle := &types.Oracle{
		Address: suite.oracleAccAddr.String(),
		Status:  types.ORACLE_STATUS_ACTIVE,
	}

	err = suite.OracleKeeper.SetOracle(ctx, existingOracle)
	suite.Require().NoError(err)

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:               suite.uniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     suite.trustedBlockHeight,
		TrustedBlockHash:       suite.trustedBlockHash,
	}

	err = suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, types.ErrRegisterOracle)
}

func (suite *oracleTestSuite) TestOracleRegistrationVoteSuccess() {
	ctx := suite.Ctx

	suite.CreateOracleValidator(suite.oracleAccPubKey, sdk.NewInt(70))
	suite.SetAccount(suite.newOracleAccPubKey)
	suite.SetValidator(suite.newOracleAccPubKey, sdk.NewInt(20))

	oracleRegistration := suite.makeNewOracleRegistration()
	err := suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	// make the correct encryptedOraclePrivKey
	encryptedOraclePrivKey, err := btcec.Encrypt(suite.nodePubKey, suite.oraclePrivKey.Serialize())
	suite.Require().NoError(err)
	// make the correct vote info
	oracleRegistrationVote := &types.OracleRegistrationVote{
		UniqueId:               suite.uniqueID,
		VoterAddress:           suite.oracleAccAddr.String(),
		VotingTargetAddress:    suite.newOracleAccAddr.String(),
		VoteOption:             types.VOTE_OPTION_YES,
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	// make the correct signature
	voteBz, err := suite.Cdc.Marshaler.Marshal(oracleRegistrationVote)
	suite.Require().NoError(err)
	oraclePrivKeySecp256k1 := secp256k1.PrivKey{
		Key: suite.oraclePrivKey.Serialize(),
	}
	signature, err := oraclePrivKeySecp256k1.Sign(voteBz)
	suite.Require().NoError(err)

	err = suite.OracleKeeper.VoteOracleRegistration(ctx, oracleRegistrationVote, signature)
	suite.Require().NoError(err)

	getOracleRegistrationVote, err := suite.OracleKeeper.GetOracleRegistrationVote(
		ctx,
		suite.uniqueID,
		suite.newOracleAccAddr.String(),
		suite.oracleAccAddr.String(),
	)
	suite.Require().NoError(err)
	suite.Require().Equal(oracleRegistrationVote, getOracleRegistrationVote)
}

func (suite *oracleTestSuite) TestOracleRegistrationVoteFailedVerifySignature() {
	ctx := suite.Ctx

	suite.CreateOracleValidator(suite.oracleAccPubKey, sdk.NewInt(70))
	suite.SetAccount(suite.newOracleAccPubKey)
	suite.SetValidator(suite.newOracleAccPubKey, sdk.NewInt(20))

	oracleRegistration := suite.makeNewOracleRegistration()
	err := suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	// make the correct encryptedOraclePrivKey
	encryptedOraclePrivKey, err := btcec.Encrypt(suite.nodePubKey, suite.oraclePrivKey.Serialize())
	suite.Require().NoError(err)
	// make the correct vote info
	oracleRegistrationVote := &types.OracleRegistrationVote{
		UniqueId:               suite.uniqueID,
		VoterAddress:           suite.oracleAccAddr.String(),
		VotingTargetAddress:    suite.newOracleAccAddr.String(),
		VoteOption:             types.VOTE_OPTION_YES,
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	// make the correct signature
	voteBz, err := suite.Cdc.Marshaler.Marshal(oracleRegistrationVote)
	suite.Require().NoError(err)
	invalidOraclePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	suite.Require().NoError(err)
	oraclePrivKeySecp256k1 := secp256k1.PrivKey{
		Key: invalidOraclePrivKey.Serialize(),
	}
	signature, err := oraclePrivKeySecp256k1.Sign(voteBz)
	suite.Require().NoError(err)

	err = suite.OracleKeeper.VoteOracleRegistration(ctx, oracleRegistrationVote, signature)
	suite.Require().ErrorIs(err, types.ErrDetectionMaliciousBehavior)
}

func (suite *oracleTestSuite) TestOracleRegistrationVoteInvalidUniqueID() {
	ctx := suite.Ctx

	suite.CreateOracleValidator(suite.oracleAccPubKey, sdk.NewInt(70))
	suite.SetAccount(suite.newOracleAccPubKey)
	suite.SetValidator(suite.newOracleAccPubKey, sdk.NewInt(20))

	oracleRegistration := suite.makeNewOracleRegistration()
	err := suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	// make the correct encryptedOraclePrivKey
	encryptedOraclePrivKey, err := btcec.Encrypt(suite.nodePubKey, suite.oraclePrivKey.Serialize())
	suite.Require().NoError(err)
	// make vote with invalid uniqueID
	invalidUniqueID := "invalidUniqueID"
	oracleRegistrationVote := &types.OracleRegistrationVote{
		UniqueId:               invalidUniqueID,
		VoterAddress:           suite.oracleAccAddr.String(),
		VotingTargetAddress:    suite.newOracleAccAddr.String(),
		VoteOption:             types.VOTE_OPTION_YES,
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	// make the correct signature
	voteBz, err := suite.Cdc.Marshaler.Marshal(oracleRegistrationVote)
	suite.Require().NoError(err)
	oraclePrivKeySecp256k1 := secp256k1.PrivKey{
		Key: suite.oraclePrivKey.Serialize(),
	}
	signature, err := oraclePrivKeySecp256k1.Sign(voteBz)
	suite.Require().NoError(err)

	err = suite.OracleKeeper.VoteOracleRegistration(ctx, oracleRegistrationVote, signature)
	suite.Require().ErrorIs(err, types.ErrOracleRegistrationVote)
	suite.Require().ErrorContains(err, fmt.Sprintf("not matched with the currently active uniqueID. expected %s, got %s", suite.uniqueID, invalidUniqueID))
}

func (suite *oracleTestSuite) TestOracleRegistrationVoteInvalidGenesisOracleStatus() {
	ctx := suite.Ctx

	suite.SetAccount(suite.oracleAccPubKey)
	suite.SetValidator(suite.oracleAccPubKey, sdk.NewInt(70))
	err := suite.OracleKeeper.SetOracle(ctx, &types.Oracle{
		Address:  suite.oracleAccAddr.String(),
		Status:   types.ORACLE_STATUS_JAILED,
		Uptime:   0,
		JailedAt: nil,
	})
	suite.Require().NoError(err)

	suite.SetAccount(suite.newOracleAccPubKey)
	suite.SetValidator(suite.newOracleAccPubKey, sdk.NewInt(20))

	oracleRegistration := suite.makeNewOracleRegistration()
	err = suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	// make the correct encryptedOraclePrivKey
	encryptedOraclePrivKey, err := btcec.Encrypt(suite.nodePubKey, suite.oraclePrivKey.Serialize())
	suite.Require().NoError(err)
	// make vote with invalid uniqueID
	oracleRegistrationVote := &types.OracleRegistrationVote{
		UniqueId:               suite.uniqueID,
		VoterAddress:           suite.oracleAccAddr.String(),
		VotingTargetAddress:    suite.newOracleAccAddr.String(),
		VoteOption:             types.VOTE_OPTION_YES,
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	// make the correct signature
	voteBz, err := suite.Cdc.Marshaler.Marshal(oracleRegistrationVote)
	suite.Require().NoError(err)
	oraclePrivKeySecp256k1 := secp256k1.PrivKey{
		Key: suite.oraclePrivKey.Serialize(),
	}
	signature, err := oraclePrivKeySecp256k1.Sign(voteBz)
	suite.Require().NoError(err)

	err = suite.OracleKeeper.VoteOracleRegistration(ctx, oracleRegistrationVote, signature)
	suite.Require().ErrorIs(err, types.ErrOracleRegistrationVote)
	suite.Require().ErrorContains(err, "this oracle is not in 'ACTIVE' state")
}

func (suite *oracleTestSuite) TestOracleRegistrationVoteInvalidOracleRegistrationStatus() {
	ctx := suite.Ctx

	suite.CreateOracleValidator(suite.oracleAccPubKey, sdk.NewInt(70))
	suite.SetAccount(suite.newOracleAccPubKey)
	suite.SetValidator(suite.newOracleAccPubKey, sdk.NewInt(20))

	oracleRegistration := suite.makeNewOracleRegistration()
	oracleRegistration.Status = types.ORACLE_REGISTRATION_STATUS_REJECTED
	err := suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	// make the correct encryptedOraclePrivKey
	encryptedOraclePrivKey, err := btcec.Encrypt(suite.nodePubKey, suite.oraclePrivKey.Serialize())
	suite.Require().NoError(err)
	// make vote with invalid uniqueID
	oracleRegistrationVote := &types.OracleRegistrationVote{
		UniqueId:               suite.uniqueID,
		VoterAddress:           suite.oracleAccAddr.String(),
		VotingTargetAddress:    suite.newOracleAccAddr.String(),
		VoteOption:             types.VOTE_OPTION_YES,
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	// make the correct signature
	voteBz, err := suite.Cdc.Marshaler.Marshal(oracleRegistrationVote)
	suite.Require().NoError(err)
	oraclePrivKeySecp256k1 := secp256k1.PrivKey{
		Key: suite.oraclePrivKey.Serialize(),
	}
	signature, err := oraclePrivKeySecp256k1.Sign(voteBz)
	suite.Require().NoError(err)

	err = suite.OracleKeeper.VoteOracleRegistration(ctx, oracleRegistrationVote, signature)
	suite.Require().ErrorIs(err, types.ErrOracleRegistrationVote)
	suite.Require().ErrorContains(err, "the currently voted oracle's status is not 'VOTING_PERIOD'")
}

func (suite *oracleTestSuite) TestOracleRegistrationEmittedEvent() {
	ctx := suite.Ctx

	suite.CreateOracleValidator(suite.oracleAccPubKey, sdk.NewInt(70))
	suite.SetAccount(suite.newOracleAccPubKey)
	suite.SetValidator(suite.newOracleAccPubKey, sdk.NewInt(20))

	msg := &types.MsgRegisterOracle{
		UniqueId:               suite.uniqueID,
		OracleAddress:          suite.newOracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     1,
		TrustedBlockHash:       []byte("trustedBlockHash"),
	}

	err := suite.OracleKeeper.RegisterOracle(ctx, msg)
	suite.Require().NoError(err)

	oracleRegistration, err := suite.OracleKeeper.GetOracleRegistration(ctx, msg.UniqueId, msg.OracleAddress)
	suite.Require().NoError(err)
	suite.Require().Equal(msg.UniqueId, oracleRegistration.UniqueId)
	suite.Require().Equal(msg.OracleAddress, oracleRegistration.Address)
	suite.Require().Equal(msg.NodePubKey, oracleRegistration.NodePubKey)
	suite.Require().Equal(msg.NodePubKeyRemoteReport, oracleRegistration.NodePubKeyRemoteReport)
	suite.Require().Equal(msg.TrustedBlockHeight, oracleRegistration.TrustedBlockHeight)
	suite.Require().Equal(msg.TrustedBlockHash, oracleRegistration.TrustedBlockHash)

	events := ctx.EventManager().Events()
	suite.Require().Equal(1, len(events))
	suite.Require().Equal(types.EventTypeRegistrationVote, events[0].Type)
	eventAttributes := events[0].Attributes
	suite.Require().Equal(2, len(eventAttributes))
	suite.Require().Equal(types.AttributeKeyVoteStatus, string(eventAttributes[0].Key))
	suite.Require().Equal(types.AttributeValueVoteStatusStarted, string(eventAttributes[0].Value))
	suite.Require().Equal(types.AttributeKeyOracleAddress, string(eventAttributes[1].Key))
	suite.Require().Equal(msg.OracleAddress, string(eventAttributes[1].Value))
}
