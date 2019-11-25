package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

const (
	QueryGetCdps              = "cdps"
	QueryGetParams            = "params"
	RestOwner                 = "owner"
	RestCollateralDenom       = "collateralDenom"
	RestUnderCollateralizedAt = "underCollateralizedAt"
)

type QueryCdpsParams struct {
	CollateralDenom       string         // get CDPs with this collateral denom
	Owner                 sdk.AccAddress // get CDPs belonging to this owner
	UnderCollateralizedAt sdk.Dec        // get CDPs that will be below the liquidation ratio when the collateral is at this price.
}

type ModifyCdpRequestBody struct {
	BaseReq rest.BaseReq `json:"base_req"`
	Cdp     CDP          `json:"cdp"`
}
