package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Querier routes for the cdp module
const (
	QueryGetCdp                     = "cdp"
	QueryGetCdpDeposits             = "deposits"
	QueryGetCdps                    = "cdps"
	QueryGetV2Cdps                  = "v2cdps"
	QueryGetCdpsByCollateralization = "ratio"
	QueryGetParams                  = "params"
	QueryGetAccounts                = "accounts"
	RestOwner                       = "owner"
	RestCollateralType              = "collateral-type"
	RestRatio                       = "ratio"
)

// QueryCdpsParams params for query /cdp/cdps
type QueryCdpsParams struct {
	CollateralType string // get CDPs with this collateral type
}

// NewQueryCdpsParams returns QueryCdpsParams
func NewQueryCdpsParams(collateralType string) QueryCdpsParams {
	return QueryCdpsParams{
		CollateralType: collateralType,
	}
}

// QueryV2CdpsParams is the params for a filtered CDP query
type QueryV2CdpsParams struct {
	Page            int            `json:"page" yaml:"page"`
	Limit           int            `json:"limit" yaml:"limit"`
	CollateralDenom string         `json:"collateral_denom" yaml:"collateral_denom"`
	Owner           sdk.AccAddress `json:"owner" yaml:"owner"`
	ID              uint64         `json:"id" yaml:"id"`
}

// NewQueryV2CdpsParams creates a new QueryV2CdpsParams
func NewQueryV2CdpsParams(page, limit int, collateralDenom string, owner sdk.AccAddress, id uint64) QueryV2CdpsParams {
	return QueryV2CdpsParams{
		Page:            page,
		Limit:           limit,
		CollateralDenom: collateralDenom,
		Owner:           owner,
		ID:              id,
	}
}

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
