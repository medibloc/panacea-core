package keeper_test

import (
	"encoding/base64"
	"sort"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datadeal/testutil"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
)

type dealTestSuite struct {
	testutil.DataDealBaseTestSuite

	defaultFunds sdk.Coins

	sellerAccPrivKey cryptotypes.PrivKey
	sellerAccPubKey  cryptotypes.PubKey
	sellerAccAddr    sdk.AccAddress

	oraclePrivKey *btcec.PrivateKey
	oraclePubKey  *btcec.PublicKey

	oracleAccPrivKey cryptotypes.PrivKey
	oracleAccPubKey  cryptotypes.PubKey
	oracleAccAddr    sdk.AccAddress

	verifiableCID1 string
	verifiableCID2 string
	verifiableCID3 string

	buyerAccAddr sdk.AccAddress
}

func TestDealTestSuite(t *testing.T) {
	suite.Run(t, new(dealTestSuite))
}

func (suite *dealTestSuite) BeforeTest(_, _ string) {
	suite.verifiableCID1 = "verifiableCID"
	suite.verifiableCID2 = "verifiableCID2"
	suite.verifiableCID3 = "verifiableCID3"

	suite.buyerAccAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	suite.defaultFunds = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)))

	testDeal := suite.MakeTestDeal(1, suite.buyerAccAddr)
	err := suite.DataDealKeeper.SetNextDealNumber(suite.Ctx, 2)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDeal(suite.Ctx, testDeal)
	suite.Require().NoError(err)

	suite.oracleAccPrivKey = secp256k1.GenPrivKey()
	suite.oracleAccPubKey = suite.oracleAccPrivKey.PubKey()
	suite.oracleAccAddr = sdk.AccAddress(suite.oracleAccPubKey.Address())

	suite.oraclePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.oraclePubKey = suite.oraclePrivKey.PubKey()

	suite.sellerAccPrivKey = secp256k1.GenPrivKey()
	suite.sellerAccPubKey = suite.sellerAccPrivKey.PubKey()
	suite.sellerAccAddr = sdk.AccAddress(suite.sellerAccPubKey.Address())

	suite.OracleKeeper.SetParams(suite.Ctx, oracletypes.Params{
		OraclePublicKey:          base64.StdEncoding.EncodeToString(suite.oraclePubKey.SerializeCompressed()),
		OraclePubKeyRemoteReport: "",
		UniqueId:                 "",
		OracleCommissionRate:     sdk.NewDecWithPrec(1, 1),
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

func (suite *dealTestSuite) TestCreateNewDeal() {

	err := suite.FundAccount(suite.Ctx, suite.buyerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	budget := &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(10000000)}

	msgCreateDeal := &types.MsgCreateDeal{
		DataSchema:   []string{"http://jsonld.com"},
		Budget:       budget,
		MaxNumData:   10000,
		BuyerAddress: suite.buyerAccAddr.String(),
	}

	buyer, err := sdk.AccAddressFromBech32(msgCreateDeal.BuyerAddress)
	suite.Require().NoError(err)

	dealID, err := suite.DataDealKeeper.CreateDeal(suite.Ctx, buyer, msgCreateDeal)
	suite.Require().NoError(err)

	expectedId, err := suite.DataDealKeeper.GetNextDealNumberAndIncrement(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(dealID, expectedId-uint64(1))

	deal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, dealID)
	suite.Require().NoError(err)
	suite.Require().Equal(deal.GetDataSchema(), msgCreateDeal.GetDataSchema())
	suite.Require().Equal(deal.GetBudget(), msgCreateDeal.GetBudget())
	suite.Require().Equal(deal.GetMaxNumData(), msgCreateDeal.GetMaxNumData())
	suite.Require().Equal(deal.GetBuyerAddress(), msgCreateDeal.GetBuyerAddress())
	suite.Require().Equal(deal.GetStatus(), types.DEAL_STATUS_ACTIVE)
}

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
	newDataSale := suite.MakeNewDataSale(suite.sellerAccAddr, suite.verifiableCID1)

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
	newDataSale := suite.MakeNewDataSale(suite.sellerAccAddr, suite.verifiableCID1)
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
	newDataSale := suite.MakeNewDataSale(suite.sellerAccAddr, suite.verifiableCID1)
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

func (suite dealTestSuite) TestSellDataDealNotExists() {
	msgSellData := &types.MsgSellData{
		DealId:        2,
		VerifiableCid: suite.verifiableCID1,
		SellerAddress: suite.sellerAccAddr.String(),
	}

	err := suite.DataDealKeeper.SellData(suite.Ctx, msgSellData)
	suite.Require().Error(err, types.ErrSellData)
}

func (suite dealTestSuite) TestSellDataDealStatusNotActive() {
	msgSellData := &types.MsgSellData{
		DealId:        1,
		VerifiableCid: suite.verifiableCID1,
		SellerAddress: suite.sellerAccAddr.String(),
	}

	deal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, msgSellData.DealId)
	suite.Require().NoError(err)

	deal.Status = types.DEAL_STATUS_INACTIVE
	err = suite.DataDealKeeper.SetDeal(suite.Ctx, deal)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.SellData(suite.Ctx, msgSellData)
	suite.Require().Error(err, types.ErrSellData)

	deal.Status = types.DEAL_STATUS_COMPLETED
	err = suite.DataDealKeeper.SetDeal(suite.Ctx, deal)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.SellData(suite.Ctx, msgSellData)
	suite.Require().Error(err, types.ErrSellData)
}

func (suite dealTestSuite) TestGetAllDataSalesList() {
	type dataSaleKey struct {
		verifiableCID string
		dealID        uint64
	}
	dataSaleKeys := make([]dataSaleKey, 0)

	dataSale1 := suite.MakeNewDataSale(suite.sellerAccAddr, suite.verifiableCID1)
	err := suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale1)
	suite.Require().NoError(err)
	dataSaleKeys = append(dataSaleKeys, dataSaleKey{dataSale1.VerifiableCid, dataSale1.DealId})

	dataSale2 := suite.MakeNewDataSale(suite.sellerAccAddr, suite.verifiableCID2)
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale2)
	suite.Require().NoError(err)
	dataSaleKeys = append(dataSaleKeys, dataSaleKey{dataSale2.VerifiableCid, dataSale2.DealId})

	dataSale3 := suite.MakeNewDataSale(suite.sellerAccAddr, suite.verifiableCID3)
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

