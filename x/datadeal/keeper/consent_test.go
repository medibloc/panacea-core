package keeper_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
)

type consentTestSuite struct {
	testsuite.TestSuite

	uniqueID string

	defaultFunds    sdk.Coins
	consumerAccAddr sdk.AccAddress

	oracleAccPrivKey cryptotypes.PrivKey
	oracleAccPubKey  cryptotypes.PubKey
	oracleAccAddr    sdk.AccAddress

	providerAccPrivKey cryptotypes.PrivKey
	providerAccPubKey  cryptotypes.PubKey
	providerAccAddr    sdk.AccAddress

	oraclePrivKey        *btcec.PrivateKey
	oraclePubKey         *btcec.PublicKey
	invalidOraclePrivKey *btcec.PrivateKey
	invalidOraclePubKey  *btcec.PublicKey

	dataHash string
}

func TestCertificateTestSuite(t *testing.T) {
	suite.Run(t, new(consentTestSuite))
}

func (suite *consentTestSuite) BeforeTest(_, _ string) {
	suite.consumerAccAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	suite.defaultFunds = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)))

	suite.uniqueID = "uniqueID"
	suite.oracleAccPrivKey = secp256k1.GenPrivKey()
	suite.oracleAccPubKey = suite.oracleAccPrivKey.PubKey()
	suite.oracleAccAddr = sdk.AccAddress(suite.oracleAccPubKey.Address())

	suite.providerAccPrivKey = secp256k1.GenPrivKey()
	suite.providerAccPubKey = suite.providerAccPrivKey.PubKey()
	suite.providerAccAddr = sdk.AccAddress(suite.providerAccPubKey.Address())

	suite.oraclePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.oraclePubKey = suite.oraclePrivKey.PubKey()
	suite.invalidOraclePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.invalidOraclePubKey = suite.invalidOraclePrivKey.PubKey()

	suite.dataHash = "dataHash"

	suite.OracleKeeper.SetParams(suite.Ctx, oracletypes.Params{
		OraclePublicKey:          base64.StdEncoding.EncodeToString(suite.oraclePubKey.SerializeCompressed()),
		OraclePubKeyRemoteReport: "",
		UniqueId:                 suite.uniqueID,
	})

	err := suite.DataDealKeeper.SetNextDealNumber(suite.Ctx, 1)
	suite.Require().NoError(err)
}

func (suite *consentTestSuite) createSampleDeal(budgetAmount, maxNumData uint64) uint64 {
	err := suite.FundAccount(suite.Ctx, suite.consumerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	budget := &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewIntFromUint64(budgetAmount)}

	msgCreateDeal := &types.MsgCreateDeal{
		DataSchema:      []string{"http://jsonld.com"},
		Budget:          budget,
		MaxNumData:      maxNumData,
		ConsumerAddress: suite.consumerAccAddr.String(),
	}

	dealID, err := suite.DataDealKeeper.CreateDeal(suite.Ctx, msgCreateDeal)
	suite.Require().NoError(err)

	return dealID
}

func (suite *consentTestSuite) storeSampleOracle(address, uniqueID string, commissionRate sdk.Dec) *oracletypes.Oracle {
	oracle := &oracletypes.Oracle{
		OracleAddress:        address,
		UniqueId:             uniqueID,
		Endpoint:             "https://my-validator.org",
		OracleCommissionRate: commissionRate,
	}
	err := suite.OracleKeeper.SetOracle(suite.Ctx, oracle)
	suite.Require().NoError(err)

	return oracle
}

