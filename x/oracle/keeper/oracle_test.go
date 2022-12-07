package keeper_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	endpoint             string
	oracleCommissionRate sdk.Dec
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

	suite.OracleKeeper.SetParams(suite.Ctx, types.Params{
		OraclePublicKey:          base64.StdEncoding.EncodeToString(suite.oraclePubKey.SerializeCompressed()),
		OraclePubKeyRemoteReport: "",
		UniqueId:                 suite.uniqueID,
	})
}

func (suite *oracleTestSuite) TestApproveOracleRegistrationSuccess() {
	ctx := suite.Ctx

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:               suite.uniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     suite.trustedBlockHeight,
		TrustedBlockHash:       suite.trustedBlockHash,
		Endpoint:               suite.endpoint,
		OracleCommissionRate:   suite.oracleCommissionRate,
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
	suite.Require().Equal(2, len(eventVoteAttributes))
	suite.Require().Equal(types.AttributeKeyOracleAddress, string(eventVoteAttributes[0].Key))
	suite.Require().Equal(suite.oracleAccAddr.String(), string(eventVoteAttributes[0].Value))
	suite.Require().Equal(types.AttributeKeyEncryptedOraclePrivKey, string(eventVoteAttributes[1].Key))
	suite.Require().Equal(string(encryptedOraclePrivKey), string(eventVoteAttributes[1].Value))

	getOracle, err := suite.OracleKeeper.GetOracle(ctx, suite.oracleAccAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(suite.uniqueID, getOracle.UniqueId)
	suite.Require().Equal(suite.oracleAccAddr.String(), getOracle.OracleAddress)
	suite.Require().Equal(suite.endpoint, getOracle.Endpoint)
	suite.Require().Equal(suite.oracleCommissionRate, getOracle.OracleCommissionRate)
}

func (suite *oracleTestSuite) TestApproveOracleRegistrationFailedInvalidUniqueID() {
	ctx := suite.Ctx

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:               suite.uniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     suite.trustedBlockHeight,
		TrustedBlockHash:       suite.trustedBlockHash,
		Endpoint:               suite.endpoint,
		OracleCommissionRate:   suite.oracleCommissionRate,
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
	suite.Require().Error(err, types.ErrInvalidUniqueId)
}

func (suite *oracleTestSuite) TestApproveOracleRegistrationFailedInvalidSignature() {
	ctx := suite.Ctx

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:               suite.uniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     suite.trustedBlockHeight,
		TrustedBlockHash:       suite.trustedBlockHash,
		Endpoint:               suite.endpoint,
		OracleCommissionRate:   suite.oracleCommissionRate,
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

	oracle := types.NewOracle(suite.oracleAccAddr.String(), suite.uniqueID, suite.endpoint, suite.oracleCommissionRate)
	err := suite.OracleKeeper.SetOracle(ctx, oracle)
	suite.Require().NoError(err)

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:               suite.uniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     suite.trustedBlockHeight,
		TrustedBlockHash:       suite.trustedBlockHash,
		Endpoint:               suite.endpoint,
		OracleCommissionRate:   suite.oracleCommissionRate,
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
