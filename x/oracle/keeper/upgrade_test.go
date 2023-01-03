package keeper_test

import (
	"encoding/base64"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type oracleUpgradeTestSuite struct {
	testsuite.TestSuite

	oracleAccPrivKey cryptotypes.PrivKey
	oracleAccPubKey  cryptotypes.PubKey
	oracleAccAddr    sdk.AccAddress

	approverAccPrivKey cryptotypes.PrivKey
	approverAccPubKey  cryptotypes.PubKey
	approverAccAddr    sdk.AccAddress

	oraclePrivKey *btcec.PrivateKey
	oraclePubKey  *btcec.PublicKey

	nodePrivKey *btcec.PrivateKey
	nodePubKey  *btcec.PublicKey

	nodePubKeyRemoteReport []byte

	currentUniqueID string
	upgradeUniqueID string
}

func TestOracleUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(oracleUpgradeTestSuite))
}

func (suite *oracleUpgradeTestSuite) BeforeTest(_, _ string) {
	suite.oracleAccPrivKey = secp256k1.GenPrivKey()
	suite.oracleAccPubKey = suite.oracleAccPrivKey.PubKey()
	suite.oracleAccAddr = sdk.AccAddress(suite.oracleAccPubKey.Address())

	suite.approverAccPrivKey = secp256k1.GenPrivKey()
	suite.approverAccPubKey = suite.approverAccPrivKey.PubKey()
	suite.approverAccAddr = sdk.AccAddress(suite.approverAccPubKey.Address())

	suite.oraclePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.oraclePubKey = suite.oraclePrivKey.PubKey()

	suite.nodePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.nodePubKey = suite.nodePrivKey.PubKey()

	suite.nodePubKeyRemoteReport = []byte("nodePubKeyRemoteReport")

	suite.currentUniqueID = "currentUniqueID"
	suite.upgradeUniqueID = "upgradeUniqueID"

	suite.OracleKeeper.SetParams(suite.Ctx, types.Params{
		OraclePublicKey:          base64.StdEncoding.EncodeToString(suite.oraclePubKey.SerializeCompressed()),
		OraclePubKeyRemoteReport: "",
		UniqueId:                 suite.currentUniqueID,
	})
}

func (suite *oracleUpgradeTestSuite) TestOracleUpgradeInfo() {
	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: suite.upgradeUniqueID,
		Height:   10000000,
	}

	require.NoError(suite.T(), suite.OracleKeeper.SetOracleUpgradeInfo(suite.Ctx, upgradeInfo))

	getUpgradeInfo, err := suite.OracleKeeper.GetOracleUpgradeInfo(suite.Ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), upgradeInfo, getUpgradeInfo)
}

func (suite *oracleUpgradeTestSuite) TestEmptyOracleUpgradeInfo() {
	upgradeInfo, err := suite.OracleKeeper.GetOracleUpgradeInfo(suite.Ctx)
	require.ErrorIs(suite.T(), err, types.ErrOracleUpgradeInfoNotFound)
	require.Nil(suite.T(), upgradeInfo)
}

func (suite *oracleUpgradeTestSuite) TestApplyUpgradeSuccess() {
	ctx := suite.Ctx

	params := types.DefaultParams()
	params.UniqueId = suite.currentUniqueID
	suite.OracleKeeper.SetParams(ctx, params)
	suite.Require().Equal(params.UniqueId, suite.OracleKeeper.GetParams(ctx).UniqueId)

	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: suite.upgradeUniqueID,
		Height:   1,
	}

	suite.Require().NoError(suite.OracleKeeper.ApplyUpgrade(ctx, upgradeInfo))
	suite.Require().Equal(upgradeInfo.UniqueId, suite.OracleKeeper.GetParams(ctx).UniqueId)
}

