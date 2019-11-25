package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/kava-labs/kava/x/liquidator/types"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/liquidator/outstandingdebt", queryDebtHandlerFn(cliCtx)).Methods("GET")
	// r.HandleFunc("liquidator/burn", surplusAuctionHandlerFn(cdc, cliCtx).Methods("POST"))
}

func queryDebtHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/liquidator/%s", types.QueryGetOutstandingDebt), nil)
		cliCtx = cliCtx.WithHeight(height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res) // write JSON to response writer
	}
}
