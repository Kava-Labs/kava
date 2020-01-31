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
	r.HandleFunc(fmt.Sprintf("/%s/rawprices/{%s}", types.ModuleName, RestMarketID), queryRawPricesHandler(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/price/{%s}", types.ModuleName, RestMarketID), queryPriceHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/markets", types.ModuleName), queryMarketsHandler(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/parameters", types.ModuleName), queryParamsHandlerFn(cliCtx)).Methods("GET")
}

func queryRawPricesHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramMarketID := vars[RestMarketID]
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/rawprices/%s", types.ModuleName, paramMarketID), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryPriceHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramMarketID := vars[RestMarketID]

		queryPriceParams := types.NewQueryPriceParams(paramMarketID)

		bz, err := cliCtx.Codec.MarshalJSON(queryPriceParams)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryPrice), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryMarketsHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryMarkets), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryParamsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the params
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryGetParams), nil)
		cliCtx = cliCtx.WithHeight(height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		// Return the params
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
