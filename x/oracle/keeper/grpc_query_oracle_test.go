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

	oracleAccPrivKey              cryptotypes.PrivKey
	oracleAccPubKey               cryptotypes.PubKey
	oracleAccAddr                 sdk.AccAddress
	oracleEndpoint                string
	oracleCommissionRate          sdk.Dec
	oracleCommissionMaxRate       sdk.Dec
	oracleCommissionMaxChangeRate sdk.Dec

	oracle2AccPrivKey              cryptotypes.PrivKey
	oracle2AccPubKey               cryptotypes.PubKey
	oracle2AccAddr                 sdk.AccAddress
	oracle2Endpoint                string
	oracle2CommissionRate          sdk.Dec
	oracle2CommissionMaxRate       sdk.Dec
	oracle2CommissionMaxChangeRate sdk.Dec

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
	suite.oracleCommissionRate = sdk.NewDecWithPrec(1, 1)          // 0.1
	suite.oracleCommissionMaxRate = sdk.NewDecWithPrec(2, 1)       // 0.2
	suite.oracleCommissionMaxChangeRate = sdk.NewDecWithPrec(1, 2) // 0.01

	suite.oracle2AccPrivKey = secp256k1.GenPrivKey()
	suite.oracle2AccPubKey = suite.oracle2AccPrivKey.PubKey()
	suite.oracle2AccAddr = sdk.AccAddress(suite.oracle2AccPubKey.Address())
	suite.oracle2Endpoint = "https://my-validator2.org"
	suite.oracle2CommissionRate = sdk.NewDecWithPrec(1, 1)          // 0.1
	suite.oracle2CommissionMaxRate = sdk.NewDecWithPrec(3, 1)       // 0.3
	suite.oracle2CommissionMaxChangeRate = sdk.NewDecWithPrec(2, 2) // 0.02

	suite.nodePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.nodePubKey = suite.nodePrivKey.PubKey()
}

func (suite *queryOracleTestSuite) TestOracles() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	oracle := types.NewOracle(suite.oracleAccAddr.String(), suite.uniqueID, suite.oracleEndpoint, suite.oracleCommissionRate, suite.oracleCommissionMaxRate, suite.oracleCommissionMaxChangeRate, ctx.BlockTime())
	err := oracleKeeper.SetOracle(ctx, oracle)
	suite.Require().NoError(err)

	oracle2 := types.NewOracle(suite.oracle2AccAddr.String(), suite.uniqueID, suite.oracle2Endpoint, suite.oracle2CommissionRate, suite.oracle2CommissionMaxRate, suite.oracle2CommissionMaxChangeRate, ctx.BlockTime())
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
			suite.Require().Equal(ctx.BlockTime(), oracle.UpdateTime)
			suite.Require().Equal(suite.oracleCommissionRate, oracle.OracleCommissionRate)
			suite.Require().Equal(suite.oracleCommissionMaxRate, oracle.OracleCommissionMaxRate)
			suite.Require().Equal(suite.oracleCommissionMaxChangeRate, oracle.OracleCommissionMaxChangeRate)
		case suite.oracle2AccAddr.String():
			suite.Require().Equal(suite.uniqueID, oracle.UniqueId)
			suite.Require().Equal(suite.oracle2Endpoint, oracle.Endpoint)
			suite.Require().Equal(ctx.BlockTime(), oracle.UpdateTime)
			suite.Require().Equal(suite.oracle2CommissionRate, oracle.OracleCommissionRate)
			suite.Require().Equal(suite.oracle2CommissionMaxRate, oracle.OracleCommissionMaxRate)
			suite.Require().Equal(suite.oracle2CommissionMaxChangeRate, oracle.OracleCommissionMaxChangeRate)
		default:
			panic("not found oracle address. address: " + oracle.OracleAddress)
		}
	}
}

