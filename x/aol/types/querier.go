package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// query endpoints supported by the aol querier
const (
	QueryListTopic  = "listTopic"
	QueryTopic      = "topic"
	QueryListWriter = "listWriter"
	QueryWriter     = "writer"
	QueryRecord     = "record"
)

type QueryListTopicParams struct {
	Owner sdk.AccAddress
}

func NewQueryListTopicParams(owner sdk.AccAddress) *QueryListTopicParams {
	return &QueryListTopicParams{Owner: owner}
}

type QueryTopicParams struct {
	Owner     sdk.AccAddress
	TopicName string
}

func NewQueryTopicParams(owner sdk.AccAddress, topicName string) *QueryTopicParams {
	return &QueryTopicParams{Owner: owner, TopicName: topicName}
}

type QueryListWriterParams struct {
	Owner     sdk.AccAddress
	TopicName string
}

func NewQueryListWriterParams(owner sdk.AccAddress, topicName string) *QueryListWriterParams {
	return &QueryListWriterParams{Owner: owner, TopicName: topicName}
}

type QueryWriterParams struct {
	Owner     sdk.AccAddress
	TopicName string
	Writer    sdk.AccAddress
}

func NewQueryWriterParams(owner sdk.AccAddress, topicName string, writer sdk.AccAddress) *QueryWriterParams {
	return &QueryWriterParams{Owner: owner, TopicName: topicName, Writer: writer}
}

type QueryRecordParams struct {
	Owner     sdk.AccAddress
	TopicName string
	Offset    uint64
}

func NewQueryRecordParams(owner sdk.AccAddress, topicName string, offset uint64) *QueryRecordParams {
	return &QueryRecordParams{Owner: owner, TopicName: topicName, Offset: offset}
}
