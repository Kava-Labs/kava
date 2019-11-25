package types

import(
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

const (
	QueryGetOutstandingDebt = "outstanding_debt" // Get the outstanding seized debt
)

type SeizeAndStartCollateralAuctionRequest struct {
	BaseReq         rest.BaseReq   `json:"base_req"`
	Sender          sdk.AccAddress `json:"sender"`
	CdpOwner        sdk.AccAddress `json:"cdp_owner"`
	CollateralDenom string         `json:"collateral_denom"`
}

type StartDebtAuctionRequest struct {
	BaseReq rest.BaseReq   `json:"base_req"`
	Sender  sdk.AccAddress `json:"sender"` // TODO use baseReq.From instead?
}