package oracle_test

import (
	"testing"
	"time"

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

	uniqueID  string
	uniqueID2 string
	endpoint  string

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
	suite.uniqueID2 = "uniqueID2"
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
				sdk.NewDecWithPrec(2, 1),
				sdk.NewDecWithPrec(1, 2),
				time.Unix(0, 0).UTC(),
			),
			*types.NewOracle(
				suite.oracle2AccAddr.String(),
				suite.uniqueID,
				suite.endpoint,
				sdk.NewDecWithPrec(1, 1),
				sdk.NewDecWithPrec(2, 1),
				sdk.NewDecWithPrec(1, 2),
				time.Unix(0, 0).UTC(),
			),
		},
		OracleRegistrations: []types.OracleRegistration{
			{
				UniqueId:                      suite.uniqueID,
				OracleAddress:                 suite.oracleAccAddr.String(),
				NodePubKey:                    suite.nodePubKey.SerializeCompressed(),
				NodePubKeyRemoteReport:        []byte("nodePubKeyRemoteReport"),
				TrustedBlockHeight:            10,
				TrustedBlockHash:              nil,
				Endpoint:                      suite.endpoint,
				OracleCommissionRate:          sdk.NewDecWithPrec(1, 1),
				OracleCommissionMaxRate:       sdk.NewDecWithPrec(2, 1),
				OracleCommissionMaxChangeRate: sdk.NewDecWithPrec(1, 2),
				EncryptedOraclePrivKey:        nil,
			},
			{
				UniqueId:                      suite.uniqueID,
				OracleAddress:                 suite.oracle2AccAddr.String(),
				NodePubKey:                    suite.nodePubKey.SerializeCompressed(),
				NodePubKeyRemoteReport:        []byte("nodePubKeyRemoteReport"),
				TrustedBlockHeight:            10,
				TrustedBlockHash:              nil,
				Endpoint:                      suite.endpoint,
				OracleCommissionRate:          sdk.NewDecWithPrec(1, 1),
				OracleCommissionMaxRate:       sdk.NewDecWithPrec(2, 1),
				OracleCommissionMaxChangeRate: sdk.NewDecWithPrec(1, 2),
				EncryptedOraclePrivKey:        nil,
			},
		},
		OracleUpgrades: []types.OracleUpgrade{
			{
				UniqueId:               suite.uniqueID2,
				OracleAddress:          suite.oracleAccAddr.String(),
				NodePubKey:             suite.nodePubKey.SerializeCompressed(),
				NodePubKeyRemoteReport: []byte("nodePubKeyRemoteReport"),
				TrustedBlockHeight:     10,
				TrustedBlockHash:       nil,
				EncryptedOraclePrivKey: nil,
			},
			{
				UniqueId:               suite.uniqueID2,
				OracleAddress:          suite.oracle2AccAddr.String(),
				NodePubKey:             suite.nodePubKey.SerializeCompressed(),
				NodePubKeyRemoteReport: []byte("nodePubKeyRemoteReport"),
				TrustedBlockHeight:     10,
				TrustedBlockHash:       nil,
				EncryptedOraclePrivKey: nil,
			},
		},
		OracleUpgradeInfo: &types.OracleUpgradeInfo{
			UniqueId: suite.uniqueID2,
			Height:   100,
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

	getOracleUpgrade, err := suite.OracleKeeper.GetOracleUpgrade(suite.Ctx, suite.uniqueID2, suite.oracleAccAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.OracleUpgrades[0], *getOracleUpgrade)
	getOracleUpgrade, err = suite.OracleKeeper.GetOracleUpgrade(suite.Ctx, suite.uniqueID2, suite.oracle2AccAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.OracleUpgrades[1], *getOracleUpgrade)

	getOracleUpgradeInfo, err := suite.OracleKeeper.GetOracleUpgradeInfo(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.OracleUpgradeInfo, getOracleUpgradeInfo)

}

func (suite *genesisTestSuite) TestExportGenesis() {
	ora := &types.Oracle{
		OracleAddress:                 suite.oracleAccAddr.String(),
		UniqueId:                      suite.uniqueID,
		Endpoint:                      suite.endpoint,
		UpdateTime:                    time.Unix(0, 0).UTC(),
		OracleCommissionRate:          sdk.NewDecWithPrec(1, 1),
		OracleCommissionMaxRate:       sdk.NewDecWithPrec(2, 1),
		OracleCommissionMaxChangeRate: sdk.NewDecWithPrec(1, 2),
	}
	err := suite.OracleKeeper.SetOracle(suite.Ctx, ora)
	suite.Require().NoError(err)

	oraRegistration := &types.OracleRegistration{
		UniqueId:                      suite.uniqueID,
		OracleAddress:                 suite.oracleAccAddr.String(),
		NodePubKey:                    suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport:        []byte("nodePubKeyRemoteReport"),
		TrustedBlockHeight:            10,
		TrustedBlockHash:              nil,
		Endpoint:                      suite.endpoint,
		OracleCommissionRate:          sdk.NewDecWithPrec(1, 1),
		OracleCommissionMaxRate:       sdk.NewDecWithPrec(2, 1),
		OracleCommissionMaxChangeRate: sdk.NewDecWithPrec(1, 2),
		EncryptedOraclePrivKey:        nil,
	}

	err = suite.OracleKeeper.SetOracleRegistration(suite.Ctx, oraRegistration)
	suite.Require().NoError(err)

	oracleUpgrade := &types.OracleUpgrade{
		UniqueId:               suite.uniqueID2,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: []byte("nodePubKeyRemoteReport"),
		TrustedBlockHeight:     10,
		TrustedBlockHash:       nil,
		EncryptedOraclePrivKey: nil,
	}

	err = suite.OracleKeeper.SetOracleUpgrade(suite.Ctx, oracleUpgrade)
	suite.Require().NoError(err)

	oracleUpgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: suite.uniqueID2,
		Height:   100,
	}

	err = suite.OracleKeeper.SetOracleUpgradeInfo(suite.Ctx, oracleUpgradeInfo)
	suite.Require().NoError(err)

	params := types.DefaultParams()
	suite.OracleKeeper.SetParams(suite.Ctx, params)

	genesisStatus := oracle.ExportGenesis(suite.Ctx, suite.OracleKeeper)
	suite.Require().Equal(*ora, genesisStatus.Oracles[0])
	suite.Require().Equal(*oraRegistration, genesisStatus.OracleRegistrations[0])
	suite.Require().Equal(*oracleUpgrade, genesisStatus.OracleUpgrades[0])
	suite.Require().Equal(oracleUpgradeInfo, genesisStatus.OracleUpgradeInfo)
	suite.Require().Equal(params, genesisStatus.Params)
}
