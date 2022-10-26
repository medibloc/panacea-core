package keeper_test

import (
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/medibloc/panacea-core/v2/x/oracle/testutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"

	"github.com/stretchr/testify/suite"
)

type queryOracleTestSuite struct {
	testutil.OracleBaseTestSuite

	uniqueID string

	oracleAccPrivKey cryptotypes.PrivKey
	oracleAccPubKey  cryptotypes.PubKey
	oracleAccAddr    sdk.AccAddress

	newOracleAccPrivKey cryptotypes.PrivKey
	newOracleAccPubKey  cryptotypes.PubKey
	newOracleAccAddr    sdk.AccAddress

	nodePrivKey *btcec.PrivateKey
	nodePubKey  *btcec.PublicKey
}

func TestQueryOracleTestSuite(t *testing.T) {
	suite.Run(t, new(queryOracleTestSuite))
}

func (suite *queryOracleTestSuite) BeforeTest(_, _ string) {
	suite.uniqueID = "correctUniqueID"

	suite.oracleAccPrivKey = secp256k1.GenPrivKey()
	suite.oracleAccPubKey = suite.oracleAccPrivKey.PubKey()
	suite.oracleAccAddr = sdk.AccAddress(suite.oracleAccPubKey.Address())

	suite.newOracleAccPrivKey = secp256k1.GenPrivKey()
	suite.newOracleAccPubKey = suite.newOracleAccPrivKey.PubKey()
	suite.newOracleAccAddr = sdk.AccAddress(suite.newOracleAccPubKey.Address())

	suite.nodePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.nodePubKey = suite.nodePrivKey.PubKey()
}

func (suite *queryOracleTestSuite) makeNewOracleRegistration() *types.OracleRegistration {
	return &types.OracleRegistration{
		UniqueId:               suite.uniqueID,
		Address:                suite.newOracleAccAddr.String(),
		NodePubKey:             suite.nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: nil,
		TrustedBlockHeight:     0,
		TrustedBlockHash:       nil,
		EncryptedOraclePrivKey: nil,
		Status:                 types.ORACLE_REGISTRATION_STATUS_VOTING_PERIOD,
		VotingPeriod: &types.VotingPeriod{
			VotingStartTime: time.Now(),
			VotingEndTime:   time.Now().Add(5 * time.Second),
		},
	}
}

func (suite *queryOracleTestSuite) TestOracle() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	suite.CreateOracleValidator(suite.oracleAccPubKey, sdk.NewInt(100))

	req := types.QueryOracleRequest{
		Address: suite.oracleAccAddr.String(),
	}
	res, err := oracleKeeper.Oracle(sdk.WrapSDKContext(ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.oracleAccAddr.String(), res.Oracle.Address)
	suite.Require().Equal(types.ORACLE_STATUS_ACTIVE, res.Oracle.Status)
	suite.Require().Equal(uint64(0), res.Oracle.Uptime)
	suite.Require().Nil(res.Oracle.JailedAt)
}

func (suite *queryOracleTestSuite) TestOracles() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	suite.CreateOracleValidator(suite.oracleAccPubKey, sdk.NewInt(70))
	suite.CreateOracleValidator(suite.newOracleAccPubKey, sdk.NewInt(30))

	req := types.QueryOraclesRequest{
		Pagination: &query.PageRequest{},
	}

	res, err := oracleKeeper.Oracles(sdk.WrapSDKContext(ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(res.Oracles))
	for _, oracle := range res.Oracles {
		switch oracle.Address {
		case suite.oracleAccAddr.String():
			suite.Require().Equal(suite.oracleAccAddr.String(), oracle.Address)
			suite.Require().Equal(types.ORACLE_STATUS_ACTIVE, oracle.Status)
			suite.Require().Equal(uint64(0), oracle.Uptime)
			suite.Require().Nil(oracle.JailedAt)
		case suite.newOracleAccAddr.String():
			suite.Require().Equal(suite.newOracleAccAddr.String(), oracle.Address)
			suite.Require().Equal(types.ORACLE_STATUS_ACTIVE, oracle.Status)
			suite.Require().Equal(uint64(0), oracle.Uptime)
			suite.Require().Nil(oracle.JailedAt)
		default:
			panic("not found oracle address. address: " + oracle.Address)
		}
	}
}

