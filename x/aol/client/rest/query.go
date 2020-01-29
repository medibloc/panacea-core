package rest

import (
	"github.com/medibloc/panacea-core/x/aol/internal/keeper"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
	"github.com/medibloc/panacea-core/x/aol/internal/types"
)

const(
	RouteListTopic  = "custom/aol/listTopic"
	RouteTopic      = "custom/aol/topic"
	RouteListWriter = "custom/aol/listWriter"
	RouteWriter     = "custom/aol/writer"
	RouteRecord     = "custom/aol/record"
)

func listTopicHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32owner := vars["ownerAddr"]
		ownerAddr, err := sdk.AccAddressFromBech32(bech32owner)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		params := keeper.QueryListTopicParams{
			Owner: ownerAddr,
		}

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(RouteListTopic, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getTopicHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32owner := vars["ownerAddr"]
		ownerAddr, err := sdk.AccAddressFromBech32(bech32owner)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		topicName := vars["topic"]

		params := keeper.QueryTopicParams{
			Owner:     ownerAddr,
			TopicName: topicName,
		}

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(RouteTopic, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func listWriterHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32owner := vars["ownerAddr"]
		ownerAddr, err := sdk.AccAddressFromBech32(bech32owner)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		topicName := vars["topic"]

		params := keeper.QueryListWriterParams{
			Owner:     ownerAddr,
			TopicName: topicName,
		}

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(RouteListWriter, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getWriterHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32owner := vars["ownerAddr"]
		ownerAddr, err := sdk.AccAddressFromBech32(bech32owner)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		topicName := vars["topic"]
		bech32writer := vars["writerAddr"]
		writerAddr, err := sdk.AccAddressFromBech32(bech32writer)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		params := keeper.QueryWriterParams{
			Owner:     ownerAddr,
			TopicName: topicName,
			Writer:    writerAddr,
		}

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(RouteWriter, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getRecordHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32owner := vars["ownerAddr"]
		ownerAddr, err := sdk.AccAddressFromBech32(bech32owner)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		topicName := vars["topic"]
		strOffset := vars["offset"]
		offset, err := strconv.ParseUint(strOffset, 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		params := keeper.QueryRecordParams{
			Owner:     ownerAddr,
			TopicName: topicName,
			Offset:    offset,
		}

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(RouteRecord, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		var recordData types.Record
		err = cliCtx.Codec.UnmarshalJSON(res, &recordData)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		if recordData.IsEmpty() {
			rest.WriteErrorResponse(w, http.StatusNotFound, "record not found")
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