func (suite *queryOracleTestSuite) TestOracle() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	oracle := types.NewOracle(suite.oracleAccAddr.String(), suite.uniqueID, suite.oracleEndpoint, suite.oracleCommissionRate, suite.oracleCommissionMaxRate, suite.oracleCommissionMaxChangeRate, ctx.BlockTime())
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
		UniqueId:                      suite.uniqueID,
		OracleAddress:                 suite.oracleAccAddr.String(),
		NodePubKey:                    suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport:        remoteReport,
		TrustedBlockHeight:            10,
		TrustedBlockHash:              trustedBlockHash,
		Endpoint:                      suite.oracleEndpoint,
		OracleCommissionRate:          suite.oracleCommissionRate,
		OracleCommissionMaxRate:       suite.oracleCommissionMaxRate,
		OracleCommissionMaxChangeRate: suite.oracleCommissionMaxChangeRate,
		EncryptedOraclePrivKey:        nil,
	}

	oracleRegistration2 := &types.OracleRegistration{
		UniqueId:                      suite.uniqueID,
		OracleAddress:                 suite.oracle2AccAddr.String(),
		NodePubKey:                    suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport:        remoteReport,
		TrustedBlockHeight:            10,
		TrustedBlockHash:              trustedBlockHash,
		Endpoint:                      suite.oracle2Endpoint,
		OracleCommissionRate:          suite.oracle2CommissionRate,
		OracleCommissionMaxRate:       suite.oracle2CommissionMaxRate,
		OracleCommissionMaxChangeRate: suite.oracle2CommissionMaxChangeRate,
		EncryptedOraclePrivKey:        nil,
	}

	anotherUniqueID := "uniqueID2"
	oracleRegistration3 := &types.OracleRegistration{
		UniqueId:                      anotherUniqueID,
		OracleAddress:                 suite.oracle2AccAddr.String(),
		NodePubKey:                    suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport:        remoteReport,
		TrustedBlockHeight:            10,
		TrustedBlockHash:              trustedBlockHash,
		Endpoint:                      suite.oracle2Endpoint,
		OracleCommissionRate:          suite.oracle2CommissionRate,
		OracleCommissionMaxRate:       suite.oracle2CommissionMaxRate,
		OracleCommissionMaxChangeRate: suite.oracle2CommissionMaxChangeRate,
		EncryptedOraclePrivKey:        nil,
	}

	err := oracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)
	err = oracleKeeper.SetOracleRegistration(ctx, oracleRegistration2)
	suite.Require().NoError(err)
	err = oracleKeeper.SetOracleRegistration(ctx, oracleRegistration3)
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
			suite.Require().Equal(suite.oracleCommissionMaxRate, oracleRegistration.OracleCommissionMaxRate)
			suite.Require().Equal(suite.oracleCommissionMaxChangeRate, oracleRegistration.OracleCommissionMaxChangeRate)
		case suite.oracle2AccAddr.String():
			suite.Require().Equal(suite.uniqueID, oracleRegistration.UniqueId)
			suite.Require().Equal(suite.nodePubKey.SerializeCompressed(), oracleRegistration.NodePubKey)
			suite.Require().Equal(remoteReport, oracleRegistration.NodePubKeyRemoteReport)
			suite.Require().Equal(int64(10), oracleRegistration.TrustedBlockHeight)
			suite.Require().Equal(trustedBlockHash, oracleRegistration.TrustedBlockHash)
			suite.Require().Equal(suite.oracle2Endpoint, oracleRegistration.Endpoint)
			suite.Require().Equal(suite.oracle2CommissionRate, oracleRegistration.OracleCommissionRate)
			suite.Require().Equal(suite.oracle2CommissionMaxRate, oracleRegistration.OracleCommissionMaxRate)
			suite.Require().Equal(suite.oracle2CommissionMaxChangeRate, oracleRegistration.OracleCommissionMaxChangeRate)
		default:
			panic("not found oracle address. address: " + oracleRegistration.OracleAddress)
		}
	}

	req.UniqueId = anotherUniqueID
	res, err = oracleKeeper.OracleRegistrations(sdk.WrapSDKContext(ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(res.OracleRegistrations))

	suite.Require().Equal(suite.oracle2AccAddr.String(), res.OracleRegistrations[0].OracleAddress)
	suite.Require().Equal(anotherUniqueID, res.OracleRegistrations[0].UniqueId)
	suite.Require().Equal(suite.nodePubKey.SerializeCompressed(), res.OracleRegistrations[0].NodePubKey)
	suite.Require().Equal(remoteReport, res.OracleRegistrations[0].NodePubKeyRemoteReport)
	suite.Require().Equal(int64(10), res.OracleRegistrations[0].TrustedBlockHeight)
	suite.Require().Equal(trustedBlockHash, res.OracleRegistrations[0].TrustedBlockHash)
	suite.Require().Equal(suite.oracle2Endpoint, res.OracleRegistrations[0].Endpoint)
	suite.Require().Equal(suite.oracle2CommissionRate, res.OracleRegistrations[0].OracleCommissionRate)
	suite.Require().Equal(suite.oracle2CommissionMaxRate, res.OracleRegistrations[0].OracleCommissionMaxRate)
	suite.Require().Equal(suite.oracle2CommissionMaxChangeRate, res.OracleRegistrations[0].OracleCommissionMaxChangeRate)
}

