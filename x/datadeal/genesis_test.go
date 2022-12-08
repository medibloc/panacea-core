package datadeal_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datadeal"
	"github.com/medibloc/panacea-core/v2/x/datadeal/testutil"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/stretchr/testify/suite"
)

type genesisTestSuite struct {
	testutil.DataDealBaseTestSuite

	consumerAccAddr sdk.AccAddress
	defaultFunds    sdk.Coins
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}

func (suite *genesisTestSuite) BeforeTest(_, _ string) {
	suite.consumerAccAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	suite.defaultFunds = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)))
}

func (suite *genesisTestSuite) TestInitGenesis() {
	deal1 := suite.MakeTestDeal(1, suite.consumerAccAddr, 100)
	deal2 := suite.MakeTestDeal(2, suite.consumerAccAddr, 100)

	genesis := types.GenesisState{
		Deals:          []types.Deal{*deal1, *deal2},
		NextDealNumber: 1,
	}

	datadeal.InitGenesis(suite.Ctx, suite.DataDealKeeper, genesis)

	getDeal1, err := suite.DataDealKeeper.GetDeal(suite.Ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.Deals[0], *getDeal1)

	getDeal2, err := suite.DataDealKeeper.GetDeal(suite.Ctx, 2)
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.Deals[1], *getDeal2)
}

func (suite *genesisTestSuite) TestExportGenesis() {
	deal1 := suite.MakeTestDeal(1, suite.consumerAccAddr, 100)
	deal2 := suite.MakeTestDeal(2, suite.consumerAccAddr, 100)

	genesis := types.GenesisState{
		Deals:          []types.Deal{*deal1},
		NextDealNumber: 2,
	}

	msgCreateDeal := &types.MsgCreateDeal{
		DataSchema:      deal2.DataSchema,
		Budget:          deal2.Budget,
		MaxNumData:      deal2.MaxNumData,
		ConsumerAddress: deal2.ConsumerAddress,
	}

	datadeal.InitGenesis(suite.Ctx, suite.DataDealKeeper, genesis)

	err := suite.FundAccount(suite.Ctx, suite.consumerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	_, err = suite.DataDealKeeper.CreateDeal(suite.Ctx, msgCreateDeal)
	suite.Require().NoError(err)

	genesisStatus := datadeal.ExportGenesis(suite.Ctx, suite.DataDealKeeper)

	suite.Require().Equal(deal1.Id, genesisStatus.Deals[0].Id)
	suite.Require().Equal(deal2.Id, genesisStatus.Deals[1].Id)
	suite.Require().Equal(deal1.Address, genesisStatus.Deals[0].Address)
	suite.Require().Equal(deal2.Address, genesisStatus.Deals[1].Address)
	suite.Require().Equal(deal1.ConsumerAddress, genesisStatus.Deals[0].ConsumerAddress)
	suite.Require().Equal(deal2.ConsumerAddress, genesisStatus.Deals[1].ConsumerAddress)
	suite.Require().Equal(deal1.DataSchema, genesisStatus.Deals[0].DataSchema)
	suite.Require().Equal(deal2.DataSchema, genesisStatus.Deals[1].DataSchema)
	suite.Require().Equal(deal1.Budget, genesisStatus.Deals[0].Budget)
	suite.Require().Equal(deal2.Budget, genesisStatus.Deals[1].Budget)
	suite.Require().Equal(uint64(3), genesisStatus.NextDealNumber)
}
