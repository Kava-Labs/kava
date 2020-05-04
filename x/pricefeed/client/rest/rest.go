package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

const (
	RestMarketID = "market_id"
)

// PostPriceReq defines the properties of a PostPrice request's body.
type PostPriceReq struct {
	BaseReq  rest.BaseReq `json:"base_req"`
	MarketID string       `json:"market_id"`
	Price    string       `json:"price"`
	Expiry   string       `json:"expiry"`
}

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r)
}
