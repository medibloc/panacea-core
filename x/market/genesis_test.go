package market_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/market"
	"github.com/medibloc/panacea-core/v2/x/market/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"testing"
)

var acc1 = secp256k1.GenPrivKey().PubKey().Address()

type genesisTestSuite struct {
	testsuite.TestSuite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}

func (suite *genesisTestSuite) TestMarketInitGenesis() {
	newDeal, err := makeTestDeal()
	suite.Require().NoError(err)

	market.InitGenesis(suite.Ctx, suite.MarketKeeper, types.GenesisState{
		Deals: map[uint64]*types.Deal{
			newDeal.GetDealId(): &newDeal,
		},
		NextDealNumber: 2,
	})

	suite.Require().Equal(suite.MarketKeeper.GetNextDealNumberAndIncrement(suite.Ctx), uint64(2))

	dealStored, err := suite.MarketKeeper.GetDeal(suite.Ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(newDeal.GetDealId(), dealStored.GetDealId())
	suite.Require().Equal(newDeal.GetDealAddress(), dealStored.GetDealAddress())
	suite.Require().Equal(newDeal.GetBudget(), dealStored.GetBudget())
	suite.Require().Equal(newDeal.GetWantDataCount(), dealStored.GetWantDataCount())
	suite.Require().Equal(newDeal.GetOwner(), dealStored.GetOwner())
	suite.Require().Equal(newDeal.GetStatus(), dealStored.GetStatus())

	_, err = suite.MarketKeeper.GetDeal(suite.Ctx, 2)
	suite.Require().NoError(err)

}

func makeTestDeal() (types.Deal, error) {
	return types.Deal{
		DealId:               1,
		DealAddress:          types.NewDealAddress(1).String(),
		DataSchema:           nil,
		Budget:               &sdk.Coin{Denom: "umed", Amount: sdk.NewInt(10000000)},
		TrustedDataValidator: nil,
		WantDataCount:        10000,
		CompleteDataCount:    0,
		Owner:                acc1.String(),
		Status:               "ACTIVE",
	}, nil
}
