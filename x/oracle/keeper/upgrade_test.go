package keeper_test

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/require"
)

func (suite *oracleTestSuite) TestOracleUpgradeInfo() {
	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: "upgradeUniqueID",
		Height:   10000000,
	}

	require.NoError(suite.T(), suite.OracleKeeper.SetOracleUpgradeInfo(suite.Ctx, upgradeInfo))

	getUpgradeInfo, err := suite.OracleKeeper.GetOracleUpgradeInfo(suite.Ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), upgradeInfo, getUpgradeInfo)
}

func (suite *oracleTestSuite) TestEmptyOracleUpgradeInfo() {
	upgradeInfo, err := suite.OracleKeeper.GetOracleUpgradeInfo(suite.Ctx)
	require.ErrorIs(suite.T(), err, types.ErrOracleUpgradeInfoNotFound)
	require.Nil(suite.T(), upgradeInfo)
}

func (suite *oracleTestSuite) TestUpgradeOracleSuccess() {
	ctx := suite.Ctx
	upgradeUniqueID := "UpgradeUniqueID"
	nonce := []byte("nonce")
	votingPeriod := suite.OracleKeeper.GetVotingPeriod(ctx)

	suite.CreateOracleValidator(suite.oracleAccPubKey, sdk.NewInt(70))

	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: upgradeUniqueID,
		Height:   100000,
	}
	require.NoError(suite.T(), suite.OracleKeeper.SetOracleUpgradeInfo(ctx, upgradeInfo))

	msg := &types.MsgUpgradeOracle{
		UniqueId:               upgradeUniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     suite.trustedBlockHeight,
		TrustedBlockHash:       suite.trustedBlockHash,
		Nonce:                  nonce,
	}
	require.NoError(suite.T(), suite.OracleKeeper.UpgradeOracle(ctx, msg))

	oracleRegistration, err := suite.OracleKeeper.GetOracleRegistration(ctx, upgradeUniqueID, suite.oracleAccAddr.String())
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), upgradeUniqueID, oracleRegistration.UniqueId)
	require.Equal(suite.T(), suite.oracleAccAddr.String(), oracleRegistration.Address)
	require.Equal(suite.T(), suite.nodePubKey.SerializeCompressed(), oracleRegistration.NodePubKey)
	require.Equal(suite.T(), suite.nodePubKeyRemoteReport, oracleRegistration.NodePubKeyRemoteReport)
	require.Equal(suite.T(), suite.trustedBlockHeight, oracleRegistration.TrustedBlockHeight)
	require.Equal(suite.T(), suite.trustedBlockHash, oracleRegistration.TrustedBlockHash)
	require.Equal(suite.T(), types.ORACLE_REGISTRATION_STATUS_VOTING_PERIOD, oracleRegistration.Status)
	require.Equal(suite.T(), votingPeriod, oracleRegistration.VotingPeriod)

	events := suite.Ctx.EventManager().Events()
	suite.Require().Equal(1, len(events))
	suite.Require().Equal(types.EventTypeUpgradeVote, events[0].Type)

	eventVoteAttributes := events[0].Attributes
	suite.Require().Equal(types.AttributeKeyUniqueID, string(eventVoteAttributes[0].Key))
	suite.Require().Equal(upgradeUniqueID, string(eventVoteAttributes[0].Value))
	suite.Require().Equal(types.AttributeKeyVoteStatus, string(eventVoteAttributes[1].Key))
	suite.Require().Equal(types.AttributeValueVoteStatusStarted, string(eventVoteAttributes[1].Value))
	suite.Require().Equal(types.AttributeKeyOracleAddress, string(eventVoteAttributes[2].Key))
	suite.Require().Equal(suite.oracleAccAddr.String(), string(eventVoteAttributes[2].Value))
}

func (suite *oracleTestSuite) TestUpgradeOracleNotSameUniqueID() {
	ctx := suite.Ctx
	upgradeUniqueID := "UpgradeUniqueID"
	nonce := []byte("nonce")

	suite.CreateOracleValidator(suite.oracleAccPubKey, sdk.NewInt(70))

	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: upgradeUniqueID,
		Height:   100000,
	}
	require.NoError(suite.T(), suite.OracleKeeper.SetOracleUpgradeInfo(ctx, upgradeInfo))

	msg := &types.MsgUpgradeOracle{
		UniqueId:               "wrongUniqueID",
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     suite.trustedBlockHeight,
		TrustedBlockHash:       suite.trustedBlockHash,
		Nonce:                  nonce,
	}

	err := suite.OracleKeeper.UpgradeOracle(ctx, msg)
	require.ErrorIs(suite.T(), err, types.ErrUpgradeOracle)

	events := suite.Ctx.EventManager().Events()
	suite.Require().Equal(0, len(events))
}

func (suite *oracleTestSuite) TestUpgradeOracleFailedValidatorNotFound() {
	ctx := suite.Ctx
	nonce := []byte("nonce")

	msg := &types.MsgUpgradeOracle{
		UniqueId:               "wrongUniqueID",
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     suite.trustedBlockHeight,
		TrustedBlockHash:       suite.trustedBlockHash,
		Nonce:                  nonce,
	}

	err := suite.OracleKeeper.UpgradeOracle(ctx, msg)
	suite.Require().Error(err, types.ErrValidatorNotFound)
}

