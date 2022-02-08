package market_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/market"
	"github.com/medibloc/panacea-core/v2/x/market/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"testing"
)

var acc1 = secp256k1.GenPrivKey().PubKey().Address()
var privKey = secp256k1.GenPrivKey()
var acc2 = privKey.PubKey().Address()

type genesisTestSuite struct {
	testsuite.TestSuite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}

func (suite *genesisTestSuite) TestMarketInitGenesis() {
	newDeal := makeTestDeal()
	newDataCert := makeTestDataCert()

	dataCertificateKey := types.GetKeyPrefixDataCertificate(newDataCert.UnsignedCert.DealId, newDataCert.UnsignedCert.DataHash)
	stringDataCertificateKey := string(dataCertificateKey[:])

	market.InitGenesis(suite.Ctx, suite.MarketKeeper, types.GenesisState{
		Deals: map[uint64]*types.Deal{
			newDeal.GetDealId(): &newDeal,
		},
		DataCertificate: map[string]*types.DataValidationCertificate{
			stringDataCertificateKey: &newDataCert,
		},
		NextDealNumber: 2,
	})

	suite.Require().Equal(suite.MarketKeeper.GetNextDealNumberAndIncrement(suite.Ctx), uint64(2))

	dealStored, err := suite.MarketKeeper.GetDeal(suite.Ctx, 1)
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

	_, err = suite.MarketKeeper.GetDeal(suite.Ctx, 2)
	suite.Require().Error(err)

	dataCertificateStored, err := suite.MarketKeeper.GetDataCertificate(suite.Ctx, newDataCert)
	suite.Require().NoError(err)
	suite.Require().Equal(newDataCert.GetSignature(), dataCertificateStored.GetSignature())
	suite.Require().Equal(newDataCert.UnsignedCert.GetDealId(), dataCertificateStored.UnsignedCert.GetDealId())
	suite.Require().Equal(newDataCert.UnsignedCert.GetDataHash(), dataCertificateStored.UnsignedCert.GetDataHash())
	suite.Require().Equal(newDataCert.UnsignedCert.GetEncryptedDataUrl(), dataCertificateStored.UnsignedCert.GetEncryptedDataUrl())
	suite.Require().Equal(newDataCert.UnsignedCert.GetDataValidatorAddress(), dataCertificateStored.UnsignedCert.GetDataValidatorAddress())
	suite.Require().Equal(newDataCert.UnsignedCert.GetRequesterAddress(), dataCertificateStored.UnsignedCert.GetRequesterAddress())
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