func (suite *consentTestSuite) TestSubmitConsentSuccess() {
	budgetAmount := uint64(10000)
	dealID := suite.createSampleDeal(budgetAmount, 10)
	deal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, dealID)
	suite.Require().NoError(err)

	oracleCommissionRate := sdk.NewDecWithPrec(1, 1) // 10%
	suite.storeSampleOracle(suite.oracleAccAddr.String(), suite.uniqueID, oracleCommissionRate)

	unsignedCert := &types.UnsignedCertificate{
		Cid:             "cid",
		UniqueId:        suite.uniqueID,
		OracleAddress:   suite.oracleAccAddr.String(),
		DealId:          dealID,
		ProviderAddress: suite.providerAccAddr.String(),
		DataHash:        suite.dataHash,
	}

	unsignedCertBz, err := unsignedCert.Marshal()
	suite.Require().NoError(err)

	sign, err := suite.oraclePrivKey.Sign(unsignedCertBz)

	suite.Require().NoError(err)

	consent := &types.Consent{
		Certificate: &types.Certificate{
			UnsignedCertificate: unsignedCert,
			Signature:           sign.Serialize(),
		},
	}

	providerBalance := suite.BankKeeper.GetBalance(suite.Ctx, suite.providerAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.ZeroInt(), providerBalance.Amount)

	oracleBalance := suite.BankKeeper.GetBalance(suite.Ctx, suite.oracleAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.ZeroInt(), oracleBalance.Amount)

	dealAccAddr, err := sdk.AccAddressFromBech32(deal.Address)
	suite.Require().NoError(err)
	dealBalance := suite.BankKeeper.GetBalance(suite.Ctx, dealAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewIntFromUint64(budgetAmount), dealBalance.Amount)

	err = suite.DataDealKeeper.SubmitConsent(suite.Ctx, consent)
	suite.Require().NoError(err)

	providerBalance = suite.BankKeeper.GetBalance(suite.Ctx, suite.providerAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewInt(900), providerBalance.Amount)

	oracleBalance = suite.BankKeeper.GetBalance(suite.Ctx, suite.oracleAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewInt(100), oracleBalance.Amount)

	dealAccAddr, err = sdk.AccAddressFromBech32(deal.Address)
	suite.Require().NoError(err)
	dealBalance = suite.BankKeeper.GetBalance(suite.Ctx, dealAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewInt(9000), dealBalance.Amount)

	deal, err = suite.DataDealKeeper.GetDeal(suite.Ctx, dealID)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), deal.CurNumData)
	suite.Require().Equal(types.DEAL_STATUS_ACTIVE, deal.Status)
}

func (suite *consentTestSuite) TestSubmitConsentChangeStatusComplete() {
	budgetAmount := uint64(10000)
	dealID := suite.createSampleDeal(budgetAmount, 1)
	deal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, dealID)
	suite.Require().NoError(err)

	oracleCommissionRate := sdk.NewDecWithPrec(1, 1) // 10%
	suite.storeSampleOracle(suite.oracleAccAddr.String(), suite.uniqueID, oracleCommissionRate)

	unsignedCert := &types.UnsignedCertificate{
		Cid:             "cid",
		UniqueId:        suite.uniqueID,
		OracleAddress:   suite.oracleAccAddr.String(),
		DealId:          dealID,
		ProviderAddress: suite.providerAccAddr.String(),
		DataHash:        suite.dataHash,
	}

	unsignedCertBz, err := unsignedCert.Marshal()
	suite.Require().NoError(err)

	sign, err := suite.oraclePrivKey.Sign(unsignedCertBz)
	suite.Require().NoError(err)

	consent := &types.Consent{
		Certificate: &types.Certificate{
			UnsignedCertificate: unsignedCert,
			Signature:           sign.Serialize(),
		},
	}

	providerBalance := suite.BankKeeper.GetBalance(suite.Ctx, suite.providerAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.ZeroInt(), providerBalance.Amount)

	oracleBalance := suite.BankKeeper.GetBalance(suite.Ctx, suite.oracleAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.ZeroInt(), oracleBalance.Amount)

	dealAccAddr, err := sdk.AccAddressFromBech32(deal.Address)
	suite.Require().NoError(err)
	dealBalance := suite.BankKeeper.GetBalance(suite.Ctx, dealAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewIntFromUint64(budgetAmount), dealBalance.Amount)

	err = suite.DataDealKeeper.SubmitConsent(suite.Ctx, consent)
	suite.Require().NoError(err)

	providerBalance = suite.BankKeeper.GetBalance(suite.Ctx, suite.providerAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewInt(9000), providerBalance.Amount)

	oracleBalance = suite.BankKeeper.GetBalance(suite.Ctx, suite.oracleAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.NewInt(1000), oracleBalance.Amount)

	dealAccAddr, err = sdk.AccAddressFromBech32(deal.Address)
	suite.Require().NoError(err)
	dealBalance = suite.BankKeeper.GetBalance(suite.Ctx, dealAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(sdk.ZeroInt(), dealBalance.Amount)

	deal, err = suite.DataDealKeeper.GetDeal(suite.Ctx, dealID)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), deal.CurNumData)
	suite.Require().Equal(types.DEAL_STATUS_COMPLETED, deal.Status)
}