func (suite dealTestSuite) TestDataVerificationVoteSuccess() {
	oracleAccAddr := sdk.AccAddress(suite.oracleAccPubKey.Address().Bytes())
	oracleAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, oracleAccAddr)
	suite.Require().NoError(oracleAccount.SetPubKey(suite.oracleAccPubKey))
	suite.AccountKeeper.SetAccount(suite.Ctx, oracleAccount)

	err := suite.OracleKeeper.SetOracle(suite.Ctx, &oracletypes.Oracle{
		Address:  suite.oracleAccAddr.String(),
		Status:   oracletypes.ORACLE_STATUS_ACTIVE,
		Uptime:   0,
		JailedAt: nil,
	})
	suite.Require().NoError(err)

	dataSale := suite.MakeNewDataSale(suite.sellerAccAddr, suite.verifiableCID1)
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale)
	suite.Require().NoError(err)

	dataVerificationVote := &types.DataVerificationVote{
		VoterAddress:  suite.oracleAccAddr.String(),
		DealId:        1,
		VerifiableCid: suite.verifiableCID1,
		VoteOption:    oracletypes.VOTE_OPTION_YES,
	}

	voteBz, err := suite.Cdc.Marshaler.Marshal(dataVerificationVote)
	suite.Require().NoError(err)

	oraclePrivKeySecp256k1 := secp256k1.PrivKey{
		Key: suite.oraclePrivKey.Serialize(),
	}
	signature, err := oraclePrivKeySecp256k1.Sign(voteBz)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.VoteDataVerification(suite.Ctx, dataVerificationVote, signature)
	suite.Require().NoError(err)

	getDataVerificationVote, err := suite.DataDealKeeper.GetDataVerificationVote(suite.Ctx, suite.verifiableCID1, suite.oracleAccAddr.String(), 1)
	suite.Require().NoError(err)
	suite.Require().Equal(dataVerificationVote, getDataVerificationVote)
}

func (suite dealTestSuite) TestDataVerificationVoteFailedVerifySignature() {
	oracleAccAddr := sdk.AccAddress(suite.oracleAccPubKey.Address().Bytes())
	oracleAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, oracleAccAddr)
	suite.Require().NoError(oracleAccount.SetPubKey(suite.oracleAccPubKey))
	suite.AccountKeeper.SetAccount(suite.Ctx, oracleAccount)

	dataSale := suite.MakeNewDataSale(suite.sellerAccAddr, suite.verifiableCID1)
	err := suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale)
	suite.Require().NoError(err)

	dataVerificationVote := &types.DataVerificationVote{
		VoterAddress:  suite.oracleAccAddr.String(),
		DealId:        1,
		VerifiableCid: suite.verifiableCID1,
		VoteOption:    oracletypes.VOTE_OPTION_YES,
	}

	voteBz, err := suite.Cdc.Marshaler.Marshal(dataVerificationVote)
	suite.Require().NoError(err)
	invalidVoterPrivKey, err := btcec.NewPrivateKey(btcec.S256())
	suite.Require().NoError(err)

	oraclePrivKeySecp256k1 := secp256k1.PrivKey{
		Key: invalidVoterPrivKey.Serialize(),
	}
	signature, err := oraclePrivKeySecp256k1.Sign(voteBz)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.VoteDataVerification(suite.Ctx, dataVerificationVote, signature)
	suite.Require().Error(err, oracletypes.ErrDetectionMaliciousBehavior)
}

