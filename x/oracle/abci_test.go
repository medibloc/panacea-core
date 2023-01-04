package oracle_test

import (
	"encoding/base64"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/oracle"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
)

type abciTestSuite struct {
	testsuite.TestSuite

	uniqueID string
}

func TestAbciTestSuite(t *testing.T) {
	suite.Run(t, new(abciTestSuite))
}

func (suite *abciTestSuite) BeforeTest(_, _ string) {
	ctx := suite.Ctx
	suite.uniqueID = "uniqueID"

	oraclePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	suite.Require().NoError(err)

	suite.OracleKeeper.SetParams(ctx, types.Params{
		OraclePublicKey:          base64.StdEncoding.EncodeToString(oraclePrivKey.PubKey().SerializeCompressed()),
		OraclePubKeyRemoteReport: base64.StdEncoding.EncodeToString([]byte("oraclePubKeyRemoteReport")),
		UniqueId:                 suite.uniqueID,
	})
}

func (suite *abciTestSuite) TestOracleUpgradeSuccess() {
	ctx := suite.Ctx
	ctx = ctx.WithBlockHeight(1)

	suite.Require().Equal(suite.uniqueID, suite.OracleKeeper.GetParams(ctx).UniqueId)

	upgradeUniqueID := "upgradeUniqueID"

	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: upgradeUniqueID,
		Height:   10,
	}

	suite.Require().NoError(suite.OracleKeeper.SetOracleUpgradeInfo(ctx, upgradeInfo))

	ctx = ctx.WithBlockHeight(10)

	oracle.EndBlocker(ctx, suite.OracleKeeper)

	suite.Require().Equal(upgradeUniqueID, suite.OracleKeeper.GetParams(ctx).UniqueId)

}

func (suite *abciTestSuite) TestOracleUpgradeFailedBeforeReachUpgradeHeight() {
	ctx := suite.Ctx
	ctx = ctx.WithBlockHeight(1)

	suite.Require().Equal(suite.uniqueID, suite.OracleKeeper.GetParams(ctx).UniqueId)

	upgradeUniqueID := "upgradeUniqueID"

	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: upgradeUniqueID,
		Height:   10,
	}

	suite.Require().NoError(suite.OracleKeeper.SetOracleUpgradeInfo(ctx, upgradeInfo))

	ctx = ctx.WithBlockHeight(9)

	oracle.EndBlocker(ctx, suite.OracleKeeper)

	suite.Require().Equal(suite.uniqueID, suite.OracleKeeper.GetParams(ctx).UniqueId)
}
