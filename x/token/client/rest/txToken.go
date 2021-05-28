package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/medibloc/panacea-core/x/token/types"
)

type issueTokenRequest struct {
	BaseReq          rest.BaseReq `json:"base_req"`
	Name             string       `json:"name"`
	ShortSymbol      string       `json:"shortSymbol"`
	TotalSupplyMicro string       `json:"totalSupplyMicro"`
	Mintable         bool         `json:"mintable"`
	OwnerAddress     string       `json:"ownerAddress"`
}

func issueTokenHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req issueTokenRequest
		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		totalSupplyMicro, ok := sdk.NewIntFromString(req.TotalSupplyMicro)
		if !ok {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid total supply in micro: %s", req.TotalSupplyMicro))
		}

		msg := types.NewMsgIssueToken(
			req.Name,
			req.ShortSymbol,
			totalSupplyMicro,
			req.Mintable,
			req.OwnerAddress,
		)

		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}