func (suite *oracleUpgradeTestSuite) TestUpgradeOracleSuccess() {
	ctx := suite.Ctx

	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: suite.upgradeUniqueID,
		Height:   10,
	}

	suite.Require().NoError(suite.OracleKeeper.SetOracleUpgradeInfo(ctx, upgradeInfo))

	oracle := &types.Oracle{
		OracleAddress:                 suite.oracleAccAddr.String(),
		UniqueId:                      suite.currentUniqueID,
		Endpoint:                      "test.com",
		UpdateTime:                    ctx.BlockTime(),
		OracleCommissionRate:          sdk.NewDecWithPrec(1, 1),
		OracleCommissionMaxRate:       sdk.NewDecWithPrec(2, 1),
		OracleCommissionMaxChangeRate: sdk.NewDecWithPrec(1, 2),
	}

	suite.Require().NoError(suite.OracleKeeper.SetOracle(ctx, oracle))

	msgOracleUpgrade := &types.MsgUpgradeOracle{
		UniqueId:               suite.upgradeUniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     int64(1),
		TrustedBlockHash:       []byte("trustedBlockHash"),
	}

	suite.Require().NoError(suite.OracleKeeper.UpgradeOracle(ctx, msgOracleUpgrade))

	upgrade, err := suite.OracleKeeper.GetOracleUpgrade(ctx, suite.upgradeUniqueID, suite.oracleAccAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(msgOracleUpgrade.OracleAddress, upgrade.OracleAddress)
	suite.Require().Equal(msgOracleUpgrade.UniqueId, upgrade.UniqueId)
	suite.Require().Equal(msgOracleUpgrade.NodePubKey, upgrade.NodePubKey)
	suite.Require().Equal(msgOracleUpgrade.NodePubKeyRemoteReport, upgrade.NodePubKeyRemoteReport)
	suite.Require().Equal(msgOracleUpgrade.TrustedBlockHeight, upgrade.TrustedBlockHeight)
	suite.Require().Equal(msgOracleUpgrade.TrustedBlockHash, upgrade.TrustedBlockHash)

	events := suite.Ctx.EventManager().Events()
	suite.Require().Equal(1, len(events))
	suite.Require().Equal(types.EventTypeUpgrade, events[0].Type)

	eventVoteAttributes := events[0].Attributes
	suite.Require().Equal(2, len(eventVoteAttributes))
	suite.Require().Equal(types.AttributeKeyUniqueID, string(eventVoteAttributes[0].Key))
	suite.Require().Equal(suite.upgradeUniqueID, string(eventVoteAttributes[0].Value))
	suite.Require().Equal(types.AttributeKeyOracleAddress, string(eventVoteAttributes[1].Key))
	suite.Require().Equal(suite.oracleAccAddr.String(), string(eventVoteAttributes[1].Value))
}

func (suite *oracleUpgradeTestSuite) TestUpgradeOracleFailedNotMatchedUniqueID() {
	ctx := suite.Ctx

	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: suite.upgradeUniqueID,
		Height:   10,
	}

	suite.Require().NoError(suite.OracleKeeper.SetOracleUpgradeInfo(ctx, upgradeInfo))

	oracle := &types.Oracle{
		OracleAddress:                 suite.oracleAccAddr.String(),
		UniqueId:                      suite.currentUniqueID,
		Endpoint:                      "test.com",
		UpdateTime:                    ctx.BlockTime(),
		OracleCommissionRate:          sdk.NewDecWithPrec(1, 1),
		OracleCommissionMaxRate:       sdk.NewDecWithPrec(2, 1),
		OracleCommissionMaxChangeRate: sdk.NewDecWithPrec(1, 2),
	}

	suite.Require().NoError(suite.OracleKeeper.SetOracle(ctx, oracle))

	msgOracleUpgrade := &types.MsgUpgradeOracle{
		UniqueId:               "upgradeUniqueID2",
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     int64(1),
		TrustedBlockHash:       []byte("trustedBlockHash"),
	}

	err := suite.OracleKeeper.UpgradeOracle(ctx, msgOracleUpgrade)
	suite.Require().Error(err, types.ErrUpgradeOracle)
	suite.Require().ErrorContains(err, "does not match the upgrade uniqueID")
}

func (suite *oracleUpgradeTestSuite) TestUpgradeOracleFailedNoRegisteredOracle() {
	ctx := suite.Ctx

	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: suite.upgradeUniqueID,
		Height:   10,
	}

	suite.Require().NoError(suite.OracleKeeper.SetOracleUpgradeInfo(ctx, upgradeInfo))

	msgOracleUpgrade := &types.MsgUpgradeOracle{
		UniqueId:               suite.upgradeUniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     int64(1),
		TrustedBlockHash:       []byte("trustedBlockHash"),
	}

	err := suite.OracleKeeper.UpgradeOracle(ctx, msgOracleUpgrade)
	suite.Require().Error(err, types.ErrUpgradeOracle)
	suite.Require().ErrorContains(err, "is not registered oracle")
}

