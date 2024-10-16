package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// QueryCdpsParams is the params for a filtered CDP query
type QueryCdpsParams struct {
	Page           int               `json:"page" yaml:"page"`
	Limit          int               `json:"limit" yaml:"limit"`
	CollateralType string            `json:"collateral_type" yaml:"collateral_type"`
	Owner          sdk.AccAddress    `json:"owner" yaml:"owner"`
	ID             uint64            `json:"id" yaml:"id"`
	Ratio          sdkmath.LegacyDec `json:"ratio" yaml:"ratio"`
}

// NewQueryCdpsParams creates a new QueryCdpsParams
func NewQueryCdpsParams(page, limit int, collateralType string, owner sdk.AccAddress, id uint64, ratio sdkmath.LegacyDec) QueryCdpsParams {
	return QueryCdpsParams{
		Page:           page,
		Limit:          limit,
		CollateralType: collateralType,
		Owner:          owner,
		ID:             id,
		Ratio:          ratio,
	}
}
