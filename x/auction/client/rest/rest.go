package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

// REST Variable names
// nolint
const (
	RestType  = "type"
	RestOwner = "owner"
	RestDenom = "denom"
	RestPhase = "phase"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r)
}

// PlaceBidReq defines the properties of a bid request's body
type PlaceBidReq struct {
	BaseReq rest.BaseReq `json:"base_req"`
	Amount  sdk.Coin     `json:"amount"`
}
