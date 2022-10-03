package types

import "github.com/cosmos/cosmos-sdk/types/query"

// NewQueryParamsRequest returns a new QueryParamsRequest
func NewQueryParamsRequest() *QueryParamsRequest {
	return &QueryParamsRequest{}
}

// NewQueryVaultsRequest returns a new QueryVaultsRequest
func NewQueryVaultsRequest() *QueryVaultsRequest {
	return &QueryVaultsRequest{}
}

// NewQueryVaultRequest returns a new QueryVaultRequest
func NewQueryVaultRequest(denom string) *QueryVaultRequest {
	return &QueryVaultRequest{
		Denom: denom,
	}
}

// NewQueryDepositsRequest returns a new QueryDepositsRequest
func NewQueryDepositsRequest(
	depositor string,
	denom string,
	ValueInStakedTokens bool,
	pagination *query.PageRequest,
) *QueryDepositsRequest {
	return &QueryDepositsRequest{
		Depositor:           depositor,
		Denom:               denom,
		ValueInStakedTokens: ValueInStakedTokens,
		Pagination:          pagination,
	}
}
