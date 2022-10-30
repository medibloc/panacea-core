package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/testutil"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
)

type keeperTestSuite struct {
	testutil.OracleBaseTestSuite

	uniqueID         string
	oracleAccPrivKey cryptotypes.PrivKey
	oracleAccPubKey  cryptotypes.PubKey
	oracleAccAddr    sdk.AccAddress

	oracleAccPrivKey2 cryptotypes.PrivKey
	oracleAccPubKey2  cryptotypes.PubKey
	oracleAccAddr2    sdk.AccAddress

	oracleAccPrivKey3 cryptotypes.PrivKey
	oracleAccPubKey3  cryptotypes.PubKey
	oracleAccAddr3    sdk.AccAddress
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(keeperTestSuite))
}

func (suite *keeperTestSuite) BeforeTest(_, _ string) {
	suite.uniqueID = "correctUniqueID"

	suite.oracleAccPrivKey = secp256k1.GenPrivKey()
	suite.oracleAccPubKey = suite.oracleAccPrivKey.PubKey()
	suite.oracleAccAddr = sdk.AccAddress(suite.oracleAccPubKey.Address())

	suite.oracleAccPrivKey2 = secp256k1.GenPrivKey()
	suite.oracleAccPubKey2 = suite.oracleAccPrivKey2.PubKey()
	suite.oracleAccAddr2 = sdk.AccAddress(suite.oracleAccPubKey2.Address())

	suite.oracleAccPrivKey3 = secp256k1.GenPrivKey()
	suite.oracleAccPubKey3 = suite.oracleAccPrivKey3.PubKey()
	suite.oracleAccAddr3 = sdk.AccAddress(suite.oracleAccPubKey3.Address())
}

func (suite *keeperTestSuite) TestClosedVoteQueue() {
	now := time.Now()
	suite.OracleKeeper.AddOracleRegistrationQueue(suite.Ctx, suite.uniqueID, suite.oracleAccAddr, now.Add(1*time.Second))
	suite.OracleKeeper.AddOracleRegistrationQueue(suite.Ctx, suite.uniqueID, suite.oracleAccAddr2, now.Add(3*time.Second))
	suite.OracleKeeper.AddOracleRegistrationQueue(suite.Ctx, suite.uniqueID, suite.oracleAccAddr3, now.Add(5*time.Second))

	iter := suite.OracleKeeper.GetClosedOracleRegistrationQueueIterator(suite.Ctx, now)
	suite.Require().False(iter.Valid())

	iter = suite.OracleKeeper.GetClosedOracleRegistrationQueueIterator(suite.Ctx, now.Add(2*time.Second))

	addrList := make([]sdk.AccAddress, 0)
	for ; iter.Valid(); iter.Next() {
		addrList = append(addrList, iter.Value())
	}
	suite.Require().Equal(1, len(addrList))
	suite.Require().Equal(suite.oracleAccAddr, addrList[0])

	iter = suite.OracleKeeper.GetClosedOracleRegistrationQueueIterator(suite.Ctx, now.Add(4*time.Second))
	addrList = make([]sdk.AccAddress, 0)
	for ; iter.Valid(); iter.Next() {
		addrList = append(addrList, iter.Value())
	}
	suite.Require().Equal(2, len(addrList))
	suite.Require().Equal(suite.oracleAccAddr, addrList[0])
	suite.Require().Equal(suite.oracleAccAddr2, addrList[1])

	iter = suite.OracleKeeper.GetClosedOracleRegistrationQueueIterator(suite.Ctx, now.Add(6*time.Second))
	addrList = make([]sdk.AccAddress, 0)
	for ; iter.Valid(); iter.Next() {
		addrList = append(addrList, iter.Value())
	}
	suite.Require().Equal(3, len(addrList))
	suite.Require().Equal(suite.oracleAccAddr, addrList[0])
	suite.Require().Equal(suite.oracleAccAddr2, addrList[1])
	suite.Require().Equal(suite.oracleAccAddr3, addrList[2])

	// remove first queue and check
	suite.OracleKeeper.RemoveOracleRegistrationQueue(suite.Ctx, suite.uniqueID, suite.oracleAccAddr, now.Add(1*time.Second))

	iter = suite.OracleKeeper.GetClosedOracleRegistrationQueueIterator(suite.Ctx, now.Add(6*time.Second))
	addrList = make([]sdk.AccAddress, 0)
	for ; iter.Valid(); iter.Next() {
		addrList = append(addrList, iter.Value())
	}
	suite.Require().Equal(2, len(addrList))
	suite.Require().Equal(suite.oracleAccAddr2, addrList[0])
	suite.Require().Equal(suite.oracleAccAddr3, addrList[1])
}

func (suite *keeperTestSuite) TestIterateOracleValidator() {
	pubKey := secp256k1.GenPrivKey().PubKey()
	address := sdk.AccAddress(pubKey.Address().Bytes()).String()
	tokens := sdk.NewInt(70)
	suite.CreateOracleValidator(pubKey, tokens)
	pubKey2 := secp256k1.GenPrivKey().PubKey()
	address2 := sdk.AccAddress(pubKey2.Address().Bytes()).String()
	tokens2 := sdk.NewInt(20)
	suite.CreateOracleValidator(pubKey2, tokens2)
	pubKey3 := secp256k1.GenPrivKey().PubKey()
	address3 := sdk.AccAddress(pubKey3.Address().Bytes()).String()
	tokens3 := sdk.NewInt(10)
	suite.CreateOracleValidator(pubKey3, tokens3)

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
