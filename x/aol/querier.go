package aol

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// query endpoints supported by the aol querier
const (
	QueryListTopic  = "listTopic"
	QueryTopic      = "topic"
	QueryListWriter = "listWriter"
	QueryWriter     = "writer"
	QueryRecord     = "record"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryListTopic:
			return queryListTopic(ctx, path[1:], req, keeper)
		case QueryTopic:
			return queryTopic(ctx, path[1:], req, keeper)
		case QueryListWriter:
			return queryListWriter(ctx, path[1:], req, keeper)
		case QueryWriter:
			return queryWriter(ctx, path[1:], req, keeper)
		case QueryRecord:
			return queryRecord(ctx, path[1:], req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown aol query endpoint")
		}
	}
}

type QueryListTopicParams struct {
	Owner sdk.AccAddress
}

func queryListTopic(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params QueryListTopicParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, k.ListTopic(ctx, params.Owner))
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

type QueryTopicParams struct {
	Owner     sdk.AccAddress
	TopicName string
}

func queryTopic(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params QueryTopicParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, k.GetTopic(ctx, params.Owner, params.TopicName))
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

type QueryListWriterParams struct {
	Owner     sdk.AccAddress
	TopicName string
}

func queryListWriter(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params QueryListWriterParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, k.ListWriter(ctx, params.Owner, params.TopicName))
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

type QueryWriterParams struct {
	Owner     sdk.AccAddress
	TopicName string
	Writer    sdk.AccAddress
}

func queryWriter(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params QueryWriterParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, k.GetWriter(ctx, params.Owner, params.TopicName, params.Writer))
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

type QueryRecordParams struct {
	Owner     sdk.AccAddress
	TopicName string
	Offset    uint64
}

func queryRecord(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params QueryRecordParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, k.GetRecord(ctx, params.Owner, params.TopicName, params.Offset))
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}
