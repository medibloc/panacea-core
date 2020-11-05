package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/medibloc/panacea-core/x/token/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/gorilla/mux"
)

const (
	RouteToken      = "custom/token/token"
	RouteListTokens = "custom/token/listTokens"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	// Token API
	r.HandleFunc(
		"/api/v1/token/tokens/{symbol}",
		getTokenHandlerFn(cliCtx),
	).Methods("GET")
	r.HandleFunc(
		"/api/v1/token/tokens",
		listTokensHandlerFn(cliCtx),
	).Methods("GET")
}

func getTokenHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		symbol := types.Symbol(vars["symbol"])

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewQueryTokenParams(symbol)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData(RouteToken, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		cliCtx = cliCtx.WithHeight(height)

		var token types.Token
		if err := cliCtx.Codec.UnmarshalJSON(res, &token); err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		if token.Empty() {
			rest.WriteErrorResponse(w, http.StatusNotFound, "token not found")
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func listTokensHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.Query(RouteListTokens)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		cliCtx = cliCtx.WithHeight(height)

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
