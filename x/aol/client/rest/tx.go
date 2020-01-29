package rest

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/medibloc/panacea-core/x/aol/internal/types"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
)

type createTopicReq struct {
	BaseReq     rest.BaseReq `json:"base_req"`
	TopicName        string  `json:"topic_name"`
	Description string       `json:"topic_description"`
	Owner       string       `json:"owner"`
}

func createTopicHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createTopicReq

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		addr, err := sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		// create the message
		msg := types.NewMsgCreateTopic(req.TopicName, req.Description, addr)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

type addWriterReq struct {
	BaseReq     rest.BaseReq `json:"base_req"`
	TopicName   string       `json:"topic_name"`
	Key         string       `json:"key"`
	Value       string       `json:"value"`
	Writer      string       `json:"writer_address"`
	Owner       string       `json:"owner_address"`
	FeePayer    string       `json:"fee_payer_address"`
}

func addWriterHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req addWriterReq

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		key := []byte(req.Key)
		value := []byte(req.Value)

		writerAddr, err := sdk.AccAddressFromBech32(req.Writer)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		ownerAddr, err := sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		feePayerAddr, err := sdk.AccAddressFromBech32(req.FeePayer)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		// create the message
		msg := types.NewMsgAddRecord(req.TopicName, key, value, writerAddr, ownerAddr, feePayerAddr)

		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

type deleteWriterReq struct {
	BaseReq     rest.BaseReq `json:"base_req"`
	TopicName   string       `json:"topic_name"`
	Writer      string       `json:"writer_address"`
	Owner       string       `json:"owner_address"`
}

func deleteWriterHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req deleteWriterReq

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		writerAddr, err := sdk.AccAddressFromBech32(req.Writer)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		ownerAddr, err := sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		// create the message
		msg := types.NewMsgDeleteWriter(req.TopicName, writerAddr, ownerAddr)

		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

type addRecordReq struct {
	BaseReq     rest.BaseReq `json:"base_req"`
	TopicName   string       `json:"topic_name"`
	Key         string       `json:"key"`
	Value       string       `json:"value"`
	Writer      string       `json:"writer_address"`
	Owner       string       `json:"owner_address"`
	FeePayer    string       `json:"fee_payer_address"`
}

func addRecordHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req addRecordReq

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		key := []byte(req.Key)
		value := []byte(req.Value)

		writerAddr, err := sdk.AccAddressFromBech32(req.Writer)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		ownerAddr, err := sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		feePayerAddr, err := sdk.AccAddressFromBech32(req.FeePayer)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		// create the message
		msg := types.NewMsgAddRecord(req.TopicName, key, value, writerAddr, ownerAddr, feePayerAddr)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}
