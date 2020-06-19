package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Querier routes for the cdp module
const (
	QueryGetCdp                     = "cdp"
	QueryGetCdpDeposits             = "deposits"
	QueryGetCdps                    = "cdps"
	QueryGetCdpsByCollateralization = "ratio"
	QueryGetParams                  = "params"
	QueryTotalSupply                = "total-supply"
	QueryGetAccounts                = "accounts"
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

// QueryCdpParams params for query /cdp/cdp
type QueryCdpParams struct {
	CollateralDenom string         // get CDPs with this collateral denom
	Owner           sdk.AccAddress // get CDPs belonging to this owner
}

// NewQueryCdpParams returns QueryCdpParams
func NewQueryCdpParams(owner sdk.AccAddress, denom string) QueryCdpParams {
	return QueryCdpParams{
		Owner:           owner,
		CollateralDenom: denom,
	}
}

// QueryCdpDeposits params for query /cdp/deposits
type QueryCdpDeposits struct {
	CollateralDenom string         // get CDPs with this collateral denom
	Owner           sdk.AccAddress // get CDPs belonging to this owner
}

// NewQueryCdpDeposits returns QueryCdpDeposits
func NewQueryCdpDeposits(owner sdk.AccAddress, denom string) QueryCdpDeposits {
	return QueryCdpDeposits{
		Owner:           owner,
		CollateralDenom: denom,
	}
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

type QueryGetAccountsResponse struct {
	Cdp         sdk.AccAddress `json:"cdp" yaml:"cdp"`
	Liquidator  sdk.AccAddress `json:"liquidator" yaml:"liquidator"`
	SavingsRate sdk.AccAddress `json:"savings_rate" yaml:"savings_rate"`
}
