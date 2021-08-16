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
	r.HandleFunc("/incentive/claim-cdp", postClaimHandlerFn(cliCtx, usdxMintingGenerator)).Methods("POST")
	r.HandleFunc("/incentive/claim-cdp-vesting", postClaimVVestingHandlerFn(cliCtx, usdxMintingVVGenerator)).Methods("POST")

	r.HandleFunc("/incentive/claim-hard", postClaimHandlerFn(cliCtx, hardGenerator)).Methods("POST")
	r.HandleFunc("/incentive/claim-hard-vesting", postClaimVVestingHandlerFn(cliCtx, hardVVGenerator)).Methods("POST")

	r.HandleFunc("/incentive/claim-delegator", postClaimHandlerFn(cliCtx, delegatorGenerator)).Methods("POST")
	r.HandleFunc("/incentive/claim-delegator-vesting", postClaimVVestingHandlerFn(cliCtx, delegatorVVGenerator)).Methods("POST")

	r.HandleFunc("/incentive/claim-swap", postClaimHandlerFn(cliCtx, swapGenerator)).Methods("POST")
	r.HandleFunc("/incentive/claim-swap-vesting", postClaimVVestingHandlerFn(cliCtx, swapVVGenerator)).Methods("POST")
}

func usdxMintingGenerator(req PostClaimReq) (sdk.Msg, error) {
	if len(req.DenomsToClaim) != 1 {
		return nil, fmt.Errorf("must only claim %s denom for usdx minting rewards, got '%s", types.USDXMintingRewardDenom, req.DenomsToClaim)
	}
	return types.NewMsgClaimUSDXMintingReward(req.Sender, req.DenomsToClaim[0].MultiplierName), nil
}
func hardGenerator(req PostClaimReq) (sdk.Msg, error) {
	return types.NewMsgClaimHardReward(req.Sender, req.DenomsToClaim...), nil
}
func delegatorGenerator(req PostClaimReq) (sdk.Msg, error) {
	return types.NewMsgClaimDelegatorReward(req.Sender, req.DenomsToClaim...), nil
}
func swapGenerator(req PostClaimReq) (sdk.Msg, error) {
	return types.NewMsgClaimSwapReward(req.Sender, req.DenomsToClaim...), nil
}

func postClaimHandlerFn(cliCtx context.CLIContext, msgGenerator func(req PostClaimReq) (sdk.Msg, error)) http.HandlerFunc {
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

		msg, err := msgGenerator(requestBody)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, requestBody.BaseReq, []sdk.Msg{msg})
	}
}

func usdxMintingVVGenerator(req PostClaimVVestingReq) (sdk.Msg, error) {
	if len(req.DenomsToClaim) != 1 {
		return nil, fmt.Errorf("must only claim %s denom for usdx minting rewards, got '%s", types.USDXMintingRewardDenom, req.DenomsToClaim)
	}
	return types.NewMsgClaimUSDXMintingRewardVVesting(req.Sender, req.Receiver, req.DenomsToClaim[0].MultiplierName), nil
}
func hardVVGenerator(req PostClaimVVestingReq) (sdk.Msg, error) {
	return types.NewMsgClaimHardRewardVVesting(req.Sender, req.Receiver, req.DenomsToClaim...), nil
}
func delegatorVVGenerator(req PostClaimVVestingReq) (sdk.Msg, error) {
	return types.NewMsgClaimDelegatorRewardVVesting(req.Sender, req.Receiver, req.DenomsToClaim...), nil
}
func swapVVGenerator(req PostClaimVVestingReq) (sdk.Msg, error) {
	return types.NewMsgClaimSwapRewardVVesting(req.Sender, req.Receiver, req.DenomsToClaim...), nil
}

func postClaimVVestingHandlerFn(cliCtx context.CLIContext, msgGenerator func(req PostClaimVVestingReq) (sdk.Msg, error)) http.HandlerFunc {
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

		msg, err := msgGenerator(requestBody)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, requestBody.BaseReq, []sdk.Msg{msg})
	}
}