func (suite *queryOracleTestSuite) TestOracleRegistration() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	oracleRegistration := &types.OracleRegistration{
		UniqueId:                      suite.uniqueID,
		OracleAddress:                 suite.oracleAccAddr.String(),
		NodePubKey:                    suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport:        []byte("nodePubKeyRemoteReport"),
		TrustedBlockHeight:            10,
		TrustedBlockHash:              []byte("hash"),
		Endpoint:                      suite.oracleEndpoint,
		OracleCommissionRate:          suite.oracleCommissionRate,
		OracleCommissionMaxRate:       suite.oracleCommissionMaxRate,
		OracleCommissionMaxChangeRate: suite.oracleCommissionMaxChangeRate,
		EncryptedOraclePrivKey:        nil,
	}
	err := oracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)
}

func (suite *queryOracleTestSuite) TestOracleParams() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	oracleKeeper.SetParams(ctx, types.DefaultParams())

	req := types.QueryOracleParamsRequest{}
	res, err := oracleKeeper.Params(sdk.WrapSDKContext(ctx), &req)
	suite.Require().NoError(err)

	suite.Require().Equal(types.DefaultParams(), *res.Params)
}

func (suite *queryOracleTestSuite) TestOracleUpgradeInfo() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: "UpgradeUniqueID",
		Height:   1000000,
	}
	suite.Require().NoError(oracleKeeper.SetOracleUpgradeInfo(ctx, upgradeInfo))

	getUpgradeInfo, err := oracleKeeper.OracleUpgradeInfo(sdk.WrapSDKContext(ctx), nil)
	suite.Require().NoError(err)
	suite.Require().Equal(upgradeInfo.UniqueId, getUpgradeInfo.OracleUpgradeInfo.UniqueId)
	suite.Require().Equal(upgradeInfo.Height, getUpgradeInfo.OracleUpgradeInfo.Height)
}

