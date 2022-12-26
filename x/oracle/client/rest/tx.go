package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

type PlanRequest struct {
	BaseReq         rest.BaseReq `json:"base_req" yaml:"base_req"`
	Title           string       `json:"title" yaml:"title"`
	Description     string       `json:"description" yaml:"description"`
	Deposit         sdk.Coins    `json:"deposit" yaml:"deposit"`
	UpgradeHeight   int64        `json:"upgrade_height" yaml:"upgrade_height"`
	UpgradeUniqueID string       `json:"upgrade_unique_id" yaml:"upgrade_unique_id"`
}

func ProposalRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "oracle",
		Handler:  newPostPlanHandler(clientCtx),
	}
}

func newPostPlanHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PlanRequest

		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		plan := types.Plan{
			UniqueId: req.UpgradeUniqueID,
			Height:   req.UpgradeHeight,
		}
		content := types.NewOracleUpgradeProposal(req.Title, req.Description, plan)
		msg, err := govtypes.NewMsgSubmitProposal(content, req.Deposit, fromAddr)
		if rest.CheckBadRequestError(w, err) {
			return
		}
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}
