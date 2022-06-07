package datadeal_test

import (
	"testing"

	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/datadeal"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
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

func (suite *genesisTestSuite) TestDataDealInitGenesis() {
	newDeal := makeTestDeal()
	newDataCert := makeTestDataCert()

	dataCertKey := types.GetKeyPrefixDataCert(newDataCert.UnsignedCert.DealId, newDataCert.UnsignedCert.DataHash)
	stringDataCertKey := string(dataCertKey)

	datadeal.InitGenesis(suite.Ctx, suite.DataDealKeeper, types.GenesisState{
		Deals: map[uint64]types.Deal{
			newDeal.GetDealId(): newDeal,
		},
		DataCerts: map[string]types.DataCert{
			stringDataCertKey: newDataCert,
		},
		NextDealNumber: 2,
	})

	nextNum, err := suite.DataDealKeeper.GetNextDealNumberAndIncrement(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(nextNum, uint64(2))

	dealStored, err := suite.DataDealKeeper.GetDeal(suite.Ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(newDeal.GetDealId(), dealStored.GetDealId())
	suite.Require().Equal(newDeal.GetDealAddress(), dealStored.GetDealAddress())
	suite.Require().Equal(newDeal.GetDataSchema(), dealStored.GetDataSchema())
	suite.Require().Equal(newDeal.GetBudget(), dealStored.GetBudget())
	suite.Require().Equal(newDeal.GetTrustedOracles(), dealStored.GetTrustedOracles())
	suite.Require().Equal(newDeal.GetCurNumData(), dealStored.GetCurNumData())
	suite.Require().Equal(newDeal.GetMaxNumData(), dealStored.GetMaxNumData())
	suite.Require().Equal(newDeal.GetOwner(), dealStored.GetOwner())
	suite.Require().Equal(newDeal.GetStatus(), dealStored.GetStatus())

	_, err = suite.DataDealKeeper.GetDeal(suite.Ctx, 2)
	suite.Require().Error(err)

	dataCertStored, err := suite.DataDealKeeper.GetDataCert(suite.Ctx, 1, string(newDataCert.UnsignedCert.DataHash))
	suite.Require().NoError(err)
	suite.Require().Equal(newDataCert.GetSignature(), dataCertStored.GetSignature())
	suite.Require().Equal(newDataCert.UnsignedCert.GetDealId(), dataCertStored.UnsignedCert.GetDealId())
	suite.Require().Equal(newDataCert.UnsignedCert.GetDataHash(), dataCertStored.UnsignedCert.GetDataHash())
	suite.Require().Equal(newDataCert.UnsignedCert.GetEncryptedDataUrl(), dataCertStored.UnsignedCert.GetEncryptedDataUrl())
	suite.Require().Equal(newDataCert.UnsignedCert.GetOracleAddress(), dataCertStored.UnsignedCert.GetOracleAddress())
	suite.Require().Equal(newDataCert.UnsignedCert.GetRequesterAddress(), dataCertStored.UnsignedCert.GetRequesterAddress())
}

func (suite *genesisTestSuite) TestDataDealExportGenesis() {
	newDeal := makeTestDeal()
	newDataCert := makeTestDataCert()

	dataCertKey := types.GetKeyPrefixDataCert(newDataCert.UnsignedCert.DealId, newDataCert.UnsignedCert.DataHash)
	stringDataCertKey := string(dataCertKey)

	datadeal.InitGenesis(suite.Ctx, suite.DataDealKeeper, types.GenesisState{
		Deals: map[uint64]types.Deal{
			newDeal.GetDealId(): newDeal,
		},
		DataCerts: map[string]types.DataCert{
			stringDataCertKey: newDataCert,
		},
		NextDealNumber: 2,
	})

	err := suite.BankKeeper.AddCoins(suite.Ctx, acc1, defaultFunds)
	suite.Require().NoError(err)

	tempDeal := types.Deal{
		DataSchema:     []string{"http://jsonld.com"},
		Budget:         &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(10000000)},
		MaxNumData:     10000,
		TrustedOracles: []string{acc1.String()},
		Owner:          acc1.String(),
	}

	oracle := oracletypes.Oracle{
		Address:  acc1.String(),
		Endpoint: "https://my-oracle.org",
	}

	err = suite.OracleKeeper.SetOracle(suite.Ctx, oracle)
	suite.Require().NoError(err)

	_, err = suite.DataDealKeeper.CreateDeal(suite.Ctx, acc1, tempDeal)
	suite.Require().NoError(err)

	newDataCert2 := makeTestDataCert2()
	_, err = suite.DataDealKeeper.SellData(suite.Ctx, acc2, newDataCert2)
	suite.Require().NoError(err)

	genesis := datadeal.ExportGenesis(suite.Ctx, suite.DataDealKeeper)
	suite.Require().Equal(genesis.NextDealNumber, uint64(3))
	suite.Require().Len(genesis.Deals, 2)
	suite.Require().Len(genesis.DataCerts, 2)
}

func makeTestDeal() types.Deal {
	return types.Deal{
		DealId:         1,
		DealAddress:    types.NewDealAddress(1).String(),
		DataSchema:     nil,
		Budget:         &sdk.Coin{Denom: "umed", Amount: sdk.NewInt(10000000)},
		TrustedOracles: nil,
		MaxNumData:     10000,
		CurNumData:     0,
		Owner:          acc1.String(),
		Status:         "ACTIVE",
	}
}

func makeTestDataCert() types.DataCert {
	uCert := types.UnsignedDataCert{
		DealId:           2,
		DataHash:         []byte("1a312c123x23"),
		EncryptedDataUrl: []byte("https://panacea.org/a/123.json"),
		OracleAddress:    acc1.String(),
		RequesterAddress: acc2.String(),
	}

	marshal, err := uCert.Marshal()
	if err != nil {
		return types.DataCert{}
	}

	sign, err := privKey.Sign(marshal)
	if err != nil {
		return types.DataCert{}
	}

	return types.DataCert{
		UnsignedCert: &uCert,
		Signature:    sign,
	}
}

func makeTestDataCert2() types.DataCert {
	uCert := types.UnsignedDataCert{
		DealId:           2,
		DataHash:         []byte("1a312c1223x2fs3"),
		EncryptedDataUrl: []byte("https://panacea.org/a/123.json"),
		OracleAddress:    acc1.String(),
		RequesterAddress: acc2.String(),
	}

	marshal, err := uCert.Marshal()
	if err != nil {
		return types.DataCert{}
	}

	sign, err := privKey.Sign(marshal)
	if err != nil {
		return types.DataCert{}
	}

	return types.DataCert{
		UnsignedCert: &uCert,
		Signature:    sign,
	}
}
