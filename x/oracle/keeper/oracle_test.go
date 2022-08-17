package keeper_test

import (
	"fmt"
	"testing"
	"time"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
)

var (
	uniqueID             = "correctUniqueID"
	genesisOraclePrivKey = secp256k1.GenPrivKey()
	genesisOraclePubKey  = genesisOraclePrivKey.PubKey()
	genesisOracleAcc     = sdk.AccAddress(genesisOraclePubKey.Address())

	newOraclePrivKey = secp256k1.GenPrivKey()
	newOraclePubKey  = newOraclePrivKey.PubKey()
	newOracleAcc     = sdk.AccAddress(newOraclePubKey.Address())

	oraclePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	oraclePubKey     = oraclePrivKey.PubKey()

	nodePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	nodePubKey     = nodePrivKey.PubKey()

	nodePubKeyRemoteReport = []byte("nodePubKeyRemoteReport")

	valPubKey = secp256k1.GenPrivKey().PubKey()
	valAddr   = valPubKey.Address()

	trustedBlockHeight = int64(1)
	trustedBlockHash   = []byte("trustedBlockHash")
)

type oracleTestSuite struct {
	testsuite.TestSuite
}

func TestOracleTestSuite(t *testing.T) {
	suite.Run(t, new(oracleTestSuite))
}

func (suite *oracleTestSuite) BeforeTest(_, _ string) {
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

func (suite *oracleTestSuite) setOracleAccount(pubKey cryptotypes.PubKey) {
	address := sdk.AccAddress(pubKey.Address())
	oracleAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, address)
	err := oracleAccount.SetPubKey(pubKey)
	suite.Require().NoError(err)
	suite.AccountKeeper.SetAccount(suite.Ctx, oracleAccount)
}

func makeNewOracleRegistration() *types.OracleRegistration {
	return &types.OracleRegistration{
		UniqueId:               uniqueID,
		Address:                newOracleAcc.String(),
		NodePubKey:             nodePubKey.SerializeCompressed(),
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

func (suite oracleTestSuite) makeNewValidator(operator sdk.ValAddress, pubKey cryptotypes.PubKey) *stakingtypes.Validator {
	v, err := stakingtypes.NewValidator(operator, pubKey, stakingtypes.Description{})
	suite.Require().NoError(err)
	return &v
}

func (suite oracleTestSuite) TestRegisterOracleSuccess() {
	ctx := suite.Ctx

	// set validator
	validator := suite.makeNewValidator(sdk.ValAddress(valAddr), valPubKey)
	suite.StakingKeeper.SetValidator(ctx, *validator)

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:               uniqueID,
		OracleAddress:          sdk.AccAddress(valAddr).String(),
		NodePubKey:             nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: nodePubKeyRemoteReport,
		TrustedBlockHeight:     trustedBlockHeight,
		TrustedBlockHash:       trustedBlockHash,
	}

	err := suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().NoError(err)

	votingPeriod := suite.OracleKeeper.GetVotingPeriod(ctx)

	oracleFromKeeper, err := suite.OracleKeeper.GetOracleRegistration(ctx, uniqueID, sdk.AccAddress(valAddr).String())
	suite.Require().NoError(err)
	suite.Require().Equal(uniqueID, oracleFromKeeper.UniqueId)
	suite.Require().Equal(sdk.AccAddress(valAddr).String(), oracleFromKeeper.Address)
	suite.Require().Equal(nodePubKey.SerializeCompressed(), oracleFromKeeper.NodePubKey)
	suite.Require().Equal(nodePubKeyRemoteReport, oracleFromKeeper.NodePubKeyRemoteReport)
	suite.Require().Equal(trustedBlockHeight, oracleFromKeeper.TrustedBlockHeight)
	suite.Require().Equal(trustedBlockHash, oracleFromKeeper.TrustedBlockHash)
	suite.Require().Nil(oracleFromKeeper.EncryptedOraclePrivKey)
	suite.Require().Equal(types.ORACLE_REGISTRATION_STATUS_VOTING_PERIOD, oracleFromKeeper.Status)
	suite.Require().Equal(votingPeriod, oracleFromKeeper.VotingPeriod)
	suite.Require().Nil(oracleFromKeeper.TallyResult)
}

func (suite oracleTestSuite) TestRegisterOracleFailedValidatorNotFound() {
	ctx := suite.Ctx

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:               uniqueID,
		OracleAddress:          sdk.AccAddress(valAddr).String(),
		NodePubKey:             nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: nodePubKeyRemoteReport,
		TrustedBlockHeight:     trustedBlockHeight,
		TrustedBlockHash:       trustedBlockHash,
	}

	err := suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, types.ErrValidatorNotFound)
}

