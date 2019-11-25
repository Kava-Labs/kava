package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/kava-labs/kava/x/cdp/types"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/cdps", modifyCdpHandlerFn(cliCtx)).Methods("PUT")
}

func modifyCdpHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode PUT request body
		var requestBody types.ModifyCdpRequestBody
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &requestBody) {
			return
		}
		requestBody.BaseReq = requestBody.BaseReq.Sanitize()
		if !requestBody.BaseReq.ValidateBasic(w) {
			return
		}

		// Get the stored CDP
		querierParams := types.QueryCdpsParams{
			Owner:           requestBody.Cdp.Owner,
			CollateralDenom: requestBody.Cdp.CollateralDenom,
		}
		querierParamsBz, err := cliCtx.Codec.MarshalJSON(querierParams)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/cdp/%s", types.QueryGetCdps), querierParamsBz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		cliCtx = cliCtx.WithHeight(height)
		var cdps types.CDPs
		err = cliCtx.Codec.UnmarshalJSON(res, &cdps)
		if len(cdps) != 1 || err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Calculate CDP updates
		collateralDelta := requestBody.Cdp.CollateralAmount.Sub(cdps[0].CollateralAmount)
		debtDelta := requestBody.Cdp.Debt.Sub(cdps[0].Debt)

		// Create and return msg
		msg := types.NewMsgCreateOrModifyCDP(
			requestBody.Cdp.Owner,
			requestBody.Cdp.CollateralDenom,
			collateralDelta,
			debtDelta,
		)
		utils.WriteGenerateStdTxResponse(w, cliCtx, requestBody.BaseReq, []sdk.Msg{msg})
	}
}
