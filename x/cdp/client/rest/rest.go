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

// PostCdpReq defines the properties of cdp request's body.
type PostCdpReq struct {
	BaseReq    rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Owner      sdk.AccAddress `json:"owner" yaml:"owner"`
	Collateral sdk.Coins      `json:"collateral" yaml:"collateral"`
	Principal  sdk.Coins      `json:"principal" yaml:"principal"`
}

// PostDepositReq defines the properties of cdp request's body.
type PostDepositReq struct {
	BaseReq    rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Owner      sdk.AccAddress `json:"owner" yaml:"owner"`
	Depositor  sdk.AccAddress `json:"depositor" yaml:"depositor"`
	Collateral sdk.Coins      `json:"collateral" yaml:"collateral"`
}

// PostWithdrawalReq defines the properties of cdp request's body.
type PostWithdrawalReq struct {
	BaseReq    rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Owner      sdk.AccAddress `json:"owner" yaml:"owner"`
	Depositor  sdk.AccAddress `json:"depositor" yaml:"depositor"`
	Collateral sdk.Coins      `json:"collateral" yaml:"collateral"`
}

// PostDrawReq defines the properties of cdp request's body.
type PostDrawReq struct {
	BaseReq   rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Owner     sdk.AccAddress `json:"owner" yaml:"owner"`
	Denom     string         `json:"denom" yaml:"denom"`
	Principal sdk.Coins      `json:"principal" yaml:"principal"`
}

// PostRepayReq defines the properties of cdp request's body.
type PostRepayReq struct {
	BaseReq rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Owner   sdk.AccAddress `json:"owner" yaml:"owner"`
	Denom   string         `json:"denom" yaml:"denom"`
	Payment sdk.Coins      `json:"payment" yaml:"payment"`
}
