package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	// r.HandleFunc("/bep3", postHTLTHandlerFn(cliCtx)).Methods("POST")

}

// Action TX body
// type <Action>Req struct {
// 	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
//
// }
