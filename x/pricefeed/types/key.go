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

	// RawPriceFeedPrefix prefix for the raw pricefeed of an asset
	RawPriceFeedPrefix = StoreKey + ":raw:"

	// CurrentPricePrefix prefix for the current price of an asset
	CurrentPricePrefix = StoreKey + ":currentprice:"

	// AssetPrefix Prefix for the assets in the pricefeed system
	AssetPrefix = StoreKey + ":assets"

	// OraclePrefix store prefix for the oracle accounts
	OraclePrefix = StoreKey + ":oracles"
)
