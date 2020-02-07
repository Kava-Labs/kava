package types

const (
	// QueryGetHTLT command for getting the information about a particular HTLT
	QueryGetHTLT = "htlt"
	// QueryGetHTLTs command for getting a list of HTLTs
	QueryGetHTLTs  = "htlts"
	QueryGetParams = "params"
)

// // Params for query 'custom/atomicswap/swapid'
// type QuerySwapByID struct {
// 	SwapID SwapBytes
// }

// Params for query 'custom/atomicswap/swapcreator'
// type QuerySwapByCreatorParams struct {
// 	Creator AccAddress
// 	Limit   int64
// 	Offset  int64
// }

// Params for query 'custom/atomicswap/swaprecipient'
// type QuerySwapByRecipientParams struct {
// 	Recipient AccAddress
// 	Limit     int64
// 	Offset    int64
// }

// implement fmt.Stringer
// func (n QueryResHTLTs) String() string {
// 	return strings.Join(n[:], "\n")
// }

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