func (suite *consentTestSuite) TestSubmitConsentNotRegisteredOracle() {
	budgetAmount := uint64(10000)
	dealID := suite.createSampleDeal(budgetAmount, 1)

	unsignedCert := &types.UnsignedCertificate{
		Cid:             "cid",
		UniqueId:        suite.uniqueID,
		OracleAddress:   suite.providerAccAddr.String(),
		DealId:          dealID,
		ProviderAddress: suite.providerAccAddr.String(),
		DataHash:        suite.dataHash,
	}

	unsignedCertBz, err := unsignedCert.Marshal()
	suite.Require().NoError(err)

	sign, err := suite.oraclePrivKey.Sign(unsignedCertBz)
	suite.Require().NoError(err)

	consent := &types.Consent{
		Certificate: &types.Certificate{
			UnsignedCertificate: unsignedCert,
			Signature:           sign.Serialize(),
		},
	}

	err = suite.DataDealKeeper.SubmitConsent(suite.Ctx, consent)
	suite.Require().ErrorIs(err, types.ErrSubmitConsent)
	suite.Require().ErrorContains(err, fmt.Sprintf("failed to oracle validation. address(%s)", suite.providerAccAddr.String()))
}

func (suite *consentTestSuite) TestSubmitConsentNotSameUniqueIDOfOracle() {
	budgetAmount := uint64(10000)
	dealID := suite.createSampleDeal(budgetAmount, 1)

	oracleCommissionRate := sdk.NewDecWithPrec(1, 1) // 10%
	suite.storeSampleOracle(suite.oracleAccAddr.String(), "invalidUniqueID", oracleCommissionRate)

	unsignedCert := &types.UnsignedCertificate{
		Cid:             "cid",
		UniqueId:        suite.uniqueID,
		OracleAddress:   suite.oracleAccAddr.String(),
		DealId:          dealID,
		ProviderAddress: suite.providerAccAddr.String(),
		DataHash:        suite.dataHash,
	}

	unsignedCertBz, err := unsignedCert.Marshal()
	suite.Require().NoError(err)

	sign, err := suite.oraclePrivKey.Sign(unsignedCertBz)
	suite.Require().NoError(err)

	consent := &types.Consent{
		Certificate: &types.Certificate{
			UnsignedCertificate: unsignedCert,
			Signature:           sign.Serialize(),
		},
	}

	err = suite.DataDealKeeper.SubmitConsent(suite.Ctx, consent)
	suite.Require().ErrorIs(err, types.ErrSubmitConsent)
	suite.Require().ErrorContains(err, "is not active an oracle.")
}

func (suite *consentTestSuite) TestSubmitConsentInvalidSignature() {
	budgetAmount := uint64(10000)
	dealID := suite.createSampleDeal(budgetAmount, 1)

	oracleCommissionRate := sdk.NewDecWithPrec(1, 1) // 10%
	suite.storeSampleOracle(suite.oracleAccAddr.String(), suite.uniqueID, oracleCommissionRate)

	unsignedCert := &types.UnsignedCertificate{
		Cid:             "cid",
		UniqueId:        suite.uniqueID,
		OracleAddress:   suite.oracleAccAddr.String(),
		DealId:          dealID,
		ProviderAddress: suite.providerAccAddr.String(),
		DataHash:        suite.dataHash,
	}

	unsignedCertBz, err := unsignedCert.Marshal()
	suite.Require().NoError(err)

	sign, err := suite.invalidOraclePrivKey.Sign(unsignedCertBz)
	suite.Require().NoError(err)

	consent := &types.Consent{
		Certificate: &types.Certificate{
			UnsignedCertificate: unsignedCert,
			Signature:           sign.Serialize(),
		},
	}

	err = suite.DataDealKeeper.SubmitConsent(suite.Ctx, consent)
	suite.Require().ErrorIs(err, types.ErrSubmitConsent)
	suite.Require().ErrorContains(err, "failed to signature validation")
}

