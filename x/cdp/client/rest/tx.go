package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/kava-labs/kava/x/cdp/types"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/cdp", postCdpHandlerFn(cliCtx)).Methods("PUT")
	r.HandleFunc("/cdp/{owner}/{denom}/deposits", postDepositHandlerFn(cliCtx)).Methods("PUT")
	r.HandleFunc("/cdp/{owner}/{denom}/withdraw", postWithdrawHandlerFn(cliCtx)).Methods("PUT")
	r.HandleFunc("/cdp/{owner}/{denom}/draw", postDrawHandlerFn(cliCtx)).Methods("PUT")
	r.HandleFunc("/cdp/{owner}/{denom}/repay", postRepayHandlerFn(cliCtx)).Methods("PUT")

}

func postCdpHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode PUT request body
		var requestBody PostCdpReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &requestBody) {
			return
		}
		requestBody.BaseReq = requestBody.BaseReq.Sanitize()
		if !requestBody.BaseReq.ValidateBasic(w) {
			return
		}

		// Create and return msg
		msg := types.NewMsgCreateCDP(
			requestBody.Owner,
			requestBody.Collateral,
			requestBody.Principal,
		)
		utils.WriteGenerateStdTxResponse(w, cliCtx, requestBody.BaseReq, []sdk.Msg{msg})
	}
}

func postDepositHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode PUT request body
		var requestBody PostDepositReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &requestBody) {
			return
		}
		requestBody.BaseReq = requestBody.BaseReq.Sanitize()
		if !requestBody.BaseReq.ValidateBasic(w) {
			return
		}

		// Create and return msg
		msg := types.NewMsgDeposit(
			requestBody.Owner,
			requestBody.Depositor,
			requestBody.Collateral,
		)
		utils.WriteGenerateStdTxResponse(w, cliCtx, requestBody.BaseReq, []sdk.Msg{msg})
	}
}

func postWithdrawHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode PUT request body
		var requestBody PostWithdrawalReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &requestBody) {
			return
		}
		requestBody.BaseReq = requestBody.BaseReq.Sanitize()
		if !requestBody.BaseReq.ValidateBasic(w) {
			return
		}

		// Create and return msg
		msg := types.NewMsgWithdraw(
			requestBody.Owner,
			requestBody.Depositor,
			requestBody.Collateral,
		)
		utils.WriteGenerateStdTxResponse(w, cliCtx, requestBody.BaseReq, []sdk.Msg{msg})
	}
}

func postDrawHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode PUT request body
		var requestBody PostDrawReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &requestBody) {
			return
		}
		requestBody.BaseReq = requestBody.BaseReq.Sanitize()
		if !requestBody.BaseReq.ValidateBasic(w) {
			return
		}

		// Create and return msg
		msg := types.NewMsgDrawDebt(
			requestBody.Owner,
			requestBody.Denom,
			requestBody.Principal,
		)
		utils.WriteGenerateStdTxResponse(w, cliCtx, requestBody.BaseReq, []sdk.Msg{msg})
	}
}

func postRepayHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode PUT request body
		var requestBody PostRepayReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &requestBody) {
			return
		}
		requestBody.BaseReq = requestBody.BaseReq.Sanitize()
		if !requestBody.BaseReq.ValidateBasic(w) {
			return
		}

		// Create and return msg
		msg := types.NewMsgRepayDebt(
			requestBody.Owner,
			requestBody.Denom,
			requestBody.Payment,
		)
		utils.WriteGenerateStdTxResponse(w, cliCtx, requestBody.BaseReq, []sdk.Msg{msg})
	}
}
