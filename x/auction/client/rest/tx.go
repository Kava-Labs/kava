package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	"github.com/kava-labs/kava/x/auction/types"
)

func registerTxRoutes(cliCtx client.Context, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/auctions/{%s}/bids", types.ModuleName, restAuctionID), postPlaceBidHandlerFn(cliCtx)).Methods("POST")
}

func postPlaceBidHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var requestBody PlaceBidReq
		if !rest.ReadRESTReq(w, req, cliCtx.LegacyAmino, &requestBody) {
			return
		}

		baseReq := requestBody.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		bidderAddr, err := sdk.AccAddressFromBech32(requestBody.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Get auction ID from url
		auctionID, ok := rest.ParseUint64OrReturnBadRequest(w, mux.Vars(req)[restAuctionID])
		if !ok {
			return
		}

		// Create and return a StdTx
		msg := types.NewMsgPlaceBid(auctionID, bidderAddr.String(), requestBody.Amount)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, baseReq, &msg)
	}
}
