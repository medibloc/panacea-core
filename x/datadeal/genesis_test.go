package datadeal_test

import (
	"encoding/base64"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datadeal"
	"github.com/medibloc/panacea-core/v2/x/datadeal/testutil"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
)

type genesisTestSuite struct {
	testutil.DataDealBaseTestSuite

	buyerAccAddr  sdk.AccAddress
	sellerAccAddr sdk.AccAddress
	oracleAccAddr sdk.AccAddress
	defaultFunds  sdk.Coins

	oraclePrivKey *btcec.PrivateKey
	oraclePubKey  *btcec.PublicKey

	verifiableCID1 string
	verifiableCID2 string

	dataHash1 string
	dataHash2 string
	dataHash3 string
	dataHash4 string

	deliveryCID1 string
	deliveryCID2 string
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}

func (suite *genesisTestSuite) BeforeTest(_, _ string) {

	suite.buyerAccAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	suite.sellerAccAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	suite.oracleAccAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	suite.defaultFunds = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)))
	suite.verifiableCID1 = "verifiableCID"
	suite.verifiableCID2 = "verifiableCID2"
	suite.dataHash1 = "dataHash1"
	suite.dataHash2 = "dataHash2"
	suite.dataHash3 = "dataHash3"
	suite.dataHash4 = "dataHash4"
	suite.deliveryCID1 = "deliveryCID"
	suite.deliveryCID2 = "deliveryCID2"

	suite.oraclePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.oraclePubKey = suite.oraclePrivKey.PubKey()

	err := suite.OracleKeeper.SetOracle(suite.Ctx, &oracletypes.Oracle{
		Address:  suite.oracleAccAddr.String(),
		Status:   oracletypes.ORACLE_STATUS_ACTIVE,
		Uptime:   0,
		JailedAt: nil,
	})
	suite.Require().NoError(err)

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