func (suite dealTestSuite) TestDataVerificationInvalidDataSaleStatus() {
	oracleAccAddr := sdk.AccAddress(suite.oracleAccPubKey.Address().Bytes())
	oracleAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, oracleAccAddr)
	suite.Require().NoError(oracleAccount.SetPubKey(suite.oracleAccPubKey))
	suite.AccountKeeper.SetAccount(suite.Ctx, oracleAccount)

	err := suite.OracleKeeper.SetOracle(suite.Ctx, &oracletypes.Oracle{
		Address:  suite.oracleAccAddr.String(),
		Status:   oracletypes.ORACLE_STATUS_ACTIVE,
		Uptime:   0,
		JailedAt: nil,
	})
	suite.Require().NoError(err)

	dataSale := suite.MakeNewDataSale(suite.sellerAccAddr, suite.verifiableCID1)
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale)
	suite.Require().NoError(err)

	getDataSale, err := suite.DataDealKeeper.GetDataSale(suite.Ctx, suite.verifiableCID1, 1)
	suite.Require().NoError(err)
	getDataSale.Status = types.DATA_SALE_STATUS_COMPLETED
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, getDataSale)
	suite.Require().NoError(err)

	dataVerificationVote := &types.DataVerificationVote{
		VoterAddress:  suite.oracleAccAddr.String(),
		DealId:        1,
		VerifiableCid: suite.verifiableCID1,
		VoteOption:    oracletypes.VOTE_OPTION_YES,
	}

	voteBz, err := suite.Cdc.Marshaler.Marshal(dataVerificationVote)
	suite.Require().NoError(err)
	oraclePrivKeySecp256k1 := secp256k1.PrivKey{
		Key: suite.oraclePrivKey.Serialize(),
	}
	signature, err := oraclePrivKeySecp256k1.Sign(voteBz)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.VoteDataVerification(suite.Ctx, dataVerificationVote, signature)
	suite.Require().Error(err, types.ErrDataVerificationVote)
	suite.Require().ErrorContains(err, "the current voted data's status is not 'VERIFICATION_VOTING_PERIOD'")

	getDataSale.Status = types.DATA_SALE_STATUS_DELIVERY_VOTING_PERIOD
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, getDataSale)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.VoteDataVerification(suite.Ctx, dataVerificationVote, signature)
	suite.Require().Error(err, types.ErrDataVerificationVote)
	suite.Require().ErrorContains(err, "the current voted data's status is not 'VERIFICATION_VOTING_PERIOD'")

	getDataSale.Status = types.DATA_SALE_STATUS_FAILED
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, getDataSale)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.VoteDataVerification(suite.Ctx, dataVerificationVote, signature)
	suite.Require().Error(err, types.ErrDataVerificationVote)
	suite.Require().ErrorContains(err, "the current voted data's status is not 'VERIFICATION_VOTING_PERIOD'")
}

func (suite dealTestSuite) TestDataVerificationInvalidGenesisOracleStatus() {
	oracleAccAddr := sdk.AccAddress(suite.oracleAccPubKey.Address().Bytes())
	oracleAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, oracleAccAddr)
	suite.Require().NoError(oracleAccount.SetPubKey(suite.oracleAccPubKey))
	suite.AccountKeeper.SetAccount(suite.Ctx, oracleAccount)

	err := suite.OracleKeeper.SetOracle(suite.Ctx, &oracletypes.Oracle{
		Address:  suite.oracleAccAddr.String(),
		Status:   oracletypes.ORACLE_STATUS_JAILED,
		Uptime:   0,
		JailedAt: nil,
	})
	suite.Require().NoError(err)

	dataSale := suite.MakeNewDataSale(suite.sellerAccAddr, suite.verifiableCID1)
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale)
	suite.Require().NoError(err)

	dataVerificationVote := &types.DataVerificationVote{
		VoterAddress:  suite.oracleAccAddr.String(),
		DealId:        1,
		VerifiableCid: suite.verifiableCID1,
		VoteOption:    oracletypes.VOTE_OPTION_YES,
	}

	voteBz, err := suite.Cdc.Marshaler.Marshal(dataVerificationVote)
	suite.Require().NoError(err)
	oraclePrivKeySecp256k1 := secp256k1.PrivKey{
		Key: suite.oraclePrivKey.Serialize(),
	}
	signature, err := oraclePrivKeySecp256k1.Sign(voteBz)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.VoteDataVerification(suite.Ctx, dataVerificationVote, signature)
	suite.Require().Error(err, types.ErrDataVerificationVote)
	suite.Require().ErrorContains(err, "this oracle is not in 'ACTIVE' state")
}
