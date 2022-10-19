package oracle

import (
	"testing"

	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type oracleProposalHandlerTestSuite struct {
	testsuite.TestSuite
}

func TestOracleProposalTestSuite(t *testing.T) {
	suite.Run(t, new(oracleProposalHandlerTestSuite))
}

func (suite *oracleProposalHandlerTestSuite) TestOracleUpgradeProposalHandler() {
	plan := types.Plan{
		UniqueId: "UpgradeUniqueID",
		Height:   10000,
	}
	oracleUpgradeProposal := &types.OracleUpgradeProposal{
		Title:       "OracleUpgradeTitle",
		Description: "OracleUpgradeDescription",
		Plan:        plan,
	}
	err := NewOracleProposalHandler(suite.OracleKeeper)(suite.Ctx, oracleUpgradeProposal)
	require.NoError(suite.T(), err)

	upgradeInfo, err := suite.OracleKeeper.GetOracleUpgradeInfo(suite.Ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), plan.UniqueId, upgradeInfo.UniqueId)
	require.Equal(suite.T(), plan.Height, upgradeInfo.Height)
}
