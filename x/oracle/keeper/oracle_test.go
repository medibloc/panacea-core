package keeper_test

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
)

type oracleTestSuite struct {
	testsuite.TestSuite

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

	endpoint                      string
	oracleCommissionRate          sdk.Dec
	oracleCommissionMaxRate       sdk.Dec
	oracleCommissionMaxChangeRate sdk.Dec

	newEndpoint             string
	newOracleCommissionRate sdk.Dec
}

func TestOracleTestSuite(t *testing.T) {
	suite.Run(t, new(oracleTestSuite))
}

func (suite *oracleTestSuite) BeforeTest(_, _ string) {
	suite.uniqueID = "uniqueID"
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

	suite.endpoint = "https://my-validator.org"
	suite.oracleCommissionRate = sdk.NewDecWithPrec(1, 1)
	suite.oracleCommissionMaxRate = sdk.NewDecWithPrec(2, 1)
	suite.oracleCommissionMaxChangeRate = sdk.NewDecWithPrec(1, 2)

	suite.newEndpoint = "https://my-validator2.org"
	suite.newOracleCommissionRate = sdk.NewDecWithPrec(11, 2)

	suite.OracleKeeper.SetParams(suite.Ctx, types.Params{
		OraclePublicKey:          base64.StdEncoding.EncodeToString(suite.oraclePubKey.SerializeCompressed()),
		OraclePubKeyRemoteReport: "",
		UniqueId:                 suite.uniqueID,
	})
}

func (suite *oracleTestSuite) TestRegisterOracleSuccess() {
	ctx := suite.Ctx

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:                      suite.uniqueID,
		OracleAddress:                 suite.oracleAccAddr.String(),
		NodePubKey:                    suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport:        suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:            suite.trustedBlockHeight,
		TrustedBlockHash:              suite.trustedBlockHash,
		Endpoint:                      suite.endpoint,
		OracleCommissionRate:          suite.oracleCommissionRate,
		OracleCommissionMaxRate:       suite.oracleCommissionMaxRate,
		OracleCommissionMaxChangeRate: suite.oracleCommissionMaxChangeRate,
	}

	err := suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().NoError(err)

	events := suite.Ctx.EventManager().Events()
	suite.Require().Equal(1, len(events))
	suite.Require().Equal(types.EventTypeRegistration, events[0].Type)

	eventVoteAttributes := events[0].Attributes
	suite.Require().Equal(2, len(eventVoteAttributes))
	suite.Require().Equal(types.AttributeKeyUniqueID, string(eventVoteAttributes[0].Key))
	suite.Require().Equal(suite.uniqueID, string(eventVoteAttributes[0].Value))
	suite.Require().Equal(types.AttributeKeyOracleAddress, string(eventVoteAttributes[1].Key))
	suite.Require().Equal(suite.oracleAccAddr.String(), string(eventVoteAttributes[1].Value))

	oracleFromKeeper, err := suite.OracleKeeper.GetOracleRegistration(ctx, suite.uniqueID, suite.oracleAccAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(suite.uniqueID, oracleFromKeeper.UniqueId)
	suite.Require().Equal(suite.oracleAccAddr.String(), oracleFromKeeper.OracleAddress)
	suite.Require().Equal(suite.nodePubKey.SerializeCompressed(), oracleFromKeeper.NodePubKey)
	suite.Require().Equal(suite.nodePubKeyRemoteReport, oracleFromKeeper.NodePubKeyRemoteReport)
	suite.Require().Equal(suite.trustedBlockHeight, oracleFromKeeper.TrustedBlockHeight)
	suite.Require().Equal(suite.trustedBlockHash, oracleFromKeeper.TrustedBlockHash)
	suite.Require().Equal(suite.endpoint, oracleFromKeeper.Endpoint)
	suite.Require().Equal(suite.oracleCommissionRate, oracleFromKeeper.OracleCommissionRate)
	suite.Require().Equal(suite.oracleCommissionMaxRate, oracleFromKeeper.OracleCommissionMaxRate)
	suite.Require().Equal(suite.oracleCommissionMaxChangeRate, oracleFromKeeper.OracleCommissionMaxChangeRate)
}

