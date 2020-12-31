package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

// Querier routes for the incentive module
const (
	QueryGetClaims          = "claims"
	RestClaimOwner          = "owner"
	RestClaimCollateralType = "collateral_type"
	QueryGetParams          = "parameters"
	QueryGetRewardPeriods   = "reward-periods"
	QueryGetClaimPeriods    = "claim-periods"
)

// QueryClaimsParams params for query /incentive/claims
type QueryClaimsParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
	Owner sdk.AccAddress
}

// NewQueryClaimsParams returns QueryClaimsParams
func NewQueryClaimsParams(page, limit int, owner sdk.AccAddress) QueryClaimsParams {
	return QueryClaimsParams{
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
