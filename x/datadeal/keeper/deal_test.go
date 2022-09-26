package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
)

type dealTestSuite struct {
	testsuite.TestSuite

	sellerAccPrivKey cryptotypes.PrivKey
	sellerAccPubKey  cryptotypes.PubKey
	sellerAccAddr    sdk.AccAddress

	verifiableCID []byte
}

func TestDataDealTestSuite(t *testing.T) {
	suite.Run(t, new(dealTestSuite))
}

func (suite *dealTestSuite) BeforeTest(_, _ string) {
	suite.verifiableCID = []byte("verifiableCID")

	suite.sellerAccPrivKey = secp256k1.GenPrivKey()
	suite.sellerAccPubKey = suite.sellerAccPrivKey.PubKey()
	suite.sellerAccAddr = sdk.AccAddress(suite.sellerAccPubKey.Address())

	suite.OracleKeeper.SetParams(suite.Ctx, oracletypes.Params{
		VoteParams: oracletypes.VoteParams{
			VotingPeriod: 100,
			JailPeriod:   60,
			Threshold:    sdk.NewDecWithPrec(2, 3),
		},
	})
}

func (suite dealTestSuite) makeNewDataSale() *types.DataSale {
	return &types.DataSale{
		SellerAddress: suite.sellerAccAddr.String(),
		DealId:        0,
		VerifiableCid: suite.verifiableCID,
		DeliveredCid:  nil,
		Status:        types.DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD,
		VotingPeriod: &oracletypes.VotingPeriod{
			VotingStartTime: time.Now(),
			VotingEndTime:   time.Now().Add(5 * time.Second),
		},
		VerificationTallyResult: nil,
		DeliveryTallyResult:     nil,
	}
}

func (suite dealTestSuite) TestSellDataSuccess() {
	msgSellData := &types.MsgSellData{
		DealId:        0,
		VerifiableCid: suite.verifiableCID,
		SellerAddress: suite.sellerAccAddr.String(),
	}

	err := suite.DataDealKeeper.SellData(suite.Ctx, msgSellData)
	suite.Require().NoError(err)
}