func (suite *genesisTestSuite) TestInitGenesis() {
	deal1 := suite.MakeTestDeal(1, suite.buyerAccAddr, 100)
	deal2 := suite.MakeTestDeal(2, suite.buyerAccAddr, 100)

	dataSale1 := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID1)
	dataSale2 := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash2, suite.verifiableCID2)

	dataSaleDeliveryVoting1 := suite.MakeNewDataSaleDeliveryVoting(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID1)
	dataSaleDeliveryVoting2 := suite.MakeNewDataSaleDeliveryVoting(suite.sellerAccAddr, suite.dataHash2, suite.verifiableCID2)

	dataVerificationVote1 := suite.MakeNewDataVerificationVote(suite.oracleAccAddr, suite.dataHash1)
	dataVerificationVote2 := suite.MakeNewDataVerificationVote(suite.oracleAccAddr, suite.dataHash2)

	dataDeliveryVote1 := suite.MakeNewDataDeliveryVote(suite.oracleAccAddr, dataSale1.DataHash, suite.deliveryCID1, deal1.Id)
	dataDeliveryVote2 := suite.MakeNewDataDeliveryVote(suite.oracleAccAddr, dataSale2.DataHash, suite.deliveryCID2, deal2.Id)

	verificationQueue1 := &types.DataVerificationQueueElement{
		DataHash:      dataSale1.DataHash,
		DealId:        dataSale1.DealId,
		VotingEndTime: dataSale1.VerificationVotingPeriod.VotingEndTime,
	}

	verificationQueue2 := &types.DataVerificationQueueElement{
		DataHash:      dataSale2.DataHash,
		DealId:        dataSale2.DealId,
		VotingEndTime: dataSale2.VerificationVotingPeriod.VotingEndTime,
	}

	deliveryQueue1 := &types.DataDeliveryQueueElement{
		DataHash:      dataSaleDeliveryVoting1.DataHash,
		DealId:        dataSaleDeliveryVoting1.DealId,
		VotingEndTime: dataSaleDeliveryVoting1.DeliveryVotingPeriod.VotingEndTime,
	}

	deliveryQueue2 := &types.DataDeliveryQueueElement{
		DataHash:      dataSaleDeliveryVoting2.DataHash,
		DealId:        dataSaleDeliveryVoting2.DealId,
		VotingEndTime: dataSaleDeliveryVoting2.DeliveryVotingPeriod.VotingEndTime,
	}

	genesis := types.GenesisState{
		Deals:                         []types.Deal{*deal1, *deal2},
		NextDealNumber:                3,
		DataSales:                     []types.DataSale{*dataSale1, *dataSale2},
		DataVerificationVotes:         []types.DataVerificationVote{*dataVerificationVote1, *dataVerificationVote2},
		DataDeliveryVotes:             []types.DataDeliveryVote{*dataDeliveryVote1, *dataDeliveryVote2},
		DataVerificationQueueElements: []types.DataVerificationQueueElement{*verificationQueue1, *verificationQueue2},
		DataDeliveryQueueElements:     []types.DataDeliveryQueueElement{*deliveryQueue1, *deliveryQueue2},
	}

	datadeal.InitGenesis(suite.Ctx, suite.DataDealKeeper, genesis)

	getDeal1, err := suite.DataDealKeeper.GetDeal(suite.Ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.Deals[0], *getDeal1)

	getDeal2, err := suite.DataDealKeeper.GetDeal(suite.Ctx, 2)
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.Deals[1], *getDeal2)

	getDataSale1, err := suite.DataDealKeeper.GetDataSale(suite.Ctx, suite.dataHash1, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.DataSales[0].SellerAddress, getDataSale1.SellerAddress)
	suite.Require().Equal(genesis.DataSales[0].DealId, getDataSale1.DealId)
	suite.Require().Equal(genesis.DataSales[0].VerifiableCid, getDataSale1.VerifiableCid)
	suite.Require().Equal(genesis.DataSales[0].DeliveredCid, getDataSale1.DeliveredCid)
	suite.Require().Equal(genesis.DataSales[0].Status, getDataSale1.Status)
	//suite.Require().Equal(genesis.DataSales[0].VotingPeriod.VotingStartTime.Local(), getDataSale1.VotingPeriod.VotingStartTime.Local())
	//suite.Require().Equal(genesis.DataSales[0].VotingPeriod.VotingEndTime.Local(), getDataSale1.VotingPeriod.VotingEndTime.Local())
	suite.Require().Equal(genesis.DataSales[0].VerificationTallyResult, getDataSale1.VerificationTallyResult)
	suite.Require().Equal(genesis.DataSales[0].DeliveryTallyResult, getDataSale1.DeliveryTallyResult)

	getDataSale2, err := suite.DataDealKeeper.GetDataSale(suite.Ctx, suite.dataHash2, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.DataSales[1].SellerAddress, getDataSale2.SellerAddress)
	suite.Require().Equal(genesis.DataSales[1].DealId, getDataSale2.DealId)
	suite.Require().Equal(genesis.DataSales[1].VerifiableCid, getDataSale2.VerifiableCid)
	suite.Require().Equal(genesis.DataSales[1].DeliveredCid, getDataSale2.DeliveredCid)
	suite.Require().Equal(genesis.DataSales[1].Status, getDataSale2.Status)
	//suite.Require().Equal(genesis.DataSales[1].VotingPeriod.VotingStartTime, getDataSale2.VotingPeriod.VotingStartTime.Local())
	//suite.Require().Equal(genesis.DataSales[1].VotingPeriod.VotingEndTime, getDataSale2.VotingPeriod.VotingEndTime.Local())
	suite.Require().Equal(genesis.DataSales[1].VerificationTallyResult, getDataSale2.VerificationTallyResult)
	suite.Require().Equal(genesis.DataSales[1].DeliveryTallyResult, getDataSale2.DeliveryTallyResult)

	getDataVerificationVote1, err := suite.DataDealKeeper.GetDataVerificationVote(suite.Ctx, suite.dataHash1, suite.oracleAccAddr.String(), 1)
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.DataVerificationVotes[0], *getDataVerificationVote1)

	getDataVerificationVote2, err := suite.DataDealKeeper.GetDataVerificationVote(suite.Ctx, suite.dataHash2, suite.oracleAccAddr.String(), 1)
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.DataVerificationVotes[1], *getDataVerificationVote2)
	suite.Require().Equal(genesis.Deals[1], *getDeal2)

	getDataDeliveryVote1, err := suite.DataDealKeeper.GetDataDeliveryVote(suite.Ctx, suite.dataHash1, suite.oracleAccAddr.String(), 1)
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.DataDeliveryVotes[0], *getDataDeliveryVote1)

	getDataDeliveryVote2, err := suite.DataDealKeeper.GetDataDeliveryVote(suite.Ctx, suite.dataHash2, suite.oracleAccAddr.String(), 2)
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.DataDeliveryVotes[1], *getDataDeliveryVote2)

	dataVerificationQueueElements, err := suite.DataDealKeeper.GetAllDataVerificationQueueElements(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.DataVerificationQueueElements[0].DataHash, dataVerificationQueueElements[0].DataHash)
	suite.Require().Equal(genesis.DataVerificationQueueElements[0].DealId, dataVerificationQueueElements[0].DealId)
	suite.Require().Equal(genesis.DataVerificationQueueElements[1].DataHash, dataVerificationQueueElements[1].DataHash)
	suite.Require().Equal(genesis.DataVerificationQueueElements[1].DealId, dataVerificationQueueElements[1].DealId)

	dataDeliveryQueueElements, err := suite.DataDealKeeper.GetAllDataDeliveryQueueElements(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.DataDeliveryQueueElements[0].DataHash, dataDeliveryQueueElements[0].DataHash)
	suite.Require().Equal(genesis.DataDeliveryQueueElements[0].DealId, dataDeliveryQueueElements[0].DealId)
	suite.Require().Equal(genesis.DataDeliveryQueueElements[1].DataHash, dataDeliveryQueueElements[1].DataHash)
	suite.Require().Equal(genesis.DataDeliveryQueueElements[1].DealId, dataDeliveryQueueElements[1].DealId)
}

