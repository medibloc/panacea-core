package keeper_test

import (
	"encoding/base64"
	"sort"
	"strconv"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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

	dataHash1 string
	dataHash2 string
	dataHash3 string

	buyerAccAddr sdk.AccAddress
}

func TestDealTestSuite(t *testing.T) {
	suite.Run(t, new(dealTestSuite))
}

func (suite *dealTestSuite) BeforeTest(_, _ string) {
	suite.verifiableCID1 = "verifiableCID"
	suite.verifiableCID2 = "verifiableCID2"
	suite.verifiableCID3 = "verifiableCID3"

	suite.dataHash1 = "dataHash1"
	suite.dataHash2 = "dataHash2"
	suite.dataHash3 = "dataHash3"

	suite.buyerAccAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	suite.defaultFunds = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)))

	testDeal := suite.MakeTestDeal(1, suite.buyerAccAddr, 100)
	err := suite.DataDealKeeper.SetNextDealNumber(suite.Ctx, 2)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDeal(suite.Ctx, testDeal)
	suite.Require().NoError(err)

	suite.oraclePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.oraclePubKey = suite.oraclePrivKey.PubKey()

	suite.oracleAccPrivKey = secp256k1.GenPrivKey()
	suite.oracleAccPubKey = suite.oracleAccPrivKey.PubKey()
	suite.oracleAccAddr = sdk.AccAddress(suite.oracleAccPubKey.Address())

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
	suite.DataDealKeeper.SetParams(suite.Ctx, types.DefaultParams())

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

func (suite *dealTestSuite) TestCheckDealCurNumDataAndIncrement() {
	err := suite.FundAccount(suite.Ctx, suite.buyerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	budget := &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(10000000)}

	msgCreateDeal := &types.MsgCreateDeal{
		DataSchema:   []string{"http://jsonld.com"},
		Budget:       budget,
		MaxNumData:   1,
		BuyerAddress: suite.buyerAccAddr.String(),
	}

	buyer, err := sdk.AccAddressFromBech32(msgCreateDeal.BuyerAddress)
	suite.Require().NoError(err)

	dealID, err := suite.DataDealKeeper.CreateDeal(suite.Ctx, buyer, msgCreateDeal)
	suite.Require().NoError(err)

	check, err := suite.DataDealKeeper.IsDealCompleted(suite.Ctx, dealID)
	suite.Require().NoError(err)
	suite.Equal(false, check)

	err = suite.DataDealKeeper.IncrementCurNumDataAtDeal(suite.Ctx, dealID)
	suite.Require().NoError(err)
	updatedDeal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, dealID)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), updatedDeal.CurNumData)

	check, err = suite.DataDealKeeper.IsDealCompleted(suite.Ctx, dealID)
	suite.Require().NoError(err)
	suite.Require().Equal(true, check)

}

func (suite *dealTestSuite) TestSellDataSuccess() {
	msgSellData := &types.MsgSellData{
		DealId:        1,
		VerifiableCid: suite.verifiableCID1,
		DataHash:      suite.dataHash1,
		SellerAddress: suite.sellerAccAddr.String(),
	}

	err := suite.DataDealKeeper.SellData(suite.Ctx, msgSellData)
	suite.Require().NoError(err)

	dataSale, err := suite.DataDealKeeper.GetDataSale(suite.Ctx, suite.dataHash1, uint64(1))
	suite.Require().NoError(err)

	suite.Require().Equal(dataSale.VerifiableCid, suite.verifiableCID1)
	suite.Require().Equal(dataSale.DealId, uint64(1))
	suite.Require().Equal(dataSale.VerificationVotingPeriod, suite.OracleKeeper.GetVotingPeriod(suite.Ctx))
	suite.Require().Equal(dataSale.Status, types.DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD)
	suite.Require().Equal(dataSale.SellerAddress, suite.sellerAccAddr.String())
	suite.Require().Equal(dataSale.DataHash, suite.dataHash1)

	events := suite.Ctx.EventManager().Events()

	suite.Require().Equal(3, len(events[0].Attributes))
	suite.Require().Equal(types.EventTypeDataVerificationVote, events[0].Type)
	suite.Require().Equal(types.AttributeKeyVoteStatus, string(events[0].Attributes[0].Key))
	suite.Require().Equal(types.AttributeValueVoteStatusStarted, string(events[0].Attributes[0].Value))
	suite.Require().Equal(types.AttributeKeyDealID, string(events[0].Attributes[1].Key))
	suite.Require().Equal(types.AttributeKeyDataHash, string(events[0].Attributes[2].Key))
}

