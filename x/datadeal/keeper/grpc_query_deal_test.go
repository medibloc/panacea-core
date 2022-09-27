package keeper_test

import (
	"encoding/base64"
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

type queryDealTestSuite struct {
	testsuite.TestSuite

	sellerAccPrivKey cryptotypes.PrivKey
	sellerAccPubKey  cryptotypes.PubKey
	sellerAccAddr    sdk.AccAddress

	verifiableCID string
}

func TestQueryDealTestSuite(t *testing.T) {
	suite.Run(t, new(queryDealTestSuite))
}

func (suite *queryDealTestSuite) BeforeTest(_, _ string) {
	verifiableCIDbz := []byte("verifiableCID")
	suite.verifiableCID = base64.StdEncoding.EncodeToString(verifiableCIDbz)

	suite.sellerAccPrivKey = secp256k1.GenPrivKey()
	suite.sellerAccPubKey = suite.sellerAccPrivKey.PubKey()
	suite.sellerAccAddr = sdk.AccAddress(suite.sellerAccPubKey.Address())

}

func (suite queryDealTestSuite) makeNewDataSale() *types.DataSale {
	return &types.DataSale{
		SellerAddress: suite.sellerAccAddr.String(),
		DealId:        1,
		VerifiableCid: suite.verifiableCID,
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

func (suite queryDealTestSuite) TestDataSale() {
	newDataSale := suite.makeNewDataSale()
	err := suite.DataDealKeeper.SetDataSale(suite.Ctx, newDataSale)
	suite.Require().NoError(err)

	req := types.QueryDataSaleRequest{
		DealId:        1,
		VerifiableCid: suite.verifiableCID,
	}

	res, err := suite.DataDealKeeper.DataSale(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Equal(newDataSale.DealId, res.DataSale.DealId)
	suite.Require().Equal(newDataSale.VerifiableCid, res.DataSale.VerifiableCid)
}
