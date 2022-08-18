package keeper_test

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type tallyTestSuite struct {
	testsuite.TestSuite
}

func TestTallyTestSuite(t *testing.T) {
	suite.Run(t, new(tallyTestSuite))
}

func (suite *tallyTestSuite) BeforeTest(_, _ string) {
	ctx := suite.Ctx

	suite.OracleKeeper.SetParams(ctx, types.Params{
		OraclePublicKey:          oraclePubKey.SerializeCompressed(),
		OraclePubKeyRemoteReport: nil,
		UniqueId:                 uniqueID,
		VoteParams: types.VoteParams{
			VotingPeriod: 100,
			JailPeriod:   60,
			Quorum:       sdk.NewDecWithPrec(2, 3),
		},
		SlashParams: types.SlashParams{
			SlashFractionDowntime: sdk.NewDecWithPrec(3, 1),
			SlashFractionForgery:  sdk.NewDecWithPrec(1, 1),
		},
	})
}

// createOracleValidator defines to register Oracle and Validator by default.
func (suite *tallyTestSuite) createOracleValidator(pubKey cryptotypes.PubKey, amount sdk.Int) {
	oracleAccAddr := sdk.AccAddress(pubKey.Address().Bytes())
	oracleAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, oracleAccAddr)
	err := oracleAccount.SetPubKey(pubKey)
	suite.Require().NoError(err)
	suite.AccountKeeper.SetAccount(suite.Ctx, oracleAccount)
	varAddr := sdk.ValAddress(pubKey.Address().Bytes())
	validator, err := stakingtypes.NewValidator(varAddr, pubKey, stakingtypes.Description{})
	suite.Require().NoError(err)
	validator = validator.UpdateStatus(stakingtypes.Bonded)
	validator, _ = validator.AddTokensFromDel(amount)

	suite.StakingKeeper.SetValidator(suite.Ctx, validator)

	oracle := &types.Oracle{
		Address:  oracleAccAddr.String(),
		Status:   types.ORACLE_STATUS_ACTIVE,
		Uptime:   0,
		JailedAt: nil,
	}

	err = suite.OracleKeeper.SetOracle(suite.Ctx, oracle)
	suite.Require().NoError(err)
}

func (suite *tallyTestSuite) TestTally() {
	ctx := suite.Ctx
	oraclePubKey := secp256k1.GenPrivKey().PubKey()
	oracleAccAddr := sdk.AccAddress(oraclePubKey.Address().Bytes())
	oracleTokens := sdk.NewInt(30)

	suite.createOracleValidator(oraclePubKey, oracleTokens)

	newOraclePubKey := secp256k1.GenPrivKey().PubKey()
	newOracleAccAddr := sdk.AccAddress(newOraclePubKey.Address().Bytes())

	nodePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	suite.Require().NoError(err)

	oracleRegistration := &types.OracleRegistration{
		UniqueId:               uniqueID,
		Address:                newOracleAccAddr.String(),
		NodePubKey:             nodePrivKey.PubKey().SerializeCompressed(),
		NodePubKeyRemoteReport: []byte("nodePubKey"),
		TrustedBlockHeight:     1,
		TrustedBlockHash:       []byte("Hash"),
		Status:                 types.ORACLE_REGISTRATION_STATUS_VOTING_PERIOD,
		VotingPeriod: &types.VotingPeriod{
			VotingStartTime: time.Now(),
			VotingEndTime:   time.Now(),
		},
	}
	err = suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	consensusValue := []byte("encPriv1")
	vote := &types.OracleRegistrationVote{
		UniqueId:               uniqueID,
		VoterAddress:           oracleAccAddr.String(),
		VotingTargetAddress:    newOracleAccAddr.String(),
		VoteOption:             types.VOTE_OPTION_YES,
		EncryptedOraclePrivKey: consensusValue,
	}
	err = suite.OracleKeeper.SetOracleRegistrationVote(suite.Ctx, vote)
	require.NoError(suite.T(), err)

	oracleVotes, err := suite.OracleKeeper.GetAllOracleRegistrationVoteList(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(oracleVotes))

	iter := suite.OracleKeeper.GetOracleRegistrationVoteIterator(suite.Ctx, uniqueID, newOracleAccAddr.String())
	tallyResult, err := suite.GetTallyKeeper().Tally(
		suite.Ctx,
		iter,
		&types.OracleRegistrationVote{},
		func(vote types.Vote) error {
			return suite.OracleKeeper.RemoveOracleRegistrationVote(suite.Ctx, vote.(*types.OracleRegistrationVote))
		},
	)
	suite.Require().NoError(err)

	suite.Require().Equal(oracleTokens, tallyResult.Yes)
	suite.Require().Equal(sdk.ZeroInt(), tallyResult.No)
	suite.Require().Equal(0, len(tallyResult.InvalidYes))
	suite.Require().Equal(consensusValue, tallyResult.ConsensusValue)

	oracleVotes, err = suite.OracleKeeper.GetAllOracleRegistrationVoteList(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(oracleVotes))
}