func (suite *dealTestSuite) TestReSellData() {
	newDataSale := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID1)

	newDataSale.Status = types.DATA_SALE_STATUS_VERIFICATION_FAILED

	err := suite.DataDealKeeper.SetDataSale(suite.Ctx, newDataSale)
	suite.Require().NoError(err)

	msgSellData := &types.MsgSellData{
		DealId:        1,
		VerifiableCid: newDataSale.VerifiableCid,
		DataHash:      newDataSale.DataHash,
		SellerAddress: newDataSale.SellerAddress,
	}

	err = suite.DataDealKeeper.SellData(suite.Ctx, msgSellData)
	suite.Require().NoError(err)

	dataSale, err := suite.DataDealKeeper.GetDataSale(suite.Ctx, suite.dataHash1, uint64(1))
	suite.Require().NoError(err)
	suite.Require().Equal(dataSale.VerifiableCid, suite.verifiableCID1)

	suite.Require().Equal(dataSale.DealId, uint64(1))
	suite.Require().Equal(dataSale.VerificationVotingPeriod, suite.OracleKeeper.GetVotingPeriod(suite.Ctx))
	suite.Require().Equal(dataSale.Status, types.DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD)
	suite.Require().Equal(dataSale.SellerAddress, suite.sellerAccAddr.String())
	suite.Require().Equal(dataSale.DataHash, suite.dataHash1)

	events := suite.Ctx.EventManager().Events()

	suite.Require().Equal(3, len(events[0].Attributes))
	suite.Require().Equal(types.EventTypeDataVerificationVote, events[0].Type)
	suite.Require().Equal(types.AttributeKeyVoteStatus, string(events[0].Attributes[0].Key))
	suite.Require().Equal(types.AttributeValueVoteStatusStarted, string(events[0].Attributes[0].Value))
	suite.Require().Equal(types.AttributeKeyDealID, string(events[0].Attributes[1].Key))
	suite.Require().Equal(types.AttributeKeyDataHash, string(events[0].Attributes[2].Key))

}

func (suite *dealTestSuite) TestReSellDataFailed() {
	newDataSale := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID1)

	newDataSale.Status = types.DATA_SALE_STATUS_VERIFICATION_FAILED

	err := suite.DataDealKeeper.SetDataSale(suite.Ctx, newDataSale)
	suite.Require().NoError(err)

	sellerAddr2 := secp256k1.GenPrivKey().PubKey().Address()

	msgSellData := &types.MsgSellData{
		DealId:        1,
		VerifiableCid: newDataSale.VerifiableCid,
		DataHash:      newDataSale.DataHash,
		SellerAddress: sellerAddr2.String(),
	}

	err = suite.DataDealKeeper.SellData(suite.Ctx, msgSellData)
	suite.Require().Error(err, types.ErrSellData)
}

func (suite *dealTestSuite) TestReSellDataNotFound() {
	newDataSale := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID1)

	newDataSale.Status = types.DATA_SALE_STATUS_VERIFICATION_FAILED

	err := suite.DataDealKeeper.SetDataSale(suite.Ctx, newDataSale)
	suite.Require().NoError(err)

	msgSellData := &types.MsgSellData{
		DealId:        2,
		VerifiableCid: newDataSale.VerifiableCid,
		DataHash:      newDataSale.DataHash,
		SellerAddress: newDataSale.SellerAddress,
	}

	err = suite.DataDealKeeper.SellData(suite.Ctx, msgSellData)
	suite.Require().Error(err, types.ErrSellData)
}

