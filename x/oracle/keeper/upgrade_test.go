package keeper_test

import (
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

func (suite oracleTestSuite) TestUpgradeOracleFailedValidatorNotFound() {
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

func (suite oracleTestSuite) TestUpgradeOracleFailedValidatorJailed() {
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

func (suite oracleTestSuite) TestApplyUpgradeSuccess() {
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
