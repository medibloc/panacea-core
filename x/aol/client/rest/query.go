package rest

import (
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
	"github.com/medibloc/panacea-core/x/aol"
	"github.com/medibloc/panacea-core/x/aol/types"
)

const (
	RouteListTopic  = "custom/aol/listTopic"
	RouteTopic      = "custom/aol/topic"
	RouteListWriter = "custom/aol/listWriter"
	RouteWriter     = "custom/aol/writer"
	RouteRecord     = "custom/aol/record"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec, queryRoute string) {
	// Topic API
	r.HandleFunc(
		"/api/v1/aol/{ownerAddr}/topics",
		listTopicHandlerFn(cliCtx, cdc),
	).Methods("GET")
	r.HandleFunc(
		"/api/v1/aol/{ownerAddr}/topics/{topic}",
		getTopicHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// ACL API
	r.HandleFunc(
		"/api/v1/aol/{ownerAddr}/topics/{topic}/acl",
		listWriterHandlerFn(cliCtx, cdc),
	).Methods("GET")
	r.HandleFunc(
		"/api/v1/aol/{ownerAddr}/topics/{topic}/acl/{writerAddr}",
		getWriterHandlerFn(cliCtx, cdc),
	).Methods("GET")

	// Record API
	r.HandleFunc(
		"/api/v1/aol/{ownerAddr}/topics/{topic}/records/{offset}",
		getRecordHandlerFn(cliCtx, cdc),
	).Methods("GET")

}

func listTopicHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32owner := vars["ownerAddr"]
		ownerAddr, err := sdk.AccAddressFromBech32(bech32owner)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		params := aol.QueryListTopicParams{
			Owner: ownerAddr,
		}

		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := cliCtx.QueryWithData(RouteListTopic, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func getTopicHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32owner := vars["ownerAddr"]
		ownerAddr, err := sdk.AccAddressFromBech32(bech32owner)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		topicName := vars["topic"]

		params := aol.QueryTopicParams{
			Owner:     ownerAddr,
			TopicName: topicName,
		}

		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := cliCtx.QueryWithData(RouteTopic, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func listWriterHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32owner := vars["ownerAddr"]
		ownerAddr, err := sdk.AccAddressFromBech32(bech32owner)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		topicName := vars["topic"]

		params := aol.QueryListWriterParams{
			Owner:     ownerAddr,
			TopicName: topicName,
		}

		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := cliCtx.QueryWithData(RouteListWriter, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func getWriterHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
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

		params := aol.QueryWriterParams{
			Owner:     ownerAddr,
			TopicName: topicName,
			Writer:    writerAddr,
		}

		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := cliCtx.QueryWithData(RouteWriter, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func getRecordHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
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

		params := aol.QueryRecordParams{
			Owner:     ownerAddr,
			TopicName: topicName,
			Offset:    offset,
		}

		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := cliCtx.QueryWithData(RouteRecord, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		var recordData types.Record
		err = cdc.UnmarshalJSON(res, &recordData)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		if recordData.IsEmpty() {
			rest.WriteErrorResponse(w, http.StatusNotFound, "record not found")
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}
