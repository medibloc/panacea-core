package keeper_test

import (
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
)

type queryOracleTestSuite struct {
	testsuite.TestSuite

	uniqueID string

	oracleAccPrivKey     cryptotypes.PrivKey
	oracleAccPubKey      cryptotypes.PubKey
	oracleAccAddr        sdk.AccAddress
	oracleEndpoint       string
	oracleCommissionRate sdk.Dec

	oracle2AccPrivKey     cryptotypes.PrivKey
	oracle2AccPubKey      cryptotypes.PubKey
	oracle2AccAddr        sdk.AccAddress
	oracle2Endpoint       string
	oracle2CommissionRate sdk.Dec

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
	suite.oracleEndpoint = "https://my-validator.org"
	suite.oracleCommissionRate = sdk.NewDecWithPrec(1, 1)

	suite.oracle2AccPrivKey = secp256k1.GenPrivKey()
	suite.oracle2AccPubKey = suite.oracle2AccPrivKey.PubKey()
	suite.oracle2AccAddr = sdk.AccAddress(suite.oracle2AccPubKey.Address())
	suite.oracle2Endpoint = "https://my-validator2.org"
	suite.oracle2CommissionRate = sdk.NewDecWithPrec(1, 2)

	suite.nodePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.nodePubKey = suite.nodePrivKey.PubKey()
}

func (suite *queryOracleTestSuite) TestOracles() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	oracle := types.NewOracle(suite.oracleAccAddr.String(), suite.uniqueID, suite.oracleEndpoint, suite.oracleCommissionRate)
	err := oracleKeeper.SetOracle(ctx, oracle)
	suite.Require().NoError(err)

	oracle2 := types.NewOracle(suite.oracle2AccAddr.String(), suite.uniqueID, suite.oracle2Endpoint, suite.oracle2CommissionRate)
	err = oracleKeeper.SetOracle(ctx, oracle2)
	suite.Require().NoError(err)

	req := types.QueryOraclesRequest{
		Pagination: &query.PageRequest{},
	}

	res, err := oracleKeeper.Oracles(sdk.WrapSDKContext(ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(res.Oracles))
	for _, oracle := range res.Oracles {
		switch oracle.OracleAddress {
		case suite.oracleAccAddr.String():
			suite.Require().Equal(suite.uniqueID, oracle.UniqueId)
			suite.Require().Equal(suite.oracleEndpoint, oracle.Endpoint)
			suite.Require().Equal(suite.oracleCommissionRate, oracle.OracleCommissionRate)
		case suite.oracle2AccAddr.String():
			suite.Require().Equal(suite.uniqueID, oracle.UniqueId)
			suite.Require().Equal(suite.oracle2Endpoint, oracle.Endpoint)
			suite.Require().Equal(suite.oracle2CommissionRate, oracle.OracleCommissionRate)
		default:
			panic("not found oracle address. address: " + oracle.OracleAddress)
		}
	}
}

func (suite *queryOracleTestSuite) TestOracle() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	oracle := types.NewOracle(suite.oracleAccAddr.String(), suite.uniqueID, suite.oracleEndpoint, suite.oracleCommissionRate)
	err := oracleKeeper.SetOracle(ctx, oracle)
	suite.Require().NoError(err)

	req := types.QueryOracleRequest{
		OracleAddress: suite.oracleAccAddr.String(),
	}
	res, err := oracleKeeper.Oracle(sdk.WrapSDKContext(ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Equal(oracle, res.Oracle)
}

func (suite *queryOracleTestSuite) TestOracleRegistrations() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	remoteReport := []byte("nodePubKeyRemoteReport")
	trustedBlockHash := []byte("hash")

	oracleRegistration := &types.OracleRegistration{
		UniqueId:               suite.uniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: remoteReport,
		TrustedBlockHeight:     10,
		TrustedBlockHash:       trustedBlockHash,
		Endpoint:               suite.oracleEndpoint,
		OracleCommissionRate:   suite.oracleCommissionRate,
	}

	oracleRegistration2 := &types.OracleRegistration{
		UniqueId:               suite.uniqueID,
		OracleAddress:          suite.oracle2AccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: remoteReport,
		TrustedBlockHeight:     10,
		TrustedBlockHash:       trustedBlockHash,
		Endpoint:               suite.oracle2Endpoint,
		OracleCommissionRate:   suite.oracle2CommissionRate,
	}

	err := oracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)
	err = oracleKeeper.SetOracleRegistration(ctx, oracleRegistration2)
	suite.Require().NoError(err)

	req := types.QueryOracleRegistrationsRequest{
		UniqueId:   suite.uniqueID,
		Pagination: &query.PageRequest{},
	}
	res, err := oracleKeeper.OracleRegistrations(sdk.WrapSDKContext(ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(res.OracleRegistrations))
	for _, oracleRegistration := range res.OracleRegistrations {
		switch oracleRegistration.OracleAddress {
		case suite.oracleAccAddr.String():
			suite.Require().Equal(suite.uniqueID, oracleRegistration.UniqueId)
			suite.Require().Equal(suite.nodePubKey.SerializeCompressed(), oracleRegistration.NodePubKey)
			suite.Require().Equal(remoteReport, oracleRegistration.NodePubKeyRemoteReport)
			suite.Require().Equal(int64(10), oracleRegistration.TrustedBlockHeight)
			suite.Require().Equal(trustedBlockHash, oracleRegistration.TrustedBlockHash)
			suite.Require().Equal(suite.oracleEndpoint, oracleRegistration.Endpoint)
			suite.Require().Equal(suite.oracleCommissionRate, oracleRegistration.OracleCommissionRate)
		case suite.oracle2AccAddr.String():
			suite.Require().Equal(suite.uniqueID, oracleRegistration.UniqueId)
			suite.Require().Equal(suite.nodePubKey.SerializeCompressed(), oracleRegistration.NodePubKey)
			suite.Require().Equal(remoteReport, oracleRegistration.NodePubKeyRemoteReport)
			suite.Require().Equal(int64(10), oracleRegistration.TrustedBlockHeight)
			suite.Require().Equal(trustedBlockHash, oracleRegistration.TrustedBlockHash)
			suite.Require().Equal(suite.oracle2Endpoint, oracleRegistration.Endpoint)
			suite.Require().Equal(suite.oracle2CommissionRate, oracleRegistration.OracleCommissionRate)
		default:
			panic("not found oracle address. address: " + oracleRegistration.OracleAddress)
		}
	}
}

func (suite *queryOracleTestSuite) TestOracleRegistration() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	oracleRegistration := &types.OracleRegistration{
		UniqueId:               suite.uniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: []byte("nodePubKeyRemoteReport"),
		TrustedBlockHeight:     10,
		TrustedBlockHash:       []byte("hash"),
		Endpoint:               suite.oracleEndpoint,
		OracleCommissionRate:   suite.oracleCommissionRate,
	}
	err := oracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

}
