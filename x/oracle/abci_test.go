package oracle_test

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/oracle"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type abciTestSuite struct {
	testsuite.TestSuite

	uniqueID string

	oraclePubKey  cryptotypes.PubKey
	oracleAddr    sdk.AccAddress
	oraclePubKey2 cryptotypes.PubKey
	oracleAddr2   sdk.AccAddress
	oraclePubKey3 cryptotypes.PubKey
	oracleAddr3   sdk.AccAddress

	newOraclePubKey cryptotypes.PubKey
	newOracleAddr   sdk.AccAddress
}

func TestAbciTestSuite(t *testing.T) {
	suite.Run(t, new(abciTestSuite))
}

func (suite *abciTestSuite) BeforeTest(_, _ string) {
	ctx := suite.Ctx
	suite.uniqueID = "uniqueID"

	suite.oraclePubKey = secp256k1.GenPrivKey().PubKey()
	suite.oracleAddr = sdk.AccAddress(suite.oraclePubKey.Address())
	suite.oraclePubKey2 = secp256k1.GenPrivKey().PubKey()
	suite.oracleAddr2 = sdk.AccAddress(suite.oraclePubKey2.Address())
	suite.oraclePubKey3 = secp256k1.GenPrivKey().PubKey()
	suite.oracleAddr3 = sdk.AccAddress(suite.oraclePubKey3.Address())

	suite.createOracleValidator(suite.oraclePubKey, sdk.NewInt(50))
	suite.createOracleValidator(suite.oraclePubKey2, sdk.NewInt(30))
	suite.createOracleValidator(suite.oraclePubKey3, sdk.NewInt(10))

	suite.newOraclePubKey = secp256k1.GenPrivKey().PubKey()
	suite.newOracleAddr = sdk.AccAddress(suite.newOraclePubKey.Address())

	oraclePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	suite.Require().NoError(err)
	suite.OracleKeeper.SetParams(ctx, types.Params{
		OraclePublicKey:          oraclePrivKey.PubKey().SerializeCompressed(),
		OraclePubKeyRemoteReport: []byte("oraclePubKeyRemoteReport"),
		UniqueId:                 uniqueID,
		VoteParams: types.VoteParams{
			VotingPeriod: 100,
			JailPeriod:   60,
			Quorum:       sdk.NewDecWithPrec(1, 3),
		},
		SlashParams: types.SlashParams{
			SlashFractionDowntime: sdk.NewDecWithPrec(3, 1),
			SlashFractionForgery:  sdk.NewDecWithPrec(1, 1),
		},
	})
}

func (suite *abciTestSuite) createOracleValidator(pubKey cryptotypes.PubKey, amount sdk.Int) {
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

	err = suite.OracleKeeper.SetOracle(suite.Ctx, &types.Oracle{
		Address:  oracleAccAddr.String(),
		Status:   types.ORACLE_STATUS_ACTIVE,
		Uptime:   0,
		JailedAt: nil,
	})
	suite.Require().NoError(err)
}

func (suite abciTestSuite) TestEndBlocker() {
	ctx := suite.Ctx

	nodePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	suite.Require().NoError(err)
	oracleRegistration := &types.OracleRegistration{
		UniqueId:               suite.uniqueID,
		Address:                suite.newOracleAddr.String(),
		NodePubKey:             nodePrivKey.PubKey().SerializeCompressed(),
		NodePubKeyRemoteReport: []byte("nodePubKeyRemoteReport"),
		TrustedBlockHeight:     10,
		TrustedBlockHash:       []byte("trustedBlockHash"),
		Status:                 types.ORACLE_REGISTRATION_STATUS_VOTING_PERIOD,
		VotingPeriod: &types.VotingPeriod{
			VotingStartTime: time.Now().Add(-2 * time.Second),
			VotingEndTime:   time.Now().Add(-1 * time.Second),
		},
	}
	err = suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	suite.OracleKeeper.AddOracleRegistrationQueue(
		ctx,
		suite.uniqueID,
		suite.newOracleAddr,
		oracleRegistration.VotingPeriod.VotingEndTime,
	)

	vote := types.OracleRegistrationVote{
		UniqueId:               suite.uniqueID,
		VoterAddress:           suite.oracleAddr.String(),
		VotingTargetAddress:    suite.newOracleAddr.String(),
		VoteOption:             types.VOTE_OPTION_VALID,
		EncryptedOraclePrivKey: []byte("encryptedOraclePrivKey"),
	}
	vote2 := types.OracleRegistrationVote{
		UniqueId:               suite.uniqueID,
		VoterAddress:           suite.oracleAddr2.String(),
		VotingTargetAddress:    suite.newOracleAddr.String(),
		VoteOption:             types.VOTE_OPTION_VALID,
		EncryptedOraclePrivKey: []byte("encryptedOraclePrivKey"),
	}
	vote3 := types.OracleRegistrationVote{
		UniqueId:               suite.uniqueID,
		VoterAddress:           suite.oracleAddr3.String(),
		VotingTargetAddress:    suite.newOracleAddr.String(),
		VoteOption:             types.VOTE_OPTION_VALID,
		EncryptedOraclePrivKey: []byte("encryptedOraclePrivKey"),
	}
	err = suite.OracleKeeper.SetOracleRegistrationVote(ctx, &vote)
	suite.Require().NoError(err)
	err = suite.OracleKeeper.SetOracleRegistrationVote(ctx, &vote2)
	suite.Require().NoError(err)
	err = suite.OracleKeeper.SetOracleRegistrationVote(ctx, &vote3)
	suite.Require().NoError(err)

	oracle.EndBlocker(suite.Ctx, suite.OracleKeeper)

	oracleRegistration, err = suite.OracleKeeper.GetOracleRegistration(ctx, suite.uniqueID, suite.newOracleAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(types.ORACLE_REGISTRATION_STATUS_PASSED, oracleRegistration.Status)

	newOracle, err := suite.OracleKeeper.GetOracle(ctx, suite.newOracleAddr.String())
	suite.Require().NoError(err)
	suite.Require().Equal(suite.newOracleAddr.String(), newOracle.Address)
	suite.Require().Equal(types.ORACLE_STATUS_ACTIVE, newOracle.Status)
}
