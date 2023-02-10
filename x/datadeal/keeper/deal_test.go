package keeper_test

import (
	"encoding/json"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datadeal/testutil"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/stretchr/testify/suite"
)

var (
	intFilterType = "integer"
)

type dealTestSuite struct {
	testutil.DataDealBaseTestSuite

	defaultFunds    sdk.Coins
	consumerAccAddr sdk.AccAddress
	providerAccAddr sdk.AccAddress

	pdBz []byte
}

func TestDealTestSuite(t *testing.T) {
	suite.Run(t, new(dealTestSuite))
}

func (suite *dealTestSuite) BeforeTest(_, _ string) {
	suite.consumerAccAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	suite.providerAccAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	suite.defaultFunds = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000)))

	testDeal := suite.MakeTestDeal(1, suite.consumerAccAddr, 100)
	err := suite.DataDealKeeper.SetNextDealNumber(suite.Ctx, 2)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDeal(suite.Ctx, testDeal)
	suite.Require().NoError(err)

	required := presexch.Required
	pd := &presexch.PresentationDefinition{
		ID:      "c1b88ce1-8460-4baf-8f16-4759a2f055fd",
		Purpose: "To sell you a drink we need to know that you are an adult.",
		InputDescriptors: []*presexch.InputDescriptor{{
			ID:      "age_descriptor",
			Purpose: "Your age should be greater or equal to 18.",
			// required temporarily in v0.1.8 for schema verification.
			// schema will be optional by supporting presentation exchange v2
			// https://github.com/hyperledger/aries-framework-go/commit/66d9bf30de2f5cd6116adaac27f277b45077f26f
			Schema: []*presexch.Schema{{
				URI:      "https://www.w3.org/2018/credentials#VerifiableCredential",
				Required: false,
			}, {
				URI:      "https://w3id.org/security/bbs/v1",
				Required: false,
			}},
			Constraints: &presexch.Constraints{
				LimitDisclosure: &required,
				Fields: []*presexch.Field{
					{
						Path: []string{"$.credentialSubject.age"},
						Filter: &presexch.Filter{
							Type:    &intFilterType,
							Minimum: 18,
							Maximum: 30,
						},
					},
				},
			},
		}},
	}

	suite.pdBz, err = json.Marshal(pd)
	suite.Require().NoError(err)
}

func (suite *dealTestSuite) TestCreateNewDeal() {
	err := suite.FundAccount(suite.Ctx, suite.consumerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	budget := &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(10000000)}

	msgCreateDeal := &types.MsgCreateDeal{
		DataSchema:      []string{"http://jsonld.com"},
		Budget:          budget,
		MaxNumData:      10000,
		ConsumerAddress: suite.consumerAccAddr.String(),
		AgreementTerms: []*types.AgreementTerm{
			{
				Id:          1,
				Required:    true,
				Title:       "title",
				Description: "description",
			},
		},
		PresentationDefinition: suite.pdBz,
	}

	dealID, err := suite.DataDealKeeper.CreateDeal(suite.Ctx, msgCreateDeal)
	suite.Require().NoError(err)

	expectedId, err := suite.DataDealKeeper.GetAndIncreaseNextDealNumber(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(dealID, expectedId-uint64(1))

	deal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, dealID)
	suite.Require().NoError(err)
	suite.Require().Equal(deal.GetDataSchema(), msgCreateDeal.GetDataSchema())
	suite.Require().Equal(deal.GetBudget(), msgCreateDeal.GetBudget())
	suite.Require().Equal(deal.GetMaxNumData(), msgCreateDeal.GetMaxNumData())
	suite.Require().Equal(deal.GetConsumerAddress(), msgCreateDeal.GetConsumerAddress())
	suite.Require().Equal(deal.GetAgreementTerms(), msgCreateDeal.GetAgreementTerms())
	suite.Require().Equal(deal.GetStatus(), types.DEAL_STATUS_ACTIVE)
}

