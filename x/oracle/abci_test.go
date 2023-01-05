package oracle_test

import (
	"encoding/base64"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/oracle"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
)

type abciTestSuite struct {
	testsuite.TestSuite

	oracleAccPrivKey cryptotypes.PrivKey
	oracleAccPubKey  cryptotypes.PubKey
	oracleAccAddr    sdk.AccAddress

	oracle2AccPrivKey cryptotypes.PrivKey
	oracle2AccPubKey  cryptotypes.PubKey
	oracle2AccAddr    sdk.AccAddress

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

func TestAbciTestSuite(t *testing.T) {
	suite.Run(t, new(abciTestSuite))
}

func (suite *abciTestSuite) BeforeTest(_, _ string) {

	suite.oracleAccPrivKey = secp256k1.GenPrivKey()
	suite.oracleAccPubKey = suite.oracleAccPrivKey.PubKey()
	suite.oracleAccAddr = sdk.AccAddress(suite.oracleAccPubKey.Address())

	suite.oracle2AccPrivKey = secp256k1.GenPrivKey()
	suite.oracle2AccPubKey = suite.oracle2AccPrivKey.PubKey()
	suite.oracle2AccAddr = sdk.AccAddress(suite.oracle2AccPubKey.Address())

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

func (suite *abciTestSuite) TestOracleUpgradeSuccess() {
	ctx := suite.Ctx
	ctx = ctx.WithBlockHeight(1)

	suite.Require().Equal(suite.currentUniqueID, suite.OracleKeeper.GetParams(ctx).UniqueId)

	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: suite.upgradeUniqueID,
		Height:   10,
	}

	suite.Require().NoError(suite.OracleKeeper.SetOracleUpgradeInfo(ctx, upgradeInfo))

	// set oracles
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

	oracle1 := &types.Oracle{
		OracleAddress:                 suite.oracleAccAddr.String(),
		UniqueId:                      suite.currentUniqueID,
		Endpoint:                      "test.com",
		UpdateTime:                    ctx.BlockTime(),
		OracleCommissionRate:          sdk.NewDecWithPrec(1, 1),
		OracleCommissionMaxRate:       sdk.NewDecWithPrec(2, 1),
		OracleCommissionMaxChangeRate: sdk.NewDecWithPrec(1, 2),
	}

	oracle2 := &types.Oracle{
		OracleAddress:                 suite.oracle2AccAddr.String(),
		UniqueId:                      suite.currentUniqueID,
		Endpoint:                      "test.com",
		UpdateTime:                    ctx.BlockTime(),
		OracleCommissionRate:          sdk.NewDecWithPrec(1, 1),
		OracleCommissionMaxRate:       sdk.NewDecWithPrec(2, 1),
		OracleCommissionMaxChangeRate: sdk.NewDecWithPrec(1, 2),
	}
	suite.Require().NoError(suite.OracleKeeper.SetOracle(ctx, oracle1))
	suite.Require().NoError(suite.OracleKeeper.SetOracle(ctx, oracle2))

	// upgrade oracle1, oracle2
	msgOracleUpgrade := &types.MsgUpgradeOracle{
		UniqueId:               suite.upgradeUniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     int64(1),
		TrustedBlockHash:       []byte("trustedBlockHash"),
	}
	msgOracle2Upgrade := &types.MsgUpgradeOracle{
		UniqueId:               suite.upgradeUniqueID,
		OracleAddress:          suite.oracle2AccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     int64(1),
		TrustedBlockHash:       []byte("trustedBlockHash"),
	}

	suite.Require().NoError(suite.OracleKeeper.UpgradeOracle(ctx, msgOracleUpgrade))
	suite.Require().NoError(suite.OracleKeeper.UpgradeOracle(ctx, msgOracle2Upgrade))

	// approve oracle upgrade requests
	encryptedOraclePrivKey, err := btcec.Encrypt(suite.nodePubKey, suite.oraclePrivKey.Serialize())
	suite.Require().NoError(err)

	approveOracleUpgrade := &types.ApprovalSharingOracleKey{
		ApproverUniqueId:       suite.currentUniqueID,
		ApproverOracleAddress:  suite.approverAccAddr.String(),
		TargetUniqueId:         suite.upgradeUniqueID,
		TargetOracleAddress:    suite.oracleAccAddr.String(),
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	approveOracleUpgradeBz, err := suite.Cdc.Marshaler.Marshal(approveOracleUpgrade)
	suite.Require().NoError(err)
	signature, err := suite.oraclePrivKey.Sign(approveOracleUpgradeBz)
	suite.Require().NoError(err)

	msgApproveOracleUpgrade := types.NewMsgApproveOracleUpgrade(approveOracleUpgrade, signature.Serialize())
	suite.Require().NoError(suite.OracleKeeper.ApproveOracleUpgrade(ctx, msgApproveOracleUpgrade))

	approveOracle2Upgrade := &types.ApprovalSharingOracleKey{
		ApproverUniqueId:       suite.currentUniqueID,
		ApproverOracleAddress:  suite.approverAccAddr.String(),
		TargetUniqueId:         suite.upgradeUniqueID,
		TargetOracleAddress:    suite.oracle2AccAddr.String(),
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	approveOracle2UpgradeBz, err := suite.Cdc.Marshaler.Marshal(approveOracle2Upgrade)
	suite.Require().NoError(err)
	signature, err = suite.oraclePrivKey.Sign(approveOracle2UpgradeBz)
	suite.Require().NoError(err)

	msgApproveOracle2Upgrade := types.NewMsgApproveOracleUpgrade(approveOracle2Upgrade, signature.Serialize())
	suite.Require().NoError(suite.OracleKeeper.ApproveOracleUpgrade(ctx, msgApproveOracle2Upgrade))

	// ApplyUpgrade
	ctx = ctx.WithBlockHeight(10)
	oracle.BeginBlocker(ctx, suite.OracleKeeper)

	// check uniqueID change
	suite.Require().Equal(suite.upgradeUniqueID, suite.OracleKeeper.GetParams(ctx).UniqueId)
	getOracle, err := suite.OracleKeeper.GetOracle(ctx, suite.oracleAccAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(suite.upgradeUniqueID, getOracle.UniqueId)
	getOracle, err = suite.OracleKeeper.GetOracle(ctx, suite.oracle2AccAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(suite.upgradeUniqueID, getOracle.UniqueId)
	getOracle, err = suite.OracleKeeper.GetOracle(ctx, suite.approverAccAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(suite.currentUniqueID, getOracle.UniqueId)
}

func (suite *abciTestSuite) TestOracleUpgradeFailedBeforeReachUpgradeHeight() {
	ctx := suite.Ctx
	ctx = ctx.WithBlockHeight(1)

	suite.Require().Equal(suite.currentUniqueID, suite.OracleKeeper.GetParams(ctx).UniqueId)

	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: suite.upgradeUniqueID,
		Height:   10,
	}

	suite.Require().NoError(suite.OracleKeeper.SetOracleUpgradeInfo(ctx, upgradeInfo))

	ctx = ctx.WithBlockHeight(9)

	oracle.BeginBlocker(ctx, suite.OracleKeeper)

	suite.Require().Equal(suite.currentUniqueID, suite.OracleKeeper.GetParams(ctx).UniqueId)
}
