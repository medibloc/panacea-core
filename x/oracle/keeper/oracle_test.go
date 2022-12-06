package keeper_test

import (
	"fmt"
	"testing"

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
}

func (suite *oracleTestSuite) TestRegisterOracleSuccess() {
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
}

func (suite *oracleTestSuite) TestRegisterOracleFailedValidateToMsgOracleRegistration() {
	ctx := suite.Ctx

	msgRegisterOracle := &types.MsgRegisterOracle{
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: suite.nodePubKeyRemoteReport,
		TrustedBlockHeight:     suite.trustedBlockHeight,
		TrustedBlockHash:       suite.trustedBlockHash,
		Endpoint:               suite.endpoint,
		OracleCommissionRate:   suite.oracleCommissionRate,
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
	msgRegisterOracle.Endpoint = ""
	err = suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, sdkerrors.ErrInvalidRequest)
	suite.Require().ErrorContains(err, "endpoint is empty")

	msgRegisterOracle.Endpoint = suite.endpoint
	msgRegisterOracle.OracleCommissionRate = sdk.NewInt(-1).ToDec()
	err = suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, sdkerrors.ErrInvalidRequest)
	suite.Require().ErrorContains(err, "oracle commission rate cannot be negative")

	events := suite.Ctx.EventManager().Events()
	suite.Require().Equal(0, len(events))
}

func (suite *oracleTestSuite) TestRegisterOracleAlreadyExistOracle() {
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

	err = suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, types.ErrOracleRegistration)
	suite.Require().ErrorContains(err, fmt.Sprintf("already registered oracle. address(%s)", msgRegisterOracle.OracleAddress))
}