func (suite *oracleUpgradeTestSuite) TestApproveOracleUpgradeSuccess() {
	// requester : suite.oracleAcc
	// approver : suite.approverAcc

	ctx := suite.Ctx

	// set oracle upgrade info
	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: suite.upgradeUniqueID,
		Height:   10,
	}

	suite.Require().NoError(suite.OracleKeeper.SetOracleUpgradeInfo(ctx, upgradeInfo))

	// set approver oracle
	approverOracle := &types.Oracle{
		OracleAddress:                 suite.approverAccAddr.String(),
		UniqueId:                      suite.currentUniqueID,
		Endpoint:                      "iam-approver.com",
		UpdateTime:                    ctx.BlockTime(),
		OracleCommissionRate:          sdk.NewDecWithPrec(1, 1),
		OracleCommissionMaxRate:       sdk.NewDecWithPrec(2, 1),
		OracleCommissionMaxChangeRate: sdk.NewDecWithPrec(1, 2),
	}

	suite.Require().NoError(suite.OracleKeeper.SetOracle(ctx, approverOracle))

	oracle := &types.Oracle{
		OracleAddress:                 suite.oracleAccAddr.String(),
		UniqueId:                      suite.currentUniqueID,
		Endpoint:                      "test.com",
		UpdateTime:                    ctx.BlockTime(),
		OracleCommissionRate:          sdk.NewDecWithPrec(1, 1),
		OracleCommissionMaxRate:       sdk.NewDecWithPrec(2, 1),
		OracleCommissionMaxChangeRate: sdk.NewDecWithPrec(1, 2),
	}

	suite.Require().NoError(suite.OracleKeeper.SetOracle(ctx, oracle))

	msgOracleUpgrade := &types.MsgUpgradeOracle{
		UniqueId:               suite.upgradeUniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     int64(1),
		TrustedBlockHash:       []byte("trustedBlockHash"),
	}

	suite.Require().NoError(suite.OracleKeeper.UpgradeOracle(ctx, msgOracleUpgrade))

	// approve oracle upgrade
	encryptedOraclePrivKey, err := btcec.Encrypt(suite.nodePubKey, suite.oraclePrivKey.Serialize())
	suite.Require().NoError(err)

	approveOracleUpgrade := &types.ApprovalSharingOracleKey{
		ApproverUniqueId:       suite.currentUniqueID,
		ApproverOracleAddress:  suite.approverAccAddr.String(),
		TargetUniqueId:         suite.upgradeUniqueID,
		TargetOracleAddress:    suite.oracleAccAddr.String(),
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	approveOracleRegistrationBz, err := suite.Cdc.Marshaler.Marshal(approveOracleUpgrade)
	suite.Require().NoError(err)
	signature, err := suite.oraclePrivKey.Sign(approveOracleRegistrationBz)
	suite.Require().NoError(err)

	msgApproveOracleUpgrade := types.NewMsgApproveOracleUpgrade(approveOracleUpgrade, signature.Serialize())

	err = suite.OracleKeeper.ApproveOracleUpgrade(ctx, msgApproveOracleUpgrade)
	suite.Require().NoError(err)

	// check events
	events := ctx.EventManager().Events()

	requiredEvents := map[string]bool{
		types.EventTypeApproveOracleUpgrade: false,
	}

	for _, e := range events {
		if e.Type == types.EventTypeApproveOracleUpgrade {
			requiredEvents[e.Type] = true
			suite.Require().Equal(types.AttributeKeyOracleAddress, string(e.Attributes[0].Key))
			suite.Require().Equal(suite.oracleAccAddr.String(), string(e.Attributes[0].Value))
			suite.Require().Equal(types.AttributeKeyUniqueID, string(e.Attributes[1].Key))
			suite.Require().Equal(suite.upgradeUniqueID, string(e.Attributes[1].Value))
		}
	}

	for _, v := range requiredEvents {
		suite.Require().True(v)
	}
}

