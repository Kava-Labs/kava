package types

// price Takes an [assetcode] and returns CurrentPrice for that asset
// pricefeed Takes an [assetcode] and returns the raw []PostedPrice for that asset
// assets Returns []Assets in the pricefeed system

const (
	// QueryPrice command for current price queries
	QueryPrice = "price"
	// QueryRawPrices command for raw price queries
	QueryRawPrices = "rawprices"
	// QueryMarkets command for assets query
	QueryMarkets = "markets"
	// QueryGetParams command for params query
	QueryGetParams = "parameters"
)

// QueryPriceParams fields for querying prices
type QueryPriceParams struct {
	MarketID string
}

// NewQueryPriceParams creates a new instance of QueryPriceParams
func NewQueryPriceParams(marketID string) QueryPriceParams {
	return QueryPriceParams{
		MarketID: marketID,
	}
}
