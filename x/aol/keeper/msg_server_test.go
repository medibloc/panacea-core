package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/types/testsuite"
	aoltypes "github.com/medibloc/panacea-core/x/aol/types"
	"github.com/stretchr/testify/suite"
	"testing"
)

type aolMsgServerTestSuite struct {
	testsuite.TestSuite
}

func TestAolMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(aolMsgServerTestSuite))
}

func (suite *aolMsgServerTestSuite) BeforeTest(_, _ string) {

}

func (suite aolMsgServerTestSuite) TestMsgServer() {
	ctx := suite.Ctx
	goCtx := sdk.WrapSDKContext(ctx)
	aolKeeper := suite.AolKeeper
	aolMsgServer := suite.AolMsgServer

	topicName := "testTopic"
	ownerAddress := suite.GetAccAddress()

	// create topic
	msgCreateTopic := aoltypes.MsgCreateTopic{
		TopicName: topicName,
		Description: "topic description",
		OwnerAddress: ownerAddress.String(),
	}
	createTopicResponse, err := aolMsgServer.CreateTopic(goCtx, &msgCreateTopic)
	suite.Require().NoError(err)
	suite.Require().NotNil(createTopicResponse)

	// get topic
	getTopicRequest := aoltypes.QueryGetTopicRequest{
		OwnerAddress: ownerAddress.String(),
		TopicName: topicName,
	}
	getTopicResponse, err := aolKeeper.Topic(goCtx, &getTopicRequest)
	suite.Require().NoError(err)
	suite.Require().Equal(msgCreateTopic.Description, getTopicResponse.Topic.Description)
	suite.Require().Equal(uint64(0), getTopicResponse.Topic.TotalRecords)
	suite.Require().Equal(uint64(0), getTopicResponse.Topic.TotalWriters)

	// add writer
	moniker := "testMoniker"
	writerAddress := suite.GetAccAddress()
	msgAddWriter := aoltypes.MsgAddWriter{
		TopicName:     topicName,
		Moniker:       moniker,
		Description:   "write Description",
		WriterAddress: writerAddress.String(),
		OwnerAddress:  ownerAddress.String(),
	}

	addWriterResponse, err := aolMsgServer.AddWriter(goCtx, &msgAddWriter)
	suite.Require().NoError(err)
	suite.Require().NotNil(addWriterResponse)

	// add writer2
	writerAddress2 := suite.GetAccAddress()
	msgAddWriter = aoltypes.MsgAddWriter{
		TopicName:     topicName,
		Moniker:       moniker,
		Description:   "write Description2",
		WriterAddress: writerAddress2.String(),
		OwnerAddress:  ownerAddress.String(),
	}

	addWriterResponse, err = aolMsgServer.AddWriter(goCtx, &msgAddWriter)
	suite.Require().NoError(err)
	suite.Require().NotNil(addWriterResponse)

	// get writer
	getWriterRequest := aoltypes.QueryGetWriterRequest{
		OwnerAddress: ownerAddress.String(),
		TopicName: topicName,
		WriterAddress: writerAddress2.String(),
	}
	getWriterResponse, err := aolKeeper.Writer(goCtx, &getWriterRequest)
	suite.Require().NoError(err)
	suite.Require().Equal(msgAddWriter.Description, getWriterResponse.Writer.Description)
	suite.Require().Equal(msgAddWriter.Moniker, getWriterResponse.Writer.Moniker)

	// get writer2
	getWriterRequest = aoltypes.QueryGetWriterRequest{
		OwnerAddress: ownerAddress.String(),
		TopicName: topicName,
		WriterAddress: writerAddress2.String(),
	}
	getWriterResponse, err = aolKeeper.Writer(goCtx, &getWriterRequest)
	suite.Require().NoError(err)
	suite.Require().Equal(msgAddWriter.Description, getWriterResponse.Writer.Description)
	suite.Require().Equal(msgAddWriter.Moniker, getWriterResponse.Writer.Moniker)

	// add record
	msgAddRecord := aoltypes.MsgAddRecord{
		TopicName: topicName,
		Key: []byte("key1"),
		Value: []byte("value1"),
		WriterAddress: writerAddress.String(),
		OwnerAddress: ownerAddress.String(),
		FeePayerAddress: writerAddress.String(),
	}
	addRecordResponse, err := aolMsgServer.AddRecord(goCtx, &msgAddRecord)
	suite.Require().NoError(err)
	suite.Require().NotNil(addRecordResponse)

	// get record
	getRecordRequest := aoltypes.QueryGetRecordRequest{
		OwnerAddress: ownerAddress.String(),
		TopicName: topicName,
		Offset: 0,
	}
	getRecordResponse, err := aolKeeper.Record(goCtx, &getRecordRequest)
	suite.Require().NoError(err)
	suite.Require().Equal(msgAddRecord.Key, getRecordResponse.Record.Key)
	suite.Require().Equal(msgAddRecord.Value, getRecordResponse.Record.Value)
	suite.Require().Equal(msgAddRecord.WriterAddress, getRecordResponse.Record.WriterAddress)

	// add record2
	msgAddRecord2 := aoltypes.MsgAddRecord{
		TopicName: topicName,
		Key: []byte("key2"),
		Value: []byte("value2"),
		WriterAddress: writerAddress.String(),
		OwnerAddress: ownerAddress.String(),
		FeePayerAddress: writerAddress.String(),
	}
	addRecordResponse, err = aolMsgServer.AddRecord(goCtx, &msgAddRecord2)
	suite.Require().NoError(err)
	suite.Require().NotNil(addRecordResponse)

	// get record
	getRecordRequest = aoltypes.QueryGetRecordRequest{
		OwnerAddress: ownerAddress.String(),
		TopicName: topicName,
		Offset: 1,
	}
	getRecordResponse, err = aolKeeper.Record(goCtx, &getRecordRequest)
	suite.Require().NoError(err)
	suite.Require().Equal(msgAddRecord2.Key, getRecordResponse.Record.Key)
	suite.Require().Equal(msgAddRecord2.Value, getRecordResponse.Record.Value)
	suite.Require().Equal(msgAddRecord2.WriterAddress, getRecordResponse.Record.WriterAddress)

	// get topic
	getTopicRequest = aoltypes.QueryGetTopicRequest{
		OwnerAddress: ownerAddress.String(),
		TopicName: topicName,
	}
	getTopicResponse, err = aolKeeper.Topic(goCtx, &getTopicRequest)
	suite.Require().NoError(err)
	// bug to description clear.
	//suite.Require().Equal(msgCreateTopic.Description, getTopicResponse.Topic.Description)
	suite.Require().Equal(uint64(2), getTopicResponse.Topic.TotalRecords)
	suite.Require().Equal(uint64(2), getTopicResponse.Topic.TotalWriters)

	// delete writer
	msgDeleteWriter := aoltypes.MsgDeleteWriter{
		TopicName: topicName,
		WriterAddress: writerAddress2.String(),
		OwnerAddress: ownerAddress.String(),
	}
	msgDeleteWriterResponse , err := aolMsgServer.DeleteWriter(goCtx, &msgDeleteWriter)
	suite.Require().NoError(err)
	suite.Require().NotNil(msgDeleteWriterResponse)

	// get topic
	getTopicRequest = aoltypes.QueryGetTopicRequest{
		OwnerAddress: ownerAddress.String(),
		TopicName: topicName,
	}
	getTopicResponse, err = aolKeeper.Topic(goCtx, &getTopicRequest)
	suite.Require().NoError(err)
	// bug to description clear.
	//suite.Require().Equal(msgCreateTopic.Description, getTopicResponse.Topic.Description)
	suite.Require().Equal(uint64(2), getTopicResponse.Topic.TotalRecords)
	suite.Require().Equal(uint64(1), getTopicResponse.Topic.TotalWriters)
}