func (suite *dealTestSuite) TestSellDataStatusVotingPeriod() {
	newDataSale := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID1)
	newDataSale.Status = types.DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD

	err := suite.DataDealKeeper.SetDataSale(suite.Ctx, newDataSale)
	suite.Require().NoError(err)

	msgSellData := &types.MsgSellData{
		DealId:        1,
		VerifiableCid: newDataSale.VerifiableCid,
		DataHash:      newDataSale.DataHash,
		SellerAddress: newDataSale.SellerAddress,
	}

	err = suite.DataDealKeeper.SellData(suite.Ctx, msgSellData)
	suite.Require().Error(err, types.ErrSellData)
}

func (suite *dealTestSuite) TestSellDataStatusCompleted() {
	newDataSale := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID1)
	newDataSale.Status = types.DATA_SALE_STATUS_COMPLETED

	err := suite.DataDealKeeper.SetDataSale(suite.Ctx, newDataSale)
	suite.Require().NoError(err)

	msgSellData := &types.MsgSellData{
		DealId:        1,
		VerifiableCid: newDataSale.VerifiableCid,
		DataHash:      newDataSale.DataHash,
		SellerAddress: newDataSale.SellerAddress,
	}

	err = suite.DataDealKeeper.SellData(suite.Ctx, msgSellData)
	suite.Require().Error(err, types.ErrSellData)
}

func (suite *dealTestSuite) TestSellDataDealNotExists() {
	msgSellData := &types.MsgSellData{
		DealId:        2,
		VerifiableCid: suite.verifiableCID1,
		DataHash:      suite.dataHash1,
		SellerAddress: suite.sellerAccAddr.String(),
	}

	err := suite.DataDealKeeper.SellData(suite.Ctx, msgSellData)
	suite.Require().Error(err, types.ErrSellData)
}

func (suite *dealTestSuite) TestSellDataDealStatusNotActive() {
	msgSellData := &types.MsgSellData{
		DealId:        1,
		VerifiableCid: suite.verifiableCID1,
		DataHash:      suite.dataHash1,
		SellerAddress: suite.sellerAccAddr.String(),
	}

	deal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, msgSellData.DealId)
	suite.Require().NoError(err)

	deal.Status = types.DEAL_STATUS_COMPLETED
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

func (suite *dealTestSuite) TestGetAllDataSalesList() {
	type dataSaleKey struct {
		dataHash string
		dealID   uint64
	}
	dataSaleKeys := make([]dataSaleKey, 0)

	dataSale1 := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID1)
	err := suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale1)
	suite.Require().NoError(err)
	dataSaleKeys = append(dataSaleKeys, dataSaleKey{dataSale1.DataHash, dataSale1.DealId})

	dataSale2 := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash2, suite.verifiableCID2)
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale2)
	suite.Require().NoError(err)
	dataSaleKeys = append(dataSaleKeys, dataSaleKey{dataSale2.DataHash, dataSale2.DealId})

	dataSale3 := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash3, suite.verifiableCID3)
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale3)
	suite.Require().NoError(err)
	dataSaleKeys = append(dataSaleKeys, dataSaleKey{dataSale3.DataHash, dataSale3.DealId})

	allDataSaleList, err := suite.DataDealKeeper.GetAllDataSaleList(suite.Ctx)
	suite.Require().NoError(err)

	sort.Slice(dataSaleKeys, func(i, j int) bool {
		return dataSaleKeys[i].dataHash < dataSaleKeys[j].dataHash
	})

	sort.Slice(allDataSaleList, func(i, j int) bool {
		return allDataSaleList[i].DataHash < allDataSaleList[j].DataHash
	})

	for i, dataSaleKey := range dataSaleKeys {
		dataSale, err := suite.DataDealKeeper.GetDataSale(suite.Ctx, dataSaleKey.dataHash, dataSaleKey.dealID)
		suite.Require().NoError(err)

		suite.Require().Equal(dataSale.VerifiableCid, allDataSaleList[i].VerifiableCid)
		suite.Require().Equal(dataSale.DealId, allDataSaleList[i].DealId)
		suite.Require().Equal(dataSale.Status, allDataSaleList[i].Status)
		suite.Require().Equal(dataSale.VerificationVotingPeriod, allDataSaleList[i].VerificationVotingPeriod)
		suite.Require().Equal(dataSale.SellerAddress, allDataSaleList[i].SellerAddress)
		suite.Require().Equal(dataSale.DataHash, allDataSaleList[i].DataHash)
	}
}

