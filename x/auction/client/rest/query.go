package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

// r.HandleFunc(fmt.Sprintf("/auction/bid/{%s}/{%s}/{%s}/{%s}", restAuctionID, restBidder, restBid, restLot), bidHandlerFn(cdc, cliCtx)).Methods("PUT")

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/auction/getauctions"), queryGetAuctionsHandlerFn(cliCtx)).Methods("GET")

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
