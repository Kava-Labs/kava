package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Querier routes for the cdp module
const (
	QueryGetCdp                     = "cdp"
	QueryGetCdps                    = "cdps"
	QueryGetCdpsByCollateralization = "ratio"
	QueryGetParams                  = "params"
	RestOwner                       = "owner"
	RestCollateralDenom             = "collateral-denom"
	RestRatio                       = "ratio"
)

// QueryCdpsParams params for query /cdp/cdps
type QueryCdpsParams struct {
	CollateralDenom string // get CDPs with this collateral denom
}

// NewQueryCdpsParams returns QueryCdpsParams
func NewQueryCdpsParams(denom string) QueryCdpsParams {
	return QueryCdpsParams{
		CollateralDenom: denom,
	}
}

// NewQueryCdpParams returns QueryCdpParams
func NewQueryCdpParams(owner sdk.AccAddress, denom string) QueryCdpParams {
	return QueryCdpParams{
		Owner:           owner,
		CollateralDenom: denom,
	}
}

// QueryCdpParams params for query /cdp/cdp
type QueryCdpParams struct {
	CollateralDenom string         // get CDPs with this collateral denom
	Owner           sdk.AccAddress // get CDPs belonging to this owner
}

// QueryCdpsByRatioParams params for query /cdp/cdps/ratio
type QueryCdpsByRatioParams struct {
	CollateralDenom string  // get CDPs with this collateral denom
	Ratio           sdk.Dec // get CDPs below this collateral:debt ratio
}

// NewQueryCdpsByRatioParams returns QueryCdpsByRatioParams
func NewQueryCdpsByRatioParams(denom string, ratio sdk.Dec) QueryCdpsByRatioParams {
	return QueryCdpsByRatioParams{
		CollateralDenom: denom,
		Ratio:           ratio,
	}
}
