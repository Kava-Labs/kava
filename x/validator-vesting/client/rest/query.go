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

}

func getCirculatingSupplyHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the query height
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.Query(fmt.Sprintf("vesting/circulatingsupply", types.QueryCirculatingSupply))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		// do not write as json, write direct output
		// rest.PostProcessResponse(w, cliCtx, res)

		w.Write(res) // write the result directly
	}

}