func (suite *queryOracleTestSuite) TestOracleUpgrades() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	remoteReport := []byte("nodePubKeyRemoteReport")
	trustedBlockHash := []byte("hash")

	oracleUpgrade := &types.OracleUpgrade{
		UniqueId:               suite.uniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: remoteReport,
		TrustedBlockHeight:     10,
		TrustedBlockHash:       trustedBlockHash,
		EncryptedOraclePrivKey: nil,
	}

	oracleUpgrade2 := &types.OracleUpgrade{
		UniqueId:               suite.uniqueID,
		OracleAddress:          suite.oracle2AccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: remoteReport,
		TrustedBlockHeight:     10,
		TrustedBlockHash:       trustedBlockHash,
		EncryptedOraclePrivKey: nil,
	}

	anotherUniqueID := "uniqueID2"
	oracleUpgrade3 := &types.OracleUpgrade{
		UniqueId:               anotherUniqueID,
		OracleAddress:          suite.oracle2AccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: remoteReport,
		TrustedBlockHeight:     10,
		TrustedBlockHash:       trustedBlockHash,
		EncryptedOraclePrivKey: nil,
	}

	err := oracleKeeper.SetOracleUpgrade(ctx, oracleUpgrade)
	suite.Require().NoError(err)
	err = oracleKeeper.SetOracleUpgrade(ctx, oracleUpgrade2)
	suite.Require().NoError(err)
	err = oracleKeeper.SetOracleUpgrade(ctx, oracleUpgrade3)
	suite.Require().NoError(err)

	req := types.QueryOracleUpgradesRequest{
		UniqueId:   suite.uniqueID,
		Pagination: &query.PageRequest{},
	}
	res, err := oracleKeeper.OracleUpgrades(sdk.WrapSDKContext(ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(res.OracleUpgrades))
	for _, oracleUpgrade := range res.OracleUpgrades {
		switch oracleUpgrade.OracleAddress {
		case suite.oracleAccAddr.String():
			suite.Require().Equal(suite.uniqueID, oracleUpgrade.UniqueId)
			suite.Require().Equal(suite.nodePubKey.SerializeCompressed(), oracleUpgrade.NodePubKey)
			suite.Require().Equal(remoteReport, oracleUpgrade.NodePubKeyRemoteReport)
			suite.Require().Equal(int64(10), oracleUpgrade.TrustedBlockHeight)
			suite.Require().Equal(trustedBlockHash, oracleUpgrade.TrustedBlockHash)
		case suite.oracle2AccAddr.String():
			suite.Require().Equal(suite.uniqueID, oracleUpgrade.UniqueId)
			suite.Require().Equal(suite.nodePubKey.SerializeCompressed(), oracleUpgrade.NodePubKey)
			suite.Require().Equal(remoteReport, oracleUpgrade.NodePubKeyRemoteReport)
			suite.Require().Equal(int64(10), oracleUpgrade.TrustedBlockHeight)
			suite.Require().Equal(trustedBlockHash, oracleUpgrade.TrustedBlockHash)
		default:
			panic("not found oracle address. address: " + oracleUpgrade.OracleAddress)
		}
	}

	req.UniqueId = anotherUniqueID
	res, err = oracleKeeper.OracleUpgrades(sdk.WrapSDKContext(ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(res.OracleUpgrades))

	suite.Require().Equal(suite.oracle2AccAddr.String(), res.OracleUpgrades[0].OracleAddress)
	suite.Require().Equal(anotherUniqueID, res.OracleUpgrades[0].UniqueId)
	suite.Require().Equal(suite.nodePubKey.SerializeCompressed(), res.OracleUpgrades[0].NodePubKey)
	suite.Require().Equal(remoteReport, res.OracleUpgrades[0].NodePubKeyRemoteReport)
	suite.Require().Equal(int64(10), res.OracleUpgrades[0].TrustedBlockHeight)
	suite.Require().Equal(trustedBlockHash, res.OracleUpgrades[0].TrustedBlockHash)
}

func (suite *queryOracleTestSuite) TestOracleUpgrade() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	oracleUpgrade := &types.OracleUpgrade{
		UniqueId:               suite.uniqueID,
		OracleAddress:          suite.oracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: []byte("nodePubKeyRemoteReport"),
		TrustedBlockHeight:     10,
		TrustedBlockHash:       []byte("hash"),
		EncryptedOraclePrivKey: nil,
	}

	suite.Require().NoError(oracleKeeper.SetOracleUpgrade(ctx, oracleUpgrade))

	getOracleUpgrade, err := oracleKeeper.GetOracleUpgrade(ctx, suite.uniqueID, suite.oracleAccAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(oracleUpgrade.UniqueId, getOracleUpgrade.UniqueId)
	suite.Require().Equal(oracleUpgrade.OracleAddress, getOracleUpgrade.OracleAddress)
	suite.Require().Equal(oracleUpgrade.NodePubKey, getOracleUpgrade.NodePubKey)
	suite.Require().Equal(oracleUpgrade.NodePubKeyRemoteReport, getOracleUpgrade.NodePubKeyRemoteReport)
	suite.Require().Equal(oracleUpgrade.TrustedBlockHeight, getOracleUpgrade.TrustedBlockHeight)
	suite.Require().Equal(oracleUpgrade.TrustedBlockHash, getOracleUpgrade.TrustedBlockHash)
}