func (suite *dealTestSuite) TestDataVerificationVoteSuccess() {
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

	dataSale := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID1)
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale)
	suite.Require().NoError(err)

	dataVerificationVote := suite.MakeNewDataVerificationVote(suite.oracleAccAddr, suite.dataHash1)

	voteBz, err := suite.Cdc.Marshaler.Marshal(dataVerificationVote)
	suite.Require().NoError(err)

	oraclePrivKeySecp256k1 := secp256k1.PrivKey{
		Key: suite.oraclePrivKey.Serialize(),
	}
	signature, err := oraclePrivKeySecp256k1.Sign(voteBz)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.VoteDataVerification(suite.Ctx, dataVerificationVote, signature)
	suite.Require().NoError(err)

	getDataVerificationVote, err := suite.DataDealKeeper.GetDataVerificationVote(suite.Ctx, suite.dataHash1, suite.oracleAccAddr.String(), 1)
	suite.Require().NoError(err)
	suite.Require().Equal(dataVerificationVote, getDataVerificationVote)
}

func (suite *dealTestSuite) TestDataVerificationVoteFailedVerifySignature() {
	oracleAccAddr := sdk.AccAddress(suite.oracleAccPubKey.Address().Bytes())
	oracleAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, oracleAccAddr)
	suite.Require().NoError(oracleAccount.SetPubKey(suite.oracleAccPubKey))
	suite.AccountKeeper.SetAccount(suite.Ctx, oracleAccount)

	dataSale := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID1)
	err := suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale)
	suite.Require().NoError(err)

	dataVerificationVote := suite.MakeNewDataVerificationVote(suite.oracleAccAddr, suite.dataHash1)

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
	suite.Require().ErrorIs(err, types.ErrDataVerificationVote)
	suite.Require().ErrorContains(err, "failed to signature validation")
}

func (suite *dealTestSuite) TestDataVerificationInvalidDataSaleStatus() {
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

	dataSale := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID1)
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale)
	suite.Require().NoError(err)

	getDataSale, err := suite.DataDealKeeper.GetDataSale(suite.Ctx, suite.dataHash1, 1)
	suite.Require().NoError(err)
	getDataSale.Status = types.DATA_SALE_STATUS_COMPLETED
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, getDataSale)
	suite.Require().NoError(err)

	dataVerificationVote := suite.MakeNewDataVerificationVote(suite.oracleAccAddr, suite.dataHash1)

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

	getDataSale.Status = types.DATA_SALE_STATUS_DELIVERY_FAILED
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, getDataSale)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.VoteDataVerification(suite.Ctx, dataVerificationVote, signature)
	suite.Require().Error(err, types.ErrDataVerificationVote)
	suite.Require().ErrorContains(err, "the current voted data's status is not 'VERIFICATION_VOTING_PERIOD'")
}

