package testutil

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
)

type DataDealBaseTestSuite struct {
	testsuite.TestSuite
}

func (suite *DataDealBaseTestSuite) MakeTestDeal(dealID uint64, buyerAddr sdk.AccAddress) types.Deal {
	return types.Deal{
		Id:           dealID,
		Address:      types.NewDealAddress(dealID).String(),
		DataSchema:   []string{"http://jsonld.com"},
		Budget:       &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(1000000000)},
		MaxNumData:   10000,
		CurNumData:   0,
		BuyerAddress: buyerAddr.String(),
		Status:       types.DEAL_STATUS_ACTIVE,
	}
}

func (suite *DataDealBaseTestSuite) MakeNewDataSale(sellerAddr sdk.AccAddress, verifiableCID string) *types.DataSale {
	return &types.DataSale{
		SellerAddress: sellerAddr.String(),
		DealId:        1,
		VerifiableCid: verifiableCID,
		DeliveredCid:  "",
		Status:        types.DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD,
		VotingPeriod: &oracletypes.VotingPeriod{
			VotingStartTime: time.Now(),
			VotingEndTime:   time.Now().Add(5 * time.Second),
		},
		VerificationTallyResult: nil,
		DeliveryTallyResult:     nil,
	}
}

// TODO: MakeNewDataVerificationVote will be used when PR #438 https://github.com/medibloc/panacea-core/pull/438 merged.
//func (suite *DataDealBaseTestSuite) MakeNewDataVerificationVote(voterAddr, sellerAddr sdk.AccAddress, verifiableCID string) *types.DataVerificationVote {
//	return &types.DataVerificationVote{
//		VoterAddress:  voterAddr.String(),
//		SellerAddress: sellerAddr.String(),
//		DealId:        1,
//		VerifiableCid: verifiableCID,
//		VoteOption:    oracletypes.VOTE_OPTION_YES,
//	}
//}
