package app

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// SimulateRequest represents attributes of a tx that will be simulated
type SimulateRequest struct {
	Msgs []sdk.Msg   `json:"msgs"`
	Fee  auth.StdFee `json:"fee"`
	Memo string      `json:"memo"`
}

// RegisterSimulateRoutes registers a tx simulate route to a mux router with
// a provided cli context
func RegisterSimulateRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/tx/simulate", postAppSimulateHandlerFn(cliCtx)).Methods("POST")
}

// postAppSimulateHandlerFn handles tx simulate requests and returns the height and
// output of the simulation
func postAppSimulateHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SimulateRequest
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		tx := auth.NewStdTx(
			req.Msgs,
			req.Fee,
			[]auth.StdSignature{{}},
			req.Memo,
		)

		txBz, err := cliCtx.Codec.MarshalBinaryLengthPrefixed(tx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		bz, height, err := cliCtx.QueryWithData("/app/simulate", txBz)
		cliCtx = cliCtx.WithHeight(height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var simRes sdk.SimulationResponse
		if err := cliCtx.Codec.UnmarshalBinaryBare(bz, &simRes); err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, simRes)
	}
}