func (suite *oracleTestSuite) TestRegisterOracleFailedValidateToMsgOracleRegistration() {
	ctx := suite.Ctx

	msgRegisterOracle := &types.MsgRegisterOracle{
		OracleAddress:                 suite.oracleAccAddr.String(),
		NodePubKey:                    suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport:        suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:            suite.trustedBlockHeight,
		TrustedBlockHash:              suite.trustedBlockHash,
		Endpoint:                      suite.endpoint,
		OracleCommissionRate:          suite.oracleCommissionRate,
		OracleCommissionMaxRate:       suite.oracleCommissionMaxRate,
		OracleCommissionMaxChangeRate: suite.oracleCommissionMaxChangeRate,
	}

	err := suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, sdkerrors.ErrInvalidRequest)
	suite.Require().ErrorContains(err, "uniqueID is empty")

	msgRegisterOracle.UniqueId = suite.uniqueID
	msgRegisterOracle.NodePubKey = nil
	err = suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, sdkerrors.ErrInvalidRequest)
	suite.Require().ErrorContains(err, "node public key is empty")

	msgRegisterOracle.NodePubKey = []byte("invalidNodePubKey")
	err = suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, sdkerrors.ErrInvalidRequest)
	suite.Require().ErrorContains(err, "invalid node public key")

	msgRegisterOracle.NodePubKey = suite.nodePubKey.SerializeCompressed()
	msgRegisterOracle.NodePubKeyRemoteReport = nil
	err = suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, sdkerrors.ErrInvalidRequest)
	suite.Require().ErrorContains(err, "remote report of node public key is empty")

	msgRegisterOracle.NodePubKeyRemoteReport = suite.nodePubKeyRemoteReport
	msgRegisterOracle.TrustedBlockHeight = 0
	err = suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, sdkerrors.ErrInvalidRequest)
	suite.Require().ErrorContains(err, "trusted block height must be greater than zero")

	msgRegisterOracle.TrustedBlockHeight = suite.trustedBlockHeight
	msgRegisterOracle.TrustedBlockHash = nil
	err = suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, sdkerrors.ErrInvalidRequest)
	suite.Require().ErrorContains(err, "trusted block hash should not be nil")

	msgRegisterOracle.TrustedBlockHash = suite.trustedBlockHash
	msgRegisterOracle.OracleCommissionRate = sdk.NewInt(-1).ToDec()
	err = suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, sdkerrors.ErrInvalidRequest)
	suite.Require().ErrorContains(err, "oracleCommissionRate must be between 0 and 1")

	msgRegisterOracle.OracleCommissionRate = sdk.NewInt(2).ToDec()
	err = suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, sdkerrors.ErrInvalidRequest)
	suite.Require().ErrorContains(err, "oracleCommissionRate must be between 0 and 1")

	events := suite.Ctx.EventManager().Events()
	suite.Require().Equal(0, len(events))
}

func (suite *oracleTestSuite) TestRegisterOracleAlreadyExistOracle() {
	ctx := suite.Ctx

	oracle := types.NewOracle(suite.oracleAccAddr.String(), suite.uniqueID, suite.endpoint, suite.oracleCommissionRate, suite.oracleCommissionMaxRate, suite.oracleCommissionMaxChangeRate, ctx.BlockTime())
	err := suite.OracleKeeper.SetOracle(ctx, oracle)
	suite.Require().NoError(err)

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:                      suite.uniqueID,
		OracleAddress:                 suite.oracleAccAddr.String(),
		NodePubKey:                    suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport:        suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:            suite.trustedBlockHeight,
		TrustedBlockHash:              suite.trustedBlockHash,
		Endpoint:                      suite.endpoint,
		OracleCommissionRate:          suite.oracleCommissionRate,
		OracleCommissionMaxRate:       suite.oracleCommissionMaxRate,
		OracleCommissionMaxChangeRate: suite.oracleCommissionMaxChangeRate,
	}

	err = suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, types.ErrOracleRegistration)
	suite.Require().ErrorContains(err, fmt.Sprintf("already registered oracle. address(%s)", msgRegisterOracle.OracleAddress))
}