func (suite oracleTestSuite) TestRegisterOracleFailedValidatorJailed() {
	ctx := suite.Ctx

	// set jailed validator
	validator := suite.makeNewValidator(sdk.ValAddress(valAddr), valPubKey)
	validator.Jailed = true
	suite.StakingKeeper.SetValidator(ctx, *validator)

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:               uniqueID,
		OracleAddress:          sdk.AccAddress(valAddr).String(),
		NodePubKey:             nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: nodePubKeyRemoteReport,
		TrustedBlockHeight:     trustedBlockHeight,
		TrustedBlockHash:       trustedBlockHash,
	}

	err := suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, types.ErrJailedValidator)
}

func (suite oracleTestSuite) TestRegisterOracleFailedInvalidUniqueID() {
	ctx := suite.Ctx

	// set validator
	validator := suite.makeNewValidator(sdk.ValAddress(valAddr), valPubKey)
	suite.StakingKeeper.SetValidator(ctx, *validator)

	msgRegisterOracle := &types.MsgRegisterOracle{
		UniqueId:               "invalidUniqueID",
		OracleAddress:          sdk.AccAddress(valAddr).String(),
		NodePubKey:             nodePubKey.SerializeCompressed(),
		NodePubKeyRemoteReport: nodePubKeyRemoteReport,
		TrustedBlockHeight:     trustedBlockHeight,
		TrustedBlockHash:       trustedBlockHash,
	}

	err := suite.OracleKeeper.RegisterOracle(ctx, msgRegisterOracle)
	suite.Require().Error(err, types.ErrRegisterOracle)
}

