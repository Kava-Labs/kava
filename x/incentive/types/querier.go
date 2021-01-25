package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

// Querier routes for the incentive module
const (
	QueryGetCdpClaims       = "cdp-claims"
	QueryGetHardClaims      = "hard-claims"
	RestClaimOwner          = "owner"
	RestClaimCollateralType = "collateral_type"
	QueryGetParams          = "parameters"
	QueryGetRewardPeriods   = "reward-periods"
	QueryGetClaimPeriods    = "claim-periods"
)

// QueryCdpClaimsParams params for query /incentive/claims
type QueryCdpClaimsParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
	Owner sdk.AccAddress
}

// NewQueryCdpClaimsParams returns QueryCdpClaimsParams
func NewQueryCdpClaimsParams(page, limit int, owner sdk.AccAddress) QueryCdpClaimsParams {
	return QueryCdpClaimsParams{
		Page:  page,
		Limit: limit,
		Owner: owner,
	}
}

// QueryHardClaimsParams params for query /incentive/claims
type QueryHardClaimsParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
	Owner sdk.AccAddress
}

// NewQueryHardClaimsParams returns QueryHardClaimsParams
func NewQueryHardClaimsParams(page, limit int, owner sdk.AccAddress) QueryHardClaimsParams {
	return QueryHardClaimsParams{
		Page:  page,
		Limit: limit,
		Owner: owner,
	}
}

// PostClaimReq defines the properties of claim transaction's request body.
type PostClaimReq struct {
	BaseReq        rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Sender         sdk.AccAddress `json:"sender" yaml:"sender"`
	MultiplierName string         `json:"multiplier_name" yaml:"multiplier_name"`
}
