package types

const (
	// QueryGetHTLT command for getting the information about a particular HTLT
	QueryGetHTLT = "htlt"
	// QueryGetHTLTs command for getting a list of HTLTs
	QueryGetHTLTs  = "htlts"
	QueryGetParams = "params"
)

// QueryHTLTsParams is the params for a HTLTs query
type QueryHTLTsParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
}

// NewQueryHTLTsParams creates a new QueryHTLTsParams
func NewQueryHTLTsParams(page int, limit int) QueryHTLTsParams {
	return QueryHTLTsParams{
		Page:  page,
		Limit: limit,
	}
}
