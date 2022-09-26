package keeper_test

import (
	"testing"

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
		OraclePublicKey:          "",
		OraclePubKeyRemoteReport: "",
		UniqueId:                 "",
		VoteParams: oracletypes.VoteParams{
			VotingPeriod: 100,
			JailPeriod:   60,
			Threshold:    sdk.NewDecWithPrec(2, 3),
		},
		SlashParams: oracletypes.SlashParams{
			SlashFractionDowntime: sdk.NewDecWithPrec(3, 1),
			SlashFractionForgery:  sdk.NewDecWithPrec(1, 1),
		},
	})
}

//TODO: The test will be complemented when CreateDeal and VoteDataSale done.
func (suite dealTestSuite) TestSellDataSuccess() {
	msgSellData := &types.MsgSellData{
		DealId:        1,
		VerifiableCid: suite.verifiableCID,
		SellerAddress: suite.sellerAccAddr.String(),
	}

	err := suite.DataDealKeeper.SellData(suite.Ctx, msgSellData)
	suite.Require().NoError(err)

	dataSale, err := suite.DataDealKeeper.GetDataSale(suite.Ctx, suite.verifiableCID, uint64(1))
	suite.Require().NoError(err)

	suite.Require().Equal(dataSale.VerifiableCid, suite.verifiableCID)
	suite.Require().Equal(dataSale.DealId, uint64(1))
	suite.Require().Equal(dataSale.VotingPeriod, suite.OracleKeeper.GetVotingPeriod(suite.Ctx))
	suite.Require().Equal(dataSale.Status, types.DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD)
	suite.Require().Equal(dataSale.SellerAddress, suite.sellerAccAddr.String())
}
