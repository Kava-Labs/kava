package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/gorilla/mux"

	"github.com/kava-labs/kava/x/issuance/types"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/issue", types.ModuleName), postIssueTokensHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/redeem", types.ModuleName), postRedeemTokensHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/block", types.ModuleName), postBlockAddressHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/unblock", types.ModuleName), postUnblockAddressHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/pause", types.ModuleName), postPauseHandlerFn(cliCtx)).Methods("POST")
}

func postIssueTokensHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody PostIssueReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &requestBody) {
			return
		}

		requestBody.BaseReq = requestBody.BaseReq.Sanitize()
		if !requestBody.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(requestBody.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgIssueTokens(
			fromAddr,
			requestBody.Tokens,
			requestBody.Receiver,
		)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, requestBody.BaseReq, []sdk.Msg{msg})
	}
}

func postRedeemTokensHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody PostRedeemReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &requestBody) {
			return
		}

		requestBody.BaseReq = requestBody.BaseReq.Sanitize()
		if !requestBody.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(requestBody.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgRedeemTokens(
			fromAddr,
			requestBody.Tokens,
		)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, requestBody.BaseReq, []sdk.Msg{msg})
	}
}

func postBlockAddressHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody PostBlockAddressReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &requestBody) {
			return
		}

		requestBody.BaseReq = requestBody.BaseReq.Sanitize()
		if !requestBody.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(requestBody.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgBlockAddress(
			fromAddr,
			requestBody.Denom,
			requestBody.Address,
		)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, requestBody.BaseReq, []sdk.Msg{msg})
	}
}

func postUnblockAddressHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody PostUnblockAddressReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &requestBody) {
			return
		}

		requestBody.BaseReq = requestBody.BaseReq.Sanitize()
		if !requestBody.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(requestBody.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgUnblockAddress(
			fromAddr,
			requestBody.Denom,
			requestBody.Address,
		)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, requestBody.BaseReq, []sdk.Msg{msg})
	}
}

func postPauseHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody PostPauseReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &requestBody) {
			return
		}

		requestBody.BaseReq = requestBody.BaseReq.Sanitize()
		if !requestBody.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(requestBody.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgSetPauseStatus(
			fromAddr,
			requestBody.Denom,
			requestBody.Status,
		)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, requestBody.BaseReq, []sdk.Msg{msg})
	}
}