func (suite *oracleTestSuite) TestRegisterOracleNotSameUniqueID() {
	ctx := suite.Ctx

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:                      "wrongUniqueID",
		OracleAddress:                 suite.oracleAccAddr.String(),
		NodePubKey:                    suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport:        suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:            suite.trustedBlockHeight,
		TrustedBlockHash:              suite.trustedBlockHash,
		Endpoint:                      suite.endpoint,
		OracleCommissionRate:          suite.oracleCommissionRate,
		OracleCommissionMaxRate:       suite.oracleCommissionMaxRate,
		OracleCommissionMaxChangeRate: suite.oracleCommissionMaxChangeRate,
	}

	err := suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, types.ErrOracleRegistration)
	suite.Require().ErrorContains(err, "is not match the currently active uniqueID")
}

func (suite *oracleTestSuite) TestApproveOracleRegistrationSuccess() {
	ctx := suite.Ctx

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:                      suite.uniqueID,
		OracleAddress:                 suite.oracleAccAddr.String(),
		NodePubKey:                    suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport:        suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:            suite.trustedBlockHeight,
		TrustedBlockHash:              suite.trustedBlockHash,
		Endpoint:                      suite.endpoint,
		OracleCommissionRate:          suite.oracleCommissionRate,
		OracleCommissionMaxRate:       suite.oracleCommissionMaxRate,
		OracleCommissionMaxChangeRate: suite.oracleCommissionMaxChangeRate,
	}

	err := suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().NoError(err)

	encryptedOraclePrivKey, err := btcec.Encrypt(suite.nodePubKey, suite.oraclePrivKey.Serialize())
	suite.Require().NoError(err)

	approveOracleRegistration := &types.ApproveOracleRegistration{
		UniqueId:               suite.uniqueID,
		ApproverOracleAddress:  suite.oracleAccAddr.String(),
		TargetOracleAddress:    suite.oracleAccAddr.String(),
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	approveOracleRegistrationBz, err := suite.Cdc.Marshaler.Marshal(approveOracleRegistration)
	suite.Require().NoError(err)
	signature, err := suite.oraclePrivKey.Sign(approveOracleRegistrationBz)
	suite.Require().NoError(err)

	msgApproveOracleRegistration := types.NewMsgApproveOracleRegistration(approveOracleRegistration, signature.Serialize())

	err = suite.OracleKeeper.ApproveOracleRegistration(ctx, msgApproveOracleRegistration)
	suite.Require().NoError(err)

	events := suite.Ctx.EventManager().Events()
	suite.Require().Equal(2, len(events))
	suite.Require().Equal(types.EventTypeApproveOracleRegistration, events[1].Type)

	eventVoteAttributes := events[1].Attributes
	suite.Require().Equal(1, len(eventVoteAttributes))
	suite.Require().Equal(types.AttributeKeyOracleAddress, string(eventVoteAttributes[0].Key))
	suite.Require().Equal(suite.oracleAccAddr.String(), string(eventVoteAttributes[0].Value))

	getOracle, err := suite.OracleKeeper.GetOracle(ctx, suite.oracleAccAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(suite.uniqueID, getOracle.UniqueId)
	suite.Require().Equal(suite.oracleAccAddr.String(), getOracle.OracleAddress)
	suite.Require().Equal(suite.endpoint, getOracle.Endpoint)
	suite.Require().Equal(suite.oracleCommissionRate, getOracle.OracleCommissionRate)

	getOracleRegistration, err := suite.OracleKeeper.GetOracleRegistration(suite.Ctx, suite.uniqueID, suite.oracleAccAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(approveOracleRegistration.EncryptedOraclePrivKey, getOracleRegistration.EncryptedOraclePrivKey)
}

func (suite *oracleTestSuite) TestApproveOracleRegistrationFailedInvalidUniqueID() {
	ctx := suite.Ctx

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:                      suite.uniqueID,
		OracleAddress:                 suite.oracleAccAddr.String(),
		NodePubKey:                    suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport:        suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:            suite.trustedBlockHeight,
		TrustedBlockHash:              suite.trustedBlockHash,
		Endpoint:                      suite.endpoint,
		OracleCommissionRate:          suite.oracleCommissionRate,
		OracleCommissionMaxRate:       suite.oracleCommissionMaxRate,
		OracleCommissionMaxChangeRate: suite.oracleCommissionMaxChangeRate,
	}

	err := suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().NoError(err)

	encryptedOraclePrivKey, err := btcec.Encrypt(suite.nodePubKey, suite.oraclePrivKey.Serialize())
	suite.Require().NoError(err)

	approveOracleRegistration := &types.ApproveOracleRegistration{
		UniqueId:               "uniqueID2",
		ApproverOracleAddress:  suite.oracleAccAddr.String(),
		TargetOracleAddress:    suite.oracleAccAddr.String(),
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}
	approveOracleRegistrationBz, err := suite.Cdc.Marshaler.Marshal(approveOracleRegistration)
	suite.Require().NoError(err)
	signature, err := suite.oraclePrivKey.Sign(approveOracleRegistrationBz)
	suite.Require().NoError(err)

	msgApproveOracleRegistration := types.NewMsgApproveOracleRegistration(approveOracleRegistration, signature.Serialize())

	err = suite.OracleKeeper.ApproveOracleRegistration(ctx, msgApproveOracleRegistration)
	suite.Require().Error(err, types.ErrInvalidUniqueID)
}

func (suite *oracleTestSuite) TestApproveOracleRegistrationFailedInvalidSignature() {
	ctx := suite.Ctx

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:                      suite.uniqueID,
		OracleAddress:                 suite.oracleAccAddr.String(),
		NodePubKey:                    suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport:        suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:            suite.trustedBlockHeight,
		TrustedBlockHash:              suite.trustedBlockHash,
		Endpoint:                      suite.endpoint,
		OracleCommissionRate:          suite.oracleCommissionRate,
		OracleCommissionMaxRate:       suite.oracleCommissionMaxRate,
		OracleCommissionMaxChangeRate: suite.oracleCommissionMaxChangeRate,
	}

	err := suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().NoError(err)

	encryptedOraclePrivKey, err := btcec.Encrypt(suite.nodePubKey, suite.oraclePrivKey.Serialize())
	suite.Require().NoError(err)

	approveOracleRegistration := &types.ApproveOracleRegistration{
		UniqueId:               "uniqueID",
		ApproverOracleAddress:  suite.oracleAccAddr.String(),
		TargetOracleAddress:    suite.oracleAccAddr.String(),
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}
	approveOracleRegistrationBz, err := suite.Cdc.Marshaler.Marshal(approveOracleRegistration)
	suite.Require().NoError(err)
	newPrivateKey, err := btcec.NewPrivateKey(btcec.S256())
	suite.Require().NoError(err)
	signature, err := newPrivateKey.Sign(approveOracleRegistrationBz)
	suite.Require().NoError(err)

	msgApproveOracleRegistration := types.NewMsgApproveOracleRegistration(approveOracleRegistration, signature.Serialize())

	err = suite.OracleKeeper.ApproveOracleRegistration(ctx, msgApproveOracleRegistration)
	suite.Require().Error(err, "failed to signature validation")
}

