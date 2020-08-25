package rest

import (
	"net/http"

	"github.com/medibloc/panacea-core/x/did"

	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/medibloc/panacea-core/x/did/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/gorilla/mux"
)

const (
	RouteDID = "custom/did/did"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec, queryRoute string) {
	// Topic API
	r.HandleFunc(
		"/api/v1/did/{did}",
		getDIDHandlerFn(cliCtx, cdc),
	).Methods("GET")
}

func getDIDHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := types.DID(vars["did"])
		if !id.Valid() {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "Invalid DID")
			return
		}

		params := did.QueryDIDParams{DID: id}
		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := cliCtx.QueryWithData(RouteDID, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}
