package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	clientrest "github.com/cosmos/cosmos-sdk/client/rest"
)

// RegisterRoutes registers kavadist-related REST handlers to a router
func RegisterRoutes(cliCtx client.Context, rtr *mux.Router) {
	r := clientrest.WithHTTPDeprecationHeaders(rtr)
	registerQueryRoutes(cliCtx, r)
}
