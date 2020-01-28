package types

import (
	"strings"
)

const (
	// QueryGetKHTLT command for getting the information about a particular HTLT
	QueryGetKHTLT = "htlt"
	// QueryGetKHTLTs command for getting a list of KHTLTs
	QueryGetKHTLTs = "htlts"
	QueryGetParams = "params"
)

// QueryResKHTLTs Result Payload for a KHTLTs query
type QueryResKHTLTs []string

// implement fmt.Stringer
func (n QueryResKHTLTs) String() string {
	return strings.Join(n[:], "\n")
}

// QueryKHTLTsParams is the params for a KHTLTs query
type QueryKHTLTsParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
}

// NewQueryKHTLTsParams creates a new QueryKHTLTsParams
func NewQueryKHTLTsParams(page int, limit int) QueryKHTLTsParams {
	return QueryKHTLTsParams{
		Page:  page,
		Limit: limit,
	}
}
