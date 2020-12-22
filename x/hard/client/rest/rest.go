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
	RestOwner       = "owner"
	RestDenom       = "deposit-denom"
	RestType        = "deposit-type"
	RestBorrowDenom = "borrow-denom"
	RestName        = "name"
)

// RegisterRoutes registers hard-related REST handlers to a router
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r)
}

// PostCreateDepositReq defines the properties of a deposit create request's body
type PostCreateDepositReq struct {
	BaseReq rest.BaseReq   `json:"base_req" yaml:"base_req"`
	From    sdk.AccAddress `json:"from" yaml:"from"`
	Amount  sdk.Coins      `json:"amount" yaml:"amount"`
}

// PostCreateWithdrawReq defines the properties of a deposit withdraw request's body
type PostCreateWithdrawReq struct {
	BaseReq rest.BaseReq   `json:"base_req" yaml:"base_req"`
	From    sdk.AccAddress `json:"from" yaml:"from"`
	Amount  sdk.Coins      `json:"amount" yaml:"amount"`
}

// PostClaimReq defines the properties of a claim reward request's body
type PostClaimReq struct {
	BaseReq        rest.BaseReq   `json:"base_req" yaml:"base_req"`
	From           sdk.AccAddress `json:"from" yaml:"from"`
	Receiver       sdk.AccAddress `json:"receiver" yaml:"receiver"`
	DepositDenom   string         `json:"deposit_denom" yaml:"deposit_denom"`
	MultiplierName string         `json:"multiplier_name" yaml:"multiplier_name"`
	ClaimType      string         `json:"claim_type" yaml:"claim_type"`
}
