package types

const (
	// ModuleName The name that will be used throughout the module
	ModuleName = "pricefeed"

	// StoreKey Top level store key where all module items will be stored
	StoreKey = ModuleName

	// RouterKey Top level router key
	RouterKey = ModuleName

	// QuerierRoute is the querier route for gov
	QuerierRoute = ModuleName

	// DefaultParamspace default namestore
	DefaultParamspace = ModuleName
)

var (
	// CurrentPricePrefix prefix for the current price of an asset
	CurrentPricePrefix = []byte{0x00}

	// RawPriceFeedPrefix prefix for the raw pricefeed of an asset
	RawPriceFeedPrefix = []byte{0x01}

	// RawPriceFeedSuffix suffix for the rawpricefeed of an asset
	RawPriceFeedSuffix = []byte{0x00}
)

// CurrentPriceKey returns the prefix for the current price
func CurrentPriceKey(marketId string) []byte {
	return append(CurrentPricePrefix, []byte(marketId)...)
}

// RawPriceMarketKey returns the prefix for the raw price for a single market
func RawPriceMarketKey(marketId string) []byte {
	return append(append(RawPriceFeedPrefix, []byte(marketId)...), RawPriceFeedSuffix...)
}

// RawPriceKey returns the prefix for the raw price
func RawPriceKey(marketId string, oracleAddr string) []byte {
	parts := [][]byte{
		RawPriceFeedPrefix,
		[]byte(marketId),
		RawPriceFeedSuffix,
		[]byte(oracleAddr),
	}

	var key []byte

	for _, part := range parts {
		key = append(key, part...)
	}

	return key
}