func (suite *oracleTestSuite) TestApproveOracleRegistrationFailedAlreadyExistOracle() {
	ctx := suite.Ctx

	oracle := types.NewOracle(suite.oracleAccAddr.String(), suite.uniqueID, suite.endpoint, suite.oracleCommissionRate, suite.oracleCommissionMaxRate, suite.oracleCommissionMaxChangeRate, ctx.BlockTime())
	err := suite.OracleKeeper.SetOracle(ctx, oracle)
	suite.Require().NoError(err)

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:                      suite.uniqueID,
		OracleAddress:                 suite.oracleAccAddr.String(),
		NodePubKey:                    suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport:        suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:            suite.trustedBlockHeight,
		TrustedBlockHash:              suite.trustedBlockHash,
		Endpoint:                      suite.endpoint,
		OracleCommissionRate:          suite.oracleCommissionRate,
		OracleCommissionMaxRate:       suite.oracleCommissionMaxRate,
		OracleCommissionMaxChangeRate: suite.oracleCommissionMaxChangeRate,
	}

	oracleRegistration := types.NewOracleRegistration(msgRegisterOracle)

	err = suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	encryptedOraclePrivKey, err := btcec.Encrypt(suite.nodePubKey, suite.oraclePrivKey.Serialize())
	suite.Require().NoError(err)

	approveOracleRegistration := &types.ApproveOracleRegistration{
		UniqueId:               "uniqueID",
		ApproverOracleAddress:  suite.oracleAccAddr.String(),
		TargetOracleAddress:    suite.oracleAccAddr.String(),
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}
	approveOracleRegistrationBz, err := suite.Cdc.Marshaler.Marshal(approveOracleRegistration)
	suite.Require().NoError(err)
	signature, err := suite.oraclePrivKey.Sign(approveOracleRegistrationBz)
	suite.Require().NoError(err)

	msgApproveOracleRegistration := types.NewMsgApproveOracleRegistration(approveOracleRegistration, signature.Serialize())

	err = suite.OracleKeeper.ApproveOracleRegistration(ctx, msgApproveOracleRegistration)
	suite.Require().Error(err, types.ErrOracleRegistration)
	suite.Require().ErrorContains(err, fmt.Sprintf("already registered oracle. address(%s)", msgRegisterOracle.OracleAddress))
}

