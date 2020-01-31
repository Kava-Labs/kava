package rest

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/kava-labs/kava/x/pricefeed/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/rawprices", types.ModuleName), postPriceHandler(cliCtx)).Methods("PUT")
}

func postPriceHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PostPriceReq

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		addr, err := sdk.AccAddressFromBech32(baseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		price, err := sdk.NewDecFromStr(req.Price)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		expiryInt, ok := sdk.NewIntFromString(req.Expiry)
		if !ok {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "invalid expiry")
			return
		}
		expiry := tmtime.Canonical(time.Unix(expiryInt.Int64(), 0))

		// create the message
		msg := types.NewMsgPostPrice(addr, req.MarketID, price, expiry)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}

}
