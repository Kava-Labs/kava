package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

// define routes that get registered by the main application
func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/rawprices/{%s}", types.ModuleName, restName), queryRawPricesHandler(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/currentprice/{%s}", types.ModuleName, restName), queryCurrentPriceHandler(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/markets", types.ModuleName), queryMarketsHandler(cliCtx)).Methods("GET")
}

func queryRawPricesHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramType := vars[restName]
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/rawprices/%s", paramType), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryCurrentPriceHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramType := vars[restName]
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/price/%s", paramType), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryMarketsHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/markets/"), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