func (suite *oracleTestSuite) TestUpgradeOracleFailedValidatorJailed() {
	ctx := suite.Ctx
	nonce := []byte("nonce")

	// set jailed validator
	suite.SetValidator(suite.oracleAccPubKey, sdk.NewInt(70))
	suite.StakingKeeper.Jail(ctx, suite.oracleAccAddr.Bytes())

	msg := &types.MsgUpgradeOracle{
		UniqueId:               "wrongUniqueID",
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     suite.trustedBlockHeight,
		TrustedBlockHash:       suite.trustedBlockHash,
		Nonce:                  nonce,
	}

	err := suite.OracleKeeper.UpgradeOracle(ctx, msg)
	suite.Require().Error(err, types.ErrJailedValidator)
}

func (suite *oracleTestSuite) TestApplyUpgradeSuccess() {
	ctx := suite.Ctx

	params := types.DefaultParams()
	params.UniqueId = "orgUniqueID"
	suite.OracleKeeper.SetParams(ctx, params)
	suite.Require().Equal(params.UniqueId, suite.OracleKeeper.GetParams(ctx).UniqueId)

	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: "upgradeUniqueID",
		Height:   1,
	}

	suite.Require().NoError(suite.OracleKeeper.ApplyUpgrade(ctx, upgradeInfo))
	suite.Require().Equal(upgradeInfo.UniqueId, suite.OracleKeeper.GetParams(ctx).UniqueId)

	events := suite.Ctx.EventManager().Events()
	suite.Require().Equal(1, len(events))
	suite.Require().Equal(types.EventTypeUpgradeVote, events[0].Type)
	suite.Require().Equal(types.AttributeKeyUniqueID, string(events[0].Attributes[0].Key))
	suite.Require().Equal(upgradeInfo.UniqueId, string(events[0].Attributes[0].Value))
}

func (suite *oracleTestSuite) TestOracleUpgradeVoteSuccess() {
	ctx := suite.Ctx

	suite.CreateOracleValidator(suite.oracleAccPubKey, sdk.NewInt(70))
	suite.SetAccount(suite.newOracleAccPubKey)
	suite.SetValidator(suite.newOracleAccPubKey, sdk.NewInt(20))

	upgradeUniqueID := "upgradeUniqueID"

	oracleRegistration := suite.makeNewOracleRegistration()
	oracleRegistration.UniqueId = upgradeUniqueID
	oracleRegistration.RegistrationType = types.ORACLE_REGISTRATION_TYPE_UPGRADE
	err := suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: upgradeUniqueID,
		Height:   1000,
	}
	err = suite.OracleKeeper.SetOracleUpgradeInfo(ctx, upgradeInfo)
	suite.Require().NoError(err)

	// make the correct encryptedOraclePrivKey
	encryptedOraclePrivKey, err := btcec.Encrypt(suite.nodePubKey, suite.oraclePrivKey.Serialize())
	suite.Require().NoError(err)
	// make the correct vote info
	oracleRegistrationVote := &types.OracleRegistrationVote{
		UniqueId:               upgradeUniqueID,
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
		upgradeUniqueID,
		suite.newOracleAccAddr.String(),
		suite.oracleAccAddr.String(),
	)
	suite.Require().NoError(err)
	suite.Require().Equal(oracleRegistrationVote, getOracleRegistrationVote)
}

func (suite *oracleTestSuite) TestOracleUpgradeVoteFailedVerifySignature() {
	ctx := suite.Ctx

	suite.CreateOracleValidator(suite.oracleAccPubKey, sdk.NewInt(70))
	suite.SetAccount(suite.newOracleAccPubKey)
	suite.SetValidator(suite.newOracleAccPubKey, sdk.NewInt(20))

	upgradeUniqueID := "upgradeUniqueID"

	oracleRegistration := suite.makeNewOracleRegistration()
	oracleRegistration.UniqueId = upgradeUniqueID
	oracleRegistration.RegistrationType = types.ORACLE_REGISTRATION_TYPE_UPGRADE
	err := suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: upgradeUniqueID,
		Height:   1000,
	}
	err = suite.OracleKeeper.SetOracleUpgradeInfo(ctx, upgradeInfo)
	suite.Require().NoError(err)

	// make the correct encryptedOraclePrivKey
	encryptedOraclePrivKey, err := btcec.Encrypt(suite.nodePubKey, suite.oraclePrivKey.Serialize())
	suite.Require().NoError(err)
	// make the correct vote info
	oracleRegistrationVote := &types.OracleRegistrationVote{
		UniqueId:               upgradeUniqueID,
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

func (suite *oracleTestSuite) TestOracleUpgradeVoteInvalidUniqueID() {
	ctx := suite.Ctx

	suite.CreateOracleValidator(suite.oracleAccPubKey, sdk.NewInt(70))
	suite.SetAccount(suite.newOracleAccPubKey)
	suite.SetValidator(suite.newOracleAccPubKey, sdk.NewInt(20))

	upgradeUniqueID := "upgradeUniqueID"

	oracleRegistration := suite.makeNewOracleRegistration()
	oracleRegistration.UniqueId = upgradeUniqueID
	oracleRegistration.RegistrationType = types.ORACLE_REGISTRATION_TYPE_UPGRADE
	err := suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: upgradeUniqueID,
		Height:   1000,
	}
	err = suite.OracleKeeper.SetOracleUpgradeInfo(ctx, upgradeInfo)
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
	suite.Require().ErrorIs(err, types.ErrOracleRegistrationVote, "oracle registration not found")
}