func (suite *consentTestSuite) TestSubmitConsentNotExistDeal() {
	oracleCommissionRate := sdk.NewDecWithPrec(1, 1) // 10%
	suite.storeSampleOracle(suite.oracleAccAddr.String(), suite.uniqueID, oracleCommissionRate)

	unsignedCert := &types.UnsignedCertificate{
		Cid:             "cid",
		UniqueId:        suite.uniqueID,
		OracleAddress:   suite.oracleAccAddr.String(),
		DealId:          1,
		ProviderAddress: suite.providerAccAddr.String(),
		DataHash:        suite.dataHash,
	}

	unsignedCertBz, err := unsignedCert.Marshal()
	suite.Require().NoError(err)

	sign, err := suite.oraclePrivKey.Sign(unsignedCertBz)
	suite.Require().NoError(err)

	consent := &types.Consent{
		Certificate: &types.Certificate{
			UnsignedCertificate: unsignedCert,
			Signature:           sign.Serialize(),
		},
	}

	err = suite.DataDealKeeper.SubmitConsent(suite.Ctx, consent)
	suite.Require().ErrorIs(err, types.ErrSubmitConsent)
	suite.Require().ErrorContains(err, "failed to get deal.")
}

func (suite *consentTestSuite) TestSubmitConsentAlreadyDealStatusComplete() {
	suite.TestSubmitConsentChangeStatusComplete()

	unsignedCert := &types.UnsignedCertificate{
		Cid:             "cid",
		UniqueId:        suite.uniqueID,
		OracleAddress:   suite.oracleAccAddr.String(),
		DealId:          1,
		ProviderAddress: suite.providerAccAddr.String(),
		DataHash:        "dataHash2",
	}

	unsignedCertBz, err := unsignedCert.Marshal()
	suite.Require().NoError(err)

	sign, err := suite.oraclePrivKey.Sign(unsignedCertBz)

	suite.Require().NoError(err)

	consent := &types.Consent{
		Certificate: &types.Certificate{
			UnsignedCertificate: unsignedCert,
			Signature:           sign.Serialize(),
		},
	}

	err = suite.DataDealKeeper.SubmitConsent(suite.Ctx, consent)
	suite.Require().ErrorIs(err, types.ErrSubmitConsent)
	suite.Require().ErrorContains(err, "deal status is not ACTIVE")
}

func (suite *consentTestSuite) TestSubmitConsentExistSameCertificate() {
	suite.TestSubmitConsentSuccess()

	unsignedCert := &types.UnsignedCertificate{
		Cid:             "cid",
		UniqueId:        suite.uniqueID,
		OracleAddress:   suite.oracleAccAddr.String(),
		DealId:          1,
		ProviderAddress: suite.providerAccAddr.String(),
		DataHash:        suite.dataHash,
	}

	unsignedCertBz, err := unsignedCert.Marshal()
	suite.Require().NoError(err)

	sign, err := suite.oraclePrivKey.Sign(unsignedCertBz)
	suite.Require().NoError(err)

	consent := &types.Consent{
		Certificate: &types.Certificate{
			UnsignedCertificate: unsignedCert,
			Signature:           sign.Serialize(),
		},
	}

	err = suite.DataDealKeeper.SubmitConsent(suite.Ctx, consent)
	suite.Require().ErrorIs(err, types.ErrSubmitConsent)
	suite.Require().ErrorContains(err, "already provided consent: error while submit consent")
}

func (suite *consentTestSuite) TestSubmitConsentNotSameUniqueIDOfCertificate() {
	budgetAmount := uint64(10000)
	dealID := suite.createSampleDeal(budgetAmount, 1)

	oracleCommissionRate := sdk.NewDecWithPrec(1, 1) // 10%
	suite.storeSampleOracle(suite.oracleAccAddr.String(), suite.uniqueID, oracleCommissionRate)

	invalidUniqueID := "invalidUniqueID"
	unsignedCert := &types.UnsignedCertificate{
		Cid:             "cid",
		UniqueId:        invalidUniqueID,
		OracleAddress:   suite.oracleAccAddr.String(),
		DealId:          dealID,
		ProviderAddress: suite.providerAccAddr.String(),
		DataHash:        suite.dataHash,
	}

	unsignedCertBz, err := unsignedCert.Marshal()
	suite.Require().NoError(err)

	sign, err := suite.oraclePrivKey.Sign(unsignedCertBz)
	suite.Require().NoError(err)

	consent := &types.Consent{
		Certificate: &types.Certificate{
			UnsignedCertificate: unsignedCert,
			Signature:           sign.Serialize(),
		},
	}

	err = suite.DataDealKeeper.SubmitConsent(suite.Ctx, consent)
	suite.Require().ErrorIs(err, types.ErrSubmitConsent)
	suite.Require().ErrorContains(err, fmt.Sprintf("does not match active uniqueID. certificateUniqueID(%s) activeUniqueID(%s)", invalidUniqueID, suite.uniqueID))
}
