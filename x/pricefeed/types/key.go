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
)

// CurrentPriceKey returns the prefix for the current price
func CurrentPriceKey(marketId string) []byte {
	return append(CurrentPricePrefix, []byte(marketId)...)
}

// RawPriceKey returns the prefix for the raw price
func RawPriceKey(marketId string, oracleAddr string) []byte {
	return append(append(RawPriceFeedPrefix, []byte(marketId)...), []byte(oracleAddr)...)
}
