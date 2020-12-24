package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

// REST Variable names
// nolint
const (
	RestOwner          = "owner"
	RestCollateralType = "collateral-type"
	RestID             = "id"
	RestRatio          = "ratio"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r)
}

// PostCdpReq defines the properties of cdp request's body.
type PostCdpReq struct {
	BaseReq        rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Sender         sdk.AccAddress `json:"sender" yaml:"sender"`
	Collateral     sdk.Coin       `json:"collateral" yaml:"collateral"`
	CollateralType string         `json:"collateral_type" yaml:"collateral_type"`
	Principal      sdk.Coin       `json:"principal" yaml:"principal"`
}

// PostDepositReq defines the properties of cdp request's body.
type PostDepositReq struct {
	BaseReq        rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Owner          sdk.AccAddress `json:"owner" yaml:"owner"`
	Depositor      sdk.AccAddress `json:"depositor" yaml:"depositor"`
	Collateral     sdk.Coin       `json:"collateral" yaml:"collateral"`
	CollateralType string         `json:"collateral_type" yaml:"collateral_type"`
}

// PostWithdrawalReq defines the properties of cdp request's body.
type PostWithdrawalReq struct {
	BaseReq        rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Owner          sdk.AccAddress `json:"owner" yaml:"owner"`
	Depositor      sdk.AccAddress `json:"depositor" yaml:"depositor"`
	Collateral     sdk.Coin       `json:"collateral" yaml:"collateral"`
	CollateralType string         `json:"collateral_type" yaml:"collateral_type"`
}

// PostDrawReq defines the properties of cdp request's body.
type PostDrawReq struct {
	BaseReq        rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Owner          sdk.AccAddress `json:"owner" yaml:"owner"`
	CollateralType string         `json:"collateral_type" yaml:"collateral_type"`
	Principal      sdk.Coin       `json:"principal" yaml:"principal"`
}

// PostRepayReq defines the properties of cdp request's body.
type PostRepayReq struct {
	BaseReq        rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Owner          sdk.AccAddress `json:"owner" yaml:"owner"`
	CollateralType string         `json:"collateral_type" yaml:"collateral_type"`
	Payment        sdk.Coin       `json:"payment" yaml:"payment"`
}

// PostLiquidateReq defines the properties of cdp liquidation request's body.
type PostLiquidateReq struct {
	BaseReq        rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Owner          sdk.AccAddress `json:"owner" yaml:"owner"`
	CollateralType string         `json:"collateral_type" yaml:"collateral_type"`
}
