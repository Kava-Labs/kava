package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
	"github.com/kava-labs/kava/x/validator-vesting/internal/types"
)

// define routes that get registered by the main application
func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {

	r.HandleFunc("/vesting/circulatingsupply", getCirculatingSupplyHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/vesting/totalsupply", getTotalSupplyHandlerFn(cliCtx)).Methods("GET")

}

func getTotalSupplyHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewBaseQueryParams(page, limit)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryTotalSupply)
		res, height, err := cliCtx.QueryWithData(route, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		// directly write output instead of putting in json
		w.Write(res)
		// rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getCirculatingSupplyHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewBaseQueryParams(page, limit)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryCirculatingSupply)
		res, height, err := cliCtx.QueryWithData(route, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		// directly write output instead of putting in json
		w.Write(res)
		// rest.PostProcessResponse(w, cliCtx, res)
	}

}
