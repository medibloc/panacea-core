package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	aoltypes "github.com/medibloc/panacea-core/v2/x/aol/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

var (
	pubKey = secp256k1.GenPrivKey().PubKey()

	address = sdk.AccAddress(pubKey.Address())
)

type topicTestSuite struct {
	testsuite.TestSuite
}

func TestTopicKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(topicTestSuite))
}

func (suite topicTestSuite) TestOneTopic() {
	ctx := suite.Ctx
	aolKeeper := suite.AolKeeper

	topicName := "topicName"
	key := aoltypes.TopicCompositeKey{
		OwnerAddress: address,
		TopicName:    topicName,
	}
	topic := aoltypes.Topic{
		Description: "test description",
	}

	aolKeeper.SetTopic(ctx, key, topic)

	// verify HasTopic
	suite.Require().True(aolKeeper.HasTopic(ctx, key))

	// verify GetTopic
	resultTopic := aolKeeper.GetTopic(ctx, key)
	suite.Require().Equal(topic.Description, resultTopic.Description)
	suite.Require().Equal(uint64(0), resultTopic.TotalRecords)
	suite.Require().Equal(uint64(0), resultTopic.TotalWriters)

	resultKeys, resultAllTopics := aolKeeper.GetAllTopics(ctx)
	suite.Require().Equal(1, len(resultKeys))
	suite.Require().Equal(address, resultKeys[0].OwnerAddress)
	suite.Require().Equal(topicName, resultKeys[0].TopicName)
	suite.Require().Equal(1, len(resultAllTopics))
	suite.Require().Equal(topic.Description, resultAllTopics[0].Description)
	suite.Require().Equal(uint64(0), resultAllTopics[0].TotalRecords)
	suite.Require().Equal(uint64(0), resultAllTopics[0].TotalWriters)
}

func (suite topicTestSuite) TestMultiTopic() {
	ctx := suite.Ctx
	aolKeeper := suite.AolKeeper

	topicName := "topicName"
	topicName2 := "topicName2"
	key := aoltypes.TopicCompositeKey{
		OwnerAddress: address,
		TopicName:    topicName,
	}
	key2 := aoltypes.TopicCompositeKey{
		OwnerAddress: address,
		TopicName:    topicName2,
	}
	topic := aoltypes.Topic{
		Description: "test description",
	}
	topic2 := aoltypes.Topic{
		Description: "test description2",
	}

	aolKeeper.SetTopic(ctx, key, topic)
	aolKeeper.SetTopic(ctx, key2, topic2)

	// verify HasTopic
	suite.Require().True(aolKeeper.HasTopic(ctx, key))
	suite.Require().True(aolKeeper.HasTopic(ctx, key2))

	// verify GetTopic
	resultTopic := aolKeeper.GetTopic(ctx, key)
	suite.Require().Equal(topic.Description, resultTopic.Description)
	suite.Require().Equal(uint64(0), resultTopic.TotalRecords)
	suite.Require().Equal(uint64(0), resultTopic.TotalWriters)
	resultTopic2 := aolKeeper.GetTopic(ctx, key2)
	suite.Require().Equal(topic2.Description, resultTopic2.Description)
	suite.Require().Equal(uint64(0), resultTopic2.TotalRecords)
	suite.Require().Equal(uint64(0), resultTopic2.TotalWriters)

	// verify GetAllTopic
	resultKeys, resultAllTopics := aolKeeper.GetAllTopics(ctx)
	suite.Require().Equal(2, len(resultKeys))
	suite.Require().Equal(address, resultKeys[0].OwnerAddress)
	suite.Require().Equal(topicName, resultKeys[0].TopicName)
	suite.Require().Equal(address, resultKeys[1].OwnerAddress)
	suite.Require().Equal(topicName2, resultKeys[1].TopicName)

	suite.Require().Equal(2, len(resultAllTopics))
	suite.Require().Equal(topic.Description, resultAllTopics[0].Description)
	suite.Require().Equal(uint64(0), resultAllTopics[0].TotalRecords)
	suite.Require().Equal(uint64(0), resultAllTopics[0].TotalWriters)
	suite.Require().Equal(topic2.Description, resultAllTopics[1].Description)
	suite.Require().Equal(uint64(0), resultAllTopics[1].TotalRecords)
	suite.Require().Equal(uint64(0), resultAllTopics[1].TotalWriters)
}
