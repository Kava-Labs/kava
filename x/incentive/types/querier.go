package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

// Querier routes for the incentive module
const (
	QueryGetClaims = "claims"
	RestClaimOwner = "owner"
	RestClaimDenom = "denom"
	QueryGetParams = "parameters"
)

// QueryClaimsParams params for query /incentive/claims
type QueryClaimsParams struct {
	Owner sdk.AccAddress
	Denom string
}

// NewQueryClaimsParams returns QueryClaimsParams
func NewQueryClaimsParams(owner sdk.AccAddress, denom string) QueryClaimsParams {
	return QueryClaimsParams{
		Owner: owner,
		Denom: denom,
	}
}

// PostClaimReq defines the properties of claim transaction's request body.
type PostClaimReq struct {
	BaseReq rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Sender  sdk.AccAddress `json:"sender" yaml:"sender"`
	Denom   string         `json:"denom" yaml:"denom"`
}