func (suite *genesisTestSuite) TestExportGenesis() {
	deal1 := suite.MakeTestDeal(1, suite.buyerAccAddr, 100)
	deal2 := suite.MakeTestDeal(2, suite.buyerAccAddr, 100)

	dataSale1 := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash1, suite.verifiableCID1)
	dataSale2 := suite.MakeNewDataSale(suite.sellerAccAddr, suite.dataHash2, suite.verifiableCID2)

	dataVerificationVote1 := suite.MakeNewDataVerificationVote(suite.oracleAccAddr, suite.dataHash1)
	dataVerificationVote2 := suite.MakeNewDataVerificationVote(suite.oracleAccAddr, suite.dataHash2)

	dataSaleDeliveryVoting1 := suite.MakeNewDataSaleDeliveryVoting(suite.sellerAccAddr, suite.dataHash3, suite.verifiableCID1)
	dataSaleDeliveryVoting2 := suite.MakeNewDataSaleDeliveryVoting(suite.sellerAccAddr, suite.dataHash4, suite.verifiableCID2)

	dataDeliveryVote1 := suite.MakeNewDataDeliveryVote(suite.oracleAccAddr, suite.dataHash3, suite.deliveryCID1, 1)
	dataDeliveryVote2 := suite.MakeNewDataDeliveryVote(suite.oracleAccAddr, suite.dataHash4, suite.deliveryCID2, 1)

	voteBz, err := suite.Cdc.Marshaler.Marshal(dataVerificationVote2)
	suite.Require().NoError(err)

	oraclePrivKeySecp256k1 := secp256k1.PrivKey{Key: suite.oraclePrivKey.Serialize()}
	signature, err := oraclePrivKeySecp256k1.Sign(voteBz)
	suite.Require().NoError(err)

	voteBz2, err := suite.Cdc.Marshaler.Marshal(dataDeliveryVote2)
	suite.Require().NoError(err)
	signature2, err := oraclePrivKeySecp256k1.Sign(voteBz2)
	suite.Require().NoError(err)

	verificationQueueElement1 := &types.DataVerificationQueueElement{
		DataHash:      dataSale1.DataHash,
		DealId:        dataSale1.DealId,
		VotingEndTime: dataSale1.VerificationVotingPeriod.VotingEndTime,
	}

	verificationQueueElement2 := &types.DataVerificationQueueElement{
		DataHash:      dataSale2.DataHash,
		DealId:        dataSale2.DealId,
		VotingEndTime: dataSale2.VerificationVotingPeriod.VotingEndTime,
	}

	deliveryQueueElement1 := &types.DataDeliveryQueueElement{
		DataHash:      dataSaleDeliveryVoting1.DataHash,
		DealId:        dataSaleDeliveryVoting1.DealId,
		VotingEndTime: dataSaleDeliveryVoting1.DeliveryVotingPeriod.VotingEndTime,
	}

	deliveryQueueElement2 := &types.DataDeliveryQueueElement{
		DataHash:      dataSaleDeliveryVoting2.DataHash,
		DealId:        dataSaleDeliveryVoting2.DealId,
		VotingEndTime: dataSaleDeliveryVoting2.DeliveryVotingPeriod.VotingEndTime,
	}

	genesis := types.GenesisState{
		Deals:                         []types.Deal{*deal1},
		NextDealNumber:                2,
		DataSales:                     []types.DataSale{*dataSale1},
		DataVerificationVotes:         []types.DataVerificationVote{*dataVerificationVote1},
		DataDeliveryVotes:             []types.DataDeliveryVote{*dataDeliveryVote1},
		DataVerificationQueueElements: []types.DataVerificationQueueElement{*verificationQueueElement1},
		DataDeliveryQueueElements:     []types.DataDeliveryQueueElement{*deliveryQueueElement1},
	}

	msgCreateDeal := &types.MsgCreateDeal{
		DataSchema:   deal2.DataSchema,
		Budget:       deal2.Budget,
		MaxNumData:   deal2.MaxNumData,
		BuyerAddress: deal2.BuyerAddress,
	}

	msgSellData := &types.MsgSellData{
		DealId:        1,
		VerifiableCid: dataSale2.VerifiableCid,
		DataHash:      dataSale2.DataHash,
		SellerAddress: dataSale2.SellerAddress,
	}

	datadeal.InitGenesis(suite.Ctx, suite.DataDealKeeper, genesis)

	err = suite.FundAccount(suite.Ctx, suite.buyerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	_, err = suite.DataDealKeeper.CreateDeal(suite.Ctx, suite.buyerAccAddr, msgCreateDeal)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.SellData(suite.Ctx, msgSellData)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.VoteDataVerification(suite.Ctx, dataVerificationVote2, signature)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.SetDataSale(suite.Ctx, dataSaleDeliveryVoting2)
	suite.Require().NoError(err)

	suite.DataDealKeeper.AddDataDeliveryQueue(suite.Ctx, dataSaleDeliveryVoting2.DataHash, dataSaleDeliveryVoting2.DealId, dataSaleDeliveryVoting2.DeliveryVotingPeriod.VotingEndTime)

	err = suite.DataDealKeeper.VoteDataDelivery(suite.Ctx, dataDeliveryVote2, signature2)
	suite.Require().NoError(err)

	genesisStatus := datadeal.ExportGenesis(suite.Ctx, suite.DataDealKeeper)

	suite.Require().Equal(deal1.Id, genesisStatus.Deals[0].Id)
	suite.Require().Equal(deal2.Id, genesisStatus.Deals[1].Id)
	suite.Require().Equal(deal1.Address, genesisStatus.Deals[0].Address)
	suite.Require().Equal(deal2.Address, genesisStatus.Deals[1].Address)
	suite.Require().Equal(deal1.BuyerAddress, genesisStatus.Deals[0].BuyerAddress)
	suite.Require().Equal(deal2.BuyerAddress, genesisStatus.Deals[1].BuyerAddress)
	suite.Require().Equal(deal1.DataSchema, genesisStatus.Deals[0].DataSchema)
	suite.Require().Equal(deal2.DataSchema, genesisStatus.Deals[1].DataSchema)
	suite.Require().Equal(deal1.Budget, genesisStatus.Deals[0].Budget)
	suite.Require().Equal(deal2.Budget, genesisStatus.Deals[1].Budget)
	suite.Require().Equal(uint64(3), genesisStatus.NextDealNumber)

	suite.Require().Equal(dataSale1.SellerAddress, genesisStatus.DataSales[0].SellerAddress)
	suite.Require().Equal(dataSale2.SellerAddress, genesisStatus.DataSales[1].SellerAddress)
	suite.Require().Equal(dataSale1.DealId, genesisStatus.DataSales[0].DealId)
	suite.Require().Equal(dataSale2.DealId, genesisStatus.DataSales[1].DealId)
	suite.Require().Equal(dataSale1.VerifiableCid, genesisStatus.DataSales[0].VerifiableCid)
	suite.Require().Equal(dataSale2.VerifiableCid, genesisStatus.DataSales[1].VerifiableCid)
	suite.Require().Equal(dataSale1.DeliveredCid, genesisStatus.DataSales[0].DeliveredCid)
	suite.Require().Equal(dataSale2.DeliveredCid, genesisStatus.DataSales[1].DeliveredCid)
	suite.Require().Equal(dataSale1.Status, genesisStatus.DataSales[0].Status)
	suite.Require().Equal(dataSale2.Status, genesisStatus.DataSales[1].Status)
	suite.Require().Equal(dataSale1.VerificationTallyResult, genesisStatus.DataSales[0].VerificationTallyResult)
	suite.Require().Equal(dataSale2.VerificationTallyResult, genesisStatus.DataSales[1].VerificationTallyResult)
	suite.Require().Equal(dataSale1.DeliveryTallyResult, genesisStatus.DataSales[0].DeliveryTallyResult)
	suite.Require().Equal(dataSale2.DeliveryTallyResult, genesisStatus.DataSales[1].DeliveryTallyResult)

	suite.Require().Equal(dataVerificationVote1.VoterAddress, genesisStatus.DataVerificationVotes[0].VoterAddress)
	suite.Require().Equal(dataVerificationVote2.VoterAddress, genesisStatus.DataVerificationVotes[1].VoterAddress)
	suite.Require().Equal(dataVerificationVote1.DataHash, genesisStatus.DataVerificationVotes[0].DataHash)
	suite.Require().Equal(dataVerificationVote2.DataHash, genesisStatus.DataVerificationVotes[1].DataHash)
	suite.Require().Equal(dataVerificationVote1.DealId, genesisStatus.DataVerificationVotes[0].DealId)
	suite.Require().Equal(dataVerificationVote2.DealId, genesisStatus.DataVerificationVotes[1].DealId)
	suite.Require().Equal(dataVerificationVote1.VoteOption, genesisStatus.DataVerificationVotes[0].VoteOption)
	suite.Require().Equal(dataVerificationVote2.VoteOption, genesisStatus.DataVerificationVotes[1].VoteOption)

	suite.Require().Equal(dataDeliveryVote1.VoterAddress, genesisStatus.DataDeliveryVotes[0].VoterAddress)
	suite.Require().Equal(dataDeliveryVote2.VoterAddress, genesisStatus.DataDeliveryVotes[1].VoterAddress)
	suite.Require().Equal(dataDeliveryVote1.DataHash, genesisStatus.DataDeliveryVotes[0].DataHash)
	suite.Require().Equal(dataDeliveryVote2.DataHash, genesisStatus.DataDeliveryVotes[1].DataHash)
	suite.Require().Equal(dataDeliveryVote1.DealId, genesisStatus.DataDeliveryVotes[0].DealId)
	suite.Require().Equal(dataDeliveryVote2.DealId, genesisStatus.DataDeliveryVotes[1].DealId)
	suite.Require().Equal(dataDeliveryVote1.VoteOption, genesisStatus.DataDeliveryVotes[0].VoteOption)
	suite.Require().Equal(dataDeliveryVote2.VoteOption, genesisStatus.DataDeliveryVotes[1].VoteOption)

	suite.Require().Equal(verificationQueueElement2.DataHash, genesisStatus.DataVerificationQueueElements[0].DataHash)
	suite.Require().Equal(verificationQueueElement2.DealId, genesisStatus.DataVerificationQueueElements[0].DealId)
	suite.Require().Equal(verificationQueueElement1.DataHash, genesisStatus.DataVerificationQueueElements[1].DataHash)
	suite.Require().Equal(verificationQueueElement1.DealId, genesisStatus.DataVerificationQueueElements[1].DealId)

	suite.Require().Equal(deliveryQueueElement1.DataHash, genesisStatus.DataDeliveryQueueElements[0].DataHash)
	suite.Require().Equal(deliveryQueueElement1.DealId, genesisStatus.DataDeliveryQueueElements[0].DealId)
	suite.Require().Equal(deliveryQueueElement2.DataHash, genesisStatus.DataDeliveryQueueElements[1].DataHash)
	suite.Require().Equal(deliveryQueueElement2.DealId, genesisStatus.DataDeliveryQueueElements[1].DealId)
}