func (suite *dealTestSuite) TestDataVerificationInvalidGenesisOracleStatus() {
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

	dataSale := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID1)
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale)
	suite.Require().NoError(err)

	dataVerificationVote := suite.MakeNewDataVerificationVote(suite.oracleAccAddr, suite.dataHash1)

	voteBz, err := suite.Cdc.Marshaler.Marshal(dataVerificationVote)
	suite.Require().NoError(err)
	oraclePrivKeySecp256k1 := secp256k1.PrivKey{
		Key: suite.oraclePrivKey.Serialize(),
	}
	signature, err := oraclePrivKeySecp256k1.Sign(voteBz)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.VoteDataVerification(suite.Ctx, dataVerificationVote, signature)
	suite.Require().Error(err, types.ErrDataVerificationVote)
	suite.Require().Error(err, types.ErrOracleNotActive)
}

func (suite *dealTestSuite) TestGetAllDataVerificationVoteList() {
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

	type dataVerificationVoteKey struct {
		dealID       uint64
		dataHash     string
		voterAddress sdk.AccAddress
	}
	dataVerificationVoteKeys := make([]dataVerificationVoteKey, 0)

	verificationVote1 := suite.MakeNewDataVerificationVote(suite.oracleAccAddr, suite.dataHash1)
	err = suite.DataDealKeeper.SetDataVerificationVote(suite.Ctx, verificationVote1)
	suite.Require().NoError(err)
	voterAcc, err := sdk.AccAddressFromBech32(verificationVote1.VoterAddress)
	suite.Require().NoError(err)
	dataVerificationVoteKeys = append(dataVerificationVoteKeys, dataVerificationVoteKey{verificationVote1.DealId, verificationVote1.DataHash, voterAcc})

	verificationVote2 := suite.MakeNewDataVerificationVote(suite.oracleAccAddr, suite.dataHash2)
	err = suite.DataDealKeeper.SetDataVerificationVote(suite.Ctx, verificationVote2)
	suite.Require().NoError(err)
	voterAcc, err = sdk.AccAddressFromBech32(verificationVote2.VoterAddress)
	suite.Require().NoError(err)
	dataVerificationVoteKeys = append(dataVerificationVoteKeys, dataVerificationVoteKey{verificationVote2.DealId, verificationVote2.DataHash, voterAcc})

	verificationVote3 := suite.MakeNewDataVerificationVote(suite.oracleAccAddr, suite.dataHash3)
	err = suite.DataDealKeeper.SetDataVerificationVote(suite.Ctx, verificationVote3)
	suite.Require().NoError(err)
	voterAcc, err = sdk.AccAddressFromBech32(verificationVote3.VoterAddress)
	suite.Require().NoError(err)
	dataVerificationVoteKeys = append(dataVerificationVoteKeys, dataVerificationVoteKey{verificationVote3.DealId, verificationVote3.DataHash, voterAcc})

	allDataVerificationVoteList, err := suite.DataDealKeeper.GetAllDataVerificationVoteList(suite.Ctx)
	suite.Require().NoError(err)

	sort.Slice(dataVerificationVoteKeys, func(i, j int) bool {
		return dataVerificationVoteKeys[i].dataHash < dataVerificationVoteKeys[j].dataHash
	})

	sort.Slice(allDataVerificationVoteList, func(i, j int) bool {
		return allDataVerificationVoteList[i].DataHash < allDataVerificationVoteList[j].DataHash
	})

	for i, dataVerificationVoteKey := range dataVerificationVoteKeys {
		dataVerificationVote, err := suite.DataDealKeeper.GetDataVerificationVote(suite.Ctx, dataVerificationVoteKey.dataHash, dataVerificationVoteKey.voterAddress.String(), dataVerificationVoteKey.dealID)
		suite.Require().NoError(err)

		suite.Require().Equal(dataVerificationVote.DataHash, allDataVerificationVoteList[i].DataHash)
		suite.Require().Equal(dataVerificationVote.DealId, allDataVerificationVoteList[i].DealId)
		suite.Require().Equal(dataVerificationVote.VoterAddress, allDataVerificationVoteList[i].VoterAddress)
		suite.Require().Equal(dataVerificationVote.VoteOption, allDataVerificationVoteList[i].VoteOption)
	}
}

