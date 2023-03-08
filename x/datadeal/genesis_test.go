package datadeal_test

import (
	"encoding/base64"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
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

	consumerAccAddr sdk.AccAddress
	providerAccAddr sdk.AccAddress

	oracleAccPrivKey cryptotypes.PrivKey
	oracleAccPubKey  cryptotypes.PubKey

	oraclePrivKey *btcec.PrivateKey
	oraclePubKey  *btcec.PublicKey
	oracleAccAddr sdk.AccAddress

	defaultFunds sdk.Coins
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}

func (suite *genesisTestSuite) BeforeTest(_, _ string) {
	suite.consumerAccAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	suite.providerAccAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	suite.oracleAccPrivKey = secp256k1.GenPrivKey()
	suite.oracleAccPubKey = suite.oracleAccPrivKey.PubKey()
	suite.oracleAccAddr = sdk.AccAddress(suite.oracleAccPubKey.Address())

	suite.oraclePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.oraclePubKey = suite.oraclePrivKey.PubKey()

	suite.OracleKeeper.SetParams(suite.Ctx, oracletypes.Params{
		OraclePublicKey:          base64.StdEncoding.EncodeToString(suite.oraclePubKey.SerializeCompressed()),
		OraclePubKeyRemoteReport: "",
		UniqueId:                 "uniqueID",
	})

	oracle := &oracletypes.Oracle{
		OracleAddress:        suite.oracleAccAddr.String(),
		UniqueId:             "uniqueID",
		Endpoint:             "https://my-validator.org",
		OracleCommissionRate: sdk.NewDecWithPrec(1, 1),
	}
	err := suite.OracleKeeper.SetOracle(suite.Ctx, oracle)
	suite.Require().NoError(err)

	suite.defaultFunds = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)))
}

func (suite *genesisTestSuite) TestInitGenesis() {
	deal1 := suite.MakeTestDeal(1, suite.consumerAccAddr, 100)
	deal2 := suite.MakeTestDeal(2, suite.consumerAccAddr, 100)

	unsignedCert1 := &types.UnsignedCertificate{
		UniqueId:        "uniqueID",
		OracleAddress:   suite.oracleAccAddr.String(),
		DealId:          1,
		ProviderAddress: suite.providerAccAddr.String(),
		DataHash:        "dataHash",
	}

	unsignedCertBz1, err := unsignedCert1.Marshal()
	suite.Require().NoError(err)

	sign1, err := suite.oraclePrivKey.Sign(unsignedCertBz1)

	suite.Require().NoError(err)

	consent1 := &types.Consent{
		DealId: deal1.Id,
		Certificate: &types.Certificate{
			UnsignedCertificate: unsignedCert1,
			Signature:           sign1.Serialize(),
		},
		Agreements: suite.MakeTestAgreements(deal1),
	}

	unsignedCert2 := &types.UnsignedCertificate{
		UniqueId:        "uniqueID",
		OracleAddress:   suite.oracleAccAddr.String(),
		DealId:          2,
		ProviderAddress: suite.providerAccAddr.String(),
		DataHash:        "dataHash",
	}

	unsignedCertBz2, err := unsignedCert1.Marshal()
	suite.Require().NoError(err)

	sign2, err := suite.oraclePrivKey.Sign(unsignedCertBz2)

	suite.Require().NoError(err)

	consent2 := &types.Consent{
		DealId: deal2.Id,
		Certificate: &types.Certificate{
			UnsignedCertificate: unsignedCert2,
			Signature:           sign2.Serialize(),
		},
		Agreements: suite.MakeTestAgreements(deal2),
	}

	genesis := types.GenesisState{
		Deals:          []types.Deal{*deal1, *deal2},
		NextDealNumber: 1,
		Consents:       []types.Consent{*consent1, *consent2},
	}

	datadeal.InitGenesis(suite.Ctx, suite.DataDealKeeper, genesis)

	getDeal1, err := suite.DataDealKeeper.GetDeal(suite.Ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.Deals[0], *getDeal1)

	getDeal2, err := suite.DataDealKeeper.GetDeal(suite.Ctx, 2)
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.Deals[1], *getDeal2)

	getConsent1, err := suite.DataDealKeeper.GetConsent(suite.Ctx, 1, "dataHash")
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.Consents[0], *getConsent1)

	getConsent2, err := suite.DataDealKeeper.GetConsent(suite.Ctx, 2, "dataHash")
	suite.Require().NoError(err)
	suite.Require().Equal(genesis.Consents[1], *getConsent2)
}

