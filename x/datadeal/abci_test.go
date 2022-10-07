package datadeal_test

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal"
	"github.com/medibloc/panacea-core/v2/x/datadeal/testutil"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
)

type abciTestSuite struct {
	testutil.DataDealBaseTestSuite

	sellerAccPrivKey cryptotypes.PrivKey
	sellerAccPubKey  cryptotypes.PubKey
	sellerAccAddr    sdk.AccAddress

	verifiableCID string

	oraclePubKey  cryptotypes.PubKey
	oracleAddr    sdk.AccAddress
	oraclePubKey2 cryptotypes.PubKey
	oracleAddr2   sdk.AccAddress
	oraclePubKey3 cryptotypes.PubKey
	oracleAddr3   sdk.AccAddress

	uniqueID string
}

func TestAbciTestSuite(t *testing.T) {
	suite.Run(t, new(abciTestSuite))
}

func (suite *abciTestSuite) BeforeTest(_, _ string) {

	ctx := suite.Ctx
	suite.uniqueID = "uniqueID"

	suite.sellerAccPrivKey = secp256k1.GenPrivKey()
	suite.sellerAccPubKey = suite.sellerAccPrivKey.PubKey()
	suite.sellerAccAddr = sdk.AccAddress(suite.sellerAccPubKey.Address())

	suite.verifiableCID = "verifiableCID"

	suite.oraclePubKey = secp256k1.GenPrivKey().PubKey()
	suite.oracleAddr = sdk.AccAddress(suite.oraclePubKey.Address())
	suite.oraclePubKey2 = secp256k1.GenPrivKey().PubKey()
	suite.oracleAddr2 = sdk.AccAddress(suite.oraclePubKey2.Address())
	suite.oraclePubKey3 = secp256k1.GenPrivKey().PubKey()
	suite.oracleAddr3 = sdk.AccAddress(suite.oraclePubKey3.Address())

	oraclePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	suite.Require().NoError(err)
	suite.OracleKeeper.SetParams(ctx, oracletypes.Params{
		OraclePublicKey:          base64.StdEncoding.EncodeToString(oraclePrivKey.PubKey().SerializeCompressed()),
		OraclePubKeyRemoteReport: base64.StdEncoding.EncodeToString([]byte("oraclePubKeyRemoteReport")),
		UniqueId:                 suite.uniqueID,
		OracleCommissionRate:     sdk.NewDecWithPrec(1, 1),
		VoteParams: oracletypes.VoteParams{
			VotingPeriod: 100,
			JailPeriod:   60,
			Threshold:    sdk.NewDec(2).Quo(sdk.NewDec(3)),
		},
		SlashParams: oracletypes.SlashParams{
			SlashFractionDowntime: sdk.NewDecWithPrec(3, 1),
			SlashFractionForgery:  sdk.NewDecWithPrec(1, 1),
		},
	})

}

func (suite abciTestSuite) TestDataDeliveryEndBlockerVotePass() {
	ctx := suite.Ctx

	suite.CreateOracleValidator(suite.oraclePubKey, sdk.NewInt(70))
	suite.CreateOracleValidator(suite.oraclePubKey2, sdk.NewInt(20))
	suite.CreateOracleValidator(suite.oraclePubKey3, sdk.NewInt(10))

	dataSale := &types.DataSale{
		SellerAddress: suite.sellerAccAddr.String(),
		DealId:        1,
		VerifiableCid: suite.verifiableCID,
		DeliveredCid:  "",
		Status:        types.DATA_SALE_STATUS_DELIVERY_VOTING_PERIOD,
		VotingPeriod: &oracletypes.VotingPeriod{
			VotingStartTime: time.Now().Add(-2 * time.Second),
			VotingEndTime:   time.Now().Add(-1 * time.Second),
		},
		VerificationTallyResult: nil,
		DeliveryTallyResult:     nil,
	}

	err := suite.DataDealKeeper.SetDataSale(ctx, dataSale)
	suite.Require().NoError(err)

	suite.DataDealKeeper.AddDataDeliveryQueue(
		ctx,
		dataSale.VerifiableCid,
		dataSale.DealId,
		dataSale.VotingPeriod.VotingEndTime,
	)

	vote := types.DataDeliveryVote{
		VoterAddress:  suite.oracleAddr.String(),
		DealId:        1,
		VerifiableCid: suite.verifiableCID,
		DeliveredCid:  "deliveredCID",
		VoteOption:    oracletypes.VOTE_OPTION_YES,
	}

	vote2 := types.DataDeliveryVote{
		VoterAddress:  suite.oracleAddr2.String(),
		DealId:        1,
		VerifiableCid: suite.verifiableCID,
		DeliveredCid:  "deliveredCID",
		VoteOption:    oracletypes.VOTE_OPTION_YES,
	}

	vote3 := types.DataDeliveryVote{
		VoterAddress:  suite.oracleAddr3.String(),
		DealId:        1,
		VerifiableCid: suite.verifiableCID,
		DeliveredCid:  "deliveredCID",
		VoteOption:    oracletypes.VOTE_OPTION_YES,
	}

	err = suite.DataDealKeeper.SetDataDeliveryVote(ctx, &vote)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDataDeliveryVote(ctx, &vote2)
	suite.Require().NoError(err)
	err = suite.DataDealKeeper.SetDataDeliveryVote(ctx, &vote3)
	suite.Require().NoError(err)

	votes, err := suite.DataDealKeeper.GetAllDataDeliveryVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(3, len(votes))

	datadeal.EndBlocker(ctx, suite.DataDealKeeper)

	updatedDataSale, err := suite.DataDealKeeper.GetDataSale(ctx, dataSale.VerifiableCid, dataSale.DealId)
	suite.Require().NoError(err)
	suite.Require().Equal(types.DATA_SALE_STATUS_COMPLETED, updatedDataSale.Status)
	suite.Require().Equal("deliveredCID", updatedDataSale.DeliveredCid)

	votes, err = suite.DataDealKeeper.GetAllDataDeliveryVoteList(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(votes))

	events := ctx.EventManager().Events()
	suite.Require().Equal(1, len(events))
	suite.Require().Equal(types.EventTypeDataDeliveryVote, events[0].Type)
	eventAttributes := events[0].Attributes
	suite.Require().Equal(2, len(eventAttributes))
	suite.Require().Equal(types.AttributeKeyVoteStatus, string(eventAttributes[0].Key))
	suite.Require().Equal(types.AttributeValueVoteStatusEnded, string(eventAttributes[0].Value))
	suite.Require().Equal(types.AttributeKeyDeliveredCID, string(eventAttributes[1].Key))

}