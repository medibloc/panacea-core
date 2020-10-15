package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/medibloc/panacea-core/x/did/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/gorilla/mux"
)

const (
	RouteDID = "custom/did/did"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	// Topic API
	r.HandleFunc(
		"/api/v1/did/{did}",
		getDIDHandlerFn(cliCtx),
	).Methods("GET")
}

func getDIDHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := types.DID(vars["did"])
		if !id.Valid() {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "Invalid DID")
			return
		}

		params := types.NewQueryDIDParams(id)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData(RouteDID, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		cliCtx = cliCtx.WithHeight(height)

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
