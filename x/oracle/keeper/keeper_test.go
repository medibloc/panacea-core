package keeper_test

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func (m keeperTestSuite) TestEndVoteQueue() {
	now := time.Now()
	m.OracleKeeper.AddOracleRegistrationVoteQueue(m.Ctx, uniqueID, oracle1Acc, now.Add(1*time.Second))
	m.OracleKeeper.AddOracleRegistrationVoteQueue(m.Ctx, uniqueID, oracle2Acc, now.Add(3*time.Second))
	m.OracleKeeper.AddOracleRegistrationVoteQueue(m.Ctx, uniqueID, oracle3Acc, now.Add(5*time.Second))

	iter := m.OracleKeeper.GetEndOracleRegistrationVoteQueueIterator(m.Ctx, now)
	m.Require().False(iter.Valid())

	iter = m.OracleKeeper.GetEndOracleRegistrationVoteQueueIterator(m.Ctx, now.Add(2*time.Second))

	addrList := make([]sdk.AccAddress, 0)
	for ; iter.Valid(); iter.Next() {
		addrList = append(addrList, iter.Value())
	}
	m.Require().Equal(1, len(addrList))
	m.Require().Equal(oracle1Acc, addrList[0])

	iter = m.OracleKeeper.GetEndOracleRegistrationVoteQueueIterator(m.Ctx, now.Add(4*time.Second))
	addrList = make([]sdk.AccAddress, 0)
	for ; iter.Valid(); iter.Next() {
		addrList = append(addrList, iter.Value())
	}
	m.Require().Equal(2, len(addrList))
	m.Require().Equal(oracle1Acc, addrList[0])
	m.Require().Equal(oracle2Acc, addrList[1])

	iter = m.OracleKeeper.GetEndOracleRegistrationVoteQueueIterator(m.Ctx, now.Add(6*time.Second))
	addrList = make([]sdk.AccAddress, 0)
	for ; iter.Valid(); iter.Next() {
		addrList = append(addrList, iter.Value())
	}
	m.Require().Equal(3, len(addrList))
	m.Require().Equal(oracle1Acc, addrList[0])
	m.Require().Equal(oracle2Acc, addrList[1])
	m.Require().Equal(oracle3Acc, addrList[2])

	// remove first queue and check
	m.OracleKeeper.RemoveOracleRegistrationVoteQueue(m.Ctx, uniqueID, oracle1Acc, now.Add(1*time.Second))

	iter = m.OracleKeeper.GetEndOracleRegistrationVoteQueueIterator(m.Ctx, now.Add(6*time.Second))
	addrList = make([]sdk.AccAddress, 0)
	for ; iter.Valid(); iter.Next() {
		addrList = append(addrList, iter.Value())
	}
	m.Require().Equal(2, len(addrList))
	m.Require().Equal(oracle2Acc, addrList[0])
	m.Require().Equal(oracle3Acc, addrList[1])
}

func (m keeperTestSuite) Test() {
	lenTime := len(sdk.FormatTimeBytes(time.Now()))

	key := types.GetOracleRegistrationVoteQueueKey(uniqueID, oracle1Acc, time.Now())
	fmt.Println(len(key))
	fmt.Println(oracle1Acc.String())
	fmt.Println(sdk.AccAddress(key[len(key)-20:]).String())
	fmt.Println(string(key[1+lenTime+1 : len(key)-21]))
}
