package keeper_test

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"

	"github.com/stretchr/testify/suite"

	"github.com/medibloc/panacea-core/v2/types/testsuite"
)

type queryOracleTestSuite struct {
	testsuite.TestSuite

	uniqueID string

	oracleAccPrivKey cryptotypes.PrivKey
	oracleAccPubKey  cryptotypes.PubKey
	oracleAccAddr    sdk.AccAddress

	newOracleAccPrivKey cryptotypes.PrivKey
	newOracleAccPubKey  cryptotypes.PubKey
	newOracleAccAddr    sdk.AccAddress

	nodePrivKey *btcec.PrivateKey
	nodePubKey  *btcec.PublicKey
}

func TestQueryOracleTestSuite(t *testing.T) {
	suite.Run(t, new(queryOracleTestSuite))
}

func (suite *queryOracleTestSuite) BeforeTest(_, _ string) {
	suite.uniqueID = "correctUniqueID"

	suite.oracleAccPrivKey = secp256k1.GenPrivKey()
	suite.oracleAccPubKey = suite.oracleAccPrivKey.PubKey()
	suite.oracleAccAddr = sdk.AccAddress(suite.oracleAccPubKey.Address())

	suite.newOracleAccPrivKey = secp256k1.GenPrivKey()
	suite.newOracleAccPubKey = suite.newOracleAccPrivKey.PubKey()
	suite.newOracleAccAddr = sdk.AccAddress(suite.newOracleAccPubKey.Address())

	suite.nodePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.nodePubKey = suite.nodePrivKey.PubKey()
}

func (suite queryOracleTestSuite) makeNewOracleRegistration() *types.OracleRegistration {
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

func (suite queryOracleTestSuite) TestOracleRegistration() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	newOracleRegistration := suite.makeNewOracleRegistration()

	err := oracleKeeper.SetOracleRegistration(ctx, newOracleRegistration)
	suite.Require().NoError(err)

	req := types.QueryOracleRegistrationRequest{
		UniqueId: newOracleRegistration.UniqueId,
		Address:  newOracleRegistration.Address,
	}

	res, err := oracleKeeper.OracleRegistration(sdk.WrapSDKContext(ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Equal(newOracleRegistration.UniqueId, res.OracleRegistration.UniqueId)
	suite.Require().Equal(newOracleRegistration.Address, res.OracleRegistration.Address)
	suite.Require().Equal(newOracleRegistration.NodePubKey, res.OracleRegistration.NodePubKey)
	suite.Require().Equal(newOracleRegistration.NodePubKeyRemoteReport, res.OracleRegistration.NodePubKeyRemoteReport)
	suite.Require().Equal(newOracleRegistration.TrustedBlockHash, res.OracleRegistration.TrustedBlockHash)
	suite.Require().Equal(newOracleRegistration.TrustedBlockHeight, res.OracleRegistration.TrustedBlockHeight)
}

func (suite queryOracleTestSuite) TestOracleParams() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	oracleKeeper.SetParams(ctx, types.DefaultParams())

	req := types.QueryOracleParamsRequest{}
	res, err := oracleKeeper.Params(sdk.WrapSDKContext(ctx), &req)
	suite.Require().NoError(err)

	suite.Require().Equal(types.DefaultParams(), *res.Params)
}
