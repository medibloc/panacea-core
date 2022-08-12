package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"

	"github.com/stretchr/testify/suite"

	"github.com/medibloc/panacea-core/v2/types/testsuite"
)

type queryOracleTestSuite struct {
	testsuite.TestSuite
}

func TestQueryOracleTestSuite(t *testing.T) {
	suite.Run(t, new(queryOracleTestSuite))
}

func (suite queryOracleTestSuite) TestOracleRegistration() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	newOracleRegistration := makeNewOracleRegistration()

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
