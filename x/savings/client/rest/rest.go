package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r)
}

// PostCreateDepositReq defines the properties of a deposit create request's body
type PostCreateDepositReq struct {
	BaseReq rest.BaseReq   `json:"base_req" yaml:"base_req"`
	From    sdk.AccAddress `json:"from" yaml:"from"`
	Amount  sdk.Coins      `json:"amount" yaml:"amount"`
}
