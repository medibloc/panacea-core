package datadeal_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/datadeal"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"testing"
)

var (
	acc1                   = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	privKey                = secp256k1.GenPrivKey()
	acc2                   = sdk.AccAddress(privKey.PubKey().Address())
	defaultFunds sdk.Coins = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)))
)

type genesisTestSuite struct {
	testsuite.TestSuite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}

func (suite *genesisTestSuite) TestDatadealInitGenesis() {
	newDeal := makeTestDeal()
	newDataCert := makeTestDataCert()

	dataCertificateKey := types.GetKeyPrefixDataCertificate(newDataCert.UnsignedCert.DealId, newDataCert.UnsignedCert.DataHash)
	stringDataCertificateKey := string(dataCertificateKey)

	datadeal.InitGenesis(suite.Ctx, suite.DatadealKeeper, types.GenesisState{
		Deals: map[uint64]*types.Deal{
			newDeal.GetDealId(): &newDeal,
		},
		DataCertificates: map[string]*types.DataValidationCertificate{
			stringDataCertificateKey: &newDataCert,
		},
		NextDealNumber: 2,
	})

	suite.Require().Equal(suite.DatadealKeeper.GetNextDealNumberAndIncrement(suite.Ctx), uint64(2))

	dealStored, err := suite.DatadealKeeper.GetDeal(suite.Ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(newDeal.GetDealId(), dealStored.GetDealId())
	suite.Require().Equal(newDeal.GetDealAddress(), dealStored.GetDealAddress())
	suite.Require().Equal(newDeal.GetDataSchema(), dealStored.GetDataSchema())
	suite.Require().Equal(newDeal.GetBudget(), dealStored.GetBudget())
	suite.Require().Equal(newDeal.GetTrustedDataValidators(), dealStored.GetTrustedDataValidators())
	suite.Require().Equal(newDeal.GetCurNumData(), dealStored.GetCurNumData())
	suite.Require().Equal(newDeal.GetMaxNumData(), dealStored.GetMaxNumData())
	suite.Require().Equal(newDeal.GetOwner(), dealStored.GetOwner())
	suite.Require().Equal(newDeal.GetStatus(), dealStored.GetStatus())

	_, err = suite.DatadealKeeper.GetDeal(suite.Ctx, 2)
	suite.Require().Error(err)

	dataCertificateStored, err := suite.DatadealKeeper.GetDataCertificate(suite.Ctx, newDataCert)
	suite.Require().NoError(err)
	suite.Require().Equal(newDataCert.GetSignature(), dataCertificateStored.GetSignature())
	suite.Require().Equal(newDataCert.UnsignedCert.GetDealId(), dataCertificateStored.UnsignedCert.GetDealId())
	suite.Require().Equal(newDataCert.UnsignedCert.GetDataHash(), dataCertificateStored.UnsignedCert.GetDataHash())
	suite.Require().Equal(newDataCert.UnsignedCert.GetEncryptedDataUrl(), dataCertificateStored.UnsignedCert.GetEncryptedDataUrl())
	suite.Require().Equal(newDataCert.UnsignedCert.GetDataValidatorAddress(), dataCertificateStored.UnsignedCert.GetDataValidatorAddress())
	suite.Require().Equal(newDataCert.UnsignedCert.GetRequesterAddress(), dataCertificateStored.UnsignedCert.GetRequesterAddress())
}

func (suite *genesisTestSuite) TestDatadealExportGenesis() {
	newDeal := makeTestDeal()
	newDataCert := makeTestDataCert()

	dataCertificateKey := types.GetKeyPrefixDataCertificate(newDataCert.UnsignedCert.DealId, newDataCert.UnsignedCert.DataHash)
	stringDataCertificateKey := string(dataCertificateKey)

	datadeal.InitGenesis(suite.Ctx, suite.DatadealKeeper, types.GenesisState{
		Deals: map[uint64]*types.Deal{
			newDeal.GetDealId(): &newDeal,
		},
		DataCertificates: map[string]*types.DataValidationCertificate{
			stringDataCertificateKey: &newDataCert,
		},
		NextDealNumber: 2,
	})

	err := suite.BankKeeper.AddCoins(suite.Ctx, acc1, defaultFunds)
	suite.Require().NoError(err)

	tempDeal := types.Deal{
		DataSchema:            []string{"http://jsonld.com"},
		Budget:                &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(10000000)},
		MaxNumData:            10000,
		TrustedDataValidators: []string{acc1.String()},
		Owner:                 acc1.String(),
	}

	_, err = suite.DatadealKeeper.CreateNewDeal(suite.Ctx, acc1, tempDeal)
	suite.Require().NoError(err)

	newDataCert2 := makeTestDataCert2()
	_, err = suite.DatadealKeeper.SellOwnData(suite.Ctx, acc2, newDataCert2)
	suite.Require().NoError(err)

	genesis := datadeal.ExportGenesis(suite.Ctx, suite.DatadealKeeper)
	suite.Require().Equal(genesis.NextDealNumber, uint64(3))
	suite.Require().Len(genesis.Deals, 2)
	suite.Require().Len(genesis.DataCertificates, 2)
}

func makeTestDeal() types.Deal {
	return types.Deal{
		DealId:                1,
		DealAddress:           types.NewDealAddress(1).String(),
		DataSchema:            nil,
		Budget:                &sdk.Coin{Denom: "umed", Amount: sdk.NewInt(10000000)},
		TrustedDataValidators: nil,
		MaxNumData:            10000,
		CurNumData:            0,
		Owner:                 acc1.String(),
		Status:                "ACTIVE",
	}
}

func makeTestDataCert() types.DataValidationCertificate {
	uCert := types.UnsignedDataValidationCertificate{
		DealId:               2,
		DataHash:             []byte("1a312c123x23"),
		EncryptedDataUrl:     []byte("https://panacea.org/a/123.json"),
		DataValidatorAddress: acc1.String(),
		RequesterAddress:     acc2.String(),
	}

	marshal, err := uCert.Marshal()
	if err != nil {
		return types.DataValidationCertificate{}
	}

	sign, err := privKey.Sign(marshal)
	if err != nil {
		return types.DataValidationCertificate{}
	}

	return types.DataValidationCertificate{
		UnsignedCert: &uCert,
		Signature:    sign,
	}
}

func makeTestDataCert2() types.DataValidationCertificate {
	uCert := types.UnsignedDataValidationCertificate{
		DealId:               2,
		DataHash:             []byte("1a312c1223x2fs3"),
		EncryptedDataUrl:     []byte("https://panacea.org/a/123.json"),
		DataValidatorAddress: acc1.String(),
		RequesterAddress:     acc2.String(),
	}

	marshal, err := uCert.Marshal()
	if err != nil {
		return types.DataValidationCertificate{}
	}

	sign, err := privKey.Sign(marshal)
	if err != nil {
		return types.DataValidationCertificate{}
	}

	return types.DataValidationCertificate{
		UnsignedCert: &uCert,
		Signature:    sign,
	}
}
