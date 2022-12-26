package keeper_test

import (
	"testing"

	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type oracleUpgradeTestSuite struct {
	testsuite.TestSuite
}

func TestOracleUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(oracleUpgradeTestSuite))
}

func (suite *oracleUpgradeTestSuite) TestOracleUpgradeInfo() {
	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: "upgradeUniqueID",
		Height:   10000000,
	}

	require.NoError(suite.T(), suite.OracleKeeper.SetOracleUpgradeInfo(suite.Ctx, upgradeInfo))

	getUpgradeInfo, err := suite.OracleKeeper.GetOracleUpgradeInfo(suite.Ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), upgradeInfo, getUpgradeInfo)
}

func (suite *oracleUpgradeTestSuite) TestEmptyOracleUpgradeInfo() {
	upgradeInfo, err := suite.OracleKeeper.GetOracleUpgradeInfo(suite.Ctx)
	require.ErrorIs(suite.T(), err, types.ErrOracleUpgradeInfoNotFound)
	require.Nil(suite.T(), upgradeInfo)
}

func (suite *oracleUpgradeTestSuite) TestApplyUpgradeSuccess() {
	ctx := suite.Ctx

	params := types.DefaultParams()
	params.UniqueId = "orgUniqueID"
	suite.OracleKeeper.SetParams(ctx, params)
	suite.Require().Equal(params.UniqueId, suite.OracleKeeper.GetParams(ctx).UniqueId)

	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: "upgradeUniqueID",
		Height:   1,
	}

	suite.Require().NoError(suite.OracleKeeper.ApplyUpgrade(ctx, upgradeInfo))
	suite.Require().Equal(upgradeInfo.UniqueId, suite.OracleKeeper.GetParams(ctx).UniqueId)
}
