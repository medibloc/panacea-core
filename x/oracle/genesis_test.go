package oracle_test

import (
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

type genesisTestSuite struct {
	testsuite.TestSuite

	uniqueID string
	endpoint string

	oracleAccPrivKey cryptotypes.PrivKey
	oracleAccPubKey  cryptotypes.PubKey
	oracleAccAddr    sdk.AccAddress

	oracle2AccPrivKey cryptotypes.PrivKey
	oracle2AccPubKey  cryptotypes.PubKey
	oracle2AccAddr    sdk.AccAddress

	nodePrivKey *btcec.PrivateKey
	nodePubKey  *btcec.PublicKey
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}

func (suite *genesisTestSuite) BeforeTest(_, _ string) {
	suite.uniqueID = "uniqueID"
	suite.endpoint = "https://my-validator.org"

	suite.oracleAccPrivKey = secp256k1.GenPrivKey()
	suite.oracleAccPubKey = suite.oracleAccPrivKey.PubKey()
	suite.oracleAccAddr = sdk.AccAddress(suite.oracleAccPubKey.Address())

	suite.oracle2AccPrivKey = secp256k1.GenPrivKey()
	suite.oracle2AccPubKey = suite.oracle2AccPrivKey.PubKey()
	suite.oracle2AccAddr = sdk.AccAddress(suite.oracle2AccPubKey.Address())

	suite.nodePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.nodePubKey = suite.nodePrivKey.PubKey()
}

func (suite *genesisTestSuite) TestInitGenesis() {
	genesis := types.GenesisState{
		Oracles: []types.Oracle{
			*types.NewOracle(
				suite.oracleAccAddr.String(),
				suite.uniqueID,
				suite.endpoint,
				sdk.NewDecWithPrec(1, 1),
			),
			*types.NewOracle(
				suite.oracle2AccAddr.String(),
				suite.uniqueID,
				suite.endpoint,
				sdk.NewDecWithPrec(1, 2),
			),
		},
		OracleRegistrations: []types.OracleRegistration{
			{
				UniqueId:               suite.uniqueID,
				OracleAddress:          suite.oracleAccAddr.String(),
				NodePubKey:             suite.nodePubKey.SerializeCompressed(),
				NodePubKeyRemoteReport: []byte("nodePubKeyRemoteReport"),
				TrustedBlockHeight:     10,
				TrustedBlockHash:       nil,
				Endpoint:               suite.endpoint,
				OracleCommissionRate:   sdk.NewDecWithPrec(1, 1),
			},
			{
				UniqueId:               suite.uniqueID,
				OracleAddress:          suite.oracle2AccAddr.String(),
				NodePubKey:             suite.nodePubKey.SerializeCompressed(),
				NodePubKeyRemoteReport: []byte("nodePubKeyRemoteReport"),
				TrustedBlockHeight:     10,
				TrustedBlockHash:       nil,
				Endpoint:               suite.endpoint,
				OracleCommissionRate:   sdk.NewDecWithPrec(1, 2),
			},
		},
		Params: types.DefaultParams(),
	}

	oracle.InitGenesis(suite.Ctx, suite.OracleKeeper, genesis)

	getOracle, err := suite.OracleKeeper.GetOracle(suite.Ctx, suite.oracleAccAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.Oracles[0], *getOracle)
	getOracle, err = suite.OracleKeeper.GetOracle(suite.Ctx, suite.oracle2AccAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.Oracles[1], *getOracle)

	getOracleRegistration, err := suite.OracleKeeper.GetOracleRegistration(suite.Ctx, suite.uniqueID, suite.oracleAccAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.OracleRegistrations[0], *getOracleRegistration)
	getOracleRegistration, err = suite.OracleKeeper.GetOracleRegistration(suite.Ctx, suite.uniqueID, suite.oracle2AccAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.OracleRegistrations[1], *getOracleRegistration)
}

func (suite *genesisTestSuite) TestExportGenesis() {
	ora := &types.Oracle{
		OracleAddress:        suite.oracleAccAddr.String(),
		UniqueId:             suite.uniqueID,
		Endpoint:             suite.endpoint,
		OracleCommissionRate: sdk.NewDecWithPrec(1, 1),
	}
	err := suite.OracleKeeper.SetOracle(suite.Ctx, ora)
	suite.Require().NoError(err)

	oraRegistration := &types.OracleRegistration{
		UniqueId:               suite.uniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: []byte("nodePubKeyRemoteReport"),
		TrustedBlockHeight:     10,
		TrustedBlockHash:       nil,
		Endpoint:               suite.endpoint,
		OracleCommissionRate:   sdk.NewDecWithPrec(1, 1),
	}
	err = suite.OracleKeeper.SetOracleRegistration(suite.Ctx, oraRegistration)
	suite.Require().NoError(err)

	params := types.DefaultParams()
	suite.OracleKeeper.SetParams(suite.Ctx, params)

	genesisStatus := oracle.ExportGenesis(suite.Ctx, suite.OracleKeeper)
	suite.Require().Equal(*ora, genesisStatus.Oracles[0])
	suite.Require().Equal(*oraRegistration, genesisStatus.OracleRegistrations[0])
	suite.Require().Equal(params, genesisStatus.Params)
}
