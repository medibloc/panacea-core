package oracle_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/oracle"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

var oracle1 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

type genesisTestSuite struct {
	testsuite.TestSuite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}

func (suite genesisTestSuite) TestDataPoolInitGenesis() {
	var oracles []types.Oracle

	tempOracle := makeSampleOracle()

	oracles = append(oracles, tempOracle)

	genState := &types.GenesisState{
		Oracles: map[string]types.Oracle{
			tempOracle.Address: tempOracle,
		},
	}

	oracle.InitGenesis(suite.Ctx, suite.OracleKeeper, *genState)

	// check oracle
	oracleFromKeeper, err := suite.OracleKeeper.GetOracle(suite.Ctx, oracle1)
	suite.Require().NoError(err)
	suite.Require().Equal(tempOracle, oracleFromKeeper)

	// check all oracles
	oraclesFromKeeper, err := suite.OracleKeeper.GetAllOracles(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(oraclesFromKeeper, oracles)
	suite.Require().Len(oraclesFromKeeper, 1)
}

func (suite genesisTestSuite) TestOracleExportGenesis() {
	// register oracle
	tempOracle := makeSampleOracle()

	err := suite.OracleKeeper.SetOracle(suite.Ctx, tempOracle)
	suite.Require().NoError(err)

	genesisState := oracle.ExportGenesis(suite.Ctx, suite.OracleKeeper)
	suite.Require().Len(genesisState.Oracles, 1)

	oracleAddr, err := sdk.AccAddressFromBech32(tempOracle.Address)
	suite.Require().NoError(err)

	oracleKey := types.GetKeyPrefixOracle(oracleAddr)

	suite.Require().Equal(genesisState.Oracles[string(oracleKey)].Address, tempOracle.Address)
	suite.Require().Equal(genesisState.Oracles[string(oracleKey)].Endpoint, tempOracle.Endpoint)
}

func makeSampleOracle() types.Oracle {
	return types.Oracle{
		Address:  oracle1.String(),
		Endpoint: "https://my-oracle.org",
	}
}
