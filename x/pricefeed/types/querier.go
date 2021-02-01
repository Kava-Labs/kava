package types

// price Takes an [assetcode] and returns CurrentPrice for that asset
// pricefeed Takes an [assetcode] and returns the raw []PostedPrice for that asset
// assets Returns []Assets in the pricefeed system

const (
	// QueryGetParams command for params query
	QueryGetParams = "parameters"
	// QueryMarkets command for assets query
	QueryMarkets = "markets"
	// QueryOracles command for oracles query
	QueryOracles = "oracles"
	// QueryRawPrices command for raw price queries
	QueryRawPrices = "rawprices"
	// QueryPrice command for price queries
	QueryPrice = "price"
	// QueryPrices command for quering all prices
	QueryPrices = "prices"
)

// QueryWithMarketIDParams fields for querying information from a specific market
type QueryWithMarketIDParams struct {
	MarketID string
}

// NewQueryWithMarketIDParams creates a new instance of QueryWithMarketIDParams
func NewQueryWithMarketIDParams(marketID string) QueryWithMarketIDParams {
	return QueryWithMarketIDParams{
		MarketID: marketID,
	}
}
