package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client"

	"github.com/kava-labs/kava/x/incentive/types"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/incentive/claim", postClaimHandlerFn(cliCtx)).Methods("POST")

}

func postClaimHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody types.PostClaimReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &requestBody) {
			return
		}
		requestBody.BaseReq = requestBody.BaseReq.Sanitize()
		if !requestBody.BaseReq.ValidateBasic(w) {
			return
		}
		msg := types.NewMsgClaimReward(requestBody.Sender, requestBody.Denom)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, requestBody.BaseReq, []sdk.Msg{msg})
	}
}
