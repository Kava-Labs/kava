package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	// r.HandleFunc("/bep3", postHTLTHandlerFn(cliCtx)).Methods("POST")

}
