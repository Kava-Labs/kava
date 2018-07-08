package rest

import (
	"github.com/gorilla/mux"
	//"github.com/tendermint/go-crypto/keys"
	//"github.com/cosmos/cosmos-sdk/client/context"
	//"github.com/cosmos/cosmos-sdk/wire"
)

// RegisterRoutes registers paychan-related REST handlers to a router
func RegisterRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	//r.HandleFunc("/accounts/{address}/send", SendRequestHandlerFn(cdc, kb, ctx)).Methods("POST")
}

// handler functions ...
// create paychan
// close paychan
// get paychan(s)
// send paychan payment
// get balance from receiver
// get balance from local storage
// handle incoming payment
