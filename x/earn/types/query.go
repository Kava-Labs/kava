package types

import "github.com/cosmos/cosmos-sdk/types/query"

// NewQueryParamsRequest returns a new QueryParamsRequest
func NewQueryParamsRequest() *QueryParamsRequest {
	return &QueryParamsRequest{}
}

// NewQueryVaultsRequest returns a new QueryVaultsRequest
func NewQueryVaultsRequest(denom string) *QueryVaultsRequest {
	return &QueryVaultsRequest{
		Denom: denom,
	}
}

// NewQueryDepositsRequest returns a new QueryDepositsRequest
func NewQueryDepositsRequest(
	owner string,
	denom string,
	pagination *query.PageRequest,
) *QueryDepositsRequest {
	return &QueryDepositsRequest{
		Owner:      owner,
		Denom:      denom,
		Pagination: pagination,
	}
}
