package rest

import (
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
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
func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	// Topic API
	r.HandleFunc(
		"/api/v1/aol/{ownerAddr}/topics",
		listTopicHandlerFn(cliCtx),
	).Methods("GET")
	r.HandleFunc(
		"/api/v1/aol/{ownerAddr}/topics/{topic}",
		getTopicHandlerFn(cliCtx),
	).Methods("GET")

	// ACL API
	r.HandleFunc(
		"/api/v1/aol/{ownerAddr}/topics/{topic}/acl",
		listWriterHandlerFn(cliCtx),
	).Methods("GET")
	r.HandleFunc(
		"/api/v1/aol/{ownerAddr}/topics/{topic}/acl/{writerAddr}",
		getWriterHandlerFn(cliCtx),
	).Methods("GET")

	// Record API
	r.HandleFunc(
		"/api/v1/aol/{ownerAddr}/topics/{topic}/records/{offset}",
		getRecordHandlerFn(cliCtx),
	).Methods("GET")

}

func listTopicHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32owner := vars["ownerAddr"]
		ownerAddr, err := sdk.AccAddressFromBech32(bech32owner)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		params := types.QueryListTopicParams{
			Owner: ownerAddr,
		}

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData(RouteListTopic, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		cliCtx = cliCtx.WithHeight(height)

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getTopicHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32owner := vars["ownerAddr"]
		ownerAddr, err := sdk.AccAddressFromBech32(bech32owner)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		topicName := vars["topic"]

		params := types.QueryTopicParams{
			Owner:     ownerAddr,
			TopicName: topicName,
		}

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData(RouteTopic, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		cliCtx = cliCtx.WithHeight(height)

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func listWriterHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32owner := vars["ownerAddr"]
		ownerAddr, err := sdk.AccAddressFromBech32(bech32owner)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		topicName := vars["topic"]

		params := types.QueryListWriterParams{
			Owner:     ownerAddr,
			TopicName: topicName,
		}

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData(RouteListWriter, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		cliCtx = cliCtx.WithHeight(height)

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getWriterHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
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

		params := types.QueryWriterParams{
			Owner:     ownerAddr,
			TopicName: topicName,
			Writer:    writerAddr,
		}

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData(RouteWriter, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		cliCtx = cliCtx.WithHeight(height)

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getRecordHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
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

		params := types.QueryRecordParams{
			Owner:     ownerAddr,
			TopicName: topicName,
			Offset:    offset,
		}

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData(RouteRecord, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		cliCtx = cliCtx.WithHeight(height)

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
