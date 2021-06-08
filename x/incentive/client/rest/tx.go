package rest

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/kava-labs/kava/x/incentive/types"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/incentive/claim-cdp", postClaimCdpHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc("/incentive/claim-cdp-vesting", postClaimCdpVVestingHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc("/incentive/claim-hard", postClaimHardHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc("/incentive/claim-hard-vesting", postClaimHardVVestingHandlerFn(cliCtx)).Methods("POST")

}

func postClaimCdpHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody PostClaimReq
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

		if !bytes.Equal(fromAddr, requestBody.Sender) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, fmt.Sprintf("expected: %s, got: %s", fromAddr, requestBody.Sender))
			return
		}

		msg := types.NewMsgClaimUSDXMintingReward(requestBody.Sender, requestBody.MultiplierName)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, requestBody.BaseReq, []sdk.Msg{msg})
	}
}

func postClaimCdpVVestingHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody PostClaimVVestingReq
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

		if !bytes.Equal(fromAddr, requestBody.Sender) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, fmt.Sprintf("expected: %s, got: %s", fromAddr, requestBody.Sender))
			return
		}

		msg := types.NewMsgClaimUSDXMintingRewardVVesting(requestBody.Sender, requestBody.Receiver, requestBody.MultiplierName)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, requestBody.BaseReq, []sdk.Msg{msg})
	}
}

func postClaimHardHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody PostClaimReq
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

		if !bytes.Equal(fromAddr, requestBody.Sender) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, fmt.Sprintf("expected: %s, got: %s", fromAddr, requestBody.Sender))
			return
		}

		msg := types.NewMsgClaimHardReward(requestBody.Sender, requestBody.MultiplierName)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, requestBody.BaseReq, []sdk.Msg{msg})
	}
}

func postClaimHardVVestingHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody PostClaimVVestingReq
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

		if !bytes.Equal(fromAddr, requestBody.Sender) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, fmt.Sprintf("expected: %s, got: %s", fromAddr, requestBody.Sender))
			return
		}

		msg := types.NewMsgClaimHardRewardVVesting(requestBody.Sender, requestBody.Receiver, requestBody.MultiplierName)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, requestBody.BaseReq, []sdk.Msg{msg})
	}
}