func (suite *genesisTestSuite) TestExportGenesis() {
	deal1 := suite.MakeTestDeal(1, suite.consumerAccAddr, 100)
	deal2 := suite.MakeTestDeal(2, suite.consumerAccAddr, 100)

	unsignedCert1 := &types.UnsignedCertificate{
		UniqueId:        "uniqueID",
		OracleAddress:   suite.oracleAccAddr.String(),
		DealId:          1,
		ProviderAddress: suite.providerAccAddr.String(),
		DataHash:        "dataHash",
	}

	unsignedCertBz1, err := unsignedCert1.Marshal()
	suite.Require().NoError(err)

	sign1, err := suite.oraclePrivKey.Sign(unsignedCertBz1)

	suite.Require().NoError(err)

	consent1 := &types.Consent{
		DealId: deal1.Id,
		Certificate: &types.Certificate{
			UnsignedCertificate: unsignedCert1,
			Signature:           sign1.Serialize(),
		},
		Agreements: suite.MakeTestAgreements(deal1),
	}

	genesis := types.GenesisState{
		Deals:          []types.Deal{*deal1},
		NextDealNumber: 2,
		Consents:       []types.Consent{*consent1},
	}

	msgCreateDeal := &types.MsgCreateDeal{
		DataSchema:      deal2.DataSchema,
		Budget:          deal2.Budget,
		MaxNumData:      deal2.MaxNumData,
		ConsumerAddress: deal2.ConsumerAddress,
		AgreementTerms:  deal2.AgreementTerms,
	}

	unsignedCert2 := &types.UnsignedCertificate{
		UniqueId:        "uniqueID",
		OracleAddress:   suite.oracleAccAddr.String(),
		DealId:          2,
		ProviderAddress: suite.providerAccAddr.String(),
		DataHash:        "dataHash",
	}

	unsignedCertBz2, err := unsignedCert1.Marshal()
	suite.Require().NoError(err)

	sign2, err := suite.oraclePrivKey.Sign(unsignedCertBz2)

	suite.Require().NoError(err)

	consent2 := &types.Consent{
		DealId: deal2.Id,
		Certificate: &types.Certificate{
			UnsignedCertificate: unsignedCert2,
			Signature:           sign2.Serialize(),
		},
		Agreements: suite.MakeTestAgreements(deal2),
	}

	msgSubmitConsent := &types.MsgSubmitConsent{
		Consent: consent2,
	}

	datadeal.InitGenesis(suite.Ctx, suite.DataDealKeeper, genesis)

	err = suite.FundAccount(suite.Ctx, suite.consumerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	_, err = suite.DataDealKeeper.CreateDeal(suite.Ctx, msgCreateDeal)
	suite.Require().NoError(err)

	err = suite.DataDealKeeper.SubmitConsent(suite.Ctx, msgSubmitConsent.Consent)
	suite.Require().NoError(err)

	genesisStatus := datadeal.ExportGenesis(suite.Ctx, suite.DataDealKeeper)

	suite.Require().Equal(deal1.Id, genesisStatus.Deals[0].Id)
	suite.Require().Equal(deal2.Id, genesisStatus.Deals[1].Id)
	suite.Require().Equal(deal1.Address, genesisStatus.Deals[0].Address)
	suite.Require().Equal(deal2.Address, genesisStatus.Deals[1].Address)
	suite.Require().Equal(deal1.ConsumerAddress, genesisStatus.Deals[0].ConsumerAddress)
	suite.Require().Equal(deal2.ConsumerAddress, genesisStatus.Deals[1].ConsumerAddress)
	suite.Require().Equal(deal1.DataSchema, genesisStatus.Deals[0].DataSchema)
	suite.Require().Equal(deal2.DataSchema, genesisStatus.Deals[1].DataSchema)
	suite.Require().Equal(deal1.Budget, genesisStatus.Deals[0].Budget)
	suite.Require().Equal(deal2.Budget, genesisStatus.Deals[1].Budget)
	suite.Require().Equal(uint64(3), genesisStatus.NextDealNumber)

	suite.Require().Equal(*consent1, genesisStatus.Consents[0])
	suite.Require().Equal(*consent2, genesisStatus.Consents[1])
}
