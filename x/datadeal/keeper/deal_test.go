package keeper_test

import (
	"sort"

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

	verifiableCID1 string
	verifiableCID2 string
	verifiableCID3 string
}

func TestDataDealTestSuite(t *testing.T) {
	suite.Run(t, new(dealTestSuite))
}

func (suite *dealTestSuite) BeforeTest(_, _ string) {
	suite.verifiableCID1 = "verifiableCID"

	suite.verifiableCID2 = "verifiableCID2"

	suite.verifiableCID3 = "verifiableCID3"

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

func (suite dealTestSuite) makeNewDataSale(verifiableCID string) *types.DataSale {
	return &types.DataSale{
		SellerAddress: suite.sellerAccAddr.String(),
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

//TODO: The test will be complemented when CreateDeal and VoteDataSale done.
func (suite dealTestSuite) TestSellDataSuccess() {
	msgSellData := &types.MsgSellData{
		DealId:        1,
		VerifiableCid: suite.verifiableCID1,
		SellerAddress: suite.sellerAccAddr.String(),
	}

	err := suite.DataDealKeeper.SellData(suite.Ctx, msgSellData)
	suite.Require().NoError(err)

	dataSale, err := suite.DataDealKeeper.GetDataSale(suite.Ctx, suite.verifiableCID1, uint64(1))
	suite.Require().NoError(err)

	suite.Require().Equal(dataSale.VerifiableCid, suite.verifiableCID1)
	suite.Require().Equal(dataSale.DealId, uint64(1))
	suite.Require().Equal(dataSale.VotingPeriod, suite.OracleKeeper.GetVotingPeriod(suite.Ctx))
	suite.Require().Equal(dataSale.Status, types.DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD)
	suite.Require().Equal(dataSale.SellerAddress, suite.sellerAccAddr.String())
}

func (suite dealTestSuite) TestSellDataStatusFailed() {
	newDataSale := suite.makeNewDataSale(suite.verifiableCID1)
	newDataSale.Status = types.DATA_SALE_STATUS_FAILED

	err := suite.DataDealKeeper.SetDataSale(suite.Ctx, newDataSale)
	suite.Require().NoError(err)

	msgSellData := &types.MsgSellData{
		DealId:        1,
		VerifiableCid: newDataSale.VerifiableCid,
		SellerAddress: newDataSale.SellerAddress,
	}

	err = suite.DataDealKeeper.SellData(suite.Ctx, msgSellData)
	suite.Require().NoError(err)
}

func (suite dealTestSuite) TestSellDataStatusVotingPeriod() {
	newDataSale := suite.makeNewDataSale(suite.verifiableCID1)
	newDataSale.Status = types.DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD

	err := suite.DataDealKeeper.SetDataSale(suite.Ctx, newDataSale)
	suite.Require().NoError(err)

	msgSellData := &types.MsgSellData{
		DealId:        1,
		VerifiableCid: newDataSale.VerifiableCid,
		SellerAddress: newDataSale.SellerAddress,
	}

	err = suite.DataDealKeeper.SellData(suite.Ctx, msgSellData)
	suite.Require().Error(err, types.ErrSellData)
}

func (suite dealTestSuite) TestSellDataStatusCompleted() {
	newDataSale := suite.makeNewDataSale(suite.verifiableCID1)
	newDataSale.Status = types.DATA_SALE_STATUS_COMPLETED

	err := suite.DataDealKeeper.SetDataSale(suite.Ctx, newDataSale)
	suite.Require().NoError(err)

	msgSellData := &types.MsgSellData{
		DealId:        1,
		VerifiableCid: newDataSale.VerifiableCid,
		SellerAddress: newDataSale.SellerAddress,
	}

	err = suite.DataDealKeeper.SellData(suite.Ctx, msgSellData)
	suite.Require().Error(err, types.ErrSellData)
}

func (suite dealTestSuite) TestGetAllDataSalesList() {
	type dataSaleKey struct {
		verifiableCID string
		dealID        uint64
	}
	dataSaleKeys := make([]dataSaleKey, 0)

	dataSale1 := suite.makeNewDataSale(suite.verifiableCID1)
	err := suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale1)
	suite.Require().NoError(err)
	dataSaleKeys = append(dataSaleKeys, dataSaleKey{dataSale1.VerifiableCid, dataSale1.DealId})

	dataSale2 := suite.makeNewDataSale(suite.verifiableCID2)
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale2)
	suite.Require().NoError(err)
	dataSaleKeys = append(dataSaleKeys, dataSaleKey{dataSale2.VerifiableCid, dataSale2.DealId})

	dataSale3 := suite.makeNewDataSale(suite.verifiableCID3)
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale3)
	suite.Require().NoError(err)
	dataSaleKeys = append(dataSaleKeys, dataSaleKey{dataSale3.VerifiableCid, dataSale3.DealId})

	allDataSaleList, err := suite.DataDealKeeper.GetAllDataSaleList(suite.Ctx)
	suite.Require().NoError(err)

	sort.Slice(dataSaleKeys, func(i, j int) bool {
		return dataSaleKeys[i].verifiableCID < dataSaleKeys[j].verifiableCID
	})

	sort.Slice(allDataSaleList, func(i, j int) bool {
		return allDataSaleList[i].VerifiableCid < allDataSaleList[j].VerifiableCid
	})

	for i, dataSaleKey := range dataSaleKeys {
		dataSale, err := suite.DataDealKeeper.GetDataSale(suite.Ctx, dataSaleKey.verifiableCID, dataSaleKey.dealID)
		suite.Require().NoError(err)

		suite.Require().Equal(dataSale.VerifiableCid, allDataSaleList[i].VerifiableCid)
		suite.Require().Equal(dataSale.DealId, allDataSaleList[i].DealId)
		suite.Require().Equal(dataSale.Status, allDataSaleList[i].Status)
		suite.Require().Equal(dataSale.VotingPeriod, allDataSaleList[i].VotingPeriod)
		suite.Require().Equal(dataSale.SellerAddress, allDataSaleList[i].SellerAddress)
	}
}
