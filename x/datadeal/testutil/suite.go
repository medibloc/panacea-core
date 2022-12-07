package testutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

type DataDealBaseTestSuite struct {
	testsuite.TestSuite
}

func (suite *DataDealBaseTestSuite) MakeTestDeal(dealID uint64, consumerAddress sdk.AccAddress, maxNumData uint64) *types.Deal {
	return &types.Deal{
		Id:              dealID,
		Address:         types.NewDealAddress(dealID).String(),
		DataSchema:      []string{"http://jsonld.com"},
		Budget:          &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(1000000000)},
		MaxNumData:      maxNumData,
		CurNumData:      0,
		ConsumerAddress: consumerAddress.String(),
		Status:          types.DEAL_STATUS_ACTIVE,
	}
}
