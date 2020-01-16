package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/kava-labs/kava/x/auction/types"
)

// r.HandleFunc(fmt.Sprintf("/auction/bid/{%s}/{%s}/{%s}/{%s}", restAuctionID, restBidder, restBid, restLot), bidHandlerFn(cdc, cliCtx)).Methods("PUT")

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/auction/getauctions"), queryGetAuctionsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/auction/params", getParamsHandlerFn(cliCtx)).Methods("GET")
}

func queryGetAuctionsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, height, err := cliCtx.QueryWithData("/custom/auction/getauctions", nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getParamsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the params
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/auction/%s", types.QueryGetParams), nil)
		cliCtx = cliCtx.WithHeight(height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		// Return the params
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