func (suite *oracleUpgradeTestSuite) TestApproveOracleUpgradeFailAlreadyApproved() {
	// requester : suite.oracleAcc
	// approver : suite.approverAcc

	ctx := suite.Ctx

	// set oracle upgrade info
	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: suite.upgradeUniqueID,
		Height:   10,
	}

	suite.Require().NoError(suite.OracleKeeper.SetOracleUpgradeInfo(ctx, upgradeInfo))

	// set approver oracle
	approverOracle := &types.Oracle{
		OracleAddress:                 suite.approverAccAddr.String(),
		UniqueId:                      suite.currentUniqueID,
		Endpoint:                      "iam-approver.com",
		UpdateTime:                    ctx.BlockTime(),
		OracleCommissionRate:          sdk.NewDecWithPrec(1, 1),
		OracleCommissionMaxRate:       sdk.NewDecWithPrec(2, 1),
		OracleCommissionMaxChangeRate: sdk.NewDecWithPrec(1, 2),
	}

	suite.Require().NoError(suite.OracleKeeper.SetOracle(ctx, approverOracle))

	oracle := &types.Oracle{
		OracleAddress:                 suite.oracleAccAddr.String(),
		UniqueId:                      suite.currentUniqueID,
		Endpoint:                      "test.com",
		UpdateTime:                    ctx.BlockTime(),
		OracleCommissionRate:          sdk.NewDecWithPrec(1, 1),
		OracleCommissionMaxRate:       sdk.NewDecWithPrec(2, 1),
		OracleCommissionMaxChangeRate: sdk.NewDecWithPrec(1, 2),
	}

	suite.Require().NoError(suite.OracleKeeper.SetOracle(ctx, oracle))

	msgOracleUpgrade := &types.MsgUpgradeOracle{
		UniqueId:               suite.upgradeUniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     int64(1),
		TrustedBlockHash:       []byte("trustedBlockHash"),
	}

	suite.Require().NoError(suite.OracleKeeper.UpgradeOracle(ctx, msgOracleUpgrade))

	// approve oracle upgrade
	encryptedOraclePrivKey, err := btcec.Encrypt(suite.nodePubKey, suite.oraclePrivKey.Serialize())
	suite.Require().NoError(err)

	approveOracleUpgrade := &types.ApprovalSharingOracleKey{
		ApproverUniqueId:       suite.currentUniqueID,
		ApproverOracleAddress:  suite.approverAccAddr.String(),
		TargetUniqueId:         suite.upgradeUniqueID,
		TargetOracleAddress:    suite.oracleAccAddr.String(),
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	approveOracleRegistrationBz, err := suite.Cdc.Marshaler.Marshal(approveOracleUpgrade)
	suite.Require().NoError(err)
	signature, err := suite.oraclePrivKey.Sign(approveOracleRegistrationBz)
	suite.Require().NoError(err)

	msgApproveOracleUpgrade := types.NewMsgApproveOracleUpgrade(approveOracleUpgrade, signature.Serialize())

	err = suite.OracleKeeper.ApproveOracleUpgrade(ctx, msgApproveOracleUpgrade)
	suite.Require().NoError(err)

	// do second approve
	secondApprove := &types.ApprovalSharingOracleKey{
		ApproverUniqueId:       suite.currentUniqueID,
		ApproverOracleAddress:  suite.oracleAccAddr.String(),
		TargetUniqueId:         suite.upgradeUniqueID,
		TargetOracleAddress:    suite.oracleAccAddr.String(),
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	secondOracleUpgradeBz, err := suite.Cdc.Marshaler.Marshal(secondApprove)
	suite.Require().NoError(err)
	secondSig, err := suite.oraclePrivKey.Sign(secondOracleUpgradeBz)
	suite.Require().NoError(err)

	secondMsgApproveOracleUpgrade := types.NewMsgApproveOracleUpgrade(secondApprove, secondSig.Serialize())
	err = suite.OracleKeeper.ApproveOracleUpgrade(ctx, secondMsgApproveOracleUpgrade)
	suite.Require().Error(types.ErrApproveOracleUpgrade)
	suite.Require().ErrorContains(err, "already approved oracle upgrade")
}