func (suite *dealTestSuite) TestDataDeliveryVoteSuccess() {
	ctx := suite.Ctx

	oracleAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, suite.oracleAccAddr)
	suite.Require().NoError(oracleAccount.SetPubKey(suite.oracleAccPubKey))
	suite.AccountKeeper.SetAccount(suite.Ctx, oracleAccount)

	err := suite.OracleKeeper.SetOracle(suite.Ctx, &oracletypes.Oracle{
		Address:  suite.oracleAccAddr.String(),
		Status:   oracletypes.ORACLE_STATUS_ACTIVE,
		Uptime:   0,
		JailedAt: nil,
	})
	suite.Require().NoError(err)

	dataSale := suite.MakeNewDataSaleDeliveryVoting(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID1)
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale)
	suite.Require().NoError(err)

	dataDeliveryVote := &types.DataDeliveryVote{
		VoterAddress: suite.oracleAccAddr.String(),
		DealId:       dataSale.DealId,
		DataHash:     dataSale.DataHash,
		DeliveredCid: "test",
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	voteBz, err := suite.Cdc.Marshaler.Marshal(dataDeliveryVote)
	suite.Require().NoError(err)

	oraclePrivKeySecp256k1 := secp256k1.PrivKey{
		Key: suite.oraclePrivKey.Serialize(),
	}

	signature, err := oraclePrivKeySecp256k1.Sign(voteBz)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.VoteDataDelivery(ctx, dataDeliveryVote, signature)
	suite.Require().NoError(err)

	getDataDeliveryVote, err := suite.DataDealKeeper.GetDataDeliveryVote(
		ctx,
		suite.dataHash1,
		suite.oracleAccAddr.String(),
		dataSale.DealId,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(dataDeliveryVote, getDataDeliveryVote)
}

func (suite *dealTestSuite) TestDataDeliveryVoteFailedVerifySignature() {
	ctx := suite.Ctx

	oracleAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, suite.oracleAccAddr)
	suite.Require().NoError(oracleAccount.SetPubKey(suite.oracleAccPubKey))
	suite.AccountKeeper.SetAccount(suite.Ctx, oracleAccount)

	err := suite.OracleKeeper.SetOracle(suite.Ctx, &oracletypes.Oracle{
		Address:  suite.oracleAccAddr.String(),
		Status:   oracletypes.ORACLE_STATUS_ACTIVE,
		Uptime:   0,
		JailedAt: nil,
	})
	suite.Require().NoError(err)

	dataSale := suite.MakeNewDataSaleDeliveryVoting(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID1)
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale)
	suite.Require().NoError(err)

	dataDeliveryVote := &types.DataDeliveryVote{
		VoterAddress: suite.oracleAccAddr.String(),
		DealId:       dataSale.DealId,
		DataHash:     dataSale.DataHash,
		DeliveredCid: "test",
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	voteBz, err := suite.Cdc.Marshaler.Marshal(dataDeliveryVote)
	suite.Require().NoError(err)

	invalidPrivKey := secp256k1.GenPrivKey()
	// sign with invalid priv key
	signature, err := invalidPrivKey.Sign(voteBz)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.VoteDataDelivery(ctx, dataDeliveryVote, signature)
	suite.Require().ErrorIs(err, types.ErrDataDeliveryVote)
	suite.Require().ErrorContains(err, "failed to signature validation")
}

