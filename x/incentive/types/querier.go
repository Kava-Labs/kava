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
	Owner          sdk.AccAddress
	CollateralType string
}

// NewQueryClaimsParams returns QueryClaimsParams
func NewQueryClaimsParams(owner sdk.AccAddress, collateralType string) QueryClaimsParams {
	return QueryClaimsParams{
		Owner:          owner,
		CollateralType: collateralType,
	}
}

// PostClaimReq defines the properties of claim transaction's request body.
type PostClaimReq struct {
	BaseReq        rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Sender         sdk.AccAddress `json:"sender" yaml:"sender"`
	CollateralType string         `json:"collateral_type" yaml:"collateral_type"`
	MultiplierName string         `json:"multiplier_name" yaml:"multiplier_name"`
}
