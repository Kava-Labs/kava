package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	"github.com/kava-labs/kava/x/issuance/types"
)

func registerTxRoutes(cliCtx client.Context, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/issue", types.ModuleName), postIssueTokensHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/redeem", types.ModuleName), postRedeemTokensHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/block", types.ModuleName), postBlockAddressHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/unblock", types.ModuleName), postUnblockAddressHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/pause", types.ModuleName), postPauseHandlerFn(cliCtx)).Methods("POST")
}

func postIssueTokensHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody PostIssueReq
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &requestBody) {
			return
		}

		baseReq := requestBody.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(requestBody.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgIssueTokens(
			fromAddr.String(),
			requestBody.Tokens,
			requestBody.Receiver.String(),
		)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, baseReq, msg)
	}
}

func postRedeemTokensHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody PostRedeemReq
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &requestBody) {
			return
		}

		baseReq := requestBody.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(requestBody.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgRedeemTokens(
			fromAddr.String(),
			requestBody.Tokens,
		)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, baseReq, msg)
	}
}

func postBlockAddressHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody PostBlockAddressReq
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &requestBody) {
			return
		}

		baseReq := requestBody.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(requestBody.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgBlockAddress(
			fromAddr.String(),
			requestBody.Denom,
			requestBody.Address.String(),
		)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, baseReq, msg)
	}
}

func postUnblockAddressHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody PostUnblockAddressReq
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &requestBody) {
			return
		}

		baseReq := requestBody.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(requestBody.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgUnblockAddress(
			fromAddr.String(),
			requestBody.Denom,
			requestBody.Address.String(),
		)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, baseReq, msg)
	}
}

func postPauseHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody PostPauseReq
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &requestBody) {
			return
		}

		baseReq := requestBody.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(requestBody.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgSetPauseStatus(
			fromAddr.String(),
			requestBody.Denom,
			requestBody.Status,
		)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, baseReq, msg)
	}
}