func (suite *dealTestSuite) TestDataDeliveryVoteFailedInvalidStatus() {
	ctx := suite.Ctx

	oracleAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, suite.oracleAccAddr)
	suite.Require().NoError(oracleAccount.SetPubKey(suite.oracleAccPubKey))
	suite.AccountKeeper.SetAccount(suite.Ctx, oracleAccount)

	err := suite.OracleKeeper.SetOracle(suite.Ctx, &oracletypes.Oracle{
		Address:  suite.oracleAccAddr.String(),
		Status:   oracletypes.ORACLE_STATUS_ACTIVE,
		Uptime:   0,
		JailedAt: nil,
	})
	suite.Require().NoError(err)

	// dataSale that status is DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD
	dataSale := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID1)
	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale)
	suite.Require().NoError(err)

	dataDeliveryVote := &types.DataDeliveryVote{
		VoterAddress: suite.oracleAccAddr.String(),
		DealId:       dataSale.DealId,
		DataHash:     dataSale.DataHash,
		DeliveredCid: "test",
		VoteOption:   oracletypes.VOTE_OPTION_YES,
	}

	voteBz, err := suite.Cdc.Marshaler.Marshal(dataDeliveryVote)
	suite.Require().NoError(err)

	oraclePrivKeySecp256k1 := secp256k1.PrivKey{
		Key: suite.oraclePrivKey.Serialize(),
	}

	signature, err := oraclePrivKeySecp256k1.Sign(voteBz)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.VoteDataDelivery(ctx, dataDeliveryVote, signature)
	suite.Require().ErrorIs(err, types.ErrDataDeliveryVote)
}

func (suite *dealTestSuite) TestRequestDeactivateDeal() {
	ctx := suite.Ctx

	err := suite.FundAccount(ctx, suite.buyerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	getDeal, err := suite.DataDealKeeper.GetDeal(ctx, 1)
	suite.Require().NoError(err)

	dealAccAddr, err := sdk.AccAddressFromBech32(getDeal.Address)
	suite.Require().NoError(err)

	// Sending Budget from buyer to deal
	err = suite.BankKeeper.SendCoins(suite.Ctx, suite.buyerAccAddr, dealAccAddr, sdk.NewCoins(*getDeal.Budget))
	suite.Require().NoError(err)

	msgDeactivateDeal := &types.MsgDeactivateDeal{
		DealId:           1,
		RequesterAddress: suite.buyerAccAddr.String(),
	}

	err = suite.DataDealKeeper.RequestDeactivateDeal(ctx, msgDeactivateDeal)
	suite.Require().NoError(err)

	getDeal, err = suite.DataDealKeeper.GetDeal(ctx, 1)
	suite.Require().NoError(err)

	suite.Require().Equal(getDeal.Status, types.DEAL_STATUS_DEACTIVATING)

}

func (suite *dealTestSuite) TestRequestDeactivateDealInvalidRequester() {
	ctx := suite.Ctx

	msgDeactivateDeal := &types.MsgDeactivateDeal{
		DealId:           1,
		RequesterAddress: suite.sellerAccAddr.String(),
	}

	err := suite.DataDealKeeper.RequestDeactivateDeal(ctx, msgDeactivateDeal)
	suite.Require().ErrorIs(err, types.ErrDealDeactivate)
	suite.Require().ErrorContains(err, "only buyer can deactivate deal")
}

func (suite *dealTestSuite) TestRequestDeactivateDealStatusNotActive() {
	ctx := suite.Ctx

	getDeal, err := suite.DataDealKeeper.GetDeal(ctx, 1)
	suite.Require().NoError(err)

	getDeal.Status = types.DEAL_STATUS_COMPLETED

	err = suite.DataDealKeeper.SetDeal(ctx, getDeal)
	suite.Require().NoError(err)

	msgDeactivateDeal := &types.MsgDeactivateDeal{
		DealId:           1,
		RequesterAddress: suite.buyerAccAddr.String(),
	}

	err = suite.DataDealKeeper.RequestDeactivateDeal(ctx, msgDeactivateDeal)
	suite.Require().ErrorIs(err, types.ErrDealDeactivate)
	suite.Require().ErrorContains(err, "deal's status is not 'ACTIVE'")
}

func (suite *dealTestSuite) TestRequestDataDeliveryVoteSuccess() {
	ctx := suite.Ctx

	dataSale := suite.MakeNewDataSaleDeliveryFailed(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID1)
	err := suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale)
	suite.Require().NoError(err)

	msgRequestDataDeliveryVote := &types.MsgReRequestDataDeliveryVote{
		DealId:           1,
		DataHash:         dataSale.DataHash,
		RequesterAddress: suite.buyerAccAddr.String(),
	}

	err = suite.DataDealKeeper.ReRequestDataDeliveryVote(ctx, msgRequestDataDeliveryVote)
	suite.Require().NoError(err)

	dataSale, err = suite.DataDealKeeper.GetDataSale(ctx, dataSale.DataHash, 1)
	suite.Require().NoError(err)

	suite.Require().Equal(types.DATA_SALE_STATUS_DELIVERY_VOTING_PERIOD, dataSale.Status)

	events := ctx.EventManager().Events()
	requiredEvents := map[string]bool{
		types.EventTypeDataDeliveryVote: false,
	}

	for _, e := range events {
		if e.Type == types.EventTypeDataDeliveryVote {
			requiredEvents[e.Type] = true
			suite.Require().Equal(3, len(e.Attributes))
			suite.Require().Equal(types.AttributeKeyVoteStatus, string(e.Attributes[0].Key))
			suite.Require().Equal(types.AttributeValueVoteStatusStarted, string(e.Attributes[0].Value))
			suite.Require().Equal(types.AttributeKeyDataHash, string(e.Attributes[1].Key))
			suite.Require().Equal(types.AttributeKeyDealID, string(e.Attributes[2].Key))
		}
	}

	for _, v := range requiredEvents {
		suite.Require().True(v)
	}
}

