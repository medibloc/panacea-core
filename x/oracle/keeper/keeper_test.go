package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

var (
	oracle1PrivKey = secp256k1.GenPrivKey()
	oracle1PubKey  = oracle1PrivKey.PubKey()
	oracle1Acc     = sdk.AccAddress(oracle1PubKey.Address())

	oracle2PrivKey = secp256k1.GenPrivKey()
	oracle2PubKey  = oracle2PrivKey.PubKey()
	oracle2Acc     = sdk.AccAddress(oracle2PubKey.Address())

	oracle3PrivKey = secp256k1.GenPrivKey()
	oracle3PubKey  = oracle3PrivKey.PubKey()
	oracle3Acc     = sdk.AccAddress(oracle3PubKey.Address())
)

type keeperTestSuite struct {
	testsuite.TestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(keeperTestSuite))
}

func (m keeperTestSuite) TestClosedVoteQueue() {
	now := time.Now()
	m.OracleKeeper.AddOracleRegistrationQueue(m.Ctx, uniqueID, oracle1Acc, now.Add(1*time.Second))
	m.OracleKeeper.AddOracleRegistrationQueue(m.Ctx, uniqueID, oracle2Acc, now.Add(3*time.Second))
	m.OracleKeeper.AddOracleRegistrationQueue(m.Ctx, uniqueID, oracle3Acc, now.Add(5*time.Second))

	iter := m.OracleKeeper.GetClosedOracleRegistrationQueueIterator(m.Ctx, now)
	m.Require().False(iter.Valid())

	iter = m.OracleKeeper.GetClosedOracleRegistrationQueueIterator(m.Ctx, now.Add(2*time.Second))

	addrList := make([]sdk.AccAddress, 0)
	for ; iter.Valid(); iter.Next() {
		addrList = append(addrList, iter.Value())
	}
	m.Require().Equal(1, len(addrList))
	m.Require().Equal(oracle1Acc, addrList[0])

	iter = m.OracleKeeper.GetClosedOracleRegistrationQueueIterator(m.Ctx, now.Add(4*time.Second))
	addrList = make([]sdk.AccAddress, 0)
	for ; iter.Valid(); iter.Next() {
		addrList = append(addrList, iter.Value())
	}
	m.Require().Equal(2, len(addrList))
	m.Require().Equal(oracle1Acc, addrList[0])
	m.Require().Equal(oracle2Acc, addrList[1])

	iter = m.OracleKeeper.GetClosedOracleRegistrationQueueIterator(m.Ctx, now.Add(6*time.Second))
	addrList = make([]sdk.AccAddress, 0)
	for ; iter.Valid(); iter.Next() {
		addrList = append(addrList, iter.Value())
	}
	m.Require().Equal(3, len(addrList))
	m.Require().Equal(oracle1Acc, addrList[0])
	m.Require().Equal(oracle2Acc, addrList[1])
	m.Require().Equal(oracle3Acc, addrList[2])

	// remove first queue and check
	m.OracleKeeper.RemoveOracleRegistrationQueue(m.Ctx, uniqueID, oracle1Acc, now.Add(1*time.Second))

	iter = m.OracleKeeper.GetClosedOracleRegistrationQueueIterator(m.Ctx, now.Add(6*time.Second))
	addrList = make([]sdk.AccAddress, 0)
	for ; iter.Valid(); iter.Next() {
		addrList = append(addrList, iter.Value())
	}
	m.Require().Equal(2, len(addrList))
	m.Require().Equal(oracle2Acc, addrList[0])
	m.Require().Equal(oracle3Acc, addrList[1])
}

func (suite *keeperTestSuite) createOracleValidator(pubKey cryptotypes.PubKey, amount sdk.Int) {
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

func (suite *keeperTestSuite) TestIterateOracleValidator() {
	pubKey := secp256k1.GenPrivKey().PubKey()
	address := sdk.AccAddress(pubKey.Address().Bytes()).String()
	tokens := sdk.NewInt(70)
	suite.createOracleValidator(pubKey, tokens)
	pubKey2 := secp256k1.GenPrivKey().PubKey()
	address2 := sdk.AccAddress(pubKey2.Address().Bytes()).String()
	tokens2 := sdk.NewInt(20)
	suite.createOracleValidator(pubKey2, tokens2)
	pubKey3 := secp256k1.GenPrivKey().PubKey()
	address3 := sdk.AccAddress(pubKey3.Address().Bytes()).String()
	tokens3 := sdk.NewInt(10)
	suite.createOracleValidator(pubKey3, tokens3)

	infos := make(map[string]*types.OracleValidatorInfo)
	suite.OracleKeeper.IterateOracleValidator(suite.Ctx, func(info *types.OracleValidatorInfo) bool {
		infos[info.Address] = info
		return false
	})

	suite.Require().Equal(3, len(infos))
	suite.Require().Equal(address, infos[address].Address)
	suite.Require().True(infos[address].OracleActivated)
	suite.Require().Equal(tokens, infos[address].BondedTokens)
	suite.Require().False(infos[address].ValidatorJailed)
	suite.Require().Equal(address2, infos[address2].Address)
	suite.Require().True(infos[address2].OracleActivated)
	suite.Require().Equal(tokens2, infos[address2].BondedTokens)
	suite.Require().False(infos[address2].ValidatorJailed)
	suite.Require().Equal(address3, infos[address3].Address)
	suite.Require().True(infos[address3].OracleActivated)
	suite.Require().Equal(tokens3, infos[address3].BondedTokens)
	suite.Require().False(infos[address3].ValidatorJailed)
}