func (suite *oracleTestSuite) TestOracleRegistrationVoteSuccess() {
	ctx := suite.Ctx

	suite.setOracleAccount(genesisOraclePubKey)
	suite.setOracleAccount(newOraclePubKey)

	// make the correct genesis oracle
	oracle := &types.Oracle{
		Address:  genesisOracleAcc.String(),
		Status:   types.ORACLE_STATUS_ACTIVE,
		Uptime:   0,
		JailedAt: nil,
	}

	err := suite.OracleKeeper.SetOracle(ctx, oracle)
	suite.Require().NoError(err)

	oracleRegistration := makeNewOracleRegistration()
	err = suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	// make the correct encryptedOraclePrivKey
	encryptedOraclePrivKey, err := btcec.Encrypt(nodePubKey, oraclePrivKey.Serialize())
	suite.Require().NoError(err)
	// make the correct vote info
	oracleRegistrationVote := &types.OracleRegistrationVote{
		UniqueId:               uniqueID,
		VoterAddress:           genesisOracleAcc.String(),
		VotingTargetAddress:    newOracleAcc.String(),
		VoteOption:             types.VOTE_OPTION_VALID,
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	// make the correct signature
	voteBz, err := suite.Cdc.Marshaler.Marshal(oracleRegistrationVote)
	suite.Require().NoError(err)
	oraclePrivKeySecp256k1 := secp256k1.PrivKey{
		Key: oraclePrivKey.Serialize(),
	}
	signature, err := oraclePrivKeySecp256k1.Sign(voteBz)
	suite.Require().NoError(err)

	err = suite.OracleKeeper.VoteOracleRegistration(ctx, oracleRegistrationVote, signature)
	suite.Require().NoError(err)

	getOracleRegistrationVote, err := suite.OracleKeeper.GetOracleRegistrationVote(ctx, uniqueID, newOracleAcc.String(), genesisOracleAcc.String())
	suite.Require().NoError(err)
	suite.Require().Equal(oracleRegistrationVote, getOracleRegistrationVote)
}

func (suite *oracleTestSuite) TestOracleRegistrationVoteFailedVerifySignature() {
	ctx := suite.Ctx

	suite.setOracleAccount(genesisOraclePubKey)
	suite.setOracleAccount(newOraclePubKey)

	// make the correct genesis oracle
	oracle := &types.Oracle{
		Address:  genesisOracleAcc.String(),
		Status:   types.ORACLE_STATUS_ACTIVE,
		Uptime:   0,
		JailedAt: nil,
	}

	err := suite.OracleKeeper.SetOracle(ctx, oracle)
	suite.Require().NoError(err)

	oracleRegistration := makeNewOracleRegistration()
	err = suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	// make the correct encryptedOraclePrivKey
	encryptedOraclePrivKey, err := btcec.Encrypt(nodePubKey, oraclePrivKey.Serialize())
	suite.Require().NoError(err)
	// make the correct vote info
	oracleRegistrationVote := &types.OracleRegistrationVote{
		UniqueId:               uniqueID,
		VoterAddress:           genesisOracleAcc.String(),
		VotingTargetAddress:    newOracleAcc.String(),
		VoteOption:             types.VOTE_OPTION_VALID,
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	// make the correct signature
	voteBz, err := suite.Cdc.Marshaler.Marshal(oracleRegistrationVote)
	suite.Require().NoError(err)
	invalidOraclePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	suite.Require().NoError(err)
	oraclePrivKeySecp256k1 := secp256k1.PrivKey{
		Key: invalidOraclePrivKey.Serialize(),
	}
	signature, err := oraclePrivKeySecp256k1.Sign(voteBz)
	suite.Require().NoError(err)

	err = suite.OracleKeeper.VoteOracleRegistration(ctx, oracleRegistrationVote, signature)
	suite.Require().ErrorIs(err, types.ErrDetectionMaliciousBehavior)
}

func (suite *oracleTestSuite) TestOracleRegistrationVoteInvalidUniqueID() {
	ctx := suite.Ctx

	suite.setOracleAccount(genesisOraclePubKey)
	suite.setOracleAccount(newOraclePubKey)

	// make the correct genesis oracle
	oracle := &types.Oracle{
		Address:  genesisOracleAcc.String(),
		Status:   types.ORACLE_STATUS_ACTIVE,
		Uptime:   0,
		JailedAt: nil,
	}

	err := suite.OracleKeeper.SetOracle(ctx, oracle)
	suite.Require().NoError(err)

	oracleRegistration := makeNewOracleRegistration()
	err = suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	// make the correct encryptedOraclePrivKey
	encryptedOraclePrivKey, err := btcec.Encrypt(nodePubKey, oraclePrivKey.Serialize())
	suite.Require().NoError(err)
	// make vote with invalid uniqueID
	invalidUniqueID := "invalidUniqueID"
	oracleRegistrationVote := &types.OracleRegistrationVote{
		UniqueId:               invalidUniqueID,
		VoterAddress:           genesisOracleAcc.String(),
		VotingTargetAddress:    newOracleAcc.String(),
		VoteOption:             types.VOTE_OPTION_VALID,
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	// make the correct signature
	voteBz, err := suite.Cdc.Marshaler.Marshal(oracleRegistrationVote)
	suite.Require().NoError(err)
	oraclePrivKeySecp256k1 := secp256k1.PrivKey{
		Key: oraclePrivKey.Serialize(),
	}
	signature, err := oraclePrivKeySecp256k1.Sign(voteBz)
	suite.Require().NoError(err)

	err = suite.OracleKeeper.VoteOracleRegistration(ctx, oracleRegistrationVote, signature)
	suite.Require().ErrorIs(err, types.ErrOracleRegistrationVote)
	suite.Require().ErrorContains(err, fmt.Sprintf("not matched with the currently active uniqueID. expected %s, got %s", uniqueID, invalidUniqueID))
}

func (suite *oracleTestSuite) TestOracleRegistrationVoteInvalidGenesisOracleStatus() {
	ctx := suite.Ctx

	suite.setOracleAccount(genesisOraclePubKey)
	suite.setOracleAccount(newOraclePubKey)

	// make the correct genesis oracle
	oracle := &types.Oracle{
		Address:  genesisOracleAcc.String(),
		Status:   types.ORACLE_STATUS_JAILED,
		Uptime:   0,
		JailedAt: nil,
	}

	err := suite.OracleKeeper.SetOracle(ctx, oracle)
	suite.Require().NoError(err)

	oracleRegistration := makeNewOracleRegistration()
	err = suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	// make the correct encryptedOraclePrivKey
	encryptedOraclePrivKey, err := btcec.Encrypt(nodePubKey, oraclePrivKey.Serialize())
	suite.Require().NoError(err)
	// make vote with invalid uniqueID
	oracleRegistrationVote := &types.OracleRegistrationVote{
		UniqueId:               uniqueID,
		VoterAddress:           genesisOracleAcc.String(),
		VotingTargetAddress:    newOracleAcc.String(),
		VoteOption:             types.VOTE_OPTION_VALID,
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	// make the correct signature
	voteBz, err := suite.Cdc.Marshaler.Marshal(oracleRegistrationVote)
	suite.Require().NoError(err)
	oraclePrivKeySecp256k1 := secp256k1.PrivKey{
		Key: oraclePrivKey.Serialize(),
	}
	signature, err := oraclePrivKeySecp256k1.Sign(voteBz)
	suite.Require().NoError(err)

	err = suite.OracleKeeper.VoteOracleRegistration(ctx, oracleRegistrationVote, signature)
	suite.Require().ErrorIs(err, types.ErrOracleRegistrationVote)
	suite.Require().ErrorContains(err, "this oracle is not in 'ACTIVE' state")
}

func (suite *oracleTestSuite) TestOracleRegistrationVoteInvalidOracleRegistrationStatus() {
	ctx := suite.Ctx

	suite.setOracleAccount(genesisOraclePubKey)
	suite.setOracleAccount(newOraclePubKey)

	// make the correct genesis oracle
	oracle := &types.Oracle{
		Address:  genesisOracleAcc.String(),
		Status:   types.ORACLE_STATUS_ACTIVE,
		Uptime:   0,
		JailedAt: nil,
	}

	err := suite.OracleKeeper.SetOracle(ctx, oracle)
	suite.Require().NoError(err)

	oracleRegistration := makeNewOracleRegistration()
	oracleRegistration.Status = types.ORACLE_REGISTRATION_STATUS_REJECTED
	err = suite.OracleKeeper.SetOracleRegistration(ctx, oracleRegistration)
	suite.Require().NoError(err)

	// make the correct encryptedOraclePrivKey
	encryptedOraclePrivKey, err := btcec.Encrypt(nodePubKey, oraclePrivKey.Serialize())
	suite.Require().NoError(err)
	// make vote with invalid uniqueID
	oracleRegistrationVote := &types.OracleRegistrationVote{
		UniqueId:               uniqueID,
		VoterAddress:           genesisOracleAcc.String(),
		VotingTargetAddress:    newOracleAcc.String(),
		VoteOption:             types.VOTE_OPTION_VALID,
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	// make the correct signature
	voteBz, err := suite.Cdc.Marshaler.Marshal(oracleRegistrationVote)
	suite.Require().NoError(err)
	oraclePrivKeySecp256k1 := secp256k1.PrivKey{
		Key: oraclePrivKey.Serialize(),
	}
	signature, err := oraclePrivKeySecp256k1.Sign(voteBz)
	suite.Require().NoError(err)

	err = suite.OracleKeeper.VoteOracleRegistration(ctx, oracleRegistrationVote, signature)
	suite.Require().ErrorIs(err, types.ErrOracleRegistrationVote)
	suite.Require().ErrorContains(err, "the currently voted oracle's status is not 'VOTING_PERIOD'")
}