func (suite *dealTestSuite) TestRequestDataDeliveryVoteFailedInvalidStatus() {
	ctx := suite.Ctx

	dataSale := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID1)
	err := suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSale)
	suite.Require().NoError(err)

	msgReRequestDataDeliveryVote := &types.MsgReRequestDataDeliveryVote{
		DealId:           1,
		DataHash:         dataSale.DataHash,
		RequesterAddress: suite.buyerAccAddr.String(),
	}

	err = suite.DataDealKeeper.ReRequestDataDeliveryVote(ctx, msgReRequestDataDeliveryVote)
	suite.Require().ErrorIs(err, types.ErrReRequestDataDeliveryVote)
}

func (suite *dealTestSuite) TestDeactivateDeal() {
	ctx := suite.Ctx

	err := suite.FundAccount(suite.Ctx, suite.buyerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	deactivatingDeal := &types.Deal{
		Id:           2,
		Address:      types.NewDealAddress(2).String(),
		DataSchema:   []string{"http://jsonld.com"},
		Budget:       &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(1000000000)},
		MaxNumData:   100,
		CurNumData:   0,
		BuyerAddress: suite.buyerAccAddr.String(),
		Status:       types.DEAL_STATUS_DEACTIVATING,
	}

	coins := sdk.NewCoins(*deactivatingDeal.Budget)

	dealAddress, err := sdk.AccAddressFromBech32(deactivatingDeal.Address)
	suite.Require().NoError(err)

	acc := suite.AccountKeeper.NewAccount(ctx, authtypes.NewModuleAccount(
		authtypes.NewBaseAccountWithAddress(
			dealAddress,
		),
		"deal"+strconv.FormatUint(deactivatingDeal.Id, 10)),
	)
	suite.AccountKeeper.SetAccount(ctx, acc)

	err = suite.BankKeeper.SendCoins(ctx, suite.buyerAccAddr, dealAddress, coins)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.SetDeal(ctx, deactivatingDeal)
	suite.Require().NoError(err)

	beforeBuyerBalance := suite.BankKeeper.GetBalance(ctx, suite.buyerAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(9000000000)), beforeBuyerBalance)

	err = suite.DataDealKeeper.DeactivateDeal(ctx, 2)
	suite.Require().NoError(err)

	afterBuyerBalance := suite.BankKeeper.GetBalance(ctx, suite.buyerAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)), afterBuyerBalance)

	getDeal, err := suite.DataDealKeeper.GetDeal(ctx, 2)
	suite.Require().NoError(err)

	suite.Require().Equal(getDeal.Status, types.DEAL_STATUS_DEACTIVATED)

}
