package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/medibloc/panacea-core/x/currency/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/gorilla/mux"
)

const (
	RouteIssuance = "custom/currency/issuance"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	// Topic API
	r.HandleFunc(
		"/api/v1/currency/issuance/{denom}",
		getIssuanceHandlerFn(cliCtx),
	).Methods("GET")
}

func getIssuanceHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		denom := vars["denom"]

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewQueryIssuanceParams(denom)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData(RouteIssuance, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		cliCtx = cliCtx.WithHeight(height)

		var issuance types.Issuance
		if err := cliCtx.Codec.UnmarshalJSON(res, &issuance); err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		if issuance.Empty() {
			rest.WriteErrorResponse(w, http.StatusNotFound, "issuance not found")
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
