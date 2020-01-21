package types

// price Takes an [assetcode] and returns CurrentPrice for that asset
// pricefeed Takes an [assetcode] and returns the raw []PostedPrice for that asset
// assets Returns []Assets in the pricefeed system

const (
	// QueryCurrentPrice command for current price queries
	QueryCurrentPrice = "price"
	// QueryRawPrices command for raw price queries
	QueryRawPrices = "rawprices"
	// QueryMarkets command for assets query
	QueryMarkets = "markets"
	// QueryGetParams command for params query
	QueryGetParams = "params"
)

// QueryPricesParams fields for querying prices
type QueryPricesParams struct {
	MarketID string
}