func (suite *queryOracleTestSuite) TestOracleRegistration() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	newOracleRegistration := suite.makeNewOracleRegistration()

	err := oracleKeeper.SetOracleRegistration(ctx, newOracleRegistration)
	suite.Require().NoError(err)

	req := types.QueryOracleRegistrationRequest{
		UniqueId: newOracleRegistration.UniqueId,
		Address:  newOracleRegistration.Address,
	}

	res, err := oracleKeeper.OracleRegistration(sdk.WrapSDKContext(ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Equal(newOracleRegistration.UniqueId, res.OracleRegistration.UniqueId)
	suite.Require().Equal(newOracleRegistration.Address, res.OracleRegistration.Address)
	suite.Require().Equal(newOracleRegistration.NodePubKey, res.OracleRegistration.NodePubKey)
	suite.Require().Equal(newOracleRegistration.NodePubKeyRemoteReport, res.OracleRegistration.NodePubKeyRemoteReport)
	suite.Require().Equal(newOracleRegistration.TrustedBlockHash, res.OracleRegistration.TrustedBlockHash)
	suite.Require().Equal(newOracleRegistration.TrustedBlockHeight, res.OracleRegistration.TrustedBlockHeight)
}

func (suite *queryOracleTestSuite) TestOracleRegistrationVote() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	vote := types.OracleRegistrationVote{
		UniqueId:               suite.uniqueID,
		VoterAddress:           suite.oracleAccAddr.String(),
		VotingTargetAddress:    suite.newOracleAccAddr.String(),
		VoteOption:             types.VOTE_OPTION_YES,
		EncryptedOraclePrivKey: []byte("encryptedOraclePrivKey"),
	}
	err := oracleKeeper.SetOracleRegistrationVote(ctx, &vote)
	suite.Require().NoError(err)

	req := types.QueryOracleRegistrationVoteRequest{
		UniqueId:            suite.uniqueID,
		VoterAddress:        suite.oracleAccAddr.String(),
		VotingTargetAddress: suite.newOracleAccAddr.String(),
	}
	res, err := oracleKeeper.OracleRegistrationVote(sdk.WrapSDKContext(ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Equal(vote.UniqueId, res.OracleRegistrationVote.UniqueId)
	suite.Require().Equal(vote.VoterAddress, res.OracleRegistrationVote.VoterAddress)
	suite.Require().Equal(vote.VotingTargetAddress, res.OracleRegistrationVote.VotingTargetAddress)
}

func (suite *queryOracleTestSuite) TestOracleParams() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	oracleKeeper.SetParams(ctx, types.DefaultParams())

	req := types.QueryOracleParamsRequest{}
	res, err := oracleKeeper.Params(sdk.WrapSDKContext(ctx), &req)
	suite.Require().NoError(err)

	suite.Require().Equal(types.DefaultParams(), *res.Params)
}

func (suite *queryOracleTestSuite) TestOracleUpgradeInfo() {
	ctx := suite.Ctx
	oracleKeeper := suite.OracleKeeper

	upgradeInfo := &types.OracleUpgradeInfo{
		UniqueId: "UpgradeUniqueID",
		Height:   1000000,
	}
	suite.Require().NoError(oracleKeeper.SetOracleUpgradeInfo(ctx, upgradeInfo))

	getUpgradeInfo, err := oracleKeeper.OracleUpgradeInfo(sdk.WrapSDKContext(ctx), nil)
	suite.Require().NoError(err)
	suite.Require().Equal(upgradeInfo.UniqueId, getUpgradeInfo.OracleUpgradeInfo.UniqueId)
	suite.Require().Equal(upgradeInfo.Height, getUpgradeInfo.OracleUpgradeInfo.Height)
}