func (suite *oracleTestSuite) TestUpdateOracleInfoSuccess() {
	ctx := suite.Ctx

	oracle := types.NewOracle(suite.oracleAccAddr.String(), suite.uniqueID, suite.endpoint, suite.oracleCommissionRate, suite.oracleCommissionMaxRate, suite.oracleCommissionMaxChangeRate, ctx.BlockTime())
	err := suite.OracleKeeper.SetOracle(ctx, oracle)
	suite.Require().NoError(err)

	msgUpdateOracleInfo := &types.MsgUpdateOracleInfo{
		OracleAddress:        suite.oracleAccAddr.String(),
		Endpoint:             suite.newEndpoint,
		OracleCommissionRate: suite.newOracleCommissionRate,
	}
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(30 * time.Hour))
	err = suite.OracleKeeper.UpdateOracleInfo(ctx, msgUpdateOracleInfo)
	suite.Require().NoError(err)

	getOracle, err := suite.OracleKeeper.GetOracle(ctx, suite.oracleAccAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(suite.newEndpoint, getOracle.Endpoint)
	suite.Require().Equal(suite.newOracleCommissionRate, getOracle.OracleCommissionRate)
	suite.Require().Equal(ctx.BlockTime(), getOracle.UpdateTime)
}

func (suite *oracleTestSuite) TestUpdateOracleInfoSuccessWithNoCommissionChange() {
	ctx := suite.Ctx

	oracle := types.NewOracle(suite.oracleAccAddr.String(), suite.uniqueID, suite.endpoint, suite.oracleCommissionRate, suite.oracleCommissionMaxRate, suite.oracleCommissionMaxChangeRate, ctx.BlockTime())
	err := suite.OracleKeeper.SetOracle(ctx, oracle)
	suite.Require().NoError(err)

	msgUpdateOracleInfo := &types.MsgUpdateOracleInfo{
		OracleAddress:        suite.oracleAccAddr.String(),
		Endpoint:             suite.newEndpoint,
		OracleCommissionRate: suite.oracleCommissionRate,
	}
	err = suite.OracleKeeper.UpdateOracleInfo(ctx, msgUpdateOracleInfo)
	suite.Require().NoError(err)

	getOracle, err := suite.OracleKeeper.GetOracle(ctx, suite.oracleAccAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(suite.newEndpoint, getOracle.Endpoint)
	suite.Require().Equal(suite.oracleCommissionRate, getOracle.OracleCommissionRate)
	suite.Require().Equal(ctx.BlockTime(), getOracle.UpdateTime)
}

func (suite *oracleTestSuite) TestUpdateOracleInfoFailedUpdateTime() {
	ctx := suite.Ctx

	oracle := types.NewOracle(suite.oracleAccAddr.String(), suite.uniqueID, suite.endpoint, suite.oracleCommissionRate, suite.oracleCommissionMaxRate, suite.oracleCommissionMaxChangeRate, ctx.BlockTime())
	err := suite.OracleKeeper.SetOracle(ctx, oracle)
	suite.Require().NoError(err)

	msgUpdateOracleInfo := &types.MsgUpdateOracleInfo{
		OracleAddress:        suite.oracleAccAddr.String(),
		Endpoint:             suite.newEndpoint,
		OracleCommissionRate: suite.newOracleCommissionRate,
	}

	err = suite.OracleKeeper.UpdateOracleInfo(ctx, msgUpdateOracleInfo)
	suite.Require().Error(err, types.ErrUpdateOracle)
	suite.Require().ErrorContains(err, "commission cannot be changed more than once in 24h")
}

func (suite *oracleTestSuite) TestUpdateOracleInfoFailedGTMaxChangeRate() {
	ctx := suite.Ctx

	oracle := types.NewOracle(suite.oracleAccAddr.String(), suite.uniqueID, suite.endpoint, suite.oracleCommissionRate, suite.oracleCommissionMaxRate, suite.oracleCommissionMaxChangeRate, ctx.BlockTime())
	err := suite.OracleKeeper.SetOracle(ctx, oracle)
	suite.Require().NoError(err)

	msgUpdateOracleInfo := &types.MsgUpdateOracleInfo{
		OracleAddress:        suite.oracleAccAddr.String(),
		Endpoint:             suite.newEndpoint,
		OracleCommissionRate: sdk.NewDecWithPrec(12, 2),
	}
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(30 * time.Hour))
	err = suite.OracleKeeper.UpdateOracleInfo(ctx, msgUpdateOracleInfo)
	suite.Require().Error(err, types.ErrUpdateOracle)
	suite.Require().ErrorContains(err, "commission cannot be changed more than max change rate")
}

