package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

// REST variable names
// nolint
const (
	RestDenom = "denom"
)

// RegisterRoutes registers incentive-related REST handlers to a router
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r)
}

// PostClaimReq defines the properties of claim transaction's request body.
type PostClaimReq struct {
	BaseReq        rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Sender         sdk.AccAddress `json:"sender" yaml:"sender"`
	MultiplierName string         `json:"multiplier_name" yaml:"multiplier_name"`
	DenomsToClaim  []string       `json:"denoms_to_claim" yaml:"denoms_to_claim"`
}

// PostClaimReq defines the properties of claim transaction's request body.
type PostClaimVVestingReq struct {
	BaseReq        rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Sender         sdk.AccAddress `json:"sender" yaml:"sender"`
	Receiver       sdk.AccAddress `json:"receiver" yaml:"receiver"`
	MultiplierName string         `json:"multiplier_name" yaml:"multiplier_name"`
}
