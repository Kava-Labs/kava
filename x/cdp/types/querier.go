package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Querier routes for the cdp module
const (
	QueryGetCdp                     = "cdp"
	QueryGetCdps                    = "cdps"
	QueryGetCdpDeposits             = "deposits"
	QueryGetCdpsByCollateralization = "ratio"          // legacy query, maintained for REST API
	QueryGetCdpsByCollateralType    = "collateralType" // legacy query, maintained for REST API
	QueryGetParams                  = "params"
	QueryGetAccounts                = "accounts"
	QueryGetTotalPrincipal          = "totalPrincipal"
	QueryGetTotalCollateral         = "totalCollateral"
	RestOwner                       = "owner"
	RestCollateralType              = "collateral-type"
	RestRatio                       = "ratio"
)

// QueryCdpParams params for query /cdp/cdp
type QueryCdpParams struct {
	CollateralType string         // get CDPs with this collateral type
	Owner          sdk.AccAddress // get CDPs belonging to this owner
}

// NewQueryCdpParams returns QueryCdpParams
func NewQueryCdpParams(owner sdk.AccAddress, collateralType string) QueryCdpParams {
	return QueryCdpParams{
		Owner:          owner,
		CollateralType: collateralType,
	}
}

// QueryCdpsParams is the params for a filtered CDP query
type QueryCdpsParams struct {
	Page           int            `json:"page" yaml:"page"`
	Limit          int            `json:"limit" yaml:"limit"`
	CollateralType string         `json:"collateral_type" yaml:"collateral_type"`
	Owner          sdk.AccAddress `json:"owner" yaml:"owner"`
	ID             uint64         `json:"id" yaml:"id"`
	Ratio          sdk.Dec        `json:"ratio" yaml:"ratio"`
}

// NewQueryCdpsParams creates a new QueryCdpsParams
func NewQueryCdpsParams(page, limit int, collateralType string, owner sdk.AccAddress, id uint64, ratio sdk.Dec) QueryCdpsParams {
	return QueryCdpsParams{
		Page:           page,
		Limit:          limit,
		CollateralType: collateralType,
		Owner:          owner,
		ID:             id,
		Ratio:          ratio,
	}
}

// QueryCdpDeposits params for query /cdp/deposits
type QueryCdpDeposits struct {
	CollateralType string         // get CDPs with this collateral type
	Owner          sdk.AccAddress // get CDPs belonging to this owner
}

// NewQueryCdpDeposits returns QueryCdpDeposits
func NewQueryCdpDeposits(owner sdk.AccAddress, collateralType string) QueryCdpDeposits {
	return QueryCdpDeposits{
		Owner:          owner,
		CollateralType: collateralType,
	}
}

// QueryCdpsByCollateralTypeParams params for query /cdp/cdps/{denom}
type QueryCdpsByCollateralTypeParams struct {
	CollateralType string // get CDPs with this collateral type
}

// NewQueryCdpsByCollateralTypeParams returns QueryCdpsByCollateralTypeParams
func NewQueryCdpsByCollateralTypeParams(collateralType string) QueryCdpsByCollateralTypeParams {
	return QueryCdpsByCollateralTypeParams{
		CollateralType: collateralType,
	}
}

// QueryCdpsByRatioParams params for query /cdp/cdps/ratio
type QueryCdpsByRatioParams struct {
	CollateralType string
	Ratio          sdk.Dec // get CDPs below this collateral:debt ratio
}

// NewQueryCdpsByRatioParams returns QueryCdpsByRatioParams
func NewQueryCdpsByRatioParams(collateralType string, ratio sdk.Dec) QueryCdpsByRatioParams {
	return QueryCdpsByRatioParams{
		CollateralType: collateralType,
		Ratio:          ratio,
	}
}

// QueryGetTotalPrincipalParams params for query /cdp/totalPrincipal
type QueryGetTotalPrincipalParams struct {
	CollateralType string
}

// NewQueryGetTotalPrincipalParams returns QueryGetTotalPrincipalParams
func NewQueryGetTotalPrincipalParams(collateralType string) QueryGetTotalPrincipalParams {
	return QueryGetTotalPrincipalParams{
		CollateralType: collateralType,
	}
}

// QueryGetTotalCollateralParams params for query /cdp/totalCollateral
type QueryGetTotalCollateralParams struct {
	CollateralType string
}

// NewQueryGetTotalCollateralParams returns QueryGetTotalCollateralParams
func NewQueryGetTotalCollateralParams(collateralType string) QueryGetTotalCollateralParams {
	return QueryGetTotalCollateralParams{
		CollateralType: collateralType,
	}
}