func (suite *dealTestSuite) TestCreateDealInvalidPD() {
	err := suite.FundAccount(suite.Ctx, suite.consumerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	budget := &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(10000000)}

	invalidPD := &presexch.PresentationDefinition{
		ID:      "c1b88ce1-8460-4baf-8f16-4759a2f055fd",
		Purpose: "To sell you a drink we need to know that you are an adult.",
		InputDescriptors: []*presexch.InputDescriptor{{
			Schema: []*presexch.Schema{{
				URI:      "https://www.w3.org/2018/credentials#VerifiableCredential",
				Required: false,
			}, {
				URI:      "https://w3id.org/security/bbs/v1",
				Required: false,
			}},
		}},
	}
	invalidPDBz, err := json.Marshal(invalidPD)
	suite.Require().NoError(err)

	msgCreateDeal := &types.MsgCreateDeal{
		DataSchema:      []string{"http://jsonld.com"},
		Budget:          budget,
		MaxNumData:      10000,
		ConsumerAddress: suite.consumerAccAddr.String(),
		AgreementTerms: []*types.AgreementTerm{
			{
				Id:          1,
				Required:    true,
				Title:       "title",
				Description: "description",
			},
		},
		PresentationDefinition: invalidPDBz,
	}

	err = msgCreateDeal.ValidateBasic()
	suite.Require().ErrorContains(err, "invalid presentation definition")
}

func (suite *dealTestSuite) TestCheckDealCurNumDataAndIncrement() {
	err := suite.FundAccount(suite.Ctx, suite.consumerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	budget := &sdk.Coin{Denom: assets.MicroMedDenom, Amount: sdk.NewInt(10000000)}

	msgCreateDeal := &types.MsgCreateDeal{
		DataSchema:      []string{"http://jsonld.com"},
		Budget:          budget,
		MaxNumData:      1,
		ConsumerAddress: suite.consumerAccAddr.String(),
	}

	dealID, err := suite.DataDealKeeper.CreateDeal(suite.Ctx, msgCreateDeal)
	suite.Require().NoError(err)

	deal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, dealID)
	suite.Require().NoError(err)

	check := deal.IsCompleted()
	suite.Equal(false, check)

	err = suite.DataDealKeeper.IncreaseCurNumDataOfDeal(suite.Ctx, dealID)
	suite.Require().NoError(err)
	updatedDeal, err := suite.DataDealKeeper.GetDeal(suite.Ctx, dealID)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), updatedDeal.CurNumData)

	check = updatedDeal.IsCompleted()
	suite.Require().Equal(true, check)
}

func (suite *dealTestSuite) TestRequestDeactivateDeal() {
	ctx := suite.Ctx

	err := suite.FundAccount(ctx, suite.consumerAccAddr, suite.defaultFunds)
	suite.Require().NoError(err)

	getDeal, err := suite.DataDealKeeper.GetDeal(ctx, 1)
	suite.Require().NoError(err)

	dealAccAddr, err := sdk.AccAddressFromBech32(getDeal.Address)
	suite.Require().NoError(err)

	// Sending Budget from buyer to deal
	err = suite.BankKeeper.SendCoins(suite.Ctx, suite.consumerAccAddr, dealAccAddr, sdk.NewCoins(*getDeal.Budget))
	suite.Require().NoError(err)

	msgDeactivateDeal := &types.MsgDeactivateDeal{
		DealId:           1,
		RequesterAddress: suite.consumerAccAddr.String(),
	}

	// Consumer Balance = Original Consumer Balance(10000000000umed) - Deal's Budget(1000000000umed) --> 9000000000umed
	beforeConsumerBalance := suite.BankKeeper.GetBalance(suite.Ctx, suite.consumerAccAddr, assets.MicroMedDenom)

	err = suite.DataDealKeeper.DeactivateDeal(ctx, msgDeactivateDeal)
	suite.Require().NoError(err)

	getDeal, err = suite.DataDealKeeper.GetDeal(ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(getDeal.Status, types.DEAL_STATUS_INACTIVE)

	// After deactivating a deal, the consumer get the refund from deal.
	afterConsumerBalance := suite.BankKeeper.GetBalance(suite.Ctx, suite.consumerAccAddr, assets.MicroMedDenom)
	suite.Require().Equal(beforeConsumerBalance.Add(*getDeal.Budget), afterConsumerBalance)
}

func (suite *dealTestSuite) TestRequestDeactivateDealInvalidRequester() {
	ctx := suite.Ctx

	msgDeactivateDeal := &types.MsgDeactivateDeal{
		DealId:           1,
		RequesterAddress: suite.providerAccAddr.String(),
	}

	err := suite.DataDealKeeper.DeactivateDeal(ctx, msgDeactivateDeal)
	suite.Require().ErrorIs(err, types.ErrDeactivateDeal)
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
		RequesterAddress: suite.consumerAccAddr.String(),
	}

	err = suite.DataDealKeeper.DeactivateDeal(ctx, msgDeactivateDeal)
	suite.Require().ErrorIs(err, types.ErrDeactivateDeal)
}