func (suite *oracleTestSuite) TestUpdateOracleInfoFailedGTMaxRate() {
	ctx := suite.Ctx

	oracle := types.NewOracle(suite.oracleAccAddr.String(), suite.uniqueID, suite.endpoint, suite.oracleCommissionRate, suite.oracleCommissionMaxRate, sdk.NewDecWithPrec(3, 1), ctx.BlockTime())
	err := suite.OracleKeeper.SetOracle(ctx, oracle)
	suite.Require().NoError(err)

	msgUpdateOracleInfo := &types.MsgUpdateOracleInfo{
		OracleAddress:        suite.oracleAccAddr.String(),
		Endpoint:             suite.newEndpoint,
		OracleCommissionRate: sdk.NewDecWithPrec(3, 1),
	}
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(30 * time.Hour))
	err = suite.OracleKeeper.UpdateOracleInfo(ctx, msgUpdateOracleInfo)
	suite.Require().Error(err, types.ErrUpdateOracle)
	suite.Require().ErrorContains(err, "commission cannot be more than the max rate")
}

func (suite *oracleTestSuite) TestUpdateOracleInfoFailedNegativeRate() {
	ctx := suite.Ctx

	oracle := types.NewOracle(suite.oracleAccAddr.String(), suite.uniqueID, suite.endpoint, suite.oracleCommissionRate, suite.oracleCommissionMaxRate, suite.oracleCommissionMaxChangeRate, ctx.BlockTime())
	err := suite.OracleKeeper.SetOracle(ctx, oracle)
	suite.Require().NoError(err)

	msgUpdateOracleInfo := &types.MsgUpdateOracleInfo{
		OracleAddress:        suite.oracleAccAddr.String(),
		Endpoint:             suite.newEndpoint,
		OracleCommissionRate: sdk.NewDecWithPrec(-1, 1),
	}
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(30 * time.Hour))
	err = suite.OracleKeeper.UpdateOracleInfo(ctx, msgUpdateOracleInfo)
	suite.Require().Error(err, types.ErrUpdateOracle)
	suite.Require().ErrorContains(err, "commission must be positive")
}
