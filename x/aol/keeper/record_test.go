package keeper_test

import (
	"testing"

	"github.com/cometbft/cometbft/crypto/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	aoltypes "github.com/medibloc/panacea-core/v2/x/aol/types"
	"github.com/stretchr/testify/suite"
)

var (
	recordPubKey = secp256k1.GenPrivKey().PubKey()

	recordAddress = sdk.AccAddress(recordPubKey.Address())
)

type recordTestSuite struct {
	testsuite.TestSuite
}

func TestRecordKeeper(t *testing.T) {
	suite.Run(t, new(recordTestSuite))
}

func (suite *recordTestSuite) TestOneRecord() {
	ctx := suite.Ctx
	aolKeeper := suite.AolKeeper

	topicName := "topicName"
	key := aoltypes.RecordCompositeKey{
		OwnerAddress: recordAddress,
		TopicName:    topicName,
		Offset:       0,
	}
	record := aoltypes.Record{
		Key:           []byte("recordKey"),
		Value:         []byte("recordValue"),
		NanoTimestamp: ctx.BlockTime().UnixNano(),
		WriterAddress: address.String(),
	}
	aolKeeper.SetRecord(ctx, key, record)

	suite.Require().True(aolKeeper.HasRecord(ctx, key))

	resultRecord := aolKeeper.GetRecord(ctx, key)
	suite.Require().Equal(record, resultRecord)

	resultKeys, resultRecords := aolKeeper.GetAllRecords(ctx)
	suite.Require().Equal(1, len(resultKeys))
	suite.Require().Equal([]aoltypes.RecordCompositeKey{key}, resultKeys)
	suite.Require().Equal(1, len(resultRecords))
	suite.Require().Equal([]aoltypes.Record{record}, resultRecords)
}

func (suite *recordTestSuite) TestMultiRecord() {
	ctx := suite.Ctx
	aolKeeper := suite.AolKeeper

	topicName := "topicName"
	key := aoltypes.RecordCompositeKey{
		OwnerAddress: recordAddress,
		TopicName:    topicName,
		Offset:       0,
	}
	key2 := aoltypes.RecordCompositeKey{
		OwnerAddress: recordAddress,
		TopicName:    topicName,
		Offset:       1,
	}
	record := aoltypes.Record{
		Key:           []byte("recordKey"),
		Value:         []byte("recordValue"),
		NanoTimestamp: ctx.BlockTime().UnixNano(),
		WriterAddress: address.String(),
	}
	record2 := aoltypes.Record{
		Key:           []byte("recordKey2"),
		Value:         []byte("recordValue2"),
		NanoTimestamp: ctx.BlockTime().UnixNano(),
		WriterAddress: address.String(),
	}
	aolKeeper.SetRecord(ctx, key, record)
	aolKeeper.SetRecord(ctx, key2, record2)

	suite.Require().True(aolKeeper.HasRecord(ctx, key))
	suite.Require().True(aolKeeper.HasRecord(ctx, key2))

	resultRecord := aolKeeper.GetRecord(ctx, key)
	suite.Require().Equal(record, resultRecord)
	resultRecord2 := aolKeeper.GetRecord(ctx, key2)
	suite.Require().Equal(record2, resultRecord2)

	resultKeys, resultRecords := aolKeeper.GetAllRecords(ctx)
	suite.Require().Equal(2, len(resultKeys))
	suite.Require().Contains(resultKeys, key)
	suite.Require().Contains(resultKeys, key2)
	suite.Require().Equal(2, len(resultRecords))
	suite.Require().Contains(resultRecords, record)
	suite.Require().Contains(resultRecords, record2)
}
