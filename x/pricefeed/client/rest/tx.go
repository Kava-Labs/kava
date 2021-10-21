package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/gorilla/mux"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/postprice", types.ModuleName), postPriceHandlerFn(cliCtx)).Methods("POST")

}

func postPriceHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
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

		expiryInt, err := strconv.ParseInt(req.Expiry, 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid expiry %s: %s", req.Expiry, err))
			return
		}

		if expiryInt > types.MaxExpiry {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid expiry; got %d, max: %d", expiryInt, types.MaxExpiry))
			return
		}

		expiry := tmtime.Canonical(time.Unix(expiryInt, 0))

		msg := types.NewMsgPostPrice(addr, req.MarketID, price, expiry)
		if err = msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}
