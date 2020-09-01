package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r)
}

// PostIssueReq defines the properties of an issue token request's body
type PostIssueReq struct {
	BaseReq  rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Tokens   sdk.Coin       `json:"tokens" yaml:"tokens"`
	Receiver sdk.AccAddress `json:"receiver" yaml:"receiver"`
}

// PostRedeemReq defines the properties of a redeem token request's body
type PostRedeemReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Tokens  sdk.Coin     `json:"tokens" yaml:"tokens"`
}

// PostBlockAddressReq defines the properties of a block address request's body
type PostBlockAddressReq struct {
	BaseReq rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Address sdk.AccAddress `json:"blocked_address" yaml:"blocked_address"`
	Denom   string         `json:"denom" yaml:"denom"`
}

// PostUnblockAddressReq defines the properties of a unblock address request's body
type PostUnblockAddressReq struct {
	BaseReq rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Address sdk.AccAddress `json:"blocked_address" yaml:"blocked_address"`
	Denom   string         `json:"denom" yaml:"denom"`
}

// PostPauseReq defines the properties of a pause request's body
type PostPauseReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Denom   string       `json:"denom" yaml:"denom"`
	Status  bool         `json:"status" yaml:"status"`
}
