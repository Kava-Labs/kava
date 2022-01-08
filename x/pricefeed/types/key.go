package types

import sdk "github.com/cosmos/cosmos-sdk/types"

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
func CurrentPriceKey(marketID string) []byte {
	return append(CurrentPricePrefix, []byte(marketID)...)
}

// RawPriceIteratorKey returns the prefix for the raw price for a single market
func RawPriceIteratorKey(marketID string) []byte {
	return append(
		RawPriceFeedPrefix,
		lengthPrefixWithByte([]byte(marketID))...,
	)
}

// RawPriceKey returns the prefix for the raw price
func RawPriceKey(marketID string, oracleAddr sdk.AccAddress) []byte {
	return append(
		RawPriceIteratorKey(marketID),
		lengthPrefixWithByte(oracleAddr)...,
	)
}

// lengthPrefixWithByte returns the input bytes prefixes with one byte containing its length.
// It panics if the input is greater than 255 in length.
func lengthPrefixWithByte(bz []byte) []byte {
	length := len(bz)

	if length > 255 {
		panic("cannot length prefix more than 255 bytes with single byte")
	}

	return append([]byte{byte(length)}, bz...)
}
