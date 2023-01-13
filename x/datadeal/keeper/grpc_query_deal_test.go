package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/medibloc/panacea-core/v2/x/datadeal/testutil"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/stretchr/testify/suite"
)

type queryDealTestSuite struct {
	testutil.DataDealBaseTestSuite

	oracleAccAddr   sdk.AccAddress
	providerAccAddr sdk.AccAddress
	consumerAccAddr sdk.AccAddress
}

func TestQueryDealTest(t *testing.T) {
	suite.Run(t, new(queryDealTestSuite))
}

func (suite *queryDealTestSuite) BeforeTest(_, _ string) {
	suite.oracleAccAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	suite.providerAccAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	suite.consumerAccAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
}

func (suite *queryDealTestSuite) TestQueryDeal() {
	deal := suite.MakeTestDeal(1, suite.consumerAccAddr, 100)
	err := suite.DataDealKeeper.SetDeal(suite.Ctx, deal)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetNextDealNumber(suite.Ctx, 2)
	suite.Require().NoError(err)

	req := types.QueryDealRequest{
		DealId: deal.Id,
	}
	res, err := suite.DataDealKeeper.Deal(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(deal, res.Deal)
}

func (suite *queryDealTestSuite) TestQueryDeals() {
	deal := suite.MakeTestDeal(1, suite.consumerAccAddr, 100)
	err := suite.DataDealKeeper.SetDeal(suite.Ctx, deal)
	suite.Require().NoError(err)

	deal2 := suite.MakeTestDeal(2, suite.consumerAccAddr, 10)
	err = suite.DataDealKeeper.SetDeal(suite.Ctx, deal2)
	suite.Require().NoError(err)

	req := types.QueryDealsRequest{}
	res, err := suite.DataDealKeeper.Deals(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(2, len(res.Deals))
	suite.Require().Equal(deal, res.Deals[0])
	suite.Require().Equal(deal2, res.Deals[1])
}

func (suite *queryDealTestSuite) TestQueryConsent() {
	consent := &types.Consent{
		DealId: 1,
		Certificate: &types.Certificate{
			UnsignedCertificate: &types.UnsignedCertificate{
				Cid:             "cid1",
				OracleAddress:   suite.oracleAccAddr.String(),
				DealId:          1,
				ProviderAddress: suite.providerAccAddr.String(),
				DataHash:        "dataHash",
			},
			Signature: []byte("signature"),
		},
		Agreements: []*types.Agreement{{TermId: 1, Agreement: true}},
	}

	err := suite.DataDealKeeper.SetConsent(suite.Ctx, consent)
	suite.Require().NoError(err)

	req := &types.QueryConsent{
		DealId:   consent.Certificate.UnsignedCertificate.DealId,
		DataHash: consent.Certificate.UnsignedCertificate.DataHash,
	}
	res, err := suite.DataDealKeeper.Consent(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(consent, res.Consent)
}

func (suite *queryDealTestSuite) TestQueryConsents() {
	consent := &types.Consent{
		DealId: 1,
		Certificate: &types.Certificate{
			UnsignedCertificate: &types.UnsignedCertificate{
				Cid:             "cid1",
				OracleAddress:   suite.oracleAccAddr.String(),
				DealId:          1,
				ProviderAddress: suite.providerAccAddr.String(),
				DataHash:        "dataHash",
			},
			Signature: []byte("signature"),
		},
		Agreements: []*types.Agreement{{TermId: 1, Agreement: true}},
	}

	err := suite.DataDealKeeper.SetConsent(suite.Ctx, consent)
	suite.Require().NoError(err)

	consent2 := &types.Consent{
		DealId: 1,
		Certificate: &types.Certificate{
			UnsignedCertificate: &types.UnsignedCertificate{
				Cid:             "cid2",
				OracleAddress:   suite.oracleAccAddr.String(),
				DealId:          1,
				ProviderAddress: suite.providerAccAddr.String(),
				DataHash:        "dataHash2",
			},
			Signature: []byte("signature"),
		},
	}

	err = suite.DataDealKeeper.SetConsent(suite.Ctx, consent2)
	suite.Require().NoError(err)

	consent3 := &types.Consent{
		DealId: 1,
		Certificate: &types.Certificate{
			UnsignedCertificate: &types.UnsignedCertificate{
				Cid:             "cid2",
				OracleAddress:   suite.oracleAccAddr.String(),
				DealId:          2,
				ProviderAddress: suite.providerAccAddr.String(),
				DataHash:        "dataHash2",
			},
			Signature: []byte("signature"),
		},
	}

	err = suite.DataDealKeeper.SetConsent(suite.Ctx, consent3)
	suite.Require().NoError(err)

	req := &types.QueryConsents{
		DealId:     1,
		Pagination: &query.PageRequest{},
	}
	res, err := suite.DataDealKeeper.Consents(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(res.Consents))
	suite.Require().Equal(consent, res.Consents[0])
	suite.Require().Equal(consent2, res.Consents[1])

	req.DealId = 2
	res, err = suite.DataDealKeeper.Consents(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(res.Consents))
	suite.Require().Equal(consent3, res.Consents[0])
}
