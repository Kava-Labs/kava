package rest

import (
	"github.com/cosmos/cosmos-sdk/client/context"

	"github.com/gorilla/mux"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/validator-vesting/circulatingsupply", getCircuatingSupplyHandlerFn(cliCtx)).Methods("GET")

}
